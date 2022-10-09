package main

import (
	"resolver/compiler"
	"resolver/engine"
	"strings"
)

var input string = `

:- father(haakon, olav).

:- father(olav, harald).
:- father(harald, 'håkon magnus').
:- father('håkon magnus', 'ingrid alexandra').


?- father(X, harald).

grandfather(X, Y) :- father(X, Z), father(Z, Y).

?- grandfather(harald, X).
`

func main() {
	st := engine.NewStore()
	compiler.Repl(st, strings.NewReader(input))
}
