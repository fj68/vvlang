package main

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: vv [run|get|clean] path")
		return
	}
	switch os.Args[1] {
	case "get":
		cmd.Get()
	case "clean":
		cmd.Clean()
	default:
		cmd.Run(stdlib)
	}
}

