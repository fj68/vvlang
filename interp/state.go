package interp

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/mod"
	"github.com/fj68/vvlang/parser"
	"github.com/fj68/vvlang/stack"
)

type State struct {
	IsTestMode     bool
	SourcePath     string
	ModuleCache    map[string]*VModule
	Env            *Env
	RetVals        stack.Stack[Value]
	Defers         [][]ast.Expr
	NativeValues   map[string]Value
	BuiltinModules map[string]map[string]Value
	NewState       func(sourcePath string) *State
}

func NewState(sourcePath string) *State {
	return &State{
		SourcePath:     sourcePath,
		ModuleCache:    make(map[string]*VModule),
		Env:            NewEnv(nil),
		NativeValues:   make(map[string]Value),
		BuiltinModules: make(map[string]map[string]Value),
		NewState:       NewState,
	}
}

func (s *State) EnsureSystemLibrary(name string, library fs.FS) error {
	checksum, err := mod.CalculateChecksumLibrary(library)
	if err != nil {
		return err
	}

	cachedPath := mod.GetPackagePath(name)

	cachedFs := os.DirFS(cachedPath)

	cachedChecksum, err := mod.CalculateChecksumLibrary(cachedFs)
	if err == nil && string(cachedChecksum) == checksum {
		return nil
	}

	vf, err := mod.OpenVersionFile()
	if err != nil {
		return err
	}

	return mod.ExtractLibrary(library, vf)
}

func Eval(sourcePath string, text []rune) error {
	s := NewState(sourcePath)
	return s.Eval(text)
}

func (s *State) RegisterNative(name string, value Value) {
	s.NativeValues[name] = value
}

func (s *State) RegisterNatives(values map[string]Value) {
	for name, value := range values {
		s.RegisterNative(name, value)
	}
}

func (s *State) RegisterBuiltinModule(name string, module map[string]Value) {
	s.BuiltinModules[name] = module
}

func (s *State) RegisterBuiltinModules(modules map[string]map[string]Value) {
	for name, module := range modules {
		s.RegisterBuiltinModule(name, module)
	}
}

func (s *State) EvalTest(text []rune) error {
	module, err := parser.Parse(text)
	if err != nil {
		return err
	}
	return s.evalTestModule(module)
}

func (s *State) Eval(text []rune) error {
	module, err := parser.Parse(text)
	if err != nil {
		return err
	}
	return s.evalModule(module)
}

func (s *State) pushDeferScope() {
	s.Defers = append(s.Defers, []ast.Expr{})
}

func (s *State) popDeferScope() error {
	if len(s.Defers) == 0 {
		return nil
	}
	scope := s.Defers[len(s.Defers)-1]
	s.Defers = s.Defers[:len(s.Defers)-1]

	// Execute defers in LIFO order
	for i := len(scope) - 1; i >= 0; i-- {
		_, err := s.evalExpr(scope[i])
		if err != nil {
			return err
		}
	}
	return nil
}

var ErrBreak = fmt.Errorf("break")
var ErrContinue = fmt.Errorf("continue")
var ErrReturn = fmt.Errorf("return")

func (s *State) evalTestModule(module *ast.Module) (err error) {
	s.IsTestMode = true
	s.pushDeferScope()
	defer func() {
		deferErr := s.popDeferScope()
		if err == nil {
			err = deferErr
		}
	}()

	for _, stmt := range module.Statements {
		if _, ok := stmt.(*ast.TestStmt); !ok {
			// run only test stmts, and skip other top-level stmts during test evaluation
			continue
		}
		if err := s.evalStmt(stmt); err != nil {
			if err == ErrReturn {
				// Top-level return: stop program execution but do not treat as an error
				return nil
			}
			return err
		}
	}
	return nil
}

func (s *State) evalModule(module *ast.Module) (err error) {
	s.pushDeferScope()
	defer func() {
		deferErr := s.popDeferScope()
		if err == nil {
			err = deferErr
		}
	}()

	for _, stmt := range module.Statements {
		if err := s.evalStmt(stmt); err != nil {
			if err == ErrReturn {
				// Top-level return: stop program execution but do not treat as an error
				return nil
			}
			return err
		}
	}
	return nil
}

