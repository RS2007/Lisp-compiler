## Parser for lisp

- Lisp is fundamentally based on S-expressions
- Something along the lines of `(+ 1 (+ 2 3))`
- Grammar: 
    - |expr| -> (op |expr| |expr|)
    - |function| -> (def |ident| (|ident|,|ident|...) |expr| )
- Recursive parser:
    - Start at left paren, skip one step and parse the operand  
    - then recursive call to parse expression again
    - store them in a tree


## Codegen
- Only support two numbers, so need three registers
- In case of an integer literal: 
    - load the value into registers X0 and X1
    - push them to stack
- In case of an sexpr:
    - pre create subroutines(double generation of subroutines!!)
    - use the assumption that the latest entries in the stack are the operands
    - pretty much a stack virtual machine
    - add all the subroutines to prefix

