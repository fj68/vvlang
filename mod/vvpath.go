package mod

import (
	"os"
	"path/filepath"
	"strings"
)

// EnsureGlobalModuleCache ensures that the $VVPATH/.cache directory exists.
func EnsureGlobalModuleCache() error {
	cacheDir := GetCachePath()
	return os.MkdirAll(cacheDir, 0755)
}

func GetVVPath() string {
	if p := os.Getenv("VVPATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".vv" // fallback to current dir's .vv if home is unknown
	}
	return filepath.Join(home, ".vv")
}

func GetCachePath() string {
	return filepath.Join(GetVVPath(), ".cache")
}

func GetPackagePath(name string) string {
	return filepath.Join(GetCachePath(), name)
}

func GetVersionPath() string {
	return filepath.Join(GetVVPath(), "files.json")
}

func ResolveModulePath(sourcePath, path string) (target string, err error) {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		dir := filepath.Dir(sourcePath)
		target, err = filepath.Abs(filepath.Join(dir, path))
	} else {
		target, err = filepath.Abs(GetPackagePath(path))
	}
	if err != nil {
		return
	}
	// We return the target path even if Stat fails, so the caller can decide
	// whether to download it (if it's a remote module) or report an error.
	if _, err = os.Stat(target); err != nil {
		return target, err
	}
	return
}