func (s *State) evalBody(body []ast.Stmt) error {
	for _, stmt := range body {
		if err := s.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) evalStmt(stmt ast.Stmt) error {
	switch v := stmt.(type) {
	case *ast.ExprStmt:
		return s.evalExprStmt(v)
	case *ast.ReturnStmt:
		return s.evalReturnStmt(v)
	case *ast.VarDeclStmt:
		return s.evalVarDeclStmt(v)
	case *ast.AssignStmt:
		return s.evalAssignStmt(v)
	case *ast.BlockStmt:
		return s.evalBlockStmt(v)
	case *ast.VarAssignStmt:
		return s.evalVarAssignStmt(v)
	case *ast.IfStmt:
		return s.evalIfStmt(v)
	case *ast.BreakStmt:
		return ErrBreak
	case *ast.ContinueStmt:
		return ErrContinue
	case *ast.WhileStmt:
		return s.evalWhileStmt(v)
	case *ast.TestStmt:
		return s.evalTestStmt(v)
	case *ast.AssertStmt:
		return s.evalAssertStmt(v)
	case *ast.DeferStmt:
		return s.evalDeferStmt(v)
	case *ast.ExternStmt:
		return s.evalExternStmt(v)
	case *ast.ImportStmt:
		return s.evalImportStmt(v)
	default:
		return fmt.Errorf("unknown stmt: %s", v.Inspect())
	}
}

func (s *State) evalReturnStmt(stmt *ast.ReturnStmt) error {
	if stmt.Value == nil {
		s.RetVals.Push(VNull{})
		return ErrReturn
	}
	value, err := s.evalExpr(stmt.Value)
	if err != nil {
		return err
	}
	s.RetVals.Push(value)
	return ErrReturn
}

func (s *State) evalIfStmt(stmt *ast.IfStmt) error {
	v, err := s.evalExpr(stmt.Cond)
	if err != nil {
		return err
	}
	cond, ok := v.(VBool)
	if !ok {
		return fmt.Errorf("expected bool, but got %s", v.Type())
	}
	if cond {
		return s.evalStmt(stmt.Then)
	}
	if stmt.Else == nil {
		return nil
	}
	return s.evalStmt(stmt.Else)
}

func (s *State) evalExprStmt(stmt *ast.ExprStmt) error {
	_, err := s.evalExpr(stmt.Expr)
	if err != nil {
		return err
	}
	return nil
}

func (s *State) evalVarDeclStmt(stmt *ast.VarDeclStmt) error {
	v, err := s.evalExpr(stmt.Body)
	if err != nil {
		return err
	}
	s.Env.Set(stmt.Name, v)
	return nil
}

func (s *State) evalAssignStmt(stmt *ast.AssignStmt) error {
	v, err := s.evalExpr(stmt.Body)
	if err != nil {
		return err
	}
	s.Env.SetOuter(stmt.Name, v)
	return nil
}

func (s *State) evalBlockStmt(stmt *ast.BlockStmt) (err error) {
	oldEnv := s.Env
	s.Env = NewEnv(s.Env)
	defer func() { s.Env = oldEnv }()

	s.pushDeferScope()
	defer func() {
		deferErr := s.popDeferScope()
		if err == nil {
			err = deferErr
		}
	}()

	return s.evalBody(stmt.Body)
}

func (s *State) evalVarAssignStmt(stmt *ast.VarAssignStmt) error {
	v, err := s.evalExpr(stmt.Body)
	if err != nil {
		return err
	}
	s.Env.SetOuter(stmt.Name, v)
	return nil
}

func (s *State) evalExpr(expr ast.Expr) (Value, error) {
	switch v := expr.(type) {
	case *ast.BoolLiteralExpr:
		return VBool(v.Value), nil
	case *ast.NumberLiteralExpr:
		return VNumber(v.Value), nil
	case *ast.StringLiteralExpr:
		return VString(v.Value), nil
	case *ast.RecordLiteralExpr:
		return s.evalRecordLiteralExpr(v)
	case *ast.NullLiteralExpr:
		return VNull{}, nil
	case *ast.FunLiteralExpr:
		return s.evalFunLiteralExpr(v)
	case *ast.FunCallExpr:
		return s.evalFunCallExpr(v)
	case *ast.VarRefExpr:
		return s.evalVarRefExpr(v)
	case *ast.InfixExpr:
		return s.evalInfixExpr(v)
	case *ast.ListLiteralExpr:
		return s.evalListLiteralExpr(v)
	case *ast.IndexExpr:
		return s.evalIndexExpr(v)
	case *ast.SliceExpr:
		return s.evalSliceExpr(v)
	case *ast.PrefixExpr:
		return s.evalPrefixExpr(v)
	case *ast.FieldAccessExpr:
		return s.evalFieldAccessExpr(v)
	case *ast.TypeExpr:
		return s.evalTypeExpr(v)
	case *ast.NotExpr:
		return s.evalNotExpr(v)
	case *ast.StrExpr:
		return s.evalStrExpr(v)
	case *ast.InterpolatedStringLiteralExpr:
		return s.evalInterpolatedStringLiteralExpr(v)
	default:
		return nil, fmt.Errorf("unknown expr: %s", v.Inspect())
	}
}

func (s *State) evalTypeExpr(expr *ast.TypeExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	t := v.Type().String()
	// Remap "number" to "int" or "float" based on value if necessary?
	// The issue says: 'null', 'bool', 'int', 'string', 'float', 'list', 'record'
	if t == "number" {
		if n, ok := v.(VNumber); ok {
			if float64(int64(n)) == float64(n) {
				return VString("int"), nil
			}
			return VString("float"), nil
		}
	}
	return VString(t), nil
}

func (s *State) evalNotExpr(expr *ast.NotExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	b, ok := v.(VBool)
	if !ok {
		return nil, fmt.Errorf("argument for not() is expected bool, but got %s", v.Type())
	}
	return VBool(!bool(b)), nil
}

func (s *State) evalStrExpr(expr *ast.StrExpr) (Value, error) {
	v, err := s.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	return VString(v.Str()), nil
}

func (s *State) evalInterpolatedStringLiteralExpr(expr *ast.InterpolatedStringLiteralExpr) (Value, error) {
	var b strings.Builder
	b.WriteString(expr.Texts[0])
	for i, valueExpr := range expr.Values {
		v, err := s.evalExpr(valueExpr)
		if err != nil {
			return nil, err
		}
		b.WriteString(v.Str())
		b.WriteString(expr.Texts[i+1])
	}
	return VString(b.String()), nil
}

func (s *State) evalFunLiteralExpr(expr *ast.FunLiteralExpr) (Value, error) {
	return &VUserFun{s.Env, expr.Args, expr.Body}, nil
}

func (s *State) evalRecordLiteralExpr(expr *ast.RecordLiteralExpr) (Value, error) {
	m := map[string]Value{}
	for key, field := range expr.Fields {
		val, err := s.evalExpr(field)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return &VRecord{Fields: m}, nil
}

func (s *State) evalFunCallExpr(expr *ast.FunCallExpr) (Value, error) {
	f, err := s.evalExpr(expr.Fun)
	if err != nil {
		return nil, err
	}
	if f, ok := f.(*VUserFun); ok {
		args, err := s.evalArgs(expr.Args)
		if err != nil {
			return nil, err
		}
		return s.callUserFun(f, args)
	}
	if f, ok := f.(VBuiltinFun); ok {
		args, err := s.evalArgs(expr.Args)
		if err != nil {
			return nil, err
		}
		return s.callBuiltinFun(f, args)
	}
	return nil, fmt.Errorf("unable to call %s", f.Type())
}

func (s *State) evalArgs(exprs []ast.Expr) ([]Value, error) {
	var args []Value
	for _, expr := range exprs {
		value, err := s.evalExpr(expr)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}
	return args, nil
}

func (s *State) callUserFun(f *VUserFun, args []Value) (val Value, err error) {
	if len(f.Args) != len(args) {
		return nil, fmt.Errorf("not enough or too much arguments")
	}

	oldEnv := s.Env
	s.Env = NewEnv(f.CapturedEnv)
	defer func() { s.Env = oldEnv }()

	s.pushDeferScope()
	defer func() {
		deferErr := s.popDeferScope()
		if err == nil {
			err = deferErr
		}
	}()

	for i, arg := range args {
		s.Env.Values[f.Args[i]] = arg
	}

	if err = s.evalBody(f.Body); err != nil && err != ErrReturn {
		return nil, err
	}

	return s.RetVals.Pop(), nil
}

func (s *State) evalWhileStmt(stmt *ast.WhileStmt) error {
	for {
		v, err := s.evalExpr(stmt.Cond)
		if err != nil {
			return err
		}
		cond, ok := v.(VBool)
		if !ok {
			return fmt.Errorf("expected bool, but got %s", v.Type())
		}
		if !cond {
			break
		}
		err = s.evalStmt(stmt.Body)
		if err == ErrBreak {
			return nil
		}
		if err == ErrReturn {
			return err
		}
		if err != nil && err != ErrContinue {
			return err
		}
	}
	return nil
}
func (s *State) callBuiltinFun(f VBuiltinFun, args []Value) (Value, error) {
	oldEnv := s.Env
	s.Env = NewEnv(s.Env)
	defer func() { s.Env = oldEnv }()

	return f(s, args)
}

func (s *State) evalVarRefExpr(expr *ast.VarRefExpr) (Value, error) {
	value, err := s.Env.Get(expr.Name)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *State) evalInfixExpr(expr *ast.InfixExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := s.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case "+":
		return s.evalAddExpr(left, right)
	case "-":
		return s.evalSubExpr(left, right)
	case "*":
		return s.evalMulExpr(left, right)
	case "/":
		return s.evalDivExpr(left, right)
	case "==":
		return s.evalEqualExpr(left, right)
	case "<":
		return s.evalLessThanExpr(left, right)
	case "<=":
		return s.evalLessThanEqualExpr(left, right)
	case "and":
		return s.evalAndExpr(left, right)
	case "or":
		return s.evalOrExpr(left, right)
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Op)
	}
}

