package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const version = "0.1.0"

const helpText = `goc — mini Go compiler v` + version + `

USAGE:
  goc [flags] <source.go>

FLAGS:
  -o <file>   Output binary name (default: a.out)
  -S          Emit C source only (do not invoke gcc)
  -ast        Print AST and exit
  -lex        Print tokens and exit
  -v          Verbose: print stage timings
  -h          Show this help

SUPPORTED LANGUAGE FEATURES:
  • package / import declarations
  • func declarations with typed parameters and return values
  • Variables:  var x int  /  x := expr  /  x = expr
  • Types:  int  string  bool
  • Control flow:  if / else / else if
  • Loops:  for { }  /  for cond { }  /  for init; cond; post { }
  • break / continue
  • Arithmetic:  + - * / %
  • Comparison:  == != < > <= >=
  • Logic:  && || !
  • fmt.Print / fmt.Println / fmt.Printf / fmt.Sprintf
  • User-defined function calls

EXAMPLE:
  goc -o hello hello.go && ./hello
`

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "goc: "+format+"\n", a...)
	os.Exit(1)
}

func main() {
	var (
		output   = "a.out"
		emitC    = false
		printAST = false
		printLex = false
		verbose  = false
		srcFile  = ""
	)

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Print(helpText)
			os.Exit(0)
		case "-o":
			i++
			if i >= len(args) {
				die("-o requires an argument")
			}
			output = args[i]
		case "-S":
			emitC = true
		case "-ast":
			printAST = true
		case "-lex":
			printLex = true
		case "-v":
			verbose = true
		default:
			if strings.HasPrefix(args[i], "-") {
				die("unknown flag: %s", args[i])
			}
			if srcFile != "" {
				die("only one source file supported")
			}
			srcFile = args[i]
		}
	}

	if srcFile == "" {
		fmt.Print(helpText)
		os.Exit(1)
	}

	// ── Read source ──────────────────────────────────────────────────────────
	src, err := os.ReadFile(srcFile)
	if err != nil {
		die("cannot read %s: %v", srcFile, err)
	}
	logf := func(format string, a ...interface{}) {
		if verbose {
			fmt.Fprintf(os.Stderr, "[goc] "+format+"\n", a...)
		}
	}

	// ── Lex ──────────────────────────────────────────────────────────────────
	logf("lexing %s", srcFile)
	lexer := NewLexer(string(src))
	tokens, err := lexer.Tokenize()
	if err != nil {
		die("lex error: %v", err)
	}

	if printLex {
		for _, t := range tokens {
			fmt.Println(t)
		}
		return
	}

	// ── Parse ─────────────────────────────────────────────────────────────────
	logf("parsing")
	parser := NewParser(tokens)
	prog, err := parser.ParseProgram()
	if err != nil {
		die("parse error: %v", err)
	}

	if printAST {
		for _, fn := range prog.Funcs {
			PrintAST(fn, 0)
		}
		return
	}

	// ── Codegen ───────────────────────────────────────────────────────────────
	logf("generating C")
	cg := NewCodegen()
	cSrc, err := cg.Generate(prog)
	if err != nil {
		die("codegen error: %v", err)
	}

	if emitC {
		// Write .c file next to output
		cFile := strings.TrimSuffix(output, filepath.Ext(output)) + ".c"
		if err := os.WriteFile(cFile, []byte(cSrc), 0644); err != nil {
			die("cannot write %s: %v", cFile, err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", cFile)
		return
	}

	// ── Compile C via gcc ─────────────────────────────────────────────────────
	logf("compiling C via gcc")
	tmpC, err := os.CreateTemp("", "goc_*.c")
	if err != nil {
		die("temp file: %v", err)
	}
	defer os.Remove(tmpC.Name())

	if _, err := tmpC.WriteString(cSrc); err != nil {
		die("write temp: %v", err)
	}
	tmpC.Close()

	cmd := exec.Command("gcc", "-O2", "-o", output, tmpC.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// On failure, dump the C source for debugging
		if verbose {
			fmt.Fprintln(os.Stderr, "\n─── generated C source ───")
			fmt.Fprintln(os.Stderr, cSrc)
		}
		die("gcc failed: %v", err)
	}
	logf("wrote binary %s", output)
}
