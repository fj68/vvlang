package lexer

import "testing"

func TestAssignmentLex(t *testing.T) {
	text := "x = 1"
	expected := []*Token{
		{TIdent, "x", Pos{0, 1}},
		{TAssign, "=", Pos{2, 3}},
		{TDigit, "1", Pos{4, 5}},
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
