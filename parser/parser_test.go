package parser

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	text := "fun add(a, b) while true do if get_key() == 'enter' do return a + b else var x = 0.8 return x end end var x = 1 return add(x, 0.5)"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
		return
	}
	var b strings.Builder
	for _, stmt := range program {
		b.WriteString(stmt.Inspect())
	}
	s := b.String()
	t.Log(s)
}

func TestParse(t *testing.T) {
	text := "fun add(a, b) return a + b end var x = 1 return add(x, 0.5)"
	v, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, stmt := range v {
		b.WriteString(stmt.Inspect())
	}
	s := b.String()
	t.Log(s)
}
