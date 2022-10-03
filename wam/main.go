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

// Global state for evaluation

type store struct {
	// Interned constants.
	atoms map[string]*atom

	// Database of rules.  This is indexed by functor and arity of the head.
	rules map[*atom]map[int][]*rule
}

func newStore() *store {
	return &store{
		atoms: make(map[string]*atom),
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

// Locals are indices into a rib of variables for the current rule

type local struct {
	slot int
}

type env []variable

func (a *local) toString(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("V%d", a.slot))
}

func (a *local) resolve(e env) any {
	v := &e[a.slot]
	return v.resolveVar()
}

// Rough plan
//
// The meaning of "variable" becomes "local variable" which holds the index into a rib.
// Every rule f(A) :- g(B), h(C) is evaluated in a rib holding A, B, C
// After var-var unification these slots can forward to other variables
// A variable representation is therefor (rib,index) where rib is just a rib object
// The variable slot also holds a possible value, obviously
// GC ensures that the ribs are kept alive
//
// A query is then a term + a rib holding the variables that are free in the query,
// and if the query succeeds we print those vars.
//
// Evaluation is pretty much CPS because this allows failure to be encoded simply: the
// eventual continuation prints the variables of the query, but when made to fail we just
// backtrack into the recursion.

// type term union {
//   *structure
//   *local
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

// The name does not need to be here, it can be stored externally in a R/O structure,
// and basically just for queries - normal ribs don't need it at all, except for
// debugging, and in that case it can be reconstructed from the rule.

// It would be incredibly sweet for this to be just one pointer.  But this just introduces
// complexity.  For now, we want the zero value to be meaningful.  So perhaps define
// that the variable is unbound or forwarding if val is nil; it is unbound and the end
// of a chain if uvar is nil too.
//
// Ergo the test always starts with testing val; if it is nil, this is a variable to be
// resolved further.
//
// With this fix, a zero variable is a meaningful variable and needs no initializer,
// which means make([]variable, n) is totally fine.
//
// At the same time, I said that a term would be represented using indices into a rib for
// the variables...  not pointers...
//
// So execution resolves from an index into a *variable and after that it's *variable all
// the way down.  The index is just so that we don't need to reconstruct a term every time
// we enter a rule, the term representation is immutable.

type variable struct {
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

/*
func (st *store) newVar(name string) *variable {
	v := &variable{name: st.intern(name)}
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
*/

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
	v := &variable{name: q.st.intern(name)}
	v.uvar = v
	q.vars = append(q.vars, v)
	return v
}

type rule struct {
	locals int
	head   *structure
	body   []*structure
}

func newFact(head *structure) *rule {
	return &rule{head: head, body: []*structure{}}
}

func unify(e env, lhs term, rhs term) bool {
	var t1, t2 any
	if v1, ok := lhs.(*local); ok {
		t1 = v1.resolve(e)
	} else {
		t1 = lhs
	}
	if v2, ok := rhs.(*local); ok {
		t2 = v2.resolve(e)
	} else {
		t2 = rhs
	}
	if v1, ok := t1.(*variable); ok {
		if v2, ok := t2.(*variable); ok {
			// Arbitrarily make the second point to the first
			v2.uvar = v1
			return true
		}
		v1.uvar = nil
		v1.val = rhs
		return true
	}
	if v2, ok := t2.(*variable); ok {
		v2.uvar = nil
		v2.val = lhs
		return true
	}
	if s1, ok := t1.(*structure); ok {
		if s2, ok := t2.(*structure); ok {
			if s1.functor != s2.functor || len(s1.subterms) != len(s2.subterms) {
				return false
			}
			for i := 0; i < len(s1.subterms); i++ {
				if !unify(e, s1.subterms[i], s2.subterms[i]) {
					return false
				}
			}
			return true
		}
		return false
	}
	if a1, ok := t1.(*atom); ok {
		if a2, ok := t2.(*atom); ok {
			return a1 == a2
		}
		return false
	}
	if n1, ok := t1.(*number); ok {
		if n2, ok := t2.(*number); ok {
			return n1.value == n2.value
		}
		return false
	}
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
	// a query is just a rib whose lifetime is controlled
	// q.newVar creates a new local but also somehow records it so that we can print them out
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
