;;; 16-bit doubly-recursive fibonacci
	
	;; os entry points
	
wboot	equ	0000h
syscall	equ	0005h

	;; os service numbers in c
	
prchar	equ	2 		; print char in e
pr10	equ	5		; print decimal value in de
	
	org	0100h

	;; stack
	ld	hl, (syscall+1)
	ld	sp, hl

	;; print(fib(14))
	ld	de, 14
	call	fib
	ld	c, pr10
	call	syscall
	ld	c, prchar
	ld	e, 13
	call	syscall
	ld	c, prchar
	ld	e, 10
	call	syscall
	
	;; exit
	jp	wboot

	;; 16-bit input n in de
	;; 16-bit result in de
fib:	push	de
	pop	hl
	ld	bc, 2
	and	a
	sbc	hl, bc
	ret	m
	
fib2:	dec	de
	push	de
	call	fib
	pop	hl
	push	de
	ex	de, hl
	dec	de
	call	fib
	pop	hl
	add	hl, de
	ex	de, hl
	ret
