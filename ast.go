package main

import "fmt"

// ── Type system ───────────────────────────────────────────────────────────────

type GoType int

const (
	TypeUnknown GoType = iota
	TypeInt
	TypeString
	TypeBool
	TypeVoid
)

func (t GoType) String() string {
	switch t {
	case TypeInt:
		return "int"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeVoid:
		return "void"
	default:
		return "unknown"
	}
}

func (t GoType) CName() string {
	switch t {
	case TypeInt:
		return "long long"
	case TypeString:
		return "const char*"
	case TypeBool:
		return "int"
	case TypeVoid:
		return "void"
	default:
		return "void*"
	}
}

// ── AST interfaces ────────────────────────────────────────────────────────────

type Node interface {
	nodeTag()
	Pos() (int, int) // line, col
}

type Stmt interface {
	Node
	stmtTag()
}

type Expr interface {
	Node
	exprTag()
}

// ── Program ───────────────────────────────────────────────────────────────────

type Program struct {
	Package string
	Imports []string
	Funcs   []*FuncDecl
}

// ── Declarations ──────────────────────────────────────────────────────────────

type Param struct {
	Name string
	Type GoType
}

type FuncDecl struct {
	Name       string
	Params     []Param
	ReturnType GoType
	Body       *BlockStmt
	Line, Col  int
}

func (n *FuncDecl) nodeTag()       {}
func (n *FuncDecl) Pos() (int, int) { return n.Line, n.Col }

// ── Statements ────────────────────────────────────────────────────────────────

type BlockStmt struct {
	Stmts     []Stmt
	Line, Col int
}

func (n *BlockStmt) nodeTag()        {}
func (n *BlockStmt) stmtTag()        {}
func (n *BlockStmt) Pos() (int, int) { return n.Line, n.Col }

type VarDeclStmt struct {
	Name      string
	Type      GoType // may be TypeUnknown (inferred)
	Value     Expr   // may be nil
	Line, Col int
}

func (n *VarDeclStmt) nodeTag()        {}
func (n *VarDeclStmt) stmtTag()        {}
func (n *VarDeclStmt) Pos() (int, int) { return n.Line, n.Col }

type AssignStmt struct {
	Target    Expr
	Value     Expr
	Line, Col int
}

func (n *AssignStmt) nodeTag()        {}
func (n *AssignStmt) stmtTag()        {}
func (n *AssignStmt) Pos() (int, int) { return n.Line, n.Col }

type ShortVarDeclStmt struct {
	Names     []string
	Values    []Expr
	Line, Col int
}

func (n *ShortVarDeclStmt) nodeTag()        {}
func (n *ShortVarDeclStmt) stmtTag()        {}
func (n *ShortVarDeclStmt) Pos() (int, int) { return n.Line, n.Col }

type ReturnStmt struct {
	Value     Expr // nil for bare return
	Line, Col int
}

func (n *ReturnStmt) nodeTag()        {}
func (n *ReturnStmt) stmtTag()        {}
func (n *ReturnStmt) Pos() (int, int) { return n.Line, n.Col }

type IfStmt struct {
	Cond      Expr
	Then      *BlockStmt
	Else      Stmt // *BlockStmt or *IfStmt or nil
	Line, Col int
}

func (n *IfStmt) nodeTag()        {}
func (n *IfStmt) stmtTag()        {}
func (n *IfStmt) Pos() (int, int) { return n.Line, n.Col }

// ForStmt covers:  for { }  /  for cond { }  /  for init; cond; post { }
type ForStmt struct {
	Init      Stmt // may be nil
	Cond      Expr // may be nil (infinite loop)
	Post      Stmt // may be nil
	Body      *BlockStmt
	Line, Col int
}

func (n *ForStmt) nodeTag()        {}
func (n *ForStmt) stmtTag()        {}
func (n *ForStmt) Pos() (int, int) { return n.Line, n.Col }

type BreakStmt struct {
	Line, Col int
}

func (n *BreakStmt) nodeTag()        {}
func (n *BreakStmt) stmtTag()        {}
func (n *BreakStmt) Pos() (int, int) { return n.Line, n.Col }

type ContinueStmt struct {
	Line, Col int
}

func (n *ContinueStmt) nodeTag()        {}
func (n *ContinueStmt) stmtTag()        {}
func (n *ContinueStmt) Pos() (int, int) { return n.Line, n.Col }

type ExprStmt struct {
	Expr      Expr
	Line, Col int
}

