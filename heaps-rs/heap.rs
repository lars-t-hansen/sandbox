// A heap data structure, and utilities for using it as a priority queue.

use std::cmp;

struct Node<Val, Weight> {
    weight: Weight,
    val: Val
}

pub struct Heap<Val, Weight: Copy + cmp::PartialOrd> {
    xs: Vec<Node<Val, Weight>>
}

impl<Val, Weight: Copy + cmp::PartialOrd> Heap<Val, Weight> {
    pub fn new() -> Heap<Val, Weight> {
        Heap::<Val, Weight> {
            xs: Vec::<Node<Val, Weight>>::new()
        }
    }

    pub fn len(&self) -> usize {
        self.xs.len()
    }

    pub fn max_weight(&self) -> Weight {
        assert!(self.len() > 0);
        self.xs[0].weight   
    }

    pub fn extract_max(&mut self) -> (Val, Weight) {
        let l = self.xs.len();
        assert!(l > 0);
        self.xs.as_mut_slice().swap(0, l-1);
        let max = self.xs.pop().expect("POP");
        if l > 2 {
            self.heapify(0)
        }
        (max.val, max.weight)
    }

    pub fn insert(&mut self, weight: Weight, t: Val) {
        self.xs.push(Node{weight:weight, val:t});
        let mut i = self.xs.len() - 1;
        while i > 0 && self.xs[i].weight > self.xs[parent(i)].weight {
            self.xs.as_mut_slice().swap(i, parent(i));
            i = parent(i);
        }
    }

    fn heapify(&mut self, mut loc: usize) {
        loop {
            let mut greatest = loc;
            let l = left(loc);
            if l < self.xs.len() && self.xs[l].weight > self.xs[loc].weight {
                greatest = l;
            }
            let r = right(loc);
            if r < self.xs.len() && self.xs[r].weight > self.xs[greatest].weight {
                greatest = r;
            }
            if greatest == loc {
                break;
            }
            self.xs.as_mut_slice().swap(loc, greatest);
            loc = greatest;
        }
    }
}

fn parent(loc: usize) -> usize {
    (loc - 1) / 2
}

fn left(loc: usize) -> usize {
    (loc * 2) + 1
}

fn right(loc: usize) -> usize {
    (loc + 1) * 2
}