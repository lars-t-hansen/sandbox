/* Pthreads + SIMD mandelbrot */
/* Usage: mandelsimd [-jN] */

/* This creates a thread pool and a shared work queue and farms out work.  The workers write
   directly into the result array and signal completion to the coordinator.  Each worker uses
   SIMD operations.  */

/* For SIMD128 and SIMD256, float_t must be float */
/* For SIMD128 we could also support ARM64 but for now only INTEL */
/*#define SIMD128*/
#define SIMD256
#define INTEL
#define DEFAULT_THREADS 4

#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <sys/time.h>
#include <inttypes.h>
#include <assert.h>
#if defined(SIMD128) && defined(INTEL)
# include <emmintrin.h>
# include <smmintrin.h>
# include <xmmintrin.h>
#endif
#if defined(SIMD256) && defined(INTEL)
# include <emmintrin.h>
# include <immintrin.h>
# include <smmintrin.h>
# include <xmmintrin.h>
#endif

static unsigned num_threads = DEFAULT_THREADS;

/* For SIMD128 (b/c I'm lazy):
   The canvas size must be divisible by 4 in the x dimension
   The tile size must be divisible by 4 in the x dimension
   However the tile size need not divide the canvas size in the x dimension

   For SIMD256, ditto but 8 instead of 4
*/

/* Canvas size in pixels */
#define WIDTH 1400
#define HEIGHT 800

/* Size of work item tiles along each dimension.  Cache contention should not be a big deal on this
   program but a 32-wide slice (with a four-byte item, for 128 bytes per tile along X) is at least
   friendly.  To do better we would need to know the line size of the cache.  Most likely, at this
   tile size, it doesn't matter at all.  In fact, work items that are too small will lead to too
   much contention.
 */
static const unsigned TILEX = 32;
static const unsigned TILEY = 32;

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

static unsigned iterations[HEIGHT][WIDTH];

static void from_rgb(unsigned rgb, unsigned* r, unsigned* g, unsigned* b) {
  *r = (rgb >> 16) & 255;
  *g = (rgb >> 8) & 255;
  *b = rgb & 255;
}

static inline float_t scale(float_t v, float_t rng, float_t min, float_t max) {
  return min + v*(max-min)/rng;
}

