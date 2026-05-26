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
	Ambiguous       bool     `json:"ambiguous"`
	Clarification   string   `json:"clarification"`
	EditIntent      bool     `json:"edit_intent"`
	SearchKeywords  string   `json:"search_keywords"`
	SummaryIntent   bool     `json:"summary_intent"`
	SummaryStart    string   `json:"summary_start"`
	SummaryEnd      string   `json:"summary_end"`
	DeleteIntent    bool     `json:"delete_intent"`
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
IMPORTANT — EXISTING MEETING CONTEXT:
The user already has a meeting being scheduled with these details:
%s

The user's new message is a FOLLOW-UP to modify this meeting. They may want to:
- Change the title/name: "call it standup", "rename to sync", "change the name to weekly review"
- Change the time: "make it 3pm instead", "change to tomorrow"
- Add/remove attendees: "add Lisa too", "remove John"
- Change the duration: "make it 30 minutes"
- Provide missing info: "John", "at 2pm"

CRITICAL MERGE RULES:
1. Start from the existing meeting JSON above as a baseline.
2. ONLY change the specific field(s) the user mentions. Keep ALL other fields exactly as they are.
3. If the user says "call it X", "rename to X", "change name/title to X", update ONLY the title to X.
4. If the user changes time, update ONLY datetime. Keep title, attendees, duration unchanged.
5. If the user adds attendees, APPEND to the existing list. Keep title, datetime, duration unchanged.
6. NEVER reset fields to defaults unless the user explicitly asks.

If the user's message is AMBIGUOUS — you cannot tell whether they want to modify the existing meeting or start a new one — set ambiguous=true and clarification to a short question asking the user what they meant. Do NOT guess. Examples of ambiguous input:
- "meeting with Bob at 2pm" (could be editing attendees/time OR a new meeting)
- "standup" (could be a title change OR a new meeting called standup)

Only start a completely fresh meeting if the user's message is CLEARLY an entirely new meeting request (explicitly mentions different people AND a different time AND a new topic).
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
  "reject_reason": "",
  "ambiguous": false,
  "clarification": "",
  "edit_intent": false,
  "search_keywords": "",
  "summary_intent": false,
  "summary_start": "",
  "summary_end": "",
  "delete_intent": false
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
- Do NOT include the creator/user in the attendees list
- EDIT DETECTION: If the user wants to modify an EXISTING event on their calendar (not a draft), set edit_intent=true. Keywords: "move my", "reschedule my", "change my", "update my", "cancel my", "edit my", "push back my". Put relevant search terms in search_keywords (e.g. "standup", "meeting with John"). When edit_intent=true, still extract the NEW values (new time, new title, etc.) into the regular fields — leave unchanged fields empty.
- DELETE DETECTION: If the user wants to DELETE, CANCEL, or REMOVE an existing event, set delete_intent=true AND edit_intent=true. Use the same search_keywords field to identify which event. Keywords: "delete my", "cancel my", "remove my", "drop my". When delete_intent=true, all other meeting fields can be left at defaults.
- SUMMARY DETECTION: If the user wants to VIEW or SUMMARIZE their schedule, set summary_intent=true. Keywords: "what's my schedule", "what do I have", "show my meetings", "my agenda", "am I free", "what's on my calendar", "summarize my week". Set summary_start and summary_end as RFC3339 datetime strings with +07:00 offset for the requested range. Examples: "today" = start of today to end of today, "this week" = today to end of this week (Sunday), "next 3 days" = today to 3 days from now. Maximum range is 1 month. When summary_intent=true, all other meeting fields can be left at defaults.`,
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
