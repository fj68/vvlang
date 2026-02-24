package mod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	repoPart := parts[2]
	version := ""

	// Check for version tag in repo name
	if strings.Contains(repoPart, "@") {
		repoSplit := strings.SplitN(repoPart, "@", 2)
		repoPart = repoSplit[0]
		version = repoSplit[1]
	}

	file := ""
	if len(parts) > 3 {
		file = strings.Join(parts[3:], "/")
	}

	return &RemoteModule{
		Domain:  domain,
		User:    user,
		Repo:    repoPart,
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
	cmd := exec.Command("git", "clone", repoURL, fullCacheDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, string(out))
	}

	// Checkout version if specified
	if rm.Version != "" {
		cmd := exec.Command("git", "-C", fullCacheDir, "checkout", rm.Version)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git checkout %s failed: %s\n%s", rm.Version, err, string(out))
		}
	}

	// Update files.json with checksums for the downloaded files
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

		sum, err := CalculateChecksum(path)
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
