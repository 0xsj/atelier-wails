package command_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/app/command"
	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/memory"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

var errStore = faults.New(faults.Unavailable, "preference store unavailable").
	WithType("preferences.store_unavailable")

// recording keeps every published event and can observe the store at
// publication time.
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

// fault fails before any effect with a classified or unclassified error.
type fault struct{ err error }

func (f fault) Replace(context.Context, domain.Scope, domain.Key, domain.Value, domain.Expected) (domain.ReplaceResult, error) {
	return domain.ReplaceResult{}, f.err
}

func (f fault) Remove(context.Context, domain.Scope, domain.Key, domain.Expected) (domain.RemoveResult, error) {
	return domain.RemoveResult{}, f.err
}

// counting records calls and honors an already-canceled context.
type counting struct {
	inner command.Store
	calls int
}

func (c *counting) Replace(ctx context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.ReplaceResult{}, err
	}
	return c.inner.Replace(ctx, scope, key, value, expected)
}

func (c *counting) Remove(ctx context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error) {
	c.calls++
	if err := ctx.Err(); err != nil {
		return domain.RemoveResult{}, err
	}
	return c.inner.Remove(ctx, scope, key, expected)
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

func revision(t *testing.T, n uint64) domain.Expected {
	t.Helper()
	e, err := domain.ExpectRevision(n)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func service(t *testing.T, store command.Store, publisher command.Publisher) *command.Service {
	t.Helper()
	s, err := command.New(store, publisher)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// requireFailure covers A12 for every refusal.
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
	store := memory.New()
	_, err := command.New(nil, command.Discard{})
	if !faults.IsKind(err, faults.Internal) || faults.DiagnosticTypeOf(err) != "preferences.missing_dependency" {
		t.Fatalf("nil store: %v", err)
	}
	_, err = command.New(store, nil)
	if !faults.IsKind(err, faults.Internal) || faults.DiagnosticTypeOf(err) != "preferences.missing_dependency" {
		t.Fatalf("nil publisher: %v", err)
	}
	if _, err := command.New(store, command.Discard{}); err != nil {
		t.Fatal(err)
	}
}

func TestA02CreatePublishesOnce(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	result, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.Changed || result.Entry.Revision() != 1 {
		t.Fatalf("create: %+v", result)
	}
	if len(published.events) != 1 || published.events[0].Name() != domain.ChangedEvent || published.events[0].Revision() != 1 {
		t.Fatalf("published: %+v", published.events)
	}
	entry, found, err := store.Read(ctx, domain.Global(), k)
	if err != nil || !found || entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatal("stored entry")
	}
}

func TestA03UnchangedPublishesNothing(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	if _, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent()); err != nil {
		t.Fatal(err)
	}
	result, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), revision(t, 1))
	if err != nil || result.Status != domain.Unchanged || result.Entry.Revision() != 1 {
		t.Fatalf("unchanged: %+v %v", result, err)
	}
	if len(published.events) != 1 {
		t.Fatalf("unchanged published: %+v", published.events)
	}
}

func TestA04ConflictLeavesStateAndPublishesNothing(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	if _, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent()); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []domain.Expected{domain.ExpectAbsent(), revision(t, 9)} {
		_, err := s.Replace(ctx, domain.Global(), k, text(t, "light"), expected)
		requireFailure(t, err, "preferences.replace", faults.Conflict, "preferences.conflict", "editor.theme", "light")
	}
	if len(published.events) != 1 {
		t.Fatal("conflict published")
	}
	entry, found, _ := store.Read(ctx, domain.Global(), k)
	if !found || entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatal("conflict mutated state")
	}
}

