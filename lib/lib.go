package lib

import (
	"io/fs"

	"github.com/fj68/vvlang/interp"
)

type Lib struct {
	Name    string
	FS      fs.FS
	Natives map[string]map[string]interp.Value
}
