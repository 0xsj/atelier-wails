// Package storetest contains the reusable logical contract for Workspace
// store adapters. It is test-only infrastructure; production code does not
// depend on it.
package storetest

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/app/command"
	"github.com/0xsj/atelier-wails/internal/workspace/app/query"
	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Store is the union of the command and query capabilities an adapter must
// satisfy.
type Store interface {
	command.Store
	query.Store
}

// Factory opens a fresh, empty, independent store for one scenario.
type Factory func(t *testing.T) Store

const (
	firstWorkspace  = "01900000-0000-7000-8000-000000000001"
	secondWorkspace = "01900000-0000-7000-8000-000000000002"
	thirdWorkspace  = "01900000-0000-7000-8000-000000000003"
)

// Run executes every shared memory-adapter scenario.
func Run(t *testing.T, open Factory) {
	t.Helper()
	scenarios := []struct {
		name string
		run  func(*testing.T, Factory)
	}{
		{"WM01FreshIsEmpty", wm01},
		{"WM02RegisterReadList", wm02},
		{"WM03LocationUniquenessAndRestoreCollision", wm03},
		{"WM04LifecycleAndRevisionConflict", wm04},
		{"WM05AbsenceAndInvalidInput", wm05},
		{"WM06DeterministicOrderingAndFilters", wm06},
		{"WM07IndependentInstances", wm07},
		{"WM08OutputIsolation", wm08},
		{"WM09ConcurrentRegistrations", wm09},
		{"WM10ConcurrentConditionalRenames", wm10},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) { scenario.run(t, open) })
	}
}

