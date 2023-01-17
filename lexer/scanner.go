package lexer

import (
	"strings"
)

type Scanner struct {
	Text []rune
	Pos  Pos
	buf  strings.Builder
}

func NewScanner(text []rune) *Scanner {
	return &Scanner{Text: text}
}

func (s *Scanner) IsEOF() bool {
	return len(s.Text) <= s.Pos.End
}

func (s *Scanner) Current() rune {
	if s.IsEOF() {
		return 0
	}
	return s.Text[s.Pos.End]
}

func (s *Scanner) Peek(n int) string {
	if len(s.Text) < s.Pos.End+n {
		return ""
	}
	return string(s.Text[s.Pos.End : s.Pos.End+n])
}

func (s *Scanner) Advance(n int) {
	if len(s.Text) < s.Pos.End+n {
		return
	}
	text := s.Text[s.Pos.End : s.Pos.End+n]
	s.buf.WriteString(string(text))
	s.Pos.End += n
}

func (s *Scanner) Skip(n int) {
	if len(s.Text) < s.Pos.End+n {
		return
	}
	s.Pos.End += n
}

func (s *Scanner) Replace(r rune) {
	s.buf.WriteRune(r)
	s.Skip(1)
}

func (s *Scanner) Flush() (string, Pos) {
	buf := s.buf.String()
	pos := s.Pos.Copy()
	s.buf.Reset()
	s.Pos.Start = s.Pos.End
	return buf, pos
}
