package lexer

import (
	"testing"
)

func TestLexerExtern(t *testing.T) {
	text := "extern 'native' fun f(a, b) extern 'native' let v"
	expected := []TokenType{
		TExtern, TLiteral, TFun, TIdent, TLParen, TIdent, TComma, TIdent, TRParen,
		TExtern, TLiteral, TLet, TIdent,
	}
	lex := New([]rune(text))
	for i, exp := range expected {
		tok, err := lex.Next()
		if err != nil {
			t.Fatal(err)
		}
		if tok.Type != exp {
			t.Fatalf("%d: expected %v, but got %v", i, exp, tok.Type)
		}
	}
}
