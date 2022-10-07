%{
package compiler

import (
	"fmt"
    "io"
	"strconv"
)

type astVar struct {
	lineno int
	name   string
	index  int
}

type astAtom struct {
	lineno int
	name   string
}

type astNumber struct {
	lineno int
	value  int64
}

type astStruct struct {
	lineno     int
	name       string
	components []astTerm
}

// type astTerm union {
//   *astAtom
//   *astNumber
//   *astStruct
//   *astVar
// }

type astTerm interface {
	fmt.Stringer
	line() int
}

type astQuery struct {
	vars []*astVar
	body []astTerm
}

type astFact struct {
	head astTerm
}

type astRule struct {
	vars []*astVar
	head astTerm
	body []astTerm
}

// type astPhrase union {
//   *astFact
//   *astRule
//   *astQuery
//   nil
// }

type astPhrase interface {
	fmt.Stringer
	line() int
}
%}

%union { 
    name string
    terms []astTerm
    term astTerm
    phrases []astPhrase
    phrase astPhrase
}

%start Program

%token <name> T_ATOM T_NUMBER T_VARNAME
%left <name> T_INFIX_OP
%token T_LPAREN T_RPAREN T_COMMA T_PERIOD T_FACT_OP
%left T_QUERY_OP

%type <phrases> Phrases
%type <phrase> Phrase Fact Query Rule
%type <terms> Terms
%type <term> Term Struct Atom Number Variable

%%

Program : Phrases
            {
				setResult(yylex, $1)
            }
        ;
Phrases :  /* empty */
            {
                $$ = []astPhrase{}
            }
        | Phrases Phrase
            {
                $$ = append($1, $2)
            }
        ;
Phrase  : Fact | Query | Rule ;
Fact    : T_FACT_OP Struct T_PERIOD
            {
				vars := getVars(yylex)
				if len(vars) != 0 {
					yylex.Error("Facts should not have free variables")
					// TODO: how to recover here?
				}
	            $$ = &astFact{$2}
            }
        ;
Query   : T_QUERY_OP Terms T_PERIOD
            {
                $$ = &astQuery{getVars(yylex), $2}
            }
        ;
Rule    : Struct T_FACT_OP Terms T_PERIOD
            {
                $$ = &astRule{getVars(yylex), $1, $3}
            }
        ;
Term    : Struct | Atom | Number | Variable ;
Terms   : Term
            {
                $$ = []astTerm{$1}
            }
        | Terms T_COMMA Term
            {
                $$ = append($1, $3)
            }
        ;
Struct  : T_ATOM T_LPAREN Terms T_RPAREN
            {
                $$ = &astStruct{lineno(yylex), $1, $3}
            }
        | Term T_INFIX_OP Term
            {
                $$ = &astStruct{lineno(yylex), $2, []astTerm{$1, $3}}
            }
        ;
Atom    : T_ATOM
            {
                $$ = &astAtom{lineno(yylex), $1}
            }
        ;
Number  : T_NUMBER
            {
                val, err := strconv.ParseInt($1, 10, 64)
				if err != nil {
					yylex.Error("numeric overflow")
					// TODO: how to recover here?
				}
                $$ = &astNumber{lineno(yylex), val}
            }
        ;
Variable : T_VARNAME
            {
				$$ = newVariable(yylex, $1)
            }
        ;

%%

/////////////////////////////////////////////////////////////////////////////////////////////////
//
// AST

func (n *astVar) String() string {
	return fmt.Sprintf("[%s %d]", n.name, n.index)
}

func (n *astVar) line() int {
	return n.lineno
}

func (n *astAtom) String() string {
	return n.name
}

func (n *astAtom) line() int {
	return n.lineno
}

func (n *astNumber) String() string {
	return fmt.Sprint(n.value)
}

func (n *astNumber) line() int {
	return n.lineno
}

func (n *astStruct) String() string {
	args := ""
	for _, c := range n.components {
		if args != "" {
			args = args + ","
		}
		args = args + c.String()
	}
	return n.name + "(" + args + ")"
}

func (n *astStruct) line() int {
	return n.lineno
}


func (f *astQuery) String() string {
	return "?- " + termsToString(f.body) + "."
}

func (f *astQuery) line() int {
	return f.body[0].line()
}

func (f *astFact) String() string {
	return ":-" + f.head.String() + "."
}

func (f *astFact) line() int {
	return f.head.line()
}

func (f *astRule) String() string {
	return f.head.String() + " :- " + termsToString(f.body) + "."
}

func (f *astRule) line() int {
	return f.head.line()
}

