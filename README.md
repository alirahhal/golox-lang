## Stack-Based VM implementation of the lox language

golox-lang is a stack based VM implementation (compiles to bytecode) of the lox language designed by [Bob Nystrom](https://github.com/munificent) in his book [Crafting Interpreters](http://www.craftinginterpreters.com/), using GoLang.

## Environment setup

You just need to have Go and Make installed on your sytem to get up and running. You can follow the instructions [here](https://go.dev/doc/install) to install Go on your system. On windows, you can use the choco package manager to install Make `choco install make`.

## Basic Usage

Run Ripple:

```Make
make run
```

Run with a specific sample file:

```Make
make run file=samples/basic.lox
```

## Grammar

The [context-free-grammar](docs/context-free-grammar.md) file contains the grammar of the whole language.

### Current Status

- [x] Expressions
- [x] Statements
- [x] Variables(Global&Local)
- [x] Control flow
- [x] Functions
- [ ] Closures
- [x] Classes
