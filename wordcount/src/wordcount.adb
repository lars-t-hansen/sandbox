--  -*- indent-tabs-mode: nil -*-
--
--  Usage: wordcount [input-file]
--    Reads from stdin if there is no input-file

with Ada.Characters.Latin_1;
with Ada.Command_Line;
with Ada.Strings.Maps;
with Ada.Text_IO;

procedure Wordcount is
   package Nat_Text_IO is new Ada.Text_IO.Integer_IO (Natural);

   use Ada.Characters;
   use Ada.Command_Line;
   use Ada.Strings.Maps;
   use Ada.Text_IO;
   use Nat_Text_IO;

   type State_Type is (Outside, Inside);

   Named_File : File_Type;
   Chars : Natural := 0;
   Words : Natural := 0;
   Lines : Natural := 0;
   Has_File : Boolean := False;
   Whitespace : constant Character_Set :=
         To_Set (Latin_1.Space) or
         To_Set (Latin_1.LF) or
         To_Set (Latin_1.FF) or
         To_Set (Latin_1.CR) or
         To_Set (Latin_1.HT) or
         To_Set (Latin_1.VT);

   --  To match `wc`, a character is in a word if it is not whitespace.
   --  Not actually sure the whitespace set above completely matches, but
   --  it should be OK.

   function Is_Inside_Char (c : Character) return Boolean is
   begin
      return (not Is_In (c, Whitespace));
   end Is_Inside_Char;

   --  At_Eof and Read_Line appear to be necessary since I can't take the
   --  address of Named_File to unify named-file with Standard_Input.

   function At_Eof return Boolean is
   begin
      if Has_File then
         return End_Of_File (Named_File);
      end if;
      return End_Of_File (Standard_Input.all);
   end At_Eof;

   function Read_Line return String is
   begin
      if Has_File then
         return Get_Line (Named_File);
      end if;
      return Get_Line (Standard_Input.all);
   end Read_Line;

begin
   if Argument_Count >= 1 then
      Open (Named_File, In_File, Argument (1));
      Has_File := True;
   end if;
   while not At_Eof loop
      declare
         s : constant String := Read_Line;
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
   if Has_File then
      Close (Named_File);
   end if;
end Wordcount;
