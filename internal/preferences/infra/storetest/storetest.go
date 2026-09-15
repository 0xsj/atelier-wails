// Package storetest is the reusable logical scenario suite for preference
// store adapters (memory CONTRACT.md, M01–M12) plus deterministic fault
// wrappers. It is imported only by tests; it is not a production dependency.
package storetest

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/app/command"
	"github.com/0xsj/atelier-wails/internal/preferences/app/query"
	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Store is the union of the two application ports one adapter satisfies.
type Store interface {
	command.Store
	query.Store
}

// Factory opens a fresh, empty, independent store for one subtest.
type Factory func(t *testing.T) Store

// Unavailable is the failure FailBefore reports.
var Unavailable = faults.New(faults.Unavailable, "preference store unavailable").
	WithType("preferences.store_unavailable")

// Unacknowledged is the failure LoseAcknowledgement reports after a write.
var Unacknowledged = faults.New(faults.Timeout, "preference store did not acknowledge").
	WithType("preferences.store_unacknowledged")

// Run executes every contract scenario against fresh stores from open.
func Run(t *testing.T, open Factory) {
	t.Helper()
	scenarios := []struct {
		name string
		run  func(*testing.T, Factory)
	}{
		{"M01FreshIsEmpty", m01},
		{"M02CreateThenRead", m02},
		{"M03DuplicateCreateAndEqualReplace", m03},
		{"M04RevisionConflicts", m04},
		{"M05Replacement", m05},
		{"M06RemoveThenRecreate", m06},
		{"M07DeterministicOrdering", m07},
		{"M08ScopeIsolation", m08},
		{"M09IndependentInstances", m09},
		{"M10OutputIsolation", m10},
		{"M11ConcurrentConditionalWrites", m11},
		{"M12FaultWrappers", m12},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) { scenario.run(t, open) })
	}
}

// FailBefore refuses every operation with failure before touching inner.
func FailBefore(inner Store, failure error) Store {
	return failBefore{inner: inner, failure: failure}
}

// LoseAcknowledgement performs writes on inner and reports failure instead of
// the result whenever the write changed state. Reads pass through so a caller
// can observe the committed state.
func LoseAcknowledgement(inner Store, failure error) Store {
	return loseAck{inner: inner, failure: failure}
}

type failBefore struct {
	inner   Store
	failure error
}

func (f failBefore) Read(context.Context, domain.Scope, domain.Key) (domain.Entry, bool, error) {
	return domain.Entry{}, false, f.failure
}

func (f failBefore) List(context.Context, domain.Scope) ([]domain.Entry, error) {
	return nil, f.failure
}

func (f failBefore) Replace(context.Context, domain.Scope, domain.Key, domain.Value, domain.Expected) (domain.ReplaceResult, error) {
	return domain.ReplaceResult{}, f.failure
}

func (f failBefore) Remove(context.Context, domain.Scope, domain.Key, domain.Expected) (domain.RemoveResult, error) {
	return domain.RemoveResult{}, f.failure
}

type loseAck struct {
	inner   Store
	failure error
}

func (l loseAck) Read(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error) {
	return l.inner.Read(ctx, scope, key)
}

func (l loseAck) List(ctx context.Context, scope domain.Scope) ([]domain.Entry, error) {
	return l.inner.List(ctx, scope)
}

func (l loseAck) Replace(ctx context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error) {
	result, err := l.inner.Replace(ctx, scope, key, value, expected)
	if err != nil || result.Status != domain.Changed {
		return result, err
	}
	return domain.ReplaceResult{}, l.failure
}

func (l loseAck) Remove(ctx context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error) {
	result, err := l.inner.Remove(ctx, scope, key, expected)
	if err != nil || !result.Removed {
		return result, err
	}
	return domain.RemoveResult{}, l.failure
}

const (
	firstWorkspace  = "01900000-0000-7000-8000-000000000001"
	secondWorkspace = "01900000-0000-7000-8000-000000000002"
)

