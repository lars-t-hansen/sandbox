--  -*- fill-column: 80; indent-tabs-mode: nil -*-

--  Huffman encoder, in Ada.
--
--  Usage:
--     huff [-o output-file] input-file
--
--  If there's no -o argument then the output file is input-file.huff.
--
--  Note, this program has only the encoding part, and currently only single-threaded.
--
--  TODO: Add at least a multithreaded encoder.
--
--  Since this is a programming exercise, it works by reading 64KB blocks and
--  compressing them individually; the output file consists of compressed blocks.
--  Also, we don't care about micro-efficiencies in representing the dictionary
--  in the file or in complicated fallback schemes, more could be done.
-- 
--  A compressed block is represented as
--    number of dictionary entries: u16 > 0 (max value is really 256)
--    run of dictionary entries sorted descending by frequency with ties
--    broken by lower byte values first:
--      value: u8
--      frequency: u32 (max value is really 65536)
--    number of encoded bytes: u32 (max value is really 65536)
--    number of bytes used for encoded bytes: u32 (max value 65536)
--    bytes, the number of which is encoded by previous field
-- 
--  An uncompressed block can be written under some circumstances, it is represented as
--    0: u16
--    number of bytes: u32 (really max 65536)
--    bytes, the number of which is encoded by previous field
--
-- The format has some problems; see the comment block in the Rust version.

with Interfaces; use Interfaces;
with Heap;
with Ada.Command_Line;
with Ada.Containers.Generic_Constrained_Array_Sort;
with Ada.Sequential_IO;
with Ada.Text_IO;
with Ada.Strings.Unbounded; use Ada.Strings.Unbounded;
with Ada.Unchecked_Deallocation;