func TestA05InvalidInputsRefusedBeforeStore(t *testing.T) {
	ctx := context.Background()
	store := &counting{inner: memory.New()}
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	value := text(t, "dark")
	badExpected := domain.Expected{}
	cases := []struct {
		name  string
		typ   string
		call  func() error
		leaks []string
	}{
		{"zero scope", "preferences.invalid_scope", func() error {
			_, err := s.Replace(ctx, domain.Scope{}, k, value, domain.ExpectAbsent())
			return err
		}, []string{"editor.theme", "dark"}},
		{"zero key", "preferences.invalid_key", func() error {
			_, err := s.Replace(ctx, domain.Global(), domain.Key{}, value, domain.ExpectAbsent())
			return err
		}, []string{"dark"}},
		{"zero value", "preferences.invalid_value", func() error {
			_, err := s.Replace(ctx, domain.Global(), k, domain.Value{}, domain.ExpectAbsent())
			return err
		}, []string{"editor.theme"}},
		{"remove zero scope", "preferences.invalid_scope", func() error {
			_, err := s.Remove(ctx, domain.Scope{}, k, badExpected)
			return err
		}, []string{"editor.theme"}},
		{"remove zero key", "preferences.invalid_key", func() error {
			_, err := s.Remove(ctx, domain.Global(), domain.Key{}, badExpected)
			return err
		}, nil},
	}
	for _, c := range cases {
		operation := "preferences.replace"
		if strings.HasPrefix(c.name, "remove") {
			operation = "preferences.remove"
		}
		requireFailure(t, c.call(), operation, faults.Invalid, c.typ, c.leaks...)
	}
	if store.calls != 0 {
		t.Fatalf("store called %d times", store.calls)
	}
	if len(published.events) != 0 {
		t.Fatal("invalid input published")
	}
}

func TestA06RemovePresentThenAbsent(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	if _, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent()); err != nil {
		t.Fatal(err)
	}
	removed, err := s.Remove(ctx, domain.Global(), k, revision(t, 1))
	if err != nil || !removed.Removed || removed.Revision != 2 {
		t.Fatalf("remove: %+v %v", removed, err)
	}
	if len(published.events) != 2 || published.events[1].Name() != domain.RemovedEvent || published.events[1].Revision() != 2 {
		t.Fatalf("removed event: %+v", published.events)
	}
	if _, found, _ := store.Read(ctx, domain.Global(), k); found {
		t.Fatal("entry still present")
	}
	again, err := s.Remove(ctx, domain.Global(), k, domain.ExpectAbsent())
	if err != nil || again.Removed || again.Event != nil {
		t.Fatalf("absent remove: %+v %v", again, err)
	}
	if len(published.events) != 2 {
		t.Fatal("absent remove published")
	}
}

func TestA07StoreFailurePassesThrough(t *testing.T) {
	ctx := context.Background()
	published := &recording{}
	s := service(t, fault{err: errStore}, published)
	k := key(t, "editor.theme")
	_, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	requireFailure(t, err, "preferences.replace", faults.Unavailable, "preferences.store_unavailable", "editor.theme", "dark")
	if !errors.Is(err, errStore) {
		t.Fatal("condition identity lost")
	}
	_, err = s.Remove(ctx, domain.Global(), k, revision(t, 1))
	requireFailure(t, err, "preferences.remove", faults.Unavailable, "preferences.store_unavailable", "editor.theme")
	if len(published.events) != 0 {
		t.Fatal("failure published")
	}

	plain := errors.New("driver exploded")
	s = service(t, fault{err: plain}, published)
	_, err = s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	if err == nil || !errors.Is(err, plain) {
		t.Fatal("unclassified error lost")
	}
	if _, classified := faults.KindOf(err); classified {
		t.Fatal("unclassified error was classified")
	}
}

func TestA08PublisherSeesCommittedState(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	k := key(t, "editor.theme")
	var seen []uint64
	published := &recording{onPublish: func(domain.Event) {
		entry, found, err := store.Read(ctx, domain.Global(), k)
		if err != nil || !found {
			t.Error("publisher ran before commit")
			return
		}
		seen = append(seen, entry.Revision())
	}}
	s := service(t, store, published)
	if _, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1)); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("observed revisions: %v", seen)
	}
}

func TestA11CanceledContextPassesThrough(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &counting{inner: memory.New()}
	published := &recording{}
	s := service(t, store, published)
	k := key(t, "editor.theme")
	_, err := s.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	requireFailure(t, err, "preferences.replace", faults.Canceled, "")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation cause lost")
	}
	_, err = s.Remove(ctx, domain.Global(), k, revision(t, 1))
	requireFailure(t, err, "preferences.remove", faults.Canceled, "")
	if len(published.events) != 0 {
		t.Fatal("canceled call published")
	}
}
