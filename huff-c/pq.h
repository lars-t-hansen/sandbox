/* -*- c-basic-offset: 4; indent-tabs-mode: nil; fill-column: 100 -*-
 *
 * Generic bounded-length priority queue type.  The elements of the queue MUST be bitwise copyable.
 */

#ifndef pq_h_included
#define pq_h_included

#include <stdint.h>

/* The structure contents are private */

typedef struct pq {
    void*     elements;         /* Storage, not owned by us */
    size_t    element_size;     /* The size of each element */
    size_t    element_count;    /* The length of storage */
    size_t    length;           /* PUBLIC: The number of active elements */
    int       (*greater)(const void* a, const void* b);
} pq_t;

/* Borrow the `elements` array mutably and use it for the priority queue.  Each element is
 * `elements_size` in length, and there are `num_elements` of them.  Only the first `length`
 * elements have values, the rest are garbage.
 */
void pq_new(pq_t* pq, void* elements, size_t element_size, size_t num_elements,
            size_t initialized_length, int (*greater)(const void* a, const void* b));

/* Number of elements in the heap */
size_t pq_length(pq_t* pq);

/* Insert the element.  The element will be copied into the queue.  Abort if the heap is full. */
void pq_insert(pq_t* pq, const void* element);

/* Extract the largest element.  The element will be copied into *storage, which must be large
 * enough to receive it.  Abort if the heap is empty.
 */
void pq_extract_max(pq_t* pq, void* storage);

#endif /* !pq_h_included */
