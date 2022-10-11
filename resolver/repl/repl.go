package repl

import (
	"os"
	"resolver/engine"
)

type reader interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

func processQuerySuccess(names []*engine.Atom, vars []engine.Varslot) bool {
	for i, n := range names {
		if n != nil {
			os.Stdout.WriteString(n.String() + "=" + vars[i].String() + "\n")
		}
	}
	os.Stdout.WriteString("yes\n")
	// But here we could return false if we want to induce failure and look for another result
	return true
}

func processQueryFailure() {
	os.Stdout.WriteString("no\n")
}

func Repl(st *engine.Store, r reader) {
	ctx := newParser(st, processQuerySuccess, processQueryFailure)
	t := newTokenizer(r, ctx)
	if yyParse(t) != 0 {
		panic("Parse failed")
	}
}
