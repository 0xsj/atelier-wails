package persistence_test

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/persistence"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/storetest"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const (
	firstWorkspace  = "01900000-0000-7000-8000-000000000001"
	secondWorkspace = "01900000-0000-7000-8000-000000000002"
)

func freshPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "preferences.json")
}

func open(t *testing.T, path string) *persistence.Store {
	t.Helper()
	store, err := persistence.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func key(t *testing.T, text string) domain.Key {
	t.Helper()
	k, err := domain.NewKey(text)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func text(t *testing.T, value string) domain.Value {
	t.Helper()
	v, err := domain.Text(value)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func workspace(t *testing.T, value string) domain.Scope {
	t.Helper()
	parsed, err := id.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := domain.ForWorkspace(parsed)
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func revision(t *testing.T, n uint64) domain.Expected {
	t.Helper()
	e, err := domain.ExpectRevision(n)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func create(t *testing.T, store *persistence.Store, scope domain.Scope, k domain.Key, value domain.Value) {
	t.Helper()
	if _, err := store.Replace(context.Background(), scope, k, value, domain.ExpectAbsent()); err != nil {
		t.Fatalf("create %s: %v", k, err)
	}
}

func requireFailure(t *testing.T, err error, typ string, path string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s", typ)
	}
	if !faults.IsKind(err, faults.Unavailable) || faults.DiagnosticTypeOf(err) != typ {
		t.Fatalf("classification: %v (%s)", err, faults.DiagnosticTypeOf(err))
	}
	if strings.Contains(faults.Message(err), path) || strings.Contains(faults.Message(err), "editor") {
		t.Fatalf("public message leaks: %q", faults.Message(err))
	}
	if faults.DetailsOf(err)["path"] != path {
		t.Fatalf("path detail: %v", faults.DetailsOf(err))
	}
}

// S01: the shared suite over fresh temporary paths.
func TestS01Contract(t *testing.T) {
	storetest.Run(t, func(t *testing.T) storetest.Store { return open(t, freshPath(t)) })
}

func TestS02Restart(t *testing.T) {
	ctx := context.Background()
	path := freshPath(t)
	store := open(t, path)
	global, first, second := domain.Global(), workspace(t, firstWorkspace), workspace(t, secondWorkspace)
	values := map[string]domain.Value{
		"text":  text(t, "dark"),
		"empty": text(t, ""),
		"flag":  domain.Bool(true),
		"min":   domain.Int(math.MinInt64),
		"max":   domain.Int(math.MaxInt64),
	}
	for _, scope := range []domain.Scope{global, first, second} {
		for name, value := range values {
			create(t, store, scope, key(t, name), value)
		}
	}
	if _, err := store.Replace(ctx, first, key(t, "text"), text(t, "light"), revision(t, 1)); err != nil {
		t.Fatal(err)
	}
	before := map[domain.Scope][]domain.Entry{}
	for _, scope := range []domain.Scope{global, first, second} {
		entries, err := store.List(ctx, scope)
		if err != nil || len(entries) != len(values) {
			t.Fatalf("list before restart: %v %v", entries, err)
		}
		before[scope] = entries
	}

	reopened := open(t, path)
	for scope, expected := range before {
		entries, err := reopened.List(ctx, scope)
		if err != nil || len(entries) != len(expected) {
			t.Fatalf("list after restart: %v %v", entries, err)
		}
		for i := range expected {
			if entries[i] != expected[i] {
				t.Fatalf("entry differs after restart: %+v vs %+v", entries[i], expected[i])
			}
		}
	}
	entry, found, _ := reopened.Read(ctx, first, key(t, "text"))
	if !found || entry.Revision() != 2 || !entry.Value().Equal(text(t, "light")) {
		t.Fatalf("replaced entry after restart: %+v", entry)
	}
	if _, err := reopened.Replace(ctx, first, key(t, "text"), text(t, "dim"), revision(t, 2)); err != nil {
		t.Fatalf("replace at preserved revision: %v", err)
	}
}

func TestS03AbsentFile(t *testing.T) {
	ctx := context.Background()
	path := freshPath(t)
	store := open(t, path)
	k := key(t, "editor.theme")
	if _, found, err := store.Read(ctx, domain.Global(), k); err != nil || found {
		t.Fatal("fresh store not empty")
	}
	if entries, err := store.List(ctx, domain.Global()); err != nil || len(entries) != 0 {
		t.Fatal("fresh store lists entries")
	}
	if _, err := store.Replace(ctx, domain.Global(), k, text(t, "dark"), revision(t, 3)); err == nil {
		t.Fatal("conflicting write accepted")
	}
	if _, err := store.Remove(ctx, domain.Global(), k, domain.ExpectAbsent()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file exists before the first change: %v", err)
	}
	create(t, store, domain.Global(), k, text(t, "dark"))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing after the first change: %v", err)
	}
}

func TestS04CorruptDocuments(t *testing.T) {
	valid := func(entry string) string {
		return `{"format":1,"entries":[` + entry + `]}`
	}
	global := `{"scope":{"kind":"global"},"key":"%s","value":{"kind":"%s","%s":"%s"},"revision":"%s"}`
	_ = global
	cases := []struct {
		name, content, typ string
	}{
		{"not json", `{`, "preferences.storage_corrupt"},
		{"format 2", `{"format":2,"entries":[]}`, "preferences.storage_unsupported"},
		{"unknown key", `{"format":1,"entries":[],"extra":true}`, "preferences.storage_corrupt"},
		{"bad key", valid(`{"scope":{"kind":"global"},"key":"Bad","value":{"kind":"text","text":"x"},"revision":"1"}`), "preferences.storage_corrupt"},
		{"zero revision", valid(`{"scope":{"kind":"global"},"key":"a","value":{"kind":"text","text":"x"},"revision":"0"}`), "preferences.storage_corrupt"},
		{"duplicate", valid(`{"scope":{"kind":"global"},"key":"a","value":{"kind":"text","text":"x"},"revision":"1"},{"scope":{"kind":"global"},"key":"a","value":{"kind":"text","text":"y"},"revision":"1"}`), "preferences.storage_corrupt"},
		{"bad int", valid(`{"scope":{"kind":"global"},"key":"a","value":{"kind":"int","int":"x"},"revision":"1"}`), "preferences.storage_corrupt"},
		{"trailing", `{"format":1,"entries":[]}{}`, "preferences.storage_corrupt"},
	}
	for _, c := range cases {
		path := freshPath(t)
		if err := os.WriteFile(path, []byte(c.content), 0o600); err != nil {
			t.Fatal(err)
		}
		store, err := persistence.Open(path)
		if store != nil {
			t.Fatalf("%s: store returned for a corrupt document", c.name)
		}
		requireFailure(t, err, c.typ, path)
		after, _ := os.ReadFile(path)
		if string(after) != c.content {
			t.Fatalf("%s: file changed by open", c.name)
		}
	}
}

func TestS05WriteFailures(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "preferences.json")
	store := open(t, path)
	k := key(t, "editor.theme")
	create(t, store, domain.Global(), k, text(t, "dark"))
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	_, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1))
	requireFailure(t, err, "preferences.storage_unavailable", path)
	entry, _, _ := store.Read(ctx, domain.Global(), k)
	if entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatalf("memory changed after failed write: %+v", entry)
	}
	reopened := open(t, path)
	entry, _, _ = reopened.Read(ctx, domain.Global(), k)
	if entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatalf("file changed after failed write: %+v", entry)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1)); err != nil {
		t.Fatalf("write after permission restored: %v", err)
	}

	missing := filepath.Join(t.TempDir(), "nope", "preferences.json")
	orphan := open(t, missing)
	_, err = orphan.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	requireFailure(t, err, "preferences.storage_unavailable", missing)
}