func (n *ExprStmt) nodeTag()        {}
func (n *ExprStmt) stmtTag()        {}
func (n *ExprStmt) Pos() (int, int) { return n.Line, n.Col }

type IncDecStmt struct {
	Target    Expr
	Op        string // "++" or "--"
	Line, Col int
}

func (n *IncDecStmt) nodeTag()        {}
func (n *IncDecStmt) stmtTag()        {}
func (n *IncDecStmt) Pos() (int, int) { return n.Line, n.Col }

// ── Expressions ───────────────────────────────────────────────────────────────

type IntLit struct {
	Value     int64
	Line, Col int
}

func (n *IntLit) nodeTag()        {}
func (n *IntLit) exprTag()        {}
func (n *IntLit) Pos() (int, int) { return n.Line, n.Col }

type StringLit struct {
	Value     string
	Line, Col int
}

func (n *StringLit) nodeTag()        {}
func (n *StringLit) exprTag()        {}
func (n *StringLit) Pos() (int, int) { return n.Line, n.Col }

type BoolLit struct {
	Value     bool
	Line, Col int
}

func (n *BoolLit) nodeTag()        {}
func (n *BoolLit) exprTag()        {}
func (n *BoolLit) Pos() (int, int) { return n.Line, n.Col }

type Identifier struct {
	Name      string
	Line, Col int
}

func (n *Identifier) nodeTag()        {}
func (n *Identifier) exprTag()        {}
func (n *Identifier) Pos() (int, int) { return n.Line, n.Col }

type BinaryExpr struct {
	Op        string
	Left      Expr
	Right     Expr
	Line, Col int
}

func (n *BinaryExpr) nodeTag()        {}
func (n *BinaryExpr) exprTag()        {}
func (n *BinaryExpr) Pos() (int, int) { return n.Line, n.Col }

type UnaryExpr struct {
	Op        string
	Operand   Expr
	Line, Col int
}

func (n *UnaryExpr) nodeTag()        {}
func (n *UnaryExpr) exprTag()        {}
func (n *UnaryExpr) Pos() (int, int) { return n.Line, n.Col }

type CallExpr struct {
	Func      Expr   // Identifier or SelectorExpr
	Args      []Expr
	Line, Col int
}

func (n *CallExpr) nodeTag()        {}
func (n *CallExpr) exprTag()        {}
func (n *CallExpr) Pos() (int, int) { return n.Line, n.Col }

// pkg.Name  e.g. fmt.Println
type SelectorExpr struct {
	Object    Expr
	Field     string
	Line, Col int
}

func (n *SelectorExpr) nodeTag()        {}
func (n *SelectorExpr) exprTag()        {}
func (n *SelectorExpr) Pos() (int, int) { return n.Line, n.Col }

// ── pretty-printer (debug) ────────────────────────────────────────────────────

func PrintAST(n Node, indent int) {
	pad := func() string { return fmt.Sprintf("%*s", indent*2, "") }
	switch v := n.(type) {
	case *FuncDecl:
		fmt.Printf("%sFuncDecl %s -> %s\n", pad(), v.Name, v.ReturnType)
		PrintAST(v.Body, indent+1)
	case *BlockStmt:
		fmt.Printf("%sBlock\n", pad())
		for _, s := range v.Stmts {
			PrintAST(s, indent+1)
		}
	case *VarDeclStmt:
		fmt.Printf("%sVarDecl %s %s\n", pad(), v.Name, v.Type)
	case *ShortVarDeclStmt:
		fmt.Printf("%sShortDecl %v\n", pad(), v.Names)
	case *AssignStmt:
		fmt.Printf("%sAssign\n", pad())
	case *ReturnStmt:
		fmt.Printf("%sReturn\n", pad())
	case *IfStmt:
		fmt.Printf("%sIf\n", pad())
	case *ForStmt:
		fmt.Printf("%sFor\n", pad())
	case *ExprStmt:
		fmt.Printf("%sExprStmt\n", pad())
		PrintAST(v.Expr, indent+1)
	case *CallExpr:
		fmt.Printf("%sCall\n", pad())
	case *BinaryExpr:
		fmt.Printf("%sBinary %s\n", pad(), v.Op)
	case *Identifier:
		fmt.Printf("%sIdent %s\n", pad(), v.Name)
	case *IntLit:
		fmt.Printf("%sInt %d\n", pad(), v.Value)
	case *StringLit:
		fmt.Printf("%sString %q\n", pad(), v.Value)
	default:
		fmt.Printf("%s<%T>\n", pad(), n)
	}
}
