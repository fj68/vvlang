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

type IntLiteralExpr struct {
	Position
	Value int64
}

func (expr *IntLiteralExpr) Inspect() string {
	return fmt.Sprintf("IntLiteralExpr{%d}", expr.Value)
}

func (expr *IntLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*IntLiteralExpr)
	if !ok {
		return false
	}
	return expr.Value == o.Value
}

type FloatLiteralExpr struct {
	Position
	Value float64
}

func (expr *FloatLiteralExpr) Inspect() string {
	return fmt.Sprintf("FloatLiteralExpr{%g}", expr.Value)
}

func (expr *FloatLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*FloatLiteralExpr)
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

type CharLiteralExpr struct {
	Position
	Value rune
}

func (expr *CharLiteralExpr) Inspect() string {
	return fmt.Sprintf("CharLiteralExpr{'%c'}", expr.Value)
}

func (expr *CharLiteralExpr) Equals(other Expr) bool {
	o, ok := other.(*CharLiteralExpr)
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

type PostfixExpr struct {
	Position
	Op   string
	Left Expr
}

func (expr *PostfixExpr) Inspect() string {
	return fmt.Sprintf("PostfixExpr{%s, \"%s\"}", expr.Left.Inspect(), expr.Op)
}

func (expr *PostfixExpr) Equals(other Expr) bool {
	o, ok := other.(*PostfixExpr)
	if !ok {
		return false
	}
	return expr.Op == o.Op && expr.Left.Equals(o.Left)
}

type InfixOp int

const (
	OpAdd InfixOp = iota
	OpSub
	OpMul
	OpDiv
	OpIDiv
	OpEqual
	OpLessThan
	OpLessThanEqual
	OpAnd
	OpOr
	OpMod
)

func (op InfixOp) String() string {
	switch op {
	case OpAdd:
		return "add"
	case OpSub:
		return "sub"
	case OpMul:
		return "mul"
	case OpDiv:
		return "div"
	case OpIDiv:
		return "idiv"
	case OpEqual:
		return "equal"
	case OpLessThan:
		return "less_than"
	case OpLessThanEqual:
		return "less_than_equal"
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	case OpMod:
		return "mod"
	}
	return "unknown"
}

type InfixExpr struct {
	Position
	Op    InfixOp
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

type BuiltinCallExpr struct {
	Position
	Op    string
	Value Expr
}

func (expr *BuiltinCallExpr) Inspect() string {
	return fmt.Sprintf("BuiltinCallExpr{\"%s\", %s}", expr.Op, expr.Value.Inspect())
}

func (expr *BuiltinCallExpr) Equals(other Expr) bool {
	o, ok := other.(*BuiltinCallExpr)
	if !ok {
		return false
	}
	return expr.Op == o.Op && expr.Value.Equals(o.Value)
}
