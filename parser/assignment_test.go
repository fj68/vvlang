package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParseVarDecl(t *testing.T) {
	text := "let x = 1"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(program.Statements))
	}
	v, ok := program.Statements[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected VarDeclStmt, got %T", program.Statements[0])
	}
	if v.Name != "x" {
		t.Fatalf("expected name 'x', got %s", v.Name)
	}
}
