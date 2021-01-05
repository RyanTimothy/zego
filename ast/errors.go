package ast

import (
	"fmt"
	"strings"

	"avidbound.com/zego/ast/term"
)

// Errors represents a series of errors encountered during parsing, compiling,
// etc.
type Errors []error

// Error represents a single error caught during parsing, compiling, etc.
type Error struct {
	File     string         `json:"file"`
	Message  string         `json:"message"`
	Location *term.Location `json:"location,omitempty"`
}

func NewError(loc *term.Location, f string, a ...interface{}) *Error {
	return &Error{
		Location: loc,
		Message:  fmt.Sprintf(f, a...),
	}
}

func (e Errors) Error() string {

	if len(e) == 0 {
		return "no error(s)"
	}

	if len(e) == 1 {
		return fmt.Sprintf("1 error occurred: %v", e[0].Error())
	}

	s := []string{}
	for _, err := range e {
		s = append(s, err.Error())
	}

	return fmt.Sprintf("%d errors occurred:\n%s", len(e), strings.Join(s, "\n"))
}

func (e *Error) Error() string {

	var prefix string

	if e.Location != nil {
		if len(e.Location.File) > 0 {
			prefix += e.Location.File + ":" + fmt.Sprint(e.Location.Line)
		} else {
			prefix += fmt.Sprint(e.Location.Line) + ":" + fmt.Sprint(e.Location.Column)
		}
	}

	msg := e.Message

	if len(prefix) > 0 {
		msg = prefix + ": " + msg
	}

	return msg
}
