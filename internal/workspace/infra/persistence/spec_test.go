package persistence_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/persistence"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/storetest"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const (
	firstWorkspace  = "01900000-0000-7000-8000-000000000001"
	secondWorkspace = "01900000-0000-7000-8000-000000000002"
)

func pathFor(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

func open(t *testing.T, path string) *persistence.Store {
	t.Helper()
	store, err := persistence.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func parse(t *testing.T, value string) id.ID {
	t.Helper()
	parsed, err := id.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func revision(t *testing.T, value uint64) domain.Expected {
	t.Helper()
	expected, err := domain.ExpectRevision(value)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}

func at(milliseconds int64) time.Time { return time.UnixMilli(milliseconds).UTC() }

func TestWP01PersistentAdapterRerunsWorkspaceContract(t *testing.T) {
	counter := 0
	storetest.Run(t, func(t *testing.T) storetest.Store {
		path := filepath.Join(t.TempDir(), "workspace.json")
		counter++
		return open(t, path)
	})
	if counter != 11 {
		t.Fatalf("expected eleven store openings across ten scenarios, got %d", counter)
	}
}

func TestWP02StateSurvivesReopen(t *testing.T) {
	ctx := context.Background()
	path := pathFor(t, "workspace.json")
	store := open(t, path)
	workspaceID := parse(t, firstWorkspace)
	if _, err := store.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Rename(ctx, workspaceID, "North", revision(t, 1), at(200)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Archive(ctx, workspaceID, revision(t, 2), at(300)); err != nil {
		t.Fatal(err)
	}
	expected, found, err := store.Read(ctx, workspaceID)
	if err != nil || !found || expected.Status() != domain.Archived {
		t.Fatalf("before restart: %+v %v %v", expected, found, err)
	}

	reopened := open(t, path)
	actual, found, err := reopened.Read(ctx, workspaceID)
	if err != nil || !found || actual != expected {
		t.Fatalf("after restart: %+v %v %v", actual, found, err)
	}
	if _, _, err := reopened.Restore(ctx, workspaceID, revision(t, 3), at(400)); err != nil {
		t.Fatal(err)
	}
	active, found, err := reopened.Read(ctx, workspaceID)
	if err != nil || !found || active.Revision() != 4 || active.Status() != domain.Active {
		t.Fatalf("post-restart mutation: %+v %v %v", active, found, err)
	}
}

func TestWP03AbsentFileIsEmptyUntilFirstCommit(t *testing.T) {
	ctx := context.Background()
	path := pathFor(t, "workspace.json")
	store := open(t, path)
	workspaceID := parse(t, firstWorkspace)
	if _, found, err := store.Read(ctx, workspaceID); err != nil || found {
		t.Fatalf("fresh read: %v %v", found, err)
	}
	if values, err := store.List(ctx, domain.All); err != nil || len(values) != 0 {
		t.Fatalf("fresh list: %v %v", values, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file exists before first commit: %v", err)
	}
	if _, err := store.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing after first commit: %v", err)
	}
}

func TestWP04CorruptDocumentsRefuseOpenWithoutRewrite(t *testing.T) {
	valid := func(records string) string {
		return `{"format":1,"workspaces":[` + records + `]}`
	}
	cases := []string{
		`{`,
		`{"format":2,"workspaces":[]}`,
		`{"format":1,"workspaces":[],"extra":true}`,
		valid(`{"id":"01900000-0000-7000-8000-000000000001","name":"Atelier","location":"/tmp/atelier","status":"active","revision":"0","created_at":"100","updated_at":"100"}`),
		valid(`{"id":"01900000-0000-7000-8000-000000000001","name":"Atelier","location":"/tmp/shared","status":"active","revision":"1","created_at":"100","updated_at":"100"},{"id":"01900000-0000-7000-8000-000000000002","name":"Two","location":"/tmp/shared","status":"active","revision":"1","created_at":"100","updated_at":"100"}`),
		`{"format":1,"workspaces":[]}{} `,
	}
	for _, content := range cases {
		path := pathFor(t, "workspace.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		store, err := persistence.Open(path)
		if store != nil || err == nil {
			t.Fatalf("document opened: %q", content)
		}
		if !faults.IsKind(err, faults.Unavailable) {
			t.Fatalf("classification for %q: %v", content, err)
		}
		if got, readErr := os.ReadFile(path); readErr != nil || string(got) != content {
			t.Fatalf("document changed for %q", content)
		}
	}
}

func TestWP05FailedWriteDoesNotChangeState(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	ctx := context.Background()
	path := pathFor(t, "workspace.json")
	store := open(t, path)
	workspaceID := parse(t, firstWorkspace)
	if _, err := store.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at(100)); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(path)
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
	_, _, err = store.Rename(ctx, workspaceID, "North", revision(t, 1), at(200))
	if err == nil || faults.DiagnosticTypeOf(err) != "workspace.storage_unavailable" {
		t.Fatalf("write failure: %v", err)
	}
	current, found, readErr := store.Read(ctx, workspaceID)
	if readErr != nil || !found || current.Revision() != 1 || current.Name() != "Atelier" {
		t.Fatalf("memory changed after failed write: %+v %v %v", current, found, readErr)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("file changed after failed write")
	}
}

func TestWP06EquivalentStateHasDeterministicBytes(t *testing.T) {
	firstPath := pathFor(t, "first.json")
	secondPath := pathFor(t, "second.json")
	first, second := open(t, firstPath), open(t, secondPath)
	firstID, secondID := parse(t, firstWorkspace), parse(t, secondWorkspace)
	if _, err := first.Register(context.Background(), secondID, "Two", "/tmp/two", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Register(context.Background(), firstID, "One", "/tmp/one", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Register(context.Background(), firstID, "One", "/tmp/one", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Register(context.Background(), secondID, "Two", "/tmp/two", at(100)); err != nil {
		t.Fatal(err)
	}
	left, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	right, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) {
		t.Fatalf("deterministic documents differ:\n%s\n%s", left, right)
	}
}
