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
- For function expressions
  - Basically functions store the link and the return address at the x29 and x30
  - Decrement the stack pointer depending on the number of local variables(Since lisp doesn't have any,its just arguments pushed to the stack)
  - Instead of this approach, can we use a simple stack?
    - push all the arguments to the stack
    - call the function, load into registers depending on the number of arguments
    - push the return values to the stack, end of function just pop the last value to the exit register

## Scope implementation for functions

- There's a main function that does not take arguments
- On evaluating a function, add that function to a global scope
- Each function has a local scope, on call evaluation, get the function object, get the argument strings in order and add them to the scope and proceed to eval the body
- function scope should inherit the global scope though(for example in recursive calls) -> Extended environment

### Function evaluation

- Extend the function environment, add some variables to the inner scope and put the outer scope as the outer field, this environment should be used for the execution of the body

- [x] Cleanup parser code(on every next char, add a comment for the logic)
- [x] Skip whitespace implementation(and use that to reimplement parse) -> test it for some examples
- [x] Add support for functions(parsing)
- [x] Add codegen for function expressions
- [x] Support LLVM IR
- [x] Compiling Fibonacci
- [x] LLVM syscalls
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

### Conditional codegeneration

- Have to push value to stack(local storage), cause llvm doesn't allow multiple assignment
- alloca instruction, store and load

### Printing error

- [x] support for array of SExpr in the function body
- [x] Parser is extremely shady, fix parser
  - [x] The `parseFunctionBody` method on the parser is extremely shady(poor implementation) , need to fix it
  - [ ] Fix other shabby parts of the code, add good asserts and code comments and unit tests

### Current Problems
-  function arguments should be pushed to the stack  
    - add the pointer symbol to the scope 
    - on codegening for the identifier, load the value from the stack using the load instruction
    - this gets rid of unnecessary adds in the generated llvm ir.
- Support for all systemcalls (for arm mac)
    - push all arguments to the stack
    - maintain information about basic syscalls(number of arguments,maybe??)
    - pop the values to the symbols in order
    - generate the asm sideffect instruction
- Add support for phi expressions in the if call, to keep track of the value
    - instead of allocating space on the stack, generate the condition
    - there should be a if label generator in case of nested ifs
    - assign a phi with the true and false conditions and their corresponding return symbols

### Problem with phi nodes
- In case of nested if statements the control flow will be very weird since 
there's no concept of nested labels in assembly
- In order to solve this problem we need to keep track of where a branch instruction should jump to

#### An example

```ll
{
    entry:
	 
	
  %sym1 = alloca i64, align 4
	store i64 %n, i64* %sym1, align 4
    
	
	%sym3 = alloca i64, align 4
	%sym5 = load i64 ,i64* %sym1,align 4
	%sym6 = add i64 2,0
	
	%sym4 = icmp slt i64 %sym5 , %sym6
		
	br i1 %sym4,label %iftrue1,label %iffalse1
	iftrue1:
	
	%sym9 = alloca i64, align 4
	%sym11 = load i64 ,i64* %sym1,align 4
	%sym12 = add i64 0,0
	
	%sym10 = icmp eq i64 %sym11 , %sym12
		
	br i1 %sym10,label %iftrue2,label %iffalse2
	iftrue2:
	
	%sym13 = add i64 1,0
	
    br label %ifresult2
	
	iffalse2:
	%sym14 = load i64 ,i64* %sym1,align 4
      br label %ifresult2
  
	ifresult2:
    %sym7 = phi i64 [%sym13,%iftrue2],[%sym14,%iffalse2]
	
    br label %ifresult1
	
	iffalse1:
	%sym18 = load i64 ,i64* %sym1,align 4
	%sym19 = add i64 1,0
	
	%sym17 = sub i64 %sym18,%sym19
		
	%sym15 = call i64 @fib(i64 %sym17)
	%sym22 = load i64 ,i64* %sym1,align 4
	%sym23 = add i64 2,0
	
	%sym21 = sub i64 %sym22,%sym23
		
	%sym16 = call i64 @fib(i64 %sym21)
	
	%sym8 = add i64 %sym15,%sym16
		
      br label %ifresult1
  
	ifresult1:
    %sym2 = phi i64 [%sym7,%iftrue1],[%sym8,%iffalse1]
	
	ret i64 %sym2
}
```
- rewrite
```ll
{
    entry:
	 
	
  %sym1 = alloca i64, align 4
	store i64 %n, i64* %sym1, align 4
    
	
	%sym3 = alloca i64, align 4
	%sym5 = load i64 ,i64* %sym1,align 4
	%sym6 = add i64 2,0
	
	%sym4 = icmp slt i64 %sym5 , %sym6
		
	br i1 %sym4,label %iftrue1,label %iffalse1
	iftrue1:
    ; queue is now [iftrue1]	
	%sym9 = alloca i64, align 4
	%sym11 = load i64 ,i64* %sym1,align 4
	%sym12 = add i64 0,0
	
	%sym10 = icmp eq i64 %sym11 , %sym12
		
	br i1 %sym10,label %iftrue2,label %iffalse2
	iftrue2:
    ; queue is now [iftrue1,iftrue2]
	
	%sym13 = add i64 1,0
	
    br label %ifresult2
	
	iffalse2:
    ; queue is now [iftrue1,iftrue2,iffalse2]
	%sym14 = load i64 ,i64* %sym1,align 4
      br label %ifresult2
  
	ifresult2:
    ; for this phi expression pop of the two of the top and push ifresult2
    ; [iftrue1,ifresult2]
    %sym7 = phi i64 [%sym13,%iftrue2],[%sym14,%iffalse2]
	
    br label %ifresult1
	
	iffalse1:
    ; [iftrue1,ifresult2,iffalse1]
	%sym18 = load i64 ,i64* %sym1,align 4
	%sym19 = add i64 1,0
	
	%sym17 = sub i64 %sym18,%sym19
		
	%sym15 = call i64 @fib(i64 %sym17)
	%sym22 = load i64 ,i64* %sym1,align 4
	%sym23 = add i64 2,0
	
	%sym21 = sub i64 %sym22,%sym23
		
	%sym16 = call i64 @fib(i64 %sym21)
	
	%sym8 = add i64 %sym15,%sym16
		
      br label %ifresult1
  
	ifresult1:
    ; pop of the first 2
    %sym2 = phi i64 [%sym7,%ifresult2],[%sym8,%iffalse1]
	
	ret i64 %sym2
}

```

- note that the `%sym2 = phi i64 [%sym7,%ifresult2],[%sym8,%iffalse1]` from 
`%sym2 = phi i64 [%sym7,%iftrue1],[%sym8,%iffalse1]`
- the solution is to maintain some kind of queue,
that should keep track of the predecessor label for the phi expression exited if labels

