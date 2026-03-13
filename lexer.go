package main

import (
	"fmt"
	"strings"
	"unicode"
)

// ── Token types ──────────────────────────────────────────────────────────────

type TokenKind int

const (
	// Literals
	TOK_INT    TokenKind = iota // 42
	TOK_STRING                  // "hello"
	TOK_BOOL                    // true / false
	TOK_IDENT                   // foo

	// Keywords
	TOK_PACKAGE
	TOK_IMPORT
	TOK_FUNC
	TOK_RETURN
	TOK_VAR
	TOK_IF
	TOK_ELSE
	TOK_FOR
	TOK_BREAK
	TOK_CONTINUE
	TOK_TRUE
	TOK_FALSE

	// Types
	TOK_TYPE_INT
	TOK_TYPE_STRING
	TOK_TYPE_BOOL

	// Operators & punctuation
	TOK_INCR        // ++
	TOK_DECR        // --
	TOK_PLUS        // +
	TOK_MINUS       // -
	TOK_STAR        // *
	TOK_SLASH       // /
	TOK_PERCENT     // %
	TOK_ASSIGN      // =
	TOK_DECL_ASSIGN // :=
	TOK_EQ          // ==
	TOK_NEQ         // !=
	TOK_LT          // <
	TOK_GT          // >
	TOK_LTE         // <=
	TOK_GTE         // >=
	TOK_AND         // &&
	TOK_OR          // ||
	TOK_NOT         // !
	TOK_COMMA       // ,
	TOK_SEMICOLON   // ;
	TOK_COLON       // :
	TOK_LPAREN      // (
	TOK_RPAREN      // )
	TOK_LBRACE      // {
	TOK_RBRACE      // }
	TOK_LBRACKET    // [
	TOK_RBRACKET    // ]
	TOK_DOT         // .

	TOK_EOF
)

var tokenNames = map[TokenKind]string{
	TOK_INCR: "++", TOK_DECR: "--",
	TOK_INT: "INT", TOK_STRING: "STRING", TOK_BOOL: "BOOL", TOK_IDENT: "IDENT",
	TOK_PACKAGE: "package", TOK_IMPORT: "import", TOK_FUNC: "func",
	TOK_RETURN: "return", TOK_VAR: "var", TOK_IF: "if", TOK_ELSE: "else",
	TOK_FOR: "for", TOK_BREAK: "break", TOK_CONTINUE: "continue",
	TOK_TRUE: "true", TOK_FALSE: "false",
	TOK_TYPE_INT: "int", TOK_TYPE_STRING: "string", TOK_TYPE_BOOL: "bool",
	TOK_PLUS: "+", TOK_MINUS: "-", TOK_STAR: "*", TOK_SLASH: "/", TOK_PERCENT: "%",
	TOK_ASSIGN: "=", TOK_DECL_ASSIGN: ":=", TOK_EQ: "==", TOK_NEQ: "!=",
	TOK_LT: "<", TOK_GT: ">", TOK_LTE: "<=", TOK_GTE: ">=",
	TOK_AND: "&&", TOK_OR: "||", TOK_NOT: "!",
	TOK_COMMA: ",", TOK_SEMICOLON: ";", TOK_COLON: ":", TOK_DOT: ".",
	TOK_LPAREN: "(", TOK_RPAREN: ")", TOK_LBRACE: "{", TOK_RBRACE: "}",
	TOK_LBRACKET: "[", TOK_RBRACKET: "]",
	TOK_EOF: "EOF",
}

func (k TokenKind) String() string {
	if s, ok := tokenNames[k]; ok {
		return s
	}
	return fmt.Sprintf("TOK(%d)", int(k))
}

var keywords = map[string]TokenKind{
	"package":  TOK_PACKAGE,
	"import":   TOK_IMPORT,
	"func":     TOK_FUNC,
	"return":   TOK_RETURN,
	"var":      TOK_VAR,
	"if":       TOK_IF,
	"else":     TOK_ELSE,
	"for":      TOK_FOR,
	"break":    TOK_BREAK,
	"continue": TOK_CONTINUE,
	"true":     TOK_TRUE,
	"false":    TOK_FALSE,
	"int":      TOK_TYPE_INT,
	"string":   TOK_TYPE_STRING,
	"bool":     TOK_TYPE_BOOL,
}

// ── Token ─────────────────────────────────────────────────────────────────────

type Token struct {
	Kind  TokenKind
	Value string
	Line  int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Kind, t.Value, t.Line, t.Col)
}

// ── Lexer ─────────────────────────────────────────────────────────────────────

type Lexer struct {
	src    []rune
	pos    int
	line   int
	col    int
	tokens []Token
}

