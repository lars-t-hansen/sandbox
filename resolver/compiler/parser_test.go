package compiler

import (
	"strings"
	"testing"
)

// Note blank line at the beginning
var parser_input string = `  
/* kongerekka. */
:- father(haakon  , olav).
:- father(olav,harald).

?- father(X, 'ingrid alexandra').

grandfather(X,Y) :- father (X, Z ), father(Z, Y).

:- n(1367).
`

func TestParser(t *testing.T) {
	items := parsePhrases(strings.NewReader(parser_input))
	k := 0
	item := items[k]
	k++
	_, ok := item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	item = items[k]
	k++
	_, ok = item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	item = items[k]
	k++
	_, ok = item.(*astQuery)
	if !ok {
		t.Fatalf("Not a query %v", item)
	}
	item = items[k]
	k++
	_, ok = item.(*astRule)
	if !ok {
		t.Fatalf("Not a rule %v", item)
	}
	item = items[k]
	k++
	_, ok = item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	if k != len(items) {
		t.Fatalf("Unexpected non-EOF %v", item)
	}
}
