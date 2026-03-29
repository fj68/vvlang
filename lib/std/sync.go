package std

import (
	"fmt"
	"sync"

	"github.com/fj68/vvlang/interp"
)

func SyncMutex(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("argument count mismatch for mutex(): expected 0, got %d", len(args))
	}
	m := &sync.Mutex{}
	return &interp.VUserData{Value: m}, nil
}

func SyncLock(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("argument count mismatch for lock(): expected 1, got %d", len(args))
	}
	ud, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("argument for lock() must be a mutex")
	}
	m, ok := ud.Value.(*sync.Mutex)
	if !ok {
		return nil, fmt.Errorf("argument for lock() must be a mutex")
	}
	m.Lock()
	return interp.NoneValue, nil
}

func SyncUnlock(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("argument count mismatch for unlock(): expected 1, got %d", len(args))
	}
	ud, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("argument for unlock() must be a mutex")
	}
	m, ok := ud.Value.(*sync.Mutex)
	if !ok {
		return nil, fmt.Errorf("argument for unlock() must be a mutex")
	}
	m.Unlock()
	return interp.NoneValue, nil
}
