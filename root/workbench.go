package root

import (
	"os"
	"path/filepath"

	wailshost "github.com/0xsj/atelier-wails/internal/host/wails"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
)

const workbenchSnapshotDocument = "workbench.json"

func workbenchSnapshotPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", faults.New(faults.Unavailable, "user configuration directory is unavailable").
			WithType("root.config_dir_unavailable").
			WithCause(err)
	}
	return filepath.Join(base, "atelier-wails", workbenchSnapshotDocument), nil
}

func composeWorkbench(path string) (*wailshost.Workbench, error) {
	if err := fileio.EnsureDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	return wailshost.NewWorkbench(path), nil
}
