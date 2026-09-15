package root

import (
	"os"
	"path/filepath"
	"testing"
)

const workspaceRegisterDocument = `{"id":"01900000-0000-7000-8000-000000000001","name":"Atelier","location":"/tmp/atelier","at":"2026-09-15T12:00:00.123Z"}`
const workspaceReadDocument = `{"id":"01900000-0000-7000-8000-000000000001"}`

func workspaceDocumentPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "atelier-wails", workspaceDocument)
}

func TestComposedWorkspaceSurvivesRestart(t *testing.T) {
	path := workspaceDocumentPath(t)
	first, err := composeWorkspace(path)
	if err != nil {
		t.Fatal(err)
	}
	if outcome := first.Register(workspaceRegisterDocument); !outcome.OK {
		t.Fatalf("register: %+v", outcome.Failure)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("workspace document missing: %v", err)
	}
	again, err := composeWorkspace(path)
	if err != nil {
		t.Fatal(err)
	}
	read := again.Read(workspaceReadDocument)
	if !read.OK || read.Value == nil || !read.Value.Found || read.Value.Workspace == nil {
		t.Fatalf("workspace did not survive restart: %+v", read)
	}
	if read.Value.Workspace.Revision != "1" || read.Value.Workspace.Name != "Atelier" {
		t.Fatalf("unexpected persisted workspace: %+v", read.Value.Workspace)
	}
}

func TestCorruptWorkspaceDocumentRefusesComposition(t *testing.T) {
	path := workspaceDocumentPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	facade, err := composeWorkspace(path)
	if facade != nil || err == nil {
		t.Fatalf("corrupt workspace document opened: facade=%v err=%v", facade, err)
	}
}
