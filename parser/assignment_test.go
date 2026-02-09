package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParseVarDecl(t *testing.T) {
	text := "x = 1"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(program) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(program))
	}
	v, ok := program[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected VarDeclStmt, got %T", program[0])
	}
	if v.Name != "x" {
		t.Fatalf("expected name 'x', got %s", v.Name)
	}
}
