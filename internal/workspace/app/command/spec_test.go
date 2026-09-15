package command_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/app/command"
	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/memory"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

var errStore = faults.New(faults.Unavailable, "workspace store unavailable").
	WithType("workspace.store_unavailable")

type recording struct {
	events    []domain.Event
	onPublish func(domain.Event)
}

func (r *recording) Publish(_ context.Context, event domain.Event) {
	r.events = append(r.events, event)
	if r.onPublish != nil {
		r.onPublish(event)
	}
}

type fault struct{ err error }

func (f fault) Register(context.Context, id.ID, string, string, time.Time) (domain.RegisterResult, error) {
	return domain.RegisterResult{}, f.err
}
func (f fault) Rename(context.Context, id.ID, string, domain.Expected, time.Time) (domain.MutationResult, bool, error) {
	return domain.MutationResult{}, false, f.err
}
func (f fault) Archive(context.Context, id.ID, domain.Expected, time.Time) (domain.MutationResult, bool, error) {
	return domain.MutationResult{}, false, f.err
}
func (f fault) Restore(context.Context, id.ID, domain.Expected, time.Time) (domain.MutationResult, bool, error) {
	return domain.MutationResult{}, false, f.err
}
func (f fault) Forget(context.Context, id.ID, domain.Expected) (domain.ForgetResult, bool, error) {
	return domain.ForgetResult{}, false, f.err
}

type counting struct {
	inner command.Store
	calls int
}

func (c *counting) Register(ctx context.Context, workspaceID id.ID, name, location string, at time.Time) (domain.RegisterResult, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.RegisterResult{}, err
	}
	return c.inner.Register(ctx, workspaceID, name, location, at)
}
func (c *counting) Rename(ctx context.Context, workspaceID id.ID, name string, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, err
	}
	return c.inner.Rename(ctx, workspaceID, name, expected, at)
}
func (c *counting) Archive(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, err
	}
	return c.inner.Archive(ctx, workspaceID, expected, at)
}
func (c *counting) Restore(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, err
	}
	return c.inner.Restore(ctx, workspaceID, expected, at)
}
func (c *counting) Forget(ctx context.Context, workspaceID id.ID, expected domain.Expected) (domain.ForgetResult, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.ForgetResult{}, false, err
	}
	return c.inner.Forget(ctx, workspaceID, expected)
}

