	;; Send a string to an output port

port	EQU	10h
	
	LHLD	hello
top:	MOV	A, M
	CPI	'$'
	JZ	done
	OUT	port
	INX	H
	JMP	top
done:	HLT
hello:	DB	'HELLO'
	DB	13,10,'$'
	END
