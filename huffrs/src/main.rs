// Huffman compressor / decompressor
//
// (Based on the Go version in `sandbox/huff`)
//
// huffer compress [-o outfile] filename
// huff [-o outfile] filename
//   Creates outfile, or if no -o option, filename.huff
//
// huffer decompress [-o outfile] filename.huff
// puff [-o outfile] filename.huff
//   Creates outfile, or if no -o option, filename
//
// Since this is a programming exercise, it works by reading 64KB blocks and
// compressing them individually; the output file consists of compressed blocks.
// Also, we don't care about micro-efficiencies in representing the dictionary
// in the file or in complicated fallback schemes, more could be done.
//
// A compressed block is represented as
//   number of dictionary entries: u16 > 0 (max value is really 256)
//   run of dictionary entries sorted descending by frequency:
//     value: u8
//     frequency: u32 (max value is really 65536)
//   number of encoded bytes: u32 (max value is really 65536)
//   number of bytes used for encoded bytes: u32 (max value 65536)
//   bytes, the number of which is encoded by previous field
//
// An uncompressed block can be written under some circumstances, it is represented as
//   0: u16
//   number of bytes: u32 (really max 65536)
//   bytes, the number of which is encoded by previous field

mod heap;

use std::fs::File;
use crate::heap::*;

fn main() {
    let (is_compress, _, in_filename, out_filename) = parse_args();
    let res = if is_compress {
        compress_file(in_filename, out_filename)
    } else {
        decompress_file(in_filename, out_filename)
    };
    match res {
        Ok(()) => {}
        _ => { panic!("{:?}", res) }
    }
}

fn parse_args() -> (bool, bool, String, String) {
    let mut args = std::env::args();
    let mut is_compress = false;
    let mut is_decompress = false;

    // Infer operation from command name, maybe
    match args.next().expect("Must have command").as_str() {
        "huff" => {
            is_compress = true;
        }
        "puff" => {
            is_decompress = true;
        }
        _ => {}
    }

    // Look for verb if operation is not implied by command name
    if !is_compress && !is_decompress {
        match args.next().expect("Must have verb").as_str() {
            "compress" => {
                is_compress = true;
            }
            "decompress" => {
                is_decompress = true;
            }
            _ => {
                panic!("Expected 'compress' or 'decompress'")
            }
        }
    }
    // Parse remaining arguments
    let in_filename;
    let mut out_filename = String::from("");
    let mut n = args.next().expect("Must have input file");
    let mut have_out_filename = false;
    if n == "-o" {
        out_filename = args.next().expect("Must have output file after -o");
        n = args.next().expect("Must have input file");
        have_out_filename = true;
    }
    in_filename = n;
    if is_decompress && !in_filename.ends_with(".huff") {
       // TODO: Also must check that filename is not empty after stripping extension
        panic!("Input file must have extension .huff")
    }
    if !have_out_filename {
        out_filename = String::from(in_filename.as_str());
        if is_compress {
            out_filename.push_str(".huff")
        } else {
            _ = out_filename.split_off(out_filename.len() - 5);
        }
    }

    //println!("{} {}", in_filename.as_str(), out_filename.as_str());

    (is_compress, is_decompress, in_filename, out_filename)
}

fn compress_file(in_filename: String, out_filename: String) -> std::io::Result<()> {
    let mut input = File::open(in_filename)?;
    let mut output = File::create(out_filename)?;
    compress_stream(&mut input, &mut output)?;
    output.sync_all()?;
    Ok(())
}

const META_SIZE: usize =
     2 /* freq table size */ +
    256*5 /* freq table max size */ +
    4 /* number of input bytes encoded */ +
    4 /* number of bytes in encoding */;

