package compiler

import (
	"strings"
	"testing"
)

var input string = `
/* kongerekka. */
:- father(haakon, olav).

?- father(X, harald).

grandfather(X, Y) :- father(X, Z), father(Z, Y).

:- n(1367).
`

func checkToken(t *testing.T, tok *tokenizer, expected token) {
	ntok := tok.next()
	if ntok.kind != expected.kind || ntok.name != expected.name {
		t.Fatalf("Line %d: Unexpected %v, expected %v", ntok.lineno, ntok, expected)
	}
}
func TestTokenizer(t *testing.T) {
	tok := newTokenizer(strings.NewReader(input))
	checkToken(t, tok, token{t_operator, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "haakon"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "olav"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_operator, 0, "?-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "harald"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "grandfather"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "Y"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_operator, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "Z"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "Z"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "Y"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_operator, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "n"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_number, 0, "1367"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_eof, 0, ""})
}
