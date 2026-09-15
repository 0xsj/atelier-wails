// Package command owns the Preferences mutations and the ports they consume:
// a conditional store and an in-process publisher for returned event facts.
// See ../CONTRACT.md. Production code never constructs an adapter here.
package command

import (
	"context"
	"fmt"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Store is the conditional write capability the commands need. An adapter
// applies the domain decision and its state change as one atomic operation
// and returns domain refusals unchanged.
type Store interface {
	Replace(ctx context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error)
	Remove(ctx context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error)
}

// Publisher receives each returned event once, synchronously, after the store
// has committed. Publication cannot fail and is not durable delivery.
type Publisher interface {
	Publish(ctx context.Context, event domain.Event)
}

// Discard is the publisher for roots without subscribers.
type Discard struct{}

func (Discard) Publish(context.Context, domain.Event) {}

// Service executes preference mutations over its ports.
type Service struct {
	store     Store
	publisher Publisher
}

// New refuses nil dependencies; use Discard when nothing subscribes.
func New(store Store, publisher Publisher) (*Service, error) {
	if store == nil || publisher == nil {
		return nil, missingDependency()
	}
	return &Service{store: store, publisher: publisher}, nil
}

// Replace conditionally creates or replaces one preference and publishes the
// changed event only when the store reports a change.
func (s *Service) Replace(ctx context.Context, scope domain.Scope, key domain.Key, value domain.Value, expected domain.Expected) (domain.ReplaceResult, error) {
	const operation = "preferences.replace"
	if err := validateTarget(scope, key, expected); err != nil {
		return domain.ReplaceResult{}, annotate(operation, err)
	}
	if !value.Valid() {
		return domain.ReplaceResult{}, annotate(operation, invalid("preferences.invalid_value"))
	}
	result, err := s.store.Replace(ctx, scope, key, value, expected)
	if err != nil {
		return domain.ReplaceResult{}, annotate(operation, err)
	}
	if result.Status == domain.Changed && result.Event != nil {
		s.publisher.Publish(ctx, *result.Event)
	}
	return result, nil
}

// Remove conditionally removes one preference. An absent entry is a successful
// no-op that publishes nothing.
func (s *Service) Remove(ctx context.Context, scope domain.Scope, key domain.Key, expected domain.Expected) (domain.RemoveResult, error) {
	const operation = "preferences.remove"
	if err := validateTarget(scope, key, expected); err != nil {
		return domain.RemoveResult{}, annotate(operation, err)
	}
	result, err := s.store.Remove(ctx, scope, key, expected)
	if err != nil {
		return domain.RemoveResult{}, annotate(operation, err)
	}
	if result.Removed && result.Event != nil {
		s.publisher.Publish(ctx, *result.Event)
	}
	return result, nil
}

func validateTarget(scope domain.Scope, key domain.Key, expected domain.Expected) error {
	if !scope.Valid() {
		return invalid("preferences.invalid_scope")
	}
	if !key.Valid() {
		return invalid("preferences.invalid_key")
	}
	if !expected.Valid() {
		return invalid("preferences.invalid_revision")
	}
	return nil
}

// annotate adds the operation name while leaving the inner classification,
// condition identity and cause reachable through Unwrap.
func annotate(operation string, err error) error {
	return fmt.Errorf("%s: %w", operation, err)
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid preference input").WithType(typ)
}

func missingDependency() error {
	return faults.New(faults.Internal, "preferences command service needs a store and a publisher").
		WithType("preferences.missing_dependency")
}