func (s *State) evalAddExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of add expression is not a number")
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of add expression is not a number")
	}
	return VNumber(lvalue + rvalue), nil
}

func (s *State) evalSubExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of sub expression is not a number")
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of sub expression is not a number")
	}
	return VNumber(lvalue - rvalue), nil
}

func (s *State) evalMulExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of mul expression is not a number")
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of mul expression is not a number")
	}
	return VNumber(lvalue * rvalue), nil
}

func (s *State) evalDivExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VNumber)
	if !ok {
		return nil, fmt.Errorf("left side value of div expression is not a number")
	}
	rvalue, ok := right.(VNumber)
	if !ok {
		return nil, fmt.Errorf("right side value of div expression is not a number")
	}
	if rvalue == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return VNumber(lvalue / rvalue), nil
}

func (s *State) evalEqualExpr(left Value, right Value) (Value, error) {
	v, err := left.Equal(right)
	if err != nil {
		return nil, err
	}
	return VBool(v), nil
}

func (s *State) evalLessThanExpr(left Value, right Value) (Value, error) {
	v, err := left.LessThan(right)
	if err != nil {
		return nil, err
	}
	return VBool(v), nil
}

func (s *State) evalLessThanEqualExpr(left Value, right Value) (Value, error) {
	v, err := s.evalEqualExpr(left, right)
	if err != nil {
		return nil, err
	}
	if bool(v.(VBool)) {
		return v, nil
	}
	return s.evalLessThanExpr(left, right)
}

