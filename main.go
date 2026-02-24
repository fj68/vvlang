package main

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: vv [run|get|clean|vendor] path")
		return
	}
	switch os.Args[1] {
	case "get":
		cmd.Get()
	case "vendor":
		cmd.Vendor()
	case "clean":
		cmd.Clean()
	default:
		cmd.Run()
	}
}
