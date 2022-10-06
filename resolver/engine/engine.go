// Unification-based resolution engine (a la Prolog)
//
// Most of this is pretty straightforward, and currently there's a quasi-CPS evaluation
// strategy in place where the success continuation is passed explicitly and the
// failure continuation uses the regular call stack.
//
// One tricky bit is how variables are managed.  Consider a rule:
//
//   f(X) :- g(h(i(X)))
//
// Here the X is fresh every time the rule is evaluated, but the X in h(i(X)) must
// always reference the X in the rib of f, not some variable in whatever context
// in which we happen to descend into h(i(X)).  The solution here is that h(i(X)) in
// effect is treated as a closure that closes over the environment that has the
// slot for X.  (An alternative is that the rule for f is cloned every time it is
// invoked and fresh variables are created and referenced from the clone.)

// TODO:
// - the obvious first step is to insert an explicit failure continuation
// - another obvious step is to introduce "cut", possibly "fail"
// - the continuations can later be reified and the engine recoded as a state machine with
//   explicit data structures

package engine

import (
	"fmt"
	"os"
	"strings"
)

func ASSERT(b bool) {
	if !b {
		panic("Assertion failed")
	}
}

// Global background state for evaluation

type Store struct {
	// Interned atoms.
	atoms map[string]*atom

	// Database of rules.  This is indexed by the functor and arity of the head.
	rules map[*atom]map[int][]*rule
}

func NewStore() *Store {
	return &Store{
		atoms: make(map[string]*atom),
		rules: make(map[*atom]map[int][]*rule),
	}
}

func (st *Store) Symbol(name string) *atom {
	if v, ok := st.atoms[name]; ok {
		return v
	}
	v := &atom{name: name}
	st.atoms[name] = v
	return v
}

func (st *Store) Number(num int64) *number {
	return &number{value: num}
}

