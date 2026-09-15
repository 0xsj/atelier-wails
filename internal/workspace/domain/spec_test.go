package domain_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const workspaceText = "01900000-0000-7000-8000-000000000001"

func workspaceID(t *testing.T) id.ID {
	t.Helper()
	value, err := id.Parse(workspaceText)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func at() time.Time { return time.UnixMilli(1_700_000_000_000).UTC() }

func later() time.Time { return time.UnixMilli(1_700_000_005_000).UTC() }

func revision(t *testing.T, n uint64) domain.Expected {
	t.Helper()
	e, err := domain.ExpectRevision(n)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func registered(t *testing.T) domain.Workspace {
	t.Helper()
	result, err := domain.Register(workspaceID(t), "Atelier", "/tmp/atelier", at())
	if err != nil {
		t.Fatal(err)
	}
	return result.Workspace
}

func archived(t *testing.T) domain.Workspace {
	t.Helper()
	result, err := domain.Archive(registered(t), revision(t, 1), later())
	if err != nil {
		t.Fatal(err)
	}
	return result.Workspace
}

func requireFailure(t *testing.T, err error, kind faults.Kind, typ string, secrets ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s", typ)
	}
	if !faults.IsKind(err, kind) || faults.DiagnosticTypeOf(err) != typ {
		t.Fatalf("classification: %v (%s)", err, faults.DiagnosticTypeOf(err))
	}
	for _, secret := range secrets {
		if secret != "" && (strings.Contains(err.Error(), secret) || strings.Contains(faults.Message(err), secret)) {
			t.Fatalf("failure echoes input: %q", err.Error())
		}
		for _, value := range faults.FieldsOf(err) {
			if secret != "" && strings.Contains(value, secret) {
				t.Fatalf("field echoes input: %q", value)
			}
		}
	}
}

func TestWD01Names(t *testing.T) {
	accepted := []string{"Atelier", "My project", "Atelier ✨", strings.Repeat("é", 120)}
	for _, name := range accepted {
		result, err := domain.Register(workspaceID(t), name, "/tmp/atelier", at())
		if err != nil || result.Workspace.Name() != name {
			t.Fatalf("accepted name %q: %v", name, err)
		}
	}
	refused := []string{"", strings.Repeat("é", 121), " Atelier", "Atelier ", "Ate\tlier", "Ate\x07lier"}
	for _, name := range refused {
		_, err := domain.Register(workspaceID(t), name, "/tmp/atelier", at())
		requireFailure(t, err, faults.Invalid, "workspace.invalid_name", strings.TrimSpace(name))
	}
}

func TestWD02Locations(t *testing.T) {
	accepted := []string{"/tmp/atelier", `C:\Users\me`, "D:/work", `\\server\share`}
	for _, location := range accepted {
		result, err := domain.Register(workspaceID(t), "Atelier", location, at())
		if err != nil || result.Workspace.Location() != location {
			t.Fatalf("accepted location %q: %v", location, err)
		}
	}
	refused := []string{"", "relative/path", "./here", "tmp", "/tmp/\x00bad"}
	for _, location := range refused {
		_, err := domain.Register(workspaceID(t), "Atelier", location, at())
		requireFailure(t, err, faults.Invalid, "workspace.invalid_location", location)
	}
}

func TestWD03Timestamps(t *testing.T) {
	zone := time.FixedZone("plus-two", 2*60*60)
	precise := time.Date(2024, 1, 2, 3, 4, 5, 678_901_234, zone)
	result, err := domain.Register(workspaceID(t), "Atelier", "/tmp/atelier", precise)
	if err != nil {
		t.Fatal(err)
	}
	want := precise.UTC().Truncate(time.Millisecond)
	if !result.Workspace.CreatedAt().Equal(want) || result.Workspace.CreatedAt().Location() != time.UTC {
		t.Fatalf("created_at: %v", result.Workspace.CreatedAt())
	}
	if result.Workspace.CreatedAt().Nanosecond()%int(time.Millisecond) != 0 {
		t.Fatal("sub-millisecond precision retained")
	}
}

func TestWD04Register(t *testing.T) {
	result, err := domain.Register(workspaceID(t), "Atelier", "/tmp/atelier", at())
	if err != nil {
		t.Fatal(err)
	}
	w := result.Workspace
	if w.ID() != workspaceID(t) || w.Status() != domain.Active || w.Revision() != 1 || !w.CreatedAt().Equal(at()) || !w.UpdatedAt().Equal(at()) || !w.Valid() {
		t.Fatalf("registered: %+v", w)
	}
	ev := result.Event
	if ev.Name() != domain.RegisteredEvent || ev.Workspace() != workspaceID(t) || ev.Revision() != 1 || ev.Title() != "Atelier" || ev.Location() != "/tmp/atelier" || ev.Status() != domain.Active {
		t.Fatalf("event: %+v", ev)
	}
	_, err = domain.Register(id.ID{}, "Atelier", "/tmp/atelier", at())
	requireFailure(t, err, faults.Invalid, "workspace.invalid_id", "Atelier", "/tmp/atelier")
}

func TestWD05Rename(t *testing.T) {
	current := registered(t)
	same, err := domain.Rename(current, "Atelier", revision(t, 1), later())
	if err != nil || same.Status != domain.Unchanged || same.Event != nil || same.Workspace != current {
		t.Fatalf("same name: %+v %v", same, err)
	}
	changed, err := domain.Rename(current, "Studio", revision(t, 1), later())
	if err != nil || changed.Status != domain.Changed || changed.Workspace.Revision() != 2 || changed.Workspace.Name() != "Studio" {
		t.Fatalf("renamed: %+v %v", changed, err)
	}
	if !changed.Workspace.UpdatedAt().Equal(later()) || !changed.Workspace.CreatedAt().Equal(at()) {
		t.Fatal("timestamps after rename")
	}
	if changed.Event == nil || changed.Event.Name() != domain.RenamedEvent || changed.Event.Title() != "Studio" || changed.Event.Revision() != 2 {
		t.Fatalf("renamed event: %+v", changed.Event)
	}
	_, err = domain.Rename(current, " bad", revision(t, 1), later())
	requireFailure(t, err, faults.Invalid, "workspace.invalid_name", "bad")
	_, err = domain.Rename(current, "Studio", revision(t, 2), later())
	requireFailure(t, err, faults.Conflict, "workspace.conflict", "Studio", "/tmp/atelier")
	_, err = domain.Rename(current, "Studio", domain.ExpectAbsent(), later())
	requireFailure(t, err, faults.Conflict, "workspace.conflict", "Studio")
	if current.Name() != "Atelier" || current.Revision() != 1 {
		t.Fatal("current mutated")
	}
}

func TestWD06Archive(t *testing.T) {
	current := registered(t)
	result, err := domain.Archive(current, revision(t, 1), later())
	if err != nil || result.Status != domain.Changed || result.Workspace.Status() != domain.Archived || result.Workspace.Revision() != 2 {
		t.Fatalf("archive: %+v %v", result, err)
	}
	if result.Event == nil || result.Event.Name() != domain.ArchivedEvent || result.Event.Status() != domain.Archived {
		t.Fatalf("archive event: %+v", result.Event)
	}
	again, err := domain.Archive(result.Workspace, revision(t, 2), later())
	if err != nil || again.Status != domain.Unchanged || again.Event != nil {
		t.Fatalf("archive again: %+v %v", again, err)
	}
	_, err = domain.Archive(current, revision(t, 9), later())
	requireFailure(t, err, faults.Conflict, "workspace.conflict")
}

func TestWD07Restore(t *testing.T) {
	current := archived(t)
	result, err := domain.Restore(current, revision(t, 2), later())
	if err != nil || result.Status != domain.Changed || result.Workspace.Status() != domain.Active || result.Workspace.Revision() != 3 {
		t.Fatalf("restore: %+v %v", result, err)
	}
	if result.Event == nil || result.Event.Name() != domain.RestoredEvent || result.Event.Status() != domain.Active {
		t.Fatalf("restore event: %+v", result.Event)
	}
	same, err := domain.Restore(result.Workspace, revision(t, 3), later())
	if err != nil || same.Status != domain.Unchanged || same.Event != nil {
		t.Fatalf("restore active: %+v %v", same, err)
	}
}

func TestWD08Forget(t *testing.T) {
	current := archived(t)
	result, err := domain.Forget(current, revision(t, 2))
	if err != nil || result.Revision != 3 || result.Event.Name() != domain.ForgottenEvent || result.Event.Status() != domain.Archived || result.Event.Revision() != 3 {
		t.Fatalf("forget: %+v %v", result, err)
	}
	_, err = domain.Forget(registered(t), revision(t, 1))
	requireFailure(t, err, faults.Conflict, "workspace.active_forget", "Atelier", "/tmp/atelier")
	_, err = domain.Forget(current, revision(t, 1))
	requireFailure(t, err, faults.Conflict, "workspace.conflict")
}

func TestWD09RevisionExhausted(t *testing.T) {
	maxActive, err := domain.Rebuild(workspaceID(t), "Atelier", "/tmp/atelier", domain.Active, math.MaxUint64, at(), at())
	if err != nil {
		t.Fatal(err)
	}
	maxArchived, _ := domain.Rebuild(workspaceID(t), "Atelier", "/tmp/atelier", domain.Archived, math.MaxUint64, at(), at())
	last := revision(t, math.MaxUint64)
	_, err = domain.Rename(maxActive, "Studio", last, later())
	requireFailure(t, err, faults.Conflict, "workspace.revision_exhausted")
	_, err = domain.Archive(maxActive, last, later())
	requireFailure(t, err, faults.Conflict, "workspace.revision_exhausted")
	_, err = domain.Restore(maxArchived, last, later())
	requireFailure(t, err, faults.Conflict, "workspace.revision_exhausted")
	_, err = domain.Forget(maxArchived, last)
	requireFailure(t, err, faults.Conflict, "workspace.revision_exhausted")
	same, err := domain.Rename(maxActive, "Atelier", last, later())
	if err != nil || same.Status != domain.Unchanged {
		t.Fatal("same-name rename at maximum revision")
	}
}

func TestWD10Rebuild(t *testing.T) {
	for _, status := range []domain.Status{domain.Active, domain.Archived} {
		w, err := domain.Rebuild(workspaceID(t), "Atelier", "/tmp/atelier", status, 5, at(), later())
		if err != nil || w.Status() != status || w.Revision() != 5 || !w.CreatedAt().Equal(at()) || !w.UpdatedAt().Equal(later()) || !w.Valid() {
			t.Fatalf("rebuild %v: %+v %v", status, w, err)
		}
	}
	cases := []struct {
		name string
		typ  string
		call func() error
	}{
		{"revision 0", "workspace.invalid_revision", func() error {
			_, err := domain.Rebuild(workspaceID(t), "Atelier", "/tmp/atelier", domain.Active, 0, at(), at())
			return err
		}},
		{"unknown status", "workspace.invalid_status", func() error {
			_, err := domain.Rebuild(workspaceID(t), "Atelier", "/tmp/atelier", domain.Status(9), 1, at(), at())
			return err
		}},
		{"invalid name", "workspace.invalid_name", func() error {
			_, err := domain.Rebuild(workspaceID(t), "", "/tmp/atelier", domain.Active, 1, at(), at())
			return err
		}},
		{"invalid location", "workspace.invalid_location", func() error {
			_, err := domain.Rebuild(workspaceID(t), "Atelier", "relative", domain.Active, 1, at(), at())
			return err
		}},
		{"zero id", "workspace.invalid_id", func() error {
			_, err := domain.Rebuild(id.ID{}, "Atelier", "/tmp/atelier", domain.Active, 1, at(), at())
			return err
		}},
	}
	for _, c := range cases {
		requireFailure(t, c.call(), faults.Invalid, c.typ, "Atelier", "/tmp/atelier")
	}
}

func TestWD11InvalidCurrent(t *testing.T) {
	var zero domain.Workspace
	_, err := domain.Rename(zero, "Studio", revision(t, 1), later())
	requireFailure(t, err, faults.Invalid, "workspace.invalid_record", "Studio")
	_, err = domain.Archive(zero, revision(t, 1), later())
	requireFailure(t, err, faults.Invalid, "workspace.invalid_record")
	_, err = domain.Restore(zero, revision(t, 1), later())
	requireFailure(t, err, faults.Invalid, "workspace.invalid_record")
	_, err = domain.Forget(zero, revision(t, 1))
	requireFailure(t, err, faults.Invalid, "workspace.invalid_record")
}
