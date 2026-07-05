package session

import (
	"testing"
	"time"

	"github.com/plexusone/omniskill/role"
)

func TestNew(t *testing.T) {
	s := New("mtg-001", "Weekly Standup")

	if s.MeetingID != "mtg-001" {
		t.Errorf("expected meeting ID 'mtg-001', got %q", s.MeetingID)
	}
	if s.MeetingName != "Weekly Standup" {
		t.Errorf("expected meeting name 'Weekly Standup', got %q", s.MeetingName)
	}
	if s.Phase != PhasePreMeeting {
		t.Errorf("expected phase %q, got %q", PhasePreMeeting, s.Phase)
	}
	if s.StartTime.IsZero() {
		t.Error("expected non-zero start time")
	}
}

func TestAddAgendaItem(t *testing.T) {
	s := New("mtg-001", "Planning")

	s.AddAgendaItem(AgendaItem{
		ID:          "item-1",
		Title:       "Sprint Goals",
		Description: "Define goals for the sprint",
		Duration:    15 * time.Minute,
		Presenter:   "John",
	})

	if len(s.Agenda) != 1 {
		t.Fatalf("expected 1 agenda item, got %d", len(s.Agenda))
	}

	item := s.Agenda[0]
	if item.Title != "Sprint Goals" {
		t.Errorf("expected title 'Sprint Goals', got %q", item.Title)
	}
	if item.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", item.Status)
	}
}

func TestAddPreRead(t *testing.T) {
	s := New("mtg-001", "Review")

	s.AddPreRead(Document{
		ID:      "doc-1",
		Title:   "Design Doc",
		Type:    "google_doc",
		URL:     "https://docs.google.com/...",
		Summary: "System architecture overview",
	})

	if len(s.PreReads) != 1 {
		t.Fatalf("expected 1 pre-read, got %d", len(s.PreReads))
	}
	if s.PreReads[0].Title != "Design Doc" {
		t.Errorf("expected title 'Design Doc', got %q", s.PreReads[0].Title)
	}
}

func TestAddParticipant(t *testing.T) {
	s := New("mtg-001", "Team Meeting")

	s.AddParticipant(Participant{
		ID:    "user-1",
		Name:  "Alice",
		Email: "alice@example.com",
		Role:  "host",
	})

	if len(s.Participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(s.Participants))
	}
	if s.Participants[0].Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", s.Participants[0].Name)
	}
}

func TestAddAction(t *testing.T) {
	s := New("mtg-001", "Planning")

	s.AddAction(role.Action{
		ID:          "action-1",
		Description: "Update the roadmap",
		Assignee:    "Bob",
		DueDate:     "2025-01-20",
		Priority:    "high",
	})

	if len(s.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(s.Actions))
	}
	if s.Actions[0].Assignee != "Bob" {
		t.Errorf("expected assignee 'Bob', got %q", s.Actions[0].Assignee)
	}
}

func TestAddDecision(t *testing.T) {
	s := New("mtg-001", "Review")

	s.AddDecision(Decision{
		ID:          "dec-1",
		Description: "Use PostgreSQL for the database",
		MadeBy:      "Team",
		Timestamp:   time.Now(),
	})

	if len(s.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(s.Decisions))
	}
	if s.Decisions[0].Description != "Use PostgreSQL for the database" {
		t.Errorf("unexpected decision description")
	}
}

func TestAddQuestion(t *testing.T) {
	s := New("mtg-001", "Q&A")

	s.AddQuestion(Question{
		ID:       "q-1",
		Question: "What's the timeline?",
		AskedBy:  "Charlie",
	})

	if len(s.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(s.Questions))
	}
	if s.Questions[0].Answered {
		t.Error("expected question to be unanswered")
	}
}

func TestOpenQuestions(t *testing.T) {
	s := New("mtg-001", "Q&A")

	s.AddQuestion(Question{ID: "q-1", Question: "Open question", Answered: false})
	s.AddQuestion(Question{ID: "q-2", Question: "Answered question", Answered: true})
	s.AddQuestion(Question{ID: "q-3", Question: "Another open", Answered: false})

	open := s.OpenQuestions()
	if len(open) != 2 {
		t.Errorf("expected 2 open questions, got %d", len(open))
	}
}

