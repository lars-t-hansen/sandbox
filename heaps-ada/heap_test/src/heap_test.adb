with Ada.Text_IO;
with Heap;

procedure Heap_Test is
   package IIO is new Ada.Text_IO.Integer_IO (Integer);

   type MyObj is limited record
      weight : Integer;
      payload : Integer;
   end record;

   type MyObjPtr is access MyObj;

   function ">" (a, b : MyObjPtr) return Boolean is
   begin
      return a.weight > b.weight;
   end ">";

   package ObjHeap is new Heap (MyObjPtr);

   hh : ObjHeap.T;
   xx : MyObjPtr;

   package IntHeap is new Heap (Integer);

   h : IntHeap.T;
   x : Integer;

begin
   IntHeap.Insert (h, 1);
   IntHeap.Insert (h, 2);
   IntHeap.Insert (h, 16);
   IntHeap.Insert (h, 8);
   IntHeap.Extract_Max (h, x);
   IIO.Put (x);
   IntHeap.Extract_Max (h, x);
   IIO.Put (x);
   IntHeap.Extract_Max (h, x);
   IIO.Put (x);
   IntHeap.Extract_Max (h, x);
   IIO.Put (x);

   ObjHeap.Insert (hh, new MyObj'(1, 10));
   ObjHeap.Insert (hh, new MyObj'(2, 20));
   ObjHeap.Insert (hh, new MyObj'(16, 160));
   ObjHeap.Insert (hh, new MyObj'(8, 80));
   ObjHeap.Extract_Max (hh, xx);
   IIO.Put (xx.payload);
   ObjHeap.Extract_Max (hh, xx);
   IIO.Put (xx.payload);
   ObjHeap.Extract_Max (hh, xx);
   IIO.Put (xx.payload);
   ObjHeap.Extract_Max (hh, xx);
   IIO.Put (xx.payload);
end Heap_Test;
