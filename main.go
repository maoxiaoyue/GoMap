package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"map/app/views"
)

func main() {
	a := app.New()
	w := a.NewWindow("Map")
	w.Resize(fyne.NewSize(800, 600))

	// 載入主畫面
	w.SetContent(views.MainView(w))

	w.ShowAndRun()
}
