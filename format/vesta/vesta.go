package vesta

import (
	"cdrgenerator/format"
	"io"
)

func init() {
	format.MustRegister(&VestaFormat{})
}

// VestaFormat implements the CDRFormat interface for Vesta 911 systems
type VestaFormat struct{}

// Name returns the format identifier
func (f *VestaFormat) Name() string {
	return "vesta"
}

// Description returns a human-readable description
func (f *VestaFormat) Description() string {
	return "Vesta 911 Call Handling System"
}

// ParseRecords parses a Vesta sample CSV file into CDR records
func (f *VestaFormat) ParseRecords(reader io.Reader) ([]format.CDRRecord, error) {
	return ParseVestaCSV(reader)
}

// GenerateRecord creates a new synthetic Vesta CDR record
func (f *VestaFormat) GenerateRecord(ctx *format.GenerationContext) (*format.CDRRecord, error) {
	return GenerateVestaRecord(ctx)
}
