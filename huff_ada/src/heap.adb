--  -*- indent-tabs-mode: nil -*-

package body Heap is

   procedure Swap (h : in out T; x, y : QueueRange) is
      tmp : V;
   begin
      tmp := h.Items (x);
      h.Items (x) := h.Items (y);
      h.Items (y) := tmp;
   end Swap;

   function Length (h : T) return Natural is
   begin
      return h.Length;
   end Length;

   function Parent (loc : QueueRange) return QueueRange is
   begin
      return QueueRange ((Natural(loc) - 1) / 2);
   end Parent;

   function Left (loc : QueueRange) return QueueRange is
   begin
      return QueueRange ((Natural(loc) * 2) + 1);
   end Left;

   function Right (loc : QueueRange) return QueueRange is
   begin
      return QueueRange ((Natural(loc) + 1) * 2);
   end Right;

   procedure Heapify (h : in out T; loc_param : QueueRange) is
      greatest, l, r, loc : QueueRange;
   begin
      loc := loc_param;
      loop
         greatest := loc;
         l := Left (loc);
         if Natural(l) < h.Length and then h.Items (l) > h.Items (greatest) then
            greatest := l;
         end if;
         r := Right (loc);
         if Natural(r) < h.Length and then h.Items (r) > h.Items (greatest) then
            greatest := r;
         end if;
         if greatest = loc then
            exit;
         end if;
         Swap (h, loc, greatest);
         loc := greatest;
      end loop;
   end Heapify;

   procedure Insert (h : in out T; x : V) is
      i : QueueRange;
   begin
      --  FIXME: This used to have a test for overflow.  That was removed
      --  when the representation was changed to use Vector.  Now that it
      --  is back to being bounded, the test needs to come back.
      h.Length := h.Length + 1;
      h.Items (QueueRange(h.Length - 1)) := x;
      i := QueueRange (h.Length - 1);
      while i > 0 and then h.Items (i) > h.Items (Parent (i)) loop
         Swap (h, i, Parent (i));
         i := Parent (i);
      end loop;
   end Insert;

   procedure Peek_Max (h : T; elt : out V) is
   begin
      if h.Length = 0 then
         raise Heap_Empty;
      end if;
      elt := h.Items (0);
   end Peek_Max;

   procedure Extract_Max (h : in out T; elt : out V) is
   begin
      if h.Length = 0 then
         raise Heap_Empty;
      end if;
      elt := h.Items (0);
      h.Items (0) := h.Items (QueueRange (h.Length - 1));
      h.Length := h.Length - 1;
      Heapify (h, 0);
   end Extract_Max;

end Heap;
