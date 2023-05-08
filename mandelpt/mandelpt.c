/* Pthreads mandelbrot */

/* This creates a thread pool and a shared work queue and farms out work.  The workers write
   directly into the result array and signal completion to the coordinator.  On my laptop I see nearly
   4x speedup over the sequential version with 4 threads (1822ms vs 459ms). */

#include <pthread.h>

#define DEFAULT_THREADS 4

static unsigned num_threads = DEFAULT_THREADS;

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

static unsigned iterations[HEIGHT * WIDTH];

#include "../mandelcommon/mandelcommon.h"

static inline float_t scale(float_t v, float_t rng, float_t min, float_t max) {
  return min + v*(max-min)/rng;
}

static void mandel_slice(unsigned start_y, unsigned lim_y, unsigned start_x, unsigned lim_x) {
#ifndef NDEBUG
  printf("Work %u %u %u %u\n", start_y, lim_y, start_x, lim_x);
#endif
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
  create_workers();
  begin_timer();
  mandel();
  end_timer("Compute");
  dump("mandelpt.ppm");
}
