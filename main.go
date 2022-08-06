package main

import (
	"context"
	"embed"
	"sync"
	"time"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed frontend/dist
var assets embed.FS


type App struct {
	ctx context.Context
	secretKey string
	wg sync.WaitGroup
	priceHandlers map[string]func(hashName string)
} 


func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.secretKey = ""
	a.priceHandlers = make(map[string]func(hashName string), 5)
	a.wg = sync.WaitGroup{}
	for {
		time.Sleep(time.Second * 1)
		for hashName, handler := range a.priceHandlers {
			a.wg.Add(1)
			go handler(hashName)
		}
		a.wg.Wait()
	}
}


func main() {
	app := &App{}
	err := wails.Run(&options.App{
		Title:            "Csgomarket trading",
		Width:            1280,
		Height:           768,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
