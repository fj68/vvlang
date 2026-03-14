package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fj68/vvlang/interp"
	"github.com/fj68/vvlang/lib"
	"github.com/fj68/vvlang/mod"
)

func Test() {
	var path string
	if len(os.Args) < 3 {
		path = "."
	} else {
		path = os.Args[2]
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var files []string
	if info.IsDir() {
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip special directories
				name := info.Name()
				if name == ".vv-modules" || name == "vendor" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(p) == ".vv" {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		files = append(files, path)
	}

	if len(files) == 0 {
		fmt.Println("no .vv files found")
		return
	}

	success := true
	for _, file := range files {
		if err, testName := testFile(file); err != nil {
			fmt.Printf("FAIL: %s [%s]\n  %v\n", file, testName, err)
			success = false
		} else {
			fmt.Printf("PASS: %s\n", file)
		}
	}

	if !success {
		os.Exit(1)
	}
}

func testFile(path string) (error, string) {
	text, err := os.ReadFile(path)
	if err != nil {
		return err, ""
	}

	cfg := mod.DefaultConfig()
	s := interp.NewState(cfg, path)
	s.RegisterBuiltinModules(lib.Std.Natives)
	if err := s.EnsureSystemLibrary(lib.Std.Name, lib.Std.FS); err != nil {
		return err, ""
	}

	defer func() {
		if r := recover(); r != nil {
			if s.CurrentTest != nil {
				fmt.Printf("PANIC: %s [%s]\n", path, s.CurrentTest.Name)
			} else {
				fmt.Printf("PANIC: %s\n", path)
			}
			panic(r)
		}
	}()

	err = s.EvalTest([]rune(string(text)))
	return err, s.CurrentTest.Name
}
