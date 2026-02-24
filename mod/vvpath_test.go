package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveModulePath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VVPATH", tmpDir)

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
			wantSuffix: filepath.Join(RemoteCacheDir, "github.com", "user", "repo", "lib.vv"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveModulePath(tt.sourcePath, tt.importPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("ResolveModulePath() error = %v, expectErr %v", err, tt.expectErr)
			}

			if len(got) < len(tt.wantSuffix) || got[len(got)-len(tt.wantSuffix):] != tt.wantSuffix {
				t.Errorf("ResolveModulePath() = %v, want suffix %v", got, tt.wantSuffix)
			}
		})
	}
}
