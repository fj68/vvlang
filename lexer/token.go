package lexer

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	TEOF TokenType = iota
	TDigit
	TIdent
	TLiteral
	TInterplated
	TComment

	// keywords
	TFun
	TReturn
	TEnd
	TWhile
	TDo
	TIf
	TElse
	TTrue
	TFalse
	TIn
	TMod
	TAnd
	TOr

	// symbols
	TEqual
	TLParen
	TRParen
	TLess
	TComma
	TLBrace
	TRBrace
	TLBracket
	TRBracket
	TPlus
	THyphen
	TAsterisk
	TSlash
	TDot
)

func (ty TokenType) String() string {
	switch ty {
	case TEOF:
		return "EOF"
	case TDigit:
		return "Digit"
	case TIdent:
		return "Ident"
	case TLiteral:
		return "Literal"
	case TInterplated:
		return "Interpolated"
	case TComment:
		return "Comment"

		// keywords
	case TFun:
		return "Fun"
	case TReturn:
		return "Return"
	case TEnd:
		return "End"
	case TWhile:
		return "While"
	case TDo:
		return "Do"
	case TIf:
		return "If"
	case TElse:
		return "Else"
	case TTrue:
		return "True"
	case TFalse:
		return "False"
	case TIn:
		return "In"
	case TMod:
		return "Mod"
	case TAnd:
		return "And"
	case TOr:
		return "Or"

	// symbols
	case TEqual:
		return "Equal"
	case TLParen:
		return "LParen"
	case TRParen:
		return "RParen"
	case TLess:
		return "Less"
	case TComma:
		return "Comma"
	case TLBrace:
		return "LBrace"
	case TRBrace:
		return "RBrace"
	case TLBracket:
		return "LBracket"
	case TRBracket:
		return "RBracket"
	case TPlus:
		return "Plus"
	case THyphen:
		return "Hyphen"
	case TAsterisk:
		return "Asterisk"
	case TSlash:
		return "Slash"
	case TDot:
		return "Dot"
	}
	return "Unknown"
}

type Token struct {
	Type TokenType
	Text string
	Pos  Pos
}

func (tok *Token) String() string {
	return fmt.Sprintf("%s{\"%s\", %s}", tok.Type, tok.Text, tok.Pos)
}

func (tok *Token) Eq(other *Token) bool {
	if tok == nil || other == nil {
		return tok == other
	}
	return tok.Type == other.Type &&
		tok.Text == other.Text &&
		tok.Pos.Eq(other.Pos)
}

var Symbols = map[rune]TokenType{
	'=': TEqual,
	'<': TLess,
	',': TComma,
	'(': TLParen,
	')': TRParen,
	'[': TLBrace,
	']': TRBrace,
	'{': TLBracket,
	'}': TRBracket,
	'+': TPlus,
	'-': THyphen,
	'*': TAsterisk,
	'/': TSlash,
	'.': TDot,
}

var Keywords = map[string]TokenType{
	"fun":    TFun,
	"return": TReturn,
	"end":    TEnd,
	"while":  TWhile,
	"do":     TDo,
	"if":     TIf,
	"else":   TElse,
	"true":   TTrue,
	"false":  TFalse,
	"in":     TIn,
	"mod":    TMod,
	"and":    TAnd,
	"or":     TOr,
}

var Comments = map[string]string{
	"//": "\n",
	"/*": "*/",
}

func IsIdentLetter(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