func (st *Store) assert(r *rule) {
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

func (st *Store) lookup(functor *atom, arity int) []*rule {
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

// Atoms are names with object identity.

type atom struct {
	name string
}

func (a *atom) String() string {
	return a.name
}

func (a *atom) ruleTermTag() string {
	return "atom"
}

func (a *atom) valueTermTag() string {
	return "atom"
}

// Numbers are i64, for now

type number struct {
	value int64
}

func (a *number) String() string {
	return fmt.Sprint(a.value)
}

func (a *number) ruleTermTag() string {
	return "number"
}

func (a *number) valueTermTag() string {
	return "number"
}

// Locals are indices into a rib of variables for the current rule.  (In principle
// the local could also carry a name.)

type local struct {
	slot int
}

func (a *local) String() string {
	return fmt.Sprintf("V%d", a.slot)
}

func (a *local) ruleTermTag() string {
	return "local"
}

// Varslots are storage for variables.  They are allocated inside ribs, which are themselves
// allocated when predicates are evaluated.
//
// If `val` is not nil then it is the value held in this slot.  Otherwise, `next` is either nil,
// in which case this is the canonical varslot for a variable, or it points to another varslot
// that this varslot has been unified with.

type varslot struct {
	next *varslot
	val  valueTerm
}

func (v *varslot) String() string {
	ASSERT(v != nil)
	if v.val != nil {
		return "[value " + v.val.String() + "]"
	}
	return "[varslot]"
}

func (v *varslot) valueTermTag() string {
	return "[varslot]"
}

type rib []varslot

// Structures represent facts or predicates.

type unboundStruct struct {
	functor  *atom
	subterms []ruleTerm
}

func (v *unboundStruct) String() string {
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

func (a *unboundStruct) ruleTermTag() string {
	return "struct"
}

func bind(t ruleTerm, e rib) valueTerm {
	ASSERT(t != nil)
	switch x := t.(type) {
	case *unboundStruct:
		return &boundStruct{env: e, s: x}
	case *atom:
		return x
	case *number:
		return x
	case *local:
		return &e[x.slot]
	default:
		panic("NYI")
	}
}

func bind_terms(ts []ruleTerm, e rib) []valueTerm {
	vs := make([]valueTerm, len(ts))
	for i, t := range ts {
		vs[i] = bind(t, e)
	}
	return vs
}

type boundStruct struct {
	env rib
	s   *unboundStruct
}

func (v *boundStruct) valueTermTag() string {
	return "struct"
}

func (v *boundStruct) String() string {
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

type ruleTerm interface {
	fmt.Stringer
	ruleTermTag() string
}

type rule struct {
	locals  int
	arity   int
	functor *atom
	formals []ruleTerm
	body    []ruleTerm
}

type valueTerm interface {
	fmt.Stringer
	valueTermTag() string
}

// "Resolving" a variable iterates until it finds a value or an unbound varslot at the end of the
// chain, the canonical varslot.  Exactly one of the return values is not nil.

func (v *varslot) resolve() (valueTerm, *varslot) {
	for v.val == nil && v.next != nil {
		v = v.next
	}
	if v.val != nil {
		return v.val, nil
	}
	return nil, v
}

// Evaluation is quasi-CPS-based for now, this is not very efficient but is semantically clean.
// If unification succeeds locally then the success continuation is invoked, and if there are
// no effects to undo then that invocation can be a tail call.  If there are effects then the
// invocation is a non-tail call - the failure continuation is encoded in the call stack.  If
// the success continuation returns false then we undo the effects.

func unify(val1 valueTerm, val2 valueTerm, onSuccess func() bool) bool {
	var var1, var2 *varslot
	// TODO: As an optimization we want the varslots in the rib to be updated to point to the
	// canonical var here so that we don't have to search as many steps later.
	if ub1, ok := val1.(*varslot); ok {
		val1, var1 = ub1.resolve()
	}
	if ub2, ok := val2.(*varslot); ok {
		val2, var2 = ub2.resolve()
	}
	if var1 != nil {
		if var2 != nil {
			if var1 != var2 {
				ASSERT(var1.next == nil && var2.next == nil)
				ASSERT(var1.val == nil && var2.val == nil)
				// Arbitrarily make the second point to the first
				var2.next = var1
				if !onSuccess() {
					var2.next = nil
					return false
				}
				return true
			}
		}
		ASSERT(var1.next == nil && var1.val == nil)
		var1.val = val2
		if !onSuccess() {
			var1.val = nil
			return false
		}
		return true
	}
	if var2 != nil {
		ASSERT(var2.next == nil && var2.val == nil)
		var2.val = val1
		if !onSuccess() {
			var2.val = nil
			return false
		}
		return true
	}
	if s1, ok := val1.(*boundStruct); ok {
		if s2, ok := val2.(*boundStruct); ok {
			if s1.s.functor != s2.s.functor || len(s1.s.subterms) != len(s2.s.subterms) {
				return false
			}
			return unify_terms(bind_terms(s1.s.subterms, s1.env), bind_terms(s2.s.subterms, s2.env), onSuccess)
		}
		return false
	}
	if a1, ok := val1.(*atom); ok {
		if a2, ok := val2.(*atom); ok {
			if a1 == a2 {
				return onSuccess()
			}
		}
		return false
	}
	if n1, ok := val1.(*number); ok {
		if n2, ok := val2.(*number); ok {
			if n1.value == n2.value {
				return onSuccess()
			}
		}
		return false
	}
	return false
}

func unify_terms(s1 []valueTerm, s2 []valueTerm, onSuccess func() bool) bool {
	if len(s1) == 0 {
		return onSuccess()
	}
	return unify(s1[0], s2[0], func /* onSuccess */ () bool {
		return unify_terms(s1[1:], s2[1:], onSuccess)
	})
}

func (st *Store) evaluateConjunct(e rib, ts []ruleTerm, onSuccess func() bool) bool {
	if len(ts) == 0 {
		return onSuccess()
	}
	switch t := ts[0].(type) {
	case *number, *atom, *local:
		return onSuccess()
	case *unboundStruct:
		candidates := st.lookup(t.functor, len(t.subterms))
		return st.evaluateDisjunct(bind_terms(t.subterms, e), candidates, func /* onSuccess */ () bool {
			return st.evaluateConjunct(e, ts[1:], onSuccess)
		})
	default:
		panic(fmt.Sprintf("No such structure %v", t))
	}
}

func (st *Store) evaluateDisjunct(actuals []valueTerm, disjuncts []*rule, onSuccess func() bool) bool {
	for _, r := range disjuncts {
		ASSERT(len(actuals) == r.arity)
		newRib := make(rib, r.locals)
		res := unify_terms(actuals, bind_terms(r.formals, newRib), func /* onSuccess */ () bool {
			return st.evaluateConjunct(newRib, r.body, onSuccess)
		})
		if res {
			return true
		}
	}
	return false
}

func (st *Store) EvaluateQuery(query []ruleTerm, names []*atom) {
	vars := make(rib, len(names))
	result := st.evaluateConjunct(vars, query, func /* onSuccess */ () bool {
		for i, n := range names {
			os.Stdout.WriteString(n.name + "=" + vars[i].String() + "\n")
		}
		return true
	})
	if result {
		os.Stdout.WriteString("yes\n")
	} else {
		os.Stdout.WriteString("no\n")
	}
}

// Convenience functions

func (st *Store) AssertFact(functor *atom, subterms ...ruleTerm) {
	st.assert(&rule{0, len(subterms), functor, subterms, []ruleTerm{}})
}

func (st *Store) AssertRule(locals []*local, head *unboundStruct, subterms ...ruleTerm) {
	st.assert(&rule{len(locals), len(head.subterms), head.functor, head.subterms, subterms})
}

func (st *Store) Vars(names ...string) ([]*atom, []*local) {
	as := make([]*atom, len(names))
	ls := make([]*local, len(names))
	for i, name := range names {
		as[i] = st.Symbol(name)
		ls[i] = &local{i}
	}
	return as, ls
}

func (st *Store) QueryTerm(terms ...ruleTerm) []ruleTerm {
	return terms
}

func (st *Store) Struct(functor *atom, subterms ...ruleTerm) *unboundStruct {
	return &unboundStruct{functor, subterms}
}
