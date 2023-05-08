/* -*- fill-column: 100 -*- */
/* Cuda mandelbrot */

#include "cuda_runtime.h"

/* Default values are such that tile_x*tile_y==32, the size of a warp.  The warp will only exit once
   all threads are done, so if one thread gets stuck in a deep search while the others exit we're
   wasting time. */

static unsigned tile_y = 2;
static unsigned tile_x = 16;

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

static unsigned iterations[HEIGHT * WIDTH];

#include "../mandelcommon/mandelcommon.h"

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

  size_t nbytes = sizeof(iterations);
  unsigned *dev_iterations;
  cudaError_t err;

  /* Just sync to force initialization so that it doesn't pollute timings */
  begin_timer();
  cudaDeviceSynchronize();
  end_timer("init");

  begin_timer();
  if ((err = cudaMalloc(&dev_iterations, nbytes)) != 0) {
    fprintf(stderr, "malloc %zu bytes %d\n", nbytes, err);
    abort();
  }
  end_timer("Malloc");

  dim3 threadsPerBlock(tile_x, tile_y);
  dim3 blocksPerGrid((WIDTH+tile_x-1)/tile_x, (HEIGHT+tile_y-1)/tile_y);
  begin_timer();
  mandel_worker<<<blocksPerGrid, threadsPerBlock>>>(dev_iterations);
  cudaDeviceSynchronize();
  end_timer("Compute");

  begin_timer();
  if ((err = cudaMemcpy(iterations, dev_iterations, nbytes, cudaMemcpyDeviceToHost)) != 0) {
    fprintf(stderr, "memcpy %d\n", err);
    abort();
  }
  end_timer("Memcpy");

  begin_timer();
  cudaFree(dev_iterations);
  end_timer("Free");
}

int main(int argc, char** argv) {
  for (int i=1 ; i < argc; i++ ) {
    if (sscanf(argv[i], "-y%u", &tile_y) == 1) {
      if (tile_y == 0) {
	fprintf(stderr, "Zero rows\n");
	exit(1);
      }
      continue;
    }
    if (sscanf(argv[i], "-x%u", &tile_x) == 1) {
      if (tile_x == 0) {
	fprintf(stderr, "Zero columns\n");
	exit(1);
      }
      continue;
    }
    fprintf(stderr, "Bad option %s\n", argv[1]);
  }

  mandel();
  dump("mandelcuda.ppm");
}
