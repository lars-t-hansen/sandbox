--  -*- indent-tabs-mode: nil -*-

--  Bounded single-threaded priority queue.

generic

   type V is private;
   --  FIXME: This is wrong, we're not interested in the range here.
   --  We're interested in the max size.  The lower limit of the range
   --  is always going to have to be zero, and this is not a constraint
   --  the client should worry about.
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
