package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/lexer"
)

func (p *Parser) parseFunLiteralExpr() (ast.Expr, error) {
	if p.curToken.Type == lexer.TFun {
		if err := p.readToken(); err != nil {
			return nil, err
		}
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

	if err := p.expect(lexer.TEnd); err != nil {
		return nil, err
	}

	return &ast.FunLiteralExpr{
		Args: args,
		Body: body,
	}, nil
}

func (p *Parser) parseFunLiteralArgs() ([]string, error) {
	var args []string
	for {
		if p.curToken.Type == lexer.TRParen {
			break
		}
		name := p.curToken.Text
		if err := p.expect(lexer.TIdent); err != nil {
			return nil, err
		}
		args = append(args, name)
		if p.curToken.Type == lexer.TRParen {
			break
		}
		if err := p.expect(lexer.TComma); err != nil {
			return nil, err
		}
	}
	if err := p.expect(lexer.TRParen); err != nil {
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

func (p *Parser) parseInfixOp(op lexer.TokenType) ast.InfixOp {
	switch op {
	case lexer.TPlus:
		return ast.OpAdd
	case lexer.THyphen:
		return ast.OpSub
	case lexer.TAsterisk:
		return ast.OpMul
	case lexer.TSlash:
		return ast.OpDiv
	case lexer.TSlashColon:
		return ast.OpIDiv
	case lexer.TEqual:
		return ast.OpEqual
	case lexer.TLess:
		return ast.OpLessThan
	case lexer.TLessEq:
		return ast.OpLessThanEqual
	case lexer.TAnd:
		return ast.OpAnd
	case lexer.TOr:
		return ast.OpOr
	case lexer.TPercent:
		return ast.OpMod
	default:
		panic(fmt.Sprintf("unknown infix operator: %s", op))
	}
}

func (p *Parser) parseInfixExpr(left ast.Expr) (ast.Expr, error) {
	op := p.parseInfixOp(p.curToken.Type)
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
	switch p.curToken.Type {
	case lexer.TColon:
		// start is nil
	case lexer.TRBrace:
		// Empty brackets [] - error
		return nil, fmt.Errorf("expected index or slice expression")
	default:
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

func (p *Parser) parseBuiltinCallExpr() (ast.Expr, error) {
	op := p.curToken.Text
	pos := p.curToken.Pos
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TLParen); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TRParen); err != nil {
		return nil, err
	}
	return &ast.BuiltinCallExpr{
		Position: ast.Position{Start: &pos, End: &p.curToken.Pos},
		Op:       op,
		Value:    expr,
	}, nil
}

func (p *Parser) parseInterpolatedStringLiteralExpr() (ast.Expr, error) {
	pos := p.curToken.Pos
	text := p.curToken.Text
	if err := p.readToken(); err != nil {
		return nil, err
	}

	runes := []rune(text)
	var currentText strings.Builder

	var expr ast.Expr

	addPart := func(s string) {
		if s == "" && expr != nil {
			return
		}
		part := stringToCharListExpr(s)
		if expr == nil {
			expr = part
		} else {
			expr = &ast.InfixExpr{
				Op:    ast.OpAdd,
				Left:  expr,
				Right: part,
			}
		}
	}

	for i := 0; i < len(runes); i++ {
		if runes[i] == '{' {
			if i+1 < len(runes) && runes[i+1] == '{' {
				currentText.WriteRune('{')
				i++
				continue
			}
			// Start of interpolation
			addPart(currentText.String())
			currentText.Reset()

			i++
			start := i
			braceCount := 1
			for i < len(runes) && braceCount > 0 {
				switch runes[i] {
				case '{':
					braceCount++
				case '}':
					braceCount--
				}
				if braceCount > 0 {
					i++
				}
			}
			if braceCount > 0 {
				return nil, fmt.Errorf("missing closing brace in interpolated string")
			}
			exprText := string(runes[start:i])

			subParser := New([]rune(exprText))
			subParser.readToken()
			subParser.readToken()
			subExpr, err := subParser.parseExpr(PLowest)
			if err != nil {
				return nil, err
			}
			// Wrap in str() to ensure it becomes a list of chars
			subExpr = &ast.BuiltinCallExpr{
				Op:    "str",
				Value: subExpr,
			}
			if expr == nil {
				expr = subExpr
			} else {
				expr = &ast.InfixExpr{
					Op:    ast.OpAdd,
					Left:  expr,
					Right: subExpr,
				}
			}
		} else if runes[i] == '}' {
			if i+1 < len(runes) && runes[i+1] == '}' {
				currentText.WriteRune('}')
				i++
				continue
			}
			currentText.WriteRune('}')
		} else {
			currentText.WriteRune(runes[i])
		}
	}
	addPart(currentText.String())

	if expr == nil {
		return stringToCharListExpr(""), nil
	}

	// Update pos
	switch e := expr.(type) {
	case *ast.ListLiteralExpr:
		e.Start = &pos
		e.End = &p.curToken.Pos
	case *ast.InfixExpr:
		e.Start = &pos
		e.End = &p.curToken.Pos
	}

	return expr, nil
}

func stringToCharListExpr(s string) ast.Expr {
	runes := []rune(s)
	elems := make([]ast.Expr, len(runes))
	for i, r := range runes {
		elems[i] = &ast.CharLiteralExpr{Value: r}
	}
	return &ast.ListLiteralExpr{Elements: elems}
}

func (p *Parser) parseIntLiteralExpr() (ast.Expr, error) {
	value, err := strconv.ParseInt(p.curToken.Text, 10, 64)
	if err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.IntLiteralExpr{
		Value: value,
	}, nil
}

func (p *Parser) parseFloatLiteralExpr() (ast.Expr, error) {
	value, err := strconv.ParseFloat(p.curToken.Text, 64)
	if err != nil {
		return nil, err
	}
	if err := p.readToken(); err != nil {
		return nil, err
	}
	return &ast.FloatLiteralExpr{
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

func (p *Parser) parseLiteralExpr() (ast.Expr, error) {
	text := p.curToken.Text
	runes := []rune(text)
	if err := p.readToken(); err != nil {
		return nil, err
	}
	if len(runes) == 1 {
		return &ast.CharLiteralExpr{Value: runes[0]}, nil
	}
	return stringToCharListExpr(text), nil
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

func (p *Parser) parseGroupedExpr() (ast.Expr, error) {
	if err := p.readToken(); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr(PLowest)
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TRParen); err != nil {
		return nil, err
	}
	return expr, nil
}
