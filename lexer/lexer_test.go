package lexer

import "testing"

func TestLexer(t *testing.T) {
	text := "fun incr(v, n) /* just return v + n */ return v + n end"
	expected := []*Token{
		{TFun, "fun", Pos{0, 3}},
		{TIdent, "incr", Pos{4, 8}},
		{TLParen, "(", Pos{8, 9}},
		{TIdent, "v", Pos{9, 10}},
		{TComma, ",", Pos{10, 11}},
		{TIdent, "n", Pos{12, 13}},
		{TRParen, ")", Pos{13, 14}},
		{TReturn, "return", Pos{39, 45}},
		{TIdent, "v", Pos{46, 47}},
		{TPlus, "+", Pos{48, 49}},
		{TIdent, "n", Pos{50, 51}},
		{TEnd, "end", Pos{52, 55}},
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
