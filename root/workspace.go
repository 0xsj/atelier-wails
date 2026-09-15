package root

import (
	"os"
	"path/filepath"

	wailshost "github.com/0xsj/atelier-wails/internal/host/wails"
	"github.com/0xsj/atelier-wails/internal/workspace/app/command"
	"github.com/0xsj/atelier-wails/internal/workspace/app/query"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/persistence"
	"github.com/0xsj/atelier-wails/internal/workspace/transport/desktop"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

const workspaceDocument = "workspace.json"

func workspacePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", faults.New(faults.Unavailable, "user configuration directory is unavailable").
			WithType("root.config_dir_unavailable").WithCause(err)
	}
	return filepath.Join(base, "atelier-wails", workspaceDocument), nil
}

func composeWorkspace(path string) (*wailshost.Workspace, error) {
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
	return wailshost.NewWorkspace(handler)
}
