package mod

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()

	projectRoot := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, cfg.ModFile), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(projectRoot, "src", "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startPath string
		want      string
		wantErr   bool
	}{
		{
			name:      "at root",
			startPath: projectRoot,
			want:      projectRoot,
			wantErr:   false,
		},
		{
			name:      "in subdir",
			startPath: subDir,
			want:      projectRoot,
			wantErr:   false,
		},
		{
			name:      "outside",
			startPath: tmpDir,
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.FindProjectRoot(tt.startPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProjectRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindProjectRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVendor(t *testing.T) {
	vvpath := t.TempDir()
	t.Setenv("VVPATH", vvpath)
	cfg := DefaultConfig()

	// 1. Setup cache
	repoPath := filepath.Join(vvpath, cfg.CacheDir, "github.com", "user", "repo@v1.0.0")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, "lib.vv"), []byte("module code"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Setup project
	projectRoot := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldCwd)

	// 3. Run Vendor
	if err := cfg.Vendor("github.com/user/repo@v1.0.0"); err != nil {
		t.Fatalf("Vendor() error = %v", err)
	}

	// 4. Verify results
	destDir := filepath.Join(projectRoot, cfg.VendorDir, "github.com", "user", "repo")
	if _, err := os.Stat(filepath.Join(destDir, "lib.vv")); err != nil {
		t.Errorf("vendored file not found: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, cfg.ModFile)); err != nil {
		t.Errorf("%s not found: %v", cfg.ModFile, err)
	}

	// 5. Test resolution prioritizing vendored
	resolved, err := cfg.ResolveModulePath(filepath.Join(projectRoot, "main.vv"), "github.com/user/repo/lib.vv")
	if err != nil {
		t.Fatalf("ResolveModulePath() error = %v", err)
	}
	expected := filepath.Join(destDir, "lib.vv")
	if resolved != expected {
		t.Errorf("ResolveModulePath() = %v, want %v", resolved, expected)
	}
}

func TestCollectDependenciesCascading(t *testing.T) {
	vvpath := t.TempDir()
	t.Setenv("VVPATH", vvpath)
	cfg := DefaultConfig()

	// Setup cached modules
	// ModA -> ModB
	modAPath := filepath.Join(vvpath, cfg.CacheDir, "github.com", "user", "modA")
	modBPath := filepath.Join(vvpath, cfg.CacheDir, "github.com", "user", "modB")
	os.MkdirAll(modAPath, 0755)
	os.MkdirAll(modBPath, 0755)

	os.WriteFile(filepath.Join(modAPath, "a.vv"), []byte(`import b from "github.com/user/modB/b.vv"`), 0644)
	os.WriteFile(filepath.Join(modBPath, "b.vv"), []byte(`print("hello from B")`), 0644)

	// Setup project
	projectRoot := t.TempDir()
	os.WriteFile(filepath.Join(projectRoot, "main.vv"), []byte(`import a from "github.com/user/modA/a.vv"`), 0644)

	deps, err := cfg.CollectDependencies(projectRoot)
	if err != nil {
		t.Fatalf("CollectDependencies() error = %v", err)
	}

	sort.Strings(deps)
	expected := []string{"github.com/user/modA", "github.com/user/modB"}
	sort.Strings(expected)

	if len(deps) != len(expected) {
		t.Errorf("got %d deps, want %d", len(deps), len(expected))
	}
	for i := range deps {
		if deps[i] != expected[i] {
			t.Errorf("deps[%d] = %v, want %v", i, deps[i], expected[i])
		}
	}
}
