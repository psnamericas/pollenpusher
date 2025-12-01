package viper

import (
	"fmt"
	"strings"
	"time"

	"cdrgenerator/format"
)

// GenerateViperRecord creates a synthetic Viper CDR record
func GenerateViperRecord(ctx *format.GenerationContext) (*format.CDRRecord, error) {
	callNum := ctx.NextCallNumber()

	ani := ctx.RandomPhoneNumber()
	location := ctx.RandomLocation()
	carrier := ctx.RandomCarrier()
	agent := ctx.RandomAgent()

	now := ctx.CurrentTime
	if now.IsZero() {
		now = time.Now()
	}

	// Random call duration between 30 seconds and 5 minutes
	duration := ctx.RandomDuration(30, 300)

	// Generate trunk and call IDs
	trunkNum := ctx.Random.Intn(10) + 1
	trunkGroup := "911"
	trunkName := fmt.Sprintf("SIP%03d", trunkNum)
	callID := fmt.Sprintf("911%03d-%05d-%s", trunkNum, callNum%100000, now.Format("20060102150405"))

	// Position/Station numbers
	posNum := ctx.Random.Intn(20) + 1
	stnNum := 2000 + posNum

	// Queue number
	queueNum := 6000 + ctx.Random.Intn(10) + 1

	// Format timestamps
	beginFormat := "01/02/06 15:04:05.000"
	aliDateFormat := "15:04  01/02"

	var lines []string

	// CDR BEGIN marker
	lines = append(lines, fmt.Sprintf("===== CDR BEGIN : %s =====", now.Format(beginFormat)))

	// System ID and trunk info
	lines = append(lines, fmt.Sprintf("00:00:00.000 [  TS] SYSTEM ID = %s", strings.ToLower(ctx.SystemID)))
	lines = append(lines, fmt.Sprintf("00:00:00.000 [VoIP] Incoming Call(ID: %s) Offered on Trunk %s/%s-%s",
		callID, trunkName, ani[:10], trunkName))
	lines = append(lines, fmt.Sprintf("00:00:00.000 [  TS] Trunk Group = %s", trunkGroup))
	lines = append(lines, "00:00:00.000 [VoIP] Call Presented")
	lines = append(lines, fmt.Sprintf("00:00:00.000 [VoIP] ANI: (40)'%s' [VALID] PseudoANI: '' [NONE]", ani))
	lines = append(lines, "00:00:00.000 [  TS] Initial ALI Request for ANI : "+ani)

	// External call identifier
	externalID := fmt.Sprintf("urn:nena:uid:callid:%s:inbcf.indigital.net", generateRandomID(ctx, 20))
	lines = append(lines, fmt.Sprintf("00:00:00.075 [VoIP] External Call-Identifier <%s>", externalID))

	// Call connected and routing
	lines = append(lines, "00:00:00.104 [VoIP] Call Connected")
	lines = append(lines, fmt.Sprintf("00:00:00.108 [VoIP] Routing call QUEUE = %d", queueNum))

	// ALI response
	lines = append(lines, fmt.Sprintf("00:00:01.696 [ PAS] Initial ALI Response received / ALI TYPE = 1"))

	// Call terminated
	durationMs := duration.Milliseconds()
	durationStr := formatDuration(duration)
	lines = append(lines, fmt.Sprintf("%s [VoIP] Caller Disconnected Before Supervision", durationStr))
	lines = append(lines, fmt.Sprintf("%s [VoIP] Call Terminated", formatDuration(duration+73*time.Millisecond)))
	lines = append(lines, fmt.Sprintf("%s [  TS] Call Completed", formatDuration(duration+73*time.Millisecond)))

	// Empty line before ALI block
	lines = append(lines, "")
	lines = append(lines, "=====   Initial ALI   ====")
	lines = append(lines, "")

	// ALI data block
	lines = append(lines, fmt.Sprintf("(%s) %s   %s",
		formatPhoneParens(ani), now.Format(aliDateFormat), ""))
	lines = append(lines, fmt.Sprintf("%s                   ", carrier.Name))
	lines = append(lines, fmt.Sprintf("%-16s", location.Address[:min(16, len(location.Address))]))
	lines = append(lines, fmt.Sprintf("%s - %s SECTOR     ",
		strings.ToUpper(location.Address), []string{"N", "S", "E", "W", "NE", "NW", "SE", "SW"}[ctx.Random.Intn(8)]))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("                              "))
	lines = append(lines, fmt.Sprintf("%-24s          ESN %s", location.City, location.ESN))
	lines = append(lines, fmt.Sprintf("CO=%s PSAP %02d POS# %02d   %s",
		carrier.Code, ctx.Random.Intn(50)+1, posNum, carrier.Type))
	lines = append(lines, "                                ")
	lines = append(lines, "      ")
	lines = append(lines, fmt.Sprintf("P#(%s)%s", ani[:3], ani[3:]))

	// Location confidence
	accuracy := 4.64 + ctx.Random.Float64()*50
	lines = append(lines, fmt.Sprintf(" UNC=%.2f     COP=90%%  Initia", accuracy))
	lines = append(lines, fmt.Sprintf("+%010.6f -%010.6f", location.Latitude, -location.Longitude))

	// Empty line
	lines = append(lines, "")

	// CDR END marker
	lines = append(lines, ViperCDREnd)

	// Optionally add AGENT block
	if ctx.Random.Float32() > 0.3 { // 70% chance of agent event
		lines = append(lines, "")
		agentLines := generateAgentBlock(ctx, agent, callID, posNum, stnNum, now)
		lines = append(lines, agentLines...)
	}

	return &format.CDRRecord{
		ID:        callID,
		Type:      "cdr",
		Timestamp: now,
		Duration:  time.Duration(durationMs) * time.Millisecond,
		Lines:     lines,
	}, nil
}

func generateAgentBlock(ctx *format.GenerationContext, agent format.Agent, callID string, posNum, stnNum int, now time.Time) []string {
	beginFormat := "01/02/06 15:04:05.000"

	var lines []string
	lines = append(lines, fmt.Sprintf("===== AGENT BEGIN : %s =====", now.Format(beginFormat)))
	lines = append(lines, fmt.Sprintf("ON CALL (ID: %s)", callID))
	lines = append(lines, "DIRECTION = incoming")
	lines = append(lines, fmt.Sprintf("ROUTE = Q%d", 6000+ctx.Random.Intn(10)+1))
	lines = append(lines, "VIPERNODE = PRIMARY")
	lines = append(lines, fmt.Sprintf("AGENT = %s/%s ROLE = %s", agent.Name, agent.ID, agent.Role))
	lines = append(lines, fmt.Sprintf("From  PSAP ID = %d PSAP Name = %s", ctx.Random.Intn(9000)+1000, ctx.PSAPName))
	lines = append(lines, fmt.Sprintf("POS = %04d / STN = %d", posNum, stnNum))
	lines = append(lines, ViperAgentEnd)

	return lines
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}

func formatPhoneParens(phone string) string {
	if len(phone) == 10 {
		return fmt.Sprintf("%s", phone[:3])
	}
	return phone
}

func generateRandomID(ctx *format.GenerationContext, length int) string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(chars[ctx.Random.Intn(len(chars))])
	}
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
