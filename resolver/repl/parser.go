// Code generated by goyacc -o parser.go parser.y. DO NOT EDIT.

//line parser.y:2
//go:generate goyacc -o parser.go parser.y

package repl

import __yyfmt__ "fmt"

//line parser.y:4

import (
	"fmt"
	"resolver/engine"
	"strconv"
)

//line parser.y:13
type yySymType struct {
	yys   int
	text  string
	terms []engine.RuleTerm
	term  engine.RuleTerm
}

const T_ATOM = 57346
const T_NUMBER = 57347
const T_VARNAME = 57348
const T_INFIX_OP = 57349
const T_LPAREN = 57350
const T_RPAREN = 57351
const T_COMMA = 57352
const T_PERIOD = 57353
const T_FACT_OP = 57354
const T_QUERY_OP = 57355

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"T_ATOM",
	"T_NUMBER",
	"T_VARNAME",
	"T_INFIX_OP",
	"T_LPAREN",
	"T_RPAREN",
	"T_COMMA",
	"T_PERIOD",
	"T_FACT_OP",
	"T_QUERY_OP",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser.y:93

func (t *tokenizer) Lex(lval *yySymType) (tok int) {
	tok, lval.text = t.get()
	return
}

func (t *tokenizer) Error(s string) {
	panic(fmt.Sprintf("Line %d: %s", t.lineno, s))
}

func parser(l yyLexer) *parserctx {
	return l.(*tokenizer).ctx
}

type parserctx struct {
	st *engine.Store

	// Next index for a variable in the clause
	varIndex int

	// All unique variables in the current clause.
	vars []*engine.Local

	nameMap map[string]int

	processQuerySuccess func([]*engine.Atom, []engine.Varslot) bool
	processQueryFailure func()
}

func newParser(st *engine.Store,
	processQuerySuccess func([]*engine.Atom, []engine.Varslot) bool,
	processQueryFailure func()) *parserctx {
	return &parserctx{
		st:                  st,
		varIndex:            0,
		vars:                make([]*engine.Local, 0),
		nameMap:             make(map[string]int, 0),
		processQuerySuccess: processQuerySuccess,
		processQueryFailure: processQueryFailure,
	}
}

func (p *parserctx) makeVariable(name string) *engine.Local {
	// Previously seen variable?
	if index, found := p.nameMap[name]; found {
		return p.st.NewLocal(index)
	}

	// Fresh variable
	index := p.varIndex
	p.varIndex++
	if name != "_" {
		p.nameMap[name] = index
	}
	v := p.st.NewLocal(index)
	p.vars = append(p.vars, v)
	return v
}

func (p *parserctx) hasFreeVariables() bool {
	return len(p.vars) > 0
}

func (p *parserctx) getAndClearVars() []*engine.Local {
	vs := p.vars
	p.varIndex = 0
	p.vars = p.vars[0:0]
	for k := range p.nameMap {
		delete(p.nameMap, k)
	}
	return vs
}

func (p *parserctx) evalFact(fact *engine.RuleStruct) {
	p.st.AssertFact(fact)
}

func (p *parserctx) evalQuery(query []engine.RuleTerm) {
	names := make([]*engine.Atom, len(p.nameMap))
	for k, v := range p.nameMap {
		names[v] = p.st.NewAtom(k)
	}
	p.getAndClearVars()
	p.st.EvaluateQuery(query, names, p.processQuerySuccess, p.processQueryFailure)
}

func (p *parserctx) evalRule(head *engine.RuleStruct, body []engine.RuleTerm) {
	p.st.AssertRule(p.getAndClearVars(), head, body)
}

func (p *parserctx) makeStruct(functor string, terms []engine.RuleTerm) *engine.RuleStruct {
	return p.st.NewStruct(p.st.NewAtom(functor), terms)
}

func (p *parserctx) makeNumber(n int64) *engine.Number {
	return p.st.NewNumber(n)
}

func (p *parserctx) makeAtom(name string) *engine.Atom {
	return p.st.NewAtom(name)
}

//line yacctab:1
var yyExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 38

var yyAct = [...]int8{
	20, 18, 19, 11, 10, 15, 16, 24, 11, 27,
	30, 22, 7, 9, 27, 26, 32, 27, 23, 21,
	6, 25, 8, 5, 29, 28, 4, 17, 31, 10,
	15, 16, 3, 2, 14, 13, 12, 1,
}

var yyPact = [...]int16{
	-1000, -1000, 0, -1000, -1000, -1000, -1000, 25, -11, 25,
	3, 11, -1000, -1000, -1000, -1000, -1000, -4, 25, 4,
	11, -1000, 25, 25, -1000, -1, -1000, 25, 7, -1000,
	-1000, 11, -1000,
}

var yyPgo = [...]int8{
	0, 37, 2, 0, 19, 36, 35, 34, 33, 32,
	26, 23, 20,
}

var yyR1 = [...]int8{
	0, 1, 8, 8, 9, 9, 9, 10, 11, 12,
	3, 3, 3, 3, 2, 2, 4, 4, 5, 6,
	7,
}

var yyR2 = [...]int8{
	0, 1, 0, 2, 1, 1, 1, 3, 4, 3,
	1, 1, 1, 1, 1, 3, 4, 3, 1, 1,
	1,
}

var yyChk = [...]int16{
	-1000, -1, -8, -9, -10, -11, -12, 12, -4, 13,
	4, -3, -5, -6, -7, 5, 6, -4, 12, -2,
	-3, -4, 8, 7, 11, -2, 11, 10, -2, -3,
	11, -3, 9,
}

var yyDef = [...]int8{
	2, -2, 1, 3, 4, 5, 6, 0, 10, 0,
	18, 0, 11, 12, 13, 19, 20, 10, 0, 0,
	14, 10, 0, 0, 7, 0, 9, 0, 0, 17,
	8, 15, 16,
}

var yyTok1 = [...]int8{
	1,
}

var yyTok2 = [...]int8{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:35
		{
			if parser(yylex).hasFreeVariables() {
				yylex.Error("Facts should not have free variables")
				// TODO: how to recover or continue here if Error returns?
			}
			parser(yylex).evalFact(yyDollar[2].term.(*engine.RuleStruct))
		}
	case 8:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:44
		{
			parser(yylex).evalRule(yyDollar[1].term.(*engine.RuleStruct), yyDollar[3].terms)
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:49
		{
			parser(yylex).evalQuery(yyDollar[2].terms)
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:55
		{
			yyVAL.terms = []engine.RuleTerm{yyDollar[1].term}
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:59
		{
			yyVAL.terms = append(yyDollar[1].terms, yyDollar[3].term)
		}
	case 16:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:64
		{
			yyVAL.term = parser(yylex).makeStruct(yyDollar[1].text, yyDollar[3].terms)
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:68
		{
			yyVAL.term = parser(yylex).makeStruct(yyDollar[2].text, []engine.RuleTerm{yyDollar[1].term, yyDollar[3].term})
		}
	case 18:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:73
		{
			yyVAL.term = parser(yylex).makeAtom(yyDollar[1].text)
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:78
		{
			val, err := strconv.ParseInt(yyDollar[1].text, 10, 64)
			if err != nil {
				yylex.Error(fmt.Sprintf("Numeric overflow: %s", yyDollar[1].text))
				// TODO: how to recover or continue here if Error returns?
			}
			yyVAL.term = parser(yylex).makeNumber(val)
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:88
		{
			yyVAL.term = parser(yylex).makeVariable(yyDollar[1].text)
		}
	}
	goto yystack /* stack new state and value */
}
