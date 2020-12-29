package ast

import "avidbound.com/zego/ast/term"

// Errors represents a series of errors encountered during parsing, compiling,
// etc.
type Errors []*Error

// Error represents a single error caught during parsing, compiling, etc.
type Error struct {
	Message  string         `json:"message"`
	Location *term.Location `json:"location,omitempty"`
}
