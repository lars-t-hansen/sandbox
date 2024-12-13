%option noyywrap
%option yylineno

%{
#include <stdio.h>

enum {
    tEnd, tLe, tLt, tGe, tGt, tMatch, tEq, tAnd, tOr, tNot, tLparen, tRparen,
    tIdent, tString, tDuration, tNumber, 
};

%}

D [0-9]

/* Consider that it is not possible to use Go regexes to implement this directly because
   they will not look for longest match, but first match, in a disjunction.  It's possible
   that the rules can be sorted so that we always do longest match first, this basically
   means expanding the NFAs into a list of disjuncts and walking them in parallel.  It
   seems easier to just do the DFA construction?  */

%%

[ \t\r\n]                                 /* nothing */
\<=                                       return tLe ;
\<                                        return tLt ;
>=                                        return tGe ;
>                                         return tGt ;
=~                                        return tMatch ;
=                                         return tEq ;
and                                       return tAnd ;
or                                        return tOr ;
not                                       return tNot ;
\(                                        return tLparen ;
\)                                        return tRparen ;
[a-zA-Z_][a-zA-Z0-9_]*                    return tIdent ;
\"[^"]*\"                                 return tString ;
'[^']*'                                   return tString ;
\/[^/]*\/                                 return tString ;
`[^`]*`                                   return tString ;
{D}+[wW]({D}+[dD])?({D}+[hH])?({D}+[mM])? return tDuration;
{D}+[dD]({D}+[hH])?({D}+[mM])?            return tDuration;
{D}+[hH]({D}+[mM])?                       return tDuration ;
{D}+[mM]                                  return tDuration ;
[+-]?{D}+(\.{D}+)?([eE][-+]?{D}+)?        return tNumber ;
.                                         fprintf(stderr, "ERROR: %d: Bad token %s\n", yylineno, yytext) ;

%%

int main(int argc, char** argv) {
    for (int c = yylex() ; c != tEnd ; c = yylex()) {
        if (c >= tIdent) {
	    printf("%d %s\n", c, yytext);
    	} else {
            printf("%d\n", c);
    	}
    }
}
