package main

import (
	"fmt"
	"testing"
)

func TestEditIntentDetection(t *testing.T) {
	claudePath := "claude"

	tests := []struct {
		name       string
		prompt     string
		wantEdit   bool
		wantSearch bool
	}{
		{
			name:       "move my standup",
			prompt:     "move my standup to 3pm",
			wantEdit:   true,
			wantSearch: true,
		},
		{
			name:       "reschedule meeting with John",
			prompt:     "reschedule my meeting with John to tomorrow",
			wantEdit:   true,
			wantSearch: true,
		},
		{
			name:       "new meeting creation",
			prompt:     "meeting with Sarah tomorrow at 10am",
			wantEdit:   false,
			wantSearch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMeetingPrompt(claudePath, tt.prompt, nil)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			fmt.Printf("[%s] edit_intent=%v search_keywords=%q title=%q datetime=%q attendees=%v\n",
				tt.name, result.EditIntent, result.SearchKeywords, result.Title, result.DateTime, result.Attendees)

			if result.EditIntent != tt.wantEdit {
				t.Errorf("edit_intent: got %v, want %v", result.EditIntent, tt.wantEdit)
			}
			if tt.wantSearch && result.SearchKeywords == "" {
				t.Error("expected non-empty search_keywords")
			}
		})
	}
}

func TestAmbiguityDetection(t *testing.T) {
	claudePath := "claude"

	// Step 1: Create an initial meeting
	initial, err := parseMeetingPrompt(claudePath, "meeting with John tomorrow at 2pm", nil)
	if err != nil {
		t.Fatalf("initial parse failed: %v", err)
	}
	fmt.Printf("INITIAL: title=%q attendees=%v datetime=%q duration=%d\n",
		initial.Title, initial.Attendees, initial.DateTime, initial.DurationMinutes)

	// Step 2: Ambiguous follow-up — "standup" could be title change or new meeting
	ambig, err := parseMeetingPrompt(claudePath, "standup", initial)
	if err != nil {
		t.Fatalf("ambiguous parse failed: %v", err)
	}
	fmt.Printf("AMBIGUOUS INPUT 'standup': ambiguous=%v clarification=%q title=%q attendees=%v\n",
		ambig.Ambiguous, ambig.Clarification, ambig.Title, ambig.Attendees)

	// Step 3: Clear title change — should NOT be ambiguous
	titleChange, err := parseMeetingPrompt(claudePath, "rename it to weekly sync", initial)
	if err != nil {
		t.Fatalf("title change parse failed: %v", err)
	}
	fmt.Printf("CLEAR TITLE CHANGE 'rename it to weekly sync': ambiguous=%v title=%q attendees=%v datetime=%q\n",
		titleChange.Ambiguous, titleChange.Title, titleChange.Attendees, titleChange.DateTime)

	if titleChange.Title != "weekly sync" && titleChange.Title != "Weekly Sync" && titleChange.Title != "Weekly sync" {
		t.Errorf("expected title ~ 'weekly sync', got %q", titleChange.Title)
	}
	if len(titleChange.Attendees) == 0 {
		t.Error("attendees were lost after title change")
	}

	// Step 4: Another ambiguous input — "meeting with Bob at 3pm"
	ambig2, err := parseMeetingPrompt(claudePath, "meeting with Bob at 3pm", initial)
	if err != nil {
		t.Fatalf("ambiguous2 parse failed: %v", err)
	}
	fmt.Printf("AMBIGUOUS INPUT 'meeting with Bob at 3pm': ambiguous=%v clarification=%q title=%q attendees=%v\n",
		ambig2.Ambiguous, ambig2.Clarification, ambig2.Title, ambig2.Attendees)
}

func TestSummaryIntentDetection(t *testing.T) {
	claudePath := "claude"

	tests := []struct {
		name        string
		prompt      string
		wantSummary bool
	}{
		{
			name:        "whats my schedule today",
			prompt:      "what's my schedule today?",
			wantSummary: true,
		},
		{
			name:        "show my meetings this week",
			prompt:      "show my meetings this week",
			wantSummary: true,
		},
		{
			name:        "am I free tomorrow afternoon",
			prompt:      "am I free tomorrow afternoon?",
			wantSummary: true,
		},
		{
			name:        "new meeting not summary",
			prompt:      "schedule a meeting with Bob at 2pm",
			wantSummary: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMeetingPrompt(claudePath, tt.prompt, nil)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			fmt.Printf("[%s] summary_intent=%v summary_start=%q summary_end=%q\n",
				tt.name, result.SummaryIntent, result.SummaryStart, result.SummaryEnd)

			if result.SummaryIntent != tt.wantSummary {
				t.Errorf("summary_intent: got %v, want %v", result.SummaryIntent, tt.wantSummary)
			}
			if tt.wantSummary && (result.SummaryStart == "" || result.SummaryEnd == "") {
				t.Error("expected non-empty summary_start and summary_end")
			}
		})
	}
}

func TestDeleteIntentDetection(t *testing.T) {
	claudePath := "claude"

	tests := []struct {
		name       string
		prompt     string
		wantDelete bool
		wantSearch bool
	}{
		{
			name:       "cancel my standup",
			prompt:     "cancel my standup tomorrow",
			wantDelete: true,
			wantSearch: true,
		},
		{
			name:       "delete meeting with John",
			prompt:     "delete my meeting with John",
			wantDelete: true,
			wantSearch: true,
		},
		{
			name:       "remove my 1on1",
			prompt:     "remove my 1on1",
			wantDelete: true,
			wantSearch: true,
		},
		{
			name:       "create not delete",
			prompt:     "meeting with Sarah at 3pm",
			wantDelete: false,
			wantSearch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMeetingPrompt(claudePath, tt.prompt, nil)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			fmt.Printf("[%s] delete_intent=%v edit_intent=%v search_keywords=%q\n",
				tt.name, result.DeleteIntent, result.EditIntent, result.SearchKeywords)

			if result.DeleteIntent != tt.wantDelete {
				t.Errorf("delete_intent: got %v, want %v", result.DeleteIntent, tt.wantDelete)
			}
			if tt.wantDelete && !result.EditIntent {
				t.Error("expected edit_intent=true when delete_intent=true")
			}
			if tt.wantSearch && result.SearchKeywords == "" {
				t.Error("expected non-empty search_keywords")
			}
		})
	}
}
