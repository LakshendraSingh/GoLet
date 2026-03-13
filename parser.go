package main

import (
	"fmt"
	"strconv"
)

// ── Parser ────────────────────────────────────────────────────────────────────

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TOK_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekKind() TokenKind { return p.peek().Kind }

func (p *Parser) advance() Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) expect(k TokenKind) (Token, error) {
	t := p.peek()
	if t.Kind != k {
		return t, fmt.Errorf("%d:%d: expected %s, got %s (%q)", t.Line, t.Col, k, t.Kind, t.Value)
	}
	p.advance()
	return t, nil
}

func (p *Parser) skip(k TokenKind) bool {
	if p.peekKind() == k {
		p.advance()
		return true
	}
	return false
}

// ── Top level ─────────────────────────────────────────────────────────────────

func (p *Parser) ParseProgram() (*Program, error) {
	prog := &Program{}

	// package clause
	if _, err := p.expect(TOK_PACKAGE); err != nil {
		return nil, err
	}
	nameT, err := p.expect(TOK_IDENT)
	if err != nil {
		return nil, err
	}
	prog.Package = nameT.Value
	p.skip(TOK_SEMICOLON)

	// import(s)
	for p.peekKind() == TOK_IMPORT {
		p.advance()
		if p.peekKind() == TOK_LPAREN {
			p.advance()
			for p.peekKind() != TOK_RPAREN && p.peekKind() != TOK_EOF {
				s, err := p.expect(TOK_STRING)
				if err != nil {
					return nil, err
				}
				prog.Imports = append(prog.Imports, s.Value)
				p.skip(TOK_SEMICOLON)
			}
			if _, err := p.expect(TOK_RPAREN); err != nil {
				return nil, err
			}
		} else {
			s, err := p.expect(TOK_STRING)
			if err != nil {
				return nil, err
			}
			prog.Imports = append(prog.Imports, s.Value)
		}
		p.skip(TOK_SEMICOLON)
	}

	// top-level declarations (only func for now)
	for p.peekKind() != TOK_EOF {
		switch p.peekKind() {
		case TOK_FUNC:
			fn, err := p.parseFunc()
			if err != nil {
				return nil, err
			}
			prog.Funcs = append(prog.Funcs, fn)
		default:
			t := p.peek()
			return nil, fmt.Errorf("%d:%d: unexpected top-level token %s", t.Line, t.Col, t.Kind)
		}
		p.skip(TOK_SEMICOLON)
	}
	return prog, nil
}

// ── Functions ─────────────────────────────────────────────────────────────────

func (p *Parser) parseFunc() (*FuncDecl, error) {
	t, _ := p.expect(TOK_FUNC)
	fn := &FuncDecl{Line: t.Line, Col: t.Col}

	name, err := p.expect(TOK_IDENT)
	if err != nil {
		return nil, err
	}
	fn.Name = name.Value

	if _, err := p.expect(TOK_LPAREN); err != nil {
		return nil, err
	}
	for p.peekKind() != TOK_RPAREN && p.peekKind() != TOK_EOF {
		pname, err := p.expect(TOK_IDENT)
		if err != nil {
			return nil, err
		}
		ptype, err := p.parseType()
		if err != nil {
			return nil, err
		}
		fn.Params = append(fn.Params, Param{Name: pname.Value, Type: ptype})
		p.skip(TOK_COMMA)
	}
	if _, err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}

	// optional return type
	fn.ReturnType = TypeVoid
	if p.peekKind() == TOK_TYPE_INT || p.peekKind() == TOK_TYPE_STRING || p.peekKind() == TOK_TYPE_BOOL {
		fn.ReturnType, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	fn.Body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	p.skip(TOK_SEMICOLON)
	return fn, nil
}

func (p *Parser) parseType() (GoType, error) {
	t := p.peek()
	switch t.Kind {
	case TOK_TYPE_INT:
		p.advance()
		return TypeInt, nil
	case TOK_TYPE_STRING:
		p.advance()
		return TypeString, nil
	case TOK_TYPE_BOOL:
		p.advance()
		return TypeBool, nil
	default:
		return TypeUnknown, fmt.Errorf("%d:%d: expected type, got %s", t.Line, t.Col, t.Kind)
	}
}

// ── Statements ────────────────────────────────────────────────────────────────

