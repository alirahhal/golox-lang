## Stack-Based VM implementation of the lox language

golanglox is a stack based VM implementation (compiles to bytecode) of the lox language designed by [Bob Nystrom](https://github.com/munificent) in his book [Crafting Interpreters](http://www.craftinginterpreters.com/), using GoLang.


## Usage

Run Ripple:

```Make
make run
```

Run from a sample file:

```Make
make run testfile=samples/test.lox
```

### Current Status

- [x] Expressions
- [x] Statements
- [x] Variables(Global&Local)
- [x] Control flow
- [x] Functions
- [ ] Closures
- [ ] Classes
