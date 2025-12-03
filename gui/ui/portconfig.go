package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"cdrgenerator/config"
)

// PortConfigTab represents the port configuration UI
type PortConfigTab struct {
	configPath  string
	config      *config.Config
	portList    *widget.List
	selectedIdx int
	window      fyne.Window
}

// NewPortConfigTab creates a new port configuration tab
func NewPortConfigTab(window fyne.Window) *PortConfigTab {
	return &PortConfigTab{
		configPath:  "/etc/cdrgenerator/config.json",
		selectedIdx: -1,
		window:      window,
	}
}

// Build constructs the port configuration UI
func (p *PortConfigTab) Build() *fyne.Container {
	// Load configuration
	p.loadConfig()

	// Port list
	p.portList = widget.NewList(
		func() int {
			if p.config == nil {
				return 0
			}
			return len(p.config.Ports)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if p.config != nil && id < len(p.config.Ports) {
				port := p.config.Ports[id]
				status := ""
				if port.Enabled {
					status = "[ENABLED]"
				} else {
					status = "[DISABLED]"
				}
				label.SetText(fmt.Sprintf("%s %s - %s @ %d baud", status, port.Device, port.Format, port.BaudRate))
			}
		},
	)

	p.portList.OnSelected = func(id widget.ListItemID) {
		p.selectedIdx = id
	}

	// Buttons
	addBtn := widget.NewButton("Add Port", func() {
		p.showAddPortDialog()
	})

	editBtn := widget.NewButton("Edit Port", func() {
		if p.selectedIdx >= 0 && p.selectedIdx < len(p.config.Ports) {
			p.showEditPortDialog(p.selectedIdx)
		} else {
			dialog.ShowInformation("No Selection", "Please select a port to edit", p.window)
		}
	})

	deleteBtn := widget.NewButton("Delete Port", func() {
		if p.selectedIdx >= 0 && p.selectedIdx < len(p.config.Ports) {
			p.deletePort(p.selectedIdx)
		} else {
			dialog.ShowInformation("No Selection", "Please select a port to delete", p.window)
		}
	})

	saveBtn := widget.NewButton("Save Configuration", func() {
		p.saveConfig()
	})
	saveBtn.Importance = widget.HighImportance

	reloadBtn := widget.NewButton("Reload Configuration", func() {
		p.loadConfig()
	})

	buttons := container.NewVBox(
		addBtn,
		editBtn,
		deleteBtn,
		widget.NewSeparator(),
		saveBtn,
		reloadBtn,
	)

	// Info panel
	infoLabel := widget.NewLabel("Configuration file: " + p.configPath)
	infoLabel.Wrapping = fyne.TextWrapWord

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Port Configuration"),
			widget.NewSeparator(),
			infoLabel,
		),
		nil,
		nil,
		buttons,
		container.NewScroll(p.portList),
	)
}

// loadConfig loads the configuration from file
func (p *PortConfigTab) loadConfig() {
	cfg, err := config.Load(p.configPath)
	if err != nil {
		// Try local config
		p.configPath = "configs/example-config.json"
		cfg, err = config.Load(p.configPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load config: %w", err), p.window)
			return
		}
	}

	p.config = cfg
	if p.portList != nil {
		p.portList.Refresh()
	}
}

// saveConfig saves the configuration to file
func (p *PortConfigTab) saveConfig() {
	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to marshal config: %w", err), p.window)
		return
	}

	if err := os.WriteFile(p.configPath, data, 0644); err != nil {
		dialog.ShowError(fmt.Errorf("failed to write config: %w", err), p.window)
		return
	}

	dialog.ShowInformation("Success", "Configuration saved successfully", p.window)
}

// showAddPortDialog shows a dialog to add a new port
func (p *PortConfigTab) showAddPortDialog() {
	newPort := config.PortConfig{
		Device:         "/dev/ttyS0",
		BaudRate:       9600,
		Format:         "vesta",
		Mode:           "replay",
		CallsPerMinute: 2.5,
		Enabled:        true,
	}

	p.showPortEditDialog(&newPort, func() {
		p.config.Ports = append(p.config.Ports, newPort)
		p.portList.Refresh()
	})
}

// showEditPortDialog shows a dialog to edit an existing port
func (p *PortConfigTab) showEditPortDialog(idx int) {
	port := &p.config.Ports[idx]
	p.showPortEditDialog(port, func() {
		p.portList.Refresh()
	})
}

// showPortEditDialog shows the port edit dialog
func (p *PortConfigTab) showPortEditDialog(port *config.PortConfig, onSave func()) {
	// Create form entries
	deviceEntry := widget.NewEntry()
	deviceEntry.SetText(port.Device)

	baudRateEntry := widget.NewEntry()
	baudRateEntry.SetText(strconv.Itoa(port.BaudRate))

	formatSelect := widget.NewSelect([]string{"vesta", "viper"}, func(value string) {
		port.Format = value
	})
	formatSelect.SetSelected(port.Format)

	modeSelect := widget.NewSelect([]string{"replay", "synthetic"}, func(value string) {
		port.Mode = value
	})
	modeSelect.SetSelected(port.Mode)

	sampleFileEntry := widget.NewEntry()
	sampleFileEntry.SetText(port.SampleFile)

	callsPerMinEntry := widget.NewEntry()
	callsPerMinEntry.SetText(fmt.Sprintf("%.1f", port.CallsPerMinute))

	enabledCheck := widget.NewCheck("", func(checked bool) {
		port.Enabled = checked
	})
	enabledCheck.SetChecked(port.Enabled)

	descEntry := widget.NewEntry()
	descEntry.SetText(port.Description)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Device", Widget: deviceEntry},
			{Text: "Baud Rate", Widget: baudRateEntry},
			{Text: "Format", Widget: formatSelect},
			{Text: "Mode", Widget: modeSelect},
			{Text: "Sample File", Widget: sampleFileEntry},
			{Text: "Calls/Minute", Widget: callsPerMinEntry},
			{Text: "Enabled", Widget: enabledCheck},
			{Text: "Description", Widget: descEntry},
		},
		OnSubmit: func() {
			// Validate and save
			baudRate, err := strconv.Atoi(baudRateEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid baud rate: %w", err), p.window)
				return
			}

			callsPerMin, err := strconv.ParseFloat(callsPerMinEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid calls per minute: %w", err), p.window)
				return
			}

			port.Device = deviceEntry.Text
			port.BaudRate = baudRate
			port.SampleFile = sampleFileEntry.Text
			port.CallsPerMinute = callsPerMin
			port.Description = descEntry.Text

			onSave()
		},
	}

	dialog.ShowForm("Edit Port Configuration", "Save", "Cancel", form.Items, func(submitted bool) {
		if submitted {
			form.OnSubmit()
		}
	}, p.window)
}

// deletePort deletes a port from the configuration
func (p *PortConfigTab) deletePort(idx int) {
	dialog.ShowConfirm("Delete Port", "Are you sure you want to delete this port?", func(confirmed bool) {
		if confirmed {
			p.config.Ports = append(p.config.Ports[:idx], p.config.Ports[idx+1:]...)
			p.selectedIdx = -1
			p.portList.Refresh()
		}
	}, p.window)
}