func (s *State) evalAndExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VBool)
	if !ok {
		return nil, fmt.Errorf("left side of and expr is expected bool, but got %s", left.Type())
	}
	rvalue, ok := right.(VBool)
	if !ok {
		return nil, fmt.Errorf("right side of and expr is expected bool, but got %s", right.Type())
	}
	return VBool(bool(lvalue) && bool(rvalue)), nil
}

func (s *State) evalOrExpr(left Value, right Value) (Value, error) {
	lvalue, ok := left.(VBool)
	if !ok {
		return nil, fmt.Errorf("left side of or expr is expected bool, but got %s", left.Type())
	}
	rvalue, ok := right.(VBool)
	if !ok {
		return nil, fmt.Errorf("right side of or expr is expected bool, but got %s", right.Type())
	}
	return VBool(bool(lvalue) || bool(rvalue)), nil
}

func (s *State) evalPrefixExpr(expr *ast.PrefixExpr) (Value, error) {
	right, err := s.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case "-":
		num, ok := right.(VNumber)
		if !ok {
			return nil, fmt.Errorf("cannot negate %s", right.Type())
		}
		return VNumber(-float64(num)), nil
	default:
		return nil, fmt.Errorf("unknown prefix operator: %s", expr.Op)
	}
}

func (s *State) evalListLiteralExpr(expr *ast.ListLiteralExpr) (Value, error) {
	var elements []Value
	for _, elemExpr := range expr.Elements {
		elem, err := s.evalExpr(elemExpr)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}
	return &VList{Elements: elements}, nil
}

