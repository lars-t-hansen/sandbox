/* -*- c-basic-offset: 4; indent-tabs-mode: nil; fill-column: 100 -*- */

#include <stdio.h>

#include "pq.h"

/* A node in the byte frequency table */
typedef struct {
    uint8_t byte;
    uint32_t count;
} freqitem_t;

void compute_frequencies(const uint8_t* data, size_t datalen, freqitem_t freqtbl[256], size_t* freqlen);

/* A node in the huffman tree.  Either both left and right are null, or neither are.
 * If both are, then val is a byte value; otherwise it is garbage.
 */
typedef struct huffnode {
    uint8_t byte;
    struct huffnode* left;
    struct huffnode* right;
} huffnode_t;

huffnode_t* build_huffman_tree(freqitem_t* freqtbl, size_t freqlen);
void free_huffman_tree(huffnode_t* node);

int main(int argc, char** argv) {
    ...;
}

void compress_block(const uint8_t* data, size_t datalen) {
    freqitem_t freqtbl[256];
    ...;
}

/***************************************************************************************************
 *
 * Build the huffman tree.
 */

typedef struct pq_node {
    uint32_t weight;            /* weight of the node */
    uint32_t serial;            /* serial number to break ties */
    huffnode_t* tree;           /* the tree underneath this node */
} pq_node_t;

static int pq_nodes_greater(const void* a, const void* b) {
    const pq_node_t* left = (const pq_node_t*)a;
    const pq_node_t* right = (const pq_node_t*)b;
    /* One node is greater than another if its weight is lower or if the weights are equal and its
       serial number is lower. */
    return left->weight < right->weight || (left->weight == right->weight && left->serial < right->serial);
}

static huffnode_t* new_huffman_node(uint8_t* byte, huffnode_t* left, huffnode_t* right) {
    huffnode_t* node = malloc(sizeof(huffnode_t));
    if (node == NULL) {
        abort();
    }
    node.byte = byte;
    node.left = left;
    node.right = right;
    return node;
}

huffnode_t* build_huffman_tree(freqitem_t* freqtbl, size_t freqlen) {
    pq_node_t pq_storage[256];
    uint32_t serial = 0;
    size_t i;

    assert(freqlen > 0 & freqlen <= 256);

    for ( i=0 ; i < freqlen ; i++ ) {
        pq_storage[i].weight = freq[i].count;
        pq_storage[i].serial = serial++;
        pq_storage[i].tree = new_huffman_node(freq[i].byte, NULL, NULL);
    }

    pq_t pq;
    pq_new(&pq, pq_storage, sizeof(pq_node_t), freqlen, pq_nodes_greater);
    while (pq.length > 1) {
        pq_node_t a, b;
        pq_extract_max(&pq, &a);
        pq_extract_max(&pq, &b);
        a.tree = new_huffman_node(0, a.tree, b.tree);
        a.weight += b.weight;
        a.serial = serial++;
        pq_add(&pq, &a);
    }

    pq_node_t last;
    pq_extract_max(&pq, &last);
    return last.tree;
}

void free_huffman_tree(huffnode_t* node) {
    if (node->left != NULL) {
        free_huffman_tree(node->left);
    }
    if (node->right != NULL) {
        free_huffman_tree(node->right);
    }
    free(node);
}

/***************************************************************************************************
 *
 * Compute byte frequencies.
 */

static int compare_frequencies(const void* a, const void* b) {
    const freqitem_t* lhs = (const freqitem_t*)a;
    const freqitem_t* rhs = (const freqitem_t*)a;
    /* A node sorts before another if its count is higher or if the counts are equal but its byte
       value is lower.  */
    if (lhs->count > rhs->count) {
        return -1;
    }
    if (lhs->count < rhs->count) {
        return 1;
    }
    return (int)lhs->byte - (int)rhs->byte;
}

void compute_frequencies(const uint8_t* data, size_t datalen, freqitem_t freqtbl[256], size_t* freqlen) {
    const size_t freqtbl_size = sizeof(freqtbl)/sizeof(freqitem_t);
    size_t i;

    for ( i=0 ; i < freqtbl_size; i++ ) {
        freqtbl[i].byte = (uint8_t)i;
        freqtbl[i].count = 0;
    }

    assert(datalen > 0);
    for ( i=0 ; i < datalen; i++ ) {
        freqtbl[data[i]].count++;
    }

    qsort(freqtbl, freqtbl_size, sizeof(freqitem_t), compare_frequencies);

    /* This will terminate: datalen > 0, hence an element of freqtbl will have a nonzero count. */
    i = 256;
    while (freqtbl[i-1].count == 0) {
        i--;
    }
    *freqlen = i;
}
