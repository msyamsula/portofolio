package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ParsedMeeting struct {
	Title           string   `json:"title"`
	Attendees       []string `json:"attendees"`
	DateTime        string   `json:"datetime"`
	DurationMinutes int      `json:"duration_minutes"`
	Flexible        bool     `json:"flexible"`
	Rejected        bool     `json:"rejected"`
	RejectReason    string   `json:"reject_reason"`
}

var jakartaLoc *time.Location

func init() {
	var err error
	jakartaLoc, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		jakartaLoc = time.FixedZone("WIB", 7*3600)
	}
}

func nowJakarta() time.Time {
	return time.Now().In(jakartaLoc)
}

func parseMeetingPrompt(claudePath, prompt string, previous *ParsedMeeting) (*ParsedMeeting, error) {
	now := nowJakarta()
	maxDate := now.AddDate(0, 1, 0)

	var contextBlock string
	if previous != nil {
		prevJSON, _ := json.Marshal(previous)
		contextBlock = fmt.Sprintf(`
The user already has a meeting in progress with these details:
%s

The user's new message is a FOLLOW-UP. They may be:
- Changing the time: "make it 3pm instead", "change to tomorrow"
- Adding/removing attendees: "add Lisa too", "remove John"
- Changing the title: "call it standup"
- Changing the duration: "make it 30 minutes"
- Providing missing info: "John", "at 2pm"
- Starting over entirely with a completely new meeting request

If it's a follow-up, MERGE their changes into the existing meeting and return the updated full JSON.
If it's a completely new meeting request (mentions different people AND different time), start fresh.
`, string(prevJSON))
	}

	systemPrompt := fmt.Sprintf(`You are a meeting parser. Extract meeting details from the user's natural language request.

Current date and time: %s (Asia/Jakarta, WIB, UTC+07:00)
Maximum schedulable date: %s (1 month from now). If the requested date is beyond this, set rejected=true.
%s
Output ONLY valid JSON matching this schema:
{
  "title": "meeting title or 'Meeting' if not specified",
  "attendees": ["person name 1", "person name 2"],
  "datetime": "2025-05-26T18:00:00+07:00",
  "duration_minutes": 60,
  "flexible": false,
  "rejected": false,
  "reject_reason": ""
}

Rules:
- ALL times are in Asia/Jakarta (WIB, UTC+07:00). Always use +07:00 offset in datetime.
- The user may use 12-hour format (6pm, 9am) or 24-hour format (18:00, 09:00). Both are valid, convert to 24h in the output.
- datetime MUST always end with +07:00. Example: 2025-05-26T18:00:00+07:00
- If the user says "tomorrow at 9am", compute the actual date from current date
- If no duration is given, default to 60 minutes
- If no title is given, use "Meeting"
- If the user says they are "flexible", "free", "anytime", or "no preference", set flexible=true and datetime to empty string
- If flexible=true, still extract any time constraints (like "this week", "next Tuesday") and put the earliest possible date in datetime
- attendees should be the raw names the user mentioned, not email addresses
- If the date is more than 1 month away, set rejected=true and reject_reason="Can only schedule within the next 30 days"
- Do NOT include the creator/user in the attendees list`,
		now.Format("2006-01-02 15:04:05"),
		maxDate.Format("2006-01-02"),
		contextBlock)

	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s", systemPrompt, prompt)

	cmd := exec.Command(claudePath, "-p", fullPrompt, "--output-format", "text")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("claude parse failed: %w\nstderr: %s", err, stderr.String())
	}

	output := extractJSON(stdout.String())

	var meeting ParsedMeeting
	if err := json.Unmarshal([]byte(output), &meeting); err != nil {
		return nil, fmt.Errorf("failed to parse meeting JSON: %w\nraw: %s", err, stdout.String())
	}

	return &meeting, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return s
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}
