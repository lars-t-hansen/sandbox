// Huffman compressor / decompressor
//
// (Based on the Go version in `sandbox/huff`, except this is still not multi-threaded)
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
//   run of dictionary entries sorted descending by frequency with ties
//   broken by lower byte values first:
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
use std::{cmp,env,io};

fn main() {
    let (is_compress, _, in_filename, out_filename) = match parse_args() {
        Ok(x) => x,
        Err(e) => { panic!("{}", e) }
    };
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

fn parse_args() -> Result<(/* is_compress */ bool, /* is_decompress */ bool, /* in_filename */ String, /* out_filename */ String), String> {
    let mut args = env::args();
    let mut is_compress = false;
    let mut is_decompress = false;

    // Infer operation from command name, maybe
    match args.next().unwrap().as_str() {
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
        match args.next().expect("Must have a verb").as_str() {
            "compress" => {
                is_compress = true;
            }
            "decompress" => {
                is_decompress = true;
            }
            _ => {
                panic!("Expected verb to be 'compress' or 'decompress'")
            }
        }
    }
    // Parse remaining arguments
    let mut out_filename = String::from("");
    let mut n = args.next().expect("Must have input file");
    let mut have_out_filename = false;
    if n == "-o" {
        out_filename = args.next().expect("Must have output file after -o");
        n = args.next().expect("Must have input file");
        have_out_filename = true;
    }
    let in_filename = n;
    if is_decompress && !in_filename.ends_with(".huff") {
       // TODO: Also must check that filename is not empty after stripping extension
        return Err("Input file must have extension .huff".to_string())
    }
    if !have_out_filename {
        out_filename = String::from(in_filename.as_str());
        if is_compress {
            out_filename.push_str(".huff")
        } else {
            _ = out_filename.split_off(out_filename.len() - 5);
        }
    }

    Ok((is_compress, is_decompress, in_filename, out_filename))
}

const META_SIZE: usize =
    2 /* freq table size */ +
    256*5 /* freq table max size */ +
    4 /* number of input bytes encoded */ +
    4 /* number of bytes in encoding */;

fn compress_file(in_filename: String, out_filename: String) -> io::Result<()> {
    let mut input = File::open(in_filename)?;
    let mut output = File::create(out_filename)?;
    compress_stream(&mut input, &mut output)?;
    output.sync_all()?;
    Ok(())
}

fn compress_stream(input: &mut dyn io::Read,  output: &mut dyn io::Write) -> io::Result<()> {
    let mut in_buf = Box::new([0u8; 65536]);
    let mut out_buf = Box::new([0u8; 65536]);
    let mut meta_buf = Box::new([0u8; META_SIZE]);
    let mut freq_buf = Box::new([FreqEntry{val: 0, count: 0}; 256]);
    let mut dict = Box::new([DictItem {width: 0, bits: 0}; 256]);
    loop {
        let bytes_read = input.read(in_buf.as_mut_slice())?;
        if bytes_read == 0 {
            break
        }
        let input = &in_buf.as_slice()[0..bytes_read];
        let (meta_data, out_data) = encode_block(input, freq_buf.as_mut_slice(), dict.as_mut_slice(), meta_buf.as_mut_slice(), out_buf.as_mut_slice());
        output.write(meta_data)?;
        output.write(out_data)?;
    }
    Ok(())
}

fn encode_block<'a, 'b>(input: &'b [u8], freq_buf: &mut [FreqEntry], dict: &mut [DictItem], meta_buf: &'a mut [u8], out_buf: &'b mut [u8]) ->
        (/* meta_data */ &'a [u8], /* out_data */ &'b [u8]) {
    let bytes_read = input.len();
    let freq = compute_byte_frequencies(input, freq_buf);
    let tree = build_huffman_tree(&freq);
    let have_dict = populate_dict(0, 0, &tree, dict);
    let mut did_encode = false;
    let mut bytes_encoded = 0;
    if have_dict {
        (did_encode, bytes_encoded) = compress_block(dict, input, out_buf);
    }
    let mut metaloc = 0;
    let output;
    if did_encode {
        metaloc = put(meta_buf, metaloc, 2, freq.len() as u64);
        for item in freq {
            metaloc = put(meta_buf, metaloc, 1, item.val as u64);
            metaloc = put(meta_buf, metaloc, 4, item.count as u64);
        }
        metaloc = put(meta_buf, metaloc, 4, bytes_read as u64);
        metaloc = put(meta_buf, metaloc, 4, bytes_encoded as u64);
        output = &out_buf[0..bytes_encoded];
    } else {
        metaloc = put(meta_buf, metaloc, 2, 0u64);
        metaloc = put(meta_buf, metaloc, 4, bytes_read as u64);
        output = input;
    }
    (&meta_buf[0..metaloc], output)
}

fn compress_block(dict: &[DictItem], input: &[u8], output: &mut [u8]) -> (/* success */ bool, /* bytes_encoded */ usize) {
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

fn decompress_file(in_filename: String, out_filename: String) -> io::Result<()> {
    let mut input = File::open(in_filename)?;
    let mut output = File::create(out_filename)?;
    decompress_stream(&mut input, &mut output)?;
    output.sync_all()?;
    Ok(())
}

fn decompress_stream(input: &mut dyn io::Read,  output: &mut dyn io::Write) -> io::Result<()> {
    let mut in_buf = Box::new([0u8; 65536]);
    let mut out_buf = Box::new([0u8; 65536]);
    let mut meta_buf = Box::new([0u8; META_SIZE]);
    let mut freq_buf = Box::new([FreqEntry{val: 0, count: 0}; 256]);
    loop {
        let got_metadata = read_bytes(input, 0, 2, meta_buf.as_mut_slice())?;
        if !got_metadata {
            break
        }
        let mut metaloc = 0;
        let mut item;
        (metaloc, item) = get(meta_buf.as_slice(), metaloc, 2);
        let freq_len = item as usize;
        let metabytes = if freq_len > 0 { 5*freq_len + 8 } else { 4 };
        let got_metadata = read_bytes(input, 2, metabytes, meta_buf.as_mut_slice())?;
        if !got_metadata {
            return Err(io::Error::new(io::ErrorKind::Other, "Bad metadata"));
        }
        let bytes_encoded;
        let bytes_to_decode;
        if freq_len > 0 {
            let mut freq = &mut freq_buf.as_mut_slice()[..freq_len];
            let mut i = 0;
            while i < freq_len {
                (metaloc, item) = get(meta_buf.as_slice(), metaloc, 1);
                freq[i].val = item as u8;
                (metaloc, item) = get(meta_buf.as_slice(), metaloc, 4);
                freq[i].count = item as u32;
                i += 1;
            }
            (metaloc, item) = get(meta_buf.as_slice(), metaloc, 4);
            bytes_to_decode = item as usize;
            (_, item) = get(meta_buf.as_slice(), metaloc, 4);
            bytes_encoded = item as usize;
        } else {
            (_, item) = get(meta_buf.as_slice(), metaloc, 4);
            bytes_encoded = item as usize;
            bytes_to_decode = bytes_encoded;
        }
        let got_data = read_bytes(input, 0, bytes_encoded as usize, in_buf.as_mut_slice())?;
        if !got_data {
            return Err(io::Error::new(io::ErrorKind::Other, "Bad data"));
        }
        let to_write;
        if freq_len > 0 {
            let freq = &freq_buf.as_slice()[..freq_len];
            let tree = build_huffman_tree(&freq);
            decompress_block(&tree, bytes_to_decode, &in_buf.as_slice()[..bytes_encoded], out_buf.as_mut_slice());
            to_write = &out_buf.as_slice()[..bytes_to_decode]
        } else {
            to_write = &in_buf.as_slice()[..bytes_to_decode]
        }
        // TODO: Can we write partial data?
        output.write(to_write)?;
    }
    Ok(())
}

fn decompress_block(tree: &Box<HuffTree>, bytes_to_decode: usize, in_buf: &[u8], out_buf: &mut [u8]) {
    let mut outptr = 0;
    let mut inptr = 0;
    let mut inbyte = 0u8;
    let mut inwidth = 0;
    let mut t = tree;
    loop {
        match (&t.left, &t.right) {
            (None, None) => {
                out_buf[outptr] = t.val;
                outptr += 1;
                if outptr == bytes_to_decode {
                    break
                }
                t = tree;
            }
            (&Some(ref zero), &Some(ref one)) => {
                if inwidth == 0 {
                    inbyte = in_buf[inptr];
                    inptr += 1;
                    inwidth = 8;
                }
                let bit = inbyte & 1;
                inbyte >>= 1;
                inwidth -= 1;
                t = if bit == 0 { zero } else { one };
            }
            _ => {
                panic!("Bad tree - should not happen")
            }
        }
    }
}

// Returns true if we got n bytes, false if we got zero bytes (orderly EOF), otherwise
// an error.

fn read_bytes(input: &mut dyn io::Read, atloc: usize, nbytes: usize, buf: &mut [u8]) -> io::Result<bool> {
    let mut bytes_read = 0;
    while bytes_read < nbytes {
        let n = input.read(&mut buf[atloc+bytes_read..atloc+nbytes])?;
        if n == 0 {
            if bytes_read == 0 {
                return Ok(false)
            }
            return Err(io::Error::new(io::ErrorKind::Other, "Premature EOF"));
        }
        bytes_read += n;
    }
    Ok(true)
}

// Encoding dictionary, mapping byte values to bit strings.  Only the byte values present in
// the tree will have valid entries in the dictionary.

#[derive(Clone,Copy)]
struct DictItem {
    bits: u64,      // the bit string, at most 56 bits, padded with zeroes
    width: usize,   // the number of valid bits
}

fn populate_dict(width: usize, bits: u64, tree: &Box<HuffTree>, dict: &mut [DictItem]) -> bool {
    match &tree.left {
        Some(_) => {
            return populate_dict(width+1, bits, &tree.left.as_ref().unwrap(), dict) &&
                   populate_dict(width+1, (1<<width)|bits, &tree.right.as_ref().unwrap(), dict)
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
// binary tree.  If left is not None, then right is also not None and val is invalid;
// otherwise, this is a leaf and val has the byte value.
//
// The priority queue used for building the tree must have a defined behavior when
// priorities are equal, or there can be no implementation-independent decoding.  To do this,
// we add a serial number to each node, and ties are broken with lower serial numbers first.
// For this to yield predictable trees, the input table of frequencies has to be sorted
// and has to be processed in order of increasing index.
//
// Also, the tree has to be built with the left (zero) branch always coming from the first
// node extracted and the right (one) branch coming from the second node.

struct HuffTree {
    left: Option<Box<HuffTree>>,
    right: Option<Box<HuffTree>>,
    val: u8
}

#[derive(Clone,Copy)]
struct Weight {
    serial: u32,
    weight: u32
}

fn greater_weight(a: Weight, b: Weight) -> bool {
    a.weight < b.weight || a.weight == b.weight && a.serial < b.serial
}

fn build_huffman_tree(freq: &[FreqEntry]) -> Box<HuffTree> {
    let mut priq = heap::Heap::<Box<HuffTree>, Weight>::new(greater_weight);
    let mut next_serial = 0u32;
    for i in freq {
        let t = Box::new(HuffTree { val: i.val, left: None, right: None });
        priq.insert(Weight{serial: next_serial, weight: i.count}, t);
        next_serial += 1;
    }
    while priq.len() > 1 {
        let (a, wa) = priq.extract_max();
        let (b, wb) = priq.extract_max();
        let t = Box::new(HuffTree { val: 0, left: Some(a), right: Some(b)});
        priq.insert(Weight{serial: next_serial, weight: wa.weight + wb.weight}, t);
        next_serial += 1;
    }
    priq.extract_max().0
}

// Byte frequency count.  The returned slice has counts for bytes with non-zero frequencies
// only, in descending stably sorted order.  The sorting is necessary for encoding as the order
// of the table can influence the relative priorities of nodes with equal weights during
// tree building, and also because the table is emitted into the compressed form and
// we want the output to be predictable.

#[derive(Clone,Copy)]
struct FreqEntry {
    val: u8,    // the byte value
    count: u32  // its count
}

fn compute_byte_frequencies<'a>(bytes: &[u8], freq: &'a mut [FreqEntry]) -> &'a [FreqEntry] {
    let mut i = 0;
    while i < 256 {
        freq[i].val = i as u8;
        freq[i].count = 0;
        i += 1;
    }

    for i in bytes {
        freq[*i as usize].count += 1;
    }

    // slice::sort_by is stable and will sort lower byte values before higher values,
    // for equal counts.
    freq.sort_by(|x, y| {
        if x.count > y.count {
            cmp::Ordering::Less
        } else if x.count < y.count {
            cmp::Ordering::Greater
        } else {
            cmp::Ordering::Equal
        }
    });

    i = 256;
    while i > 0 && freq[i-1].count == 0 {
        i -= 1;
    }
    &freq[0..i]
}

// Utilities

// Read n-byte value little-endian from stream at location p, return new location and
// the value read.

fn get(v: &[u8], mut p: usize, mut n: usize) -> (usize, u64) {
    let mut val : u64 = 0;
    let mut k = 0;
    while n > 0 {
        val = val | ((v[p] as u64) << k);
        k += 8;
        p += 1;
        n -= 1;
    }
    (p, val)
}

// Write n-byte value little-endian to stream, return location.

fn put(v: &mut [u8], mut p: usize, mut n: usize, mut val: u64) -> usize {
    while n > 0 {
        v[p] = val as u8;
        val >>= 8;
        n -= 1;
        p += 1;
    }
    p
}

