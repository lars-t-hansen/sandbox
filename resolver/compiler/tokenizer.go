package compiler

import (
	"fmt"
	"io"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//
// Tokenizer

type reader interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

type tokenizer struct {
	// Input characters
	input reader

	// Line number at start of next character in the input.  Private to lexer,
	// tokens and nonterminals carry their own line numbers.
	lineno int

	// The rest of this is parser context, see parser code above.
	ctx *parserctx
}

func newTokenizer(r reader, ctx *parserctx) *tokenizer {
	return &tokenizer{
		input:  r,
		lineno: 0,
		ctx:    ctx,
	}
}

func (l *tokenizer) Lex(lval *yySymType) (t int) {
	t, lval.name.text, lval.name.line = l.get()
	return
}

func (l *tokenizer) Error(s string) {
	// TODO: Line number for the error, although sometimes that comes in with the
	// message too?
	panic(fmt.Sprintf("Line %d: %s", l.lineno, s))
}

func (t *tokenizer) peekChar() rune {
	r, _, err := t.input.ReadRune()
	if err == io.EOF {
		return -1
	}
	if err != nil {
		panic(fmt.Sprintf("Line %d: Bad input: "+err.Error(), t.lineno))
	}
	t.input.UnreadRune()
	return r
}

func (t *tokenizer) getChar() rune {
	r, _, err := t.input.ReadRune()
	if err == io.EOF {
		return -1
	}
	if err != nil {
		panic(fmt.Sprintf("Line %d: Bad input: "+err.Error(), t.lineno))
	}
	return r
}

func (t *tokenizer) get() (tokval int, name string, lineno int) {
outer:
	for {
		r := t.getChar()
		if r == -1 {
			tokval, lineno = -1, t.lineno
			return
		}
		if r == '\t' || r == ' ' {
			continue
		}
		// TODO: \r and other line breaks
		if r == '\n' {
			t.lineno++
			continue
		}
		if r == '/' && t.peekChar() == '*' {
			t.getChar()
			for {
				r := t.getChar()
				if r == -1 {
					panic(fmt.Sprintf("Line %d: EOF in comment", t.lineno))
				}
				if r == '*' && t.peekChar() == '/' {
					t.getChar()
					continue outer
				}
				if r == '\n' {
					t.lineno++
				}
			}
		}
		if r == '(' {
			tokval, lineno = T_LPAREN, t.lineno
			return
			return
		}
		if r == ')' {
			tokval, lineno = T_RPAREN, t.lineno
			return
		}
		if r == '.' {
			tokval, lineno = T_PERIOD, t.lineno
			return
		}
		if r == ',' {
			tokval, lineno = T_COMMA, t.lineno
			return
		}
		if r == '-' {
			lineno = t.lineno
			if isDigitChar(t.peekChar()) {
				name = t.lexWhile(isDigitChar, "-")
				tokval = T_NUMBER
				return
			}
		}
		if r == '\'' {
			lineno = t.lineno
			name = t.lexWhile(func(r rune) bool {
				return r != -1 && r != '\'' && r != '\n' && r != '\r'
			}, "")
			if t.getChar() != '\'' {
				panic(fmt.Sprintf("Line %d: unterminated quoted atom", t.lineno))
			}
			tokval = T_ATOM
			return
		}
		if isOperatorChar(r) {
			lineno = t.lineno
			name = t.lexWhile(isOperatorChar, string(r))
			if name == "?-" {
				tokval = T_QUERY_OP
				return
			}
			if name == ":-" {
				tokval = T_FACT_OP
				return
			}
			tokval = T_INFIX_OP
			return
		}
		if isDigitChar(r) {
			lineno = t.lineno
			name = t.lexWhile(isDigitChar, string(r))
			tokval = T_NUMBER
			return
		}
		if isVarFirstChar(r) {
			lineno = t.lineno
			name = t.lexWhile(isAtomNextChar, string(r))
			tokval = T_VARNAME
			return
		}
		if isAtomFirstChar(r) {
			lineno = t.lineno
			name = t.lexWhile(isAtomNextChar, string(r))
			// TODO: This strikes me as a hack, there should be a more principled solution
			// to this somewhere.
			if name == "is" {
				tokval = T_INFIX_OP
			} else {
				tokval = T_ATOM
			}
			return
		}
		panic(fmt.Sprintf("Line %d: bad character: %v", t.lineno, r))
	}
}

// This depends on isChar() being false for -1 and newlines
func (t *tokenizer) lexWhile(isChar func(r rune) bool, s string) string {
	for isChar(t.peekChar()) {
		s = s + string(t.getChar())
	}
	return s
}

func isOperatorChar(r rune) bool {
	return r == '+' || r == '-' || r == '?' || r == ':' || r == '!' || r == '='
}

func isDigitChar(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAtomFirstChar(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isVarFirstChar(r rune) bool {
	return r >= 'A' && r <= 'Z' || r == '_'
}

func isAtomNextChar(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || r >= '0' && r <= '9'
}
