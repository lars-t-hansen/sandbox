// Single-producer multiple-workers single-consumer pipeline, with the
// producer being run on the invoking thread and the others being invoked
// on background threads.
//
// The notion here is that there is a fixed number of "items" that are
// reused (expensive to create / manage) and used to send work around
// the pipeline; there can be more of these than there are workers.
//
// The pipeline handles communication, error signalling, and so on

use std::collections::BinaryHeap;
use std::fs;
use std::sync::atomic::{self,AtomicBool};
use crossbeam_channel as channel;

// Per-worker data.
pub trait WorkerData {
    fn new() -> Self;
}

// Per-work-item data.
pub trait WorkItem<Aux: WorkerData>: Send + Ord {
    fn new() -> Self;

    // the id is used by the pipeline to indicate ordering of work items, generally
    // the other methods on WorkItem should not inspect it.
    //
    // TODO: Move the id out of the WorkItem and into some data structure here.  Then
    // the Ord requirement on WorkItem would also move into this module.
    fn id(&self) -> usize;
    fn set_id(&mut self, id: usize);

    // produce returns Ok(true) for work obtained, Ok(false) for orderly end of input
    fn produce(&mut self, input: &mut fs::File) -> Result<bool, String>;
    fn work(&mut self, aux: &mut Aux);
    fn consume(&mut self, output: &mut fs::File) -> Result<(), String>;
}

pub fn run<Aux: WorkerData, Work: WorkItem<Aux>>(num_workers: usize, queue_size: usize, input: fs::File, output: fs::File) -> Result<(), String> {
    let mut items = Vec::<Box<Work>>::with_capacity(queue_size);
    for _ in 0..queue_size {
        items.push(Box::new(Work::new()))
    }

    // Various logic depends on this equality:
    assert!(items.len() == items.capacity());

    let error_flag = AtomicBool::new(false);
    let (available_s, available_r) = channel::unbounded();
    let (ready_s, ready_r) = channel::unbounded();
    let (done_s, done_r) = channel::unbounded();
    std::thread::scope(|s| {
        let writer_thread = s.spawn(|| consumer_loop(&error_flag, ready_r, done_s, output) );
        let mut worker_threads = Vec::with_capacity(num_workers);
        for _ in 0..worker_threads.capacity() {
            let available_r = available_r.clone();
            let ready_s = ready_s.clone();
            worker_threads.push(s.spawn(|| worker_loop(available_r, ready_s)));
        }
        // These are dead so drop them, to allow the closing of channels to trigger
        // shutdown as described below.
        drop(available_r);
        drop(ready_s);
        producer_loop(&error_flag, items, done_r, available_s, input);
        for w in worker_threads {
            let _ = w.join();
        }
        let _ = writer_thread.join();

        // Obviously we could communicate something more interesting.
        if error_flag.load(atomic::Ordering::Relaxed) {
            return Err("Compression error".to_string())
        }

        Ok(())
    })
}

fn producer_loop<Aux: WorkerData, Work: WorkItem<Aux>>(error_flag: &AtomicBool, mut items:Vec<Box<Work>>,
                    done: channel::Receiver<Box<Work>>, available: channel::Sender<Box<Work>>,
                    mut input: fs::File) {
    // When this returns, whether normally or by error, it will close `available_s`.
    // That will make the encoders exit their encoding loops and trigger reliable
    // shutdown of the writer thread too.
    let mut next_read_id = 0;
    loop {
        if error_flag.load(atomic::Ordering::Relaxed) {
            return
        }
        if items.len() == 0 {
            items.push(done.recv().unwrap())
        }
        let mut item = items.pop().unwrap();
        match item.produce(&mut input) {
            Ok(got_input) => {
                if !got_input {
                    return
                }
                item.set_id(next_read_id);
                next_read_id += 1;
                available.send(item).unwrap();
            }
            Err(_) => {
                error_flag.store(true, atomic::Ordering::Relaxed);
            }
        }
    }
}

fn worker_loop<Aux: WorkerData, Work: WorkItem<Aux>>(available: channel::Receiver<Box<Work>>, ready: channel::Sender<Box<Work>>) {
    // The reader closes `available_r` to signal shutdown, and when we fail to
    // receive we exit the loop.
    //
    // When this leaves the worker loop it will close its copy of `ready_s`,
    // and once all the workers have closed, shutdown will be triggered in
    // the writer too.
    let mut aux = Aux::new();
    loop {
        match available.recv() {
            Ok(mut b) => {
                b.work(&mut aux);
                ready.send(b).unwrap();
            }
            Err(_) => { break }
        }
    }
}

fn consumer_loop<Aux: WorkerData, Work: WorkItem<Aux>>(error_flag: &AtomicBool, ready: channel::Receiver<Box<Work>>, done: channel::Sender<Box<Work>>, mut output: fs::File) {
    // The workers will shut down the `ready_s` channel and trigger shutdown of the writer.
    //
    // The writer can also shut down due to write error.  Once it discovers a write error it sets the error
    // flag and then consumes input without processing it, apart from forwarding the item to its consumer.  
    // The reader and the workers will stop producing input for the writer once they see that the error flag
    // is set.
    let mut next_write_id = 0;
    let mut queue = BinaryHeap::<Box<Work>>::new();
    let mut has_error = false;
    loop {
        match ready.recv() {
            Ok(item) => {
                queue.push(item);
                while !queue.is_empty() && queue.peek().unwrap().id() == next_write_id {
                    let mut item = queue.pop().unwrap();
                    if !has_error {
                        match item.consume(&mut output) {
                            Err(_) => {
                                has_error = true
                            }
                            Ok(_) => {}
                        }
                    }
                    item.set_id(0);
                    next_write_id += 1;
                    let _ = done.send(item);
                }
            }
            Err(_) => {
                assert!(queue.len() == 0);
                break
            }
        }
    }
    if !has_error {
        if output.sync_all().is_err() {
            error_flag.store(true, atomic::Ordering::Relaxed);
        }
    }
}
