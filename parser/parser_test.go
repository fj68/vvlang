package parser

import (
	"strings"
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParser(t *testing.T) {
	text := "fun add(a, b) while true if get_key() == 'enter' return a + b else let x = 0.8 return x end end end let x = 1 let y = add(x, 0.5)"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
		return
	}
	var b strings.Builder
	for _, stmt := range program.Statements {
		b.WriteString(stmt.Inspect())
	}
	s := b.String()
	t.Log(s)
}

func TestParse(t *testing.T) {
	text := "fun add(a, b) return a + b end let x = 1 let y = add(x, 0.5)"
	v, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, stmt := range v.Statements {
		b.WriteString(stmt.Inspect())
	}
	s := b.String()
	t.Log(s)
}

func TestParseWhile(t *testing.T) {
	text := "while 1 < 2 let x = 1 end"
	v, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Statements) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(v.Statements))
	}
	if _, ok := v.Statements[0].(*ast.WhileStmt); !ok {
		t.Fatalf("expected WhileStmt, got %T", v.Statements[0])
	}
}
