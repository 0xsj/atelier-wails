package wailshost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkbenchSnapshotRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workbench.json")
	facade := NewWorkbench(path)

	document, err := facade.Load()
	if err != nil {
		t.Fatalf("load missing snapshot: %v", err)
	}
	if document != nil {
		t.Fatalf("missing snapshot = %q, want nil", *document)
	}

	if err := facade.Save(`{"version":1}`); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	document, err = facade.Load()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if document == nil {
		t.Fatal("loaded snapshot = nil, want a document")
	}
	if *document != `{"version":1}` {
		t.Fatalf("loaded snapshot = %q, want %q", *document, `{"version":1}`)
	}
}

func TestWorkbenchSnapshotReplacesWithoutTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "workbench.json")
	facade := NewWorkbench(path)
	if err := facade.Save("first"); err != nil {
		t.Fatalf("save first snapshot: %v", err)
	}
	if err := facade.Save("second"); err != nil {
		t.Fatalf("save second snapshot: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(data) != "second" {
		t.Fatalf("snapshot = %q, want second", data)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read snapshot directory: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Fatalf("temporary file remains: %s", entry.Name())
		}
	}
}