func termsToString(ts []astTerm) string {
	s := ""
	for _, term := range ts {
		if s != "" {
			s = s + ", "
		}
		s = s + term.String()
	}
	return s
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//
// Tokenizer

type reader interface {
	ReadRune() (rune, int, error)
	UnreadRune() error
}

type tokenizer2 struct {
	// Input characters
	input  reader

	// Line number at start of previous token returned
	lineno int

	// The rest of this is parser context, see parser code further down.
	ctx *parserctx
}

func newTokenizer2(r reader, ctx *parserctx) *tokenizer2 {
    return &tokenizer2{
		input: r, 
		lineno: 0,
		ctx: ctx,
	}
}

func (l *tokenizer2) Lex(lval *yySymType) int {
	t, name := l.get()
	lval.name = name
	return t
}

func (l *tokenizer2) Error(s string) {
	panic(s)
}

func lineno(l yyLexer) int {
	return l.(*tokenizer2).lineno
}

func (t *tokenizer2) peekChar() rune {
	r, _, err := t.input.ReadRune()
	if err == io.EOF {
		return -1
	}
	if err != nil {
		panic(fmt.Sprintf("Line %d: Bad input: "+err.Error(), t.lineno))
	}
	t.input.UnreadRune()
	return r
}

func (t *tokenizer2) getChar() rune {
	r, _, err := t.input.ReadRune()
	if err == io.EOF {
		return -1
	}
	if err != nil {
		panic(fmt.Sprintf("Line %d: Bad input: "+err.Error(), t.lineno))
	}
	return r
}

func (t *tokenizer2) get() (int, string) {
outer:
	for {
		r := t.getChar()
		if r == -1 {
			return -1, ""
		}
		if r == '\t' || r == ' ' {
			continue
		}
		if r == '\n' {
			t.lineno++
			continue
		}
		if r == '/' && t.peekChar() == '*' {
			t.getChar()
			for {
				r := t.getChar()
				if r == -1 {
					panic(fmt.Sprintf("Line %d: EOF in comment", t.lineno))
				}
				if r == '*' && t.peekChar() == '/' {
					t.getChar()
					continue outer
				}
				if r == '\n' {
					t.lineno++
				}
			}
		}
		if r == '(' {
			return T_LPAREN, ""
		}
		if r == ')' {
			return T_RPAREN, ""
		}
		if r == '.' {
			return T_PERIOD, ""
		}
		if r == ',' {
			return T_COMMA, ""
		}
		if r == '-' {
			if isDigitChar(t.peekChar()) {
				return T_NUMBER, t.lexWhile(isDigitChar, "-")
			}
		}
		if r == '\'' {
			name := t.lexWhile(func(r rune) bool {
				return r != -1 && r != '\'' && r != '\n' && r != '\r'
			}, "")
			if t.getChar() != '\'' {
				panic(fmt.Sprintf("Line %d: unterminated quoted atom", t.lineno))
			}
			return T_ATOM, name
		}
		if isOperatorChar(r) {
            name := t.lexWhile(isOperatorChar, string(r))
            if name == "?-" {
                return T_QUERY_OP, ""
            }
            if name == ":-" {
                return T_FACT_OP, ""
            }
			return T_INFIX_OP, name
		}
		if isDigitChar(r) {
			return T_NUMBER, t.lexWhile(isDigitChar, string(r))
		}
		if isVarFirstChar(r) {
			return T_VARNAME, t.lexWhile(isAtomNextChar, string(r))
		}
		if isAtomFirstChar(r) {
			// TODO: This strikes me as a hack, there should be a more principled solution
			// to this somewhere.
			name := t.lexWhile(isAtomNextChar, string(r))
			if name == "is" {
				return T_INFIX_OP, name
			}
			return T_ATOM, name
		}
		panic(fmt.Sprintf("Line %d: bad character: %v", t.lineno, r))
	}
}

// This depends on isChar() being false for -1 and newlines
func (t *tokenizer2) lexWhile(isChar func(r rune) bool, s string) string {
	for isChar(t.peekChar()) {
		s = s + string(t.getChar())
	}
	return s
}

func isOperatorChar(r rune) bool {
	return r == '+' || r == '-' || r == '?' || r == ':' || r == '!' || r == '='
}

func isDigitChar(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAtomFirstChar(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isVarFirstChar(r rune) bool {
	return r >= 'A' && r <= 'Z' || r == '_'
}

func isAtomNextChar(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || r >= '0' && r <= '9'
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//
// Parser interface

// An alternative here would be to pass in a callback and for the grammar to act on each phrase
// as it is encountered.  That might also play a little better in an interactive system.

// Sticking parser context on yylex is how the cool kids do it, but it's really not pretty.

type parserctx struct {
	// Next index for a variable in the clause
	varIndex int

	// All unique variables in the current clause.
	vars []*astVar

	nameMap map[string]int

	// Set to the slice of phrases when the parse succeeds
	result []astPhrase
}

func parsePhrases(r reader) []astPhrase {
	ctx := &parserctx{
		varIndex: 0, 
		vars: make([]*astVar, 0),
		nameMap: make(map[string]int, 0),
	}
	t := newTokenizer2(r, ctx)
	if yyParse(t) == 0 {
		return ctx.result
	}
	panic("Parse failed")
}

func setResult(l yyLexer, r []astPhrase) {
	l.(*tokenizer2).ctx.result = r
}

func newVariable(l yyLexer, name string) *astVar {
	t := l.(*tokenizer2)
	p := t.ctx
	if name == "_" {
		// Fresh anonymous variable
		index := p.varIndex
		p.varIndex++
		v := &astVar{t.lineno, name, index}
		p.vars = append(p.vars, v)
		return v
	}

	index, found := p.nameMap[name]
	if found {
		// Previously seen variable
		return &astVar{t.lineno, name, index}
	}

	// Fresh named variable
	index = p.varIndex
	p.varIndex++
	p.nameMap[name] = index
	v := &astVar{t.lineno, name, index}
	p.vars = append(p.vars, v)
	return v
}

func getVars(l yyLexer) []*astVar {
	p := l.(*tokenizer2).ctx
	vs := p.vars
	p.varIndex = 0
	p.vars = p.vars[0:0]
	for k := range p.nameMap {
		delete(p.nameMap, k)
	}
	return vs
}
