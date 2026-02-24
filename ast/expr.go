package ast

import (
	"fmt"
	"sort"
	"strings"
)

type Expr interface {
	StartPos() *Pos
	EndPos() *Pos
	Inspect() string
	Equals(other Expr) bool
}

type Position struct {
	Start *Pos
	End   *Pos
}

func (p Position) StartPos() *Pos {
	return p.Start
}

func (p Position) EndPos() *Pos {
	return p.End
}

type NumberLiteralExpr struct {
	Position
	Value float64
}

func (expr *NumberLiteralExpr) Inspect() string {
	return fmt.Sprintf("NumberLiteralExpr{%g}", expr.Value)
}

func (expr *NumberLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*NumberLiteralExpr)
	if !ok {
		return false
	}
	return expr.Value == o.Value
}

type BoolLiteralExpr struct {
	Position
	Value bool
}

func (expr *BoolLiteralExpr) Inspect() string {
	return fmt.Sprintf("BoolLiteralExpr{%t}", expr.Value)
}

func (expr *BoolLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*BoolLiteralExpr)
	if !ok {
		return false
	}
	return expr.Value == o.Value
}

type StringLiteralExpr struct {
	Position
	Value string
}

func (expr *StringLiteralExpr) Inspect() string {
	return fmt.Sprintf("StringLiteralExpr{%s}", expr.Value)
}

func (expr *StringLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*StringLiteralExpr)
	if !ok {
		return false
	}
	return expr.Value == o.Value
}

type RecordLiteralExpr struct {
	Position
	Fields map[string]Expr
}

func (expr *RecordLiteralExpr) Inspect() string {
	var keys []string
	for k := range expr.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", k, expr.Fields[k].Inspect()))
	}
	return fmt.Sprintf("RecordLiteralExpr{%s}", strings.Join(parts, ", "))
}

func (expr *RecordLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*RecordLiteralExpr)
	if !ok {
		return false
	}
	if len(expr.Fields) != len(o.Fields) {
		return false
	}
	for k, v := range expr.Fields {
		ov, ok := o.Fields[k]
		if !ok || !v.Equals(ov) {
			return false
		}
	}
	return true
}

type InterpolatedStringLiteralExpr struct {
	Position
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

func (expr *InterpolatedStringLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*InterpolatedStringLiteralExpr)
	if !ok {
		return false
	}
	if len(expr.Texts) != len(o.Texts) || len(expr.Values) != len(o.Values) {
		return false
	}
	for i := range expr.Texts {
		if expr.Texts[i] != o.Texts[i] {
			return false
		}
	}
	for i := range expr.Values {
		if !expr.Values[i].Equals(o.Values[i]) {
			return false
		}
	}
	return true
}

type FunLiteralExpr struct {
	Position
	Args []string
	Body []Stmt
}

func (expr *FunLiteralExpr) Inspect() string {
	var body []string
	for _, s := range expr.Body {
		body = append(body, s.Inspect())
	}
	return fmt.Sprintf("FunLiteralExpr{[%s], [%s]}", strings.Join(expr.Args, ", "), strings.Join(body, ", "))
}

func (expr *FunLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*FunLiteralExpr)
	if !ok {
		return false
	}
	if len(expr.Args) != len(o.Args) || len(expr.Body) != len(o.Body) {
		return false
	}
	for i := range expr.Args {
		if expr.Args[i] != o.Args[i] {
			return false
		}
	}
	for i := range expr.Body {
		if !expr.Body[i].Equals(o.Body[i]) {
			return false
		}
	}
	return true
}

