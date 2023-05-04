/* Sequential mandelbrot */

#include <stdio.h>
#include <sys/time.h>
#include <inttypes.h>

/* Canvas size in pixels */
#define WIDTH 1400
#define HEIGHT 800

/* Classic mandelbrot set */
typedef float float_t;
static const unsigned CUTOFF = 3000;
static const float_t MINY = -1;
static const float_t MAXY = 1;
static const float_t MINX = -2.5;
static const float_t MAXX = 1;

/* Off the web, a little different */
/*
typedef double float_t;
static const unsigned CUTOFF = 10000;
static const float_t MINY = -0.6065922085831237;
static const float_t MAXY = -0.606486596104741;
static const float_t MINX = -0.34853774148008254;
static const float_t MAXX = -0.34831493420245574;
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

static unsigned iterations[HEIGHT * WIDTH];

static void from_rgb(unsigned rgb, unsigned* r, unsigned* g, unsigned* b) {
  *r = (rgb >> 16) & 255;
  *g = (rgb >> 8) & 255;
  *b = rgb & 255;
}

static inline float_t scale(float_t v, float_t rng, float_t min, float_t max) {
  return min + v*(max-min)/rng;
}

static void mandel_slice(unsigned start_y, unsigned lim_y, unsigned start_x, unsigned lim_x) {
  unsigned py;
  for (py = start_y; py < lim_y; py++) {
    float_t y0 = scale(py, HEIGHT, MINY, MAXY);
    unsigned px;
    for (px = start_x; px < lim_x; px++) {
      float_t x0 = scale(px, WIDTH, MINX, MAXX);
      float_t x = 0, y = 0;
      unsigned iteration = 0;
      while (x*x+y*y <= 4 && iteration < CUTOFF) {
	float_t nx = x*x - y*y + x0;
	float_t ny = 2*x*y + y0;
	x = nx;
	y = ny;
	iteration++;
      }
      iterations[py*WIDTH + px] = iteration;
    }
  }
}

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

int main(int argc, char** argv) {
  struct timeval before, after;
  gettimeofday(&before, NULL);
  mandel_slice(0, HEIGHT, 0, WIDTH);
  gettimeofday(&after, NULL);
  int64_t delta = ((int64_t)after.tv_sec - (int64_t)before.tv_sec)*1000000 + (after.tv_usec - before.tv_usec);
  printf("Elapsed %" PRIi64 "ms\n", delta/1000);
  dump("mandelbrot.ppm");
}