static void mandel_slice(unsigned start_y, unsigned lim_y, unsigned start_x, unsigned lim_x) {
#ifndef NDEBUG
  printf("Work %u %u %u %u\n", start_y, lim_y, start_x, lim_x);
#endif

#if defined(SIMD128)
# ifdef INTEL
  /* These abstractions are roughly from WASM SIMD although we have separate int and float types here. */
  typedef __m128i i128_t;
  typedef __m128 f128_t;

#  define f32x4_splat(x) _mm_set_ps1(x)
#  define f32x4_make(a, b, c, d) _mm_set_ps(d, c, b, a)
#  define f32x4_const(a, b, c, d) _mm_set_ps(d, c, b, a)
#  define f32x4_mul(x, y) _mm_mul_ps(x, y)
#  define f32x4_add(x, y) _mm_add_ps(x, y)
#  define f32x4_sub(x, y) _mm_sub_ps(x, y)

  /* Comparison ops produce integer results */
#  define f32x4_le(x, y) (__m128i)_mm_cmpnge_ps(x, y)
#  define f32x4_gt(x, y) (__m128i)_mm_cmplt_ps(x, y)

#  define i32x4_const(a, b, c, d) _mm_set_epi32(d, c, b, a)
#  define i32x4_all_false(a) _mm_test_all_zeros(a, a)
#  define i32x4_add(x, y) _mm_add_epi32(x, y)
#  define i32x4_sub(x, y) _mm_sub_epi32(x, y)
#  define i32x4_gt(x, y) _mm_cmpgt_epi32(x, y)
#  define i128_and(x, y) _mm_and_si128(x, y)
# else
#  error Bad architecture
# endif

  assert((lim_x - start_x) % 4 == 0);
  for ( unsigned py=start_y ; py < lim_y; py++ ) {
    i128_t* addr = (i128_t*)&iterations[py][start_x];
    f128_t  y0 = f32x4_splat(scale((float)py, HEIGHT, MINY, MAXY));
    for ( float px=start_x ; px < lim_x; px+=4 ) {
      f128_t x0 = f32x4_make(scale(px,   WIDTH, MINX, MAXX),
			     scale(px+1, WIDTH, MINX, MAXX),
			     scale(px+2, WIDTH, MINX, MAXX),
			     scale(px+3, WIDTH, MINX, MAXX));
      f128_t x = f32x4_const(0, 0, 0, 0);
      f128_t y = f32x4_const(0, 0, 0, 0);
      i128_t active = i32x4_const(-1, -1, -1, -1);
      i128_t counter = i32x4_const(CUTOFF, CUTOFF, CUTOFF, CUTOFF);
      for(;;) {
	f128_t x_sq = f32x4_mul(x, x);
	f128_t y_sq = f32x4_mul(y, y);
	f128_t sum_sq = f32x4_add(x_sq, y_sq);
	active = i128_and(active, f32x4_le(sum_sq, f32x4_const(4, 4, 4, 4)));
	active = i128_and(active, i32x4_gt(counter, i32x4_const(0, 0, 0, 0)));
	if (i32x4_all_false(active)) {
	  break;
	}
	f128_t tmp = f32x4_add(f32x4_sub(x_sq, y_sq), x0);
	f128_t xy = f32x4_mul(x, y);
	y = f32x4_add(f32x4_add(xy, xy), y0);
	x = tmp;
	counter = i32x4_add(counter, active);
      }
      counter = i32x4_sub(i32x4_const(CUTOFF, CUTOFF, CUTOFF, CUTOFF), counter);
      *addr++ = counter;
    }
  }
#elif defined(SIMD256)
# ifdef INTEL
  /* These abstractions are roughly from WASM SIMD although we have separate int and float types here. */
  typedef __m256i i256_t;
  typedef __m256 f256_t;

#  define f32x8_splat(x) _mm256_set_m128(_mm_set_ps1(x), _mm_set_ps1(x))
#  define f32x8_make(a, b, c, d, e, f, g, h) _mm256_set_ps(h, g, f, e, d, c, b, a)
#  define f32x8_const(a, b, c, d, e, f, g, h) _mm256_set_ps(h, g, f, e, d, c, b, a)
#  define f32x8_mul(x, y) _mm256_mul_ps(x, y)
#  define f32x8_add(x, y) _mm256_add_ps(x, y)
#  define f32x8_sub(x, y) _mm256_sub_ps(x, y)

  /* Comparison ops produce integer results */
#  define f32x8_le(x, y) (__m256i)_mm256_cmp_ps(x, y, _CMP_LT_OS)
#  define f32x8_gt(x, y) (__m256i)_mm256_cmp_ps(x, y, _CMP_GT_OS)

#  define i32x8_const(a, b, c, d, e, f, g, h) _mm256_set_epi32(h, g, f, e, d, c, b, a)
#  define i32x8_all_false(a) _mm256_testz_si256(a, a)
#  define i32x8_add(x, y) _mm256_add_epi32(x, y)
#  define i32x8_sub(x, y) _mm256_sub_epi32(x, y)
#  define i32x8_gt(x, y) _mm256_cmpgt_epi32(x, y)
#  define i256_and(x, y) _mm256_and_si256(x, y)
# else
#  error Bad architecture
# endif

  assert((lim_x - start_x) % 8 == 0);
  for ( unsigned py=start_y ; py < lim_y; py++ ) {
    i256_t* addr = (i256_t*)&iterations[py][start_x];
    f256_t  y0 = f32x8_splat(scale((float)py, HEIGHT, MINY, MAXY));
    for ( float px=start_x ; px < lim_x; px+=8 ) {
      f256_t x0 = f32x8_make(scale(px,   WIDTH, MINX, MAXX),
			     scale(px+1, WIDTH, MINX, MAXX),
			     scale(px+2, WIDTH, MINX, MAXX),
			     scale(px+3, WIDTH, MINX, MAXX),
			     scale(px+4, WIDTH, MINX, MAXX),
			     scale(px+5, WIDTH, MINX, MAXX),
			     scale(px+6, WIDTH, MINX, MAXX),
			     scale(px+7, WIDTH, MINX, MAXX));
      f256_t x = f32x8_const(0, 0, 0, 0, 0, 0, 0, 0);
      f256_t y = f32x8_const(0, 0, 0, 0, 0, 0, 0, 0);
      i256_t active = i32x8_const(-1, -1, -1, -1, -1, -1, -1, -1);
      i256_t counter = i32x8_const(CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF);
      for(;;) {
	f256_t x_sq = f32x8_mul(x, x);
	f256_t y_sq = f32x8_mul(y, y);
	f256_t sum_sq = f32x8_add(x_sq, y_sq);
	active = i256_and(active, f32x8_le(sum_sq, f32x8_const(4, 4, 4, 4, 4, 4, 4, 4)));
	active = i256_and(active, i32x8_gt(counter, i32x8_const(0, 0, 0, 0, 0, 0, 0, 0)));
	if (i32x8_all_false(active)) {
	  break;
	}
	f256_t tmp = f32x8_add(f32x8_sub(x_sq, y_sq), x0);
	f256_t xy = f32x8_mul(x, y);
	y = f32x8_add(f32x8_add(xy, xy), y0);
	x = tmp;
	counter = i32x8_add(counter, active);
      }
      counter = i32x8_sub(i32x8_const(CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF, CUTOFF), counter);
      *addr++ = counter;
    }
  }
#else
  /* Scalar code */
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
      iterations[py][px] = iteration;
    }
  }
#endif
}

struct qitem {
  unsigned start_y;
  unsigned lim_y;
  unsigned start_x;
  unsigned lim_x;
};

