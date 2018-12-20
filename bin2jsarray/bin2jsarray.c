/* Read binary data from stdin, write as a JS array literal */

#include <stdio.h>
#include <stdlib.h>

int main(int argc, char** argv)
{
    int c;
    if (argc > 1) {
        fprintf(stderr, "Error: The program reads stdin\n");
        exit(1);
    }
    putchar('[');
    while ((c = getchar()) != EOF) {
        printf("%d,", c);
    }
    putchar(']');
    putchar('\n');
}
