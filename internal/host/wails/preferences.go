package wailshost

import (
	"context"
	"sync"

	"github.com/0xsj/atelier-wails/internal/preferences/transport/desktop"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Preferences is the Wails-bound facade for the preferences desktop
// transport. Each exported method is one wire operation: it takes the raw
// request document as a string and returns the outcome envelope, which Wails
// serializes with the transport's JSON tags. It holds no framework type.
type Preferences struct {
	handler *desktop.Handler
	mu      sync.RWMutex
	ctx     context.Context
}

// NewPreferences refuses a nil handler.
func NewPreferences(handler *desktop.Handler) (*Preferences, error) {
	if handler == nil {
		return nil, faults.New(faults.Internal, "preferences facade needs a handler").
			WithType("preferences.missing_dependency")
	}
	return &Preferences{handler: handler, ctx: context.Background()}, nil
}

// start binds the application lifetime context; unexported so Wails does not
// bind it. Before startup the facade uses a background context.
func (p *Preferences) start(ctx context.Context) {
	if ctx == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ctx = ctx
}

func (p *Preferences) context() context.Context {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ctx
}

func (p *Preferences) Read(document string) desktop.ReadOutcome {
	return p.handler.ReadDocument(p.context(), []byte(document))
}

func (p *Preferences) List(document string) desktop.ListOutcome {
	return p.handler.ListDocument(p.context(), []byte(document))
}

func (p *Preferences) Replace(document string) desktop.ReplaceOutcome {
	return p.handler.ReplaceDocument(p.context(), []byte(document))
}

func (p *Preferences) Remove(document string) desktop.RemoveOutcome {
	return p.handler.RemoveDocument(p.context(), []byte(document))
}

func (p *Preferences) Resolve(document string) desktop.ResolveOutcome {
	return p.handler.ResolveDocument(p.context(), []byte(document))
}
