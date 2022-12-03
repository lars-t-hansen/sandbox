with Ada.Text_IO;
with Heap;

procedure Heap_Test is
   package IIO is new Ada.Text_IO.Integer_IO (Integer);

   h : Heap.T;
   x : Integer;
   
begin
   Heap.Insert (h, 1);
   Heap.Insert (h, 2);
   Heap.Insert (h, 16);
   Heap.Insert (h, 8);
   Heap.Extract_Max (h, x);
   IIO.Put (x);
   Heap.Extract_Max (h, x);
   IIO.Put (x);
   Heap.Extract_Max (h, x);
   IIO.Put (x);
   Heap.Extract_Max (h, x);
   IIO.Put (x);
end Heap_Test;