func (s *State) evalIndexExpr(expr *ast.IndexExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	index, err := s.evalExpr(expr.Index)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case *VList:
		idx, ok := index.(VNumber)
		if !ok {
			return nil, fmt.Errorf("list index must be a number, got %s", index.Type())
		}
		// Convert to int, handle negative indices
		intIdx := int(float64(idx))
		if intIdx < 0 {
			intIdx = len(l.Elements) + intIdx
		}
		if intIdx < 0 || intIdx >= len(l.Elements) {
			return nil, fmt.Errorf("list index out of range: %d", intIdx)
		}
		return l.Elements[intIdx], nil
	case VString:
		idx, ok := index.(VNumber)
		if !ok {
			return nil, fmt.Errorf("string index must be a number, got %s", index.Type())
		}
		// Convert to int, handle negative indices
		intIdx := int(float64(idx))
		str := string(l)
		if intIdx < 0 {
			intIdx = len(str) + intIdx
		}
		if intIdx < 0 || intIdx >= len(str) {
			return nil, fmt.Errorf("string index out of range: %d", intIdx)
		}
		return VString(string(str[intIdx])), nil
	default:
		return nil, fmt.Errorf("cannot index %s", left.Type())
	}
}

func (s *State) evalSliceExpr(expr *ast.SliceExpr) (Value, error) {
	left, err := s.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	switch l := left.(type) {
	case *VList:
		var start, end int
		start = 0
		end = len(l.Elements)

		if expr.Start != nil {
			startVal, err := s.evalExpr(expr.Start)
			if err != nil {
				return nil, err
			}
			startNum, ok := startVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice start must be a number, got %s", startVal.Type())
			}
			start = int(float64(startNum))
			if start < 0 {
				start = len(l.Elements) + start
			}
		}

		if expr.End != nil {
			endVal, err := s.evalExpr(expr.End)
			if err != nil {
				return nil, err
			}
			endNum, ok := endVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice end must be a number, got %s", endVal.Type())
			}
			end = int(float64(endNum))
			if end < 0 {
				end = len(l.Elements) + end
			}
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if end > len(l.Elements) {
			end = len(l.Elements)
		}
		if start > end {
			start = end
		}

		return &VList{Elements: l.Elements[start:end]}, nil

	case VString:
		str := string(l)
		var start, end int
		start = 0
		end = len(str)

		if expr.Start != nil {
			startVal, err := s.evalExpr(expr.Start)
			if err != nil {
				return nil, err
			}
			startNum, ok := startVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice start must be a number, got %s", startVal.Type())
			}
			start = int(float64(startNum))
			if start < 0 {
				start = len(str) + start
			}
		}

		if expr.End != nil {
			endVal, err := s.evalExpr(expr.End)
			if err != nil {
				return nil, err
			}
			endNum, ok := endVal.(VNumber)
			if !ok {
				return nil, fmt.Errorf("slice end must be a number, got %s", endVal.Type())
			}
			end = int(float64(endNum))
			if end < 0 {
				end = len(str) + end
			}
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if end > len(str) {
			end = len(str)
		}
		if start > end {
			start = end
		}

		return VString(str[start:end]), nil

	default:
		return nil, fmt.Errorf("cannot slice %s", left.Type())
	}
}

func (s *State) evalFieldAccessExpr(expr *ast.FieldAccessExpr) (Value, error) {
	recordVal, err := s.evalExpr(expr.Record)
	if err != nil {
		return nil, err
	}
	var record *VRecord
	switch tv := recordVal.(type) {
	case *VRecord:
		record = tv
	case *VModule:
		record = tv.VRecord
	default:
		return nil, fmt.Errorf("expected record for field access, but got %s", recordVal.Type())
	}
	fieldName := string(expr.Field)
	value, ok := record.Fields[fieldName]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in record", fieldName)
	}
	return value, nil
}

func (s *State) evalAssertStmt(stmt *ast.AssertStmt) error {
	v, err := s.evalExpr(stmt.Cond)
	if err != nil {
		return err
	}
	cond, ok := v.(VBool)
	if !ok {
		return fmt.Errorf("expected bool in assert condition, but got %s", v.Type())
	}
	if !cond {
		return fmt.Errorf("assertion failed: %s", stmt.Cond.Inspect())
	}
	return nil
}

func (s *State) evalTestStmt(stmt *ast.TestStmt) error {
	// run them only in test mode, and skip during normal evaluation
	if s.IsTestMode {
		return s.evalBody(stmt.Body)
	}
	return nil
}

