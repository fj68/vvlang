package interp

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/mod"
	"github.com/fj68/vvlang/parser"
)

func (s *State) evalReturnStmt(stmt *ast.ReturnStmt) error {
	if stmt.Value == nil {
		s.RetVals.Push(NoneValue)
		return ErrReturn
	}

	// Tail call optimization check
	if call, ok := stmt.Value.(*ast.FunCallExpr); ok {
		f, err := s.evalExpr(call.Fun)
		if err != nil {
			return err
		}
		if uf, isUserFun := f.(*VUserFun); isUserFun {
			// We must construct the VTailCall with evaluated arguments
			args, err := s.evalArgs(call.Args)
			if err != nil {
				return err
			}
			s.RetVals.Push(&VTailCall{Fun: uf, Args: args})
			return ErrReturn
		}
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

func (s *State) evalRecFunDeclStmt(stmt *ast.RecFunDeclStmt) error {
	recEnv := NewEnv(s.Env)
	oldEnv := s.Env
	s.Env = recEnv
	defer func() { s.Env = oldEnv }()

	var funs []Value
	for _, decl := range stmt.Funs {
		v, err := s.evalExpr(decl.Body)
		if err != nil {
			return err
		}
		funs = append(funs, v)
		recEnv.Set(decl.Name, v)
	}

	for i, decl := range stmt.Funs {
		oldEnv.Set(decl.Name, funs[i])
	}
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
	return nil
}

func (s *State) evalImportStmt(stmt *ast.ImportStmt) error {
	targetPath, err := mod.ResolveModulePath(s.SourcePath, stmt.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if _, err := mod.ParseRemotePath(stmt.Path); err == nil {
				if getErr := mod.Get(stmt.Path); getErr != nil {
					return fmt.Errorf("module '%s' not found locally and download failed: %v", stmt.Path, getErr)
				}
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

	if modVal, ok := s.ModuleCache[targetPath]; ok {
		if modVal == nil {
			return fmt.Errorf("cyclic dependency detected: %s", targetPath)
		}
		s.Env.Set(stmt.Alias, modVal)
		return nil
	}

	visited := make(map[string]bool)
	visited[s.SourcePath] = true
	if err := s.checkCycles(targetPath, visited); err != nil {
		return err
	}

	text, err := os.ReadFile(targetPath)
	if err != nil {
		return err
	}

	module, err := parser.Parse([]rune(string(text)))
	if err != nil {
		return err
	}

	modState := s.NewState(targetPath)
	modState.ModuleCache = s.ModuleCache
	modState.IsTestMode = s.IsTestMode
	modState.BuiltinModules = s.BuiltinModules

	s.ModuleCache[targetPath] = nil
	defer func() {
		if s.ModuleCache[targetPath] == nil {
			delete(s.ModuleCache, targetPath)
		}
	}()

	for stdPath, funcs := range modState.BuiltinModules {
		if targetPath == mod.GetPackagePath(stdPath) {
			modState.RegisterNatives(funcs)
			break
		}
	}

	if err := modState.evalModule(module); err != nil {
		return err
	}

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

	s.ModuleCache[targetPath] = modRecord
	s.Env.Set(stmt.Alias, modRecord)

	return nil
}

func (s *State) checkCycles(targetPath string, visited map[string]bool) error {
	if visited[targetPath] {
		return fmt.Errorf("cyclic dependency detected: %s", targetPath)
	}
	visited[targetPath] = true
	defer delete(visited, targetPath)

	if mod, ok := s.ModuleCache[targetPath]; ok && mod != nil {
		return nil
	}

	text, err := os.ReadFile(targetPath)
	if err != nil {
		return err
	}

	p := parser.New([]rune(string(text)))
	imports, err := p.ParseImportStmtsOnly()
	if err != nil {
		return err
	}

	for _, imp := range imports {
		subPath, err := mod.ResolveModulePath(targetPath, imp.Path)
		if err != nil {
			if _, remoteErr := mod.ParseRemotePath(imp.Path); remoteErr == nil {
				continue
			}
			return err
		}
		if err := s.checkCycles(subPath, visited); err != nil {
			return err
		}
	}

	return nil
}