func TestS06DeterministicBytes(t *testing.T) {
	ctx := context.Background()
	first, second := freshPath(t), freshPath(t)
	a, b := open(t, first), open(t, second)
	ws := workspace(t, firstWorkspace)
	create(t, a, domain.Global(), key(t, "z"), text(t, "z"))
	create(t, a, ws, key(t, "m"), domain.Int(1))
	create(t, a, domain.Global(), key(t, "a"), domain.Bool(true))
	create(t, b, domain.Global(), key(t, "a"), domain.Bool(true))
	create(t, b, ws, key(t, "m"), domain.Int(1))
	create(t, b, domain.Global(), key(t, "z"), text(t, "z"))
	left, _ := os.ReadFile(first)
	right, _ := os.ReadFile(second)
	if string(left) != string(right) {
		t.Fatalf("files differ:\n%s\n%s", left, right)
	}
	want := `{"format":1,"entries":[{"scope":{"kind":"global"},"key":"a","value":{"kind":"bool","bool":true},"revision":"1"},{"scope":{"kind":"global"},"key":"z","value":{"kind":"text","text":"z"},"revision":"1"},{"scope":{"kind":"workspace","workspace_id":"` + firstWorkspace + `"},"key":"m","value":{"kind":"int","int":"1"},"revision":"1"}]}` + "\n"
	if string(left) != want {
		t.Fatalf("document bytes:\n got %s\nwant %s", left, want)
	}
	entries, _ := open(t, first).List(ctx, domain.Global())
	if len(entries) != 2 {
		t.Fatal("reopen after deterministic write")
	}
}