func NewLexer(src string) *Lexer {
	return &Lexer{src: []rune(src), line: 1, col: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekAt(offset int) rune {
	i := l.pos + offset
	if i >= len(l.src) {
		return 0
	}
	return l.src[i]
}

func (l *Lexer) advance() rune {
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) emit(kind TokenKind, val string, line, col int) {
	l.tokens = append(l.tokens, Token{Kind: kind, Value: val, Line: line, Col: col})
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.src) {
		ch := l.peek()
		line, col := l.line, l.col

		// Skip whitespace (but track newlines for auto-semicolons)
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		// Auto-semicolon insertion: Go inserts ';' after certain tokens before '\n'
		if ch == '\n' {
			l.advance()
			if len(l.tokens) > 0 {
				last := l.tokens[len(l.tokens)-1].Kind
				if last == TOK_IDENT || last == TOK_INT || last == TOK_STRING ||
					last == TOK_TRUE || last == TOK_FALSE ||
					last == TOK_RPAREN || last == TOK_RBRACE || last == TOK_RBRACKET ||
					last == TOK_RETURN || last == TOK_BREAK || last == TOK_CONTINUE ||
					last == TOK_TYPE_INT || last == TOK_TYPE_STRING || last == TOK_TYPE_BOOL {
					l.emit(TOK_SEMICOLON, ";", line, col)
				}
			}
			continue
		}

		// Line comments
		if ch == '/' && l.peekAt(1) == '/' {
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
			continue
		}

		// Block comments
		if ch == '/' && l.peekAt(1) == '*' {
			l.advance()
			l.advance()
			for l.pos < len(l.src) {
				if l.peek() == '*' && l.peekAt(1) == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
			continue
		}

		// String literal
		if ch == '"' {
			l.advance()
			var sb strings.Builder
			for l.pos < len(l.src) && l.peek() != '"' {
				c := l.advance()
				if c == '\\' {
					esc := l.advance()
					switch esc {
					case 'n':
						sb.WriteRune('\n')
					case 't':
						sb.WriteRune('\t')
					case '\\':
						sb.WriteRune('\\')
					case '"':
						sb.WriteRune('"')
					default:
						sb.WriteRune('\\')
						sb.WriteRune(esc)
					}
				} else {
					sb.WriteRune(c)
				}
			}
			if l.pos >= len(l.src) {
				return nil, fmt.Errorf("%d:%d: unterminated string", line, col)
			}
			l.advance() // closing "
			l.emit(TOK_STRING, sb.String(), line, col)
			continue
		}

		// Integer literal
		if unicode.IsDigit(ch) {
			var sb strings.Builder
			for l.pos < len(l.src) && unicode.IsDigit(l.peek()) {
				sb.WriteRune(l.advance())
			}
			l.emit(TOK_INT, sb.String(), line, col)
			continue
		}

		// Identifier or keyword
		if unicode.IsLetter(ch) || ch == '_' {
			var sb strings.Builder
			for l.pos < len(l.src) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
				sb.WriteRune(l.advance())
			}
			word := sb.String()
			if kind, ok := keywords[word]; ok {
				l.emit(kind, word, line, col)
			} else {
				l.emit(TOK_IDENT, word, line, col)
			}
			continue
		}

		// Two-character operators
		two := string([]rune{ch, l.peekAt(1)})
		switch two {
		case ":=":
			l.advance()
			l.advance()
			l.emit(TOK_DECL_ASSIGN, ":=", line, col)
			continue
		case "==":
			l.advance()
			l.advance()
			l.emit(TOK_EQ, "==", line, col)
			continue
		case "!=":
			l.advance()
			l.advance()
			l.emit(TOK_NEQ, "!=", line, col)
			continue
		case "<=":
			l.advance()
			l.advance()
			l.emit(TOK_LTE, "<=", line, col)
			continue
		case ">=":
			l.advance()
			l.advance()
			l.emit(TOK_GTE, ">=", line, col)
			continue
		case "&&":
			l.advance()
			l.advance()
			l.emit(TOK_AND, "&&", line, col)
			continue
		case "++":
			l.advance()
			l.advance()
			l.emit(TOK_INCR, "++", line, col)
			continue
		case "--":
			l.advance()
			l.advance()
			l.emit(TOK_DECR, "--", line, col)
			continue
		case "||":
			l.advance()
			l.advance()
			l.emit(TOK_OR, "||", line, col)
			continue
		}

		// Single-character tokens
		l.advance()
		switch ch {
		case '+':
			l.emit(TOK_PLUS, "+", line, col)
		case '-':
			l.emit(TOK_MINUS, "-", line, col)
		case '*':
			l.emit(TOK_STAR, "*", line, col)
		case '/':
			l.emit(TOK_SLASH, "/", line, col)
		case '%':
			l.emit(TOK_PERCENT, "%", line, col)
		case '=':
			l.emit(TOK_ASSIGN, "=", line, col)
		case '<':
			l.emit(TOK_LT, "<", line, col)
		case '>':
			l.emit(TOK_GT, ">", line, col)
		case '!':
			l.emit(TOK_NOT, "!", line, col)
		case ',':
			l.emit(TOK_COMMA, ",", line, col)
		case ';':
			l.emit(TOK_SEMICOLON, ";", line, col)
		case ':':
			l.emit(TOK_COLON, ":", line, col)
		case '.':
			l.emit(TOK_DOT, ".", line, col)
		case '(':
			l.emit(TOK_LPAREN, "(", line, col)
		case ')':
			l.emit(TOK_RPAREN, ")", line, col)
		case '{':
			l.emit(TOK_LBRACE, "{", line, col)
		case '}':
			l.emit(TOK_RBRACE, "}", line, col)
		case '[':
			l.emit(TOK_LBRACKET, "[", line, col)
		case ']':
			l.emit(TOK_RBRACKET, "]", line, col)
		default:
			return nil, fmt.Errorf("%d:%d: unexpected character %q", line, col, ch)
		}
	}
	l.emit(TOK_EOF, "", l.line, l.col)
	return l.tokens, nil
}