func workspace(t *testing.T, text string) domain.Scope {
	t.Helper()
	parsed, err := id.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	scope, err := domain.ForWorkspace(parsed)
	if err != nil {
		t.Fatal(err)
	}
	return scope
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

func create(t *testing.T, store Store, scope domain.Scope, k domain.Key, value domain.Value) domain.ReplaceResult {
	t.Helper()
	result, err := store.Replace(context.Background(), scope, k, value, domain.ExpectAbsent())
	if err != nil {
		t.Fatalf("create %s: %v", k, err)
	}
	if result.Status != domain.Changed || result.Entry.Revision() != 1 {
		t.Fatalf("create %s: %+v", k, result)
	}
	return result
}

func read(t *testing.T, store Store, scope domain.Scope, k domain.Key) (domain.Entry, bool) {
	t.Helper()
	entry, found, err := store.Read(context.Background(), scope, k)
	if err != nil {
		t.Fatalf("read %s: %v", k, err)
	}
	return entry, found
}

func list(t *testing.T, store Store, scope domain.Scope) []domain.Entry {
	t.Helper()
	entries, err := store.List(context.Background(), scope)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	return entries
}

func requireConflict(t *testing.T, err error) {
	t.Helper()
	if !faults.IsKind(err, faults.Conflict) || faults.DiagnosticTypeOf(err) != "preferences.conflict" {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func m01(t *testing.T, open Factory) {
	store := open(t)
	k := key(t, "editor.theme")
	for _, scope := range []domain.Scope{domain.Global(), workspace(t, firstWorkspace)} {
		if _, found := read(t, store, scope, k); found {
			t.Fatalf("fresh store found %s in %s", k, scope)
		}
		if entries := list(t, store, scope); len(entries) != 0 {
			t.Fatalf("fresh store listed %d entries in %s", len(entries), scope)
		}
	}
}

func m02(t *testing.T, open Factory) {
	store := open(t)
	k := key(t, "editor.theme")
	result := create(t, store, domain.Global(), k, text(t, "dark"))
	if result.Event == nil || result.Event.Name() != domain.ChangedEvent {
		t.Fatal("create event")
	}
	entry, found := read(t, store, domain.Global(), k)
	if !found || entry != result.Entry {
		t.Fatalf("read after create: %+v", entry)
	}
	entries := list(t, store, domain.Global())
	if len(entries) != 1 || entries[0] != result.Entry {
		t.Fatalf("list after create: %+v", entries)
	}
}

func m03(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	create(t, store, domain.Global(), k, text(t, "dark"))
	_, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), domain.ExpectAbsent())
	requireConflict(t, err)
	result, err := store.Replace(ctx, domain.Global(), k, text(t, "dark"), revision(t, 1))
	if err != nil || result.Status != domain.Unchanged || result.Event != nil {
		t.Fatalf("equal replace: %+v %v", result, err)
	}
	entry, found := read(t, store, domain.Global(), k)
	if !found || entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatalf("state after refusals: %+v", entry)
	}
}

func m04(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	create(t, store, domain.Global(), k, text(t, "dark"))
	_, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 2))
	requireConflict(t, err)
	_, err = store.Remove(ctx, domain.Global(), k, revision(t, 2))
	requireConflict(t, err)
	entry, found := read(t, store, domain.Global(), k)
	if !found || entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatalf("state after conflicts: %+v", entry)
	}
}

func m05(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	create(t, store, domain.Global(), k, text(t, "dark"))
	second, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1))
	if err != nil || second.Status != domain.Changed || second.Entry.Revision() != 2 {
		t.Fatalf("second: %+v %v", second, err)
	}
	if value, ok := second.Event.Value(); !ok || !value.Equal(text(t, "light")) {
		t.Fatal("second event value")
	}
	entry, _ := read(t, store, domain.Global(), k)
	if entry.Revision() != 2 || !entry.Value().Equal(text(t, "light")) {
		t.Fatalf("after second: %+v", entry)
	}
	third, err := store.Replace(ctx, domain.Global(), k, domain.Int(1), revision(t, 2))
	if err != nil || third.Entry.Revision() != 3 {
		t.Fatalf("third: %+v %v", third, err)
	}
	entry, _ = read(t, store, domain.Global(), k)
	if entry.Revision() != 3 || !entry.Value().Equal(domain.Int(1)) {
		t.Fatalf("after third: %+v", entry)
	}
}

