--  -*- indent-tabs-mode: nil -*-

with Ada.Text_IO;
with Ada.Strings.Maps;
with Ada.Characters.Latin_1;

procedure Wordcount is
   package Integer_Text_IO is new Ada.Text_IO.Integer_IO (Integer);

   use Ada.Text_IO;
   use Ada.Characters;
   use Ada.Strings.Maps;
   use Integer_Text_IO;

   type State_Type is (Outside, Inside);

   f : File_Type;
   Chars : Integer := 0;
   Words : Integer := 0;
   Lines : Integer := 0;
   Whitespace : constant Character_Set := "or" (To_Set (Latin_1.Space), To_Set (Latin_1.HT));

   --  To match `wc`, a character is in a word if it is not whitespace.

   function Is_Inside_Char (c : Character) return Boolean is
   begin
      return (not Is_In (c, Whitespace));
   end Is_Inside_Char;

begin
   Open (f, In_File, "test.txt");
   while not End_Of_File (f) loop
      declare
         s : constant String := Get_Line (f);
         State : State_Type := Outside;
      begin
         Lines := Lines + 1;
         Chars := Chars + s'Length + 1;
         for j in s'Range loop
            if State = Outside and then Is_Inside_Char (s (j)) then
               Words := Words + 1;
               State := Inside;
            elsif State = Inside and then not Is_Inside_Char (s (j)) then
               State := Outside;
            end if;
         end loop;
      end;
   end loop;
   Put (Lines);
   Put (Words);
   Put (Chars);
   New_Line;
   Close (f);
end Wordcount;
