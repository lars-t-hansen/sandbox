/* An 8080 assembler that is so simple that it can be hand-translated to asm.
 *
 * Syntax:
 *
 * All upper case mandated for keywords (but lower case ok in numbers)
 * All 7-bit ascii
 * No labels on lines by themselves
 * Whitespace allowed in obvious places
 * Hex, decimal, octal, binary numbers
 * Comments both on line and trailing
 * Five significant chars in names
 * x+y expressions in operands
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

#include <assert.h>
#include <stdio.h>  /* printf, for now */
#include <string.h> /* strlen */
#include <stdlib.h> /* exit */
#include <fcntl.h>  /* open */
#include <unistd.h> /* read, write, close */

/* The input is the entire input file.  Every line is terminated by \n, end of input by 0x00. */
/* Labels and names are max five chars.  There's no output overflow check. */

#define BUFSIZE (1024*16)       /* 2*16K = 32K */
#define NSIZE   5               /* Length of name, you can't change this easily */
#define NAMES   1024            /* 7K, if we want */
#define HERE    ((q-output)+org) /* current program location */

#define byte unsigned char
#define word unsigned short

struct name_t {
    char name[NSIZE];
    word value;
};

char           input[BUFSIZE];  /* input text, last two are <newline><nul> */
byte           output[BUFSIZE]; /* output bytes */
struct name_t  names[NAMES];    /* defined names, in no order */
const char*    p;               /* input pointer */
byte*          q;               /* output pointer */
struct name_t* namex;           /* pointer past last name */
int            pass;            /* pass number */
int            lno;             /* line number */
word           org;             /* origin */

void scan();
int  line();
void cpy(char dst[NSIZE], const char src[NSIZE]);
int  same(const char a[NSIZE], const char b[NSIZE]);
void clr(char w[NSIZE]);
int  wrd(char w[NSIZE]);
void fail(const char* msg);
void set(const char w[NSIZE], word value);
int  lookup(const char w[NSIZE], word* v);
void spc();
void def(const char* name, word value);
word value();
word val(const char w[NSIZE]);
word num();
byte bval();
word wcval();
int  op(char);
int  xnum(const char w[NSIZE], word* v);

struct simple_t {
    char name[NSIZE];
    byte op;
};

const struct simple_t simple[] = {
    {"CMA  ", 0x2F},
    {"CMC  ", 0x3F},
    {"DAA  ", 0x27},
    {"NOP  ", 0x00},
    {"RAL  ", 0x17},
    {"RAR  ", 0x1F},
    {"RLC  ", 0x07},
    {"RRC  ", 0x0F},
    {"STC  ", 0x37},
    {"*    ", 0x00},
};

struct rm8_t {
    char name[NSIZE];
    byte op;
    int  shift;
};

const struct rm8_t rm8[] = {
    {"ADC  ", 0x88, 0},
    {"ADD  ", 0x80, 0},
    {"ANA  ", 0xA0, 0},
    {"CMP  ", 0xB8, 0},
    {"DCR  ", 0x05, 3},
    {"INR  ", 0x04, 3},
    {"ORA  ", 0xB0, 0},
    {"SBB  ", 0x98, 0},
    {"SUB  ", 0x90, 0},
    {"XRA  ", 0xA8, 0},
    {"*    ", 0x00, 0},
};


int main(int argc, char** argv) {
    if (argc != 3) {
        fail("Usage");
    }
    int ind = open(argv[1], O_RDONLY);
    if (ind == -1) {
        fail("Could not open input");
    }
    ssize_t n;
    n = read(ind, input, BUFSIZE);
    close(ind);
    if (n > BUFSIZE-2) {
        fail("too much input");
    }
    /* Set up the input so that a line is always followed by \n and the eof (nul) is only ever at
     * the start of a line.  That way we only need check for eof at the beginning of a line in the
     * scanner.
     */
    input[n] = '\n';
    input[n+1] = 0;

    namex = names;

    printf("Pass 1\n");
    pass = 1;
    scan();

    printf("Pass 2\n");
    pass = 2;
    scan();

    if (q-output > BUFSIZE) {
        fail("too much output");
    }
    int outd = open(argv[2], O_WRONLY|O_CREAT, 0666);
    if (outd == -1) {
        fail("could not create output");
    }
    write(outd, output, q-output);
    close(outd);
    return 0;
}

void scan() {
    p = input;
    q = output;
    lno = 0;
    while (*p) {
        printf("Scanning at %p\n", p);
        if (line()) {
            break;
        }
    }
}

