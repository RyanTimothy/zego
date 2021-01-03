package tokens

type Token int

func (t Token) String() string {
	if t < 0 || int(t) >= len(strings) {
		return "unknown"
	}
	return strings[t]
}

const (
	Illegal Token = iota
	EOF
	EOL
	Whitespace
	Identifier
	Field
	Comment // TODO

	Package
	Import // TODO
	Else   // TODO
	Null   // TODO
	True
	False

	Number
	String
	RawString

	LBracket
	RBracket
	LBrace
	RBrace
	LParenthesis
	RParenthesis
	Comma
	Colon // TODO
	Declare
	Assign

	Add
	Subtract
	Multiply
	Divide
	Modulus
	And       // TODO
	Or        // TODO
	NEqual    // not equal
	Equal     // equal
	LT        // less than
	GT        // greater than
	LTE       // less than or equal
	GTE       // greater than or equal
	Dot       // TODO
	Semicolon // TODO
)

var strings = [...]string{
	Illegal:      "illegal",
	EOF:          "eof",
	EOL:          "eol",
	Whitespace:   "whitespace",
	Identifier:   "identifier",
	Field:        "field",
	Comment:      "comment",
	Package:      "package",
	Import:       "import",
	Else:         "else",
	Null:         "null",
	True:         "true",
	False:        "false",
	Number:       "number",
	String:       "string",
	RawString:    "rawstring",
	LBracket:     "[",
	RBracket:     "]",
	LBrace:       "{",
	RBrace:       "}",
	LParenthesis: "(",
	RParenthesis: ")",
	Comma:        ",",
	Colon:        ":",
	Declare:      "declare",
	Assign:       "=",
	Add:          "add",      // +
	Subtract:     "minus",    // -
	Multiply:     "multiply", // *
	Divide:       "divide",   // /
	Modulus:      "modulus",  // %
	And:          "and",      // &
	Or:           "or",       // |
	NEqual:       "nEqual",   // !=
	Equal:        "equal",    // ==
	LT:           "lt",
	GT:           "gt",
	LTE:          "lte",
	GTE:          "gte",
	Dot:          ".",
	Semicolon:    ";",
}

var keywords = map[string]Token{
	"package": Package,
	"import":  Import,
	"else":    Else,
	"null":    Null,
	"true":    True,
	"false":   False,
}

func Keyword(lit string) Token {
	if tok, ok := keywords[lit]; ok {
		return tok
	}
	return Identifier
}