func parse(t *testing.T, value string) id.ID {
	t.Helper()
	parsed, err := id.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func at(milliseconds int64) time.Time { return time.UnixMilli(milliseconds).UTC() }

func revision(t *testing.T, value uint64) domain.Expected {
	t.Helper()
	expected, err := domain.ExpectRevision(value)
	if err != nil {
		t.Fatal(err)
	}
	return expected
}

func requireConflict(t *testing.T, err error, typ string) {
	t.Helper()
	if !faults.IsKind(err, faults.Conflict) || faults.DiagnosticTypeOf(err) != typ {
		t.Fatalf("expected %s conflict, got %v", typ, err)
	}
}

func wm01(t *testing.T, open Factory) {
	store := open(t)
	missing := parse(t, firstWorkspace)
	if _, found, err := store.Read(context.Background(), missing); err != nil || found {
		t.Fatalf("fresh read: %v %v", found, err)
	}
	for _, filter := range []domain.ListFilter{domain.All, domain.ActiveOnly} {
		items, err := store.List(context.Background(), filter)
		if err != nil || len(items) != 0 {
			t.Fatalf("fresh list: %v %v", items, err)
		}
	}
}

func wm02(t *testing.T, open Factory) {
	store := open(t)
	workspaceID := parse(t, firstWorkspace)
	registered, err := store.Register(context.Background(), workspaceID, "Atelier", "/tmp/atelier", at(100))
	if err != nil || registered.Workspace.Revision() != 1 {
		t.Fatalf("register: %+v %v", registered, err)
	}
	read, found, err := store.Read(context.Background(), workspaceID)
	if err != nil || !found || read != registered.Workspace {
		t.Fatalf("read: %+v %v %v", read, found, err)
	}
	items, err := store.List(context.Background(), domain.ActiveOnly)
	if err != nil || len(items) != 1 || items[0] != registered.Workspace {
		t.Fatalf("list: %v %v", items, err)
	}
}

func wm03(t *testing.T, open Factory) {
	store := open(t)
	first, second := parse(t, firstWorkspace), parse(t, secondWorkspace)
	if _, err := store.Register(context.Background(), first, "One", "/tmp/shared", at(100)); err != nil {
		t.Fatal(err)
	}
	_, err := store.Register(context.Background(), second, "Two", "/tmp/shared", at(100))
	requireConflict(t, err, "workspace.location_taken")
	if _, _, err := store.Archive(context.Background(), first, revision(t, 1), at(200)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Register(context.Background(), second, "Two", "/tmp/shared", at(200)); err != nil {
		t.Fatal(err)
	}
	_, _, err = store.Restore(context.Background(), first, revision(t, 2), at(300))
	requireConflict(t, err, "workspace.location_taken")
}

func wm04(t *testing.T, open Factory) {
	store := open(t)
	workspaceID := parse(t, firstWorkspace)
	if _, err := store.Register(context.Background(), workspaceID, "Atelier", "/tmp/atelier", at(100)); err != nil {
		t.Fatal(err)
	}
	_, _, err := store.Rename(context.Background(), workspaceID, "North", revision(t, 9), at(200))
	requireConflict(t, err, "workspace.conflict")
	archived, _, err := store.Archive(context.Background(), workspaceID, revision(t, 1), at(200))
	if err != nil || archived.Status != domain.Changed || archived.Workspace.Status() != domain.Archived {
		t.Fatalf("archive: %+v %v", archived, err)
	}
	unchanged, _, err := store.Archive(context.Background(), workspaceID, revision(t, 2), at(300))
	if err != nil || unchanged.Status != domain.Unchanged || unchanged.Workspace.Revision() != 2 {
		t.Fatalf("unchanged archive: %+v %v", unchanged, err)
	}
	restored, _, err := store.Restore(context.Background(), workspaceID, revision(t, 2), at(300))
	if err != nil || restored.Status != domain.Changed || restored.Workspace.Revision() != 3 {
		t.Fatalf("restore: %+v %v", restored, err)
	}
}

func wm05(t *testing.T, open Factory) {
	store := open(t)
	missing := parse(t, firstWorkspace)
	for _, operation := range []func() (bool, error){
		func() (bool, error) {
			_, found, err := store.Rename(context.Background(), missing, "Missing", revision(t, 1), at(100))
			return found, err
		},
		func() (bool, error) {
			_, found, err := store.Archive(context.Background(), missing, revision(t, 1), at(100))
			return found, err
		},
		func() (bool, error) {
			_, found, err := store.Restore(context.Background(), missing, revision(t, 1), at(100))
			return found, err
		},
		func() (bool, error) {
			_, found, err := store.Forget(context.Background(), missing, revision(t, 1))
			return found, err
		},
	} {
		found, err := operation()
		if err != nil || found {
			t.Fatalf("absence: found=%v err=%v", found, err)
		}
	}
	bad := parse(t, secondWorkspace)
	if _, err := store.Register(context.Background(), bad, "Bad", "relative/path", at(100)); !faults.IsKind(err, faults.Invalid) || faults.DiagnosticTypeOf(err) != "workspace.invalid_location" {
		t.Fatalf("invalid location: %v", err)
	}
	if _, _, err := store.Read(context.Background(), id.ID{}); !faults.IsKind(err, faults.Invalid) {
		t.Fatalf("invalid read: %v", err)
	}
}

func wm06(t *testing.T, open Factory) {
	store := open(t)
	entries := []struct {
		value, name, location string
	}{
		{thirdWorkspace, "Zulu", "/tmp/z"},
		{firstWorkspace, "Alpha", "/tmp/a"},
		{secondWorkspace, "Middle", "/tmp/m"},
	}
	for _, entry := range entries {
		if _, err := store.Register(context.Background(), parse(t, entry.value), entry.name, entry.location, at(100)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := store.Archive(context.Background(), parse(t, secondWorkspace), revision(t, 1), at(200)); err != nil {
		t.Fatal(err)
	}
	active, err := store.List(context.Background(), domain.ActiveOnly)
	if err != nil || len(active) != 2 || active[0].ID() != parse(t, firstWorkspace) || active[1].ID() != parse(t, thirdWorkspace) {
		t.Fatalf("active order: %v %v", active, err)
	}
	all, err := store.List(context.Background(), domain.All)
	if err != nil || len(all) != 3 {
		t.Fatalf("all list: %v %v", all, err)
	}
}

func wm07(t *testing.T, open Factory) {
	first, second := open(t), open(t)
	workspaceID := parse(t, firstWorkspace)
	if _, err := first.Register(context.Background(), workspaceID, "One", "/tmp/one", at(100)); err != nil {
		t.Fatal(err)
	}
	if _, found, err := second.Read(context.Background(), workspaceID); err != nil || found {
		t.Fatalf("instances share state: %v %v", found, err)
	}
}

func wm08(t *testing.T, open Factory) {
	store := open(t)
	first, second := parse(t, firstWorkspace), parse(t, secondWorkspace)
	for _, entry := range []struct {
		workspaceID    id.ID
		name, location string
	}{{first, "One", "/tmp/one"}, {second, "Two", "/tmp/two"}} {
		if _, err := store.Register(context.Background(), entry.workspaceID, entry.name, entry.location, at(100)); err != nil {
			t.Fatal(err)
		}
	}
	items, err := store.List(context.Background(), domain.All)
	if err != nil || len(items) != 2 {
		t.Fatalf("list: %v %v", items, err)
	}
	items[0], items[1] = items[1], items[0]
	again, err := store.List(context.Background(), domain.All)
	if err != nil || again[0].ID() != first || again[1].ID() != second {
		t.Fatalf("output mutation leaked: %v %v", again, err)
	}
}

func wm09(t *testing.T, open Factory) {
	store := open(t)
	workspaceID := parse(t, firstWorkspace)
	barrier := sync.NewCond(&sync.Mutex{})
	ready := 0
	start := false
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func(index int) {
			barrier.L.Lock()
			ready++
			if ready == 8 {
				start = true
				barrier.Broadcast()
			} else {
				for !start {
					barrier.Wait()
				}
			}
			barrier.L.Unlock()
			_, err := store.Register(context.Background(), workspaceID, "Atelier", "/tmp/"+strconv.Itoa(index), at(100))
			results <- err
		}(i)
	}
	successes := 0
	for i := 0; i < 8; i++ {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("duplicate registrations succeeded %d times", successes)
	}
}

func wm10(t *testing.T, open Factory) {
	store := open(t)
	workspaceID := parse(t, firstWorkspace)
	if _, err := store.Register(context.Background(), workspaceID, "Atelier", "/tmp/atelier", at(100)); err != nil {
		t.Fatal(err)
	}
	barrier := sync.NewCond(&sync.Mutex{})
	ready := 0
	start := false
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func(index int) {
			barrier.L.Lock()
			ready++
			if ready == 8 {
				start = true
				barrier.Broadcast()
			} else {
				for !start {
					barrier.Wait()
				}
			}
			barrier.L.Unlock()
			_, _, err := store.Rename(context.Background(), workspaceID, "Atelier "+strconv.Itoa(index), revision(t, 1), at(200))
			results <- err
		}(i)
	}
	successes := 0
	for i := 0; i < 8; i++ {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("conditional renames succeeded %d times", successes)
	}
	read, found, err := store.Read(context.Background(), workspaceID)
	if err != nil || !found || read.Revision() != 2 {
		t.Fatalf("rename winner: %+v %v %v", read, found, err)
	}
}
