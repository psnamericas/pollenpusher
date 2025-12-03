package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainUI represents the main user interface
type MainUI struct {
	window     fyne.Window
	dashboard  *DashboardTab
	portConfig *PortConfigTab
	control    *ControlTab
}

// NewMainUI creates a new main UI
func NewMainUI(window fyne.Window) *MainUI {
	ui := &MainUI{
		window: window,
	}

	// Create tabs
	ui.dashboard = NewDashboardTab()
	ui.portConfig = NewPortConfigTab(window)
	ui.control = NewControlTab()

	return ui
}

// Build constructs the UI layout
func (m *MainUI) Build() *fyne.Container {
	// Create tab container
	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", m.dashboard.Build()),
		container.NewTabItem("Port Configuration", m.portConfig.Build()),
		container.NewTabItem("Service Control", m.control.Build()),
	)

	return container.NewBorder(
		m.buildHeader(),
		m.buildFooter(),
		nil,
		nil,
		tabs,
	)
}

// buildHeader creates the header section
func (m *MainUI) buildHeader() *fyne.Container {
	title := widget.NewLabelWithStyle("CDR Generator Control Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true})

	return container.NewVBox(
		title,
		widget.NewSeparator(),
	)
}

// buildFooter creates the footer section
func (m *MainUI) buildFooter() *fyne.Container {
	status := widget.NewLabel("Status: Ready")

	return container.NewVBox(
		widget.NewSeparator(),
		status,
	)
}
