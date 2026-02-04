package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParseReturnTopLevel(t *testing.T) {
	text := "return 1"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(program) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(program))
	}
	if _, ok := program[0].(*ast.ReturnStmt); !ok {
		t.Fatalf("expected ReturnStmt, got %T", program[0])
	}
}

func TestParseReturnNoValue(t *testing.T) {
	text := "return"
	program, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(program) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(program))
	}
	rtn, ok := program[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", program[0])
	}
	if rtn.Value != nil {
		t.Fatalf("expected nil value, got %v", rtn.Value)
	}
}
