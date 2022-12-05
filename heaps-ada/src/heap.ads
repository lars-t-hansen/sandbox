generic

   --  V is the type of values stored in the heap
   --
   --  Ada, like Go (and Java), does not seem to like to abstract over all types,
   --  it needs to know the type's basic structure.  Here, limit ourselves to
   --  access types, it's good enough.

   type R is limited private;
   type V is access R;

package Heap is

   type GreaterFn is access function (a, b: in V) return Boolean;

   type T is limited private;

   Heap_Empty, Heap_Full : exception;

   procedure Init (h : in out T; greater : GreaterFn);

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

   --  TODO: Heap has limited length.  At a minimum its range could be a parameter.
   type A is array (0 .. 50) of V;

   type T is limited record
      Greater : GreaterFn;
      Length  : Natural := 0;
      Items   : A;
   end record;

end Heap;
