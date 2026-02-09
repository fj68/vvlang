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
	if len(rec.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(rec.Elements))
	}
	
	// Check that both elements are RecordFields
	for i, elem := range rec.Elements {
		_, ok := elem.(*ast.RecordField)
		if !ok {
			t.Fatalf("expected RecordField at index %d, got %T", i, elem)
		}
	}
	
	// Check field names
	fieldMap := make(map[string]bool)
	for _, elem := range rec.Elements {
		if field, ok := elem.(*ast.RecordField); ok {
			fieldMap[field.Key] = true
		}
	}
	if _, ok := fieldMap["name"]; !ok {
		t.Fatalf("missing field 'name'")
	}
	if _, ok := fieldMap["key"]; !ok {
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
	if len(rec.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(rec.Elements))
	}
	field, ok := rec.Elements[0].(*ast.RecordField)
	if !ok {
		t.Fatalf("expected RecordField, got %T", rec.Elements[0])
	}
	if field.Key != "name" {
		t.Fatalf("expected field 'name', got '%s'", field.Key)
	}
}
