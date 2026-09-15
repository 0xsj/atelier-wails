package wailshost

import (
	"context"
	"sync"

	"github.com/0xsj/atelier-wails/internal/workspace/transport/desktop"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Workspace is the Wails-bound facade for the Workspace desktop transport.
// Bound methods accept one raw JSON request document and return plain outcome
// data; no framework type crosses the transport boundary.
type Workspace struct {
	handler *desktop.Handler
	mu      sync.RWMutex
	ctx     context.Context
}

func NewWorkspace(handler *desktop.Handler) (*Workspace, error) {
	if handler == nil {
		return nil, faults.New(faults.Internal, "workspace facade needs a handler").WithType("workspace.missing_dependency")
	}
	return &Workspace{handler: handler, ctx: context.Background()}, nil
}

func (w *Workspace) start(ctx context.Context) {
	if ctx == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ctx = ctx
}

func (w *Workspace) context() context.Context {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.ctx
}

func (w *Workspace) Read(document string) desktop.ReadOutcome {
	return w.handler.ReadDocument(w.context(), []byte(document))
}
func (w *Workspace) List(document string) desktop.ListOutcome {
	return w.handler.ListDocument(w.context(), []byte(document))
}
func (w *Workspace) Register(document string) desktop.RegisteredOutcome {
	return w.handler.RegisterDocument(w.context(), []byte(document))
}
func (w *Workspace) Rename(document string) desktop.MutationOutcome {
	return w.handler.RenameDocument(w.context(), []byte(document))
}
func (w *Workspace) Archive(document string) desktop.MutationOutcome {
	return w.handler.ArchiveDocument(w.context(), []byte(document))
}
func (w *Workspace) Restore(document string) desktop.MutationOutcome {
	return w.handler.RestoreDocument(w.context(), []byte(document))
}
func (w *Workspace) Forget(document string) desktop.ForgetOutcome {
	return w.handler.ForgetDocument(w.context(), []byte(document))
}