int line() {
    char w[NSIZE];  /* current word */
    int done = 0;

    lno++;

    /* Start of line */
    if (!wrd(w)) {
        goto Leol;
    }

    /* Possible label in w */
    if (*p == ':') {
        p++;
        def(w, HERE);
        if (!wrd(w)) {
            fail("want instruction or directive");
        }
        goto Linst;
    }

    /* Non-label word in w, handle directive for label-without-colon */
    {
        const char *prev = p;
        char tmp[NSIZE];
        cpy(tmp, w);
        if (wrd(w)) {
            if (same(w, "EQU  ")) {
                def(tmp, value());
                goto Leol;
            }
            if (same(w, "SET  ")) {
                set(tmp, value());
                goto Leol;
            }
        }
        /* Rollback, note wrd() can move p */
        p = prev;
        cpy(w, tmp);
    }

    /* Word in w, lbl clear */
    goto Linst;

Linst:
    /* State: w has an opcode word, lbl garbage */

    /* Directives */
    if (same(w, "DB   ")) {
        *q++ = bval();
        goto Leol;
    }
    if (same(w, "DS   ")) {
        word s = value();
        while (s--) {
            *q++ = 0x00;
        }
        goto Leol;
    }
    if (same(w, "DW   ")) {
        word v = value();
        *q++ = v & 255;
        *q++ = v >> 8;
        goto Leol;
    }
    if (same(w, "END  ")) {
        done = 1;
        goto Leol;
    }
    if (same(w, "ORG  ")) {
        word v = wcval();
        if (q == output) {
            org = v;
        } else {
            while (HERE < v) {
                *q++ = 0x00;
            }
        }
        goto Leol;
    }

    /* Opcode-only instructions */
    for (const struct simple_t *s = simple ; s->name[0] != '*' ; s++ ) {
        if (same(w, s->name)) {
            *q++ = s->op;
            goto Leol;
        }
    }

    /* 8-bit register-or-memory instructions */
    for (const struct rm8_t *i = rm8 ; i->name[0] != '*' ; i++ ) {
        if (same(w, i->name)) {
            if (!wrd(w)) {
                fail("Expected operand");
            }
            byte r;
            if (same(w, "B    ")) {
                r = 0;
            } else if (same(w, "C    ")) {
                r = 1;
            } else if (same(w, "D    ")) {
                r = 2;
            } else if (same(w, "E    ")) {
                r = 3;
            } else if (same(w, "H    ")) {
                r = 4;
            } else if (same(w, "L    ")) {
                r = 5;
            } else if (same(w, "M    ")) {
                r = 6;
            } else if (same(w, "A    ")) {
                r = 7;
            } else {
                fail("Bad operand");
            }
            *q++ = i->op | (r << i->shift);
            goto Leol;
        }
    }

    /* Load/store indirect */
    {
        byte op;
        if (same(w, "STAX ")) {
            op = 0x02;
        } else if (same(w, "LDAX ")) {
            op = 0x0A;
        } else {
            goto Lnxfr;
        }
        if (!wrd(w)) {
            fail("Expected operand");
        }
        if (same(w, "B    ")) {
            ;
        } else if (same(w, "D    ")) {
            op |= 0x10;
        } else {
            fail("Bad operand");
        }
        *q++ = op;
    Lnxfr:
        ;
    }

    /* Special */
    if (same(w, "MOV  ")) {
        // TODO: special
    }

    // TODO: Many more

    fail("Unknown instruction");

Leol:
    spc();
    if (*p != '\n') {
        fail("Junk at the end of the line");
    }
    p++;
    return done;
}

void dset(const char name[NSIZE], word value, int isdef) {
    for ( struct name_t *n = names ; n < namex ; n++ ) {
        if (same(name, n->name)) {
            if (isdef) {
                if (pass == 1) {
                    fail("Second definition of name");
                } else if (n->value != value) {
                    fail("Redefining name with different value");
                }
            }
            n->value = value;
            return;
        }
    }
    cpy(namex->name, name);
    namex->value = value;
    namex++;
}

void def(const char name[NSIZE], word value) {
    dset(name, value, 1);
}

void set(const char name[NSIZE], word value) {
    dset(name, value, 0);
}

int lookup(const char name[NSIZE], word* v) {
    for ( struct name_t *n = names ; n < namex ; n++ ) {
        if (same(name, n->name)) {
            *v = n->value;
            return 1;
        }
    }
    *v = 0;
    return 0;
}

