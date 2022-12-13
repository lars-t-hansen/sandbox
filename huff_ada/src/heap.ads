--  -*- indent-tabs-mode: nil -*-

--  Bounded single-threaded priority queue.

generic

   type V is private;
   type QueueRange is range <>;

   with function ">" (Left, Right : V) return Boolean is <>;

package Heap is

   type T is limited private;

   Heap_Empty, Heap_Full : exception;

   --  Insert x into the heap.
   --  Raises Heap_Full if the heap is full.
   procedure Insert (h : in out T; x : V);

   --  Peek at the maximum element without removing it.
   --  Raises Heap_Empty if the heap is empty.
   procedure Peek_Max (h : T; elt : out V);

   --  Extract the maximum element.
   --  Raises Heap_Empty if the heap is empty.
   procedure Extract_Max (h : in out T; elt : out V);

   --  Return the number of elements in the heap.
   function Length (h : T) return Natural;

private

   type A is array (QueueRange) of V;

   type T is limited record
      Length  : Natural := 0;
      Items   : A;
   end record;

end Heap;
