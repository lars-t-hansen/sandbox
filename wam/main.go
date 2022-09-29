package main

import (
	"strings"
)

// The L0 unifier from the book, but using Go's heap

// type argument union {
//   structure
//   variable
// }

type argument interface {
	ToString(*strings.Builder)
}

// Symbols are strings with object identity
type symbol *string

type structure struct {
	functor   symbol
	arguments []argument
}

func (v *structure) ToString(b *strings.Builder) {
	b.WriteString(*v.functor)
	if len(v.arguments) > 0 {
		b.WriteRune('(')
		for i, a := range v.arguments {
			if i > 0 {
				b.WriteRune(',')
			}
			a.ToString(b)
		}
		b.WriteRune(')')
	}
}

type variable struct {
	// Print name.
	name symbol

	// The ID of a variable is used during var-var unification to
	// make the higher-IDd variable forward to the lower-IDd
	// variable.
	id int

	// This is nil for a concrete value; a pointer to a different
	// variable after var-var unification; and a pointer to this
	// variable for an unbound variable.  The last case creates a
	// cycle, and for a refcount-friendly implementation a pointer
	// to a shared "unbound variable" sentinel might be better,
	// even if conceptually more complex.
	uvar *variable

	// This is nil for a variable.
	val *structure
}

func (v *variable) ToString(b *strings.Builder) {
	if v.val != nil {
		v.val.ToString(b)
	} else if v.uvar != v {
		v.uvar.ToString(b)
	} else {
		b.WriteString(*v.name)
	}
}

type context struct {
	symbols map[string]symbol
	varId   int
}

func newContext() *context {
	return &context{
		symbols: make(map[string]symbol),
		varId:   0,
	}
}
func (c *context) intern(s string) symbol {
	if v, ok := c.symbols[s]; ok {
		return v
	}
	v := &s
	c.symbols[s] = v
	return v
}

func (c *context) newVar(name string) *variable {
	sym := c.intern(name)
	id := c.varId
	c.varId++
	v := &variable{name: sym, id: id}
	v.uvar = v
	return v
}

func (c *context) newStruct(name string, args ...argument) *structure {
	return &structure{functor: c.intern(name), arguments: args}
}

func (c *context) newConst(name string) *structure {
	return c.newStruct(name)
}

func isVar(x argument) bool {
	// if x is a variable then
	//  resolve x to its canonical representation
	//    -- this means that it either points to the canonical variable
	//    -- slot or receives a value
	//  if x is a variable then
	//   true
	// false
}

func (c *context) uninfy(lhs argument, rhs argument) {
	if isVar(lhs) {
		if isVar(rhs) {
			unifyVarVar(lhs, rhs)
		} else {
			unifyVarVal(lhs, rhs)
		}
	} else if isVar(rhs) {
		unifyVarVal(rhs, lhs)
	} else {
		unifyValVal(lhs, rhs)
	}
}

func main() {
	c := newContext()
	A := c.newVar("A")
	B := c.newVar("B")
	f := c.newStruct("f", A, B, c.newConst("true"))
	g := c.newStruct("f", c.newConst("false"), c.newVar("D"), c.newConst("true"))
	c.unify(f, g)
}