procedure Huff_Ada is

   --  Cant_Encode is raised during encoding if internal limits are exceeded
   --  (bit string width; encoded size).  When raised, the compressor should just
   --  dump the input block verbatim on the output.

   Cant_Encode : exception;


   --  Block used to represent a byte sequence with a buffer and a varying
   --  length.

   type Buffer_Array is array (0 .. 65535) of Unsigned_8;

   type Buffer is limited record
      Length : Natural;
      It     : Buffer_Array;
   end record;


   --  Frequency computation.
   --
   --  On return from Compute_Frequencies, Freqs.Length has the number of
   --  non-zero elements in the table, and the table is sorted descending by
   --  count, with lower byte values breaking ties.

   type Freq_Item is record
      Ch    : Unsigned_8;
      Count : Natural;
   end record;

   type Freq_Array_Length is range 0 .. 256;
   type Freq_Array_Range is range 0 .. 255;
   type Freq_Array is array (Freq_Array_Range) of Freq_Item;

   type Freq_Table is limited record
      Length : Freq_Array_Length;
      It     : Freq_Array;
   end record;
 
   procedure Compute_Frequencies(Input : Buffer; Freqs : out Freq_Table) is

      function "<" (a, b: Freq_Item) return Boolean is
      begin
         return (a.Count > b.Count) or else (a.Count = b.Count and then a.Ch < b.Ch);
      end "<";

      procedure Sort_Freq is new Ada.Containers.
         Generic_Constrained_Array_Sort (Freq_Array_Range, Freq_Item, Freq_Array);

   begin
      pragma Assert (Input.Length > 0);
      for i in Freq_Array_Range loop
         Freqs.It (i).Ch := Unsigned_8 (i);
         Freqs.It (i).Count := 0;
      end loop;
      for i in 0 .. Input.Length-1 loop
         declare
            ix : constant Freq_Array_Range := Freq_Array_Range (Input.It (i));
         begin
            Freqs.It (ix).Count := Freqs.It (ix).Count + 1;
         end;
      end loop;
      Sort_Freq (Freqs.It);
      for ix in reverse Freq_Array_Range loop
         if Freqs.It (ix). Count /= 0 then
            Freqs.Length := Freq_Array_Length (ix) + 1;
            exit;
         end if;
      end loop;
   end Compute_Frequencies;


   --  Huffman tree construction.
   --
   --  In the Huff_Node, either Left and Right are both null, in which case
   --  Ch is valid, or they are both not null, in which case Ch is junk.
   --
   --  This will not raise any exceptions.  The tree will need to be freed
   --  once it's no longer needed.

   type Huff_Node;
   type Huff_Node_Ptr is access Huff_Node;
   type Huff_Node is limited record
      Ch          : Unsigned_8;
      Left, Right : Huff_Node_Ptr;
   end record;

   type Huff_Item is record
      Weight : Natural;
      Serial : Natural;
      Tree   : Huff_Node_Ptr;
   end record;

   function ">"(Left, Right : Huff_Item) return Boolean is
   begin
      return (Left.Weight < Right.Weight) or else
             (Left.Weight = Right.Weight and then Left.Serial < Right.Serial);
   end ">";

   procedure Build_Huffman_Tree(Freqs : Freq_Table; Tree : out Huff_Node_Ptr) is

     package HuffHeap is new Heap (Huff_Item);
     
     Priq : HuffHeap.T;
     Serial : Natural := 0;

   begin
      pragma Assert (Freqs.Length > 0);
      for i in 0 .. Freqs.Length - 1 loop
         declare
            ix : constant Freq_Array_Range := Freq_Array_Range (i);
         begin
            HuffHeap.Insert (Priq, Huff_Item'(Freqs.It (ix).Count,
                                              Serial,
                                              new Huff_Node'(Freqs.It (ix).Ch, null, null)));
            Serial := Serial + 1;
         end;
      end loop;
      while HuffHeap.Length (Priq) > 1 loop
         declare
            a, b : Huff_Item;
         begin
            HuffHeap.Extract_Max (Priq, a);
            HuffHeap.Extract_Max (Priq, b);
            HuffHeap.Insert (Priq, Huff_Item'(a.Weight + b.Weight,
                                              Serial,
                                              new Huff_Node'(0, a.Tree, b.Tree)));
            Serial := Serial + 1;
         end;
      end loop;
      declare
         it : Huff_Item;
      begin
         HuffHeap.Extract_Max (Priq, it);
         Tree := it.Tree;
      end;
   end Build_Huffman_Tree;

   procedure Free_Huff_Node is new Ada.Unchecked_Deallocation (Huff_Node, Huff_Node_Ptr);

   procedure Free_Huffman_Tree (Tree : in out Huff_Node_Ptr) is
   begin
      if Tree.Left /= null then
         Free_Huffman_Tree (Tree.Left);
         Free_Huffman_Tree (Tree.Right);
      end if;
      Free_Huff_Node (Tree);
   end Free_Huffman_Tree;

   --  Dictionary construction.
   --
   --  When Build_Dictionary returns normally, the items in Dict that correspond
   --  to non-zero frequencies in the tree will have valid entries.  Everything
   --  else will be garbage.
   --
   --  When Build_Dictionary runs out of bits in the encoding representation it
   --  raises Cant_Encode.

   type Dict_Item is limited record
      Bits  : Unsigned_64;
      Width : Natural;
   end record;
   
   type Dict_Range is range 0 .. 255;
   type Dictionary is array (Dict_Range) of Dict_Item;

   procedure Build_Dictionary(Tree : Huff_Node_Ptr; Dict : out Dictionary) is

      procedure Descend (Tree : Huff_Node_Ptr; Bits : Unsigned_64; Width : Natural) is
      begin
         if Tree.Left = null then
            pragma Assert (Tree.Right = null);
            --  This limit ensures that a 64-bit window will never overflow
            --  during encoding.
            if Width > 56 then
               raise Cant_Encode;
            end if;
            Dict (Dict_Range (Tree.Ch)).Width := Width;
            Dict (Dict_Range (Tree.Ch)).Bits := Bits;
         else
            Descend(Tree.Left, Bits, Width + 1);
            Descend(Tree.Right, Shift_Left(1, Width) or Bits, Width + 1);
         end if;
      end Descend;

   begin
      Descend (Tree, 0, 0);
   end Build_Dictionary;


   --  Block encoder.
   --
   --  When Encode_Block returns normally, Output.Length will have the correct
   --  encoded length.
   --
   --  When Encode_Block runs out of space, it raises Cant_Encode.

   procedure Encode_Block (Input : Buffer; Output : out Buffer; Dict : Dictionary) is
      Outptr : Natural := 0;
      Inptr  : Natural := 0;
      Bits   : Unsigned_64 := 0;
      Width  : Natural := 0;
      dix    : Dict_Range;
   begin
      while Inptr < Input.Length loop
         dix := Dict_Range (Input.It (Inptr));
         Bits := Bits or Shift_Left(Dict (dix).Bits, Width);
         Width := Width + Dict (dix).Width;
         Inptr := Inptr + 1;
         while Width >= 8 loop
            if Outptr = Output.It'Length then
               raise Cant_Encode;
            end if;
            Output.It (Outptr) := Unsigned_8 (Bits and 255);
            Outptr := Outptr + 1;
            Bits := Shift_Right (Bits, 8);
            Width := Width - 8;
         end loop;
      end loop;
      if Width > 0 then
         if Outptr = Output.It'Length then
            raise Cant_Encode;
         end if;
         Output.It (Outptr) := Unsigned_8 (Bits);
         Outptr := Outptr + 1;
      end if;
      Output.Length := Outptr;
   end Encode_Block;


   -- The MetaBuffer holds an encoding of the frequency table and some other
   -- crucial data.

   type Metadata is array (0 .. 2 + 5*256 + 2*4) of Unsigned_8;

   type MetaBuffer is limited record
      Length : Natural;
      It     : Metadata;
   end record;

   procedure Put_8(A : in out MetaBuffer; V : Unsigned_8) is
   begin
      A.It (A.Length) := V;
      A.Length := A.Length + 1;
   end Put_8;

   procedure Put_16(A : in out MetaBuffer; V : Unsigned_16) is
   begin
      A.It (A.Length) := Unsigned_8 (V and 255);
      A.It (A.Length + 1) := Unsigned_8 (Shift_Right(V, 8) and 255);
      A.Length := A.Length + 2;
   end Put_16;

   procedure Put_32(A : in out MetaBuffer; V : Unsigned_32) is
   begin
      A.It (A.Length) := Unsigned_8 (V and 255);
      A.It (A.Length + 1) := Unsigned_8 (Shift_Right(V, 8) and 255);
      A.It (A.Length + 2) := Unsigned_8 (Shift_Right(V, 16) and 255);
      A.It (A.Length + 3) := Unsigned_8 (Shift_Right(V, 24) and 255);
      A.Length := A.Length + 4;
   end Put_32;


   -- Block compressor.
   --
   -- When Compress_Block returns, the Output and the Metadata together
   -- hold the compressed representation.  No exceptions are raised.

   procedure Compress_Block (Input : Buffer; Output : out Buffer; Meta : out MetaBuffer) is
      Freqs : Freq_Table;
      Dict  : Dictionary;
      Tree  : Huff_Node_Ptr := null;

   begin
      begin
         Compute_Frequencies (Input, Freqs);
         Build_Huffman_Tree (Freqs, Tree);
         Build_Dictionary (Tree, Dict);
         Encode_Block (Input, Output, Dict);

         Meta.Length := 0;
         Put_16 (Meta, Unsigned_16 (Freqs.Length));
         for i in 0 .. Freqs.Length-1 loop
            Put_8 (Meta, Freqs.It (Freq_Array_Range (i)).Ch);
            Put_32 (Meta, Unsigned_32 (Freqs.It (Freq_Array_Range (i)).Count));
         end loop;
         Put_32 (Meta, Unsigned_32 (Input.Length));
         Put_32 (Meta, Unsigned_32 (Output.Length));

      exception
      when Cant_Encode =>
         Meta.Length := 0;
         Output.It (0 .. Input.Length-1) := Input.It (0 .. Input.Length-1);
         Output.Length := Input.Length;
         Put_16 (Meta, 0);
         Put_32 (Meta, Unsigned_32 (Output.Length));

      end;
      Free_Huffman_Tree (Tree);
   end Compress_Block;


   -- File compressor.

   procedure Compress_File (Input_Name, Output_Name : String) is
      package FIO is new Ada.Sequential_IO (Unsigned_8);
      Input_File : FIO.File_Type;
      Output_File : FIO.File_Type;
      Input, Output : Buffer;
      Meta : MetaBuffer;

   begin
      FIO.Open(Input_File, FIO.In_File, Input_Name);
      FIO.Create(Output_File, FIO.Out_File, Output_Name);
      while not FIO.End_Of_File(Input_File) loop
         --  TODO: This is completely tragic.  Surely there has got to be a better way than
         --  byte-at-a-time?  It looks like stream I/O might work, after a fashion, but
         --  not yet sure how to do that.
         Input.Length := Input.It'Last;
         for i in 0 .. Input.It'Last loop
            if FIO.End_Of_File(Input_File) then
               Input.Length := i;
               exit;
            end if;
            FIO.Read(Input_File, Input.It (i));
         end loop;
         Compress_Block (Input, Output, Meta);
         --  TODO: Ditto.
         for i in 0 .. Meta.Length-1 loop
            FIO.Write(Output_File, Meta.It (i));
         end loop;
         for i in 0 .. Output.Length-1 loop
            FIO.Write(Output_File, Output.It (i));
         end loop;
      end loop;
      FIO.Close(Output_File);
      FIO.Close(Input_File);
   end Compress_File;

   procedure Parse_Arguments (Input_Name, Output_Name: in out Unbounded_String;
                              Failed : in out Boolean) is
      use Ada.Command_Line;
      use Ada.Text_IO;

      Next_Arg : Natural := 1;
      Explicit_Output : Boolean := False;
   begin
      if Next_Arg <= Argument_Count and then Argument (Next_Arg) = "-o" then
         Next_Arg := Next_Arg + 1;
         if Next_Arg <= Argument_Count then
            Output_Name := To_Unbounded_String(Argument (Next_Arg));
            Explicit_Output := True;
            Next_Arg := Next_Arg + 1;
         else
            --  Output file required after -o
            goto Borked;
         end if;
      end if;
      if Next_Arg <= Argument_Count then
         Input_Name := To_Unbounded_String(Argument (Next_Arg));
         Next_Arg := Next_Arg + 1;
      else
         --  Input file required
         goto Borked;
      end if;
      if Next_Arg <= Argument_Count then
         --  Unconsumed arguments
         goto Borked;
      end if;
      if not Explicit_Output then
         Output_Name := Input_Name & ".huff";
      end if;
      return;

   <<Borked>>
      Put_Line ("Usage: huff [-o output-file] input-file");
      Set_Exit_Status (1);
      Failed := True;
   end Parse_Arguments;

   Failed : Boolean := False;
   Input_Name : Unbounded_String;
   Output_Name : Unbounded_String;

begin
   Parse_Arguments (Input_Name, Output_Name, Failed);
   if Failed then
      return;
   end if;
   Compress_File (To_String(Input_Name), To_String(Output_Name));
end Huff_Ada;
