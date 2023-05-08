/* -*- fill-column: 100 -*- */
/* Cuda mandelbrot */

#include "cuda_runtime.h"

/* Several interesting complications here:

   - The `iterations` array must be allocated as Cuda host memory in order to optimize the memcpy
     (otherwise the memcpy is very slow); not a free lunch though: this memory is pinned in RAM and
     we can't have an unbounded amount

   - Host malloc is also expensive and the expense grows with the size.  It is more expensive than
     the computation in our case.  So the allocation cost must be amortized.

   - The cuda initialization takes a long time and is measured separately, and in practice it means
     that we must amortize the cost of it across many runs of computation.

   - Default tile sizes are such that tile_x*tile_y==32, the size of a warp.  Empirically this
     appears pretty much to be optimal.  The warp will only exit once all threads are done, so if
     one thread gets stuck in a deep search while the others exit we're wasting time.  But it seems
     likely (fractals...)  that no particular shape of a tile will be better than another, only
     smaller tiles could be better, and then only if scheduling is free.  But my understanding is
     that no tile can effectively be smaller than a warp without basically wasting resources. */

static unsigned tile_y = 2;
static unsigned tile_x = 16;

/* TODO: Can the host memory be mapped into the cuda address space so as to avoid the memcpy?  See
   the cudaHostAllocMapped flag to cudaHostAlloc. The code runs but the output is borked, maybe need
   some kind of sync?  Or is the code incomplete.  There is cudaHostGetDevicePointer and in the
   description of that there's wording about registering things, too. */

//#define USE_MAPPED_MEMORY_IF_POSSIBLE

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

static unsigned* iterations; // Cuda host memory, [HEIGHT * WIDTH]

#include "../mandelcommon/mandelcommon.h"

__device__ inline float_t scale(float_t v, float_t rng, float_t min, float_t max) {
  return min + v*((max-min)/rng);
}

__device__ unsigned mandel_pixel(unsigned py, unsigned px) {
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

static bool can_map_memory = false;

static void initCuda() {
  /* Get the device ID and sync to force initialization so that it doesn't pollute timings.  Getting
     the device ID is by itself not enough. */
  begin_timer();
  int dev_id;
  cudaGetDevice(&dev_id);
  unsigned flags;
  cudaGetDeviceFlags(&flags);
#ifdef USE_MAPPED_MEMORY_IF_POSSIBLE
  can_map_memory = flags & cudaDeviceMapHost;
#endif
  cudaDeviceSynchronize();
  end_timer("init");
#ifndef NDEBUG
  fprintf(stderr, "device %d\n", dev);
#endif
}

static void initHostMemory() {
  size_t nbytes = HEIGHT*WIDTH*sizeof(unsigned);
  cudaError_t err;
  begin_timer();
  if (can_map_memory) {
    if ((err = cudaHostAlloc(&iterations, nbytes, cudaHostAllocMapped)) != 0) {
      fprintf(stderr, "host alloc mapped %zu bytes %d\n", nbytes, err);
      abort();
    }
  } else {
    if ((err = cudaMallocHost(&iterations, nbytes)) != 0) {
      fprintf(stderr, "host malloc %zu bytes %d\n", nbytes, err);
      abort();
    }
  }
  end_timer("Host malloc");
}

static void mandel() {
  size_t nbytes = HEIGHT*WIDTH*sizeof(unsigned);
  unsigned *dev_iterations;
  cudaError_t err;

  begin_timer();
  if ((err = cudaMalloc(&dev_iterations, nbytes)) != 0) {
    fprintf(stderr, "device malloc %zu bytes %d\n", nbytes, err);
    abort();
  }
  end_timer("Device malloc");

  dim3 threadsPerBlock(tile_x, tile_y);
  dim3 blocksPerGrid((WIDTH+tile_x-1)/tile_x, (HEIGHT+tile_y-1)/tile_y);
  begin_timer();
  mandel_worker<<<blocksPerGrid, threadsPerBlock>>>(dev_iterations);
  cudaDeviceSynchronize();
  end_timer("Compute");

  if (!can_map_memory) {
    begin_timer();
    if ((err = cudaMemcpy(iterations, dev_iterations, nbytes, cudaMemcpyDeviceToHost)) != 0) {
      fprintf(stderr, "memcpy %d\n", err);
      abort();
    }
    cudaDeviceSynchronize();
    end_timer("Memcpy");
  } else {
    // Probably need some kind of synchronization?
  }

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

  initCuda();
  initHostMemory();
  mandel();
  dump("mandelcuda.ppm");
}
