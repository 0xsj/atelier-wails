package root

import (
	"os"
	"path/filepath"

	wailshost "github.com/0xsj/atelier-wails/internal/host/wails"
	"github.com/0xsj/atelier-wails/internal/preferences/app/command"
	"github.com/0xsj/atelier-wails/internal/preferences/app/query"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/persistence"
	"github.com/0xsj/atelier-wails/internal/preferences/transport/desktop"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

const preferencesDocument = "preferences.json"

// preferencesPath is the per-user document under the platform config
// directory: ~/Library/Application Support/atelier-wails on macOS.
func preferencesPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", faults.New(faults.Unavailable, "user configuration directory is unavailable").
			WithType("root.config_dir_unavailable").
			WithCause(err)
	}
	return filepath.Join(base, "atelier-wails", preferencesDocument), nil
}

// composePreferences selects the file-backed adapter explicitly as the
// default store: a document that cannot be read refuses startup rather than
// starting empty. The memory adapter remains for tests and a deliberate
// ephemeral mode. One store serves both ports; the publisher is Discard until
// a subscriber exists.
func composePreferences(path string) (*wailshost.Preferences, error) {
	if err := fileio.EnsureDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	store, err := persistence.Open(path)
	if err != nil {
		return nil, err
	}
	commands, err := command.New(store, command.Discard{})
	if err != nil {
		return nil, err
	}
	queries, err := query.New(store)
	if err != nil {
		return nil, err
	}
	handler, err := desktop.NewHandler(commands, queries)
	if err != nil {
		return nil, err
	}
	return wailshost.NewPreferences(handler)
}
