package fileio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

func requireFailure(t *testing.T, err error, typ string, path string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s", typ)
	}
	if !faults.IsKind(err, faults.Unavailable) || faults.DiagnosticTypeOf(err) != typ {
		t.Fatalf("classification: %v (%s)", err, faults.DiagnosticTypeOf(err))
	}
	if faults.DetailsOf(err)["path"] != path {
		t.Fatalf("path detail missing: %v", faults.DetailsOf(err))
	}
	if strings.Contains(faults.Message(err), path) {
		t.Fatalf("public message echoes the path: %q", faults.Message(err))
	}
	var failure *faults.Failure
	if !errorsAs(err, &failure) || failure.Unwrap() == nil {
		t.Fatal("operating system cause missing")
	}
}

func errorsAs(err error, target **faults.Failure) bool {
	for err != nil {
		if f, ok := err.(*faults.Failure); ok {
			*target = f
			return true
		}
		wrapped, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = wrapped.Unwrap()
	}
	return false
}

func listNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestF01ReadAbsentAndDirectory(t *testing.T) {
	dir := t.TempDir()
	data, found, err := fileio.Read(filepath.Join(dir, "missing.json"))
	if err != nil || found || data != nil {
		t.Fatalf("missing: %v %v %v", data, found, err)
	}
	_, _, err = fileio.Read(dir)
	requireFailure(t, err, "fileio.read_failed", dir)
}

func TestF02ReplaceThenRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prefs.json")
	if err := fileio.Replace(path, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	data, found, err := fileio.Read(path)
	if err != nil || !found || string(data) != `{"a":1}` {
		t.Fatalf("round trip: %q %v %v", data, found, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode: %v %v", info.Mode(), err)
	}
	empty := filepath.Join(dir, "empty")
	if err := fileio.Replace(empty, nil); err != nil {
		t.Fatal(err)
	}
	data, found, err = fileio.Read(empty)
	if err != nil || !found || len(data) != 0 {
		t.Fatalf("empty round trip: %q %v %v", data, found, err)
	}
}

func TestF03ReplaceExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prefs.json")
	for _, content := range []string{"first", "second", "third"} {
		if err := fileio.Replace(path, []byte(content)); err != nil {
			t.Fatal(err)
		}
		data, _, err := fileio.Read(path)
		if err != nil || string(data) != content {
			t.Fatalf("after replace: %q %v", data, err)
		}
	}
	if names := listNames(t, dir); len(names) != 1 || names[0] != "prefs.json" {
		t.Fatalf("temporary files left behind: %v", names)
	}
}

func TestF04FailedReplaceLeavesTarget(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope", "prefs.json")
	err := fileio.Replace(missing, []byte("x"))
	requireFailure(t, err, "fileio.write_failed", missing)

	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(locked, "prefs.json")
	if err := fileio.Replace(path, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
	if os.Getuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	err = fileio.Replace(path, []byte("replacement"))
	requireFailure(t, err, "fileio.write_failed", path)
	data, _, err := fileio.Read(path)
	if err != nil || string(data) != "original" {
		t.Fatalf("target changed after failed replace: %q %v", data, err)
	}
	if names := listNames(t, locked); len(names) != 1 || names[0] != "prefs.json" {
		t.Fatalf("temporary files left behind: %v", names)
	}
}

func TestF05EnsureDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	if err := fileio.EnsureDir(nested); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(nested)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("nested: %v %v", info, err)
	}
	if err := fileio.EnsureDir(nested); err != nil {
		t.Fatalf("second call: %v", err)
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	requireFailure(t, fileio.EnsureDir(file), "fileio.dir_failed", file)
}
