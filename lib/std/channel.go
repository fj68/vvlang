package std

import (
	"fmt"
	"reflect"
	"time"

	"github.com/fj68/vvlang/interp"
)

func ChannelMake(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("argument count mismatch for make(): expected 0, got %d", len(args))
	}
	ch := make(chan interp.Value)
	return &interp.VUserData{Value: ch}, nil
}

func ChannelSend(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch for send(): expected 2, got %d", len(args))
	}
	ud, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("first argument for send() must be a channel")
	}
	ch, ok := ud.Value.(chan interp.Value)
	if !ok {
		return nil, fmt.Errorf("first argument for send() must be a channel")
	}
	ch <- args[1]
	// send doesn't return anything wrapped in Ok per user instructions
	return interp.NoneValue, nil
}

func ChannelRecv(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch for recv(): expected 2, got %d", len(args))
	}
	ud, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("first argument for recv() must be a channel")
	}
	ch, ok := ud.Value.(chan interp.Value)
	if !ok {
		return nil, fmt.Errorf("first argument for recv() must be a channel")
	}
	timeoutVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("second argument for recv() must be an int")
	}
	timeoutMs := int64(timeoutVal)

	if timeoutMs <= 0 {
		// Block indefinitely
		val := <-ch
		return interp.OkValue(val), nil
	}

	select {
	case val := <-ch:
		return interp.OkValue(val), nil
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		return interp.ErrorValue(interp.StringToValue("timeout")), nil
	}
}

func ChannelSelect(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("argument count mismatch for select(): expected 2, got %d", len(args))
	}
	vList, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("first argument for select() must be a list of channels")
	}
	timeoutVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("second argument for select() must be an int")
	}
	timeoutMs := int64(timeoutVal)

	cases := make([]reflect.SelectCase, len(vList.Elements))
	for i, elem := range vList.Elements {
		ud, ok := elem.(*interp.VUserData)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a channel", i)
		}
		ch, ok := ud.Value.(chan interp.Value)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a channel", i)
		}
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	var timeoutCh <-chan time.Time
	if timeoutMs > 0 {
		timeoutCh = time.After(time.Duration(timeoutMs) * time.Millisecond)
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(timeoutCh)})
	}

	chosen, recv, recvOK := reflect.Select(cases)

	if timeoutMs > 0 && chosen == len(cases)-1 {
		// Timeout case
		return interp.ErrorValue(interp.StringToValue("timeout")), nil
	}

	if !recvOK {
		return interp.ErrorValue(interp.StringToValue("channel closed")), nil
	}

	val := recv.Interface().(interp.Value)
	
	record := &interp.VRecord{
		Fields: map[string]interp.Value{
			"index": interp.VInt(chosen),
			"value": val,
		},
	}
	return interp.OkValue(record), nil
}
