with Ada.Text_IO;
with Heap;

procedure Heap_Test is
   package IIO is new Ada.Text_IO.Integer_IO (Integer);

   type MyObj is limited record
      weight : Integer;
      payload : Integer;
   end record;

   type MyObjPtr is access MyObj;
   package ObjHeap is new Heap (MyObj, MyObjPtr);

   function ObjGreater (a, b : MyObjPtr) return Boolean is
   begin
      return a.weight > b.weight;
   end ObjGreater;

   hh : ObjHeap.T;
   xx : MyObjPtr;

begin
   ObjHeap.Init (hh, ObjGreater'Access);
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
