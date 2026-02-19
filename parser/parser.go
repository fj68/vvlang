package parser

import (
	"fmt"
	"strconv"

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
	PEquals
	PLess
	PSum
	PProduct
	PPrefix
	PCall
	PIndex
)

var precedences = map[lexer.TokenType]Precedence{
	lexer.TIdent:    PLowest,
	lexer.TEqual:    PEquals,
	lexer.TLess:     PLess,
	lexer.TPlus:     PSum,
	lexer.THyphen:   PSum,
	lexer.TAsterisk: PProduct,
	lexer.TSlash:    PProduct,
	lexer.TLParen:   PCall,
	lexer.TLBrace:   PIndex,
	lexer.TDot:      PCall,
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
		lexer.TDigit:    p.parseDigitLiteralExpr,
		lexer.TTrue:     p.parseBoolLiteralExpr,
		lexer.TFalse:    p.parseBoolLiteralExpr,
		lexer.TLiteral:  p.parseStringLiteralExpr,
		lexer.THyphen:   p.parsePrefixExpr,
		lexer.TIdent:    p.parseVarRefExpr,
		lexer.TFun:      p.parseFunLiteralExpr,
		lexer.TLBrace:   p.parseListLiteralExpr,
		lexer.TLBracket: p.parseRecordLiteralExpr,
		lexer.TNull:     p.parseNullLiteralExpr,
	}
}

