// The L0 unifier from the book, but using Go's heap

package main

import (
	"os"
	"strings"
)

// Symbols are just strings with object identity

type symbol *string

// Global state for evaluation

type store struct {
	symbols map[string]symbol
	varId   int
	vars    []*variable
}

func newStore() *store {
	return &store{
		symbols: make(map[string]symbol),
		varId:   0,
		vars:    []*variable{},
	}
}

func (st *store) toString(b *strings.Builder) {
	for _, v := range st.vars {
		b.WriteString(*v.name)
		b.WriteString(" = ")
		v.toString(b)
		b.WriteRune('\n')
	}
}

func (st *store) intern(name string) symbol {
	if v, ok := st.symbols[name]; ok {
		return v
	}
	v := &name
	st.symbols[name] = v
	return v
}

func (st *store) newVarId() int {
	n := st.varId
	st.varId++
	return n
}

// Terms are structures or variables.  A zero-arity structure is also known
// as a constant.
//
// type term union {
//   structure
//   variable
// }

type term interface {
	toString(*strings.Builder)
}

type structure struct {
	functor  symbol
	subterms []term
}

func (st *store) newStruct(name string, args ...term) *structure {
	return &structure{functor: st.intern(name), subterms: args}
}

func (st *store) newConst(name string) *structure {
	return st.newStruct(name)
}

func (v *structure) toString(b *strings.Builder) {
	b.WriteString(*v.functor)
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
	name symbol

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
	val *structure
}

func (st *store) newVar(name string) *variable {
	v := &variable{name: st.intern(name), id: st.newVarId()}
	v.uvar = v
	st.vars = append(st.vars, v)
	return v
}

func (v *variable) toString(b *strings.Builder) {
	if v.val != nil {
		v.val.toString(b)
	} else if v.uvar != v {
		v.uvar.toString(b)
	} else {
		b.WriteString(*v.name)
	}
}

func (v *variable) resolve() term {
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

func unify(lhs term, rhs term) bool {
	if v1, ok := lhs.(*variable); ok {
		lhs = v1.resolve()
	}
	if v2, ok := rhs.(*variable); ok {
		rhs = v2.resolve()
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
		v1.val = rhs.(*structure)
		return true
	}
	if v2, ok := rhs.(*variable); ok {
		v2.uvar = nil
		v2.val = lhs.(*structure)
		return true
	}
	s1 := lhs.(*structure)
	s2 := rhs.(*structure)
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

func main() {
	var buf strings.Builder
	c := newStore()
	A := c.newVar("A")
	B := c.newVar("B")
	if !unify(A, c.newStruct("f", c.newConst("false"), B)) {
		panic("First test")
	}
	if !unify(B, c.newConst("true")) {
		panic("Second test")
	}
	D := c.newVar("D")
	if !unify(A, c.newStruct("f", D, c.newConst("true"))) {
		panic("Third test")
	}
	E := c.newVar("E")
	F := c.newVar("F")
	G := c.newVar("G")
	H := c.newVar("H")
	if !unify(c.newStruct("f", E, F), c.newStruct("f", G, H)) {
		panic("Fourth test")
	}
	if !unify(c.newStruct("f", c.newConst("hi"), c.newConst("ho")), E) {
		panic("Fifth test")
	}
	c.toString(&buf)
	os.Stdout.WriteString(buf.String())
}
