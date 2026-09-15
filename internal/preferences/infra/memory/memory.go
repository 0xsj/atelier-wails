// Package memory provides an isolated, synchronized preference store that
// implements the command and query store ports. State lasts only for the
// instance; root must select it explicitly. See CONTRACT.md.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Store is one independent in-process preference store.
type Store struct {
	mu    sync.RWMutex
	items map[domain.Scope]map[domain.Key]domain.Entry
}

// New returns an empty instance. There is no shared default and no seed.
func New() *Store { return &Store{items: make(map[domain.Scope]map[domain.Key]domain.Entry)} }

// Read reports the stored entry, or false when absent.
func (s *Store) Read(_ context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error) {
	if err := validateTarget(scope, key); err != nil {
		return domain.Entry{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.items[scope][key]
	return entry, ok, nil
}

// List returns the scope's entries in bytewise key order; unknown scopes
// list empty.
func (s *Store) List(_ context.Context, scope domain.Scope) ([]domain.Entry, error) {
	if !scope.Valid() {
		return nil, invalid("preferences.invalid_scope")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.items[scope]
	entries := make([]domain.Entry, 0, len(bucket))
	for _, entry := range bucket {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key().String() < entries[j].Key().String() })
	return entries, nil
}

// Replace applies the domain decision and its state change under one lock.
func (s *Store) Replace(_ context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := domain.DecideReplace(s.current(scope, key), scope, key, value, expected)
	if err != nil || result.Status != domain.Changed {
		return result, err
	}
	bucket := s.items[scope]
	if bucket == nil {
		bucket = make(map[domain.Key]domain.Entry)
		s.items[scope] = bucket
	}
	bucket[key] = result.Entry
	return result, nil
}

// Remove applies the domain decision and its state change under one lock.
func (s *Store) Remove(_ context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := domain.DecideRemove(s.current(scope, key), scope, key, expected)
	if err != nil || !result.Removed {
		return result, err
	}
	delete(s.items[scope], key)
	if len(s.items[scope]) == 0 {
		delete(s.items, scope)
	}
	return result, nil
}

// current must be called with the lock held. It copies the entry so the
// domain never sees the map's storage.
func (s *Store) current(scope domain.Scope, key domain.Key) *domain.Entry {
	entry, ok := s.items[scope][key]
	if !ok {
		return nil
	}
	return &entry
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

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid preference input").WithType(typ)
}
