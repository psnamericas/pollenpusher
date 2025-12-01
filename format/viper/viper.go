package viper

import (
	"cdrgenerator/format"
	"io"
)

func init() {
	format.MustRegister(&ViperFormat{})
}

// ViperFormat implements the CDRFormat interface for Viper 911 systems
type ViperFormat struct{}

// Name returns the format identifier
func (f *ViperFormat) Name() string {
	return "viper"
}

// Description returns a human-readable description
func (f *ViperFormat) Description() string {
	return "Viper VoIP 911 Call Handling System"
}

// ParseRecords parses a Viper sample CSV file into CDR records
func (f *ViperFormat) ParseRecords(reader io.Reader) ([]format.CDRRecord, error) {
	return ParseViperCSV(reader)
}

// GenerateRecord creates a new synthetic Viper CDR record
func (f *ViperFormat) GenerateRecord(ctx *format.GenerationContext) (*format.CDRRecord, error) {
	return GenerateViperRecord(ctx)
}
