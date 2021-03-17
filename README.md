# Zego
A personal introduction to a lexer, parser and rules engine based off of the [Open Policy Agent (OPA)](https://github.com/open-policy-agent/opa).

The lexer is inspired by the Rob Pike's lecture on [Lexical Scanning in Go](https://youtu.be/HxaD_trXwRE).

## How does Zego work?

Currently in development and does not yet work as intended. As-is this is just a lexer & parser without completed compiling.

## Lexer

With a package input such as
```
package test

abEquals {
    a := input.test[1]
    b := 13.5
    a == b
}
```

The lexer tokenizes the above input into this (excluding whitespace tokens in this):
```
[ 
    package, identifier["test"],

    identifier["abEquals"], lbrace,
        identifier["a"], declare, identifier["input"], field["test"], lbracket, number[1], rbracket,
        identifier["b"], declare, number[13.5],
        identifier["a"], equal, identifier["b"],
    rbrace 
]
```

The parser will parse tokens into AST structure the input below:
```
a == (b + (c - d)) * e
```

Will parse the lexer token output into this:
```
callTerm(
    opTerm("=="), 
    varTerm("a"), 
    callTerm(
        opTerm("*"), 
        callTerm(
            opTerm("+"),
            callTerm(
                opTerm("-"),
                varTerm("c"),
                varTerm("d")
            ),
            varTerm("e")
        )
    )
)
```

The compiler portion is currently incomplete.