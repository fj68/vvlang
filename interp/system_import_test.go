package interp

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/fj68/vvlang/mod"
)

func TestSystemImport(t *testing.T) {
	// 1. Test relative import
	t.Run("RelativeImport", func(t *testing.T) {
		// prepare files
		tmpDir := t.TempDir()
		mainFile := filepath.Join(tmpDir, "main.vv")
		libFile := filepath.Join(tmpDir, "lib.vv")

		err := os.WriteFile(libFile, []byte("pub let x = 42"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		s := NewState(mod.DefaultConfig(), mainFile)
		err = s.Eval([]rune("import l from './lib.vv' assert l.x == 42"))
		if err != nil {
			t.Fatalf("Relative import failed: %v", err)
		}
	})

	// 2. Test system import and automatic extraction
	t.Run("SystemImportAndExtraction", func(t *testing.T) {
		// prepare cache dir
		tmpDir := t.TempDir()
		t.Setenv("VVPATH", tmpDir)

		// prepare embedded fs
		mapFs := fstest.MapFS{
			"std/math.vv": &fstest.MapFile{
				Data: []byte("pub fun add(a, b) return a + b end"),
			},
			"std/console.vv": &fstest.MapFile{
				Data: []byte("pub extern \"native\" fun print(v)"),
			},
		}

		s := NewState(mod.DefaultConfig(), "main.vv")
		s.EnsureSystemLibrary("std", mapFs)

		err := s.Eval([]rune("import m from 'std/math.vv' assert m.add(1, 2) == 3"))
		if err != nil {
			t.Fatalf("System import failed: %v", err)
		}
	})

	// 3. Test Checksum Mismatch Restoration
	t.Run("ChecksumMismatchRestoration", func(t *testing.T) {
		// prepare cache dir
		tmpDir := t.TempDir()
		cfg := mod.DefaultConfig()
		cfg.VVPath = filepath.Join(tmpDir, ".vv")
		t.Setenv("VVPATH", cfg.VVPath)
		mathFile := filepath.Join(cfg.GetCachePath(), "std/math.vv")
		if err := os.MkdirAll(filepath.Dir(mathFile), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(mathFile, []byte("pub fun sub(a, b) return a - b end"), 0644); err != nil {
			t.Fatal(err)
		}

		// prepare embedded fs
		mapFs := fstest.MapFS{
			"std/math.vv": &fstest.MapFile{
				Data: []byte("pub fun add(a, b) return a + b end"),
			},
		}

		s := NewState(cfg, "main.vv")
		s.EnsureSystemLibrary("std", mapFs)

		path, err := cfg.ResolveModulePath("./", "std/math.vv")
		if err != nil {
			t.Fatal(err)
		}
		expected, err := filepath.Abs(cfg.GetPackagePath("std/math.vv"))
		if err != nil {
			t.Fatal(err)
		}
		// On windows, Abs might return C:\... so we check suffix or handle paths carefully
		// We'll just check if it contains the expected parts
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got %s", path)
		}
		if path != expected {
			t.Errorf("expected %s, got %s", expected, path)
		}
	})
}
