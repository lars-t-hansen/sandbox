/* An 8080 assembler that is so simple that it can be hand-translated to asm.
 *
 * Syntax:
 *
 * All lower case mandated
 * All 7-bit ascii
 * No labels on lines by themselves
 * Whitespace allowed in obvious places
 * Hex and decimal numbers
 * Comments both on line and trailing
 * Five significant chars in names
 * No expressions in operands (but names stand for their values)
 * All syntax errors are immediately fatal
 *
 *   input : lines ;
 *   lines : line eol lines
 *         | line eof
 *         ;
 *   line  : blank
 *         | instr
 *         ;
 *   blank : maybe-spaces maybe-comment ;
 *   instr : label maybe-spaces instr maybe-spaces maybe-comment
 *         | maybe-spaces instr maybe-comment
 *         ;
 *   label : name ':' ;
 *   instr : keyword
 *         | keyword maybe-spaces operand
 *         ;
 *   maybe-spaces : | space maybe-spaces ;
 *   maybe-comment : | ';' anything ;
 *   name : letter-or-digit | letter-or-digit name ;
 *   eol   : <newline> ;
 *   eof   : <nul> ;
 *   anything : <any char except newline or nul> ;
 */

/* The input is the entire input file.  Every line is terminated by \n, end of input by 0x00. */
/* Labels and names are max five chars.  There's no output overflow check. */

#define BUFSIZE (1024*16)       /* 2*16K = 32K */
#define NSIZE 5                 /* Length of name, you can't change this easily */
#define NAMES 1024              /* 7K */

struct name_t {
    char name[NSIZE];
    uint16 value;
};

char input[SIZE];
unsigned output[SIZE];
struct name_t names[NAMES];

char w[NSIZE];                  /* current word */
char lbl[NSIZE];                /* label on this line if first char is not ' ' */
char* w2;                       /* word to compare to */
char* p;                        /* input pointer */
char* q;                        /* output pointer */
const char* msg;                /* error message */
int pass;                       /* pass number */
int namex;                      /* index in names table of next free */

// All of these are 6 bytes: five bytes of opcode name followed by a single opcode byte,
// to be emitted literally, followed by NUL.
const unsigned char simple[] =
    "cma  \x2F"
    "cmc  \x3F"
    "stc  \x37";

void scan();

/* Read from stdin, write to stdout */
int main() {
    ssize_t n;
    n = read(0, input, SIZE);
    if (n == SIZE) {
        msg = "too much input";
        fail();
    }

    input[n] = 0;
    pass = 1;
    p = input;
    q = output;
    lno = 0;
    scan();

    pass = 2;
    p = input;
    q = output;
    lno = 0;
    scan();

    if (q-output > SIZE) {
        msg = "too much output";
        fail();
    }
    write(1, output, q-output);
    return 0;
}

void scan() {
    /* new line */
line:
    lno++;
    lclr();
    wclr();
    if (word()) {
        if (*p == ':') {
            p++;
            goto lbld;
        }
        goto instr;
    }
    if (*p == 0) {
        goto eof;
    }
    goto eol;

lbld:
    lcpy();
    wclr();
    // TODO: do something to define the label!
    if (!word()) {
        msg = "want instruction";
        fail();
    }
    /* fallthrough to instr */

instr:
    // w has the opcode word

    // Opcode-only instruction
    w2 = simple;
    while (*w2) {
        flag=wcomp();
        if (flag) {
            *out++ = w2[NSIZE];
            goto end;
        }
        w2 += NSIZE+1;
    }

    // Single register instructions
    w2="inr  ";
    flag=wcomp();
    if (flag) {
        op=0x04;
        shift=3;
        goto rm;
    }
    w2="dcr  ";
    flag=wcomp();
    if (flag) {
        op = 0x05;
        shift=3;
        got rm;
    }

    // Many more

    msg = "Unknown instruction";
    goto fail;

rm:
    // Register-or-memory operand
    while (*p == ' ' || *p == '\t') {
        p++;
    }
    // FIXME: look for b,c,d,e,h,l,m,a -> code in r
    *out++ = op | (r << shift);
    goto end;

end:
    // TODO: check that there's nothing left over
    return p;

estax:
    // syntax error on this line
    fail("Syntax error");
}

int wclr() {
    w[0] = ' ';
    w[1] = ' ';
    w[2] = ' ';
    w[3] = ' ';
    w[4] = ' ';
}

void lcpy() {
    lbl[0] = w[0];
    lbl[1] = w[1];
    lbl[2] = w[2];
    lbl[3] = w[3];
    lbl[4] = w[4];
}

int lclr() {
    w[0] = ' ';
    w[1] = ' ';
    w[2] = ' ';
    w[3] = ' ';
    w[4] = ' ';
}

int wcomp() {
    if (w[0] != w2[0]) return 0;
    if (w[1] != w2[1]) return 0;
    if (w[2] != w2[2]) return 0;
    if (w[3] != w2[3]) return 0;
    if (w[4] != w2[4]) return 0;
    return 1;
}

/* Fill w with the word and return 1 if a word is found, otherwise 0.  Consumes spaces and comments
 * before the word but not eol or eof.  Fails on unknown input.  When it returns 0, the next char is
 * either eof or eol.
 */
int word() {
    for (;;) {
        char c = *p;
        if (c == 0 || c == '\n') {
            return 0;
        }
        if (c == ' ' || c == '\t') {
            p++;
            continue;
        }
        if (c == ';') {
            p++;
            for (;;) {
                c = *p;
                if (c == 0 || c == '\n') {
                    return 0;
                }
                p++;
            }
        }

        if (c >= 'a' && c <= 'z') {
            p++;
            char* r = w;
            char* l = w+NSIZE;
            *r++ = c;
            for (;;) {
                c = *p;
                if (!(c >= 'a' && c <= 'z' || c >= '0' && c <= '9')) {
                    return 1;
                }
                p++;
                if (r < l) {
                    *r++ = c;
                }
            }
        }

        msg = "unknown character";
        fail();
    }
}

void fail() {
    write(2, msg, strlen(msg));
    exit(1);
}
