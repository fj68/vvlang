package interp

import (
	"fmt"

	"github.com/fj68/vvlang/ast"
)

type Scope interface {
	Declare(name string, value Value)
	Get(name string) (Value, bool)
	Set(name string, value Value) error
	Parent() Scope
	CreateClosure(args []string, body []ast.Stmt) Closure
	Names() []string
}

type Closure interface {
	CapturedScope() Scope
	Args() []string
	Body() []ast.Stmt
}

type ScopeManager interface {
	EnterScope() Scope
	EnterScopeWithParent(parent Scope) Scope
	ExitScope()
	Resolve(name string) (Value, error)
	Assign(name string, value Value) error
	Declare(name string, value Value)
	Current() Scope
}

type LexicalScope struct {
	values map[string]Value
	parent Scope
}

func NewScope(parent Scope) Scope {
	return &LexicalScope{
		values: make(map[string]Value),
		parent: parent,
	}
}

func (s *LexicalScope) Declare(name string, value Value) {
	s.values[name] = value
}

func (s *LexicalScope) Get(name string) (Value, bool) {
	v, ok := s.values[name]
	return v, ok
}

func (s *LexicalScope) Set(name string, value Value) error {
	if _, ok := s.values[name]; ok {
		s.values[name] = value
		return nil
	}
	return fmt.Errorf("variable '%s' not found in this scope", name)
}

func (s *LexicalScope) Parent() Scope {
	return s.parent
}

func (s *LexicalScope) Names() []string {
	names := make([]string, 0, len(s.values))
	for k := range s.values {
		names = append(names, k)
	}
	return names
}

func (s *LexicalScope) CreateClosure(args []string, body []ast.Stmt) Closure {
	return &UserFunctionClosure{
		capturedScope: s,
		args:          args,
		body:          body,
	}
}

type UserFunctionClosure struct {
	capturedScope Scope
	args          []string
	body          []ast.Stmt
}

func (c *UserFunctionClosure) CapturedScope() Scope {
	return c.capturedScope
}

func (c *UserFunctionClosure) Args() []string {
	return c.args
}

func (c *UserFunctionClosure) Body() []ast.Stmt {
	return c.body
}

type EnvManager struct {
	current Scope
}

func NewEnvManager(root Scope) *EnvManager {
	if root == nil {
		root = NewScope(nil)
	}
	return &EnvManager{current: root}
}

func (m *EnvManager) Current() Scope {
	return m.current
}

func (m *EnvManager) EnterScope() Scope {
	s := NewScope(m.current)
	m.current = s
	return s
}

func (m *EnvManager) EnterScopeWithParent(parent Scope) Scope {
	s := NewScope(parent)
	m.current = s
	return s
}

func (m *EnvManager) ExitScope() {
	if m.current.Parent() != nil {
		m.current = m.current.Parent()
	}
}

func (m *EnvManager) Declare(name string, value Value) {
	m.current.Declare(name, value)
}

func (m *EnvManager) Resolve(name string) (Value, error) {
	for s := m.current; s != nil; s = s.Parent() {
		if v, ok := s.Get(name); ok {
			return v, nil
		}
	}
	return nil, fmt.Errorf("variable named '%s' is not found", name)
}

func (m *EnvManager) Assign(name string, value Value) error {
	for s := m.current; s != nil; s = s.Parent() {
		if _, ok := s.Get(name); ok {
			s.Set(name, value)
			return nil
		}
	}
	return fmt.Errorf("variable named '%s' is not defined", name)
}
