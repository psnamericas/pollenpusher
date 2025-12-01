package format

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

// CDRRecord represents a single CDR record (call or agent event)
type CDRRecord struct {
	ID        string        // Unique identifier for this record
	Type      string        // Record type: "cdr", "agent", etc.
	Timestamp time.Time     // When this record occurred
	Duration  time.Duration // Call duration (for CDR records)
	Lines     []string      // The actual output lines
}

// Output returns the record formatted for serial output
func (r *CDRRecord) Output() []byte {
	var output []byte
	for _, line := range r.Lines {
		output = append(output, []byte(line)...)
		output = append(output, '\n')
	}
	return output
}

// GenerationContext provides context for synthetic record generation
type GenerationContext struct {
	SystemID     string
	PSAPName     string
	AgentPool    []Agent
	LocationPool []Location
	CarrierPool  []Carrier
	CurrentTime  time.Time
	CallNumber   int
	Random       *rand.Rand
}

// Agent represents a call taker agent
type Agent struct {
	ID   string
	Name string
	Role string
}

// Location represents a geographic location for ALI data
type Location struct {
	Address    string
	City       string
	State      string
	Township   string
	ESN        string
	Latitude   float64
	Longitude  float64
	Altitude   float64
}

// Carrier represents a phone carrier
type Carrier struct {
	Code     string // e.g., "VZW", "TMOB", "ATTMO"
	Name     string // e.g., "VERIZON", "T-MOBILE USA, INC."
	Type     string // e.g., "WPH2" (Wireless Phase 2)
}

// CDRFormat defines the interface that all CDR format handlers must implement.
// This is the primary extension point for adding new formats.
type CDRFormat interface {
	// Name returns the unique identifier for this format (e.g., "vesta", "viper")
	Name() string

	// Description returns a human-readable description
	Description() string

	// ParseRecords reads a sample CSV file and returns parsed CDR records
	// Used in replay mode
	ParseRecords(reader io.Reader) ([]CDRRecord, error)

	// GenerateRecord creates a new synthetic CDR record
	// Used in synthetic mode
	GenerateRecord(ctx *GenerationContext) (*CDRRecord, error)
}

// NewGenerationContext creates a new generation context with default data pools
func NewGenerationContext(systemID, psapName string, seed int64) *GenerationContext {
	return &GenerationContext{
		SystemID:     systemID,
		PSAPName:     psapName,
		AgentPool:    defaultAgents(),
		LocationPool: defaultLocations(),
		CarrierPool:  defaultCarriers(),
		CurrentTime:  time.Now(),
		CallNumber:   10000000,
		Random:       rand.New(rand.NewSource(seed)),
	}
}

func defaultAgents() []Agent {
	return []Agent{
		{ID: "10001", Name: "John Smith", Role: "CALL TAKER"},
		{ID: "10002", Name: "Jane Doe", Role: "CALL TAKER"},
		{ID: "10003", Name: "Mike Johnson", Role: "CALL TAKER"},
		{ID: "10004", Name: "Sarah Williams", Role: "CALL TAKER"},
		{ID: "10005", Name: "David Brown", Role: "DISPATCHER"},
		{ID: "10006", Name: "Emily Davis", Role: "DISPATCHER"},
		{ID: "10007", Name: "Chris Wilson", Role: "CALL TAKER"},
		{ID: "10008", Name: "Amanda Miller", Role: "CALL TAKER"},
		{ID: "10009", Name: "Robert Taylor", Role: "SUPERVISOR"},
		{ID: "10010", Name: "Lisa Anderson", Role: "CALL TAKER"},
	}
}

