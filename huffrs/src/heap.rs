// Priority queue using a heap.

struct Node<Val, Weight> {
    weight: Weight,
    val: Val
}

pub struct Heap<Val, Weight: Copy> {
    xs: Vec<Node<Val, Weight>>,
    greater: fn(Weight, Weight) -> bool,
}

impl<Val, Weight: Copy> Heap<Val, Weight> {
    pub fn new(greater: fn(Weight, Weight) -> bool) -> Heap<Val, Weight> {
        Heap::<Val, Weight> {
            xs: Vec::<Node<Val, Weight>>::new(),
            greater: greater
        }
    }

    pub fn len(&self) -> usize {
        self.xs.len()
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
        while i > 0 && (self.greater)(self.xs[i].weight, self.xs[parent(i)].weight) {
            self.xs.as_mut_slice().swap(i, parent(i));
            i = parent(i);
        }
    }

    fn heapify(&mut self, mut loc: usize) {
        loop {
            let mut greatest = loc;
            let l = left(loc);
            if l < self.xs.len() && (self.greater)(self.xs[l].weight, self.xs[loc].weight) {
                greatest = l;
            }
            let r = right(loc);
            if r < self.xs.len() && (self.greater)(self.xs[r].weight, self.xs[greatest].weight) {
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