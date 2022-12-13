/* This came up on a HN thread.  Suppose you have an alphabet a..z,
   and input string I of length S and an integer D, D <= S.  Find the
   first index in I of a length D substring with all different
   characters.

   The running time of this is O(S).  Every step in the outer loop
   advances by one character.  The inner loop advances unpredictably
   but never examines any input char more than once.
*/

/* Usage: window S D */

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char** argv) {
  if (argc != 3) {
    abort();
  }
  const char* I = argv[1];
  int D = atoi(argv[2]);
  int S = strlen(I);
  if (D >= S) {
    abort();
  }

  int first_c = 0;		/* start of current substring */
  int next_c = 0;		/* current candidate char */
  uint32_t cs = 0;		/* the set of chars in the window */
  int k = 0;			/* number of chars in the set */
  for (;;) {
    char ch = I[next_c++];		    /* the candidate */
    uint32_t v = (uint32_t)1 << (ch - 'a'); /* the set element for the candidate */
    if ((cs & v) == 0) {		    /* if not in set ... */
      cs |= v;				    /* ... then add it */
      k++;				    /* ... and account for it */
      if (k == D) {			    /* ... and if we have D, we're done at first_c */
	goto success;
      }
    } else {			/* ch is in the set already */
      char other;
      while ((other = I[first_c++]) != ch) { /* scan until previous occurrence of ch */
	uint32_t v = (uint32_t)1 << (other - 'a');
	cs ^= v;		/* Remove from set */
	k--;			/* and account for it */
      }
      if (first_c > S-D) {
	goto failure;
      }
    }
  }

 success:
  printf("Found it at %d\n", first_c);
  return 0;

 failure:
  printf("Did not find it\n");
  return 0;
}
