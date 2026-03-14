package cmd

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/mod"
)

func Clean() {
	if len(os.Args) < 3 {
		fmt.Println("usage: vv clean [path]")
		return
	}
	path := os.Args[2]
	cfg := mod.DefaultConfig()
	if err := cfg.Clean(path); err != nil {
		fmt.Println(err)
	}
}
