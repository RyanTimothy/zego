package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"avidbound.com/zego/ast/internal/tokens"
)

const eof = -1

type Item struct {
	Token tokens.Token // Type, such as itemNumber.
	Value string       // Value, such as "23.2".
	Pos   Pos          // The starting position, in bytes, of this item in the input string.
	Line  int          // The line number at the start of this item.
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

type blockDepth struct {
	tokens []tokens.Token
}

// lexer holds the state of the scanner.
type lexer struct {
	name           string     // the name of the input; used only for error reports
	input          string     // the string being scanned
	leftDelim      string     // start of action
	rightDelim     string     // end of action
	trimRightDelim string     // end of action with trim marker
	pos            Pos        // current position in the input
	start          Pos        // start position of this item
	width          Pos        // width of last rune read from input
	items          []Item     // scanned items
	blockDepth     blockDepth // nesting depth of { ( [ exprs
	line           int        // 1+number of newlines seen
	startLine      int        // start line of this item
}

func (b *blockDepth) push(t tokens.Token) {
	b.tokens = append(b.tokens, t)
}

func (b *blockDepth) peek() tokens.Token {
	if len(b.tokens) > 0 {
		return b.tokens[len(b.tokens)-1]
	}
	return tokens.EOF
}

func (b *blockDepth) pop() tokens.Token {
	if len(b.tokens) > 0 {
		t := b.tokens[len(b.tokens)-1]
		b.tokens = b.tokens[:len(b.tokens)-1] //pop
		return t
	}
	return tokens.EOF
}

// lex creates a new scanner for the input string.
func Lex(name, input string) []Item {
	l := &lexer{
		name:       name,
		input:      input,
		line:       1,
		startLine:  1,
		blockDepth: blockDepth{},
	}
	l.run()
	return l.items
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexScan; state != nil; {
		state = state(l)
	}
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// lexScan for token type
func lexScan(l *lexer) stateFn {

	switch r := l.next(); {
	case r == eof:
		if l.blockDepth.pop() != tokens.EOF {
			return l.errorf("unexpected EOF")
		}
		l.emit(tokens.EOF)
		return nil
	case isEndOfLine(r):
		return lexEndOfLine
	case isSpace(r):
		l.backup()
		return lexWhitespace
	case r == '#':
		return lexComment
	case r == '=':
		if l.peek() == '=' {
			l.next()
			l.emit(tokens.Equal)
		} else {
			l.emit(tokens.Assign)
		}
	case r == '!':
		if l.next() != '=' {
			return l.errorf("illegal ! character")
		}
		l.emit(tokens.NEqual)
	case r == '<':
		if l.peek() == '=' {
			l.next()
			l.emit(tokens.LTE)
		} else {
			l.emit(tokens.LT)
		}
	case r == '>':
		if l.peek() == '=' {
			l.next()
			l.emit(tokens.GTE)
		} else {
			l.emit(tokens.GT)
		}
	case r == ':':
		if l.next() != '=' {
			return l.errorf("expected :=")
		}
		l.emit(tokens.Declare)
	case r == '&':
		l.emit(tokens.Add)
	case r == '|':
		l.emit(tokens.Or)
	case r == '/':
		l.emit(tokens.Divide)
	case r == '%':
		l.emit(tokens.Modulus)
	case r == '*':
		l.emit(tokens.Multiply)
	case r == '+':
		l.emit(tokens.Add)
	case r == '-':
		l.emit(tokens.Subtract)
	case r == '"':
		return lexQuote
	case r == '`':
		return lexRawQuote
	case r == ',':
		l.emit(tokens.Comma)
	case r == '.':
		// special look-ahead for ".field" so we don't break l.backup().
		if l.pos < Pos(len(l.input)) {
			r := l.input[l.pos]
			if r < '0' || '9' < r {
				return lexField
			}
		}
		fallthrough // '.' can start a number.
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == '[':
		l.emit(tokens.LBracket)
		l.blockDepth.push(tokens.LBracket)
	case r == ']':
		if l.blockDepth.pop() != tokens.LBracket {
			return l.errorf("unexpected right bracket %#U", r)
		}
		l.emit(tokens.RBracket)
	case r == '(':
		l.emit(tokens.LParenthesis)
		l.blockDepth.push(tokens.LParenthesis)
	case r == ')':
		if l.blockDepth.pop() != tokens.LParenthesis {
			return l.errorf("unexpected right paren %#U", r)
		}
		l.emit(tokens.RParenthesis)
	case r == '{':
		l.emit(tokens.LBrace)
		l.blockDepth.push(tokens.LBrace)
	case r == '}':
		if l.blockDepth.pop() != tokens.LBrace {
			return l.errorf("unexpected right brace %#U", r)
		}
		l.emit(tokens.RBrace)
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}

	return lexScan
}

// lexSpace scans a run of new line characters and skips leading white space in new line.
func lexEndOfLine(l *lexer) stateFn {
	var r rune
	for {
		r = l.peek()
		if !isEndOfLine(r) && !isSpace(r) {
			if r == '.' { // new line before "\n.field"
				return l.errorf("expected identifier")
			}
			break
		}
		l.next()
	}
	l.emit(tokens.EOL)
	return lexScan
}

// lexSpace scans a run of space characters.
// We have not consumed the first space, which is known to be present.
func lexWhitespace(l *lexer) stateFn {
	var r rune
	for {
		r = l.peek()
		if !isSpace(r) {
			if isEndOfLine(r) {
				l.ignore()
				return lexEndOfLine
			}
			if r == '.' { // space before " .field"
				return l.errorf("expected identifier")
			}
			break
		}
		l.next()
	}
	l.emit(tokens.Whitespace)
	return lexScan
}

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	var r rune
	for {
		r = l.next()
		if isEndOfLine(r) || r == eof {
			break
		}
	}
	l.ignore()
	return lexScan
}

// lexQuote scans a quoted string.
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(tokens.String)
	return lexScan
}

