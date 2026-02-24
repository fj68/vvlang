package mod

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// WriteToCache ensures the cache folder exists and writes a file to it, concluding with the version file.
func WriteToCache(path string, data []byte, vf *VersionFile) error {
	if err := EnsureGlobalModuleCache(); err != nil {
		return err
	}

	fullPath := GetPackagePath(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return err
	}

	checksum, err := CalculateFileChecksum(fullPath)
	if err != nil {
		return err
	}

	vf.Files[path] = checksum
	return nil
}

// ExtractLibrary walks the FS and extracts all files to the destination directory.
func ExtractLibrary(library fs.FS, vf *VersionFile) error {
	return fs.WalkDir(library, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := library.Open(path)
		if err != nil {
			return err
		}
		defer data.Close()
		dataBytes, err := io.ReadAll(data)
		if err != nil {
			return err
		}

		return WriteToCache(path, dataBytes, vf)
	})
}
