package query_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/app/query"
	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/memory"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

var errStore = faults.New(faults.Unavailable, "preference store unavailable").
	WithType("preferences.store_unavailable")

type fault struct{ err error }

func (f fault) Read(context.Context, domain.Scope, domain.Key) (domain.Entry, bool, error) {
	return domain.Entry{}, false, f.err
}

func (f fault) List(context.Context, domain.Scope) ([]domain.Entry, error) {
	return nil, f.err
}

type counting struct {
	inner query.Store
	calls int
}

func (c *counting) Read(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.Entry{}, false, err
	}
	return c.inner.Read(ctx, scope, key)
}

func (c *counting) List(ctx context.Context, scope domain.Scope) ([]domain.Entry, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.inner.List(ctx, scope)
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

// seeded fills a memory store through the store's own conditional write.
func seeded(t *testing.T, entries map[string]domain.Value) *memory.Store {
	t.Helper()
	store := memory.New()
	for name, value := range entries {
		if _, err := store.Replace(context.Background(), domain.Global(), key(t, name), value, domain.ExpectAbsent()); err != nil {
			t.Fatal(err)
		}
	}
	return store
}

func service(t *testing.T, store query.Store) *query.Service {
	t.Helper()
	s, err := query.New(store)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func requireFailure(t *testing.T, err error, operation string, kind faults.Kind, typ string, secrets ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s failure", typ)
	}
	if !strings.HasPrefix(err.Error(), operation+": ") {
		t.Fatalf("operation missing: %q", err.Error())
	}
	if !faults.IsKind(err, kind) || faults.DiagnosticTypeOf(err) != typ {
		t.Fatalf("classification: %v (%s)", err, faults.DiagnosticTypeOf(err))
	}
	for _, secret := range secrets {
		if strings.Contains(faults.Message(err), secret) || strings.Contains(err.Error(), secret) {
			t.Fatalf("failure echoes input: %q", err.Error())
		}
	}
}

func TestA01Construction(t *testing.T) {
	_, err := query.New(nil)
	if !faults.IsKind(err, faults.Internal) || faults.DiagnosticTypeOf(err) != "preferences.missing_dependency" {
		t.Fatalf("nil store: %v", err)
	}
	if _, err := query.New(memory.New()); err != nil {
		t.Fatal(err)
	}
}

func TestA05InvalidInputsRefusedBeforeStore(t *testing.T) {
	ctx := context.Background()
	store := &counting{inner: memory.New()}
	s := service(t, store)
	k := key(t, "editor.theme")
	_, _, err := s.Read(ctx, domain.Scope{}, k)
	requireFailure(t, err, "preferences.read", faults.Invalid, "preferences.invalid_scope", "editor.theme")
	_, _, err = s.Read(ctx, domain.Global(), domain.Key{})
	requireFailure(t, err, "preferences.read", faults.Invalid, "preferences.invalid_key")
	_, err = s.List(ctx, domain.Scope{})
	requireFailure(t, err, "preferences.list", faults.Invalid, "preferences.invalid_scope")
	_, err = s.Resolve(ctx, domain.Global(), domain.Key{}, text(t, "dark"))
	requireFailure(t, err, "preferences.resolve", faults.Invalid, "preferences.invalid_key", "dark")
	_, err = s.Resolve(ctx, domain.Global(), k, domain.Value{})
	requireFailure(t, err, "preferences.resolve", faults.Invalid, "preferences.invalid_value", "editor.theme")
	if store.calls != 0 {
		t.Fatalf("store called %d times", store.calls)
	}
}

func TestA07StoreFailurePassesThrough(t *testing.T) {
	ctx := context.Background()
	s := service(t, fault{err: errStore})
	k := key(t, "editor.theme")
	_, found, err := s.Read(ctx, domain.Global(), k)
	requireFailure(t, err, "preferences.read", faults.Unavailable, "preferences.store_unavailable", "editor.theme")
	if found || !errors.Is(err, errStore) {
		t.Fatal("read failure reported as absence or lost identity")
	}
	entries, err := s.List(ctx, domain.Global())
	requireFailure(t, err, "preferences.list", faults.Unavailable, "preferences.store_unavailable")
	if entries != nil {
		t.Fatal("list failure returned entries")
	}
	resolved, err := s.Resolve(ctx, domain.Global(), k, text(t, "dark"))
	requireFailure(t, err, "preferences.resolve", faults.Unavailable, "preferences.store_unavailable", "editor.theme", "dark")
	if resolved.Stored || resolved.Value.Valid() {
		t.Fatal("resolve failure reported a value")
	}

	plain := errors.New("driver exploded")
	s = service(t, fault{err: plain})
	_, _, err = s.Read(ctx, domain.Global(), k)
	if err == nil || !errors.Is(err, plain) {
		t.Fatal("unclassified error lost")
	}
	if _, classified := faults.KindOf(err); classified {
		t.Fatal("unclassified error was classified")
	}
}

func TestA09ReadAndList(t *testing.T) {
	ctx := context.Background()
	store := seeded(t, map[string]domain.Value{"z.last": text(t, "z"), "a.first": text(t, "a")})
	s := service(t, store)
	entry, found, err := s.Read(ctx, domain.Global(), key(t, "z.last"))
	if err != nil || !found || !entry.Value().Equal(text(t, "z")) {
		t.Fatal("present read")
	}
	entry, found, err = s.Read(ctx, domain.Global(), key(t, "missing"))
	if err != nil || found || entry.Valid() {
		t.Fatal("absent read")
	}
	entries, err := s.List(ctx, domain.Global())
	if err != nil || len(entries) != 2 || entries[0].Key().String() != "a.first" || entries[1].Key().String() != "z.last" {
		t.Fatalf("list: %v %v", entries, err)
	}
	other, _ := domain.ForWorkspace(workspaceID(t))
	entries, err = s.List(ctx, other)
	if err != nil || len(entries) != 0 {
		t.Fatalf("empty list: %v %v", entries, err)
	}
}

func TestA10ResolveNeverWrites(t *testing.T) {
	ctx := context.Background()
	store := seeded(t, map[string]domain.Value{"editor.theme": text(t, "dark")})
	s := service(t, store)
	resolved, err := s.Resolve(ctx, domain.Global(), key(t, "editor.theme"), text(t, "light"))
	if err != nil || !resolved.Stored || resolved.Revision != 1 || !resolved.Value.Equal(text(t, "dark")) {
		t.Fatalf("stored resolve: %+v %v", resolved, err)
	}
	missing := key(t, "editor.wrap")
	resolved, err = s.Resolve(ctx, domain.Global(), missing, domain.Bool(true))
	if err != nil || resolved.Stored || resolved.Revision != 0 || !resolved.Value.Equal(domain.Bool(true)) {
		t.Fatalf("fallback resolve: %+v %v", resolved, err)
	}
	if _, found, _ := store.Read(ctx, domain.Global(), missing); found {
		t.Fatal("fallback was written")
	}
	entries, _ := store.List(ctx, domain.Global())
	if len(entries) != 1 {
		t.Fatal("scope changed by resolve")
	}
}

func TestA11CanceledContextPassesThrough(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := service(t, &counting{inner: memory.New()})
	k := key(t, "editor.theme")
	_, _, err := s.Read(ctx, domain.Global(), k)
	requireFailure(t, err, "preferences.read", faults.Canceled, "")
	_, err = s.List(ctx, domain.Global())
	requireFailure(t, err, "preferences.list", faults.Canceled, "")
	_, err = s.Resolve(ctx, domain.Global(), k, text(t, "dark"))
	requireFailure(t, err, "preferences.resolve", faults.Canceled, "")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation cause lost")
	}
}

func workspaceID(t *testing.T) id.ID {
	t.Helper()
	value, err := id.Parse("01900000-0000-7000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	return value
}