func (p *Parser) registerInfixParsers() {
	p.infixParsers = map[lexer.TokenType]InfixParser{
		lexer.TDot:      p.parseFieldAccessExpr,
		lexer.THyphen:   p.parseInfixExpr,
		lexer.TPlus:     p.parseInfixExpr,
		lexer.TEqual:    p.parseInfixExpr,
		lexer.TLessEq:   p.parseInfixExpr,
		lexer.TLess:     p.parseInfixExpr,
		lexer.TLParen:   p.parseFunCallExpr,
		lexer.TLBrace:   p.parseIndexOrSliceExpr,
		lexer.TAsterisk: p.parseInfixExpr,
		lexer.TSlash:    p.parseInfixExpr,
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

func (p *Parser) expect(ty lexer.TokenType) error {
	if p.curToken.Type != ty {
		return fmt.Errorf("expected %s, but got %s", ty, p.curToken.Type)
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
		stmts, err := p.parseToplevelStmt()
		if err != nil {
			return nil, err
		}
		program = append(program, stmts...)
	}
	return program, nil
}

func (p *Parser) parseToplevelStmt() ([]ast.Stmt, error) {
	if p.curToken.Type == lexer.TTest {
		stmt, err := p.parseTestStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	return p.parseStmt()
}

func (p *Parser) parseStmt() ([]ast.Stmt, error) {
	if p.curToken.Type == lexer.TEOF {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if p.curToken.Type == lexer.TIdent {
		if p.peekToken.Type == lexer.TAssign {
			// `x = expr` form
			stmt, err := p.parseVarAssignStmt()
			if err != nil {
				return nil, err
			}
			return []ast.Stmt{stmt}, nil
		}
		if p.peekToken.Type == lexer.TIncr || p.peekToken.Type == lexer.TDecr {
			// `x += expr` or `x -= expr` form
			stmt, err := p.parseVarIncrDecrStmt()
			if err != nil {
				return nil, err
			}
			return []ast.Stmt{stmt}, nil
		}
	}

	if p.curToken.Type == lexer.TLet {
		return p.parseVarDeclStmt()
	}

	if p.curToken.Type == lexer.TBegin {
		stmt, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TWhile {
		stmt, err := p.parseWhileStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TIf {
		stmt, err := p.parseIfStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TReturn {
		stmt, err := p.parseReturnStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TBreak {
		stmt, err := p.parseBreakStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TContinue {
		stmt, err := p.parseContinueStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TAssert {
		stmt, err := p.parseAssertStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TDefer {
		stmt, err := p.parseDeferStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}

	return []ast.Stmt{&ast.ExprStmt{Expr: expr}}, nil
}

func (p *Parser) parseBodyStmt() ([]ast.Stmt, error) {
	if p.curToken.Type == lexer.TReturn {
		stmt, err := p.parseReturnStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TBreak {
		stmt, err := p.parseBreakStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	if p.curToken.Type == lexer.TContinue {
		stmt, err := p.parseContinueStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	return p.parseStmt()
}

func (p *Parser) parseBody() ([]ast.Stmt, error) {
	var body []ast.Stmt
	for {
		if p.curToken.Type == lexer.TEOF {
			return nil, fmt.Errorf("unexpected eof while reading body")
		}
		if p.curToken.Type == lexer.TEnd {
			break
		}
		if p.curToken.Type == lexer.TElse {
			break
		}
		stmts, err := p.parseBodyStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmts...)
	}
	return body, nil
}

func (p *Parser) parseBlockStmt() (*ast.BlockStmt, error) {
	if err := p.expect(lexer.TBegin); err != nil {
		return nil, err
	}
	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TEnd); err != nil {
		return nil, err
	}
	return &ast.BlockStmt{
		Body: body,
	}, nil
}

func (p *Parser) parseBreakStmt() (*ast.BreakStmt, error) {
	if err := p.expect(lexer.TBreak); err != nil {
		return nil, err
	}
	return &ast.BreakStmt{}, nil
}

func (p *Parser) parseContinueStmt() (*ast.ContinueStmt, error) {
	if err := p.expect(lexer.TContinue); err != nil {
		return nil, err
	}
	return &ast.ContinueStmt{}, nil
}

func (p *Parser) parseDeferStmt() (*ast.DeferStmt, error) {
	if err := p.expect(lexer.TDefer); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	return &ast.DeferStmt{Body: expr}, nil
}

func (p *Parser) parseReturnStmt() (*ast.ReturnStmt, error) {
	// allow `return` without a value (e.g. `return`, or `return` followed by `end`)
	if p.peekToken.Type == lexer.TEnd || p.peekToken.Type == lexer.TEOF {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		return &ast.ReturnStmt{}, nil
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStmt{
		Value: expr,
	}, nil
}

func (p *Parser) parseVarDeclStmt() ([]ast.Stmt, error) {
	if err := p.expect(lexer.TLet); err != nil {
		return nil, err
	}
	if p.curToken.Type == lexer.TLBracket { // '{'
		return p.parseRecordDestructStmt()
	}

	name := p.curToken.Text
	if err := p.expect(lexer.TIdent); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TAssign); err != nil {
		return nil, err
	}
	body, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	return []ast.Stmt{&ast.VarDeclStmt{
		Name: name,
		Body: body,
	}}, nil
}

func (p *Parser) parseRecordDestructStmt() ([]ast.Stmt, error) {
	if err := p.expect(lexer.TLBracket); err != nil {
		return nil, err
	}

	type field struct {
		name  string
		alias string
	}
	var fields []field
	for {
		if p.curToken.Type == lexer.TIdent {
			name := p.curToken.Text
			if err := p.readToken(); err != nil {
				return nil, err
			}
			alias := ""
			if p.curToken.Type == lexer.TAs {
				if err := p.readToken(); err != nil {
					return nil, err
				}
				if p.curToken.Type != lexer.TIdent {
					return nil, fmt.Errorf("expected identifier after 'as', got %s", p.curToken.Type)
				}
				alias = p.curToken.Text
				if err := p.readToken(); err != nil {
					return nil, err
				}
			}
			fields = append(fields, field{name: name, alias: alias})
		}

		if p.curToken.Type == lexer.TRBracket { // '}'
			break
		}

		if err := p.expect(lexer.TComma); err != nil {
			return nil, err
		}

		if p.curToken.Type == lexer.TRBracket { // '}'
			break
		}
	}

	if err := p.expect(lexer.TRBracket); err != nil {
		return nil, err
	}

	if err := p.expect(lexer.TAssign); err != nil {
		return nil, err
	}

	body, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}

	var stmts []ast.Stmt
	for _, f := range fields {
		name := f.name
		if f.alias != "" {
			name = f.alias
		}
		stmts = append(stmts, &ast.VarDeclStmt{
			Name: name,
			Body: &ast.NullLiteralExpr{},
		})
	}

	tempVar := fmt.Sprintf("$temp_%d", p.curToken.Pos.Start)
	blockBody := []ast.Stmt{
		&ast.VarDeclStmt{
			Name: tempVar,
			Body: body,
		},
	}
	for _, f := range fields {
		targetName := f.name
		if f.alias != "" {
			targetName = f.alias
		}
		blockBody = append(blockBody, &ast.AssignStmt{
			Name: targetName,
			Body: &ast.FieldAccessExpr{
				Record: &ast.VarRefExpr{Name: tempVar},
				Field:  f.name,
			},
		})
	}
	stmts = append(stmts, &ast.BlockStmt{Body: blockBody})

	return stmts, nil
}

func (p *Parser) parseVarAssignStmt() (*ast.VarAssignStmt, error) {
	name := p.curToken.Text

	if err := p.expectNext(lexer.TAssign); err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}

	return &ast.VarAssignStmt{
		Name: name,
		Body: expr,
	}, nil
}

func (p *Parser) parseVarIncrDecrStmt() (*ast.VarAssignStmt, error) {
	name := p.curToken.Text
	op := p.peekToken.Type

	if err := p.readToken(); err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}

	if op == lexer.TIncr {
		return &ast.VarAssignStmt{
			Name: name,
			Body: &ast.InfixExpr{
				Op:    "+",
				Left:  &ast.VarRefExpr{Name: name},
				Right: expr,
			},
		}, nil
	}
	return &ast.VarAssignStmt{
		Name: name,
		Body: &ast.InfixExpr{
			Op:    "-",
			Left:  &ast.VarRefExpr{Name: name},
			Right: expr,
		},
	}, nil
}

func (p *Parser) parseWhileStmt() (*ast.WhileStmt, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TEnd); err != nil {
		return nil, err
	}
	return &ast.WhileStmt{
		Cond: cond,
		Body: &ast.BlockStmt{Body: body},
	}, nil
}

func (p *Parser) parseIfStmt() (*ast.IfStmt, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	thenBody, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	var elseBody *ast.BlockStmt
	if p.curToken.Type == lexer.TElse {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		eBody, err := p.parseBody()
		if err != nil {
			return nil, err
		}
		elseBody = &ast.BlockStmt{Body: eBody}
	}
	if err := p.expect(lexer.TEnd); err != nil {
		return nil, err
	}
	return &ast.IfStmt{
		Cond: cond,
		Then: &ast.BlockStmt{Body: thenBody},
		Else: elseBody,
	}, nil
}

func (p *Parser) parseFunLiteralExpr() (ast.Expr, error) {
	var name string
	if p.peekToken.Type == lexer.TIdent {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		name = p.curToken.Text
	}

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

	if err := p.expect(lexer.TEnd); err != nil {
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

	stopTokens := []lexer.TokenType{
		lexer.TEOF,
		lexer.TEnd,
	}
	for !oneOf(stopTokens, p.curToken.Type) && precedence < precedenceOf(p.curToken.Type) {
		infix, ok := p.infixParsers[p.curToken.Type]
		if !ok {
			break
		}
		expr, err = infix(expr)
		if err != nil {
			return nil, err
		}
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
		Fun:  fun,
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
		if err := p.expect(lexer.TComma); err != nil {
			return nil, err
		}
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseIndexOrSliceExpr(left ast.Expr) (ast.Expr, error) {
	// current token is TLBrace ('['
	if err := p.readToken(); err != nil {
		return nil, err
	}

	// Check for slice notation (with optional start)
	// or index notation
	var start ast.Expr
	var end ast.Expr

	// If we see a colon right away, start is nil and we have a slice like [:end]
	if p.curToken.Type == lexer.TColon {
		// start is nil
	} else if p.curToken.Type == lexer.TRBrace {
		// Empty brackets [] - error
		return nil, fmt.Errorf("expected index or slice expression")
	} else {
		// Parse the first expression
		expr, err := p.parseExpr(PLowest)
		if err != nil {
			return nil, err
		}

		// Check if it's a slice or index
		if p.curToken.Type == lexer.TColon {
			start = expr
		} else {
			// It's an index access
			if p.curToken.Type != lexer.TRBrace {
				return nil, fmt.Errorf("expected ']' after index, got %s", p.curToken.Type)
			}
			if err := p.readToken(); err != nil {
				return nil, err
			}
			return &ast.IndexExpr{
				Left:  left,
				Index: expr,
			}, nil
		}
	}

	// Parse the end expression (for slice)
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if p.curToken.Type != lexer.TRBrace {
		expr, err := p.parseExpr(PLowest)
		if err != nil {
			return nil, err
		}
		end = expr
	}

	if p.curToken.Type != lexer.TRBrace {
		return nil, fmt.Errorf("expected ']' after slice, got %s", p.curToken.Type)
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}

	return &ast.SliceExpr{
		Left:  left,
		Start: start,
		End:   end,
	}, nil
}

func (p *Parser) parseNullLiteralExpr() (ast.Expr, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.NullLiteralExpr{}, nil
}

func (p *Parser) parseDigitLiteralExpr() (ast.Expr, error) {
	value, err := strconv.ParseFloat(p.curToken.Text, 64)
	if err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.NumberLiteralExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseBoolLiteralExpr() (ast.Expr, error) {
	value := p.curToken.Type == lexer.TTrue
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.BoolLiteralExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseStringLiteralExpr() (ast.Expr, error) {
	value := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.StringLiteralExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseListLiteralExpr() (ast.Expr, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}

	var elements []ast.Expr
	for {
		if p.curToken.Type == lexer.TEOF {
			return nil, fmt.Errorf("unexpected EOF while reading list")
		}
		if p.curToken.Type == lexer.TRBrace {
			break
		}

		elem, err := p.parseExpr(PLowest)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if p.curToken.Type == lexer.TRBrace {
			break
		}

		if err := p.expect(lexer.TComma); err != nil {
			return nil, err
		}
	}

	if err := p.expect(lexer.TRBrace); err != nil {
		return nil, err
	}

	return &ast.ListLiteralExpr{
		Elements: elements,
	}, nil
}

func (p *Parser) parseRecordLiteralExpr() (ast.Expr, error) {
	// current token is TLBracket ("{")
	if err := p.readToken(); err != nil {
		return nil, err
	}
	fields := map[string]ast.Expr{}
	// empty record
	if p.curToken.Type == lexer.TRBracket {
		if err := p.readToken(); err != nil {
			return nil, err
		}
		return &ast.RecordLiteralExpr{Fields: fields}, nil
	}
	for {
		if p.curToken.Type != lexer.TIdent {
			return nil, fmt.Errorf("expected identifier for record field, got %s", p.curToken.Type)
		}
		name := p.curToken.Text
		if err := p.expectNext(lexer.TAssign); err != nil {
			return nil, err
		}
		if err := p.readToken(); err != nil {
			return nil, err
		}
		expr, err := p.parseExpr(PLowest)
		if err != nil {
			return nil, err
		}
		fields[name] = expr
		if p.curToken.Type == lexer.TRBracket {
			break
		}
		// allow optional trailing comma before the closing brace
		if p.curToken.Type == lexer.TComma {
			// consume comma
			if err := p.readToken(); err != nil {
				return nil, err
			}
			// if next token is closing brace, break (trailing comma)
			if p.curToken.Type == lexer.TRBracket {
				break
			}
			// otherwise continue to next field
			continue
		}
		return nil, fmt.Errorf("expected comma or '}', got %s", p.curToken.Type)
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.RecordLiteralExpr{Fields: fields}, nil
}

func (p *Parser) parseFieldAccessExpr(record ast.Expr) (ast.Expr, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if p.curToken.Type != lexer.TIdent {
		return nil, fmt.Errorf("expected identifier after '.', got %s", p.curToken.Type)
	}
	field := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.FieldAccessExpr{
		Record: record,
		Field:  field,
	}, nil
}

func (p *Parser) parseTestStmt() (*ast.TestStmt, error) {
	if err := p.expect(lexer.TTest); err != nil {
		return nil, err
	}
	name := p.curToken.Text
	if err := p.expect(lexer.TLiteral); err != nil {
		return nil, err
	}
	body, err := p.parseBody()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TEnd); err != nil {
		return nil, err
	}
	return &ast.TestStmt{
		Name: name,
		Body: body,
	}, nil
}

func (p *Parser) parseAssertStmt() (*ast.AssertStmt, error) {
	if err := p.expect(lexer.TAssert); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	return &ast.AssertStmt{
		Cond: cond,
	}, nil
}
