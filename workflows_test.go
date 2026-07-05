package facilitator

import (
	"context"
	"testing"

	"github.com/plexusone/omniskill/role"
)

func TestPrepareWorkflow(t *testing.T) {
	r := New(Config{})
	ctx := context.Background()

	workflows := r.Workflows()
	var prepareWorkflow role.Workflow

	for _, w := range workflows {
		if w.Name() == "prepare" {
			prepareWorkflow = w
			break
		}
	}

	if prepareWorkflow == nil {
		t.Fatal("prepare workflow not found")
	}

	// Test with missing required fields
	result, err := prepareWorkflow.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Success {
		t.Error("expected failure with missing required fields")
	}

	// Test with valid input
	result, err = prepareWorkflow.Execute(ctx, map[string]any{
		"meeting_id":   "mtg-001",
		"meeting_name": "Sprint Planning",
		"agenda_items": []any{
			map[string]any{
				"title":       "Review Goals",
				"description": "Review sprint goals",
				"duration":    float64(15),
				"presenter":   "Alice",
			},
			map[string]any{
				"title":    "Backlog Grooming",
				"duration": float64(30),
			},
		},
		"prereads": []any{
			"https://docs.google.com/doc1",
			"https://confluence.example.com/page1",
		},
	})

	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !result.Success {
		t.Error("expected success with valid input")
	}

	// Verify session was created
	session := r.Session()
	if session == nil {
		t.Fatal("expected session to be created")
	}
	if len(session.Agenda) != 2 {
		t.Errorf("expected 2 agenda items, got %d", len(session.Agenda))
	}
	if len(session.PreReads) != 2 {
		t.Errorf("expected 2 pre-reads, got %d", len(session.PreReads))
	}
}

func TestFacilitateWorkflow(t *testing.T) {
	r := New(Config{})
	ctx := context.Background()

	workflows := r.Workflows()
	var facilitateWorkflow role.Workflow

	for _, w := range workflows {
		if w.Name() == "facilitate" {
			facilitateWorkflow = w
			break
		}
	}

	if facilitateWorkflow == nil {
		t.Fatal("facilitate workflow not found")
	}

	// Test without active session
	result, err := facilitateWorkflow.Execute(ctx, map[string]any{
		"meeting_id": "mtg-001",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Success {
		t.Error("expected failure without active session")
	}

	// Start a session first
	r.StartSession("mtg-001", "Test Meeting")

	// Test with mismatched meeting ID
	result, err = facilitateWorkflow.Execute(ctx, map[string]any{
		"meeting_id": "mtg-002",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Success {
		t.Error("expected failure with mismatched meeting ID")
	}

	// Test with correct meeting ID
	result, err = facilitateWorkflow.Execute(ctx, map[string]any{
		"meeting_id": "mtg-001",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !result.Success {
		t.Error("expected success with correct meeting ID")
	}

	// Verify phase changed
	if r.Session().Phase != PhaseMeeting {
		t.Errorf("expected phase %q, got %q", PhaseMeeting, r.Session().Phase)
	}
}

func TestWrapupWorkflow(t *testing.T) {
	r := New(Config{})
	ctx := context.Background()

	workflows := r.Workflows()
	var wrapupWorkflow role.Workflow

	for _, w := range workflows {
		if w.Name() == "wrapup" {
			wrapupWorkflow = w
			break
		}
	}

	if wrapupWorkflow == nil {
		t.Fatal("wrapup workflow not found")
	}

	// Test without active session
	result, err := wrapupWorkflow.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Success {
		t.Error("expected failure without active session")
	}

	// Start a session and add some data
	session := r.StartSession("mtg-001", "Retrospective")
	session.AddDecision(Decision{
		ID:          "dec-1",
		Description: "Use weekly sprints",
		MadeBy:      "Team",
	})
	session.AddAction(role.Action{
		ID:          "action-1",
		Description: "Update documentation",
		Assignee:    "Bob",
		DueDate:     "2025-01-20",
	})
	session.AddQuestion(Question{
		ID:       "q-1",
		Question: "What about testing?",
		AskedBy:  "Charlie",
	})

	// Execute wrapup
	result, err = wrapupWorkflow.Execute(ctx, map[string]any{
		"publish_to_confluence": true,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}

	// Session should be cleared after wrapup
	if r.Session() != nil {
		t.Error("expected session to be cleared after wrapup")
	}
}

func TestWorkflowMetadata(t *testing.T) {
	r := New(Config{})

	workflows := r.Workflows()
	if len(workflows) != 3 {
		t.Fatalf("expected 3 workflows, got %d", len(workflows))
	}

	expectedWorkflows := map[string]struct {
		trigger     string
		hasRequired bool
	}{
		"prepare":    {"manual", true},
		"facilitate": {"on_meeting_start", true},
		"wrapup":     {"on_meeting_end", false},
	}

	for _, w := range workflows {
		expected, ok := expectedWorkflows[w.Name()]
		if !ok {
			t.Errorf("unexpected workflow: %s", w.Name())
			continue
		}

		if w.Trigger() != expected.trigger {
			t.Errorf("workflow %s: expected trigger %q, got %q", w.Name(), expected.trigger, w.Trigger())
		}

		if w.Description() == "" {
			t.Errorf("workflow %s: expected non-empty description", w.Name())
		}

		schema := w.InputSchema()
		if schema == nil {
			t.Errorf("workflow %s: expected non-nil input schema", w.Name())
		}
	}
}
