package mod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitClient interface {
	Clone(url, path string) error
	Checkout(repoPath, version string) error
}

type CommandGitClient struct{}

func (c *CommandGitClient) Clone(url, path string) error {
	cmd := exec.Command("git", "clone", url, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, string(out))
	}
	return nil
}

func (c *CommandGitClient) Checkout(repoPath, version string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", version)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s failed: %s\n%s", version, err, string(out))
	}
	return nil
}

var gitClient GitClient = &CommandGitClient{}

// We use a map as a set because it's O(1) for lookups and we don't care about the values.
var allowedDomains = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"bitbucket.org": true,
}

type RemoteModule struct {
	Domain  string
	User    string
	Repo    string
	Version string
	File    string
}

// ParseRemotePath parses a string like "github.com/user/repo@v1.0.0/path/to/file.vv"
func ParseRemotePath(path string) (*RemoteModule, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid remote path: %s", path)
	}

	domain := parts[0]
	if !allowedDomains[domain] {
		return nil, fmt.Errorf("domain not allowed: %s", domain)
	}

	user := parts[1]
	repoAndVersion := parts[2]
	repo := repoAndVersion
	version := ""

	if strings.Contains(repoAndVersion, "@") {
		split := strings.SplitN(repoAndVersion, "@", 2)
		repo = split[0]
		version = split[1]
	}

	file := ""
	if len(parts) > 3 {
		file = strings.Join(parts[3:], "/")
	}

	return &RemoteModule{
		Domain:  domain,
		User:    user,
		Repo:    repo,
		Version: version,
		File:    file,
	}, nil
}

// Get downloads the remote module to the cache.
func Get(path string) error {
	rm, err := ParseRemotePath(path)
	if err != nil {
		return err
	}

	// Construct cache directory: $VVPATH/.cache/domain/user/repo[@version]
	repoDirName := rm.Repo
	if rm.Version != "" {
		repoDirName += "@" + rm.Version
	}

	// The base path for the repository in the cache
	cacheBase := filepath.Join(rm.Domain, rm.User, repoDirName)
	fullCacheDir := GetPackagePath(cacheBase)

	// Check if already exists
	if _, err := os.Stat(fullCacheDir); err == nil {
		return nil // Already cached
	}

	if err := EnsureGlobalModuleCache(); err != nil {
		return err
	}

	// Clone URL
	repoURL := fmt.Sprintf("https://%s/%s/%s.git", rm.Domain, rm.User, rm.Repo)

	fmt.Printf("Downloading %s...\n", repoURL)

	// Git clone
	if err := gitClient.Clone(repoURL, fullCacheDir); err != nil {
		return err
	}

	// Checkout version if specified
	if rm.Version != "" {
		if err := gitClient.Checkout(fullCacheDir, rm.Version); err != nil {
			return err
		}
	}

	// Update GlobalSumFile with checksums for the downloaded files
	vf, err := OpenVersionFile()
	if err != nil {
		return err
	}

	// Walk the downloaded repository and register all files in files.json
	err = filepath.Walk(fullCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		sum, err := CalculateFileChecksum(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(GetCachePath(), path)
		if err != nil {
			return err
		}
		vf.Files[filepath.ToSlash(rel)] = sum
		return nil
	})
	if err != nil {
		return err
	}

	return vf.Write()
}

func Clean(path string) error {
	rm, err := ParseRemotePath(path)
	if err != nil {
		return err
	}

	repoDirName := rm.Repo
	if rm.Version != "" {
		repoDirName += "@" + rm.Version
	}

	cacheBase := filepath.Join(rm.Domain, rm.User, repoDirName)
	fullCacheDir := GetPackagePath(cacheBase)

	if _, err := os.Stat(fullCacheDir); os.IsNotExist(err) {
		return nil // Not cached, nothing to do
	}

	fmt.Printf("Cleaning %s...\n", fullCacheDir)

	if err := os.RemoveAll(fullCacheDir); err != nil {
		return err
	}

	vf, err := OpenVersionFile()
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(GetCachePath(), fullCacheDir)
	if err != nil {
		return err
	}

	prefix := filepath.ToSlash(rel)
	for k := range vf.Files {
		if strings.HasPrefix(k, prefix) {
			delete(vf.Files, k)
		}
	}

	return vf.Write()
}
