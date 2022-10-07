//////////////////////////////////////////////////////////////////////////
//
// Tokenizer

package compiler

/*
import (
	"fmt"
	"io"
)

type tkind int

const (
	t_invalid tkind = iota
	t_atom
	t_varname
	t_infix
	t_number
	t_lparen
	t_rparen
	t_comma
	t_period
	t_eof
)

type token struct {
	kind   tkind
	lineno int
	name   string
}

type reader interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

type tokenizer struct {
	input  reader
	next   token
	lineno int
}

func newTokenizer(input reader) *tokenizer {
	return &tokenizer{
		input:  input,
		lineno: 1}
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

func (t *tokenizer) peek() token {
	if t.next.kind == t_invalid {
		t.next = t.readNext()
	}
	return t.next
}

func (t *tokenizer) get() token {
	if t.next.kind != t_invalid {
		tok := t.next
		t.next.kind = t_invalid
		return tok
	}
	return t.readNext()
}

func (t *tokenizer) readNext() token {
outer:
	for {
		r := t.getChar()
		if r == -1 {
			return token{t_eof, t.lineno, ""}
		}
		if r == '\t' || r == ' ' {
			continue
		}
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
			return token{t_lparen, t.lineno, ""}
		}
		if r == ')' {
			return token{t_rparen, t.lineno, ""}
		}
		if r == '.' {
			return token{t_period, t.lineno, ""}
		}
		if r == ',' {
			return token{t_comma, t.lineno, ""}
		}
		if r == '-' {
			if isDigitChar(t.peekChar()) {
				return token{t_number, t.lineno, t.lexWhile(isDigitChar, "-")}
			}
		}
		if r == '\'' {
			name := t.lexWhile(func(r rune) bool {
				return r != -1 && r != '\'' && r != '\n' && r != '\r'
			}, "")
			if t.getChar() != '\'' {
				panic(fmt.Sprintf("Line %d: unterminated quoted atom", t.lineno))
			}
			return token{t_atom, t.lineno, name}
		}
		if isOperatorChar(r) {
			return token{t_infix, t.lineno, t.lexWhile(isOperatorChar, string(r))}
		}
		if isDigitChar(r) {
			return token{t_number, t.lineno, t.lexWhile(isDigitChar, string(r))}
		}
		if isVarFirstChar(r) {
			return token{t_varname, t.lineno, t.lexWhile(isAtomNextChar, string(r))}
		}
		if isAtomFirstChar(r) {
			// TODO: This strikes me as a hack, there should be a more principled solution
			// to this somewhere.
			name := t.lexWhile(isAtomNextChar, string(r))
			if name == "is" {
				return token{t_infix, t.lineno, name}
			}
			return token{t_atom, t.lineno, name}
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
*/
