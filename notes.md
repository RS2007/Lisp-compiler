## Parser for lisp

- Lisp is fundamentally based on S-expressions
- Something along the lines of `(+ 1 (+ 2 3))`
- Grammar: 
    - |expr| -> (op |expr| |expr|)
- Recursive parser:
    - Start at left paren, skip one step and parse the operand  
    - then recursive call to parse expression again
    - store them in a tree