fn compress_stream(input: &mut dyn std::io::Read,  output: &mut dyn std::io::Write) -> std::io::Result<()> {
    let mut in_buf = Vec::with_capacity(65536);
    let mut out_buf = Vec::with_capacity(65536);
    let mut meta_buf = Vec::with_capacity(META_SIZE);
    loop {
        let bytes_read = input.read(in_buf.as_mut_slice())?;
        if bytes_read == 0 {
            break
        }
        let input = &in_buf.as_slice()[0..bytes_read];
        let freq = compute_byte_frequencies(input);
        let tree = build_huffman_tree(freq.as_slice());
        let mut dict = Vec::with_capacity(256);
        let have_dict = populate_dict(0, 0, &tree, &mut dict);
        let mut did_encode = false;
        let mut bytes_encoded = 0;
        if have_dict {
            (did_encode, bytes_encoded) = compress_block(&dict, input, out_buf.as_mut_slice());
        }
        let mut metaloc = 0;
        if did_encode {
            metaloc = put(&mut meta_buf, metaloc, 2, freq.len());
            for item in freq {
                metaloc = put(&mut meta_buf, metaloc, 1, item.val as usize);
                metaloc = put(&mut meta_buf, metaloc, 4, item.count as usize);
            }
            metaloc = put(&mut meta_buf, metaloc, 4, bytes_read);
            metaloc = put(&mut meta_buf, metaloc, 4, bytes_encoded);
        } else {
            metaloc = put(&mut meta_buf, metaloc, 2, 0);
            metaloc = put(&mut meta_buf, metaloc, 4, bytes_read);
        }
        output.write(&meta_buf.as_slice()[0..metaloc])?;
        if did_encode {
            output.write(&out_buf.as_mut_slice()[0..bytes_encoded])?;
        } else {
            output.write(input)?;
        }
    }
    Ok(())
}

fn put(v: &mut Vec<u8>, mut p: usize, mut n: usize, mut val: usize) -> usize {
    while n > 0 {
        v[p] = val as u8;
        val >>= 8;
        n -= 1;
        p += 1;
    }
    p
}

fn compress_block(dict: &Vec<DictItem>, input: &[u8], output: &mut [u8]) -> (bool, usize) {
    let mut outptr = 0;
    let limit = output.len();
    let mut window = 0u64;
    let mut width = 0;
    for b in input {
        let e = &dict[*b as usize];
        window = window | ((e.bits as u64) << width);
        width += e.width;
        while width >= 8 {
            if outptr == limit {
                return (false, 0)
            }
            output[outptr] = window as u8;
            outptr += 1;
            window >>= 8;
            width -= 8;
        }
    }
    if width > 0 {
        output[outptr] = window as u8;
        outptr += 1;
    }
    (true, outptr)
}

fn decompress_file(in_filename: String, out_filename: String) -> std::io::Result<()> {
    panic!("No decompression yet")
}

// Encoding dictionary, mapping byte values to bit strings.

struct DictItem {
    width: usize,
    bits: u64
}

fn populate_dict(width: usize, bits: u64, tree: &Box<HuffTree>, dict: &mut Vec<DictItem>) -> bool {
    match &tree.left {
        Some(_) => {
            return populate_dict(width+1, bits, &tree.left.as_ref().expect("LEFT"), dict) &&
                   populate_dict(width+1, (1<<width)|bits, &tree.right.as_ref().expect("RIGHT"), dict)
        }
        None => {
            // "56" is an artifact of the implementation of compression, it guarantees that
            // we don't have to deal with overflow when constructing the bitstring.
            if width > 56 {
                return false
            }
            dict[tree.val as usize].bits = bits;
            dict[tree.val as usize].width = width;
            return true
        }
    } 
}

// Huffman tree, representing the encoding of byte values by the bit path to a leaf in the
// binary tree.

struct HuffTree {
    left: Option<Box<HuffTree>>,
    right: Option<Box<HuffTree>>,
    val: u8
}

fn build_huffman_tree(freq: &[FreqEntry]) -> Box<HuffTree> {
    let mut heap = Heap::<Box<HuffTree>, u32>::new(|a, b| a < b );
    for i in freq {
        let t = Box::new(HuffTree { val: i.val, left: None, right: None });
        heap.insert(i.count, t)
    }
    while heap.len() > 1 {
        let (a, wa) = heap.extract_max();
        let (b, wb) = heap.extract_max();
        let t = Box::new(HuffTree { val: 0, left: Some(a), right: Some(b)});
        heap.insert(wa + wb, t)
    }
    heap.extract_max().0
}

// Byte frequency count.  Returned vector has counts for bytes with non-zero frequencies,
// in unspecified order

#[derive(Clone,Copy)]
struct FreqEntry {
    val: u8,
    count: u32
}

fn compute_byte_frequencies(bytes: &[u8]) -> Vec<FreqEntry> {
    let mut freq = Vec::<FreqEntry>::with_capacity(256);

    for i in 0..256 {
        freq[i].val = i as u8;
    }

    for i in bytes {
        freq[*i as usize].count += 1;
    }

    // Pack nonzero entries to the start of the vector and discard the zero ones
    let mut i = 0;
    let mut j = 0;
    loop {
        while j < 256 && freq[j].count == 0 {
            j += 1;
        }
        if j == 256 {
            break;
        }
        freq[i] = freq[j];
        i += 1;
        j += 1;
    }
    freq.truncate(i);

    freq
}