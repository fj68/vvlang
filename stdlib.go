package main

import (
	"embed"
	"io/fs"
)

//go:embed lib/*
var lib embed.FS

var stdlib, err = fs.Sub(lib, "lib")

func init() {
	if err != nil {
		panic(err)
	}
}
