package compiler

import (
	"fmt"
	"io"
)

//////////////////////////////////////////////////////////////////////////
//
// Tokenizer
//

type tkind int

const (
	t_atom tkind = iota
	t_operator
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
	lineno int
}

func newTokenizer(input reader) *tokenizer {
	return &tokenizer{input: input, lineno: 1}
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

func (t *tokenizer) next() token {
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
		if isOperatorChar(r) {
			return token{t_operator, t.lineno, t.lexWhile(isOperatorChar, string(r))}
		}
		if isDigitChar(r) {
			return token{t_number, t.lineno, t.lexWhile(isDigitChar, string(r))}
		}
		if isAtomFirstChar(r) {
			return token{t_atom, t.lineno, t.lexWhile(isAtomNextChar, string(r))}
		}
		panic(fmt.Sprintf("Line %d: bad character: %v", t.lineno, r))
	}
}

// This depends on isChar() being false for -1
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
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_'
}

func isAtomNextChar(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || r >= '0' && r <= '9'
}

//////////////////////////////////////////////////////////////////////////
//
// Parser

type astNode interface {
	fmt.Stringer
}

type astVar struct {
	name string
}

func (n *astVar) String() string {
	return n.name
}

type astAtom struct {
	name string
}

func (n *astAtom) String() string {
	return n.name
}

type astNumber struct {
	value int64
}

func (n *astNumber) String() string {
	return fmt.Sprint(n.value)
}

type astStruct struct {
	name       string
	components []astNode
}

func (n *astStruct) String() string {
	args := ""
	for _, c := range n.components {
		if args != "" {
			args = args + ","
		}
		args = args + c.String()
	}
	return n.name + "(" + args + ")"
}

type astQuery struct {
	// list of vars
	// body - list of terms, probably all structs?
}

type astFact struct {
	// this has no vars
	// has only a head
}

type astRule struct {
	// list of vars
	// head
	// body - list of terms, mostly structs, cut, fail?
}

// f(G) :- h(G,A,"x"), j(A)
// is compiled into
//
//	rule{locals: 2,
//	     head: &structure{"f",[]term{&local{0}},
//	     body: []*structure{&structure{"h", []term{&local{0},&local{1},intern("x")}},
//	                        &structure{"j", []term{&local{1}}}}
//
// The input to the compiler is a head structure and a slice of subterms, also structures.
// These structures are fully ground - this is source code!  There is a
/*
func compileRule(head *astStruct, subterms []*astNode) *rule {
	//rib := make(map[string]int)

	panic("NYI")
}
*/
