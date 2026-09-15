// Package persistence stores preferences in one JSON document replaced
// atomically through the shared fileio leaf. It implements the command and
// query store ports, passes the shared store scenario suite, and adds
// restart, atomic-commit and corruption behavior. See CONTRACT.md.
package persistence

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

// Store is one file-backed preference store. Reads are served from memory;
// every change rewrites the whole document before memory is updated.
type Store struct {
	mu    sync.RWMutex
	path  string
	state items
}

// Open reads the document at path once. A missing file is an empty store;
// an unreadable or invalid file is a failure, never an empty store.
func Open(path string) (*Store, error) {
	data, found, err := fileio.Read(path)
	if err != nil {
		return nil, unavailable(path, err)
	}
	state := make(items)
	if found {
		state, err = decodeDocument(data)
		if err != nil {
			if errors.Is(err, errUnsupported) {
				return nil, faults.New(faults.Unavailable, "preference storage format is not supported").
					WithType("preferences.storage_unsupported").
					WithDetail("path", path)
			}
			return nil, faults.New(faults.Unavailable, "preference storage is corrupt").
				WithType("preferences.storage_corrupt").
				WithDetail("path", path).
				WithDetail("reason", err.Error())
		}
	}
	return &Store{path: path, state: state}, nil
}

func (s *Store) Read(_ context.Context, scope domain.Scope, key domain.Key) (domain.Entry, bool, error) {
	if err := validateTarget(scope, key); err != nil {
		return domain.Entry{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.state[scope][key]
	return entry, ok, nil
}

func (s *Store) List(_ context.Context, scope domain.Scope) ([]domain.Entry, error) {
	if !scope.Valid() {
		return nil, invalid("preferences.invalid_scope")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.state[scope]
	entries := make([]domain.Entry, 0, len(bucket))
	for _, entry := range bucket {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key().String() < entries[j].Key().String() })
	return entries, nil
}

// Replace decides under the lock, writes the whole document, and commits to
// memory only after the file replacement succeeded.
func (s *Store) Replace(_ context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := domain.DecideReplace(s.current(scope, key), scope, key, value, expected)
	if err != nil || result.Status != domain.Changed {
		return result, err
	}
	next := s.cloneState()
	bucket := next[scope]
	if bucket == nil {
		bucket = make(map[domain.Key]domain.Entry)
		next[scope] = bucket
	}
	bucket[key] = result.Entry
	if err := s.persist(next); err != nil {
		return domain.ReplaceResult{}, err
	}
	s.state = next
	return result, nil
}

// Remove mirrors Replace: decide, persist, then commit to memory.
func (s *Store) Remove(_ context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := domain.DecideRemove(s.current(scope, key), scope, key, expected)
	if err != nil || !result.Removed {
		return result, err
	}
	next := s.cloneState()
	delete(next[scope], key)
	if len(next[scope]) == 0 {
		delete(next, scope)
	}
	if err := s.persist(next); err != nil {
		return domain.RemoveResult{}, err
	}
	s.state = next
	return result, nil
}

func (s *Store) persist(next items) error {
	data, err := encodeDocument(next)
	if err != nil {
		return unavailable(s.path, err)
	}
	if err := fileio.Replace(s.path, data); err != nil {
		return unavailable(s.path, err)
	}
	return nil
}

// current must be called with the lock held.
func (s *Store) current(scope domain.Scope, key domain.Key) *domain.Entry {
	entry, ok := s.state[scope][key]
	if !ok {
		return nil
	}
	return &entry
}

func (s *Store) cloneState() items {
	next := make(items, len(s.state))
	for scope, bucket := range s.state {
		copied := make(map[domain.Key]domain.Entry, len(bucket))
		for key, entry := range bucket {
			copied[key] = entry
		}
		next[scope] = copied
	}
	return next
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

func unavailable(path string, cause error) error {
	return faults.New(faults.Unavailable, "preference storage is unavailable").
		WithType("preferences.storage_unavailable").
		WithDetail("path", path).
		WithCause(cause)
}
