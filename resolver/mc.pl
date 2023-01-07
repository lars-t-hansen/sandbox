% -*- mode: prolog -*-

% state(MLB, CLB, MB, CB, MRB, CRB, States)
% where M is Missionaries, C is Cannibals, LB is left bank, B is boat, RB is right bank
% States is a list of states we've gone through to get here

%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%
% Constraints on new states.

% There must be no more cannibals in a place than missionaries

constrain_place(left, MLB, CLB, MB, CB, MRB, CRB) :-
    M is MLB + MB,
    C is CLB + CB,
    M >= C,
    MRB >= CRB.

constrain_place(right, MLB, CLB, MB, CB, MRB, CRB) :-
    M is MB + MRB,
    C is CB + CRB,
    M >= C,
    MLB >= CLB.

constrain_place(enroute, _, _, MB, CB, _, _) :-
    MB >= CB.

% The boat holds only two no matter where it is

constrain_boat(_, _, _, MB, CB, _, _) :-
    B is MB + CB,
    B =< 2.

% The boat can't move on its own

constrain_sailing(enroute, _, _, MB, CB, _, _) :-
    B is MB + CB,
    B > 0.

constrain_sailing(left, _, _, _, _, _, _) :- true.
constrain_sailing(right, _, _, _, _, _, _) :- true.

constrain(P, MLB, CLB, MB, CB, MRB, CRB, _) :-
    constrain_place(P, MLB, CLB, MB, CB, MRB, CRB),
    constrain_boat(P, MLB, CLB, MB, CB, MRB, CRB),
    constrain_sailing(P, MLB, CLB, MB, CB, MRB, CRB).

%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%
% Legal moves

% Move one person from the left bank into the boat.

move(left, MLB, CLB, MB, CB, MRB, CRB, S) :-
    MLB > 0,
    MLB1 is MLB - 1,
    MB1 is MB + 1,
    makemove(left, MLB1, CLB, MB1, CB, MRB, CRB, S).

move(left, MLB, CLB, MB, CB, MRB, CRB, S) :-
    CLB > 0,
    CLB1 is CLB - 1,
    CB1 is CB + 1,
    makemove(left, MLB, CLB1, MB, CB1, MRB, CRB, S).

% Move one person from the boat onto the left bank

move(left, MLB, CLB, MB, CB, MRB, CRB, S) :-
    MB > 0,
    MB1 is MB - 1,
    MLB1 is MLB + 1,
    makemove(left, MLB1, CLB, MB1, CB, MRB, CRB, S).

move(left, MLB, CLB, MB, CB, MRB, CRB, S) :-
    CB > 0,
    CB1 is CB - 1,
    CLB1 is CLB + 1,
    makemove(left, MLB, CLB1, MB, CB1, MRB, CRB, S).

% Move one person from the boat onto the right bank.

move(right, MLB, CLB, MB, CB, MRB, CRB, S) :-
    MB > 0,
    MB1 is MB - 1,
    MRB1 is MRB + 1,
    makemove(right, MLB, CLB, MB1, CB, MRB1, CRB, S).

move(right, MLB, CLB, MB, CB, MRB, CRB, S) :-
    CB > 0,
    CB1 is CB - 1,
    CRB1 is CRB + 1,
    makemove(right, MLB, CLB, MB, CB1, MRB, CRB1, S).

% Move one person from the right bank into the boat.

move(right, MLB, CLB, MB, CB, MRB, CRB, S) :-
    MRB > 0,
    MRB1 is MRB - 1,
    MB1 is MB + 1,
    makemove(right, MLB, CLB, MB1, CB, MRB1, CRB, S).

move(right, MLB, CLB, MB, CB, MRB, CRB, S) :-
    CRB > 0,
    CRB1 is CRB - 1,
    CB1 is CB + 1,
    makemove(right, MLB, CLB, MB, CB1, MRB, CRB1, S).

% Move the boat

move(left, MLB, CLB, MB, CB, MRB, CRB, S) :-
    B is MB + CB,
    B > 0,
    makemove(enroute, MLB, CLB, MB, CB, MRB, CRB, S).

move(enroute, MLB, CLB, MB, CB, MRB, CRB, S) :-
    makemove(right, MLB, CLB, MB, CB, MRB, CRB, S).

move(right, MLB, CLB, MB, CB, MRB, CRB, S) :-
    B is MB + CB,
    B > 0,
    makemove(enroute, MLB, CLB, MB, CB, MRB, CRB, S).

move(enroute, MLB, CLB, MB, CB, MRB, CRB, S) :-
    makemove(left, MLB, CLB, MB, CB, MRB, CRB, S).

%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%
%
% Goal state

move(right, 0, 0, 0, 0, _, _, S) :-
    reverse(S, T),
    print(T).

% Administrative

makemove(B, MLB, CLB, MB, CB, MRB, CRB, S) :-
    member([B, MLB, CLB, MB, CB, MRB, CRB], S), !, fail.

makemove(B, MLB, CLB, MB, CB, MRB, CRB, S) :-
    constrain(B, MLB, CLB, MB, CB, MRB, CRB, S),
    move(B, MLB, CLB, MB, CB, MRB, CRB, [[B, MLB, CLB, MB, CB, MRB, CRB] | S]).

solve :-
    makemove(left, 2, 2, 0, 0, 0, 0, []).

