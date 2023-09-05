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
    - write a mov instruction for the same
- In case of an sexpr:
    - create a subroutine for the operation and then use the X0 and X1 values for the same
    - subroutine should be stored in a different string since it should be appended to the end