func (p *Parser) parseBlock() (*BlockStmt, error) {
	t, err := p.expect(TOK_LBRACE)
	if err != nil {
		return nil, err
	}
	block := &BlockStmt{Line: t.Line, Col: t.Col}
	for p.peekKind() != TOK_RBRACE && p.peekKind() != TOK_EOF {
		p.skip(TOK_SEMICOLON)
		if p.peekKind() == TOK_RBRACE {
			break
		}
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if s != nil {
			block.Stmts = append(block.Stmts, s)
		}
		p.skip(TOK_SEMICOLON)
	}
	if _, err := p.expect(TOK_RBRACE); err != nil {
		return nil, err
	}
	return block, nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	t := p.peek()
	switch t.Kind {
	case TOK_VAR:
		return p.parseVarDecl()
	case TOK_RETURN:
		return p.parseReturn()
	case TOK_IF:
		return p.parseIf()
	case TOK_FOR:
		return p.parseFor()
	case TOK_BREAK:
		p.advance()
		return &BreakStmt{Line: t.Line, Col: t.Col}, nil
	case TOK_CONTINUE:
		p.advance()
		return &ContinueStmt{Line: t.Line, Col: t.Col}, nil
	case TOK_LBRACE:
		return p.parseBlock()
	default:
		return p.parseSimpleStmt()
	}
}

