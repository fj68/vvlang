package ast

import (
	"fmt"
	"strings"
)

type Expr interface {
	Inspect() string
}

type NumberLiteralExpr struct {
	Value float64
}

func (expr *NumberLiteralExpr) Inspect() string {
	return fmt.Sprintf("NumberLiteralExpr{%g}", expr.Value)
}

type BoolLiteralExpr struct {
	Value bool
}

func (expr *BoolLiteralExpr) Inspect() string {
	return fmt.Sprintf("BoolLiteralExpr{%t}", expr.Value)
}

type StringLiteralExpr struct {
	Value string
}

func (expr *StringLiteralExpr) Inspect() string {
	return fmt.Sprintf("StringLiteralExpr{%s}", expr.Value)
}

type RecordLiteralExpr struct {
	Fields map[string]Expr
}

func (expr *RecordLiteralExpr) Inspect() string {
	var parts []string
	for k, v := range expr.Fields {
		parts = append(parts, fmt.Sprintf("%s = %s", k, v.Inspect()))
	}
	return fmt.Sprintf("RecordLiteralExpr{%s}", strings.Join(parts, ", "))
}

type InterpolatedStringLiteralExpr struct {
	Texts  []string
	Values []Expr
}

func (expr *InterpolatedStringLiteralExpr) Inspect() string {
	var b strings.Builder
	b.WriteString(expr.Texts[0])
	for i, value := range expr.Values {
		b.WriteRune('{')
		b.WriteString(value.Inspect())
		b.WriteRune('}')
		b.WriteString(expr.Texts[i+1])
	}
	return fmt.Sprintf("InterpolatedStringLiteralExpr{\"%s\"}", b.String())
}

type FunLiteralExpr struct {
	Name string
	Args []string
	Body []Stmt
}

func (expr *FunLiteralExpr) Inspect() string {
	var body []string
	for _, s := range expr.Body {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("FunLiteralExpr{\"%s\", [%s], [%s]}", expr.Name, strings.Join(expr.Args, ", "), strings.Join(body, ", "))
}

type FunCallExpr struct {
	Fun  Expr
	Args []Expr
}

func (expr *FunCallExpr) Inspect() string {
	var args []string
	for _, arg := range expr.Args {
		args = append(args, arg.Inspect())
	}
	return fmt.Sprintf("FunCallExpr{%s, [%s]}", expr.Fun.Inspect(), strings.Join(args, ", "))
}

type VarRefExpr struct {
	Name string
}

func (expr *VarRefExpr) Inspect() string {
	return fmt.Sprintf("VarRefExpr{\"%s\"}", expr.Name)
}

type PrefixExpr struct {
	Op    string
	Right Expr
}

func (expr *PrefixExpr) Inspect() string {
	return fmt.Sprintf("PrefixExpr{\"%s\", %s}", expr.Op, expr.Right.Inspect())
}

type InfixExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (expr *InfixExpr) Inspect() string {
	return fmt.Sprintf("InfixExpr{\"%s\", %s, %s}", expr.Op, expr.Left.Inspect(), expr.Right.Inspect())
}
