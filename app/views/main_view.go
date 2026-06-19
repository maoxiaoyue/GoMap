package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainView 回傳應用程式主畫面
func MainView(w fyne.Window) fyne.CanvasObject {
	//@ai:think intent=組出 Fyne 主畫面的元件樹，作為桌面模式的起始畫面（待接上選目錄與掃描流程） model=claude-opus-4-8
	title := widget.NewLabel("Welcome to Map")
	title.TextStyle = fyne.TextStyle{Bold: true}

	info := widget.NewLabel("Built with HypGo + Fyne")

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		info,
		widget.NewButton("Click me", func() {
			info.SetText("Button clicked!")
		}),
	)
}
