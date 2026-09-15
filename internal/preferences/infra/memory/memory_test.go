package memory_test

import (
	"context"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/memory"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/storetest"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// TestContract runs M01–M12 from the shared suite against fresh instances.
func TestContract(t *testing.T) {
	storetest.Run(t, func(*testing.T) storetest.Store { return memory.New() })
}

// Go-only guard: zero inputs are refused with the domain categories before
// the map is touched.
func TestInvalidReadInputs(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	key, _ := domain.NewKey("editor.theme")
	_, _, err := store.Read(ctx, domain.Scope{}, key)
	if !faults.IsKind(err, faults.Invalid) || faults.DiagnosticTypeOf(err) != "preferences.invalid_scope" {
		t.Fatalf("zero scope: %v", err)
	}
	_, _, err = store.Read(ctx, domain.Global(), domain.Key{})
	if !faults.IsKind(err, faults.Invalid) || faults.DiagnosticTypeOf(err) != "preferences.invalid_key" {
		t.Fatalf("zero key: %v", err)
	}
	_, err = store.List(ctx, domain.Scope{})
	if !faults.IsKind(err, faults.Invalid) || faults.DiagnosticTypeOf(err) != "preferences.invalid_scope" {
		t.Fatalf("zero list scope: %v", err)
	}
	_, err = store.Replace(ctx, domain.Global(), domain.Key{}, domain.Bool(true), domain.ExpectAbsent())
	if !faults.IsKind(err, faults.Invalid) || faults.DiagnosticTypeOf(err) != "preferences.invalid_key" {
		t.Fatalf("zero replace key: %v", err)
	}
	if entries, _ := store.List(ctx, domain.Global()); len(entries) != 0 {
		t.Fatal("refused write left state")
	}
}
