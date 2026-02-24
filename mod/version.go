package mod

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type VersionFile struct {
	Files map[string]string `json:"files"`
}

func OpenVersionFile() (*VersionFile, error) {
	if err := EnsureGlobalModuleCache(); err != nil {
		return nil, err
	}
	path := GetVersionPath()
	if _, err := os.Stat(path); err != nil {
		return &VersionFile{
			Files: make(map[string]string),
		}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	versionFile := &VersionFile{
		Files: make(map[string]string),
	}
	if err := json.Unmarshal(data, versionFile); err != nil {
		return nil, err
	}
	return versionFile, nil
}

func (vf *VersionFile) Write() error {
	data, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(GetVersionPath()), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(GetVersionPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

// CalculateFileChecksum computes the MD5 checksum of a single file.
func CalculateFileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateDirectoryChecksum computes the checksum of all files in a physical directory.
func CalculateDirectoryChecksum(dir string) (string, error) {
	hash := md5.New()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(hash, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateChecksumLibrary computes the checksum of all files in an FS.
func CalculateChecksumLibrary(library fs.FS) (string, error) {
	hash := md5.New()
	err := fs.WalkDir(library, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			data, err := library.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()
			if _, err := io.Copy(hash, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
