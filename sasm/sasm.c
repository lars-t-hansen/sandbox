// suppose we want a maximally simple assembler for a subset of z80 (a simple case).  we could
// simplify further by using 8080... but not yet.  this would have syntax that is very restricted:
//
// all upper (or lower) case mandated
// all 7-bit ascii
// hex and decimal as usual (or things become hard)
// comments both on line and trailing
// no labels on lines by themselves
// short names (maybe just two chars and digits)
// opcode separated from operands by exactly one tab
// spaces not allowed (or maybe allowed only after comma)
// no expressions in operands
// all syntax errors are fatal
// syntax that is not LL(0) should be clarified
//
// in the following the eol is implied
//
//   line  : blank | instr ;
//   blank : ' ' blank | '\t' blank | ';' junk;
//   junk : <list of chars that are junk> ;
//   instr : label instr | indented-instr
//   label : name ':' ;
//   name : letter-or-digit | letter-or-digit letter-or-digit ;
//   indented-instr : '\t' instr | instr ;
//   instr : "ADC\t" A ',' n eol
//         | "ADD\t"
//         | "CALL"
//         | "LD"
//         | ...
//         ;
//
// Eventually this assembler will be written in its own language so no cheating with regexes.  All values are 8
// or 16 bits, signed or unsigned.
//
// For a lot of instructions we could define our own opcodes, eg ADD A, ... is ADDA while ADD HL,
// ... is ADDHL, AND (HL) could be ANDI.  We ignore IX, IY, this changes a lot.  ADD A, r could be
// ADDAR r while ADD A, n could be ADDAN.
//
// Probably 8080 is exactly what we need: execpt for MOV, everything has at most one operand and there
// is at most one opcode byte.
//
// use 'r' for 8-bit regs, 'w' for 16-bit regs.

// Every line is terminated by *, end of input also.  If nothing is found, nothing is output.
void scan_line(char* p) {
    char c = *p++;
    char word[6];
    char label[2];
newword:
    word[0] = ' ';
    word[1] = ' ';
    word[2] = ' ';
    word[3] = ' ';
    word[4] = ' ';
    word[5] = ' ';
again:
    if (c == '*' || c == ' ' || c == '\t') {
        goto again;
    }
    if (c == ';') {
        while (*p != '*') {
            p++;
        }
        goto again;
    }
    if (c >= 'a' && c <= '9') {
        word[0] = c;
        int i = 1;
        while (*p >= 'a' && *p <= 'z' || *p >= '0' && *p <= '9') {
            if (i < 6) {
                word[i++] = *p;
            }
            p++;
        }
        if (*p == ':') {
            p++;
            goto labeled;
        }
        goto instruction;
      default:
        fail();
        return;
    }
labeled:
    label[0] = word[1];
    if (word[0] > 1) {
        label[1] = word[2];
    }
    // do something to define it
    goto newword;
instruction:
    if (same(word, "adc   ")) {
    }
    if (same(word, "addr  ")) {
        // ADD A, r
    }
    if (same(word, "addn  ")) {
    }
    if (same(word, "bit   ")) {
    }
    // and so on
}
