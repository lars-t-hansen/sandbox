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
// invoked and fresh variables are created and referenced from the clone.)  See
// values.go for more.

// TODO:
// - some built-in predicates, notably `is`
// - the obvious first step is to insert an explicit failure continuation
// - another obvious step is to introduce "cut", possibly "fail"
// - the continuations can later be reified and the engine recoded as a state machine with
//   explicit data structures

package engine

func assert(b bool) {
	if !b {
		panic("Assertion failed")
	}
}

// Evaluation is quasi-CPS-based for now, this is not very efficient but is semantically clean.
// If unification succeeds locally then the success continuation is invoked, and if there are
// no effects to undo then that invocation can be a tail call.  If there are effects then the
// invocation is a non-tail call - the failure continuation is encoded in the call stack.  If
// the success continuation returns false then we undo the effects.

func unify(val1 ValueTerm, val2 ValueTerm, onSuccess func() bool) bool {
	var var1, var2 *Varslot
	// TODO: As an optimization we want the varslots in the rib to be updated to point to the
	// canonical var here so that we don't have to search as many steps later.
	if ub1, ok := val1.(*Varslot); ok {
		val1, var1 = ub1.resolve()
	}
	if ub2, ok := val2.(*Varslot); ok {
		val2, var2 = ub2.resolve()
	}
	if var1 != nil {
		if var2 != nil {
			if var1 != var2 {
				assert(var1.next == nil && var2.next == nil)
				assert(var1.val == nil && var2.val == nil)
				// Arbitrarily make the second point to the first
				var2.next = var1
				if !onSuccess() {
					var2.next = nil
					return false
				}
				return true
			}
		}
		assert(var1.next == nil && var1.val == nil)
		var1.val = val2
		if !onSuccess() {
			var1.val = nil
			return false
		}
		return true
	}
	if var2 != nil {
		assert(var2.next == nil && var2.val == nil)
		var2.val = val1
		if !onSuccess() {
			var2.val = nil
			return false
		}
		return true
	}
	if s1, ok := val1.(*ValueStruct); ok {
		if s2, ok := val2.(*ValueStruct); ok {
			if s1.s.functor != s2.s.functor || len(s1.s.subterms) != len(s2.s.subterms) {
				return false
			}
			return unify_terms(bind_terms(s1.s.subterms, s1.env), bind_terms(s2.s.subterms, s2.env), onSuccess)
		}
		return false
	}
	if a1, ok := val1.(*Atom); ok {
		if a2, ok := val2.(*Atom); ok {
			if a1 == a2 {
				return onSuccess()
			}
		}
		return false
	}
	if n1, ok := val1.(*Number); ok {
		if n2, ok := val2.(*Number); ok {
			if n1.value == n2.value {
				return onSuccess()
			}
		}
		return false
	}
	return false
}

func unify_terms(s1 []ValueTerm, s2 []ValueTerm, onSuccess func() bool) bool {
	if len(s1) == 0 {
		return onSuccess()
	}
	return unify(s1[0], s2[0], func /* onSuccess */ () bool {
		return unify_terms(s1[1:], s2[1:], onSuccess)
	})
}

func (st *Store) evaluateConjunct(e rib, ts []RuleTerm, onSuccess func() bool) bool {
	if len(ts) == 0 {
		return onSuccess()
	}
	switch t := ts[0].(type) {
	case *Number, *Atom, *Local:
		return onSuccess()
	case *RuleStruct:
		candidates := st.lookupRule(t.functor, len(t.subterms))
		return st.evaluateDisjunct(bind_terms(t.subterms, e), candidates, func /* onSuccess */ () bool {
			return st.evaluateConjunct(e, ts[1:], onSuccess)
		})
	default:
		panic("Unknown term type")
	}
}

func (st *Store) evaluateDisjunct(actuals []ValueTerm, disjuncts []*rule, onSuccess func() bool) bool {
	for _, r := range disjuncts {
		assert(len(actuals) == r.arity)
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

func (st *Store) EvaluateQuery(query []RuleTerm, names []*Atom,
	processQuerySuccess func(names []*Atom, vars []Varslot) bool,
	processQueryFailure func()) {
	vars := make(rib, len(names))
	result := st.evaluateConjunct(vars, query, func /* onSuccess */ () bool {
		return processQuerySuccess(names, vars)
	})
	if !result {
		processQueryFailure()
	}
}
