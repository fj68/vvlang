package cmd

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/mod"
)

func Get() {
	if len(os.Args) < 3 {
		fmt.Println("usage: vv get [path]")
		return
	}
	path := os.Args[2]
	if err := mod.Get(path); err != nil {
		fmt.Println(err)
	}
}