func m06(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	create(t, store, domain.Global(), k, text(t, "dark"))
	if _, err := store.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1)); err != nil {
		t.Fatal(err)
	}
	removed, err := store.Remove(ctx, domain.Global(), k, revision(t, 2))
	if err != nil || !removed.Removed || removed.Revision != 3 || removed.Event == nil || removed.Event.Name() != domain.RemovedEvent {
		t.Fatalf("remove: %+v %v", removed, err)
	}
	if _, found := read(t, store, domain.Global(), k); found {
		t.Fatal("entry present after remove")
	}
	if entries := list(t, store, domain.Global()); len(entries) != 0 {
		t.Fatalf("list after remove: %+v", entries)
	}
	again, err := store.Remove(ctx, domain.Global(), k, domain.ExpectAbsent())
	if err != nil || again.Removed || again.Event != nil {
		t.Fatalf("absent remove: %+v %v", again, err)
	}
	create(t, store, domain.Global(), k, text(t, "dark"))
}

func m07(t *testing.T, open Factory) {
	store := open(t)
	for _, name := range []string{"z.last", "m.mid", "a-b", "a.b", "a_b", "a0", "ab"} {
		create(t, store, domain.Global(), key(t, name), text(t, name))
	}
	want := []string{"a-b", "a.b", "a0", "a_b", "ab", "m.mid", "z.last"}
	for range 2 {
		entries := list(t, store, domain.Global())
		if len(entries) != len(want) {
			t.Fatalf("list length %d", len(entries))
		}
		for i, entry := range entries {
			if entry.Key().String() != want[i] {
				t.Fatalf("order at %d: %s, want %s", i, entry.Key(), want[i])
			}
		}
	}
}

func m08(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	global, first, second := domain.Global(), workspace(t, firstWorkspace), workspace(t, secondWorkspace)
	create(t, store, global, k, text(t, "global"))
	create(t, store, first, k, text(t, "first"))
	create(t, store, second, k, text(t, "second"))
	for scope, want := range map[domain.Scope]string{global: "global", first: "first", second: "second"} {
		entry, found := read(t, store, scope, k)
		if !found || !entry.Value().Equal(text(t, want)) {
			t.Fatalf("scope %s: %+v", scope, entry)
		}
		if entries := list(t, store, scope); len(entries) != 1 || entries[0].Scope() != scope {
			t.Fatalf("scope %s list: %+v", scope, entries)
		}
	}
	if _, err := store.Remove(ctx, first, k, revision(t, 1)); err != nil {
		t.Fatal(err)
	}
	if _, found := read(t, store, first, k); found {
		t.Fatal("first workspace still present")
	}
	for _, scope := range []domain.Scope{global, second} {
		if _, found := read(t, store, scope, k); !found {
			t.Fatalf("scope %s lost its entry", scope)
		}
	}
}

func m09(t *testing.T, open Factory) {
	first, second := open(t), open(t)
	k := key(t, "editor.theme")
	create(t, first, domain.Global(), k, text(t, "dark"))
	if _, found := read(t, second, domain.Global(), k); found {
		t.Fatal("second instance shares state")
	}
	if entries := list(t, second, domain.Global()); len(entries) != 0 {
		t.Fatal("second instance lists foreign entries")
	}
	if entries := list(t, open(t), domain.Global()); len(entries) != 0 {
		t.Fatal("third instance is not empty")
	}
}

func m10(t *testing.T, open Factory) {
	store := open(t)
	a, z := key(t, "a.first"), key(t, "z.last")
	create(t, store, domain.Global(), z, text(t, "z"))
	create(t, store, domain.Global(), a, text(t, "a"))
	once, _ := read(t, store, domain.Global(), a)
	twice, _ := read(t, store, domain.Global(), a)
	if once != twice {
		t.Fatal("repeated reads differ")
	}
	original := list(t, store, domain.Global())
	snapshot := append([]domain.Entry(nil), original...)
	original[0] = domain.Entry{}
	original[0], original[1] = original[1], original[0]
	again := list(t, store, domain.Global())
	if len(again) != len(snapshot) {
		t.Fatal("list length changed")
	}
	for i := range snapshot {
		if again[i] != snapshot[i] {
			t.Fatalf("stored state changed through a returned list at %d", i)
		}
	}
}

