package parser

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestParseStmt(t *testing.T) {
	tests := []TestCase{
		{
			Name:  "simple assignment",
			Input: "let x = 1",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "x",
					Body: &ast.IntLiteralExpr{Value: 1},
				},
			},
		},
		{
			Name:  "function declaration",
			Input: "let add = fun(a, b) return a + b end",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.InfixExpr{
									Op:    "+",
									Left:  &ast.VarRefExpr{Name: "a"},
									Right: &ast.VarRefExpr{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:  "function shorthand",
			Input: "fun add(a, b) return a + b end",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.ReturnStmt{
								Value: &ast.InfixExpr{
									Op:    "+",
									Left:  &ast.VarRefExpr{Name: "a"},
									Right: &ast.VarRefExpr{Name: "b"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:  "while statement",
			Input: "while 1 < 2 let x = 1 end",
			Expected: []ast.Stmt{
				&ast.WhileStmt{
					Cond: &ast.InfixExpr{
						Op:    "<",
						Left:  &ast.IntLiteralExpr{Value: 1},
						Right: &ast.IntLiteralExpr{Value: 2},
					},
					Body: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.IntLiteralExpr{Value: 1},
							},
						},
					},
				},
			},
		},
		{
			Name:  "if statement",
			Input: "if true let x = 1 end",
			Expected: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BoolLiteralExpr{Value: true},
					Then: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.IntLiteralExpr{Value: 1},
							},
						},
					},
				},
			},
		},
		{
			Name:  "if else statement",
			Input: "if true let x = 1 else let x = 2 end",
			Expected: []ast.Stmt{
				&ast.IfStmt{
					Cond: &ast.BoolLiteralExpr{Value: true},
					Then: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.IntLiteralExpr{Value: 1},
							},
						},
					},
					Else: &ast.BlockStmt{
						Body: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: "x",
								Body: &ast.IntLiteralExpr{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			Name:        "return top level",
			Input:       "return 1",
			ExpectedErr: "return statement is not allowed here (must be inside a function or block)",
		},
		{
			Name:        "return no value",
			Input:       "return",
			ExpectedErr: "return statement is not allowed here (must be inside a function or block)",
		},
		{
			Name:  "complex nested",
			Input: "fun add(a, b) while true if get_key() == 'enter' return a + b else let x = 0.8 return x end end end",
			Expected: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: "add",
					Body: &ast.FunLiteralExpr{
						Args: []string{"a", "b"},
						Body: []ast.Stmt{
							&ast.WhileStmt{
								Cond: &ast.BoolLiteralExpr{Value: true},
								Body: &ast.BlockStmt{
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.InfixExpr{
												Op: "==",
												Left: &ast.FunCallExpr{
													Fun: &ast.VarRefExpr{Name: "get_key"},
												},
												Right: &ast.ListLiteralExpr{
													Elements: []ast.Expr{
														&ast.CharLiteralExpr{Value: 'e'},
														&ast.CharLiteralExpr{Value: 'n'},
														&ast.CharLiteralExpr{Value: 't'},
														&ast.CharLiteralExpr{Value: 'e'},
														&ast.CharLiteralExpr{Value: 'r'},
													},
												},
											},
											Then: &ast.BlockStmt{
												Body: []ast.Stmt{
													&ast.ReturnStmt{
														Value: &ast.InfixExpr{
															Op:    "+",
															Left:  &ast.VarRefExpr{Name: "a"},
															Right: &ast.VarRefExpr{Name: "b"},
														},
													},
												},
											},
											Else: &ast.BlockStmt{
												Body: []ast.Stmt{
													&ast.VarDeclStmt{
														Name: "x",
														Body: &ast.FloatLiteralExpr{Value: 0.8},
													},
													&ast.ReturnStmt{
														Value: &ast.VarRefExpr{Name: "x"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:  "valid extern fun",
			Input: "extern 'native' fun f(a, b)",
			Expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f",
					Exported: false,
				},
			},
		},
		{
			Name:  "valid extern let",
			Input: "extern 'native' let v",
			Expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "v",
					Exported: false,
				},
			},
		},
		{
			Name:  "multiple valid extern statements",
			Input: "extern 'native' fun f1()\nextern 'native' let v1\nextern 'native' fun f2()",
			Expected: []ast.Stmt{
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f1",
					Exported: false,
				},
				&ast.ExternStmt{
					Type:     "native",
					Name:     "v1",
					Exported: false,
				},
				&ast.ExternStmt{
					Type:     "native",
					Name:     "f2",
					Exported: false,
				},
			},
		},
		{
			Name:        "invalid extern placement",
			Input:       "let x = 1\nextern 'native' let v",
			ExpectedErr: "extern statement must be after import statements and before other statements",
		},
		{
			Name:        "invalid extern syntax - missing literal",
			Input:       "extern fun f()",
			ExpectedErr: "expected string literal, but got Fun",
		},
		{
			Name:        "invalid extern syntax - missing fun/let",
			Input:       "extern 'native' f()",
			ExpectedErr: "expected 'fun' or 'let' after extern literal, but got Ident",
		},
	}

	for _, tc := range tests {
		RunTest(t, tc)
	}
}
