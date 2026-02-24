package lexer

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	s *Scanner
}

func New(text []rune) *Lexer {
	return &Lexer{NewScanner(text)}
}

func (lex *Lexer) Next() (*Token, error) {
	lex.skipWhitespacesAndComments()

	if lex.s.IsEOF() {
		return lex.newToken(TEOF), nil
	}

	// Check for docstring before anything else
	if lex.s.Peek(3) == "///" {
		return lex.docstring()
	}

	r := lex.s.Current()

	for sym2, ty := range Symbols2 {
		if lex.s.Peek(len(sym2)) == sym2 {
			lex.s.Advance(len(sym2))
			return lex.newToken(ty), nil
		}
	}

	for sym, ty := range Symbols {
		if r == sym {
			lex.s.Advance(1)
			return lex.newToken(ty), nil
		}
	}

	if unicode.IsDigit(r) {
		return lex.digit()
	}

	if r == '\'' {
		return lex.literal()
	}

	if r == '"' {
		return lex.interpolated()
	}

	if IsIdentLetter(r) {
		return lex.ident()
	}

	return nil, fmt.Errorf("unexpected letter '%s'", string(r))
}

func (lex *Lexer) newToken(ty TokenType) *Token {
	text, pos := lex.s.Flush()
	return &Token{
		ty,
		text,
		pos,
	}
}

func (lex *Lexer) skipWhitespacesAndComments() {
	for !lex.s.IsEOF() {
		read := 0
		read += lex.skipWhitespaces()
		read += lex.skipComment()
		if read == 0 {
			break
		}
	}
}

func (lex *Lexer) skipWhitespaces() int {
	for !lex.s.IsEOF() && unicode.IsSpace(lex.s.Current()) {
		lex.s.Skip(1)
	}
	_, pos := lex.s.Flush() // reset start pos
	return pos.End - pos.Start
}

func (lex *Lexer) skipComment() int {
	// Check docstring marker first to avoid treating it as a regular comment
	if lex.s.Peek(3) == "///" {
		return 0 // Do not skip; Next() will emit TDocstring
	}
	for start, end := range Comments {
		if lex.s.Peek(len(start)) == start {
			return lex.comment(start, end)
		}
	}
	return 0
}

// docstring reads a `///` line and emits a TDocstring token with the trimmed content.
func (lex *Lexer) docstring() (*Token, error) {
	lex.s.Skip(3) // skip `///`
	// skip optional single space after ///
	if !lex.s.IsEOF() && lex.s.Current() == ' ' {
		lex.s.Skip(1)
	}
	// Read until end of line
	for !lex.s.IsEOF() && lex.s.Current() != '\n' {
		lex.s.Advance(1)
	}
	if !lex.s.IsEOF() {
		lex.s.Skip(1) // skip the newline
	}
	return lex.newToken(TDocstring), nil
}

func (lex *Lexer) comment(start, end string) int {
	lex.s.Skip(len(start))
	for !lex.s.IsEOF() && lex.s.Peek(len(end)) != end {
		lex.s.Skip(1)
	}
	lex.s.Skip(len(end))
	_, pos := lex.s.Flush() // reset start pos
	return pos.End - pos.Start
}

func (lex *Lexer) digit() (*Token, error) {
	for !lex.s.IsEOF() && unicode.IsDigit(lex.s.Current()) {
		lex.s.Advance(1)
	}
	if lex.s.Current() == '.' {
		lex.s.Advance(1)
		for !lex.s.IsEOF() && unicode.IsDigit(lex.s.Current()) {
			lex.s.Advance(1)
		}
	}
	return lex.newToken(TDigit), nil
}

func (lex *Lexer) literal() (*Token, error) {
	marker := lex.s.Current()
	lex.s.Skip(1)

	for !lex.s.IsEOF() && lex.s.Current() != marker {
		if lex.s.Current() == '\\' {
			lex.scanEscapeSequence()
		} else {
			lex.s.Advance(1)
		}
	}

	if lex.s.IsEOF() {
		return nil, fmt.Errorf("unexpected eof while reading string literal: %s", lex.s.Pos)
	}

	lex.s.Skip(1) // skip end marker

	return lex.newToken(TLiteral), nil
}

func (lex *Lexer) interpolated() (*Token, error) {
	marker := lex.s.Current()
	lex.s.Skip(1)

	for !lex.s.IsEOF() && lex.s.Current() != marker {
		if lex.s.Current() == '\\' {
			lex.scanEscapeSequence()
		} else {
			lex.s.Advance(1)
		}
	}

	if lex.s.IsEOF() {
		return nil, fmt.Errorf("unexpected eof while reading string literal: %s", lex.s.Pos)
	}

	lex.s.Skip(1) // skip end marker

	return lex.newToken(TInterpolated), nil
}

func (lex *Lexer) scanEscapeSequence() {
	lex.s.Skip(1) // skip marker

	r := lex.s.Current()
	switch r {
	case 't':
		lex.s.Replace('\t')
		return
	case 'r':
		lex.s.Replace('\r')
		return
	case 'n':
		lex.s.Replace('\n')
		return
	case 'b':
		lex.s.Replace('\b')
		return
	default:
		lex.s.Replace(r)
		return
	}
}

func (lex *Lexer) ident() (*Token, error) {
	for !lex.s.IsEOF() && IsIdentLetter(lex.s.Current()) {
		lex.s.Advance(1)
	}
	tok := lex.newToken(TIdent)

	// check keyword
	for kw, ty := range Keywords {
		if tok.Text == kw {
			tok.Type = ty
			return tok, nil
		}
	}

	return tok, nil
}
