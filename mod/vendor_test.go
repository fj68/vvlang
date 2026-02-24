package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()

	projectRoot := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "vv.mod"), []byte("{}"), 0644); err != nil {
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
			got, err := FindProjectRoot(tt.startPath)
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

	// 1. Setup cache
	repoPath := filepath.Join(vvpath, ".cache", "github.com", "user", "repo@v1.0.0")
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
	if err := Vendor("github.com/user/repo@v1.0.0"); err != nil {
		t.Fatalf("Vendor() error = %v", err)
	}

	// 4. Verify results
	destDir := filepath.Join(projectRoot, ".vv-modules", "github.com", "user", "repo")
	if _, err := os.Stat(filepath.Join(destDir, "lib.vv")); err != nil {
		t.Errorf("vendored file not found: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectRoot, "vv.mod")); err != nil {
		t.Errorf("vv.mod not found: %v", err)
	}

	// 5. Test resolution prioritizing vendored
	resolved, err := ResolveModulePath(filepath.Join(projectRoot, "main.vv"), "github.com/user/repo/lib.vv")
	if err != nil {
		t.Fatalf("ResolveModulePath() error = %v", err)
	}
	expected := filepath.Join(destDir, "lib.vv")
	if resolved != expected {
		t.Errorf("ResolveModulePath() = %v, want %v", resolved, expected)
	}
}
