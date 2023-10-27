# Lisp-compiler

- Learning about llvm ir by building a lisp compiler.
- Currently supports an interpret and compile mode.
  - Interpret runs a tree walking interpreter on the AST.
  - Compiler generates LLVM IR thats compiled using LLVM's `llc` to generate an executable
- Handwritten lexer+parser
- Supports the write syscall for Mac M2 AArch64 platform (Others to be added in due course)

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
- Write Syscall support

> Write syscall does'nt work with the interpret mode, since references are implemented only for LLVM IR. Instead to print just return the value from the function and subsequently main.

### Demo


https://github.com/RS2007/Lisp-compiler/assets/83483297/4f94ba7f-a707-4663-805e-02c029c435e1



> [!Warning]
> Using an If expression without an else expression will implicitly assume the else block expression to be 0.

