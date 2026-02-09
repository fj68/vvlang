package parser

import (
	"github.com/fj68/vvlang/ast"
	"testing"
)

func TestParseRecordLiteral(t *testing.T) {
	text := "{ name = 'value', key = 8 }"
	v, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(v))
	}
	exprStmt, ok := v[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", v[0])
	}
	rec, ok := exprStmt.Expr.(*ast.RecordLiteralExpr)
	if !ok {
		t.Fatalf("expected RecordLiteralExpr, got %T", exprStmt.Expr)
	}
	if len(rec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(rec.Fields))
	}
	if _, ok := rec.Fields["name"]; !ok {
		t.Fatalf("missing field 'name'")
	}
	if _, ok := rec.Fields["key"]; !ok {
		t.Fatalf("missing field 'key'")
	}
}

func TestParseRecordLiteralTrailingComma(t *testing.T) {
	text := "{ name = 'value', }"
	v, err := Parse([]rune(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(v))
	}
	exprStmt, ok := v[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", v[0])
	}
	rec, ok := exprStmt.Expr.(*ast.RecordLiteralExpr)
	if !ok {
		t.Fatalf("expected RecordLiteralExpr, got %T", exprStmt.Expr)
	}
	if len(rec.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(rec.Fields))
	}
	if _, ok := rec.Fields["name"]; !ok {
		t.Fatalf("missing field 'name'")
	}
}
