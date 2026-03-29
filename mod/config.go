package mod

import (
	"os"
	"path/filepath"
)

type Config struct {
	VVPath    string
	CacheDir  string
	VendorDir string
	ModFile   string
	SumFile   string
}

func DefaultConfig() *Config {
	vvPath := os.Getenv("VVPATH")
	if vvPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			vvPath = ".vv" // fallback to current dir's .vv if home is unknown
		} else {
			vvPath = filepath.Join(home, ".vv")
		}
	}
	return &Config{
		VVPath:    vvPath,
		CacheDir:  ".cache",
		VendorDir: ".vv-modules",
		ModFile:   "vv.mod",
		SumFile:   "vv.sum",
	}
}
