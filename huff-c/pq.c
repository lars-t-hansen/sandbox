/* -*- c-basic-offset: 4; indent-tabs-mode: nil; fill-column: 100 -*- */

#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#include "pq.h"

static char* element_pointer(pq_t* pq, size_t n) {
    return (char*)pq->elements + n * pq->element_size;
}

static void swap(pq_t* pq, size_t x, size_t y) {
    /* We need pq->element_size temp storage, but we don't want to assume C11-style variable-length
     * local arrays, we don't want to use globals or statics (so as to remain thread-safe), and we
     * don't want to use heap allocation because that means there either has to be a destructor for
     * the pq_t or we will incur allocation for every call to pq_insert() and pq_extract_max().  We
     * could have required the client to provide a temp but this is annoying.  Instead, just swap
     * piecewise.  In typical cases the loop trip count below is 1.
     */
    size_t nbytes = pq->element_size;
    char* xp = element_pointer(pq, x);
    char* yp = element_pointer(pq, y);
    while (nbytes > 0) {
        char temp[32];
        size_t to_copy = nbytes > sizeof(temp) ? sizeof(temp) : nbytes;
        memcpy(temp, xp, to_copy);
        memcpy(xp, yp, to_copy);
        memcpy(yp, temp, to_copy);
        xp += to_copy;
        yp += to_copy;
        nbytes -= to_copy;
    }
}

static size_t parent(size_t loc) {
    return (loc - 1) / 2;
}

static size_t left(size_t loc) {
    return (loc * 2) + 1;
}

static size_t right(size_t loc) {
    return (loc + 1) * 2;
}

static void heapify(pq_t* pq, size_t loc) {
    for(;;) {
        size_t greatest = loc;
        size_t l = left(loc);
        if (l < pq->length && pq->greater(element_pointer(pq, l), element_pointer(pq, greatest))) {
            greatest = l;
        }
        size_t r = right(loc);
        if (r < pq->length && pq->greater(element_pointer(pq, r), element_pointer(pq, greatest))) {
            greatest = r;
        }
        if (greatest == loc) {
            break;
        }
        swap(pq, loc, greatest);
        loc = greatest;
    }
}

void pq_new(pq_t* pq, void* elements, size_t element_size, size_t element_count,
            size_t initialized_length, int (*greater)(const void* a, const void* b))
{
    pq->elements = elements;
    pq->element_size = element_size;
    pq->element_count = element_count;
    pq->length = initialized_length;
    pq->greater = greater;
    size_t i;
    for ( i = pq->length / 2 ; i >= 0 ; i-- ) {
        heapify(pq, i);
    }
}

void pq_insert(pq_t* pq, const void* element) {
    if (pq->length == pq->element_count) {
        abort();
    }
    size_t loc = pq->length;
    pq->length++;
    memcpy(element_pointer(pq, loc), element, pq->element_size);
    while (loc > 0 && pq->greater(element_pointer(pq, loc), element_pointer(pq, parent(loc)))) {
        swap(pq, loc, parent(loc));
        loc = parent(loc);
    }
}

void pq_extract_max(pq_t* pq, void* storage) {
    if (pq->length == 0) {
        abort();
    }
    memcpy(storage, element_pointer(pq, 0), pq->element_size);
    memcpy(element_pointer(pq, 0), element_pointer(pq, pq->length - 1), pq->element_size);
    pq->length--;
    if (pq->length > 1) {
        heapify(pq, 0);
    }
}
