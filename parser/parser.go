package parser

import (
	"fmt"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/lexer"
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
	POr
	PAnd
	PEquals
	PLess
	PSum
	PProduct
	PPrefix
	PCall
	PIndex
)

var precedences = map[lexer.TokenType]Precedence{
	lexer.TIdent:      PLowest,
	lexer.TOr:         POr,
	lexer.TAnd:        PAnd,
	lexer.TEqual:      PEquals,
	lexer.TLessEq:     PLess,
	lexer.TLess:       PLess,
	lexer.TPlus:       PSum,
	lexer.THyphen:     PSum,
	lexer.TAsterisk:   PProduct,
	lexer.TSlash:      PProduct,
	lexer.TSlashColon: PProduct,
	lexer.TPercent:    PProduct,
	lexer.TLParen:     PCall,
	lexer.TLBrace:     PIndex,
	lexer.TDot:        PCall,
	lexer.TExclam:     PCall,
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

func Parse(text []rune) (*ast.Module, error) {
	p := New(text)
	return p.Parse()
}

func (p *Parser) registerPrefixParsers() {
	p.prefixParsers = map[lexer.TokenType]PrefixParser{
		lexer.TInt:          p.parseIntLiteralExpr,
		lexer.TFloat:        p.parseFloatLiteralExpr,
		lexer.TTrue:         p.parseBoolLiteralExpr,
		lexer.TFalse:        p.parseBoolLiteralExpr,
		lexer.TLiteral:      p.parseLiteralExpr,
		lexer.THyphen:       p.parsePrefixExpr,
		lexer.TIdent:        p.parseVarRefExpr,
		lexer.TFun:          p.parseFunLiteralExpr,
		lexer.TLBrace:       p.parseListLiteralExpr,
		lexer.TLBracket:     p.parseRecordLiteralExpr,
		lexer.TNot:          p.parseBuiltinCallExpr,
		lexer.TLen:          p.parseBuiltinCallExpr,
		lexer.TInterpolated: p.parseInterpolatedStringLiteralExpr,
		lexer.TLParen:       p.parseGroupedExpr,
	}
}

func (p *Parser) registerInfixParsers() {
	p.infixParsers = map[lexer.TokenType]InfixParser{
		lexer.TDot:        p.parseFieldAccessExpr,
		lexer.THyphen:     p.parseInfixExpr,
		lexer.TPlus:       p.parseInfixExpr,
		lexer.TEqual:      p.parseInfixExpr,
		lexer.TLessEq:     p.parseInfixExpr,
		lexer.TLess:       p.parseInfixExpr,
		lexer.TLParen:     p.parseFunCallExpr,
		lexer.TLBrace:     p.parseIndexOrSliceExpr,
		lexer.TAsterisk:   p.parseInfixExpr,
		lexer.TSlash:      p.parseInfixExpr,
		lexer.TSlashColon: p.parseInfixExpr,
		lexer.TPercent:    p.parseInfixExpr,
		lexer.TAnd:        p.parseInfixExpr,
		lexer.TOr:         p.parseInfixExpr,
		lexer.TExclam:     p.parsePostfixExpr,
	}
}

func (p *Parser) Parse() (*ast.Module, error) {
	return p.parseProgram()
}

func (p *Parser) ParseImportStmtsOnly() ([]*ast.ImportStmt, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	// Collect module-level docstring (appears before imports / first statement)
	// We call parseDocstring to skip it, as in parseProgram.
	_ = p.parseDocstring()

	var imports []*ast.ImportStmt
	for p.curToken.Type == lexer.TImport {
		stmt, err := p.parseImportStmt()
		if err != nil {
			return nil, err
		}
		imports = append(imports, stmt)
	}
	return imports, nil
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

func (p *Parser) expect(ty lexer.TokenType) error {
	if p.curToken.Type != ty {
		return fmt.Errorf("expected %s, but got %s", ty, p.curToken.Type)
	}
	if err := p.readToken(); err != nil {
		return err
	}
	return nil
}

func (p *Parser) expectString() error {
	if p.curToken.Type != lexer.TLiteral && p.curToken.Type != lexer.TInterpolated {
		return fmt.Errorf("expected string literal, but got %s", p.curToken.Type)
	}
	if err := p.readToken(); err != nil {
		return err
	}
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

func (p *Parser) parseDocstring() map[string]string {
	docs := make(map[string]string)
	lang := "en"
	lastLine := -1
	for p.curToken.Type == lexer.TDocstring {
		if lastLine != -1 && p.curToken.Pos.Line > lastLine+1 {
			break
		}
		lastLine = p.curToken.Pos.Line

		line := p.curToken.Text
		if len(line) >= 6 && line[:6] == "@lang " {
			lang = strings.TrimSpace(line[6:])
			// Consume the @lang token and move on
			if err := p.readToken(); err != nil {
				break
			}
			continue
		}
		docs[lang] += line + "\n"
		if err := p.readToken(); err != nil {
			break
		}
	}
	// Trim trailing newlines from each lang entry
	for k, v := range docs {
		docs[k] = strings.TrimRight(v, "\n")
	}
	if len(docs) == 0 {
		return nil
	}
	return docs
}

func (p *Parser) parseProgram() (*ast.Module, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	// Collect module-level docstring (appears before imports / first statement)
	moduleDocstring := p.parseDocstring()

	header, err := p.parseProgramHeader()
	if err != nil {
		return nil, err
	}

	module := &ast.Module{
		Statements: header,
		Exports:    make(map[string]ast.Stmt),
		Docstring:  moduleDocstring,
	}

	seenOtherStmt := false
	for {
		if p.curToken.Type == lexer.TEOF {
			break
		}
		if p.curToken.Type == lexer.TImport {
			return nil, fmt.Errorf("import statement must be at the beginning of the program")
		}

		isExtern := false
		if p.curToken.Type == lexer.TExtern {
			isExtern = true
		} else if p.curToken.Type == lexer.TPub && p.peekToken.Type == lexer.TExtern {
			isExtern = true
		}

		if isExtern {
			if seenOtherStmt {
				return nil, fmt.Errorf("extern statement must be after import statements and before other statements")
			}
		} else if p.curToken.Type != lexer.TDocstring && p.curToken.Type != lexer.TTest {
			seenOtherStmt = true
		}

		stmts, err := p.parseToplevelStmt()
		if err != nil {
			return nil, err
		}
		module.Statements = append(module.Statements, stmts...)
	}

	// Collect exports
	for _, stmt := range module.Statements {
		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			if s.Exported {
				module.Exports[s.Name] = s
			}
		case *ast.ExternStmt:
			if s.Exported {
				module.Exports[s.Name] = s
			}
		}
	}

	return module, nil
}

func (p *Parser) parseProgramHeader() ([]ast.Stmt, error) {
	var header []ast.Stmt
	// imports first
	for p.curToken.Type == lexer.TImport {
		stmt, err := p.parseImportStmt()
		if err != nil {
			return nil, err
		}
		header = append(header, stmt)
	}
	return header, nil
}

func (p *Parser) parseImportStmt() (*ast.ImportStmt, error) {
	if err := p.expect(lexer.TImport); err != nil {
		return nil, err
	}
	alias := p.curToken.Text
	if err := p.expect(lexer.TIdent); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TFrom); err != nil {
		return nil, err
	}
	path := p.curToken.Text
	if err := p.expectString(); err != nil {
		return nil, err
	}
	return &ast.ImportStmt{Alias: alias, Path: path}, nil
}
