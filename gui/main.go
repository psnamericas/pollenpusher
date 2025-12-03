package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"cdrgenerator/gui/ui"
)

func main() {
	// Create the app
	myApp := app.New()
	myWindow := myApp.NewWindow("CDR Generator Control Panel")
	myWindow.Resize(fyne.NewSize(1200, 800))

	// Create the main UI
	mainUI := ui.NewMainUI(myWindow)

	// Set up the window content
	myWindow.SetContent(mainUI.Build())
	myWindow.ShowAndRun()
}