func TestOpenActions(t *testing.T) {
	s := New("mtg-001", "Planning")

	// Action without links (open)
	s.AddAction(role.Action{ID: "a-1", Description: "Open action"})

	// Action with links (closed)
	s.AddAction(role.Action{
		ID:          "a-2",
		Description: "Linked action",
		Links: []role.ActionLink{
			{System: "jira", ID: "PROJ-123"},
		},
	})

	open := s.OpenActions()
	if len(open) != 1 {
		t.Errorf("expected 1 open action, got %d", len(open))
	}
	if open[0].ID != "a-1" {
		t.Errorf("expected action ID 'a-1', got %q", open[0].ID)
	}
}

func TestSetPhase(t *testing.T) {
	s := New("mtg-001", "Meeting")

	if s.Phase != PhasePreMeeting {
		t.Errorf("expected initial phase %q", PhasePreMeeting)
	}

	s.SetPhase(PhaseMeeting)
	if s.Phase != PhaseMeeting {
		t.Errorf("expected phase %q, got %q", PhaseMeeting, s.Phase)
	}

	s.SetPhase(PhasePostMeeting)
	if s.Phase != PhasePostMeeting {
		t.Errorf("expected phase %q, got %q", PhasePostMeeting, s.Phase)
	}
}

func TestEnd(t *testing.T) {
	s := New("mtg-001", "Meeting")

	if !s.EndTime.IsZero() {
		t.Error("expected zero end time before End()")
	}

	s.End()

	if s.EndTime.IsZero() {
		t.Error("expected non-zero end time after End()")
	}
	if s.Phase != PhasePostMeeting {
		t.Errorf("expected phase %q after End(), got %q", PhasePostMeeting, s.Phase)
	}
}

func TestDuration(t *testing.T) {
	s := New("mtg-001", "Meeting")

	// Before end, duration is time since start
	d1 := s.Duration()
	if d1 < 0 {
		t.Error("duration should be non-negative")
	}

	// Sleep briefly and check duration increases
	time.Sleep(10 * time.Millisecond)
	d2 := s.Duration()
	if d2 <= d1 {
		t.Error("duration should increase over time")
	}

	// After end, duration is fixed
	s.End()
	d3 := s.Duration()
	time.Sleep(10 * time.Millisecond)
	d4 := s.Duration()
	if d3 != d4 {
		t.Error("duration should be fixed after End()")
	}
}

func TestAddTranscript(t *testing.T) {
	s := New("mtg-001", "Recorded Meeting")

	s.AddTranscript(TranscriptSegment{
		Speaker:   "Alice",
		Text:      "Hello everyone",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Second),
		IsFinal:   true,
	})

	if len(s.Transcript) != 1 {
		t.Fatalf("expected 1 transcript segment, got %d", len(s.Transcript))
	}
	if s.Transcript[0].Speaker != "Alice" {
		t.Errorf("expected speaker 'Alice', got %q", s.Transcript[0].Speaker)
	}
}

func TestAddArtifact(t *testing.T) {
	s := New("mtg-001", "Meeting")

	s.AddArtifact(role.Artifact{
		Name:    "meeting-notes",
		Type:    "document",
		Format:  "markdown",
		Content: "# Meeting Notes\n\n...",
	})

	if len(s.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(s.Artifacts))
	}
	if s.Artifacts[0].Name != "meeting-notes" {
		t.Errorf("expected artifact name 'meeting-notes', got %q", s.Artifacts[0].Name)
	}
}

func TestAddNote(t *testing.T) {
	s := New("mtg-001", "Discussion")

	s.AddNote(Note{
		ID:        "note-1",
		Content:   "Important discussion point",
		Speaker:   "Bob",
		Timestamp: time.Now(),
		Tags:      []string{"important", "follow-up"},
	})

	if len(s.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(s.Notes))
	}
	if s.Notes[0].Content != "Important discussion point" {
		t.Errorf("unexpected note content")
	}
}

func TestClose(t *testing.T) {
	s := New("mtg-001", "Meeting")

	// Close should not error
	if err := s.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
