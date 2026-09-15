// Package query owns Workspace reads and the application read port. See
// ../CONTRACT.md.
package query

import (
	"context"
	"fmt"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Store reports read absence distinctly and returns deterministic lists.
type Store interface {
	Read(context.Context, id.ID) (domain.Workspace, bool, error)
	List(context.Context, domain.ListFilter) ([]domain.Workspace, error)
}

type Service struct{ store Store }

func New(store Store) (*Service, error) {
	if store == nil {
		return nil, faults.New(faults.Internal, "workspace query service needs a store").
			WithType("workspace.missing_dependency")
	}
	return &Service{store: store}, nil
}

func (s *Service) Read(ctx context.Context, workspaceID id.ID) (domain.Workspace, bool, error) {
	const operation = "workspace.read"
	if workspaceID.IsZero() {
		return domain.Workspace{}, false, annotate(operation, invalid("workspace.invalid_id"))
	}
	workspace, found, err := s.store.Read(ctx, workspaceID)
	if err != nil {
		return domain.Workspace{}, false, annotate(operation, err)
	}
	return workspace, found, nil
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.Workspace, error) {
	const operation = "workspace.list"
	if !filter.Valid() {
		return nil, annotate(operation, invalid("workspace.invalid_filter"))
	}
	workspaces, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, annotate(operation, err)
	}
	return workspaces, nil
}

func annotate(operation string, err error) error { return fmt.Errorf("%s: %w", operation, err) }

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid workspace input").WithType(typ)
}
