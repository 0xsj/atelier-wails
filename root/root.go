// Package root composes the native process and owns its lifetime.
package root

import (
	"embed"

	wailshost "github.com/0xsj/atelier-wails/internal/host/wails"
	"github.com/0xsj/atelier-wails/pkg/logger"
)

func Run(assets embed.FS) error {
	log, err := consoleLogger()
	if err != nil {
		return err
	}
	path, err := preferencesPath()
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "preferences.unavailable", nil))
		return err
	}
	preferences, err := composePreferences(path)
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "preferences.unavailable", nil))
		return err
	}
	snapshotPath, err := workbenchSnapshotPath()
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "workbench.snapshot.unavailable", nil))
		return err
	}
	workbench, err := composeWorkbench(snapshotPath)
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "workbench.snapshot.unavailable", nil))
		return err
	}
	workspaceDocumentPath, err := workspacePath()
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "workspace.unavailable", nil))
		return err
	}
	workspace, err := composeWorkspace(workspaceDocumentPath)
	if err != nil {
		reportLog(log.WithError(err).Log(logger.Error, "workspace.unavailable", nil))
		return err
	}
	facades := wailshost.Facades{Preferences: preferences, Workbench: workbench, Workspace: workspace}
	return runLogged(log, func() error { return wailshost.Run(assets, facades) })
}
