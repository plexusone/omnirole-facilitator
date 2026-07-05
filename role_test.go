package facilitator

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := Config{
		DefaultConfluenceSpace: "TEAM",
		DefaultAhaProduct:      "PROD-1",
		EnableTranscription:    true,
		EnableActionTracking:   true,
	}

	role := New(cfg)

	if role.Name() != "facilitator" {
		t.Errorf("expected name 'facilitator', got %q", role.Name())
	}

	if role.Description() == "" {
		t.Error("expected non-empty description")
	}

	// Check workflows are initialized
	workflows := role.Workflows()
	if len(workflows) != 3 {
		t.Errorf("expected 3 workflows, got %d", len(workflows))
	}

	// Verify workflow names
	names := make(map[string]bool)
	for _, w := range workflows {
		names[w.Name()] = true
	}
	for _, expected := range []string{"prepare", "facilitate", "wrapup"} {
		if !names[expected] {
			t.Errorf("missing workflow %q", expected)
		}
	}
}

func TestSystemPrompt(t *testing.T) {
	cfg := Config{
		DefaultConfluenceSpace: "ENGINEERING",
		DefaultAhaProduct:      "ROADMAP",
	}

	role := New(cfg)
	ctx := context.Background()

	prompt, err := role.SystemPrompt(ctx)
	if err != nil {
		t.Fatalf("SystemPrompt failed: %v", err)
	}

	// Should contain the embedded system prompt
	if prompt == "" {
		t.Error("expected non-empty system prompt")
	}

	// Should contain configuration
	if !contains(prompt, "ENGINEERING") {
		t.Error("expected prompt to contain Confluence space")
	}
	if !contains(prompt, "ROADMAP") {
		t.Error("expected prompt to contain Aha product")
	}
}

func TestPhasePrompt(t *testing.T) {
	role := New(Config{})

	tests := []struct {
		phase    Phase
		notEmpty bool
	}{
		{PhasePreMeeting, true},
		{PhaseMeeting, true},
		{PhasePostMeeting, true},
		{"unknown", false},
	}

	for _, tt := range tests {
		prompt := role.PhasePrompt(tt.phase)
		if tt.notEmpty && prompt == "" {
			t.Errorf("expected non-empty prompt for phase %q", tt.phase)
		}
		if !tt.notEmpty && prompt != "" {
			t.Errorf("expected empty prompt for phase %q", tt.phase)
		}
	}
}

func TestRequiredSkills(t *testing.T) {
	role := New(Config{})

	required := role.RequiredSkills()
	expected := []string{"meeting", "google", "confluence"}

	if len(required) != len(expected) {
		t.Errorf("expected %d required skills, got %d", len(expected), len(required))
	}

	for _, skill := range expected {
		found := false
		for _, r := range required {
			if r == skill {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing required skill %q", skill)
		}
	}
}

func TestOptionalSkills(t *testing.T) {
	role := New(Config{})

	optional := role.OptionalSkills()
	if len(optional) == 0 {
		t.Error("expected at least one optional skill")
	}

	// Should include aha, jira, etc.
	hasAha := false
	for _, s := range optional {
		if s == "aha" {
			hasAha = true
			break
		}
	}
	if !hasAha {
		t.Error("expected 'aha' in optional skills")
	}
}

func TestSessionLifecycle(t *testing.T) {
	role := New(Config{})

	// Initially no session
	if role.Session() != nil {
		t.Error("expected no initial session")
	}

	// Start session
	session := role.StartSession("mtg-123", "Sprint Planning")
	if session == nil {
		t.Fatal("StartSession returned nil")
	}

	if session.MeetingID != "mtg-123" {
		t.Errorf("expected meeting ID 'mtg-123', got %q", session.MeetingID)
	}
	if session.MeetingName != "Sprint Planning" {
		t.Errorf("expected meeting name 'Sprint Planning', got %q", session.MeetingName)
	}

	// Session should be accessible
	if role.Session() != session {
		t.Error("Session() should return the active session")
	}

	// End session
	ended := role.EndSession()
	if ended != session {
		t.Error("EndSession should return the ended session")
	}

	// Session should be cleared
	if role.Session() != nil {
		t.Error("expected no session after EndSession")
	}
}

func TestClose(t *testing.T) {
	role := New(Config{})

	// Close without session should not error
	if err := role.Close(); err != nil {
		t.Errorf("Close without session failed: %v", err)
	}

	// Close with session should not error
	role.StartSession("mtg-456", "Retrospective")
	if err := role.Close(); err != nil {
		t.Errorf("Close with session failed: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
