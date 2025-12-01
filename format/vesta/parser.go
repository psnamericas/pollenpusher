package vesta

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"cdrgenerator/format"
)

const (
	// VestaSeparator is the line separator between Vesta records
	VestaSeparator = "---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---"
)

// vestaMessage represents a single message from the Vesta CSV
type vestaMessage struct {
	SysIdent int64
	Message  string
}

// ParseVestaCSV parses a Vesta sample CSV file into CDR records
func ParseVestaCSV(reader io.Reader) ([]format.CDRRecord, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = 2
	csvReader.LazyQuotes = true

	// Read all records
	rawRecords, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip header row and parse messages
	var messages []vestaMessage
	for i, record := range rawRecords {
		if i == 0 && record[0] == "sysident" {
			continue // Skip header
		}

		sysIdent, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue // Skip invalid records
		}

		messages = append(messages, vestaMessage{
			SysIdent: sysIdent,
			Message:  record[1],
		})
	}

	// Sort messages by sysident in ascending order (oldest first for output)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].SysIdent < messages[j].SysIdent
	})

	// Group messages into record blocks (separated by VestaSeparator)
	var records []format.CDRRecord
	var currentLines []string
	var currentID string

	for _, msg := range messages {
		if msg.Message == VestaSeparator {
			// End of current record block
			if len(currentLines) > 0 {
				record := format.CDRRecord{
					ID:        currentID,
					Type:      "cdr",
					Timestamp: time.Now(),
					Lines:     currentLines,
				}
				// Add separator at the end
				record.Lines = append(record.Lines, VestaSeparator)
				records = append(records, record)
				currentLines = nil
				currentID = ""
			}
		} else if msg.Message == "" {
			// Skip empty messages
			continue
		} else {
			// Extract call ID if this is a call event line
			if strings.Contains(msg.Message, "Call ") && strings.Contains(msg.Message, "Arrives On") {
				// Extract call number like "Call 10105965"
				parts := strings.Fields(msg.Message)
				for i, part := range parts {
					if part == "Call" && i+1 < len(parts) {
						currentID = parts[i+1]
						break
					}
				}
			}
			currentLines = append(currentLines, msg.Message)
		}
	}

	// Don't forget the last record if there's no trailing separator
	if len(currentLines) > 0 {
		record := format.CDRRecord{
			ID:        currentID,
			Type:      "cdr",
			Timestamp: time.Now(),
			Lines:     currentLines,
		}
		records = append(records, record)
	}

	return records, nil
}

// ParseVestaFile is a convenience function to parse a Vesta file by path
func ParseVestaFile(path string) ([]format.CDRRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseVestaCSV(bufio.NewReader(file))
}
