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
	tok := newTokenizer(strings.NewReader(parser_input))
	p := newParser(tok)
	item := p.parsePhrase()
	_, ok := item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	item = p.parsePhrase()
	_, ok = item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	item = p.parsePhrase()
	_, ok = item.(*astQuery)
	if !ok {
		t.Fatalf("Not a query %v", item)
	}
	item = p.parsePhrase()
	_, ok = item.(*astRule)
	if !ok {
		t.Fatalf("Not a rule %v", item)
	}
	item = p.parsePhrase()
	_, ok = item.(*astFact)
	if !ok {
		t.Fatalf("Not a fact %v", item)
	}
	item = p.parsePhrase()
	if item != nil {
		t.Fatalf("Unexpected non-EOF %v", item)
	}
}
