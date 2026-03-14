package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveModulePath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VVPATH", tmpDir)
	cfg := DefaultConfig()

	// Create a dummy local file
	localFile := filepath.Join(tmpDir, "local.vv")
	if err := os.WriteFile(localFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		sourcePath string
		importPath string
		wantSuffix string
		expectErr  bool
	}{
		{
			name:       "local relative existing",
			sourcePath: filepath.Join(tmpDir, "main.vv"),
			importPath: "./local.vv",
			wantSuffix: "local.vv",
			expectErr:  false,
		},
		{
			name:       "remote path (not cached)",
			sourcePath: filepath.Join(tmpDir, "main.vv"),
			importPath: "github.com/user/repo/lib.vv",
			wantSuffix: filepath.Join(cfg.CacheDir, "github.com", "user", "repo", "lib.vv"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.ResolveModulePath(tt.sourcePath, tt.importPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("ResolveModulePath() error = %v, expectErr %v", err, tt.expectErr)
			}

			if !tt.expectErr && (len(got) < len(tt.wantSuffix) || got[len(got)-len(tt.wantSuffix):] != tt.wantSuffix) {
				t.Errorf("ResolveModulePath() = %v, want suffix %v", got, tt.wantSuffix)
			}
		})
	}
}

func TestRelativeResolver(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.vv")
	os.WriteFile(targetFile, []byte(""), 0644)

	resolver := RelativeResolver{}
	sourcePath := filepath.Join(tmpDir, "main.vv")

	got, err := resolver.Resolve(sourcePath, "./target.vv")
	if err != nil {
		t.Errorf("RelativeResolver.Resolve() error = %v", err)
	}
	if got != targetFile {
		t.Errorf("RelativeResolver.Resolve() = %v, want %v", got, targetFile)
	}

	_, err = resolver.Resolve(sourcePath, "not-relative")
	if err == nil {
		t.Error("RelativeResolver.Resolve() expected error for non-relative path")
	}
}

func TestVendorResolver(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	vendorDir := filepath.Join(tmpDir, cfg.VendorDir)
	os.MkdirAll(vendorDir, 0755)

	targetFile := filepath.Join(vendorDir, "pkg", "lib.vv")
	os.MkdirAll(filepath.Dir(targetFile), 0755)
	os.WriteFile(targetFile, []byte(""), 0644)

	resolver := VendorResolver{cfg: cfg}
	sourcePath := filepath.Join(tmpDir, "main.vv")

	got, err := resolver.Resolve(sourcePath, "pkg/lib.vv")
	if err != nil {
		t.Errorf("VendorResolver.Resolve() error = %v", err)
	}
	if got != targetFile {
		t.Errorf("VendorResolver.Resolve() = %v, want %v", got, targetFile)
	}
}

func TestCacheResolver(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VVPATH", tmpDir)
	cfg := DefaultConfig()

	cachePkgDir := filepath.Join(tmpDir, cfg.CacheDir, "github.com", "user", "repo")
	os.MkdirAll(cachePkgDir, 0755)
	targetFile := filepath.Join(cachePkgDir, "lib.vv")
	os.WriteFile(targetFile, []byte(""), 0644)

	resolver := CacheResolver{cfg: cfg}
	sourcePath := filepath.Join(tmpDir, "main.vv")

	got, err := resolver.Resolve(sourcePath, "github.com/user/repo/lib.vv")
	if err != nil {
		t.Errorf("CacheResolver.Resolve() error = %v", err)
	}
	if got != targetFile {
		t.Errorf("CacheResolver.Resolve() = %v, want %v", got, targetFile)
	}
}
