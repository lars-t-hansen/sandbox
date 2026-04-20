	;;; 16-bit doubly-recursive fibonacci
	
	;; os entry points
	
;; wboot	equ	0000h
;; syscall	equ	0005h

	;; os service numbers in c
	
;; prchar	equ	2 		; print char in e
;; pr10	equ	5		; print decimal value in de
	
	org	0100h

	;; stack
	ldhli	0006h		; ld hl, (6)
	mv	sp hl		; ld sp, hl

	;; print(fib(14))
	ldwn	de 14		; ld de, 14
	call	f
	ldrn	c 5 		; pr10
	call	5		; syscall
	ldrn	c 2		; prchar
	ldrn	e 13
	call	5		; syscall
	ldrn	c 2		; prchar
	ldrn	e 10
	call	5		; syscall
	
	;; exit
	jp	0		; wboot

	;; 16-bit input n in de
	;; 16-bit result in de
f:	pushw	de
	popw	hl
	ldw	bc 2
	and	a
	sbcw	hl bc
	ret	m
	
	decw	de
	pushw	de
	call	f
	popw	hl
	pushw	de
	exw	de hl
	decw	de
	call	f
	popw	hl
	addw	hl de
	exw	de hl
	ret
