package mod

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRemotePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    *RemoteModule
		wantErr bool
	}{
		{
			name: "github simple",
			path: "github.com/user/repo/main.vv",
			want: &RemoteModule{
				Domain:  "github.com",
				User:    "user",
				Repo:    "repo",
				Version: "",
				File:    "main.vv",
			},
			wantErr: false,
		},
		{
			name: "github with version",
			path: "github.com/user/repo@v1.0.0/lib/math.vv",
			want: &RemoteModule{
				Domain:  "github.com",
				User:    "user",
				Repo:    "repo",
				Version: "v1.0.0",
				File:    "lib/math.vv",
			},
			wantErr: false,
		},
		{
			name: "bitbucket simple",
			path: "bitbucket.org/org/proj/file.vv",
			want: &RemoteModule{
				Domain:  "bitbucket.org",
				User:    "org",
				Repo:    "proj",
				Version: "",
				File:    "file.vv",
			},
			wantErr: false,
		},
		{
			name:    "invalid domain",
			path:    "example.com/user/repo/file.vv",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "path too short",
			path:    "github.com/user",
			want:    nil,
			wantErr: true,
		},
		{
			name: "github with multiple subdirectories",
			path: "github.com/user/repo/lib/subdir/file.vv",
			want: &RemoteModule{
				Domain:  "github.com",
				User:    "user",
				Repo:    "repo",
				Version: "",
				File:    "lib/subdir/file.vv",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRemotePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRemotePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRemotePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockGitClient struct {
	CloneFunc    func(url, path string) error
	CheckoutFunc func(repoPath, version string) error
}

func (m *MockGitClient) Clone(url, path string) error {
	return m.CloneFunc(url, path)
}

func (m *MockGitClient) Checkout(repoPath, version string) error {
	return m.CheckoutFunc(repoPath, version)
}

func TestGet(t *testing.T) {
	vvpath := t.TempDir()
	t.Setenv("VVPATH", vvpath)
	cfg := DefaultConfig()

	originalGitClient := gitClient
	defer func() { gitClient = originalGitClient }()

	mock := &MockGitClient{
		CloneFunc: func(url, path string) error {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(path, "test.vv"), []byte("test"), 0644)
		},
		CheckoutFunc: func(repoPath, version string) error {
			return nil
		},
	}
	gitClient = mock

	path := "github.com/user/repo/test.vv"
	if err := cfg.Get(path); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Check if the file is cached
	cachedFile := filepath.Join(vvpath, cfg.CacheDir, "github.com", "user", "repo", "test.vv")
	if _, err := os.Stat(cachedFile); err != nil {
		t.Errorf("file not cached: %v", err)
	}

	// Check GlobalSumFile
	vf, err := cfg.OpenVersionFile()
	if err != nil {
		t.Fatalf("OpenVersionFile() error = %v", err)
	}
	relPath := filepath.Join("github.com", "user", "repo", "test.vv")
	if _, ok := vf.Files[filepath.ToSlash(relPath)]; !ok {
		t.Errorf("file not found in %s: %s", cfg.SumFile, relPath)
	}
}

func TestClean(t *testing.T) {
	vvpath := t.TempDir()
	t.Setenv("VVPATH", vvpath)
	cfg := DefaultConfig()

	// Create a dummy cached file
	repoDir := filepath.Join(vvpath, cfg.CacheDir, "github.com", "user", "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	cachedFile := filepath.Join(repoDir, "test.vv")
	if err := os.WriteFile(cachedFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add to GlobalSumFile
	vf, err := cfg.OpenVersionFile()
	if err != nil {
		t.Fatal(err)
	}
	relPath := strings.TrimPrefix(cachedFile, cfg.GetCachePath()+string(filepath.Separator))
	vf.Files[filepath.ToSlash(relPath)] = "checksum"
	if err := vf.Write(cfg); err != nil {
		t.Fatal(err)
	}

	path := "github.com/user/repo/test.vv"
	if err := cfg.Clean(path); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}

	// Check if the file is removed
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Errorf("repo directory not removed: %v", err)
	}

	// Check GlobalSumFile
	vf, err = cfg.OpenVersionFile()
	if err != nil {
		t.Fatalf("OpenVersionFile() error = %v", err)
	}
	if _, ok := vf.Files[filepath.ToSlash(relPath)]; ok {
		t.Errorf("file not removed from %s: %s", cfg.SumFile, relPath)
	}
}
