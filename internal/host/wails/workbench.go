package wailshost

import (
	"errors"
	"sync"
	"unicode/utf8"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

// Workbench is the Wails-bound file document facade for workbench restoration.
// The native side owns the config path and atomic replacement; JSON validation
// remains in the frontend model.
type Workbench struct {
	mu   sync.RWMutex
	path string
}

func NewWorkbench(path string) *Workbench {
	return &Workbench{path: path}
}

// Load returns nil when no snapshot has been saved yet.
func (w *Workbench) Load() (*string, error) {
	w.mu.RLock()
	path := w.path
	w.mu.RUnlock()
	data, found, err := fileio.Read(path)
	if err != nil {
		return nil, unavailable(err)
	}
	if !found {
		return nil, nil
	}
	if !utf8.Valid(data) {
		return nil, unavailable(errors.New("snapshot document is not valid UTF-8"))
	}
	document := string(data)
	return &document, nil
}

func (w *Workbench) Save(document string) error {
	w.mu.RLock()
	path := w.path
	w.mu.RUnlock()
	if err := fileio.Replace(path, []byte(document)); err != nil {
		return unavailable(err)
	}
	return nil
}

func unavailable(cause error) error {
	return faults.New(faults.Unavailable, "workbench snapshot unavailable").
		WithType("workbench.snapshot_unavailable").
		WithCause(cause)
}
