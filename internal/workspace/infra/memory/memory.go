// Package memory provides an isolated, synchronized Workspace registry. It
// satisfies both application ports and is the reference adapter for the
// reusable workspace-memory scenarios.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

type Store struct {
	mu    sync.RWMutex
	items map[id.ID]domain.Workspace
}

func New() *Store { return &Store{items: make(map[id.ID]domain.Workspace)} }

func (s *Store) Register(ctx context.Context, workspaceID id.ID, name, location string, at time.Time) (domain.RegisterResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.RegisterResult{}, canceled(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[workspaceID]; exists {
		return domain.RegisterResult{}, conflict("workspace.id_taken")
	}
	if s.activeLocationTaken(location, workspaceID) {
		return domain.RegisterResult{}, conflict("workspace.location_taken")
	}
	result, err := domain.Register(workspaceID, name, location, at)
	if err != nil {
		return domain.RegisterResult{}, err
	}
	s.items[workspaceID] = result.Workspace
	return result, nil
}

func (s *Store) Read(ctx context.Context, workspaceID id.ID) (domain.Workspace, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.Workspace{}, false, canceled(err)
	}
	if workspaceID.IsZero() {
		return domain.Workspace{}, false, invalid("workspace.invalid_id")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.items[workspaceID]
	return value, ok, nil
}

func (s *Store) List(ctx context.Context, filter domain.ListFilter) ([]domain.Workspace, error) {
	if err := ctx.Err(); err != nil {
		return nil, canceled(err)
	}
	if !filter.Valid() {
		return nil, invalid("workspace.invalid_filter")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]domain.Workspace, 0, len(s.items))
	for _, item := range s.items {
		if filter == domain.ActiveOnly && item.Status() != domain.Active {
			continue
		}
		values = append(values, item)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID().String() < values[j].ID().String() })
	return values, nil
}

func (s *Store) Rename(ctx context.Context, workspaceID id.ID, name string, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, canceled(err)
	}
	return s.mutate(workspaceID, func(current domain.Workspace) (domain.MutationResult, error) {
		return domain.Rename(current, name, expected, at)
	})
}

func (s *Store) Archive(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, canceled(err)
	}
	return s.mutate(workspaceID, func(current domain.Workspace) (domain.MutationResult, error) {
		return domain.Archive(current, expected, at)
	})
}

func (s *Store) Restore(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.MutationResult{}, false, canceled(err)
	}
	return s.mutate(workspaceID, func(current domain.Workspace) (domain.MutationResult, error) {
		if current.Status() == domain.Archived && s.activeLocationTaken(current.Location(), workspaceID) {
			return domain.MutationResult{}, conflict("workspace.location_taken")
		}
		return domain.Restore(current, expected, at)
	})
}

func (s *Store) Forget(ctx context.Context, workspaceID id.ID, expected domain.Expected) (domain.ForgetResult, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.ForgetResult{}, false, canceled(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[workspaceID]
	if !ok {
		return domain.ForgetResult{}, false, nil
	}
	result, err := domain.Forget(current, expected)
	if err != nil {
		return domain.ForgetResult{}, true, err
	}
	delete(s.items, workspaceID)
	return result, true, nil
}

// mutate runs one decision under the write lock and commits a change.
func (s *Store) mutate(workspaceID id.ID, decide func(domain.Workspace) (domain.MutationResult, error)) (domain.MutationResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[workspaceID]
	if !ok {
		return domain.MutationResult{}, false, nil
	}
	result, err := decide(current)
	if err != nil {
		return domain.MutationResult{}, true, err
	}
	if result.Status == domain.Changed {
		s.items[workspaceID] = result.Workspace
	}
	return result, true, nil
}

// activeLocationTaken must be called with the lock held.
func (s *Store) activeLocationTaken(location string, except id.ID) bool {
	for otherID, item := range s.items {
		if otherID != except && item.Status() == domain.Active && item.Location() == location {
			return true
		}
	}
	return false
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid workspace input").WithType(typ)
}

func conflict(typ string) error {
	return faults.New(faults.Conflict, "workspace already registered").WithType(typ)
}

func canceled(cause error) error {
	return faults.Wrap(cause, faults.Canceled, "workspace operation canceled")
}
