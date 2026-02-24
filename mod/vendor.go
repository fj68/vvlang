package mod

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Vendor copies a module from the global cache to the local .vv-modules directory.
// If path is empty, it should ideally vendor all dependencies, but for now we require a path.
func Vendor(path string) error {
	if path == "" {
		// TODO: Implement vendoring all dependencies by walking the project and finding imports.
		return fmt.Errorf("vv vendor: please specify a module path (e.g., vv vendor github.com/user/repo@v1.0.0)")
	}

	rm, err := ParseRemotePath(path)
	if err != nil {
		return err
	}

	// 1. Identify the source in the cache
	repoDirName := rm.Repo
	if rm.Version != "" {
		repoDirName += "@" + rm.Version
	}
	cacheBase := filepath.Join(rm.Domain, rm.User, repoDirName)
	sourceDir := GetPackagePath(cacheBase)
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' not found in cache. Run 'vv get %s' first", path, path)
	}

	// 2. Identify the destination (project root/.vv-modules)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectRoot, err := FindProjectRoot(cwd)
	if err != nil {
		// If vv.mod doesn't exist, use CWD as project root and we'll create vv.mod there.
		projectRoot = cwd
	}

	// Destination path: strip version suffix so imports work
	destBase := filepath.Join(rm.Domain, rm.User, rm.Repo)
	destDir := filepath.Join(projectRoot, ".vv-modules", destBase)

	// 3. Copy files from cache to .vv-modules
	fmt.Printf("Vendoring %s to %s...\n", path, destDir)
	if err := copyDir(sourceDir, destDir); err != nil {
		return err
	}

	// 4. Update vv.mod with checksums
	return updateVVMod(projectRoot)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, path[len(src):]), info.Mode())
		}
		return copyFile(path, filepath.Join(dst, path[len(src):]))
	})
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

func updateVVMod(projectRoot string) error {
	vvmodPath := filepath.Join(projectRoot, "vv.mod")
	vendorDir := filepath.Join(projectRoot, ".vv-modules")

	vf := &VersionFile{
		Files: make(map[string]string),
	}

	if _, err := os.Stat(vendorDir); err == nil {
		err = filepath.Walk(vendorDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			sum, err := CalculateFileChecksum(path)
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(vendorDir, path)
			if err != nil {
				return err
			}
			// Prefix with .vv-modules for the key in vv.mod?
			// Or just store it as is. Let's match files.json style.
			vf.Files[filepath.ToSlash(filepath.Join(".vv-modules", rel))] = sum
			return nil
		})
		if err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(vvmodPath, data, 0644)
}
