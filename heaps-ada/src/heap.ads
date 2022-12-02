--  TODO: Heap has integer values only, and these are both the elements and
--  the weights.  Really want to make this generic.

package Heap is

   type T is limited private;

   Heap_Empty, Heap_Full : exception;

   --  Insert x into the heap.
   --  Raises Heap_Full if the heap is full.
   procedure Insert (h : in out T; x : Integer);

   --  Peek at the maximum element without removing it.
   --  Raises Heap_Empty if the heap is empty.
   procedure Peek_Max (h : T; elt : out Integer);

   --  Extract the maximum element.
   --  Raises Heap_Empty if the heap is empty.
   procedure Extract_Max (h : in out T; elt : out Integer);

   --  Return the number of elements in the heap.
   function Length (h : T) return Natural;

private

   --  TODO: Heap has limited length
   type A is array (0 .. 50) of Integer;

   type T is limited record
      Length : Natural := 0;
      Items  : A;
   end record;

end Heap;
