package interp

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/mod"
	"github.com/fj68/vvlang/parser"
	"github.com/fj68/vvlang/stack"
)

type State struct {
	IsTestMode        bool
	SourcePath        string
	ModuleCache       map[string]*VModule
	Env               *Env
	RetVals           stack.Stack[Value]
	Defers            [][]ast.Expr
	NativeValues      map[string]Value
	BuiltinModules    map[string]map[string]Value
	NewState          func(sourcePath string) *State
	MaxRecursionDepth int
	StackDepth        int
}

func NewState(sourcePath string) *State {
	return &State{
		SourcePath:        sourcePath,
		ModuleCache:       make(map[string]*VModule),
		Env:               NewEnv(nil),
		NativeValues:      make(map[string]Value),
		BuiltinModules:    make(map[string]map[string]Value),
		NewState:          NewState,
		MaxRecursionDepth: 1000,
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
		switch stmt.(type) {
		case *ast.TestStmt, *ast.VarDeclStmt, *ast.ExternStmt, *ast.ImportStmt, *ast.RecFunDeclStmt:
			if err := s.evalStmt(stmt); err != nil {
				if err == ErrReturn {
					return fmt.Errorf("top-level return is not allowed")
				}
				return err
			}
		default:
			// skip other top-level stmts during test evaluation
			continue
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
				return fmt.Errorf("top-level return is not allowed")
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
	case *ast.RecFunDeclStmt:
		return s.evalRecFunDeclStmt(v)
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

func (s *State) evalExpr(expr ast.Expr) (Value, error) {
	switch v := expr.(type) {
	case *ast.BoolLiteralExpr:
		return VBool(v.Value), nil
	case *ast.IntLiteralExpr:
		return VInt(v.Value), nil
	case *ast.FloatLiteralExpr:
		return VFloat(v.Value), nil
	case *ast.CharLiteralExpr:
		return VChar(v.Value), nil
	case *ast.RecordLiteralExpr:
		return s.evalRecordLiteralExpr(v)
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
	case *ast.PostfixExpr:
		return s.evalPostfixExpr(v)
	case *ast.FieldAccessExpr:
		return s.evalFieldAccessExpr(v)
	case *ast.BuiltinCallExpr:
		return s.evalBuiltinCallExpr(v)
	default:
		return nil, fmt.Errorf("unknown expr: %s", v.Inspect())
	}
}
