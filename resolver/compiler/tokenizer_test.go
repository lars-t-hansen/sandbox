package compiler

// Note blank line at the beginning
var tokenizer_input string = `  
/* kongerekka. */
:- father(haakon  , olav).

?- father(X, 'ingrid alexandra').

grandfather(X,Y) :- father (X, Z ), father(Z, Y).

:- n(1367).
`

/*

func checkToken(t *testing.T, tok *tokenizer, expected token) {
	ntok := tok.readNext()
	if ntok.kind != expected.kind || ntok.name != expected.name {
		t.Fatalf("Line %d: Unexpected %v, expected %v", ntok.lineno, ntok, expected)
	}
}
func TestTokenizer(t *testing.T) {
	tok := newTokenizer(strings.NewReader(tokenizer_input))
	checkToken(t, tok, token{t_infix, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "haakon"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "olav"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_infix, 0, "?-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "ingrid alexandra"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "grandfather"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "Y"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_infix, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "X"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "Z"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_atom, 0, "father"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "Z"})
	checkToken(t, tok, token{t_comma, 0, ""})
	checkToken(t, tok, token{t_varname, 0, "Y"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_infix, 0, ":-"})
	checkToken(t, tok, token{t_atom, 0, "n"})
	checkToken(t, tok, token{t_lparen, 0, ""})
	checkToken(t, tok, token{t_number, 0, "1367"})
	checkToken(t, tok, token{t_rparen, 0, ""})
	checkToken(t, tok, token{t_period, 0, ""})
	checkToken(t, tok, token{t_eof, 0, ""})
	if tok.lineno != 10 {
		t.Fatalf("Line numbers are off.  Got %v, expected %v", tok.lineno, 10)
	}
}
*/
