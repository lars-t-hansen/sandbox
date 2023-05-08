/* Sequential mandelbrot */

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

static unsigned iterations[HEIGHT * WIDTH];

#include "../mandelcommon/mandelcommon.h"

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

int main(int argc, char** argv) {
  begin_timer();
  mandel_slice(0, HEIGHT, 0, WIDTH);
  end_timer("Compute");
  dump("mandelbrot.ppm");
}
