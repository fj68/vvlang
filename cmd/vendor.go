package cmd

import (
	"fmt"
	"os"

	"github.com/fj68/vvlang/mod"
)

func Vendor() {
	path := ""
	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	if err := mod.Vendor(path); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
