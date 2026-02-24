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

// FindProjectRoot looks for a vv.mod file in the current directory or its parents.
func FindProjectRoot(startPath string) (string, error) {
	curr, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(curr); err == nil && !info.IsDir() {
		curr = filepath.Dir(curr)
	}

	for {
		if _, err := os.Stat(filepath.Join(curr, "vv.mod")); err == nil {
			return curr, nil
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return "", os.ErrNotExist
}

func ResolveModulePath(sourcePath, path string) (target string, err error) {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		dir := filepath.Dir(sourcePath)
		target, err = filepath.Abs(filepath.Join(dir, path))
		if err != nil {
			return "", err
		}
		_, err = os.Stat(target)
		return target, err
	}

	// 1. Check .vv-modules in the current directory/source directory
	sourceDir := filepath.Dir(sourcePath)
	target = filepath.Join(sourceDir, ".vv-modules", path)
	if _, err = os.Stat(target); err == nil {
		return target, nil
	}

	// 2. Search upwards for vv.mod and its .vv-modules
	if root, rErr := FindProjectRoot(sourceDir); rErr == nil {
		target = filepath.Join(root, ".vv-modules", path)
		if _, err = os.Stat(target); err == nil {
			return target, nil
		}
	}

	// 3. Fallback to $VVPATH/.cache
	target, err = filepath.Abs(GetPackagePath(path))
	if err != nil {
		return "", err
	}
	_, err = os.Stat(target)
	return target, err
}