// lexRawQuote scans a raw quoted string.
func lexRawQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated raw quoted string")
		case '`':
			break Loop
		}
	}
	l.emit(tokens.RawString)
	return lexScan
}

// lexField scans a field: .Alphanumeric.
// The . has been scanned.
func lexField(l *lexer) stateFn {
	if l.atTerminator() { // Nothing interesting follows -> "."
		return l.errorf("expected field")
	}
	var r rune
	for {
		r = l.next()
		if !isAlphaNumeric(r) {
			l.backup()
			break
		}
	}
	if !l.atTerminator() {
		return l.errorf("bad character %#U", r)
	}
	l.emit(tokens.Field)
	return lexScan
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb. TODO: maybe throw error here
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case tokens.Keyword(word) != tokens.Identifier:
				l.emit(tokens.Keyword(word))
			case word[0] == '.':
				l.emit(tokens.Field)
			default:
				l.emit(tokens.Identifier)
			}
			break Loop
		}
	}
	return lexScan
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(tokens.Number)

	return lexScan
}

func (l *lexer) scanNumber() bool {
	// Is it hex?
	digits := "0123456789_"
	if l.accept("0") {
		// Note: Leading 0 does not mean octal in floats.
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF_"
		} else if l.accept("oO") {
			digits = "01234567_"
		} else if l.accept("bB") {
			digits = "01_"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if len(digits) == 10+1 && l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	if len(digits) == 16+6+1 && l.accept("pP") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces. Also catches cases
// like "$x+2" not being acceptable without a space, in case we decide one
// day to implement arithmetic.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	switch r {
	case eof, '+', '-', '/', '%', '*', '.', ',', '|', ':', ']', '[', ')', '(':
		return true
	}
	// Does r start the delimiter? This can be ambiguous (with delim=="//", $x/2 will
	// succeed but should fail) but only in extremely rare cases caused by willfully
	// bad choice of delimiter.
	if rd, _ := utf8.DecodeRuneInString(l.rightDelim); rd == r {
		return true
	}
	return false
}

// emit passes an item back to the client.
func (l *lexer) emit(t tokens.Token) {
	item := Item{
		Token: t,
		Value: l.input[l.start:l.pos],
		Pos:   l.pos,
		Line:  l.line,
	}

	l.items = append(l.items, item)
	l.start = l.pos
	l.startLine = l.line
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	item := Item{
		Token: tokens.Illegal,
		Value: fmt.Sprintf(format, args...),
		Pos:   l.start,
		Line:  l.startLine,
	}

	l.items = append(l.items, item)
	return nil
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
