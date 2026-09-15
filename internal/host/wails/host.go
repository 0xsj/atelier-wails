package wailshost

import (
	"context"
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Facades are the bound native facades root composes. Nil entries are not
// bound, so the Hello world host still runs without any domain slice.
type Facades struct {
	Preferences *Preferences
	Workbench   *Workbench
	Workspace   *Workspace
}

func Run(assets embed.FS, facades Facades) error {
	appMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
	}
	var bound []interface{}
	if facades.Preferences != nil {
		bound = append(bound, facades.Preferences)
	}
	if facades.Workbench != nil {
		bound = append(bound, facades.Workbench)
	}
	if facades.Workspace != nil {
		bound = append(bound, facades.Workspace)
	}
	return wails.Run(&options.App{
		Title: "Atelier · Wails", Width: 1000, Height: 700,
		AssetServer: &assetserver.Options{Assets: assets}, Menu: appMenu,
		Bind: bound,
		OnStartup: func(ctx context.Context) {
			if facades.Preferences != nil {
				facades.Preferences.start(ctx)
			}
			if facades.Workspace != nil {
				facades.Workspace.start(ctx)
			}
		},
	})
}
