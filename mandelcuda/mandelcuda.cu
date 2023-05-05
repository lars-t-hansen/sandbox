/* Cuda mandelbrot */

#include <stdio.h>
#include <stdlib.h>
#include <sys/time.h>
#include <inttypes.h>
#include <assert.h>
#include "cuda_runtime.h"

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

__device__ inline float_t scale(float_t v, float_t rng, float_t min, float_t max) {
  return min + v*(max-min)/rng;
}

__device__ unsigned mandel_pixel(unsigned py, unsigned px) {
  /* TODO: Overhead.  We can hoist a bunch of stuff here I think. */
  float_t y0 = scale(py, HEIGHT, MINY, MAXY);
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
  return iteration;
}

__global__ void mandel_worker(unsigned* iterations) {
  unsigned row = blockIdx.y * blockDim.y + threadIdx.y;
  unsigned col = blockIdx.x * blockDim.x + threadIdx.x;
  if (row < HEIGHT && col < WIDTH) {
    iterations[row*WIDTH + col] = mandel_pixel(row, col);
  }
}

static void mandel() {
#ifndef NDEBUG
  int dev = -87;
  cudaGetDevice(&dev);
  fprintf(stderr, "device %d\n", dev);
#endif

  size_t nbytes = sizeof(unsigned)*HEIGHT*WIDTH;
  assert(nbytes == sizeof(iterations));
  unsigned *dev_iterations;
  cudaError_t err;
  if ((err = cudaMalloc(&dev_iterations, nbytes)) != 0) {
    fprintf(stderr, "malloc %u bytes %d\n", (unsigned)nbytes, err);
    abort();
  }

  const unsigned TILEY = 4;
  const unsigned TILEX = 4;
  dim3 threadsPerBlock(TILEX, TILEY);
  dim3 blocksPerGrid((WIDTH+TILEX-1)/TILEX, (HEIGHT+TILEY-1)/TILEY);
  mandel_worker<<<blocksPerGrid, threadsPerBlock>>>(dev_iterations);

#ifndef NDEBUG
  for ( int y=0 ; y < HEIGHT; y++ ) {
    for ( int x=0 ; x < WIDTH; x++ ) {
      iterations[y*WIDTH + x] = 2;
    }
  }
#endif
  if ((err = cudaMemcpy(iterations, dev_iterations, nbytes, cudaMemcpyDeviceToHost)) != 0) {
    fprintf(stderr, "memcpy %d\n", err);
    abort();
  }
  cudaFree(dev_iterations);
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
  mandel();
  gettimeofday(&after, NULL);
  int64_t delta = ((int64_t)after.tv_sec - (int64_t)before.tv_sec)*1000000 + (after.tv_usec - before.tv_usec);
  printf("Elapsed %" PRIi64 "ms\n", delta/1000);
  dump("mandelcuda.ppm");
}
