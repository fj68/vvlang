package mod

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/parser"
)

// Vendor copies a module from the global cache to the local .vv-modules directory.
// If path is empty, it vendors all project dependencies recursively.
func Vendor(path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectRoot, err := FindProjectRoot(cwd)
	if err != nil {
		projectRoot = cwd
	}

	if path == "" {
		fmt.Println("Collecting dependencies...")
		deps, err := CollectDependencies(projectRoot)
		if err != nil {
			return err
		}
		if len(deps) == 0 {
			fmt.Println("No remote dependencies found.")
			return nil
		}
		for _, dep := range deps {
			if err := Vendor(dep); err != nil {
				return err
			}
		}
		return nil
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

	// Destination path: strip version suffix so imports work
	destBase := filepath.Join(rm.Domain, rm.User, rm.Repo)
	destDir := filepath.Join(projectRoot, VendorDir, destBase)

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
	vvmodPath := filepath.Join(projectRoot, ProjectModFile)
	vendorDir := filepath.Join(projectRoot, VendorDir)

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
			// Prefix with VendorDir for the key in vv.mod?
			// Or just store it as is. Let's match GlobalSumFile style.
			vf.Files[filepath.ToSlash(filepath.Join(VendorDir, rel))] = sum
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

func CollectDependencies(projectRoot string) ([]string, error) {
	foundModules := make(map[string]bool)
	visitedFiles := make(map[string]bool)

	var scan func(path string) error
	scan = func(path string) error {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if visitedFiles[path] {
			return nil
		}
		visitedFiles[path] = true

		if strings.Contains(path, VendorDir) {
			return nil
		}

		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		module, err := parser.Parse([]rune(string(text)))
		if err != nil {
			return fmt.Errorf("error parsing %s: %v", path, err)
		}

		for _, stmt := range module.Statements {
			if imp, ok := stmt.(*ast.ImportStmt); ok {
				if strings.HasPrefix(imp.Path, "./") || strings.HasPrefix(imp.Path, "../") {
					// Relative import: scan the imported file
					target, err := ResolveModulePath(path, imp.Path)
					if err == nil {
						if err := scan(target); err != nil {
							return err
						}
					}
				} else if _, err := ParseRemotePath(imp.Path); err == nil {
					// Remote module
					rm, _ := ParseRemotePath(imp.Path)
					modName := fmt.Sprintf("%s/%s/%s", rm.Domain, rm.User, rm.Repo)
					if rm.Version != "" {
						modName += "@" + rm.Version
					}

					if !foundModules[modName] {
						foundModules[modName] = true
						// Also scan the files in the cached module for cascading dependencies
						cachedPath, err := ResolveModulePath(path, imp.Path)
						if err == nil {
							// For remote modules, we walk the cached directory
							err = filepath.Walk(cachedPath, func(subPath string, info os.FileInfo, err error) error {
								if err != nil {
									return err
								}
								if !info.IsDir() && strings.HasSuffix(subPath, ".vv") {
									return scan(subPath)
								}
								return nil
							})
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
		return nil
	}

	// Initial project walk to find all entry points
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".vv") && !strings.Contains(path, VendorDir) {
			return scan(path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var result []string
	for mod := range foundModules {
		result = append(result, mod)
	}
	return result, nil
}