type FunCallExpr struct {
	Position
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

func (expr *FunCallExpr) Equals(other Expr) bool {
	o, ok := other.(*FunCallExpr)
	if !ok {
		return false
	}
	if !expr.Fun.Equals(o.Fun) || len(expr.Args) != len(o.Args) {
		return false
	}
	for i := range expr.Args {
		if !expr.Args[i].Equals(o.Args[i]) {
			return false
		}
	}
	return true
}

type VarRefExpr struct {
	Position
	Name string
}

func (expr *VarRefExpr) Inspect() string {
	return fmt.Sprintf("VarRefExpr{\"%s\"}", expr.Name)
}

func (expr *VarRefExpr) Equals(other Expr) bool {
	o, ok := other.(*VarRefExpr)
	if !ok {
		return false
	}
	return expr.Name == o.Name
}

type PrefixExpr struct {
	Position
	Op    string
	Right Expr
}

func (expr *PrefixExpr) Inspect() string {
	return fmt.Sprintf("PrefixExpr{\"%s\", %s}", expr.Op, expr.Right.Inspect())
}

func (expr *PrefixExpr) Equals(other Expr) bool {
	o, ok := other.(*PrefixExpr)
	if !ok {
		return false
	}
	return expr.Op == o.Op && expr.Right.Equals(o.Right)
}

type InfixExpr struct {
	Position
	Op    string
	Left  Expr
	Right Expr
}

func (expr *InfixExpr) Inspect() string {
	return fmt.Sprintf("InfixExpr{\"%s\", %s, %s}", expr.Op, expr.Left.Inspect(), expr.Right.Inspect())
}

func (expr *InfixExpr) Equals(other Expr) bool {
	o, ok := other.(*InfixExpr)
	if !ok {
		return false
	}
	return expr.Op == o.Op && expr.Left.Equals(o.Left) && expr.Right.Equals(o.Right)
}

type ListLiteralExpr struct {
	Position
	Elements []Expr
}

func (expr *ListLiteralExpr) Inspect() string {
	var elements []string
	for _, elem := range expr.Elements {
		elements = append(elements, elem.Inspect())
	}
	return fmt.Sprintf("ListLiteralExpr{[%s]}", strings.Join(elements, ", "))
}

func (expr *ListLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*ListLiteralExpr)
	if !ok {
		return false
	}
	if len(expr.Elements) != len(o.Elements) {
		return false
	}
	for i := range expr.Elements {
		if !expr.Elements[i].Equals(o.Elements[i]) {
			return false
		}
	}
	return true
}

type IndexExpr struct {
	Position
	Left  Expr
	Index Expr
}

func (expr *IndexExpr) Inspect() string {
	return fmt.Sprintf("IndexExpr{%s, %s}", expr.Left.Inspect(), expr.Index.Inspect())
}

func (expr *IndexExpr) Equals(other Expr) bool {
	o, ok := other.(*IndexExpr)
	if !ok {
		return false
	}
	return expr.Left.Equals(o.Left) && expr.Index.Equals(o.Index)
}

type SliceExpr struct {
	Position
	Left  Expr
	Start Expr
	End   Expr
}

func (expr *SliceExpr) Inspect() string {
	startStr := ""
	if expr.Start != nil {
		startStr = expr.Start.Inspect()
	}
	endStr := ""
	if expr.End != nil {
		endStr = expr.End.Inspect()
	}
	return fmt.Sprintf("SliceExpr{%s, %s, %s}", expr.Left.Inspect(), startStr, endStr)
}

func (expr *SliceExpr) Equals(other Expr) bool {
	o, ok := other.(*SliceExpr)
	if !ok {
		return false
	}
	if !expr.Left.Equals(o.Left) {
		return false
	}
	if (expr.Start == nil) != (o.Start == nil) {
		return false
	}
	if expr.Start != nil && !expr.Start.Equals(o.Start) {
		return false
	}
	if (expr.End == nil) != (o.End == nil) {
		return false
	}
	if expr.End != nil && !expr.End.Equals(o.End) {
		return false
	}
	return true
}

type FieldAccessExpr struct {
	Position
	Record Expr
	Field  string
}

func (expr *FieldAccessExpr) Inspect() string {
	return fmt.Sprintf("FieldAccessExpr{%s.%s}", expr.Record.Inspect(), expr.Field)
}

func (expr *FieldAccessExpr) Equals(other Expr) bool {
	o, ok := other.(*FieldAccessExpr)
	if !ok {
		return false
	}
	return expr.Record.Equals(o.Record) && expr.Field == o.Field
}

type NullLiteralExpr struct {
	Position
}

func (expr *NullLiteralExpr) Inspect() string {
	return "NullLiteralExpr{}"
}

func (expr *NullLiteralExpr) Equals(other Expr) bool {
	_, ok := other.(*NullLiteralExpr)
	return ok
}

type TypeExpr struct {
	Position
	Value Expr
}

func (expr *TypeExpr) Inspect() string {
	return fmt.Sprintf("TypeExpr{%s}", expr.Value.Inspect())
}

func (expr *TypeExpr) Equals(other Expr) bool {
	o, ok := other.(*TypeExpr)
	if !ok {
		return false
	}
	return expr.Value.Equals(o.Value)
}

type NotExpr struct {
	Position
	Value Expr
}

func (expr *NotExpr) Inspect() string {
	return fmt.Sprintf("NotExpr{%s}", expr.Value.Inspect())
}

func (expr *NotExpr) Equals(other Expr) bool {
	o, ok := other.(*NotExpr)
	if !ok {
		return false
	}
	return expr.Value.Equals(o.Value)
}

type StrExpr struct {
	Position
	Value Expr
}

func (expr *StrExpr) Inspect() string {
	return fmt.Sprintf("StrExpr{%s}", expr.Value.Inspect())
}

func (expr *StrExpr) Equals(other Expr) bool {
	o, ok := other.(*StrExpr)
	if !ok {
		return false
	}
	return expr.Value.Equals(o.Value)
}