func (p *Parser) parseVarDecl() (*VarDeclStmt, error) {
	t, _ := p.expect(TOK_VAR)
	name, err := p.expect(TOK_IDENT)
	if err != nil {
		return nil, err
	}
	stmt := &VarDeclStmt{Name: name.Value, Line: t.Line, Col: t.Col}

	// optional explicit type
	if p.peekKind() == TOK_TYPE_INT || p.peekKind() == TOK_TYPE_STRING || p.peekKind() == TOK_TYPE_BOOL {
		stmt.Type, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	// optional initializer
	if p.skip(TOK_ASSIGN) {
		stmt.Value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

func (p *Parser) parseReturn() (*ReturnStmt, error) {
	t, _ := p.expect(TOK_RETURN)
	stmt := &ReturnStmt{Line: t.Line, Col: t.Col}
	// if next token is on same logical line and can start an expression
	if p.peekKind() != TOK_SEMICOLON && p.peekKind() != TOK_RBRACE && p.peekKind() != TOK_EOF {
		var err error
		stmt.Value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

func (p *Parser) parseIf() (*IfStmt, error) {
	t, _ := p.expect(TOK_IF)
	stmt := &IfStmt{Line: t.Line, Col: t.Col}
	var err error
	stmt.Cond, err = p.parseExpr()
	if err != nil {
		return nil, err
	}
	stmt.Then, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	// optional else
	p.skip(TOK_SEMICOLON)
	if p.peekKind() == TOK_ELSE {
		p.advance()
		if p.peekKind() == TOK_IF {
			stmt.Else, err = p.parseIf()
		} else {
			stmt.Else, err = p.parseBlock()
		}
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

func (p *Parser) parseFor() (*ForStmt, error) {
	t, _ := p.expect(TOK_FOR)
	stmt := &ForStmt{Line: t.Line, Col: t.Col}

	if p.peekKind() == TOK_LBRACE {
		// for { } — infinite loop
	} else {
		// Try to detect 3-part for: init ; cond ; post
		// vs simple for cond { }
		// Heuristic: if we see := or = after the first expr, it's 3-part
		saved := p.pos
		_, isThreePart, err := p.tryParseSimpleStmt()
		if err != nil {
			return nil, err
		}
		_ = saved

		if isThreePart && p.peekKind() == TOK_SEMICOLON {
			// restore and parse properly
			p.pos = saved
			init, err := p.parseSimpleStmt()
			if err != nil {
				return nil, err
			}
			stmt.Init = init
			if _, err := p.expect(TOK_SEMICOLON); err != nil {
				return nil, err
			}
			if p.peekKind() != TOK_SEMICOLON {
				stmt.Cond, err = p.parseExpr()
				if err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(TOK_SEMICOLON); err != nil {
				return nil, err
			}
			if p.peekKind() != TOK_LBRACE {
				stmt.Post, err = p.parseSimpleStmt()
				if err != nil {
					return nil, err
				}
			}
		} else {
			// for cond { }
			p.pos = saved
			stmt.Cond, err = p.parseExpr()
			if err != nil {
				return nil, err
			}
		}
	}

	var err error
	stmt.Body, err = p.parseBlock()
	return stmt, err
}

// tryParseSimpleStmt parses a simple statement speculatively and returns
// (stmt, isAssignment, error). Used to decide 3-part vs condition for.
func (p *Parser) tryParseSimpleStmt() (Stmt, bool, error) {
	s, err := p.parseSimpleStmt()
	if err != nil {
		return nil, false, err
	}
	switch s.(type) {
	case *ShortVarDeclStmt, *AssignStmt, *IncDecStmt:
		return s, true, nil
	default:
		return s, false, nil
	}
}

// parseSimpleStmt handles assignment, short-var-decl, inc/dec, and expr stmts.
func (p *Parser) parseSimpleStmt() (Stmt, error) {
	t := p.peek()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	switch p.peekKind() {
	case TOK_DECL_ASSIGN:
		p.advance()
		// multi-assign: a, b := x, y  (left side was already parsed as tuple/ident)
		names, err := exprToNames(expr)
		if err != nil {
			return nil, fmt.Errorf("%d:%d: invalid short variable declaration", t.Line, t.Col)
		}
		var vals []Expr
		for {
			v, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
			if !p.skip(TOK_COMMA) {
				break
			}
		}
		return &ShortVarDeclStmt{Names: names, Values: vals, Line: t.Line, Col: t.Col}, nil

	case TOK_ASSIGN:
		p.advance()
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &AssignStmt{Target: expr, Value: val, Line: t.Line, Col: t.Col}, nil
	}

	// inc/dec
	if p.skip(TOK_INCR) {
		return &IncDecStmt{Target: expr, Op: "++", Line: t.Line, Col: t.Col}, nil
	}
	if p.skip(TOK_DECR) {
		return &IncDecStmt{Target: expr, Op: "--", Line: t.Line, Col: t.Col}, nil
	}

	return &ExprStmt{Expr: expr, Line: t.Line, Col: t.Col}, nil
}

func exprToNames(e Expr) ([]string, error) {
	switch v := e.(type) {
	case *Identifier:
		return []string{v.Name}, nil
	default:
		return nil, fmt.Errorf("expected identifier(s)")
	}
}

// ── Expressions (Pratt / precedence climbing) ─────────────────────────────────

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseBinary(0)
}

func precedence(k TokenKind) int {
	switch k {
	case TOK_OR:
		return 1
	case TOK_AND:
		return 2
	case TOK_EQ, TOK_NEQ, TOK_LT, TOK_GT, TOK_LTE, TOK_GTE:
		return 3
	case TOK_PLUS, TOK_MINUS:
		return 4
	case TOK_STAR, TOK_SLASH, TOK_PERCENT:
		return 5
	}
	return -1
}

func (p *Parser) parseBinary(minPrec int) (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		prec := precedence(p.peekKind())
		if prec < minPrec {
			break
		}
		op := p.advance()
		right, err := p.parseBinary(prec + 1)
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op.Value, Left: left, Right: right, Line: op.Line, Col: op.Col}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	t := p.peek()
	switch t.Kind {
	case TOK_MINUS:
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "-", Operand: operand, Line: t.Line, Col: t.Col}, nil
	case TOK_NOT:
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: "!", Operand: operand, Line: t.Line, Col: t.Col}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peekKind() {
		case TOK_DOT:
			dot := p.advance()
			field, err := p.expect(TOK_IDENT)
			if err != nil {
				return nil, err
			}
			expr = &SelectorExpr{Object: expr, Field: field.Value, Line: dot.Line, Col: dot.Col}
		case TOK_LPAREN:
			t := p.advance()
			var args []Expr
			for p.peekKind() != TOK_RPAREN && p.peekKind() != TOK_EOF {
				arg, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				p.skip(TOK_COMMA)
			}
			if _, err := p.expect(TOK_RPAREN); err != nil {
				return nil, err
			}
			expr = &CallExpr{Func: expr, Args: args, Line: t.Line, Col: t.Col}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parsePrimary() (Expr, error) {
	t := p.peek()
	switch t.Kind {
	case TOK_INT:
		p.advance()
		v, _ := strconv.ParseInt(t.Value, 10, 64)
		return &IntLit{Value: v, Line: t.Line, Col: t.Col}, nil
	case TOK_STRING:
		p.advance()
		return &StringLit{Value: t.Value, Line: t.Line, Col: t.Col}, nil
	case TOK_TRUE:
		p.advance()
		return &BoolLit{Value: true, Line: t.Line, Col: t.Col}, nil
	case TOK_FALSE:
		p.advance()
		return &BoolLit{Value: false, Line: t.Line, Col: t.Col}, nil
	case TOK_IDENT:
		p.advance()
		return &Identifier{Name: t.Value, Line: t.Line, Col: t.Col}, nil
	case TOK_LPAREN:
		p.advance()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(TOK_RPAREN); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("%d:%d: unexpected token %s (%q) in expression", t.Line, t.Col, t.Kind, t.Value)
	}
}
