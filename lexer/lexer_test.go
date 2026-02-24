package lexer

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			t    TokenType
			text string
		}
	}{
		{
			"basic function and assignment",
			"fun incr(v, n) x = v + n return x end",
			[]struct {
				t    TokenType
				text string
			}{
				{TFun, "fun"},
				{TIdent, "incr"},
				{TLParen, "("},
				{TIdent, "v"},
				{TComma, ","},
				{TIdent, "n"},
				{TRParen, ")"},
				{TIdent, "x"},
				{TAssign, "="},
				{TIdent, "v"},
				{TPlus, "+"},
				{TIdent, "n"},
				{TReturn, "return"},
				{TIdent, "x"},
				{TEnd, "end"},
			},
		},
		{
			"literals and booleans",
			"true false null 123 45.6 'single' \"double\"",
			[]struct {
				t    TokenType
				text string
			}{
				{TTrue, "true"},
				{TFalse, "false"},
				{TNull, "null"},
				{TDigit, "123"},
				{TDigit, "45.6"},
				{TLiteral, "single"},
				{TLiteral, "double"},
			},
		},
		{
			"string with braces (still TLiteral in this branch)",
			"\"hello {name}! {{escaped}}\"",
			[]struct {
				t    TokenType
				text string
			}{
				{TLiteral, "hello {name}! {{escaped}}"},
			},
		},
		{
			"prefix operators",
			"not a type b str c",
			[]struct {
				t    TokenType
				text string
			}{
				{TNot, "not"},
				{TIdent, "a"},
				{TType, "type"},
				{TIdent, "b"},
				{TIdent, "str"},
				{TIdent, "c"},
			},
		},
		{
			"module, imports and extern",
			"pub import x from 'y' extern 'native' fun f(a, b) extern 'native' let v",
			[]struct {
				t    TokenType
				text string
			}{
				{TPub, "pub"},
				{TImport, "import"},
				{TIdent, "x"},
				{TFrom, "from"},
				{TLiteral, "y"},
				{TExtern, "extern"},
				{TLiteral, "native"},
				{TFun, "fun"},
				{TIdent, "f"},
				{TLParen, "("},
				{TIdent, "a"},
				{TComma, ","},
				{TIdent, "b"},
				{TRParen, ")"},
				{TExtern, "extern"},
				{TLiteral, "native"},
				{TLet, "let"},
				{TIdent, "v"},
			},
		},
		{
			"record and list",
			"[1, 2] { a = b }",
			[]struct {
				t    TokenType
				text string
			}{
				{TLBrace, "["},
				{TDigit, "1"},
				{TComma, ","},
				{TDigit, "2"},
				{TRBrace, "]"},
				{TLBracket, "{"},
				{TIdent, "a"},
				{TAssign, "="},
				{TIdent, "b"},
				{TRBracket, "}"},
			},
		},
		{
			"comments",
			"// line\n/* block */ x",
			[]struct {
				t    TokenType
				text string
			}{
				{TIdent, "x"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := New([]rune(tt.input))
			for i, exp := range tt.expected {
				tok, err := lex.Next()
				if err != nil {
					t.Fatalf("%d: %v", i, err)
				}
				if tok.Type == TEOF {
					t.Fatalf("%d: unexpected EOF", i)
				}
				if tok.Type != exp.t {
					t.Errorf("%d: type mismatch: expected=%v, actual=%v", i, exp.t, tok.Type)
				}
				if tok.Text != exp.text {
					t.Errorf("%d: text mismatch: expected=%q, actual=%q", i, exp.text, tok.Text)
				}
			}
			// Check if TEOF follows
			tok, err := lex.Next()
			if err != nil {
				t.Fatal(err)
			}
			if tok.Type != TEOF {
				t.Errorf("expected TEOF, but got %v", tok.Type)
			}
		})
	}
}
