// Package query owns the Preferences reads and the read port they consume.
// See ../CONTRACT.md. Resolve merges a caller-supplied fallback and never
// writes; defaults belong to the consuming feature, not to storage.
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Store is the read capability the queries need. Read reports found-or-absent
// distinctly from failure; List returns entries sorted by key.
type Store interface {
	Read(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error)
	List(ctx context.Context, scope domain.Scope) ([]domain.Entry, error)
}

// Resolved is a value merged with a fallback. Revision is set only when Stored.
type Resolved struct {
	Value    domain.Value
	Stored   bool
	Revision uint64
}

// Service executes preference reads over its port.
type Service struct{ store Store }

// New refuses a nil store.
func New(store Store) (*Service, error) {
	if store == nil {
		return nil, missingDependency()
	}
	return &Service{store: store}, nil
}

// Read returns the stored entry, or false when absent.
func (s *Service) Read(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error) {
	const operation = "preferences.read"
	if err := validateTarget(scope, key); err != nil {
		return domain.Entry{}, false, annotate(operation, err)
	}
	entry, found, err := s.store.Read(ctx, scope, key)
	if err != nil {
		return domain.Entry{}, false, annotate(operation, err)
	}
	return entry, found, nil
}

// List returns every entry in the scope in the store's key order.
func (s *Service) List(ctx context.Context, scope domain.Scope) ([]domain.Entry, error) {
	const operation = "preferences.list"
	if !scope.Valid() {
		return nil, annotate(operation, invalid("preferences.invalid_scope"))
	}
	entries, err := s.store.List(ctx, scope)
	if err != nil {
		return nil, annotate(operation, err)
	}
	return entries, nil
}

// Resolve returns the stored value with its revision, or the fallback marked
// as not stored. It never writes the fallback.
func (s *Service) Resolve(ctx context.Context, scope domain.Scope, key domain.Key, fallback domain.Value) (Resolved, error) {
	const operation = "preferences.resolve"
	if err := validateTarget(scope, key); err != nil {
		return Resolved{}, annotate(operation, err)
	}
	if !fallback.Valid() {
		return Resolved{}, annotate(operation, invalid("preferences.invalid_value"))
	}
	entry, found, err := s.store.Read(ctx, scope, key)
	if err != nil {
		return Resolved{}, annotate(operation, err)
	}
	if !found {
		return Resolved{Value: fallback}, nil
	}
	return Resolved{Value: entry.Value(), Stored: true, Revision: entry.Revision()}, nil
}

func validateTarget(scope domain.Scope, key domain.Key) error {
	if !scope.Valid() {
		return invalid("preferences.invalid_scope")
	}
	if !key.Valid() {
		return invalid("preferences.invalid_key")
	}
	return nil
}

func annotate(operation string, err error) error {
	return fmt.Errorf("%s: %w", operation, err)
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid preference input").WithType(typ)
}

func missingDependency() error {
	return faults.New(faults.Internal, "preferences query service needs a store").
		WithType("preferences.missing_dependency")
}
