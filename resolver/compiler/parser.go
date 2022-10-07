//////////////////////////////////////////////////////////////////////////
//
// Parser
//
// The grammar here is something like this
//
// phrase ::= term ":-" terms "."
//          | ":-" term "."
//          | "?-" terms "."
// terms ::= term ("," term)*
// term ::= atom | number | varname | struct | "!"
// struct ::= atom "(" terms ")"
//          | term binop term
// binop ::= <sundry operators> | "is"
//
// with additional constraints on the terms that may appear in various places.
//
// Real prolog in addition has syntactic sugar for lists, NYI here.

package compiler

/*
import (
	"fmt"
)

type parser struct {
	toks      *tokenizer
	nameMap   map[string]int // Indices of names that are not "_"
	names     []*astVar      // All unique variables including "_"s
	nextIndex int
}

func newParser(t *tokenizer) *parser {
	return &parser{
		toks:      t,
		nameMap:   make(map[string]int),
		names:     []*astVar{},
		nextIndex: 0,
	}
}

func (p *parser) clear() {
	for k := range p.nameMap {
		delete(p.nameMap, k)
	}
	p.names = []*astVar{}
	p.nextIndex = 0
}

func (p *parser) peek(kind tkind) bool {
	return p.toks.peek().kind == kind
}

func (p *parser) get() token {
	return p.toks.get()
}

func (p *parser) eat(kind tkind, name string) bool {
	t := p.toks.peek()
	if t.kind == kind && t.name == name {
		p.get()
		return true
	}
	return false
}

func (p *parser) match(kind tkind, name string) {
	t := p.toks.peek()
	if t.kind == kind && t.name == name {
		p.get()
	} else {
		panic(fmt.Sprintf("Line %d: failed to match token", t.lineno))
	}
}

func (p *parser) eatX(kind tkind) (bool, string) {
	if p.peek(kind) {
		t := p.get()
		return true, t.name
	}
	return false, ""
}
*/
// A new astVar is created for each occurrence of a given variable, but the index
// is shared among all the instances of astVar for a given name (except "_").
//
// In the parser, the `names` list has the first instance of an astVar for a given
// index value.
/*
type astVar struct {
	lineno int
	name   string
	index  int
}

func (n *astVar) String() string {
	return fmt.Sprintf("[%s %d]", n.name, n.index)
}

func (n *astVar) line() int {
	return n.lineno
}

type astAtom struct {
	lineno int
	name   string
}

func (n *astAtom) String() string {
	return n.name
}

func (n *astAtom) line() int {
	return n.lineno
}

type astNumber struct {
	lineno int
	value  int64
}

func (n *astNumber) String() string {
	return fmt.Sprint(n.value)
}

func (n *astNumber) line() int {
	return n.lineno
}

type astStruct struct {
	lineno     int
	name       string
	components []astTerm
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

func (n *astStruct) line() int {
	return n.lineno
}
*/
// type astTerm union {
//   *astAtom
//   *astNumber
//   *astStruct
//   *astVar
// }
/*
type astTerm interface {
	fmt.Stringer
	line() int
}
func (p *parser) parseTerm0() astTerm {
	t := p.get()
	switch t.kind {
	case t_number:
		val, err := strconv.ParseInt(t.name, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Line %d: Number out of range: %s", t.lineno, t.name))
		}
		return &astNumber{t.lineno, val}
	case t_atom:
		if p.peek(t_lparen) {
			p.get()
			ts := p.parseTerms()
			p.match(t_rparen, "")
			return &astStruct{t.lineno, t.name, ts}
		}
		return &astAtom{t.lineno, t.name}
	case t_varname:
		if t.name == "_" {
			// Fresh anonymous variable
			index := p.nextIndex
			p.nextIndex++
			v := &astVar{t.lineno, t.name, index}
			p.names = append(p.names, v)
			return v
		}
		index, found := p.nameMap[t.name]
		if found {
			// Previously seen variable
			return &astVar{t.lineno, t.name, index}
		}
		// Fresh named variable
		index = p.nextIndex
		p.nextIndex++
		p.nameMap[t.name] = index
		v := &astVar{t.lineno, t.name, index}
		p.names = append(p.names, v)
		return v
	default:
		panic(fmt.Sprintf("Line %d: Unexpected token", t.lineno))
	}
}

func (p *parser) parseTerm() astTerm {
	term := p.parseTerm0()
	for p.peek(t_infix) {
		op := p.get()
		rhs := p.parseTerm0()
		term = &astStruct{term.line(), op.name, []astTerm{term, rhs}}
	}
	return term
}

func (p *parser) parseTerms() []astTerm {
	ts := []astTerm{}
	ts = append(ts, p.parseTerm())
	for p.peek(t_comma) {
		p.get()
		ts = append(ts, p.parseTerm())
	}
	return ts
}

/*
	func termsToString(ts []astTerm) string {
		s := ""
		for _, term := range ts {
			if s != "" {
				s = s + ", "
			}
			s = s + term.String()
		}
		return s
	}

	type astQuery struct {
		vars []*astVar
		body []astTerm
	}

	func (f *astQuery) String() string {
		return "?- " + termsToString(f.body) + "."
	}

	func (f *astQuery) line() int {
		return f.body[0].line()
	}

	type astFact struct {
		head astTerm
	}

	func (f *astFact) String() string {
		return ":-" + f.head.String() + "."
	}

	func (f *astFact) line() int {
		return f.head.line()
	}

	type astRule struct {
		vars []*astVar
		head astTerm
		body []astTerm
	}

	func (f *astRule) String() string {
		return f.head.String() + " :- " + termsToString(f.body) + "."
	}

	func (f *astRule) line() int {
		return f.head.line()
	}

// type astPhrase union {
//   *astFact
//   *astRule
//   *astQuery
//   nil
// }

	type astPhrase interface {
		fmt.Stringer
		line() int
	}

	func (p *parser) parsePhrase() astPhrase {
		p.clear()
		if p.eat(t_eof, "") {
			return nil
		}
		if p.eat(t_infix, ":-") {
			head := p.parseTerm0()
			p.match(t_period, "")
			if len(p.names) != 0 {
				panic(fmt.Sprintf("Line %d: Fact should not have variables: %v", head.line(), head))
			}
			return &astFact{head}
		}
		if p.eat(t_infix, "?-") {
			query := p.parseTerms()
			p.match(t_period, "")
			return &astQuery{p.names, query}
		}
		head := p.parseTerm0()
		p.match(t_infix, ":-")
		body := p.parseTerms()
		p.match(t_period, "")
		return &astRule{p.names, head, body}
	}
func (p *parser) parsePhrase() astPhrase {
	panic("Hi")
}
*/
