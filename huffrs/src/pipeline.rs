// 

// single-producer multiple-workers single-consumer pipeline, with the
// producer being run on the invoking thread and the others being invoked
// on background threads.
//
// the notion here is that there is a fixed number of "items" that are
// reused (expensive to create / manage) and used to send work around
// the pipeline; there can be more of these than there are workers.
//
// the pipeline handles communication, error signalling, and so on

use std::thread;
use crossbeam_channel as channel;

pub trait ItemBase {
    new() -> Self;

    // the id is used by the pipeline to indicate ordering of work items, generally
    // the other methods on Item should not inspect it.
    id(&self) -> usize;
    set_id(&mut self, usize: id);

    // produce returns Ok(true) for work obtained, Ok(false) for orderly end of input
    produce(&mut self, input: &mut fs::File) -> Result<bool, String>;
    work(&mut self);
    consume(&mut self, output: &mut fs::File) -> Result<(), String>;
}

pub trait WorkerData {
    new() -> Self;
}

pub struct Pipeline<Item: ItemBase + Send, Aux: WorkerData> {
}

impl<Item: ItemBase + Send, Extra: ExtraBase> Pipeline<Item, Extra> {
    fn run(num_workers: usize, queue_size: usize,  input: &mut fs::File, output: &mut fs::File) -> Result<(), String> {
        let mut items = Vec::<Item>::with_capacity(2*num_workers);
        for _ in 0..items.capacity() {
            items.push(T::new())
        }

        // Various logic depends on this equality:
        assert!(items.len() == items.capacity());
    
        let error_flag = Cell::<AtomicBool>::new(false);
        let (available_s, available_r) = channel::unbounded();
        let (ready_s, ready_r) = channel::unbounded();
        let (done_s, done_r) = channel::unbounded();
        std::thread::scope(|s| {
            let writer_thread = s.spawn(|| consumer_loop(&error_flag, ready_r, done_s) );
            let mut worker_threads = Vec::with_capacity(num_workers);
            let mut extra = Vec::with_capacity(num_workers)
            for _ in 0..worker_threads.capacity() {
                let available_r = available_r.clone();
                let ready_s = ready_s.clone();
                worker_threads.push(s.spawn(|| worker_loop(available_r, ready_s)))
                extra.push(U::new())
            }
            // These are dead so drop them, to allow the closing of channels to trigger
            // shutdown as described below.
            drop(available_r);
            drop(ready_s);
        }
        let mut p = Pipeline {
            worker_threads,
            writer_thread,
            producer: || producer_loop(&error_flag, items, done_r, available_s) }
        
        Pipeline::producer_loop(...);
        for w in self.worker_threads {
            let _ = w.join();
        }
        let _ = self.writer_thread.join();

        // Obviously we could communicate something more interesting.
        if error_flag.load(atomic::Ordering::Relaxed) {
            return Err(io::Error::new(io::ErrorKind::Other, "Compression error"))
        }
        Ok(())
    }

    fn producer_loop(&mut self, mut items:Vec<Item>, done: Receiver<Item>, available: Sender<Item>, mut input: fs::File) {
        // When this returns, whether normally or by error, it will close `available_s`.
        // That will make the encoders exit their encoding loops and trigger reliable
        // shutdown of the writer thread too.
        let mut next_read_id = 0;
        loop {
            if self.error_flag.load(atomic::Ordering::Relaxed) {
                return
            }
            if items.len() == 0 {
                items.push(done.recv().unwrap())
            }
            let mut item = items.pop().unwrap();
            // Call the producer here.  Three results: Ok() means the item was set up
            // and should be pushed to the workers with a new id.  Done() means we're
            // done and should exit.  Err() means we errored out.
            match input.read(item.in_buf.as_mut_slice()) {
                Ok(bytes_read) => {
                    if bytes_read == 0 {
                        return
                    }
                    item.in_buf_size = bytes_read;
                    item.id = next_read_id;
                    next_read_id += 1;
                    available.send(item).unwrap();
                }
                Err(_) => {
                    self.error_flag.store(true, atomic::Ordering::Relaxed);
                }
            }
        }
    }

    fn encoder_loop(encode: fn(m: &mut T, extra: &mut U) -> Result<()>, available: Receiver<Item>, ready: Sender<Item>) {
        // The reader closes `available_r` to signal shutdown, and when we fail to
        // receive we exit the loop.
        //
        // When this leaves the worker loop it will close its copy of `ready_s`,
        // and once all the workers have closed, shutdown will be triggered in
        // the writer too.
        //
        // For the workers there is per-worker storage that needs to be set up somehow
        // and passed to the worker loop, a little bit of a headache.
        let mut freq_buf = Box::new([FreqEntry{val: 0, count: 0}; 256]);
        let mut dict = Box::new([DictItem {width: 0, bits: 0}; 256]);
        loop {
            match available.recv() {
                Ok(mut b) => {
                    
                    let input = &b.in_buf.as_slice()[0..b.in_buf_size];
                    (b.meta_buf_size, b.out_buf_size) =
                        encode_block(input, freq_buf.as_mut_slice(), dict.as_mut_slice(), b.meta_buf.as_mut_slice(), b.out_buf.as_mut_slice());
                    ready.send(b).unwrap();
                }
                Err(_) => { break }
            }
        }
    }

    fn writer_loop(error_flag: &AtomicBool, ready: Receiver<Item>, done: Sender<Item>, mut output: fs::File) {
        // The workers will shut down the `ready_s` channel and trigger shutdown of the writer.
        //
        // The writer can also shut down due to write error.  Once it discovers a write error it sets the error
        // flag and then consumes input without processing it, apart from forwarding the item to its consumer.  
        // The reader and the workers will stop producing input for the writer once they see that the error flag
        // is set.
        let mut next_write_id = 0;
        let mut queue = BinaryHeap::<Item>::new();
        let mut has_error = false;
        loop {
            match ready.recv() {
                Ok(item) => {
                    queue.push(item);
                    while !queue.is_empty() && queue.peek().unwrap().id == next_write_id {
                        let mut item = queue.pop().unwrap();
        
                        if !has_error {
                            let meta_data = &item.meta_buf[..item.meta_buf_size];
                            if output.write(meta_data).is_err() {
                                error_flag.store(true, atomic::Ordering::Relaxed);
                                has_error = true;
                            }
                        }

                        if !has_error {
                            let out_data = if item.out_buf_size == 0 { 
                                &item.in_buf[..item.in_buf_size]
                            } else {
                                &item.out_buf[..item.out_buf_size]
                            };
                            if output.write(out_data).is_err() {
                                error_flag.store(true, atomic::Ordering::Relaxed);
                                has_error = true;
                            }
                        }

                        item.id = 0;
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
}
