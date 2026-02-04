package lexer

import "testing"

func TestVarLex(t *testing.T) {
	text := "var x = 1"
	expected := []*Token{
		{TVar, "var", Pos{0, 3}},
		{TIdent, "x", Pos{4, 5}},
		{TAssign, "=", Pos{6, 7}},
		{TDigit, "1", Pos{8, 9}},
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
