package viper

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
	// ViperCDRBegin marks the start of a CDR block
	ViperCDRBegin = "===== CDR BEGIN :"
	// ViperCDREnd marks the end of a CDR block
	ViperCDREnd = "===== CDR END ====="
	// ViperAgentBegin marks the start of an Agent block
	ViperAgentBegin = "===== AGENT BEGIN :"
	// ViperAgentEnd marks the end of an Agent block
	ViperAgentEnd = "===== AGENT END ====="
)

// viperMessage represents a single message from the Viper CSV
type viperMessage struct {
	SysIdent int64
	Message  string
}

// ParseViperCSV parses a Viper sample CSV file into CDR records
func ParseViperCSV(reader io.Reader) ([]format.CDRRecord, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = 2
	csvReader.LazyQuotes = true

	// Read all records
	rawRecords, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Skip header row and parse messages
	var messages []viperMessage
	for i, record := range rawRecords {
		if i == 0 && record[0] == "sysident" {
			continue // Skip header
		}

		sysIdent, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue // Skip invalid records
		}

		messages = append(messages, viperMessage{
			SysIdent: sysIdent,
			Message:  record[1],
		})
	}

	// Sort messages by sysident in ascending order (oldest first for output)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].SysIdent < messages[j].SysIdent
	})

	// Group messages into record blocks
	var records []format.CDRRecord
	var currentLines []string
	var currentID string
	var currentType string
	var inBlock bool

	for _, msg := range messages {
		trimmed := strings.TrimSpace(msg.Message)

		// Check for block markers
		if strings.HasPrefix(trimmed, ViperCDRBegin) {
			// Start of CDR block
			if inBlock && len(currentLines) > 0 {
				// Save previous block if exists
				records = append(records, format.CDRRecord{
					ID:        currentID,
					Type:      currentType,
					Timestamp: time.Now(),
					Lines:     currentLines,
				})
			}
			currentLines = []string{trimmed}
			currentType = "cdr"
			currentID = extractCallID(trimmed)
			inBlock = true
		} else if strings.HasPrefix(trimmed, ViperAgentBegin) {
			// Start of Agent block
			if inBlock && len(currentLines) > 0 {
				// Save previous block if exists
				records = append(records, format.CDRRecord{
					ID:        currentID,
					Type:      currentType,
					Timestamp: time.Now(),
					Lines:     currentLines,
				})
			}
			currentLines = []string{trimmed}
			currentType = "agent"
			currentID = ""
			inBlock = true
		} else if trimmed == ViperCDREnd || trimmed == ViperAgentEnd {
			// End of block
			if inBlock {
				currentLines = append(currentLines, trimmed)
				records = append(records, format.CDRRecord{
					ID:        currentID,
					Type:      currentType,
					Timestamp: time.Now(),
					Lines:     currentLines,
				})
				currentLines = nil
				currentID = ""
				currentType = ""
				inBlock = false
			}
		} else if inBlock && trimmed != "" {
			// Inside a block, add the line
			currentLines = append(currentLines, msg.Message)

			// Try to extract call ID from Incoming Call line
			if currentID == "" && strings.Contains(msg.Message, "Incoming Call(ID:") {
				start := strings.Index(msg.Message, "Incoming Call(ID:")
				if start >= 0 {
					start += len("Incoming Call(ID:")
					end := strings.Index(msg.Message[start:], ")")
					if end > 0 {
						currentID = strings.TrimSpace(msg.Message[start : start+end])
					}
				}
			}
		}
	}

	// Don't forget the last record if there's no trailing end marker
	if inBlock && len(currentLines) > 0 {
		records = append(records, format.CDRRecord{
			ID:        currentID,
			Type:      currentType,
			Timestamp: time.Now(),
			Lines:     currentLines,
		})
	}

	return records, nil
}

func extractCallID(line string) string {
	// CDR BEGIN lines don't typically have call ID, it comes later
	return ""
}

// ParseViperFile is a convenience function to parse a Viper file by path
func ParseViperFile(path string) ([]format.CDRRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseViperCSV(bufio.NewReader(file))
}
