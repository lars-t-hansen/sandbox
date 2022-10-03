// term ::= structure | variable
// structure ::= functor term*
// functor ::= symbol
// rule ::= head body?
// head ::= structure
// body ::= structure+
// query ::= structure
// ruleset ::= rule*

// To apply a query Q = (f t_1 .. t_n) to a ruleset is to
//  - select from the ruleset the rules R whose functor and arity match f and n
//  - if R is empty then fail
//  - for each rule S in R in order
//    - create a copy C of S with fresh variables where S has variables
//    - try to unify the head of C with Q
//    - if unification succeeds,
//       - for each structure B in the body of S in order,
//         - apply the query B to the ruleset
//       - if any application fails, then fail, otherwise succeed.

package main

import (
	"fmt"
	"os"
	"strings"
)

// Atoms are just strings with object identity

type atom struct {
	name string
}

func (a *atom) toString(b *strings.Builder) {
	b.WriteString(a.name)
}

// Numbers are numbers, for now just i64

type number struct {
	value int64
}

func (a *number) toString(b *strings.Builder) {
	b.WriteString(fmt.Sprint(a.value))
}

// Global state for evaluation

type store struct {
	// Interned constants.
	atoms map[string]*atom

	// This is probably redundant.
	varId int

	// Database of rules.  This is indexed by functor and arity of the head.
	rules map[*atom]map[int][]*rule
}

func newStore() *store {
	return &store{
		atoms: make(map[string]*atom),
		varId: 0,
		rules: make(map[*atom]map[int][]*rule),
	}
}

func (st *store) intern(name string) *atom {
	if v, ok := st.atoms[name]; ok {
		return v
	}
	v := &atom{name: name}
	st.atoms[name] = v
	return v
}

func (st *store) newVarId() int {
	n := st.varId
	st.varId++
	return n
}

func (st *store) assert(r *rule) {
	functorMap, ok := st.rules[r.head.functor]
	if !ok {
		functorMap = make(map[int][]*rule)
		st.rules[r.head.functor] = functorMap
	}
	arity := len(r.head.subterms)
	aritySlice, ok := functorMap[arity]
	if !ok {
		aritySlice = make([]*rule, 0, 4)
	}
	functorMap[arity] = append(aritySlice, r)
}

func (st *store) lookup(functor *atom, arity int) []*rule {
	functorMap, ok := st.rules[functor]
	if !ok {
		return []*rule{}
	}
	aritySlice, ok := functorMap[arity]
	if !ok {
		return []*rule{}
	}
	return aritySlice
}

func (st *store) retract(functor atom, arity int) {
	panic("NYI")
}

// type term union {
//   *structure
//   *variable
//   *atom
//   *number
// }

type term interface {
	toString(*strings.Builder)
}

type structure struct {
	functor  *atom
	subterms []term
}

func (st *store) newStruct(name string, args ...term) *structure {
	return &structure{functor: st.intern(name), subterms: args}
}

func (st *store) atom(name string) *atom {
	return st.intern(name)
}

func (v *structure) toString(b *strings.Builder) {
	v.functor.toString(b)
	if len(v.subterms) > 0 {
		b.WriteRune('(')
		for i, a := range v.subterms {
			if i > 0 {
				b.WriteRune(',')
			}
			a.toString(b)
		}
		b.WriteRune(')')
	}
}

type variable struct {
	// Print name.
	name *atom

	// The `id` of a variable is used during var-var unification to
	// make the higher-IDd variable forward to the lower-IDd
	// variable.
	id int

	// Precisely one of `uvar` and `val` is nil.

	// `uvar`` is:
	//
	// - nil for a concrete value
	// - a pointer to a different variable after var-var unification
	//   where this variable is not the canonical variable
	// - and a pointer to this variable for a canonical unbound variable.
	//
	// The last case creates a cycle, and for a refcount-friendly implementation
	// a pointer to a shared "unbound variable" sentinel might be better,
	// even if conceptually more complex.
	uvar *variable

	// `val` is:
	//
	// - non-nil for a concrete value
	// - nil in every other case
	val term
}

