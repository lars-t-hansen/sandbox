package main

import (
	E "resolver/engine"
)

func main() {
	st := E.NewStore()

	// :- father(haakon, olav).
	// :- father(olav, harald).
	// :- father(harald, 'håkon magnus').
	// :- father('håkon magnus', 'ingrid alexandra').

	father := st.Symbol("father")
	haakon := st.Symbol("haakon")
	olav := st.Symbol("olav")
	harald := st.Symbol("harald")
	krompen := st.Symbol("håkon magnus")
	prinsessa := st.Symbol("ingrid alexandra")
	st.AssertFact(father, haakon, olav)
	st.AssertFact(father, olav, harald)
	st.AssertFact(father, harald, krompen)
	st.AssertFact(father, krompen, prinsessa)

	// ?- father(X, harald).

	names, locals := st.Vars("X")
	query := st.QueryTerm(st.Struct(father, harald, locals[0]))
	st.EvaluateQuery(query, names)

	// grandfather(X, Y) :- father(X, Z), father(Z, Y).

	grandfather := st.Symbol("grandfather")
	_, locals = st.Vars("X", "Y", "Z")
	st.AssertRule(
		locals,
		st.Struct(grandfather, locals[0], locals[1]),
		/* :- */
		st.Struct(father, locals[0], locals[2]),
		st.Struct(father, locals[2], locals[1]))

	// ?- grandfather(harald, X).

	names, locals = st.Vars("X")
	query = st.QueryTerm(st.Struct(grandfather, harald, locals[0]))
	st.EvaluateQuery(query, names)
}
