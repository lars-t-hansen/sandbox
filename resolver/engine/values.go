package engine

import (
	"fmt"
	"strings"
)

// `Store`: Global background state for evaluation

type Store struct {
	// Interned atoms.
	atoms map[string]*Atom

	// Database of rules.  This is indexed by the functor and arity of the head.
	rules map[*Atom]map[int][]*rule
}

func NewStore() *Store {
	return &Store{
		atoms: make(map[string]*Atom),
		rules: make(map[*Atom]map[int][]*rule),
	}
}
func (st *Store) addRule(r *rule) {
	functorMap, ok := st.rules[r.functor]
	if !ok {
		functorMap = make(map[int][]*rule)
		st.rules[r.functor] = functorMap
	}
	aritySlice, ok := functorMap[r.arity]
	if !ok {
		aritySlice = make([]*rule, 0, 4)
	}
	functorMap[r.arity] = append(aritySlice, r)
}

func (st *Store) lookupRule(functor *Atom, arity int) []*rule {
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

// Values: A term value has two flavors, unbound (the `RuleTerm`) and bound (the `ValueTerm`),
// reflecting that when a predicate is entered its head and body must be bound to the new rib
// for the predicate.  The terms of the predicate do not contain varslot nodes but instead Local
// nodes that point into the rib; the binding operation creates a pair that associates the term
// with a rib.
//
// The `ruleTermTag` and `valueTermTag` methods on RuleTerm and ValueTerm serve to distinguish
// the types and make sure that unbound terms do not flow into resolution code.  The atomic
// types Atom and Number are both RuleTerms and ValueTerms, but Locals are only RuleTerms and
// for structures there are two kinds, RuleStruct and ValueStruct.  Binding operations transform
// a value from one domain to another.

type RuleTerm interface {
	fmt.Stringer
	ruleTermTag() string
}

type ValueTerm interface {
	fmt.Stringer
	valueTermTag() string
}

type rib []Varslot

func bind(t RuleTerm, e rib) ValueTerm {
	assert(t != nil)
	switch x := t.(type) {
	case *RuleStruct:
		return &ValueStruct{env: e, s: x}
	case *Atom:
		return x
	case *Number:
		return x
	case *Local:
		return &e[x.slot]
	default:
		panic("NYI")
	}
}

func bind_terms(ts []RuleTerm, e rib) []ValueTerm {
	vs := make([]ValueTerm, len(ts))
	for i, t := range ts {
		vs[i] = bind(t, e)
	}
	return vs
}

// `varslot`: evaluator-internal storage for variables.  They are allocated inside ribs,
// which are themselves allocated when predicates are evaluated.
//
// If `val` is not nil then it is the value held in this slot.  Otherwise, `next` is either nil,
// in which case this is the canonical varslot for a variable, or it points to another varslot
// that this varslot has been unified with.

type Varslot struct {
	next *Varslot
	val  ValueTerm
}

func (v *Varslot) String() string {
	assert(v != nil)
	if v.val != nil {
		return "[value " + v.val.String() + "]"
	}
	return "[varslot]"
}

func (v *Varslot) valueTermTag() string {
	return "[varslot]"
}

// "Resolving" a varslot iterates until it finds a value or an unbound varslot at the end of the
// chain, the canonical varslot.  Exactly one of the return values is not nil.

func (v *Varslot) resolve() (ValueTerm, *Varslot) {
	for v.val == nil && v.next != nil {
		v = v.next
	}
	if v.val != nil {
		return v.val, nil
	}
	return nil, v
}

// `Atom`: a name with object identity.

type Atom struct {
	name string
}

func (st *Store) NewAtom(name string) *Atom {
	if v, ok := st.atoms[name]; ok {
		return v
	}
	v := &Atom{name: name}
	st.atoms[name] = v
	return v
}

func (a *Atom) String() string {
	return a.name
}

func (a *Atom) ruleTermTag() string {
	return "atom"
}

func (a *Atom) valueTermTag() string {
	return "atom"
}

// `Number`: an i64 value, for now

type Number struct {
	value int64
}

func (st *Store) NewNumber(num int64) *Number {
	return &Number{value: num}
}

func (a *Number) String() string {
	return fmt.Sprint(a.value)
}

func (a *Number) ruleTermTag() string {
	return "number"
}

func (a *Number) valueTermTag() string {
	return "number"
}

// `Local`: a term that holds an index into a rib of variables for the current rule.

type Local struct {
	slot int
}

func (st *Store) NewLocal(index int) *Local {
	return &Local{index}
}

func (a *Local) String() string {
	return fmt.Sprintf("V%d", a.slot)
}

func (a *Local) ruleTermTag() string {
	return "local"
}

// `Structure`: the representation of facts, predicates, and queries.

type RuleStruct struct {
	functor  *Atom
	subterms []RuleTerm
}

func (st *Store) NewStruct(functor *Atom, subterms []RuleTerm) *RuleStruct {
	return &RuleStruct{functor, subterms}
}

func (v *RuleStruct) String() string {
	var b strings.Builder
	b.WriteString(v.functor.String())
	if len(v.subterms) > 0 {
		b.WriteRune('(')
		for i, a := range v.subterms {
			if i > 0 {
				b.WriteRune(',')
			}
			b.WriteString(a.String())
		}
		b.WriteRune(')')
	}
	return b.String()
}

func (a *RuleStruct) ruleTermTag() string {
	return "struct"
}

type ValueStruct struct {
	env rib
	s   *RuleStruct
}

func (v *ValueStruct) valueTermTag() string {
	return "struct"
}

func (v *ValueStruct) String() string {
	var b strings.Builder
	b.WriteString(v.s.functor.String())
	if len(v.s.subterms) > 0 {
		b.WriteRune('(')
		for i, a := range v.s.subterms {
			if i > 0 {
				b.WriteRune(',')
			}
			b.WriteString(bind(a, v.env).String())
		}
		b.WriteRune(')')
	}
	return b.String()
}

// Rules represent rules in the database or queries.  The head may be any term, and for ease
// of processing we've broken it out into its components.  For a query, the head is just a
// fact, we use true/0.  Rules are compiled.  The `locals` member is the number of varslots to
// allocate for the rib, representing the number of variables in the rule.

type rule struct {
	locals  int
	arity   int
	functor *Atom
	formals []RuleTerm
	body    []RuleTerm
}
