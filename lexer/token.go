package lexer

import (
	"fmt"
	"unicode"

	"github.com/fj68/vvlang/ast"
)

type TokenType int

const (
	TEOF TokenType = iota
	TInt
	TFloat
	TIdent
	TLiteral
	TInterpolated
	TComment
	TDocstring

	// keywords
	TFun
	TRec
	TBegin
	TReturn
	TEnd
	TWhile
	TIf
	TElse
	TTrue
	TFalse
	TIn
	TMod
	TAnd
	TOr
	TBreak
	TContinue
	TLet
	TTest
	TAssert
	TAs
	TDefer
	TExtern
	TPub
	TImport
	TFrom
	TNot
	TLen
	TAlso

	// symbols
	TLessEq
	TEqual
	TAssign
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
	TSlashColon
	TDot
	TColon
	TIncr
	TDecr
	TPercent
	TExclam
)

func (ty TokenType) String() string {
	switch ty {
	case TEOF:
		return "EOF"
	case TInt:
		return "Int"
	case TFloat:
		return "Float"
	case TIdent:
		return "Ident"
	case TLiteral:
		return "Literal"
	case TInterpolated:
		return "Interpolated"
	case TComment:
		return "Comment"
	case TDocstring:
		return "Docstring"

		// keywords
	case TFun:
		return "Fun"
	case TRec:
		return "Rec"
	case TBegin:
		return "Begin"
	case TReturn:
		return "Return"
	case TEnd:
		return "End"
	case TWhile:
		return "While"
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
	case TLet:
		return "Let"
	case TBreak:
		return "Break"
	case TContinue:
		return "Continue"
	case TTest:
		return "Test"
	case TAssert:
		return "Assert"
	case TAs:
		return "As"
	case TDefer:
		return "Defer"
	case TExtern:
		return "Extern"
	case TPub:
		return "Pub"
	case TImport:
		return "Import"
	case TFrom:
		return "From"
	case TNot:
		return "Not"
	case TLen:
		return "Len"
	case TAlso:
		return "Also"

	// symbols
	case TLessEq:
		return "LessEq"
	case TAssign:
		return "Assign"
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
	case TSlashColon:
		return "SlashColon"
	case TDot:
		return "Dot"
	case TColon:
		return "Colon"
	case TIncr:
		return "Incr"
	case TDecr:
		return "Decr"
	case TPercent:
		return "Percent"
	case TExclam:
		return "Exclam"
	}
	return "Unknown"
}

type Token struct {
	Type TokenType
	Text string
	Pos  ast.Pos
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
	'=': TAssign,
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
	':': TColon,
	'%': TPercent,
	'!': TExclam,
}

var Symbols2 = map[string]TokenType{
	"<=": TLessEq,
	"==": TEqual,
	"/:": TSlashColon,
	"+=": TIncr,
	"-=": TDecr,
}

var Keywords = map[string]TokenType{
	"fun":      TFun,
	"rec":      TRec,
	"begin":    TBegin,
	"return":   TReturn,
	"end":      TEnd,
	"while":    TWhile,
	"if":       TIf,
	"else":     TElse,
	"true":     TTrue,
	"false":    TFalse,
	"in":       TIn,
	"mod":      TMod,
	"and":      TAnd,
	"or":       TOr,
	"break":    TBreak,
	"continue": TContinue,
	"let":      TLet,
	"test":     TTest,
	"assert":   TAssert,
	"as":       TAs,
	"defer":    TDefer,
	"extern":   TExtern,
	"pub":      TPub,
	"import":   TImport,
	"from":     TFrom,
	"not":      TNot,
	"len":      TLen,
	"also":     TAlso,
}

var Comments = map[string]string{
	"//": "\n",
	"/*": "*/",
}

func IsIdentLetter(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
