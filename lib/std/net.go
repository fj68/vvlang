package std

import (
	"fmt"
	"io"
	"maps"
	"net"
	"time"

	"github.com/fj68/vvlang/interp"
)

func NetDial(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("dial expects 2 arguments")
	}
	networkVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("dial expects network to be a string")
	}
	addressVal, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("dial expects address to be a string")
	}

	conn, err := net.Dial(networkVal.Str(), addressVal.Str())
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(&interp.VUserData{Value: conn}), nil
}

func NetListen(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("listen expects 2 arguments")
	}
	networkVal, ok := args[0].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("listen expects network to be a string")
	}
	addressVal, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("listen expects address to be a string")
	}

	ln, err := net.Listen(networkVal.Str(), addressVal.Str())
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(&interp.VUserData{Value: ln}), nil
}

func NetAccept(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("accept expects 2 arguments")
	}
	lnVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("accept expects listener to be a userdata")
	}
	ln, ok := lnVal.Value.(net.Listener)
	if !ok {
		return nil, fmt.Errorf("invalid listener object")
	}

	handler := args[1]

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				childState := s.NewState(s.Config, s.SourcePath)
				maps.Copy(childState.NativeValues, s.NativeValues)
				maps.Copy(childState.BuiltinModules, s.BuiltinModules)
				maps.Copy(childState.ModuleCache, s.ModuleCache)

				_, _ = childState.Call(handler, []interp.Value{&interp.VUserData{Value: c}})
			}(conn)
		}
	}()

	return interp.NoneValue, nil
}

func NetRead(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("read expects 2 arguments")
	}
	connVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("read expects conn to be a userdata")
	}
	conn, ok := connVal.Value.(net.Conn)
	if !ok {
		return nil, fmt.Errorf("invalid connection object")
	}
	lenVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("read expects length to be an int")
	}

	buf := make([]byte, lenVal)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}

	ints := make([]interp.Value, n)
	for i := 0; i < n; i++ {
		ints[i] = interp.VInt(int64(buf[i]))
	}

	return interp.OkValue(&interp.VList{Elements: ints}), nil
}

func NetWrite(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("write expects 2 arguments")
	}
	connVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("write expects conn to be a userdata")
	}
	conn, ok := connVal.Value.(net.Conn)
	if !ok {
		return nil, fmt.Errorf("invalid connection object")
	}
	contentVal, ok := args[1].(*interp.VList)
	if !ok {
		return nil, fmt.Errorf("write expects content to be a string (list of chars)")
	}

	buf := make([]byte, len(contentVal.Elements))
	for i, v := range contentVal.Elements {
		if c, ok := v.(interp.VChar); ok {
			buf[i] = byte(c)
		} else if n, ok := v.(interp.VInt); ok {
			buf[i] = byte(n)
		} else {
			return nil, fmt.Errorf("write content must be list of chars or ints, got %s at index %d", v.Type(), i)
		}
	}

	n, err := conn.Write(buf)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}

	return interp.OkValue(interp.VInt(n)), nil
}

func NetClose(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("close expects 1 argument")
	}
	connVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("close expects conn to be a userdata")
	}
	closer, ok := connVal.Value.(io.Closer)
	if !ok {
		return nil, fmt.Errorf("invalid closer object")
	}

	err := closer.Close()
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}

func NetSetReadTimeout(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("set_read_timeout expects 2 arguments")
	}
	connVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("set_read_timeout expects conn to be a userdata")
	}
	conn, ok := connVal.Value.(net.Conn)
	if !ok {
		return nil, fmt.Errorf("invalid connection object")
	}
	msVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("set_read_timeout expects timeout_ms to be an int")
	}

	var d time.Time
	if msVal > 0 {
		d = time.Now().Add(time.Duration(msVal) * time.Millisecond)
	}
	err := conn.SetReadDeadline(d)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}

func NetSetWriteTimeout(s *interp.State, args []interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("set_write_timeout expects 2 arguments")
	}
	connVal, ok := args[0].(*interp.VUserData)
	if !ok {
		return nil, fmt.Errorf("set_write_timeout expects conn to be a userdata")
	}
	conn, ok := connVal.Value.(net.Conn)
	if !ok {
		return nil, fmt.Errorf("invalid connection object")
	}
	msVal, ok := args[1].(interp.VInt)
	if !ok {
		return nil, fmt.Errorf("set_write_timeout expects timeout_ms to be an int")
	}

	var d time.Time
	if msVal > 0 {
		d = time.Now().Add(time.Duration(msVal) * time.Millisecond)
	}
	err := conn.SetWriteDeadline(d)
	if err != nil {
		return interp.ErrorValue(interp.StringToValue(err.Error())), nil
	}
	return interp.OkValue(interp.NoneValue), nil
}
