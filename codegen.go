package main

var prefix = `
.global _start
.align 2

plus:
	add X2,X0,X1
	str X2,[sp,#-16]!
	ret

minus:
	sub X2,X0,X1
	str X2,[sp,#-16]!
	ret

multiply:
	mul X2,X0,X1
	str X2,[sp,#-16]!
	ret

divide:
	udiv X2,X0,X1
	str X2,[sp,#-16]!
	ret
_start:

`

var suffix = `
	mov X0,X2
	mov X16,#1
	svc #0x80
`
