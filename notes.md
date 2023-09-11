## Parser for lisp

- Lisp is fundamentally based on S-expressions
- Something along the lines of `(+ 1 (+ 2 3))`
- Grammar:
  - |expr| -> (|ident| |expr| |expr|)
  - |function| -> (def |ident| (|ident|,|ident|...) |expr| )
- List of builtin identifiers in the scope:
  - `+`,`-`,`*`,`/`: Arithmetic operators
  - User defined functions go to the scope(Main function has to be defined in the end, or functions can only be used after their declaration)
- Recursive parser:

  - Start at left paren, skip one step and parse the ident for the function operation
  - then recursive call to parse expression again
  - store them in a tree

- How to evaluate an identifier: identifier should point to an ast node during binding
  - And thus when evaluating identifier just evaluate that ASTNode

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
- For function expresssions
  - Basically functions store the link and the return address at the x29 and x30
  - Decrement the stack pointer depending on the number of local variables(Since lisp does'nt have any,its just arguments pushed to the stack)
  - Instead of this approach, can we use a simple stack?
    - push all the arguments to the stack
    - call the function, load into registers depending on the number of arguments
    - push the return values to the stack, end of function just pop the last value to the exit register

## Scope implementation for functions

- Theres a main function that does not take arguments
- On evaluating a function, add that function to a global scope
- Each function has a local scope, on call evaluation, get the function object, get the argument strings in order and add them to the scope and proceed to eval the body
- function scope should inherit the global scope though(for example in recursive calls) -> Extended environment

### Function evaluation

- Extend the function environment, add some variables to the inner scope and put the outer scope as the outer field, this environment should be used for the execution of the body

- [ ] Cleanup parser code(on every next char, add a comment for the logic)
- [x] Skip whitespace implementation(and use that to reimplement parse) -> test it for some examples
- [x] Add support for functions(parsing)
- [ ] Add codegen for function expressions
- [ ] Support LLVM IR
- [ ] Compiling Fibonacci
- [ ] LLVM syscalls
- [ ] Infinite locals and params
