package main

import (
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)


func (a *App) FolowError(hashName string) {
	if r := recover(); r != nil {
		a.RemoveItemFollow(hashName)
		err := fmt.Sprintf("Помилка відсідковування: %s", hashName)
		runtime.EventsEmit(a.ctx, "onError", err)
	}
}