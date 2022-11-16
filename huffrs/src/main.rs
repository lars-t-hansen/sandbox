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
//
// This format has a couple of bugs:
//
// - since there will be no dictionary entries whose frequency is zero, the frequency
//   value can be encoded as f-1, so that a 16-bit field is sufficient.
// - we have to perform three read operations to read a compressed block: first two
//   bytes to get the metadata length (the dictionary size), then the metadata to
//   get the size of the encoded block, then the encoded block.  This would more
//   sensibly be encoded as length-of-metadata-and-data (4 bytes) followed by data,
//   and a single read operation would get both metadata and data.

mod heap;

use std::fs::File;
use std::{cmp,env,io,process};

#[derive(PartialEq)]
enum Op {
    Compress,
    Decompress
}

fn main() {
    let (op, in_filename, out_filename) = match parse_args() {
        Ok(x) => x,
        Err(e) => { 
            eprintln!("Error: {}", e);
            process::exit(1);
        }
    };
    let res = if op == Op::Compress {
        compress_file(in_filename, out_filename)
    } else {
        decompress_file(in_filename, out_filename)
    };
    if res.is_err() {
        eprintln!("Error: {}", res.err().unwrap());
        process::exit(1);
    }
}

fn parse_args() -> Result<(Op, /* in_filename */ String, /* out_filename */ String), String> {
    let usage = "Usage: huffrs (compress|decompress) [-o outfile] infile".to_string();
    let mut args = env::args();

    // Infer operation from command name, maybe, otherwise look for verb
    let op = match args.next().unwrap_or("".to_string()).as_str() {
        "huff" => Op::Compress,
        "puff" => Op::Decompress,
        _ => match args.next().unwrap_or("".to_string()).as_str() {
                "compress" => Op::Compress,
                "decompress" => Op::Decompress,
                _ => {
                    return Err(usage);
                }
            }
    };

    // Parse remaining arguments
    let mut out_filename = String::from("");
    let mut have_out_filename = false;
    let mut n = args.next();
    if !n.is_some() {
        return Err(usage);
    }
    if n.as_ref().unwrap() == "-o" {
        n = args.next();
        if !n.is_some() {
            return Err(usage);
        }
        out_filename = n.unwrap();
        n = args.next();
        have_out_filename = true;
    }
    let in_filename = n.unwrap();
    n = args.next();
    if !n.is_none() {
        return Err(usage);
    }

    if op == Op::Decompress && !in_filename.ends_with(".huff") {
       // TODO: Also must check that filename is not empty after stripping extension
        return Err("Input file must have extension .huff".to_string())
    }
    if !have_out_filename {
        out_filename = String::from(in_filename.as_str());
        if op == Op::Compress {
            out_filename.push_str(".huff")
        } else {
            _ = out_filename.split_off(out_filename.len() - 5);
        }
    }

    Ok((op, in_filename, out_filename))
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
        metaloc = put_u16(meta_buf, metaloc, freq.len() as u16);
        for item in freq {
            metaloc = put_u8(meta_buf, metaloc, item.val);
            metaloc = put_u32(meta_buf, metaloc, item.count);
        }
        metaloc = put_u32(meta_buf, metaloc, bytes_read as u32);
        metaloc = put_u32(meta_buf, metaloc, bytes_encoded as u32);
        output = &out_buf[0..bytes_encoded];
    } else {
        metaloc = put_u16(meta_buf, metaloc, 0u16);
        metaloc = put_u32(meta_buf, metaloc, bytes_read as u32);
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
        let (_, item) = get_u16(meta_buf.as_slice(), 0);
        let freq_len = item as usize;
        let metabytes = if freq_len > 0 { 5*freq_len + 8 } else { 4 };
        let got_metadata = read_bytes(input, 2, metabytes, meta_buf.as_mut_slice())?;
        if !got_metadata {
            return Err(io::Error::new(io::ErrorKind::Other, "Bad metadata"));
        }
        let (freq, bytes_encoded, bytes_to_decode) = decode_metadata(&meta_buf.as_slice()[2..], freq_len, freq_buf.as_mut_slice());
        let got_data = read_bytes(input, 0, bytes_encoded as usize, in_buf.as_mut_slice())?;
        if !got_data {
            return Err(io::Error::new(io::ErrorKind::Other, "Bad data"));
        }
        let out_data = decode_block(freq, bytes_to_decode, &in_buf.as_slice()[..bytes_encoded], out_buf.as_mut_slice());
        write_bytes(output, out_data)?;
    }
    Ok(())
}

