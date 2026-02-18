package lexer

import (
	"testing"

	"github.com/fj68/vvlang/ast"
)

func TestLexer(t *testing.T) {
	text := "fun incr(v, n) /* just return v + n */ return v + n end"
	expected := []*Token{
		{TFun, "fun", ast.Pos{0, 3, 0, 3}},
		{TIdent, "incr", ast.Pos{4, 8, 0, 8}},
		{TLParen, "(", ast.Pos{8, 9, 0, 9}},
		{TIdent, "v", ast.Pos{9, 10, 0, 10}},
		{TComma, ",", ast.Pos{10, 11, 0, 11}},
		{TIdent, "n", ast.Pos{12, 13, 0, 13}},
		{TRParen, ")", ast.Pos{13, 14, 0, 14}},
		{TReturn, "return", ast.Pos{39, 45, 0, 45}},
		{TIdent, "v", ast.Pos{46, 47, 0, 47}},
		{TPlus, "+", ast.Pos{48, 49, 0, 49}},
		{TIdent, "n", ast.Pos{50, 51, 0, 51}},
		{TEnd, "end", ast.Pos{52, 55, 0, 55}},
	}
	lex := New([]rune(text))
	for i := 0; ; i++ {
		tok, err := lex.Next()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Type == TEOF {
			break
		}
		if !tok.Eq(expected[i]) {
			t.Fatalf("%d\n\texpected: %s\n\tactual : %s", i, expected[i], tok)
		}
	}
}