func defaultLocations() []Location {
	return []Location{
		{Address: "123 Main St", City: "Lincoln", State: "NE", Township: "Lancaster", ESN: "123456", Latitude: 40.8136, Longitude: -96.7026, Altitude: 357},
		{Address: "456 Oak Ave", City: "Omaha", State: "NE", Township: "Douglas", ESN: "234567", Latitude: 41.2565, Longitude: -95.9345, Altitude: 332},
		{Address: "789 Elm Blvd", City: "Bellevue", State: "NE", Township: "Sarpy", ESN: "345678", Latitude: 41.1544, Longitude: -95.9146, Altitude: 305},
		{Address: "321 Pine Rd", City: "Grand Island", State: "NE", Township: "Hall", ESN: "456789", Latitude: 40.9264, Longitude: -98.3420, Altitude: 566},
		{Address: "654 Cedar Ln", City: "Kearney", State: "NE", Township: "Buffalo", ESN: "567890", Latitude: 40.6993, Longitude: -99.0817, Altitude: 652},
		{Address: "987 Maple Dr", City: "Fremont", State: "NE", Township: "Dodge", ESN: "678901", Latitude: 41.4333, Longitude: -96.4981, Altitude: 373},
		{Address: "147 Birch Way", City: "Hastings", State: "NE", Township: "Adams", ESN: "789012", Latitude: 40.5861, Longitude: -98.3884, Altitude: 595},
		{Address: "258 Walnut Ct", City: "Norfolk", State: "NE", Township: "Madison", ESN: "890123", Latitude: 42.0283, Longitude: -97.4170, Altitude: 476},
		{Address: "369 Spruce Pl", City: "Columbus", State: "NE", Township: "Platte", ESN: "901234", Latitude: 41.4297, Longitude: -97.3684, Altitude: 436},
		{Address: "480 Ash St", City: "Papillion", State: "NE", Township: "Sarpy", ESN: "012345", Latitude: 41.1544, Longitude: -96.0419, Altitude: 329},
	}
}

func defaultCarriers() []Carrier {
	return []Carrier{
		{Code: "VZW", Name: "VERIZON", Type: "WPH2"},
		{Code: "TMOB", Name: "T-MOBILE USA, INC.", Type: "WPH2"},
		{Code: "ATTMO", Name: "AT&T Mobility", Type: "WPH2"},
		{Code: "SPRINT", Name: "SPRINT", Type: "WPH2"},
		{Code: "USCC", Name: "US CELLULAR", Type: "WPH2"},
	}
}

// Helper methods for GenerationContext

// RandomAgent returns a random agent from the pool
func (ctx *GenerationContext) RandomAgent() Agent {
	return ctx.AgentPool[ctx.Random.Intn(len(ctx.AgentPool))]
}

// RandomLocation returns a random location from the pool
func (ctx *GenerationContext) RandomLocation() Location {
	return ctx.LocationPool[ctx.Random.Intn(len(ctx.LocationPool))]
}

// RandomCarrier returns a random carrier from the pool
func (ctx *GenerationContext) RandomCarrier() Carrier {
	return ctx.CarrierPool[ctx.Random.Intn(len(ctx.CarrierPool))]
}

// RandomPhoneNumber generates a random 10-digit phone number
func (ctx *GenerationContext) RandomPhoneNumber() string {
	areaCode := 200 + ctx.Random.Intn(800)
	exchange := 200 + ctx.Random.Intn(800)
	subscriber := ctx.Random.Intn(10000)
	return formatPhone(areaCode, exchange, subscriber)
}

// RandomDuration returns a random duration between min and max seconds
func (ctx *GenerationContext) RandomDuration(minSec, maxSec int) time.Duration {
	if maxSec <= minSec {
		return time.Duration(minSec) * time.Second
	}
	secs := minSec + ctx.Random.Intn(maxSec-minSec)
	return time.Duration(secs) * time.Second
}

// NextCallNumber increments and returns the next call number
func (ctx *GenerationContext) NextCallNumber() int {
	ctx.CallNumber++
	return ctx.CallNumber
}

func formatPhone(areaCode, exchange, subscriber int) string {
	return fmt.Sprintf("%03d%03d%04d", areaCode, exchange, subscriber)
}
