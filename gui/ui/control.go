package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ControlTab represents the service control UI
type ControlTab struct {
	serviceName   string
	statusLabel   *widget.Label
	outputText    *widget.Entry
	window        fyne.Window
}

// NewControlTab creates a new control tab
func NewControlTab() *ControlTab {
	return &ControlTab{
		serviceName: "cdrgenerator.service",
	}
}

// Build constructs the control UI
func (c *ControlTab) Build() *fyne.Container {
	// Status display
	c.statusLabel = widget.NewLabel("Service Status: Unknown")
	c.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	statusCard := widget.NewCard("Current Status", "", c.statusLabel)

	// Control buttons
	startBtn := widget.NewButton("Start Service", func() {
		c.executeCommand("start")
	})
	startBtn.Importance = widget.SuccessImportance

	stopBtn := widget.NewButton("Stop Service", func() {
		c.executeCommand("stop")
	})
	stopBtn.Importance = widget.DangerImportance

	restartBtn := widget.NewButton("Restart Service", func() {
		c.executeCommand("restart")
	})
	restartBtn.Importance = widget.WarningImportance

	statusBtn := widget.NewButton("Check Status", func() {
		c.checkStatus()
	})

	enableBtn := widget.NewButton("Enable Auto-Start", func() {
		c.executeCommand("enable")
	})

	disableBtn := widget.NewButton("Disable Auto-Start", func() {
		c.executeCommand("disable")
	})

	controlButtons := container.NewGridWithColumns(3,
		startBtn,
		stopBtn,
		restartBtn,
	)

	autoStartButtons := container.NewGridWithColumns(2,
		enableBtn,
		disableBtn,
	)

	// Output log
	c.outputText = widget.NewMultiLineEntry()
	c.outputText.SetPlaceHolder("Command output will appear here...")
	c.outputText.Wrapping = fyne.TextWrapWord

	outputCard := widget.NewCard("Command Output", "", container.NewScroll(c.outputText))

	// Main layout
	content := container.NewBorder(
		container.NewVBox(
			statusCard,
			widget.NewSeparator(),
			widget.NewLabel("Service Control"),
			controlButtons,
			statusBtn,
			widget.NewSeparator(),
			widget.NewLabel("Auto-Start Configuration"),
			autoStartButtons,
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		outputCard,
	)

	// Initial status check
	go c.checkStatus()

	return content
}

// executeCommand executes a systemctl command
func (c *ControlTab) executeCommand(action string) {
	c.appendOutput(fmt.Sprintf("Executing: systemctl %s %s\n", action, c.serviceName))

	cmd := exec.Command("systemctl", action, c.serviceName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		c.appendOutput(fmt.Sprintf("Error: %v\n", err))
	}

	c.appendOutput(string(output))
	c.appendOutput("\n")

	// Update status after command
	c.checkStatus()
}

// checkStatus checks the current service status
func (c *ControlTab) checkStatus() {
	cmd := exec.Command("systemctl", "status", c.serviceName)
	output, err := cmd.CombinedOutput()

	statusText := string(output)

	// Parse status
	if strings.Contains(statusText, "Active: active (running)") {
		c.statusLabel.SetText("Service Status: RUNNING")
		c.statusLabel.Importance = widget.SuccessImportance
	} else if strings.Contains(statusText, "Active: inactive") {
		c.statusLabel.SetText("Service Status: STOPPED")
		c.statusLabel.Importance = widget.MediumImportance
	} else if strings.Contains(statusText, "Active: failed") {
		c.statusLabel.SetText("Service Status: FAILED")
		c.statusLabel.Importance = widget.DangerImportance
	} else {
		c.statusLabel.SetText("Service Status: UNKNOWN")
		c.statusLabel.Importance = widget.WarningImportance
	}

	if err != nil {
		c.appendOutput(fmt.Sprintf("Status check error: %v\n", err))
	}

	c.appendOutput("Status updated\n")
}

// appendOutput appends text to the output display
func (c *ControlTab) appendOutput(text string) {
	c.outputText.SetText(c.outputText.Text + text)
	// Scroll to bottom
	c.outputText.CursorRow = len(strings.Split(c.outputText.Text, "\n"))
}