/* Circular queue, protected by the mutex.  Items are taken from head, inserted at tail.  It is full
   if count == QSIZE, empty if count == 0.  The single cond is shared by the two signals "added
   element to empty queue" and "queue is no longer full".
*/
static pthread_mutex_t q_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t q_cond = PTHREAD_COND_INITIALIZER;
#define QSIZE 100
static struct qitem queue[QSIZE];
static int q_tail = 0;
static int q_head = 0;
static int q_count = 0;

static void enqueue(const struct qitem* it) {
  pthread_mutex_lock(&q_lock);
  while (q_count == QSIZE) {
    pthread_cond_wait(&q_cond, &q_lock);
  }
  queue[q_tail] = *it;
  q_tail = (q_tail + 1) % QSIZE;
  q_count++;
  if (q_count == 1) {
    /* The broadcast is needed because producer and consumer share it and we can't risk
       waking up eg a blocked consumer when what we really want to do is wake up the
       producer.  Having two conditions would alleviate this. */
    pthread_cond_broadcast(&q_cond);
  }
  pthread_mutex_unlock(&q_lock);
}

static struct qitem dequeue() {
  pthread_mutex_lock(&q_lock);
  while (q_count == 0) {
    pthread_cond_wait(&q_cond, &q_lock);
  }
  struct qitem it = queue[q_head];
  q_head = (q_head + 1) % QSIZE;
  q_count--;
  if (q_count == QSIZE-1) {
    /* See above re broadcast */
    pthread_cond_broadcast(&q_cond);
  }
  pthread_mutex_unlock(&q_lock);
  return it;
}

/* Used for termination.  The master sets items_remaining to some positive number before creating
   any work items and goes to sleep on the condition variable.  The workers decrement the number,
   whoever gets to zero signals the condition variable to wake the master.  It's assumed that the
   queue is drained at that point. */
static pthread_mutex_t c_lock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t c_cond = PTHREAD_COND_INITIALIZER;
static int items_remaining = 0;

static void signal_work_done() {
  pthread_mutex_lock(&c_lock);
  if (--items_remaining == 0) {
    pthread_cond_signal(&c_cond);
  }
  pthread_mutex_unlock(&c_lock);
}

static void init_work_to_do(int items) {
  pthread_mutex_lock(&c_lock);
  items_remaining = items;
  pthread_mutex_unlock(&c_lock);
}

static void wait_for_work_done() {
  pthread_mutex_lock(&c_lock);
  while (items_remaining > 0) {
    pthread_cond_wait(&c_cond, &c_lock);
  }
  pthread_mutex_unlock(&c_lock);
}

/* variable names required for some older compilers */
static void* mandel_worker(void* dummy) {
  for (;;) {
    struct qitem it = dequeue();
    mandel_slice(it.start_y, it.lim_y, it.start_x, it.lim_x);
    signal_work_done();
  }
}

static void create_workers() {
  int i;
  for (i=0 ; i < num_threads; i++) {
    pthread_t dummy;
    pthread_create(&dummy, NULL, mandel_worker, NULL);
  }
}

static unsigned min(unsigned a, unsigned b) {
  return a < b ? a : b;
}

static void mandel() {
  unsigned rows = (HEIGHT + (TILEY - 1)) / TILEY;
  unsigned cols = (WIDTH + (TILEX - 1)) / TILEX;

#ifndef NDEBUG
  printf("Rows %u cols %u\n", rows, cols);
#endif
  init_work_to_do(rows*cols);

  unsigned ry, cx;
  for (ry = 0; ry < rows; ry++) {
    for (cx = 0; cx < cols; cx++) {
      struct qitem it = {
	.start_y = ry*TILEY,
	.lim_y = min((ry+1)*TILEY, HEIGHT),
	.start_x = cx*TILEX,
	.lim_x = min((cx+1)*TILEX, WIDTH)
      };
      enqueue(&it);
    }
  }

  wait_for_work_done();
}

static void dump(const char* filename) {
  FILE* out = fopen(filename, "w");
  fprintf(out, "P6 %d %d 255\n", WIDTH, HEIGHT);
  unsigned y, x;
  for (y=0; y < HEIGHT; y++) {
    for ( x = 0 ; x < WIDTH; x++ ) {
      unsigned r = 0, g = 0, b = 0;
      if (iterations[y][x] < CUTOFF) {
	from_rgb(mapping[iterations[y][x] % 16], &r, &g, &b);
      }
      fputc(r, out);
      fputc(g, out);
      fputc(b, out);
    }
  }
  fclose(out);
}

int main(int argc, char** argv) {
  if (argc > 1) {
    if (sscanf(argv[1], "-j%u", &num_threads) == 1) {
      if (num_threads == 0) {
	fprintf(stderr, "Zero threads\n");
	exit(1);
      }
    } else {
      fprintf(stderr, "Bad option %s\n", argv[1]);
    }
  }
  struct timeval before, after;
  create_workers();
  gettimeofday(&before, NULL);
  mandel();
  gettimeofday(&after, NULL);
  int64_t delta = ((int64_t)after.tv_sec - (int64_t)before.tv_sec)*1000000 + (after.tv_usec - before.tv_usec);
  printf("Elapsed %" PRIi64 "ms\n", delta/1000);
  dump("mandelsimd.ppm");
}
