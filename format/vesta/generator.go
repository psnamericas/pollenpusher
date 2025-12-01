package vesta

import (
	"fmt"
	"strings"
	"time"

	"cdrgenerator/format"
)

// GenerateVestaRecord creates a synthetic Vesta CDR record
func GenerateVestaRecord(ctx *format.GenerationContext) (*format.CDRRecord, error) {
	callNum := ctx.NextCallNumber()
	callID := fmt.Sprintf("%d", callNum)

	ani := ctx.RandomPhoneNumber()
	cpn := ctx.RandomPhoneNumber()
	location := ctx.RandomLocation()
	carrier := ctx.RandomCarrier()
	_ = ctx.RandomAgent() // Reserved for future agent tracking

	now := ctx.CurrentTime
	if now.IsZero() {
		now = time.Now()
	}

	// Random call duration between 30 seconds and 5 minutes
	duration := ctx.RandomDuration(30, 300)
	endTime := now.Add(duration)

	// Generate position/device names
	deviceNum := ctx.Random.Intn(10) + 1
	posDevice := fmt.Sprintf("DCD%02d", deviceNum)
	eimDevice := fmt.Sprintf("DCDEIM911%d", ctx.Random.Intn(5)+1)
	queueName := "DCD-911"

	// Format timestamps
	dateFormat := "Jan/02/06 15:04:05 EST"
	aliDateFormat := "01/02/2006"
	aliTimeFormat := "15:04:05.0MST"

	var lines []string

	// PSAP identifier line
	lines = append(lines, fmt.Sprintf("%d %s", 3001, "Nebraska"))

	// Call event line (all events on one line, space-separated)
	callEvents := fmt.Sprintf("ANI             %s                                                      CPN             %s                                                                                                                                      Call %s   Arrives On               %s     %s %s           Goes Off Hook                            %s %s           Queue In                 %s         %s Call %s   Cellular Call                            %s Call %s   CPN: %s                          %s %s         Queue Out (Answered)     %s           %s %s          Picks Up                                 %s %s     Is Released                              %s %s          Hangs Up                 Call %s   %s %s          Releases                 Call %s   %s Call %s   Finishes                                 %s",
		ani, cpn,
		callID, eimDevice, now.Format(dateFormat),
		eimDevice, now.Format(dateFormat),
		eimDevice, queueName, now.Format(dateFormat),
		callID, now.Add(2*time.Second).Format(dateFormat),
		callID, cpn, now.Add(2*time.Second).Format(dateFormat),
		queueName, posDevice, now.Add(4*time.Second).Format(dateFormat),
		posDevice, now.Add(4*time.Second).Format(dateFormat),
		eimDevice, endTime.Format(dateFormat),
		posDevice, callID, endTime.Format(dateFormat),
		posDevice, callID, endTime.Format(dateFormat),
		callID, endTime.Format(dateFormat),
	)
	lines = append(lines, callEvents)

	// ALI Information marker
	lines = append(lines, "ALI Information")

	// Location/ALI data line
	locTech := []string{"Handset AGPS", "Handset GPS", "Hybrid Device Based", "Hybrid Unspecified"}[ctx.Random.Intn(4)]
	confidence := 90
	accuracy := 4.64 + ctx.Random.Float64()*50

	aliLine := fmt.Sprintf("%s   CBN %s    %s  %s     %sEST        %s%s                         ESN %s           %s                                                    Township:                               %s                            %sComments:                               %s                                                       %s PositionX=%+010.6f            %d%% sure callerY=%+010.6f         within %.2f metersZ=%03d+/-%.12f                                                          LAW:                                    FIR:                                    EMS:                                    LocTechn:%s            MIN:           IMIN:                    Tabular/Legacy route %s",
		formatPhoneWithDashes(ani),
		formatPhoneWithDashes(cpn),
		carrier.Code,
		now.Format(aliDateFormat),
		now.Format(aliTimeFormat),
		carrier.Type,
		carrier.Name,
		location.ESN,
		strings.ToUpper(location.Address),
		strings.ToUpper(location.Township),
		location.State,
		strings.ToUpper(location.City[:min(len(location.City), 20)]),
		"Updated",
		location.Longitude,
		confidence,
		location.Latitude,
		accuracy,
		int(location.Altitude),
		ctx.Random.Float64()*10,
		locTech,
		strings.ToUpper(location.City),
	)
	lines = append(lines, aliLine)

	// SIP Call IDs marker and ID
	lines = append(lines, "SIP Call IDs")
	sipID := generateSIPCallID(ctx)
	lines = append(lines, sipID)

	// Separator
	lines = append(lines, VestaSeparator)

	return &format.CDRRecord{
		ID:        callID,
		Type:      "cdr",
		Timestamp: now,
		Duration:  duration,
		Lines:     lines,
	}, nil
}

func formatPhoneWithDashes(phone string) string {
	if len(phone) == 10 {
		return fmt.Sprintf("%s-%s-%s", phone[:3], phone[3:6], phone[6:])
	}
	return phone
}

func generateSIPCallID(ctx *format.GenerationContext) string {
	// Generate a random base64-like SIP call ID
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var sb strings.Builder
	for i := 0; i < 22; i++ {
		sb.WriteByte(chars[ctx.Random.Intn(len(chars))])
	}
	sb.WriteString("..")
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
