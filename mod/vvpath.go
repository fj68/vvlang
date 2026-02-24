package mod

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	VendorDir      = ".vv-modules"
	ProjectModFile = "vv.mod"
	GlobalSumFile  = "vv.sum"
	RemoteCacheDir = ".cache"
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
	return filepath.Join(GetVVPath(), RemoteCacheDir)
}

func GetPackagePath(name string) string {
	return filepath.Join(GetCachePath(), name)
}

func GetVersionPath() string {
	return filepath.Join(GetVVPath(), GlobalSumFile)
}

// FindProjectRoot looks for a ProjectModFile in the current directory or its parents.
func FindProjectRoot(startPath string) (string, error) {
	curr, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(curr); err == nil && !info.IsDir() {
		curr = filepath.Dir(curr)
	}

	for {
		if _, err := os.Stat(filepath.Join(curr, ProjectModFile)); err == nil {
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

	// 1. Check VendorDir in the current directory/source directory
	sourceDir := filepath.Dir(sourcePath)
	target = filepath.Join(sourceDir, VendorDir, path)
	if _, err = os.Stat(target); err == nil {
		return target, nil
	}

	// 2. Search upwards for ProjectModFile and its VendorDir
	if root, rErr := FindProjectRoot(sourceDir); rErr == nil {
		target = filepath.Join(root, VendorDir, path)
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
