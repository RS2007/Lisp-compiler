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

- [x] Cleanup parser code(on every next char, add a comment for the logic)
- [x] Skip whitespace implementation(and use that to reimplement parse) -> test it for some examples
- [x] Add support for functions(parsing)
- [ ] Add codegen for function expressions
- [ ] Support LLVM IR
- [ ] Compiling Fibonacci
- [ ] LLVM syscalls
- [ ] Infinite locals and params

## LLVM IR generation

- All functions in our lisp version take in integers and returns integers

### Function signature generation

```lisp
  (def add_two(a b) (+ a b 2))
  (def main() (add_two 1 2))
```

```
define i32 @add_two(i32 %a,i32 %b){
  %sym1 = add i32 %a,0
  %sym2 = add i32 %b,0
  %sym3 = add i32 2,0
  %sym4 = add i32 %a,%b
  %sym5 = add i32 %sym4,%sym3
  ret i32 %sym5
}
define i32 @main(){
  %sym1 = add i32 1,0
  %sym2 = add i32 2,0
  %sym3 = call i32 @plus_two(i32 %sym1,i32 %sym2)
  ret i32 %sym5
}
```

- Need to keep track of some things

  - What is the count of sym for the next local variable
  - once these are generated, what is the mapping between the llvm variable and the variable in the lisp code(This is unnecessary since the lisp variable constraints that we have posed make them legitimate llvm variables)

- How to generate and keep track of symbols
  - Issue with codegenning for S expressions
    - Recursive generation: say `(+ 1 (+ 1 2) 3)`
    - codegen for the add, reserve two variables, that are the addendums and another to store the result
      - lets say sym1, sym2 and sym3
      - take the length of the arguments array(3 in this case)
      - its odd and therefore the you can batch it into groups of 2 and 1
      - sym4 the next available symbol becomes the result of the addition of the sym6 and sym7 which are the stores of the values 1 and the result of the addition of 1 and 2
      - sym5 becomes the addition of the values 3 and 0(0 has to be added since the length is odd)
      - sym6 will be equal to 1, since this is a simple integerNode
      - sym7 becomes the result of the addition of 1 and 2 which are sym8 and sym9
      - backtrack from there, sym5 = add sym 0
      - sym7 = add sym8 sym9
