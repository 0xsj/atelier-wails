// Package command owns Workspace mutations and the application ports they
// consume. See ../CONTRACT.md.
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Store is the conditional write capability. The adapter owns the atomic
// decision and state change; found=false means the workspace is absent.
type Store interface {
	Register(context.Context, id.ID, string, string, time.Time) (domain.RegisterResult, error)
	Rename(context.Context, id.ID, string, domain.Expected, time.Time) (domain.MutationResult, bool, error)
	Archive(context.Context, id.ID, domain.Expected, time.Time) (domain.MutationResult, bool, error)
	Restore(context.Context, id.ID, domain.Expected, time.Time) (domain.MutationResult, bool, error)
	Forget(context.Context, id.ID, domain.Expected) (domain.ForgetResult, bool, error)
}

// Publisher receives returned events synchronously after the store commits.
// Publication is in-process and non-durable.
type Publisher interface {
	Publish(context.Context, domain.Event)
}

// Discard is the publisher for roots without subscribers.
type Discard struct{}

func (Discard) Publish(context.Context, domain.Event) {}

type Service struct {
	store     Store
	publisher Publisher
}

func New(store Store, publisher Publisher) (*Service, error) {
	if store == nil || publisher == nil {
		return nil, faults.New(faults.Internal, "workspace command service needs a store and a publisher").
			WithType("workspace.missing_dependency")
	}
	return &Service{store: store, publisher: publisher}, nil
}

func (s *Service) Register(ctx context.Context, workspaceID id.ID, name, location string, at time.Time) (domain.RegisterResult, error) {
	const operation = "workspace.register"
	result, err := s.store.Register(ctx, workspaceID, name, location, at)
	if err != nil {
		return domain.RegisterResult{}, annotate(operation, err)
	}
	s.publisher.Publish(ctx, result.Event)
	return result, nil
}

func (s *Service) Rename(ctx context.Context, workspaceID id.ID, name string, expected domain.Expected, at time.Time) (domain.MutationResult, error) {
	return s.mutate(ctx, "workspace.rename", func() (domain.MutationResult, bool, error) {
		return s.store.Rename(ctx, workspaceID, name, expected, at)
	})
}

func (s *Service) Archive(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, error) {
	return s.mutate(ctx, "workspace.archive", func() (domain.MutationResult, bool, error) {
		return s.store.Archive(ctx, workspaceID, expected, at)
	})
}

func (s *Service) Restore(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, error) {
	return s.mutate(ctx, "workspace.restore", func() (domain.MutationResult, bool, error) {
		return s.store.Restore(ctx, workspaceID, expected, at)
	})
}

func (s *Service) Forget(ctx context.Context, workspaceID id.ID, expected domain.Expected) (domain.ForgetResult, bool, error) {
	const operation = "workspace.forget"
	result, found, err := s.store.Forget(ctx, workspaceID, expected)
	if err != nil {
		return domain.ForgetResult{}, false, annotate(operation, err)
	}
	if found {
		s.publisher.Publish(ctx, result.Event)
	}
	return result, found, nil
}

func (s *Service) mutate(ctx context.Context, operation string, call func() (domain.MutationResult, bool, error)) (domain.MutationResult, error) {
	result, found, err := call()
	if err != nil {
		return domain.MutationResult{}, annotate(operation, err)
	}
	if !found {
		return domain.MutationResult{}, annotate(operation, absent())
	}
	if result.Status == domain.Changed && result.Event != nil {
		s.publisher.Publish(ctx, *result.Event)
	}
	return result, nil
}

func annotate(operation string, err error) error { return fmt.Errorf("%s: %w", operation, err) }

func absent() error {
	return faults.New(faults.NotFound, "workspace not found").WithType("workspace.not_found")
}