func workspace(t *testing.T, value string) id.ID {
	t.Helper()
	parsed, err := id.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func at() time.Time    { return time.UnixMilli(100).UTC() }
func later() time.Time { return time.UnixMilli(200).UTC() }

func revision(t *testing.T, value uint64) domain.Expected {
	t.Helper()
	expected, err := domain.ExpectRevision(value)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}

func service(t *testing.T, store command.Store, publisher command.Publisher) *command.Service {
	t.Helper()
	service, err := command.New(store, publisher)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func requireFailure(t *testing.T, err error, operation string, kind faults.Kind, typ string) {
	t.Helper()
	if err == nil || !strings.HasPrefix(err.Error(), operation+": ") {
		t.Fatalf("operation missing: %v", err)
	}
	if !faults.IsKind(err, kind) || (typ != "" && faults.DiagnosticTypeOf(err) != typ) {
		t.Fatalf("classification: %v (%s)", err, faults.DiagnosticTypeOf(err))
	}
}

func TestWA01Construction(t *testing.T) {
	if _, err := command.New(nil, command.Discard{}); !faults.IsKind(err, faults.Internal) {
		t.Fatalf("nil store: %v", err)
	}
	if _, err := command.New(memory.New(), nil); !faults.IsKind(err, faults.Internal) {
		t.Fatalf("nil publisher: %v", err)
	}
	if _, err := command.New(memory.New(), command.Discard{}); err != nil {
		t.Fatal(err)
	}
}

func TestWA02RegisterPublishesOnce(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	service := service(t, store, published)
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	result, err := service.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at())
	if err != nil || result.Workspace.Revision() != 1 {
		t.Fatalf("register: %+v %v", result, err)
	}
	if len(published.events) != 1 || published.events[0].Name() != domain.RegisteredEvent {
		t.Fatalf("events: %+v", published.events)
	}
	read, found, err := store.Read(ctx, workspaceID)
	if err != nil || !found || read != result.Workspace {
		t.Fatalf("stored workspace: %+v %v %v", read, found, err)
	}
}

func TestWA03ChangedAndUnchangedPublication(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	service := service(t, store, published)
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	if _, err := service.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at()); err != nil {
		t.Fatal(err)
	}
	unchanged, err := service.Rename(ctx, workspaceID, "Atelier", revision(t, 1), later())
	if err != nil || unchanged.Status != domain.Unchanged || len(published.events) != 1 {
		t.Fatalf("unchanged rename: %+v %v", unchanged, err)
	}
	if _, err := service.Rename(ctx, workspaceID, "Atelier North", revision(t, 1), later()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Archive(ctx, workspaceID, revision(t, 2), later()); err != nil {
		t.Fatal(err)
	}
	unchanged, err = service.Archive(ctx, workspaceID, revision(t, 3), later())
	if err != nil || unchanged.Status != domain.Unchanged || len(published.events) != 3 {
		t.Fatalf("unchanged archive: %+v %v", unchanged, err)
	}
	if _, err := service.Restore(ctx, workspaceID, revision(t, 3), later()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Archive(ctx, workspaceID, revision(t, 4), later()); err != nil {
		t.Fatal(err)
	}
	if _, found, err := service.Forget(ctx, workspaceID, revision(t, 5)); err != nil || !found {
		t.Fatalf("forget: %v %v", found, err)
	}
	if len(published.events) != 6 {
		t.Fatalf("events: %+v", published.events)
	}
}

func TestWA05AbsencePolicy(t *testing.T) {
	ctx := context.Background()
	published := &recording{}
	service := service(t, memory.New(), published)
	missing := workspace(t, "01900000-0000-7000-8000-000000000002")
	_, err := service.Rename(ctx, missing, "Missing", revision(t, 1), at())
	requireFailure(t, err, "workspace.rename", faults.NotFound, "workspace.not_found")
	_, err = service.Archive(ctx, missing, revision(t, 1), at())
	requireFailure(t, err, "workspace.archive", faults.NotFound, "workspace.not_found")
	_, err = service.Restore(ctx, missing, revision(t, 1), at())
	requireFailure(t, err, "workspace.restore", faults.NotFound, "workspace.not_found")
	if _, found, err := service.Forget(ctx, missing, revision(t, 1)); err != nil || found {
		t.Fatalf("absent forget: %v %v", found, err)
	}
	if len(published.events) != 0 {
		t.Fatal("absence published")
	}
}

func TestWA06RefusalsPreserveState(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	service := service(t, store, published)
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	if _, err := service.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at()); err != nil {
		t.Fatal(err)
	}
	_, _, err := service.Forget(ctx, workspaceID, revision(t, 1))
	requireFailure(t, err, "workspace.forget", faults.Conflict, "workspace.active_forget")
	_, err = service.Rename(ctx, workspaceID, "North", revision(t, 9), later())
	requireFailure(t, err, "workspace.rename", faults.Conflict, "workspace.conflict")
	read, found, readErr := store.Read(ctx, workspaceID)
	if readErr != nil || !found || read.Revision() != 1 || len(published.events) != 1 {
		t.Fatalf("refusal mutated state: %+v %v %v", read, found, readErr)
	}
}

func TestWA07FailuresPreserveClassification(t *testing.T) {
	ctx := context.Background()
	service := service(t, fault{err: errStore}, command.Discard{})
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	_, err := service.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at())
	requireFailure(t, err, "workspace.register", faults.Unavailable, "workspace.store_unavailable")
	if !errors.Is(err, errStore) {
		t.Fatal("failure identity lost")
	}
	_, err = service.Rename(ctx, workspaceID, "North", revision(t, 1), at())
	requireFailure(t, err, "workspace.rename", faults.Unavailable, "workspace.store_unavailable")
}

func TestWA08PublisherSeesCommittedState(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	var seen []uint64
	published := &recording{onPublish: func(domain.Event) {
		read, found, err := store.Read(ctx, workspaceID)
		if err != nil || !found {
			t.Errorf("published before commit: %v %v", found, err)
			return
		}
		seen = append(seen, read.Revision())
	}}
	service := service(t, store, published)
	if _, err := service.Register(ctx, workspaceID, "Atelier", "/tmp/atelier", at()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Rename(ctx, workspaceID, "North", revision(t, 1), later()); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("observed revisions: %v", seen)
	}
}

func TestWA09CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &counting{inner: memory.New()}
	published := &recording{}
	service := service(t, store, published)
	_, err := service.Register(ctx, workspace(t, "01900000-0000-7000-8000-000000000001"), "Atelier", "/tmp/atelier", at())
	requireFailure(t, err, "workspace.register", faults.Canceled, "")
	if !errors.Is(err, context.Canceled) || len(published.events) != 0 || store.calls != 1 {
		t.Fatalf("cancellation: %v calls=%d events=%d", err, store.calls, len(published.events))
	}
}
