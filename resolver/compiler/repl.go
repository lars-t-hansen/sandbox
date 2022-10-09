package compiler

//go:generate goyacc -o parser.go parser.y

import "resolver/engine"

func Repl(st *engine.Store, r reader, writeString func(string)) {
	ctx := newParser(st, writeString)
	t := newTokenizer(r, ctx)
	if yyParse(t) != 0 {
		panic("Parse failed")
	}
}