// This will return a freq of length zero if there is no decoding to be done.

fn decode_metadata<'a>(metadata: &[u8], freq_len: usize, freq_buf: &'a mut [FreqEntry]) -> 
        (/* freq */ &'a [FreqEntry], /* bytes_encoded */ usize, /* bytes_to_decode */ usize) {
    let bytes_encoded;
    let bytes_to_decode;
    let mut freq = &mut freq_buf[..freq_len];
    let mut metaloc = 0;
    if freq_len > 0 {
        let mut i = 0;
        while i < freq_len {
            (metaloc, freq[i].val) = get_u8(metadata, metaloc);
            (metaloc, freq[i].count) = get_u32(metadata, metaloc);
            i += 1;
        }
        let mut item : u32;
        (metaloc, item) = get_u32(metadata, metaloc);
        bytes_to_decode = item as usize;
        (_, item) = get_u32(metadata, metaloc);
        bytes_encoded = item as usize;
    } else {
        let (_, item) = get_u32(metadata, metaloc);
        bytes_encoded = item as usize;
        bytes_to_decode = bytes_encoded;
    }
    (freq, bytes_encoded, bytes_to_decode)
}

fn decode_block<'a>(freq: &[FreqEntry], bytes_to_decode: usize, in_buf: &'a [u8], out_buf: &'a mut [u8]) -> &'a [u8] {
    if freq.len() > 0 {
        let tree = build_huffman_tree(&freq);
        decompress_block(&tree, bytes_to_decode, in_buf, out_buf);
        return &out_buf[..bytes_to_decode]
    }
    assert!(bytes_to_decode == in_buf.len());
    in_buf
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

#[derive(Clone,Copy,PartialEq)]
struct Weight {
    serial: u32,
    weight: u32
}

impl PartialOrd for Weight  {
    fn partial_cmp(&self, other: &Self) ->  Option<cmp::Ordering> {
        Some(if self.weight < other.weight {
            cmp::Ordering::Greater  // smaller values have higher priorities
        } else if self.weight > other.weight {
            cmp::Ordering::Less     // and vice versa
        } else if self.serial < other.serial {
            cmp::Ordering::Greater  // smaller serial numbers have higher priorities
        } else if self.serial > other.serial {
            cmp::Ordering::Less     // and vice versa
        } else {
            cmp::Ordering::Equal
        })
    }
}

fn build_huffman_tree(freq: &[FreqEntry]) -> Box<HuffTree> {
    let mut priq = heap::Heap::<Box<HuffTree>, Weight>::new();
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

// Read value little-endian from stream at location p, return new location and
// the value read.

fn get_u8(v: &[u8], p: usize) -> (usize, u8) {
    (p+1, v[p])
}

fn get_u16(v: &[u8], p: usize) -> (usize, u16) {
    (p+2, ((v[p+1] as u16) << 8) | (v[p] as u16))
}

fn get_u32(v: &[u8], p: usize) -> (usize, u32) {
    (p+4, ((v[p+3] as u32) << 24) | ((v[p+2] as u32) << 16) | ((v[p+1] as u32) << 8) | (v[p] as u32))
}

// Write value little-endian to slice at position, return new position.

fn put_u8(v: &mut [u8], p: usize, val: u8) -> usize {
    v[p] = val;
    p+1
}

fn put_u16(v: &mut [u8], p: usize, val: u16) -> usize {
    v[p] = val as u8;
    v[p+1] = (val >> 8) as u8;
    p+2
}

fn put_u32(v: &mut [u8], p: usize, val: u32) -> usize {
    v[p] = val as u8;
    v[p+1] = (val >> 8) as u8;
    v[p+2] = (val >> 16) as u8;
    v[p+3] = (val >> 24) as u8;
    p+4
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

// Try hard to write the entire slice to the output, signal error if we can't do it.

const MAX_RETRIES : usize = 1;

fn write_bytes(output: &mut dyn io::Write, out_data: &[u8]) -> io::Result<()> {
    let mut written = 0;
    let mut no_progress = 0;
    while written < out_data.len() {
        let n = output.write(&out_data[written..])?;
        if n == 0 {
            if no_progress > MAX_RETRIES {
                return Err(io::Error::new(io::ErrorKind::Other, "Could not write"));
            }
            no_progress += 1;
            continue;
        }
        written += n;
        no_progress = 0;
    }
    Ok(())
}
