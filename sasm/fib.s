;;; 16-bit doubly-recursive fibonacci
	
	;; os entry points
	
wboot	EQU	0000h
syscall	EQU	0005h

	;; os service numbers in c
	
prchar	EQU	2 		; print char in e
pr10	EQU	5		; print decimal value in de
	
	ORG	0100h

	;; stack
	LHLD	syscall+1
	SPHL

	;; print(fib(14))
	LXI	D, 14
	CALL	fib
	MVI	C, pr10
	CALL	syscall
	MVI	C, prchar
	MVI	E, 13
	CALL	syscall
	MVI	C, prchar
	MVI	E, 10
	CALL	syscall
	
	;; exit
	JMP	wboot

	;; 16-bit input n in de
	;; 16-bit result in de
	;; if n < 2 return n
fib:	MOV	A, D
	CPI	0
	JNZ	fib2
	MOV	A, E
	CPI	2
	JP	fib2
	RET

	;; return fib(n-1) + fib(n-2)
fib2:	DCX	D
	PUSH	D
	CALL	fib
	POP	H
	PUSH	D
	XCHG
	DCX	D
	CALL	fib
	POP	H
	DAD	D
	XCHG
	RET
