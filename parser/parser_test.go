package parser

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	text := "fun add(a, b) while true do if get_key() == 'enter' do return a + b else var x = 0.8 return x end end"
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
	if s != "FunLiteralExpr{\"add\", [a, b], [WhileStmt{BoolLiteralExpr{true}, IfStmt{InfixStmt{\"==\", FunCallExpr{VarRefExpr{\"get_key\"}, []}, StringLiteralExpr{enter}}, ReturnStmt{InfixStmt{\"+\", VarRefExpr{\"a\"}, VarRefExpr{\"b\"}}}, VarDeclStmt{\"x\", NumberLiteralExpr{0.8}}, ReturnStmt{VarRefExpr{\"x\"}}}}]}" {
		t.Fatal(s)
	}
}
