package parser

import (
	"fmt"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/lexer"
)

func (p *Parser) parseToplevelStmt() ([]ast.Stmt, error) {
	// Consume docstring preceding the declaration
	docstring := p.parseDocstring()

	if p.curToken.Type == lexer.TTest {
		stmt, err := p.parseTestStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	isPub := false
	if p.curToken.Type == lexer.TPub {
		isPub = true
		if err := p.readToken(); err != nil {
			return nil, err
		}
	}

	stmts, err := p.parseStmt()
	if err != nil {
		return nil, err
	}

	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			s.Exported = isPub
			s.Docstring = docstring
		case *ast.ExternStmt:
			s.Exported = isPub
			s.Docstring = docstring
		default:
			if isPub {
				return nil, fmt.Errorf("only let, fun, and extern declarations can be exported")
			}
		}
	}
	return stmts, nil
}

func (p *Parser) parseStmt() ([]ast.Stmt, error) {
	if p.curToken.Type == lexer.TEOF {
		return nil, fmt.Errorf("unexpected EOF")
	}

	if p.curToken.Type == lexer.TFun {
		if p.peekToken.Type == lexer.TIdent || p.peekToken.Type == lexer.TRec {
			return p.parseFunDeclStmt()
		}
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

	if p.curToken.Type == lexer.TExtern {
		stmt, err := p.parseExternStmt()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
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
		return nil, fmt.Errorf("return statement is not allowed here (must be inside a function or block)")
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
		switch p.curToken.Type {
		case lexer.TEOF:
			return nil, fmt.Errorf("unexpected eof while reading body")
		case lexer.TEnd, lexer.TElse, lexer.TAnd:
			return body, nil
		}
		stmts, err := p.parseBodyStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, stmts...)
	}
}

func (p *Parser) parseFunDeclStmt() ([]ast.Stmt, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}

	isRec := false
	if p.curToken.Type == lexer.TRec {
		isRec = true
		if err := p.readToken(); err != nil {
			return nil, err
		}
	}

	var funs []*ast.VarDeclStmt
	for {
		name := p.curToken.Text
		if err := p.readToken(); err != nil {
			return nil, err
		}

		if err := p.expect(lexer.TLParen); err != nil {
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

		funs = append(funs, &ast.VarDeclStmt{
			Name: name,
			Body: &ast.FunLiteralExpr{
				Args: args,
				Body: body,
			},
		})

		if isRec && p.curToken.Type == lexer.TAnd {
			if err := p.readToken(); err != nil {
				return nil, err
			}
			continue
		}

		if err := p.expect(lexer.TEnd); err != nil {
			return nil, err
		}
		break
	}

	if isRec {
		return []ast.Stmt{&ast.RecFunDeclStmt{Funs: funs}}, nil
	}
	return []ast.Stmt{funs[0]}, nil
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

func (p *Parser) parseExternStmt() (*ast.ExternStmt, error) {
	if err := p.expect(lexer.TExtern); err != nil {
		return nil, err
	}

	extType := p.curToken.Text
	if err := p.expectString(); err != nil {
		return nil, err
	}

	var name string
	switch p.curToken.Type {
	case lexer.TFun:
		if err := p.readToken(); err != nil {
			return nil, err
		}
		name = p.curToken.Text
		if err := p.expect(lexer.TIdent); err != nil {
			return nil, err
		}
		if err := p.expect(lexer.TLParen); err != nil {
			return nil, err
		}
		_, err := p.parseFunLiteralArgs()
		if err != nil {
			return nil, err
		}
	case lexer.TLet:
		if err := p.readToken(); err != nil {
			return nil, err
		}
		name = p.curToken.Text
		if err := p.expect(lexer.TIdent); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("expected 'fun' or 'let' after extern literal, but got %s", p.curToken.Type)
	}

	return &ast.ExternStmt{
		Type: extType,
		Name: name,
	}, nil
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
				Op:    ast.OpAdd,
				Left:  &ast.VarRefExpr{Name: name},
				Right: expr,
			},
		}, nil
	}
	return &ast.VarAssignStmt{
		Name: name,
		Body: &ast.InfixExpr{
			Op:    ast.OpSub,
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

func (p *Parser) parseTestStmt() (*ast.TestStmt, error) {
	if err := p.expect(lexer.TTest); err != nil {
		return nil, err
	}
	name := p.curToken.Text
	if err := p.expectString(); err != nil {
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
