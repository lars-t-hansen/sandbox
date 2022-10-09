package repl

import (
	"fmt"
	"io"
)

type tokenizer struct {
	input  reader
	lineno int
	ctx    *parserctx
}

func newTokenizer(r reader, ctx *parserctx) *tokenizer {
	return &tokenizer{
		input:  r,
		lineno: 0,
		ctx:    ctx,
	}
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

func (t *tokenizer) get() (tokval int, name string) {
outer:
	for {
		r := t.getChar()
		if r == -1 {
			tokval = -1
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
			tokval = T_LPAREN
			return
		}
		if r == ')' {
			tokval = T_RPAREN
			return
		}
		if r == '.' {
			tokval = T_PERIOD
			return
		}
		if r == ',' {
			tokval = T_COMMA
			return
		}
		if r == '-' {
			if isDigitChar(t.peekChar()) {
				name = t.lexWhile(isDigitChar, "-")
				tokval = T_NUMBER
				return
			}
		}
		if r == '\'' {
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
			name = t.lexWhile(isDigitChar, string(r))
			tokval = T_NUMBER
			return
		}
		if isVarFirstChar(r) {
			name = t.lexWhile(isAtomNextChar, string(r))
			tokval = T_VARNAME
			return
		}
		if isAtomFirstChar(r) {
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
	// TODO: inadequate
	return r == '+' || r == '-' || r == '?' || r == ':' || r == '!' || r == '='
}

func isDigitChar(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAtomFirstChar(r rune) bool {
	// TODO: any unicode "lower case" letter should be accepted.  Letters outside the
	// ascii range have to be quoted now.
	return r >= 'a' && r <= 'z'
}

func isVarFirstChar(r rune) bool {
	// TODO: any unicode "upper case" letter should be accepted
	return r >= 'A' && r <= 'Z' || r == '_'
}

func isAtomNextChar(r rune) bool {
	// TODO: any unicode letter should be accepted.  Letters outside the ascii range have to
	// be quoted now.
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || r >= '0' && r <= '9'
}
