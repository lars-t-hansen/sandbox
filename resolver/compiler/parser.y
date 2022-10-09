%{
package compiler

import (
	"fmt"
	"resolver/engine"
	"strconv"
)
%}

%union { 
    text string
    terms []engine.RuleTerm
    term engine.RuleTerm
}

%start Program

%token <text> T_ATOM T_NUMBER T_VARNAME
%left <text> T_INFIX_OP
%token T_LPAREN T_RPAREN T_COMMA T_PERIOD T_FACT_OP
%left T_QUERY_OP

%type <terms> Terms
%type <term> Term Struct Atom Number Variable

%%

Program : Phrases ;
Phrases : | Phrases Phrase ;
Phrase  : Fact | Rule | Query ;
Fact    : T_FACT_OP Struct T_PERIOD
            {
				if parser(yylex).hasFreeVariables() {
					yylex.Error("Facts should not have free variables")
					// TODO: how to recover or continue here if Error returns?
				}
	            parser(yylex).evalFact($2.(*engine.RuleStruct))
            }
        ;
Rule    : Struct T_FACT_OP Terms T_PERIOD
            {
                parser(yylex).evalRule($1.(*engine.RuleStruct), $3)
            }
        ;
Query   : T_QUERY_OP Terms T_PERIOD
            {
                parser(yylex).evalQuery($2)
            }
        ;
Term    : Struct | Atom | Number | Variable ;
Terms   : Term
            {
                $$ = []engine.RuleTerm{$1}
            }
        | Terms T_COMMA Term
            {
                $$ = append($1, $3)
            }
        ;
Struct  : T_ATOM T_LPAREN Terms T_RPAREN
            {
                $$ = parser(yylex).makeStruct($1, $3)
            }
        | Term T_INFIX_OP Term
            {
                $$ = parser(yylex).makeStruct($2, []engine.RuleTerm{$1, $3})
            }
        ;
Atom    : T_ATOM
            {
                $$ = parser(yylex).makeAtom($1)
            }
        ;
Number  : T_NUMBER
            {
                val, err := strconv.ParseInt($1, 10, 64)
				if err != nil {
					yylex.Error(fmt.Sprintf("Numeric overflow: %s", $1))
					// TODO: how to recover or continue here if Error returns?
				}
                $$ = parser(yylex).makeNumber(val)
            }
        ;
Variable : T_VARNAME
            {
				$$ = parser(yylex).makeVariable($1)
            }
        ;

%%

/////////////////////////////////////////////////////////////////////////////////////////////////
//
// Parser interface

// Sticking parser context on yylex is how the cool kids do it, but it's really not pretty.

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

	writeString func(s string)
}

func newParser(st *engine.Store, writeString func(string)) *parserctx {
	return &parserctx{
		st: st,
		varIndex: 0, 
		vars: make([]*engine.Local, 0),
		nameMap: make(map[string]int, 0),
		writeString: writeString,
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
	for i, n := range names {
		if n == nil {
			names[i] = p.st.NewAtom("_")
		}
	}
	p.getAndClearVars()
	p.st.EvaluateQuery(query, names, p.writeString)
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

func Repl(st *engine.Store, r reader, writeString func(string)) {
	ctx := newParser(st, writeString)
	t := newTokenizer(r, ctx)
	if yyParse(t) != 0 {
		panic("Parse failed")
	}
}