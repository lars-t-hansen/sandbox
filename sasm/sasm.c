// suppose we want a maximally simple assembler for a subset of z80 (a simple case).  we should
// simplify further by using 8080 assembly, which has a much simpler syntax and instruction layout.
// execpt for MOV, everything has at most one operand and there is at most one opcode byte.
//
// Meaningful restrictions:
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

/* The input is the entire input file.  Every line is terminated by \n, end of input by 0x00. */
/* Labels and names are max five chars.  There's no output overflow check. */

char w[6];                      /* current word */
char lbl[2];                    /* label on this line if first char is not ' ' */
char* w2;                       /* word to compare to */
char* p;                        /* input pointer */
char* q;                        /* output pointer */

void scan() {
    char c = *p++;
neww:
    w[0] = ' ';
    w[1] = ' ';
    w[2] = ' ';
    w[3] = ' ';
    w[4] = ' ';
    w[5] = ' ';
again:
    if (c == ' ' || c == '\t') {
        goto again;
    }
    if (c == ';') {
        while (*p != '*') {
            p++;
        }
        goto again;
    }
    if (c >= 'a' && c <= '9') {
        w[0] = c;
        int i = 1;
        while (*p >= 'a' && *p <= 'z' || *p >= '0' && *p <= '9') {
            if (i < 6) {
                w[i++] = *p;
            }
            p++;
        }
        if (*p == ':') {
            p++;
            goto lbld;
        }
        goto instr;
      default:
        fail();
        return;
    }

lbld:
    label[0] = w[1];
    if (w[0] > 1) {
        label[1] = w[2];
    }
    // do something to define it
    goto neww;

instr:
    // Opcode-only instructions
    w2 = simple;
    while (*w2) {
        flag=comp();
        if (flag) {
            *out++ = w2[5];
            goto end;
        }
        w2 += 6;
    }

    // Single register instructions
    w2="inr  ";
    flag=comp();
    if (flag) {
        // possibly r_or_memory takes the bit pattern and this is a goto, b/c there's only one-byte
        // operations?  but what about the shift amount?
        op = r_or_memory(&p);
        *out++ = 0x04 | (op << 3);
        goto end;
    }
    w2="dcr  ";
    flag=comp();
    if (flag) {
        op = rm(&p);
        *out++ = 0x05 | (op << 3);
        goto end;
    }

end:
    // TODO: check that there's nothing left over
    return p;
}

// All of these are 6 bytes: five bytes of opcode name followed by a single opcode byte,
// to be emitted literally, followed by NUL.
const unsigned char simple[] =
    "cma  \x2F"
    "cmc  \x3F"
    "stc  \x37";

// operand scanners also scan any space, a single tab is required

// second operand is 8-bit Register or Memory: b,c,d,e,h,l,m,a
// with m meaning (hl).
int rm() {
    if (*p != '\t') {
        fail();
    }
    p++;
}

int comp() {
    if (w[0] != w2[0]) return 0;
    if (w[1] != w2[1]) return 0;
    if (w[2] != w2[2]) return 0;
    if (w[3] != w2[3]) return 0;
    if (w[4] != w2[4]) return 0;
    return 1;
}