word value() {
    char w[NSIZE];
    if (!wrd(w)) {
        fail("Operand expected");
    }
    word v = val(w);
    if (op('+')) {
        if (!wrd(w)) {
            fail("Operand expected");
        }
        v += val(w);
    }
    return v;
}

byte bval() {
    word v = value();
    if ((v >> 8) == 0 || (v >> 8) == 255) {
        return v & 255;
    }
    fail("Out of range");
    return 0;
}

word val(const char w[NSIZE]) {
    word v;
    if (xnum(w, &v)) {
        return v;
    }
    if (lookup(w, &v)) {
        return v;
    }
    fail("Not a value");
    return 0;
}

word num() {
    word v;
    char w[NSIZE];
    if (!wrd(w) || !xnum(w, &v)) {
        fail("Number expected");
    }
    return v;
}

word wcval() {
    /* TODO: really "constant value", which is more elaborate */
    return num();
}

/* TODO: Must handle quoted chars */
int xnum(const char w[NSIZE], word* v) {
    if (!(w[0] >= '0' && w[0] <= '9')) {
        return 0;
    }
    int i=NSIZE-1;
    while (w[i] == ' ') {
        i--;
    }
    word tmp = 0;
    if (w[i] == 'h' || w[i] == 'H') {
        for (int j=0 ; j < i; j++ ) {
            char c = w[j];
            tmp *= 16;
            if (c >= '0' && c <= '9') {
                tmp += c - '0';
            } else if (c >= 'A' && c <= 'F') {
                tmp += c - 'A' + 10;
            } else if (c >= 'a' && c <= 'f') {
                tmp += c - 'a' + 10;
            } else {
                return 0;
            }
        }
        *v = tmp;
        return 1;
    }
    if (w[i] == 'b' || w[i] == 'B') {
        for (int j=0 ; j < i; j++ ) {
            char c = w[j];
            tmp *= 2;
            if (c = '0' || c == '1') {
                tmp += c - '0';
            } else {
                return 0;
            }
        }
        *v = tmp;
        return 1;
    }
    if (w[i] == 'o' || w[i] == 'O' || w[i] == 'q' || w[i] == 'Q') {
        for (int j=0 ; j < i; j++ ) {
            char c = w[j];
            tmp *= 8;
            if (c >= '0' && c <= '7') {
                tmp += c - '0';
            } else {
                return 0;
            }
        }
        *v = tmp;
        return 1;
    }
    for (int j=0 ; j <= i; j++ ) {
        char c = w[j];
        tmp *= 10;
        if (c >= '0' && c <= '9') {
            tmp += c - '0';
        } else {
            return 0;
        }
    }
    *v = tmp;
    return 1;
}

int op(char x) {
    spc();
    if (*p == x) {
        p++;
        return 1;
    }
    return 0;
}

/* Fill w with an alphanumeric word and return 1 if a word is found, otherwise 0.  Consumes spaces
 * and comments before the word but not eol or eof.  Fails on unknown input.  When it returns 0, the
 * next char is either eof or eol.
 */
int wrd(char w[NSIZE]) {
    clr(w);
    spc();
    char* r = w;
    char* l = w+NSIZE;
    for (;;) {
        char c = *p;
        if (!(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9')) {
            break;
        }
        if (r < l) {
            *r++ = c;
        }
        p++;
    }
    return r > w;
}

void spc() {
    while (*p == ' ' || *p == '\t') {
        p++;
    }
    if (*p == ';') {
        p++;
        while (*p != '\n') {
            p++;
        }
    }
}

void clr(char x[NSIZE]) {
    assert(NSIZE == 5);
    x[0] = ' ';
    x[1] = ' ';
    x[2] = ' ';
    x[3] = ' ';
    x[4] = ' ';
}

void cpy(char dst[NSIZE], const char src[NSIZE]) {
    assert(NSIZE == 5);
    dst[0] = src[0];
    dst[1] = src[1];
    dst[2] = src[2];
    dst[3] = src[3];
    dst[4] = src[4];
}

int same(const char a[NSIZE], const char b[NSIZE]) {
    assert(NSIZE == 5);
    if (a[0] != b[0]) return 0;
    if (a[1] != b[1]) return 0;
    if (a[2] != b[2]) return 0;
    if (a[3] != b[3]) return 0;
    if (a[4] != b[4]) return 0;
    return 1;
}

void fail(const char* msg) {
    write(2, msg, strlen(msg));
    exit(1);
}