func m11(t *testing.T, open Factory) {
	ctx := context.Background()
	store := open(t)
	k := key(t, "editor.theme")
	const contenders = 8

	creates := race(contenders, func(i int) error {
		_, err := store.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
		return err
	})
	if creates.successes != 1 || creates.conflicts != contenders-1 {
		t.Fatalf("creates: %+v", creates)
	}

	var winners sync.Map
	replaces := race(contenders, func(i int) error {
		value := domain.Int(int64(i))
		_, err := store.Replace(ctx, domain.Global(), k, value, revision(t, 1))
		if err == nil {
			winners.Store(i, value)
		}
		return err
	})
	if replaces.successes != 1 || replaces.conflicts != contenders-1 {
		t.Fatalf("replaces: %+v", replaces)
	}
	entry, found := read(t, store, domain.Global(), k)
	if !found || entry.Revision() != 2 {
		t.Fatalf("final entry: %+v", entry)
	}
	var matched bool
	winners.Range(func(_, value any) bool {
		matched = entry.Value().Equal(value.(domain.Value))
		return false
	})
	if !matched {
		t.Fatal("final value is not the winner's")
	}

	distinct := race(contenders, func(i int) error {
		_, err := store.Replace(ctx, domain.Global(), key(t, "k"+string(rune('a'+i))), domain.Int(int64(i)), domain.ExpectAbsent())
		return err
	})
	if distinct.successes != contenders || distinct.conflicts != 0 {
		t.Fatalf("distinct creates: %+v", distinct)
	}
	if entries := list(t, store, domain.Global()); len(entries) != contenders+1 {
		t.Fatalf("list after distinct creates: %d", len(entries))
	}
}

type tally struct {
	successes int
	conflicts int
	others    []error
}

func race(n int, attempt func(i int) error) tally {
	var wg sync.WaitGroup
	results := make([]error, n)
	start := make(chan struct{})
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results[i] = attempt(i)
		}()
	}
	close(start)
	wg.Wait()
	var out tally
	for _, err := range results {
		switch {
		case err == nil:
			out.successes++
		case faults.IsKind(err, faults.Conflict):
			out.conflicts++
		default:
			out.others = append(out.others, err)
		}
	}
	return out
}

func m12(t *testing.T, open Factory) {
	ctx := context.Background()
	k := key(t, "editor.theme")

	inner := open(t)
	failing := FailBefore(inner, Unavailable)
	_, err := failing.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	if !errors.Is(err, Unavailable) || !faults.IsKind(err, faults.Unavailable) || faults.DiagnosticTypeOf(err) != "preferences.store_unavailable" {
		t.Fatalf("fail before: %v", err)
	}
	if _, found := read(t, inner, domain.Global(), k); found {
		t.Fatal("fail-before wrapper touched the store")
	}

	inner = open(t)
	lossy := LoseAcknowledgement(inner, Unacknowledged)
	_, err = lossy.Replace(ctx, domain.Global(), k, text(t, "dark"), domain.ExpectAbsent())
	if !errors.Is(err, Unacknowledged) || !faults.IsKind(err, faults.Timeout) || faults.DiagnosticTypeOf(err) != "preferences.store_unacknowledged" {
		t.Fatalf("lost acknowledgement: %v", err)
	}
	entry, found := read(t, lossy, domain.Global(), k)
	if !found || entry.Revision() != 1 || !entry.Value().Equal(text(t, "dark")) {
		t.Fatalf("write was not committed behind the lost acknowledgement: %+v", entry)
	}
	_, err = lossy.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 1))
	if !errors.Is(err, Unacknowledged) {
		t.Fatalf("retry through lossy wrapper: %v", err)
	}
	entry, _ = read(t, inner, domain.Global(), k)
	if entry.Revision() != 2 || !entry.Value().Equal(text(t, "light")) {
		t.Fatalf("retry did not commit: %+v", entry)
	}
	unchanged, err := lossy.Replace(ctx, domain.Global(), k, text(t, "light"), revision(t, 2))
	if err != nil || unchanged.Status != domain.Unchanged {
		t.Fatalf("unchanged write must not lose acknowledgement: %+v %v", unchanged, err)
	}
}
