{ -*- mode: fundamental; indent-tabs-mode: nil -*- }

{ SKETCHES }

program huff(infile, outfile);

const
  BufSize = 65536;  { what every other version uses }
  MetaSize = 1290;  { 2 + (5*256) + 2*4, the max needed }

type
  Buffer = packed array [1..BufSize] of char;
  MetaBuf = packed array [1..MetaSize] of char;

  FreqItem = record
               byte, count : integer;
             end;

  DictEntry = record
                bits, width : integer;
              end;

  { Indexed by byte value }
  FreqBuffer = array [0..255] of FreqItem;
  Dictionary = array [0..255] of DictEntry;

  HuffTree = ^HuffNode;
  HuffNode = record
               byte: integer;
               left, right: HuffTree
             end;

  { For the generic heap code.  The max heap size is 10000: the store is a tree of 100 slots
    for arrays of length 100.  These values are pretty arbitrary but will exercise the code a
    bit more. }
  HeapNode = record
               weight, serial: integer;
               tree: integer
             end;
  HeapStruct = record
                 size: integer;
                 {private} serial: integer;
                 { No variable-length arrays in Pascal, one could simulate with a b-tree }
                 {private} store: array [1..100] of HeapNode;
               end;
  Hean = ^HeapStruct;

var
  infile  : file of char;
  outfile : file of char;

{- Generic I/O ------------------------------------------------------------------------------------}

procedure ReadBuffer(var {out} input: packed array [alow..ahigh] of char; var {out} nread : integer);
  var
    i : integer;
  begin
    i := alow;
    while i <= ahigh && !eof(infile) do begin
      input[i] := infile^;
      get(infile);
      i := i+1
    end;
    nread := i - alow
  end;

procedure WriteBuffer(var {in} output : packed array [alow..ahigh] of char; nwrite : integer);
  var
    i : integer;
  begin
    i := 0;
    while i < nwrite do begin
      outfile^ := output[i+alow];
      put(outfile);
      i := i+1
    end
  end;

{- Generic sort ------------------------------------------------------------------------------------}
{ The sequence being sorted is abstracted away by the high/low and the less/swap code.              }

procedure Sort(low, high: integer;
               less : function(i, j: integer): boolean;
               swap : procedure(i, j: integer));
  var
    i, j : integer;
  begin
    for i := low to high-1 do
      for j := i+1 to high do
        if less(i, j) then
          swap(i, j)
      end
    end
  end;

{- Generic heap ------------------------------------------------------------------------------------}
{ The trick here is that the data in the heap are represented by integers.  The client code must    }
{ maintain a mapping from the integer to the data.                                                  }

function NewHeap() : Heap;
  var
    h: Heap;
  begin
    new(h);
    NewHeap := h
  end;

procedure HeapInsert(h: Heap; weight, data: integer);

  procedure push(weight, data: integer);
    var
      loc, serial: integer;
    begin
      serial := h^.serial;
      h^.serial := serial + 1;
      loc := h^.size;
      if loc >= h^.storeSize then begin
        { FIXME }
      end;
      h^.size := loc + 1;
      { TODO: Grab the correct store array here based on loc, or use some indexing proc }
      h^.store[loc].weight := weight;
      h^.store[loc].serial := serial;
      h^.store[loc].data := data
    end;

  begin
  end;

procedure HeapExtractMax(h: Heap; var {out} weight, data: integer);

  procedure heapifyAtZero;
    begin
    end;

  begin
  end;

procedure HeapSwap(h: Heap; i, j: integer);
  begin
  end;

{- Frequency table ---------------------------------------------------------------------------------}

{ Build a frequency table for all the bytes and return the number of nonzero entries }
procedure ComputeFrequencies(var {in} input : Buffer;
                             inputlen : integer;
                             var {inout} freq : FreqBuffer;
                             var {out} numfreq : integer);
  var
    i, k : integer;

  { Sort into descending order by count, and ascending by byte value on ties }
  procedure SortFreq;

    function FreqLess(i, j: integer): boolean;
      begin
        FreqLess := freq[j].count > freq[i].count or
                    freq[i].count = freq[j].count and freq[i].byte > freq[j].byte
      end;

    procedure FreqSwap(i, j);
      var
        tmp: FreqItem;
      begin
        tmp := freq[i];
        freq[i] := freq[j];
        freq[j] := tmp
      end;

    begin
      Sort(0, 255, FreqLess, FreqSwap)
    end;

  begin
    for i := 0 to 255 do begin
      freq[i].byte := i;
      freq[i].count := 0
    end;
    for i := 1 to inputlen do begin
      k := ord(input[i]);
      freq[k] := freq[k] + 1
    end;
    SortFreq;
    i := 256;
    while freq[i-1].count = 0 do
      i := i - 1;
    numfreq := i
  end;

{- Huffman tree construction ----------------------------------------------------------------------}

function BuildHuffmanTree(var {in} freqbuf : FreqBuffer; freqlen : integer): HuffTree;
  type
    { probably not right }
    PqTree = ^PqNode;
    PqNode = record
               weight, serial : integer;
               tree : PqTree
             end;

  begin
  end;

{- Compressor -------------------------------------------------------------------------------------}

procedure EncodeBlock(var input  : packed array [ilow..ihigh] of char;
                      inputlen   : integer;
                      var output : packed array [olow..ohigh] of char;
                      var outptr : integer;
                      var d      : Dictionary);
  var
    inptr, bits, widthMul, outputlen : integer;
    dix : Dictentry;
  begin
    inptr := 0
    outptr := 0;
    bits := 0;
    widthMul := 1; { 2^0 }
    while inptr < inputlen do begin
      dix := d[input[inptr]];
      inptr := inptr + 1;
      bits := bits + (dix.bits * widthMul);
      widthMul := widthMul * dix.width;
      while widthMul >= 256 do begin
         if outptr = BufSiz then begin
            outptr := -1;
            goto End
         end;
         output[outptr] := bits mod 256;
         outptr := outptr + 1;
         bits := bits div 256;
         widthMul := widthMul div 256;
      end;
    end;
    if widthMul > 1 then begin
      { fixme }
    end;
  End:
    ;
  end; { EncodeBlock }

procedure CompressBlock();
  var
    freqlen : integer;
  begin
    ComputeFrequencies(inbuf, inputlen, freqbuf, freqlen);
    tree := BuildHuffmanTree(freqbuf, freqlen);
    if BuildDictionary(tree, dictbuf) then begin
    end
  end;

procedure CompressFile;
  var
    inbuf    : Buffer;
    inputlen : integer;
    outbuf   : Buffer;
    outlen   : integer;
    metabuf  : MetaBuf;
    metalen  : integer;
    freqbuf  : FreqBuffer;
    dictbuf  : Dictionary;
  begin
    while true do begin
      ReadBuffer(inbuf, inputlen);
      if inputlen = 0 then
        goto Done;
      CompressBlock(freqbuf, dictbuf, inbuf, inputlen, outbuf, outlen, metabuf, metalen);
      { This is where Pascal level 0 fails utterly: can't abstract over array length }
      WriteBuffer(metabuf, metalen);
      WriteBuffer(outbuf, outlen)
    end
  Done:;
  end;

begin
  CompressFile;
end.
