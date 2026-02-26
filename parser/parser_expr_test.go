package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParseExpr(t *testing.T) {
	tests := []TestCase{
		{
			Name:  "function call",
			Input: "let y = add(x, 0.5)",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "y",
					Body: &ast.FunCallExpr{
						Fun: &ast.VarRefExpr{Name: "add"},
						Args: []ast.Expr{
							&ast.VarRefExpr{Name: "x"},
							&ast.NumberLiteralExpr{Value: 0.5},
						},
					},
				},
			},
		},
		{
			Name:  "record literal",
			Input: "let r = { a = 1, b = 2 }",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "r",
					Body: &ast.RecordLiteralExpr{
						Fields: map[string]ast.Expr{
							"a": &ast.NumberLiteralExpr{Value: 1},
							"b": &ast.NumberLiteralExpr{Value: 2},
						},
					},
				},
			},
		},
		{
			Name:  "record literal trailing comma",
			Input: "let r = { name = 'value', }",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "r",
					Body: &ast.RecordLiteralExpr{
						Fields: map[string]ast.Expr{
							"name": &ast.ListLiteralExpr{
								Elements: []ast.Expr{
									&ast.CharLiteralExpr{Value: 'v'},
									&ast.CharLiteralExpr{Value: 'a'},
									&ast.CharLiteralExpr{Value: 'l'},
									&ast.CharLiteralExpr{Value: 'u'},
									&ast.CharLiteralExpr{Value: 'e'},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:  "len(xs)",
			Input: "let x = len(xs)",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "x",
					Body: &ast.BuiltinCallExpr{
						Op:    "len",
						Value: &ast.VarRefExpr{Name: "xs"},
					},
				},
			},
		},
		{
			Name:  "len([1, 2, 3])",
			Input: "let x = len([1, 2, 3])",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "x",
					Body: &ast.BuiltinCallExpr{
						Op: "len",
						Value: &ast.ListLiteralExpr{
							Elements: []ast.Expr{
								&ast.NumberLiteralExpr{Value: 1},
								&ast.NumberLiteralExpr{Value: 2},
								&ast.NumberLiteralExpr{Value: 3},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		RunTest(t, tc)
	}
}
