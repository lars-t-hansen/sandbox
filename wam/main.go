package main

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

type store struct {
	// Interned atoms.
	atoms map[string]*atom

	// Database of rules.  This is indexed by the functor and arity of the head.
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
	args    []ruleTerm
	body    []ruleTerm
}

type valueTerm interface {
	fmt.Stringer
	valueTermTag() string
}

// "Resolving" a variable iterates until it finds a value or an unbound varslot at the end of the
// chain, the canonical varslot.  Exactly one of the return values is not nil.

func (v *varslot) resolve(e rib) (valueTerm, *varslot) {
	for v.val == nil && v.next != nil {
		v = v.next
	}
	if v.val != nil {
		return v.val, nil
	}
	return nil, v
}

// Evaluation is CPS-based for now, this is not very efficient but is semantically clean.  If
// unification succeeds locally then the continuation is invoked, and if there are not effects
// to undo then that invocation can be a tail call.  If there are effects then the invocation
// is a non-tail call.  If the continuation returns false then we undo the effects.
//
// The idea here is that the ultimate continuation passed in from the repl prints out the current
// bindings of the variables in the query and waits for feedback from the user.  If the user
// says to fail and retry then false is returned and we backtrack.  If the user says to succeed then
// true is returned and we commit and return all the way out.

func unify(e rib, val1 valueTerm, val2 valueTerm, k func() bool) bool {
	var var1, var2 *varslot
	// TODO: As an optimization we want the varslots in the rib to be updated to point to the
	// canonical var here so that we don't have to search as many steps later.  This is not
	// important for correctness though so do it later.
	if ub1, ok := val1.(*varslot); ok {
		val1, var1 = ub1.resolve(e)
	}
	if ub2, ok := val2.(*varslot); ok {
		val2, var2 = ub2.resolve(e)
	}
	if var1 != nil {
		if var2 != nil {
			if var1 != var2 {
				ASSERT(var1.next == nil && var2.next == nil)
				ASSERT(var1.val == nil && var2.val == nil)
				// Arbitrarily make the second point to the first
				var2.next = var1
				if !k() {
					var2.next = nil
					return false
				}
				return true
			}
		}
		ASSERT(var1.next == nil && var1.val == nil)
		var1.val = val2
		if !k() {
			var1.val = nil
			return false
		}
		return true
	}
	if var2 != nil {
		ASSERT(var2.next == nil && var2.val == nil)
		var2.val = var1
		if !k() {
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
			return unify_terms(e, bind_terms(s1.s.subterms, s1.env), bind_terms(s2.s.subterms, s2.env), k)
		}
		return false
	}
	if a1, ok := val1.(*atom); ok {
		if a2, ok := val2.(*atom); ok {
			if a1 == a2 {
				return k()
			}
		}
		return false
	}
	if n1, ok := val1.(*number); ok {
		if n2, ok := val2.(*number); ok {
			if n1.value == n2.value {
				return k()
			}
		}
		return false
	}
	return false
}

func unify_terms(e rib, s1 []valueTerm, s2 []valueTerm, k func() bool) bool {
	if len(s1) == 0 {
		return k()
	}
	return unify(e, s1[0], s2[0], func() bool {
		return unify_terms(e, s1[1:], s2[1:], k)
	})
}

func (st *store) evaluate_rule(r *rule, actuals []valueTerm, k func() bool) bool {
	ASSERT(len(actuals) == r.arity)
	newRib := make(rib, r.locals)
	return unify_terms(newRib, actuals, bind_terms(r.args, newRib), func() bool {
		return st.evaluate_conjunct(newRib, r.body, k)
	})
}

func (st *store) evaluate_conjunct(e rib, ts []ruleTerm, k func() bool) bool {
	if len(ts) == 0 {
		return k()
	}
	switch t := ts[0].(type) {
	case *number, *atom, *local:
		return k()
	case *unboundStruct:
		candidates := st.lookup(t.functor, len(t.subterms))
		return st.evaluate_disjunct(e, bind_terms(t.subterms, e), candidates, func() bool {
			return st.evaluate_conjunct(e, ts[1:], k)
		})
	default:
		panic(fmt.Sprintf("No such structure %v", t))
	}
}

func (st *store) evaluate_disjunct(e rib, actuals []valueTerm, disjuncts []*rule, k func() bool) bool {
	for _, d := range disjuncts {
		if st.evaluate_rule(d, actuals, k) {
			return true
		}
	}
	return false
}

// consider f(X) :- g(h(i(X)))
//
// the X is a reference to the local rib but the X in h(i(X)) has to be represented
// as a varslot, or as a (rib, index) pair - either way it is no longer some kind
// of constant.  In the structure in the rule it is local(0) but this is insufficient.
//
// It is possible that terms can be passed along with their ribs?
//
// That is, when a rule is invoked a new entity is created that pairs the body terms
// with a rib.  Each term would have to have a reference to the rib somehow.  This
// reference would have to be maintained as the structure is decomposed and so on,
// leading to the context (the rib) for a variable being passed everywhere as part
// of the term.
//
// In particular, when a var is unified with a term, the term's context also has to
// be stored in the var.
//
// The idea of locals was to avoid having to rebuild the rule body every time we
// enter a rule.  But this has pushed the complexity elsewhere.  In truth, h(i(X)) is
// like a closure, and it needs to retain a reference to its environment to be
// evaluated properly.
//
// This complexity is not seen in f(X) :- g(X), for example, only when the variable
// is hidden inside a structure -- again, it's a closure.
//
// So, a structure is a closure that retains a reference to the environment that
// holds the closed-over variables.  In h(i(X)), the representation is h(.) + e
// and when we descend into h, it becomes i(.) + e, and when we encounter X inside
// i we look it up in e.
//
// Thus a structure in a rule is *not* the same thing as a structure value, rather
// it is like a lambda expression, being input to closure creation.

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
/*
func (st *store) newStruct(name string, args ...term) *structure {
	return &structure{functor: st.intern(name), subterms: args}
}

func (st *store) atom(name string) *atom {
	return st.intern(name)
}
*/

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
/*
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
*/
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
/*
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

func newFact(head *structure) *rule {
	return &rule{head: head, body: []*structure{}}
}
*/

/*
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
*/

func main() {
	st := newStore()

	// :- father(haakon, olav).
	// :- father(olav, harald).
	// :- father(harald, 'håkon magnus').
	// :- father('håkon magnus', 'ingrid alexandra').

	empty := []ruleTerm{}
	father := st.intern("father")
	haakon := st.intern("haakon")
	olav := st.intern("olav")
	harald := st.intern("harald")
	krompen := st.intern("håkon magnus")
	prinsessa := st.intern("ingrid alexandra")
	st.assert(&rule{0, 2, father, []ruleTerm{haakon, olav}, empty})
	st.assert(&rule{0, 2, father, []ruleTerm{olav, harald}, empty})
	st.assert(&rule{0, 2, father, []ruleTerm{harald, krompen}, empty})
	st.assert(&rule{0, 2, father, []ruleTerm{krompen, prinsessa}, empty})

	// ?- father(X, harald)

	X := st.intern("X")
	query := []ruleTerm{&unboundStruct{father, []ruleTerm{&local{0}, harald}}}
	names := []*atom{X}
	vars := make(rib, 1)
	result := st.evaluate_conjunct(vars, query, func() bool {
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