func (s *State) evalDeferStmt(stmt *ast.DeferStmt) error {
	if len(s.Defers) == 0 {
		return fmt.Errorf("no defer scope available")
	}
	s.Defers[len(s.Defers)-1] = append(s.Defers[len(s.Defers)-1], stmt.Body)
	return nil
}
func (s *State) evalExternStmt(stmt *ast.ExternStmt) error {
	if stmt.Type != "native" {
		return fmt.Errorf("extern: unsupported type %s", stmt.Type)
	}

	val, ok := s.NativeValues[stmt.Name]
	if !ok {
		return fmt.Errorf("extern native: %s not registered", stmt.Name)
	}

	s.Env.Set(stmt.Name, val)

	if stmt.Exported {
		// Mark as exported in the environment if the environment supports it,
		// or handle it during module export collection.
		// Currently VModule collection looks at the environment for names in module.Exports.
		// So we just need to make sure the name is in module.Exports, but that's already
		// handled by the parser (it populates module.Exports).
		// However, we need to ensure the value is present in s.Env.
	}
	return nil
}

func (s *State) evalImportStmt(stmt *ast.ImportStmt) error {
	targetPath, err := mod.ResolveModulePath(s.SourcePath, stmt.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// Check if it's a remote module path.
			// For now, we can infer this if it doesn't start with ./ or ../
			// A more robust check might be needed in mod.IsRemotePath()
			if _, err := mod.ParseRemotePath(stmt.Path); err == nil {
				if getErr := mod.Get(stmt.Path); getErr != nil {
					return fmt.Errorf("module '%s' not found locally and download failed: %v", stmt.Path, getErr)
				}
				// After successful download, resolve path again
				targetPath, err = mod.ResolveModulePath(s.SourcePath, stmt.Path)
				if err != nil {
					return fmt.Errorf("could not resolve module '%s' after download: %v", stmt.Path, err)
				}
			} else {
				return fmt.Errorf("module not found: %s", stmt.Path)
			}
		} else {
			return err
		}
	}

	// Check if already in cache (successful load or currently loading)
	if mod, ok := s.ModuleCache[targetPath]; ok {
		if mod == nil {
			return fmt.Errorf("cyclic dependency detected: %s", targetPath)
		}
		// If we find it in the cache, it's already evaluated
		s.Env.Set(stmt.Alias, mod)
		return nil
	}

	// Read and parse the target file
	text, err := os.ReadFile(targetPath)
	if err != nil {
		return err
	}

	module, err := parser.Parse([]rune(string(text)))
	if err != nil {
		return err
	}

	// Create a new state for the module
	// Pass the same ModuleCache to detect cycles across the whole project
	modState := s.NewState(targetPath)
	modState.ModuleCache = s.ModuleCache
	modState.IsTestMode = s.IsTestMode         // propagate test mode to imported modules
	modState.BuiltinModules = s.BuiltinModules // propagate builtin modules to imported modules

	// To detect cycles: we can add a placeholder in the cache or check a "loading" set
	// For simplicity, let's use the cache with a nil value as a "currently loading" marker
	s.ModuleCache[targetPath] = nil
	defer func() {
		if s.ModuleCache[targetPath] == nil {
			delete(s.ModuleCache, targetPath)
		}
	}()

	// Register module-specific built-ins
	for stdPath, funcs := range modState.BuiltinModules {
		if targetPath == mod.GetPackagePath(stdPath) {
			modState.RegisterNatives(funcs)
			break
		}
	}

	// Evaluate the module
	if err := modState.evalModule(module); err != nil {
		return err
	}

	// Create a VModule for exports, carrying docstring metadata
	fields := make(map[string]Value)
	for name := range module.Exports {
		val, err := modState.Env.Get(name)
		if err != nil {
			return err
		}
		fields[name] = val
	}

	fieldDocstrings := make(map[string]map[string]string)
	for _, stmt := range module.Statements {
		if vd, ok := stmt.(*ast.VarDeclStmt); ok && vd.Exported && vd.Docstring != nil {
			fieldDocstrings[vd.Name] = vd.Docstring
		}
	}

	modRecord := &VModule{
		VRecord:         &VRecord{Fields: fields},
		Docstring:       module.Docstring,
		FieldDocstrings: fieldDocstrings,
	}

	// Cache the result
	s.ModuleCache[targetPath] = modRecord

	// Bind to alias in current env
	s.Env.Set(stmt.Alias, modRecord)

	return nil
}
