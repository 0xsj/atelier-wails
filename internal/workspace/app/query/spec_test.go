package query_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/workspace/app/query"
	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/memory"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

type fault struct{ err error }

func (f fault) Read(context.Context, id.ID) (domain.Workspace, bool, error) {
	return domain.Workspace{}, false, f.err
}
func (f fault) List(context.Context, domain.ListFilter) ([]domain.Workspace, error) {
	return nil, f.err
}

type counting struct {
	inner query.Store
	calls int
}

func (c *counting) Read(ctx context.Context, workspaceID id.ID) (domain.Workspace, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.Workspace{}, false, err
	}
	return c.inner.Read(ctx, workspaceID)
}
func (c *counting) List(ctx context.Context, filter domain.ListFilter) ([]domain.Workspace, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.inner.List(ctx, filter)
}

func workspace(t *testing.T, value string) id.ID {
	t.Helper()
	parsed, err := id.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
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

func service(t *testing.T, store query.Store) *query.Service {
	t.Helper()
	service, err := query.New(store)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestWA01Construction(t *testing.T) {
	if _, err := query.New(nil); !faults.IsKind(err, faults.Internal) {
		t.Fatalf("nil store: %v", err)
	}
	if _, err := query.New(memory.New()); err != nil {
		t.Fatal(err)
	}
}

func TestWA04ReadListAndAbsence(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	service := service(t, store)
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	missing, found, err := service.Read(ctx, workspaceID)
	if err != nil || found || missing.Valid() {
		t.Fatalf("absent read: %+v %v %v", missing, found, err)
	}
	items, err := service.List(ctx, domain.All)
	if err != nil || len(items) != 0 {
		t.Fatalf("empty list: %v %v", items, err)
	}
}

func TestWA07StoreFailure(t *testing.T) {
	errStore := faults.New(faults.Unavailable, "workspace store unavailable").WithType("workspace.store_unavailable")
	service := service(t, fault{err: errStore})
	workspaceID := workspace(t, "01900000-0000-7000-8000-000000000001")
	_, found, err := service.Read(context.Background(), workspaceID)
	requireFailure(t, err, "workspace.read", faults.Unavailable, "workspace.store_unavailable")
	if found || !errors.Is(err, errStore) {
		t.Fatal("read failure reported as absence or lost identity")
	}
	items, err := service.List(context.Background(), domain.All)
	requireFailure(t, err, "workspace.list", faults.Unavailable, "workspace.store_unavailable")
	if items != nil {
		t.Fatal("list failure returned items")
	}
}

func TestWA09CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &counting{inner: memory.New()}
	service := service(t, store)
	_, _, err := service.Read(ctx, workspace(t, "01900000-0000-7000-8000-000000000001"))
	requireFailure(t, err, "workspace.read", faults.Canceled, "")
	_, err = service.List(ctx, domain.All)
	requireFailure(t, err, "workspace.list", faults.Canceled, "")
	if !errors.Is(err, context.Canceled) || store.calls != 2 {
		t.Fatalf("cancellation: %v calls=%d", err, store.calls)
	}
}

func TestWA10InvalidInputsRefusedBeforeStore(t *testing.T) {
	ctx := context.Background()
	store := &counting{inner: memory.New()}
	service := service(t, store)
	_, _, err := service.Read(ctx, id.ID{})
	requireFailure(t, err, "workspace.read", faults.Invalid, "workspace.invalid_id")
	_, err = service.List(ctx, domain.ListFilter(0))
	requireFailure(t, err, "workspace.list", faults.Invalid, "workspace.invalid_filter")
	if store.calls != 0 {
		t.Fatalf("store called %d times", store.calls)
	}
}
