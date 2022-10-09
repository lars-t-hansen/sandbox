package main

import (
	"os"
	"resolver/engine"
	"resolver/repl"
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
	repl.Repl(st, strings.NewReader(input), func(s string) { os.Stdout.WriteString(s) })
}
