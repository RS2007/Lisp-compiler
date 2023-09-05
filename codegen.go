package main

import "fmt"


func Codegen(root ASTNode) string{
	asm := `
.global _start
.align 2
		
	`
	subroutine := ""
	switch root.(type) {
		case *SExpr:
			var sexpr *SExpr
			var ok bool
			if sexpr,ok = root.(*SExpr); !ok {
				panic("Error")
			}
			asm += fmt.Sprintf(`
_start: mov X0,#%d
	mov X1,#%d
	`,sexpr.left.eval(),sexpr.right.eval())
			switch sexpr.operand {
				case "+":
					subroutine += `
plus:
	add X2,X0,X1
	ret
`
					asm += "bl plus"
					break
				case "-":
					subroutine += `
minus:
	sub X2,X0,X1
	ret
						`
					asm += "bl minus"
					break
				case "*":
					subroutine += `
multiply:
	mul X2,X0,X1
	ret
						`
					asm += "bl multiply"
					break
				case "/":
					subroutine += `
divide:
	div X2,X0,X1
	ret
					`
					asm += "bl divide"
					break

			}
	}
	asm +=  `
	mov X0,X2
	mov X16,#1
	svc #0x80
	`
	fmt.Println("Added some stuff")
	asm += fmt.Sprintf(`
	%s
	`,subroutine)
	return asm
}