func (st *store) newVar(name string) *variable {
	v := &variable{name: st.intern(name), id: st.newVarId()}
	v.uvar = v
	return v
}

func (v *variable) toString(b *strings.Builder) {
	if v.val != nil {
		v.val.toString(b)
	} else if v.uvar != v {
		v.uvar.toString(b)
	} else {
		v.name.toString(b)
	}
}

func (v *variable) resolveVar() term {
	for v.uvar != nil && v.uvar != v {
		next := v.uvar
		if next.uvar == nil {
			v.uvar = nil
			v.val = next.val
		} else {
			v.uvar = next.uvar
		}
	}
	if v.val != nil {
		return v.val
	}
	return v
}

type query struct {
	st *store

	// These are variables that are free in the query
	vars []*variable
}

func newQuery(st *store) *query {
	return &query{st: st, vars: []*variable{}}
}

func (q *query) toString(b *strings.Builder) {
	for _, v := range q.vars {
		v.name.toString(b)
		b.WriteString(" = ")
		v.toString(b)
		b.WriteRune('\n')
	}
}

func (q *query) newVar(name string) *variable {
	v := &variable{name: q.st.intern(name), id: q.st.newVarId()}
	v.uvar = v
	q.vars = append(q.vars, v)
	return v
}

type rule struct {
	head *structure
	body []*structure
}

func newFact(head *structure) *rule {
	return &rule{head: head, body: []*structure{}}
}

func unify(lhs term, rhs term) bool {
	if v1, ok := lhs.(*variable); ok {
		lhs = v1.resolveVar()
	}
	if v2, ok := rhs.(*variable); ok {
		rhs = v2.resolveVar()
	}
	if v1, ok := lhs.(*variable); ok {
		if v2, ok := rhs.(*variable); ok {
			if v1.id < v2.id {
				v2.uvar = v1
			} else if v2.id < v1.id {
				v1.uvar = v2
			}
			return true
		}
		v1.uvar = nil
		v1.val = rhs
		return true
	}
	if v2, ok := rhs.(*variable); ok {
		v2.uvar = nil
		v2.val = lhs
		return true
	}
	if s1, ok := lhs.(*structure); ok {
		if s2, ok := rhs.(*structure); ok {
			if s1.functor != s2.functor || len(s1.subterms) != len(s2.subterms) {
				return false
			}
			for i := 0; i < len(s1.subterms); i++ {
				if !unify(s1.subterms[i], s2.subterms[i]) {
					return false
				}
			}
			return true
		}
		return false
	}
	if a1, ok := lhs.(*atom); ok {
		if a2, ok := rhs.(*atom); ok {
			return a1 == a2
		}
		return false
	}
	if n1, ok := lhs.(*number); ok {
		if n2, ok := rhs.(*number); ok {
			return n1.value == n2.value
		}
		return false
	}
	// There will be a case for numbers too
	return false
}

func (q *query) ask(s *structure) bool {
	rules := q.st.lookup(s.functor, len(s.subterms))
	for _, r := range rules {
		if unify(r.head, s) {
			return true
		}
		// Unbind any query variables bound by the unification
		for _, v := range q.vars {
			v.val = nil
			v.uvar = v
		}
	}
	return false
}

func main() {
	var buf strings.Builder
	st := newStore()
	st.assert(newFact(st.newStruct("father", st.atom("haakon"), st.atom("olav"))))
	st.assert(newFact(st.newStruct("father", st.atom("olav"), st.atom("harald"))))
	st.assert(newFact(st.newStruct("father", st.atom("harald"), st.atom("håkon magnus"))))
	st.assert(newFact(st.newStruct("father", st.atom("håkon magnus"), st.atom("ingrid alexandra"))))
	q := newQuery(st)
	X := q.newVar("X")
	found := q.ask(st.newStruct("father", X, st.atom("harald")))
	if !found {
		os.Stdout.WriteString("no\n")
	} else if len(q.vars) == 0 {
		os.Stdout.WriteString("yes\n")
	} else {
		q.toString(&buf)
		os.Stdout.WriteString(buf.String())
	}
}
