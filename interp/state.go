package interp

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/mod"
	"github.com/fj68/vvlang/parser"
	"github.com/fj68/vvlang/stack"
)

type TestState struct {
	Name string
	Body []ast.Stmt
}

type State struct {
	CurrentTest       *TestState
	SourcePath        string
	ModuleCache       map[string]*VModule
	ScopeManager      ScopeManager
	RetVals           stack.Stack[Value]
	Defers            [][]ast.Stmt
	NativeValues      map[string]Value
	BuiltinModules    map[string]map[string]Value
	NewState          func(cfg *mod.Config, sourcePath string) *State
	MaxRecursionDepth int
	StackDepth        int
	Config            *mod.Config
}

func NewState(cfg *mod.Config, sourcePath string) *State {
	return &State{
		SourcePath:        sourcePath,
		ModuleCache:       make(map[string]*VModule),
		ScopeManager:      NewEnvManager(nil),
		NativeValues:      make(map[string]Value),
		BuiltinModules:    make(map[string]map[string]Value),
		NewState:          NewState,
		MaxRecursionDepth: 1000,
		Config:            cfg,
	}
}

type StateBuilder struct {
	state *State
}

func NewStateBuilder(cfg *mod.Config, sourcePath string) *StateBuilder {
	return &StateBuilder{
		state: NewState(cfg, sourcePath),
	}
}

func (b *StateBuilder) WithNative(name string, value Value) *StateBuilder {
	b.state.NativeValues[name] = value
	return b
}

func (b *StateBuilder) WithNatives(values map[string]Value) *StateBuilder {
	for name, value := range values {
		b.state.NativeValues[name] = value
	}
	return b
}

func (b *StateBuilder) WithModule(name string, module map[string]Value) *StateBuilder {
	b.state.BuiltinModules[name] = module
	return b
}

func (b *StateBuilder) WithBuiltinModules(modules map[string]map[string]Value) *StateBuilder {
	maps.Copy(b.state.BuiltinModules, modules)
	return b
}

func (b *StateBuilder) Build() *State {
	return b.state
}

func (s *State) IsTestMode() bool {
	return s.CurrentTest != nil
}

func (s *State) EnsureSystemLibrary(name string, library fs.FS) error {
	checksum, err := mod.CalculateChecksumLibrary(library)
	if err != nil {
		return err
	}

	cachedPath := s.Config.GetPackagePath(name)

	cachedFs := os.DirFS(cachedPath)

	cachedChecksum, err := mod.CalculateChecksumLibrary(cachedFs)
	if err == nil && string(cachedChecksum) == checksum {
		return nil
	}

	vf, err := s.Config.OpenVersionFile()
	if err != nil {
		return err
	}

	return s.Config.ExtractLibrary(library, vf)
}

func Eval(cfg *mod.Config, sourcePath string, text []rune) error {
	s := NewState(cfg, sourcePath)
	return s.Eval(text)
}

// registerNativesForPath populates NativeValues with any builtin module
// whose logical path (e.g. "std/bool.vv") matches targetPath. The match is
// tried two ways:
//  1. Exact match against the cached path (~/.vv/.cache/std/bool.vv)
//  2. Suffix match on the slash-normalized targetPath
//     (covers running vv test against a local source tree).
func (s *State) registerNativesForPath(targetPath string) {
	normTarget := strings.ReplaceAll(targetPath, "\\", "/")
	for stdPath, funcs := range s.BuiltinModules {
		cachedPath := s.Config.GetPackagePath(stdPath)
		normCached := strings.ReplaceAll(cachedPath, "\\", "/")
		normStd := strings.ReplaceAll(stdPath, "\\", "/")
		if normTarget == normCached || strings.HasSuffix(normTarget, "/"+normStd) {
			maps.Copy(s.NativeValues, funcs)
			break
		}
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
	s.Defers = append(s.Defers, []ast.Stmt{})
}

func (s *State) popDeferScope() error {
	if len(s.Defers) == 0 {
		return nil
	}
	scope := s.Defers[len(s.Defers)-1]
	s.Defers = s.Defers[:len(s.Defers)-1]

	// Execute defers in LIFO order
	for i := len(scope) - 1; i >= 0; i-- {
		err := s.evalStmt(scope[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *State) formatUnhandledError(v Value) string {
	if rec, ok := v.(*VRecord); ok {
		if typ, ok := rec.Fields["type"]; ok && typ.Str() == "error" {
			if val, ok := rec.Fields["value"]; ok {
				return val.Str()
			}
		}
	}
	return v.String()
}

var ErrBreak = fmt.Errorf("break")
var ErrContinue = fmt.Errorf("continue")
var ErrReturn = fmt.Errorf("return")
var ErrNoValue = fmt.Errorf("function did not return a value")

func (s *State) evalTestModule(module *ast.Module) (err error) {
	s.CurrentTest = &TestState{}
	// Register any native functions that belong to the file under test,
	// mirroring what evalImportStmt does for imported modules.
	s.registerNativesForPath(s.SourcePath)
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
					return fmt.Errorf("unhandled error: %s", s.formatUnhandledError(s.RetVals.Pop()))
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
				return fmt.Errorf("unhandled error: %s", s.formatUnhandledError(s.RetVals.Pop()))
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
	case *ast.AssignmentStmt:
		return s.evalAssignmentStmt(v)
	case *ast.BlockStmt:
		return s.evalBlockStmt(v)
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
	v, err := s.evalExprInner(expr)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrNoValue
	}
	return v, nil
}

func (s *State) evalExprInner(expr ast.Expr) (Value, error) {
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
	case *ast.FieldAccessExpr:
		return s.evalFieldAccessExpr(v)
	case *ast.BuiltinCallExpr:
		return s.evalBuiltinCallExpr(v)
	default:
		return nil, fmt.Errorf("unknown expr: %s", v.Inspect())
	}
}
