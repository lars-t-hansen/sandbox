/* Common code for all the mandelbrot programs, to be included in each program */

#include <stdio.h>
#include <sys/time.h>
#include <inttypes.h>
#include <stdlib.h>
#include <assert.h>

/* There are several free variables:
    - unsigned CUTOFF is the cutoff for search
    - size_t HEIGHT is the number of rows in the grid
    - size_t WIDTH is the number of columns in each row
    - unsigned iterations[HEIGHT * WIDTH] is the grid of iteration counts
*/

#define RGB(r, g, b) ((r << 16) | (g << 8) | b)

/* Supposedly the gradients used by the Wikipedia mandelbrot page */
static unsigned mapping[] = {
  RGB(66, 30, 15),
  RGB(25, 7, 26),
  RGB(9, 1, 47),
  RGB(4, 4, 73),
  RGB(0, 7, 100),
  RGB(12, 44, 138),
  RGB(24, 82, 177),
  RGB(57, 125, 209),
  RGB(134, 181, 229),
  RGB(211, 236, 248),
  RGB(241, 233, 191),
  RGB(248, 201, 95),
  RGB(255, 170, 0),
  RGB(204, 128, 0),
  RGB(153, 87, 0),
  RGB(106, 52, 3),
};

static void from_rgb(unsigned rgb, unsigned* r, unsigned* g, unsigned* b) {
  *r = (rgb >> 16) & 255;
  *g = (rgb >> 8) & 255;
  *b = rgb & 255;
}

/* Dump grid to a PPM file. */
static void dump(const char* filename) {
  FILE* out = fopen(filename, "w");
  fprintf(out, "P6 %d %d 255\n", WIDTH, HEIGHT);
  unsigned y, x;
  for (y=0; y < HEIGHT; y++) {
    for ( x = 0 ; x < WIDTH; x++ ) {
      unsigned r = 0, g = 0, b = 0;
      if (iterations[y*WIDTH + x] < CUTOFF) {
	from_rgb(mapping[iterations[y*WIDTH + x] % 16], &r, &g, &b);
      }
      fputc(r, out);
      fputc(g, out);
      fputc(b, out);
    }
  }
  fclose(out);
}

static struct timeval before;

/* Start a timer.  Timers don't nest */
static void begin_timer() {
  gettimeofday(&before, NULL);
}

/* End a timer and print the elapsed time, with informative text */
static void end_timer(const char* what) {
  struct timeval after;
  gettimeofday(&after, NULL);
  int64_t delta = ((int64_t)after.tv_sec - (int64_t)before.tv_sec)*1000000 + (after.tv_usec - before.tv_usec);
  printf("%s: Elapsed %" PRIi64 "ms\n", what, delta/1000);
}

