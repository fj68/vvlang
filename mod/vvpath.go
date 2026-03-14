package mod

import (
	"os"
	"path/filepath"
	"strings"
)

// EnsureGlobalModuleCache ensures that the module cache directory exists.
func (c *Config) EnsureGlobalModuleCache() error {
	cacheDir := c.GetCachePath()
	return os.MkdirAll(cacheDir, 0755)
}

func (c *Config) GetVVPath() string {
	return c.VVPath
}

func (c *Config) GetCachePath() string {
	return filepath.Join(c.GetVVPath(), c.CacheDir)
}

func (c *Config) GetPackagePath(name string) string {
	return filepath.Join(c.GetCachePath(), name)
}

func (c *Config) GetVersionPath() string {
	return filepath.Join(c.GetVVPath(), c.SumFile)
}

// FindProjectRoot looks for a ModFile in the current directory or its parents.
func (c *Config) FindProjectRoot(startPath string) (string, error) {
	curr, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(curr); err == nil && !info.IsDir() {
		curr = filepath.Dir(curr)
	}

	for {
		if _, err := os.Stat(filepath.Join(curr, c.ModFile)); err == nil {
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

type ModuleResolver interface {
	Resolve(sourcePath, importPath string) (string, error)
}

type RelativeResolver struct{}

func (r RelativeResolver) Resolve(sourcePath, importPath string) (string, error) {
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(sourcePath)
		target, err := filepath.Abs(filepath.Join(dir, importPath))
		if err != nil {
			return "", err
		}
		_, err = os.Stat(target)
		return target, err
	}
	return "", os.ErrNotExist
}

type VendorResolver struct {
	cfg *Config
}

func (r VendorResolver) Resolve(sourcePath, importPath string) (string, error) {
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return "", os.ErrNotExist
	}

	sourceDir := filepath.Dir(sourcePath)

	// 1. Check VendorDir in the current directory/source directory
	target := filepath.Join(sourceDir, r.cfg.VendorDir, importPath)
	if _, err := os.Stat(target); err == nil {
		return target, nil
	}

	// 2. Search upwards for ModFile and its VendorDir
	if root, rErr := r.cfg.FindProjectRoot(sourceDir); rErr == nil {
		target = filepath.Join(root, r.cfg.VendorDir, importPath)
		if _, err := os.Stat(target); err == nil {
			return target, nil
		}
	}

	return "", os.ErrNotExist
}

type CacheResolver struct {
	cfg *Config
}

func (r CacheResolver) Resolve(sourcePath, importPath string) (string, error) {
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return "", os.ErrNotExist
	}

	target, err := filepath.Abs(r.cfg.GetPackagePath(importPath))
	if err != nil {
		return "", err
	}
	_, err = os.Stat(target)
	return target, err
}

func (c *Config) ResolveModulePath(sourcePath, path string) (target string, err error) {
	resolvers := []ModuleResolver{
		RelativeResolver{},
		VendorResolver{cfg: c},
		CacheResolver{cfg: c},
	}

	for _, r := range resolvers {
		target, err = r.Resolve(sourcePath, path)
		if err == nil {
			return target, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", os.ErrNotExist
}
