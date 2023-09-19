# Lisp-compiler

- Learning about llvm ir by building a lisp compiler.
- Currently supports an interpret and compile mode.
  - Interpret runs a tree walking interpreter on the AST.
  - Compiler generates LLVM IR thats compiled using LLVM's `llc` to generate an executable
- Handwritten lexer+parser

## Prerequisites

- A golang compiler
- LLVM toolchain(`llc` should be in path)
- GCC/Clang(For assembling)

## Usage

```
  go build
  ./lisp-compiler interpret <name-of-file> # For running the interpreter
  ./lisp-compiler compile <name-of-file># Compiles to an executable called output
```

## Current progress

### Features

- Function expressions
- If expressions
- Integer data structures & arithmetic and comparision operators on them
- Interpret and compile modes
- Write Syscall working

### Demo




https://github.com/RS2007/Lisp-compiler/assets/83483297/d844cdf1-64ee-46ef-ad99-fd742db9fe9e

