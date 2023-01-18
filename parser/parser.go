package parser

import (
	"fmt"

	"github.com/fj68/new-lang/ast"
	"github.com/fj68/new-lang/lexer"
)

func oneOf[T comparable](xs []T, x T) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

type Precedence int

const (
	PLowest Precedence = iota
	PEquals
	PLess
	PSum
	PProduct
	PPrefix
	PCall
)

var precedences = map[lexer.TokenType]Precedence{
	lexer.TIdent:    PLowest,
	lexer.TEqual:    PEquals,
	lexer.TLess:     PLess,
	lexer.TPlus:     PSum,
	lexer.THyphen:   PSum,
	lexer.TAsterisk: PProduct,
	lexer.TSlash:    PProduct,
}

func precedenceOf(ty lexer.TokenType) Precedence {
	if v, ok := precedences[ty]; ok {
		return v
	}
	return PLowest
}

type PrefixParser func() (ast.Expr, error)
type InfixParser func(left ast.Expr) (ast.Expr, error)

type Parser struct {
	lex       *lexer.Lexer
	curToken  *lexer.Token
	peekToken *lexer.Token

	prefixParsers map[lexer.TokenType]PrefixParser
	infixParsers  map[lexer.TokenType]InfixParser
}

func New(text []rune) *Parser {
	p := &Parser{
		lex: lexer.New(text),
	}
	p.registerPrefixParsers()
	p.registerInfixParsers()
	return p
}

func Parse(text []rune) ([]ast.Stmt, error) {
	p := New(text)
	return p.Parse()
}

func (p *Parser) registerPrefixParsers() {
	p.prefixParsers = map[lexer.TokenType]PrefixParser{
		lexer.TDigit:  p.parsePrefixExpr,
		lexer.THyphen: p.parsePrefixExpr,
		lexer.TIdent:  p.parseVarRefExpr,
		lexer.TFun:    p.parseFunLiteralExpr,
	}
}

func (p *Parser) registerInfixParsers() {
	p.infixParsers = map[lexer.TokenType]InfixParser{
		lexer.TDot:    p.parseInfixExpr,
		lexer.THyphen: p.parseInfixExpr,
		lexer.TPlus:   p.parseInfixExpr,
		lexer.TLessEq: p.parseInfixExpr,
		lexer.TLess:   p.parseInfixExpr,
		lexer.TLParen: p.parseFunCallExpr,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, error) {
	return p.parseProgram()
}

func (p *Parser) readToken() error {
	tok, err := p.lex.Next()
	if err != nil {
		return err
	}
	p.curToken = p.peekToken
	p.peekToken = tok
	return nil
}

func (p *Parser) expectNext(ty lexer.TokenType) error {
	if p.peekToken.Type != ty {
		return fmt.Errorf("expected %s, but got %s", ty, p.peekToken.Type)
	}
	if err := p.readToken(); err != nil {
		return err
	}
	return nil
}

func (p *Parser) parseProgram() ([]ast.Stmt, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	var program []ast.Stmt
	for {
		if p.curToken.Type == lexer.TEOF {
			break
		}
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		program = append(program, stmt)
	}
	return program, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	if p.curToken.Type == lexer.TEOF {
		return nil, fmt.Errorf("unexpected EOF")
	}

	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}

	return &ast.ExprStmt{Expr: expr}, nil
}

func (p *Parser) parseBody() ([]ast.Stmt, error) {
	// TODO: prepare p.parseBodyStmt() instead of p.parseStmt()

	var body []ast.Stmt
	for {
		if p.curToken.Type == lexer.TEOF {
			return nil, fmt.Errorf("unexpected eof while reading body")
		}
		if p.curToken.Type == lexer.TEnd {
			break
		}
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
	}
	return body, nil
}

func (p *Parser) parseFunLiteralExpr() (ast.Expr, error) {
	if err := p.expectNext(lexer.TIdent); err != nil {
		return nil, err
	}

	name := p.curToken.Text

	if err := p.expectNext(lexer.TLParen); err != nil {
		return nil, err
	}

	args, err := p.parseFunLiteralArgs()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}

	return &ast.FunLiteralExpr{
		Name: name,
		Args: args,
		Body: body,
	}, nil
}

func (p *Parser) parseFunLiteralArgs() ([]string, error) {
	var args []string
	for {
		if p.peekToken.Type == lexer.TEOF {
			return nil, fmt.Errorf("unexpected eof while reading function arguments")
		}
		if p.peekToken.Type == lexer.TRParen {
			break
		}
		if err := p.expectNext(lexer.TIdent); err != nil {
			return nil, err
		}
		args = append(args, p.curToken.Text)
		if p.peekToken.Type == lexer.TRParen {
			break
		}
		if err := p.expectNext(lexer.TComma); err != nil {
			return nil, err
		}
	}
	// read remaining TRParen
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseExpr(precedence Precedence) (expr ast.Expr, err error) {
	prefix, ok := p.prefixParsers[p.curToken.Type]
	if !ok {
		return nil, fmt.Errorf("no prefix parser found for %s", p.curToken.Type)
	}
	expr, err = prefix()
	if err != nil {
		return nil, err
	}

	for oneOf([]lexer.TokenType{lexer.TEOF, lexer.TDo, lexer.TEnd}, p.peekToken.Type) && precedence < precedenceOf(p.peekToken.Type) {
		infix, ok := p.infixParsers[p.curToken.Type]
		if !ok {
			break
		}
		expr, err = infix(expr)
		if err != nil {
			return nil, err
		}
	}
	if p.peekToken.Type == lexer.TEOF {
		return nil, fmt.Errorf("unexpected eof while reading expr")
	}
	return expr, nil
}

func (p *Parser) parsePrefixExpr() (ast.Expr, error) {
	op := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}
	right, err := p.parseExpr(PPrefix)
	if err != nil {
		return nil, err
	}
	return &ast.PrefixExpr{
		Op:    op,
		Right: right,
	}, nil
}

func (p *Parser) parseInfixExpr(left ast.Expr) (ast.Expr, error) {
	op := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}
	right, err := p.parseExpr(PPrefix)
	if err != nil {
		return nil, err
	}
	return &ast.InfixExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}, nil
}

func (p *Parser) parseVarRefExpr() (ast.Expr, error) {
	name := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.VarRefExpr{Name: name}, nil
}

func (p *Parser) parseFunCallExpr(fun ast.Expr) (ast.Expr, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	args, err := p.parseFunCallArgs()
	if err != nil {
		return nil, err
	}
	return &ast.FunCallExpr{
		Fun: fun,
		Args: args,
	}, nil
}

func (p *Parser) parseFunCallArgs() ([]ast.Expr, error) {
	var args []ast.Expr
	for {
		if p.curToken.Type == lexer.TEOF {
			return nil, fmt.Errorf("unexpected token while reading arguments for function call")
		}
		if p.curToken.Type == lexer.TRParen {
			break
		}
		arg, err := p.parseExpr(PLowest)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.curToken.Type == lexer.TRParen {
			break
		}
		if err := p.expectNext(lexer.TComma); err != nil {
			return nil, err
		}
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return args, nil
}
