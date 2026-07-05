package facilitator

import (
	"context"
	"fmt"
	"time"

	"github.com/plexusone/omniskill/role"
)

// prepareWorkflow creates the pre-meeting preparation workflow.
func (r *FacilitatorRole) prepareWorkflow() role.Workflow {
	return &role.BaseWorkflow{
		WorkflowName:        "prepare",
		WorkflowDescription: "Prepare for a meeting by gathering pre-reads, creating agenda, and briefing participants",
		WorkflowTrigger:     "manual",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meeting_id": map[string]any{
					"type":        "string",
					"description": "Meeting identifier",
				},
				"meeting_name": map[string]any{
					"type":        "string",
					"description": "Meeting name/title",
				},
				"agenda_items": map[string]any{
					"type":        "array",
					"description": "Agenda items for the meeting",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"title":       map[string]any{"type": "string"},
							"description": map[string]any{"type": "string"},
							"duration":    map[string]any{"type": "integer", "description": "Duration in minutes"},
							"presenter":   map[string]any{"type": "string"},
						},
					},
				},
				"prereads": map[string]any{
					"type":        "array",
					"description": "Document URLs or IDs to gather as pre-reads",
					"items":       map[string]any{"type": "string"},
				},
				"confluence_space": map[string]any{
					"type":        "string",
					"description": "Confluence space for meeting notes",
				},
			},
			"required": []string{"meeting_id", "meeting_name"},
		},
		ExecuteFunc: r.executePrepare,
	}
}

// executePrepare runs the preparation workflow.
func (r *FacilitatorRole) executePrepare(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
	meetingID, _ := input["meeting_id"].(string)
	meetingName, _ := input["meeting_name"].(string)

	if meetingID == "" || meetingName == "" {
		return role.WorkflowResult{
			Success: false,
			Error:   "meeting_id and meeting_name are required",
		}, nil
	}

	// Start a new session
	session := r.StartSession(meetingID, meetingName)

	// Process agenda items
	if agendaItems, ok := input["agenda_items"].([]any); ok {
		for i, item := range agendaItems {
			if itemMap, ok := item.(map[string]any); ok {
				agendaItem := AgendaItem{
					ID:          fmt.Sprintf("agenda-%d", i+1),
					Title:       getString(itemMap, "title"),
					Description: getString(itemMap, "description"),
					Presenter:   getString(itemMap, "presenter"),
					Status:      "pending",
				}
				if duration, ok := itemMap["duration"].(float64); ok {
					agendaItem.Duration = time.Duration(duration) * time.Minute
				}
				session.AddAgendaItem(agendaItem)
			}
		}
	}

	// Process pre-reads (would use Google/Confluence skills to fetch)
	if prereads, ok := input["prereads"].([]any); ok {
		for i, preread := range prereads {
			if url, ok := preread.(string); ok {
				doc := Document{
					ID:    fmt.Sprintf("preread-%d", i+1),
					Title: fmt.Sprintf("Pre-read %d", i+1),
					Type:  "url",
					URL:   url,
					// In real implementation: fetch and summarize using Google/Confluence skill
				}
				session.AddPreRead(doc)
			}
		}
	}

	return role.WorkflowResult{
		Success: true,
		Message: fmt.Sprintf("Meeting '%s' prepared with %d agenda items and %d pre-reads",
			meetingName, len(session.Agenda), len(session.PreReads)),
		Output: map[string]any{
			"meeting_id":    meetingID,
			"meeting_name":  meetingName,
			"agenda_count":  len(session.Agenda),
			"preread_count": len(session.PreReads),
		},
	}, nil
}

// facilitateWorkflow creates the meeting facilitation workflow.
func (r *FacilitatorRole) facilitateWorkflow() role.Workflow {
	return &role.BaseWorkflow{
		WorkflowName:        "facilitate",
		WorkflowDescription: "Start facilitating a meeting - join, track actions, decisions, and discussions",
		WorkflowTrigger:     "on_meeting_start",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"meeting_id": map[string]any{
					"type":        "string",
					"description": "Meeting identifier to join",
				},
				"enable_transcription": map[string]any{
					"type":        "boolean",
					"description": "Enable real-time transcription",
				},
				"enable_action_tracking": map[string]any{
					"type":        "boolean",
					"description": "Enable automatic action item detection",
				},
			},
			"required": []string{"meeting_id"},
		},
		ExecuteFunc: r.executeFacilitate,
	}
}

// executeFacilitate runs the facilitation workflow.
func (r *FacilitatorRole) executeFacilitate(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
	meetingID, _ := input["meeting_id"].(string)

	session := r.Session()
	if session == nil {
		return role.WorkflowResult{
			Success: false,
			Error:   "no active session - run prepare workflow first",
		}, nil
	}

	if session.MeetingID != meetingID {
		return role.WorkflowResult{
			Success: false,
			Error:   fmt.Sprintf("session mismatch: expected %s, got %s", session.MeetingID, meetingID),
		}, nil
	}

	// Transition to meeting phase
	session.SetPhase(PhaseMeeting)

	// In real implementation:
	// 1. Use meeting skill to join the meeting
	// 2. Set up transcription if enabled
	// 3. Start listening for action items, decisions

	return role.WorkflowResult{
		Success: true,
		Message: fmt.Sprintf("Now facilitating meeting '%s'", session.MeetingName),
		Output: map[string]any{
			"meeting_id":   meetingID,
			"meeting_name": session.MeetingName,
			"phase":        string(session.Phase),
		},
	}, nil
}

// wrapupWorkflow creates the post-meeting wrap-up workflow.
func (r *FacilitatorRole) wrapupWorkflow() role.Workflow {
	return &role.BaseWorkflow{
		WorkflowName:        "wrapup",
		WorkflowDescription: "End the meeting and generate artifacts - notes, action items, follow-ups",
		WorkflowTrigger:     "on_meeting_end",
		WorkflowInputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"publish_to_confluence": map[string]any{
					"type":        "boolean",
					"description": "Publish meeting notes to Confluence",
				},
				"create_aha_features": map[string]any{
					"type":        "boolean",
					"description": "Create Aha features for action items",
				},
				"create_jira_issues": map[string]any{
					"type":        "boolean",
					"description": "Create Jira issues for action items",
				},
				"send_summary_email": map[string]any{
					"type":        "boolean",
					"description": "Send summary email to participants",
				},
			},
		},
		ExecuteFunc: r.executeWrapup,
	}
}

// executeWrapup runs the wrap-up workflow.
func (r *FacilitatorRole) executeWrapup(ctx context.Context, input map[string]any) (role.WorkflowResult, error) {
	session := r.Session()
	if session == nil {
		return role.WorkflowResult{
			Success: false,
			Error:   "no active session",
		}, nil
	}

	// End the session
	session.End()

	// Generate meeting notes artifact
	notes := r.generateMeetingNotes(session)
	notesArtifact := role.Artifact{
		Name:    "meeting-notes",
		Type:    "document",
		Format:  "markdown",
		Content: notes,
	}
	session.AddArtifact(notesArtifact)

	// Generate action items summary
	actionSummary := r.generateActionSummary(session)
	actionArtifact := role.Artifact{
		Name:    "action-items",
		Type:    "list",
		Format:  "markdown",
		Content: actionSummary,
	}
	session.AddArtifact(actionArtifact)

	// In real implementation:
	// 1. Publish to Confluence if requested
	// 2. Create Aha features if requested
	// 3. Create Jira issues if requested
	// 4. Send email if requested

	result := role.WorkflowResult{
		Success:   true,
		Message:   fmt.Sprintf("Meeting '%s' wrapped up - %d actions, %d decisions", session.MeetingName, len(session.Actions), len(session.Decisions)),
		Artifacts: session.Artifacts,
		Actions:   session.Actions,
		Output: map[string]any{
			"meeting_id":     session.MeetingID,
			"meeting_name":   session.MeetingName,
			"duration":       session.Duration().String(),
			"action_count":   len(session.Actions),
			"decision_count": len(session.Decisions),
			"question_count": len(session.Questions),
		},
	}

	// Clear the session
	r.EndSession()

	return result, nil
}

// generateMeetingNotes creates markdown meeting notes.
func (r *FacilitatorRole) generateMeetingNotes(session *MeetingSession) string {
	var notes string

	notes += fmt.Sprintf("# %s\n\n", session.MeetingName)
	notes += fmt.Sprintf("**Date:** %s\n", session.StartTime.Format("January 2, 2006"))
	notes += fmt.Sprintf("**Duration:** %s\n\n", session.Duration().Round(time.Minute))

	// Participants
	if len(session.Participants) > 0 {
		notes += "## Participants\n\n"
		for _, p := range session.Participants {
			notes += fmt.Sprintf("- %s", p.Name)
			if p.Role != "" {
				notes += fmt.Sprintf(" (%s)", p.Role)
			}
			notes += "\n"
		}
		notes += "\n"
	}

	// Agenda & Notes
	if len(session.Agenda) > 0 {
		notes += "## Agenda\n\n"
		for _, item := range session.Agenda {
			notes += fmt.Sprintf("### %s\n", item.Title)
			if item.Description != "" {
				notes += fmt.Sprintf("%s\n", item.Description)
			}
			if item.Notes != "" {
				notes += fmt.Sprintf("\n%s\n", item.Notes)
			}
			notes += "\n"
		}
	}

	// Decisions
	if len(session.Decisions) > 0 {
		notes += "## Decisions\n\n"
		for _, d := range session.Decisions {
			notes += fmt.Sprintf("- **%s**", d.Description)
			if d.MadeBy != "" {
				notes += fmt.Sprintf(" (by %s)", d.MadeBy)
			}
			notes += "\n"
		}
		notes += "\n"
	}

	// Action Items
	if len(session.Actions) > 0 {
		notes += "## Action Items\n\n"
		for _, a := range session.Actions {
			notes += fmt.Sprintf("- [ ] %s", a.Description)
			if a.Assignee != "" {
				notes += fmt.Sprintf(" - %s", a.Assignee)
			}
			if a.DueDate != "" {
				notes += fmt.Sprintf(" (due: %s)", a.DueDate)
			}
			notes += "\n"
		}
		notes += "\n"
	}

	// Open Questions
	openQuestions := session.OpenQuestions()
	if len(openQuestions) > 0 {
		notes += "## Open Questions\n\n"
		for _, q := range openQuestions {
			notes += fmt.Sprintf("- %s", q.Question)
			if q.AskedBy != "" {
				notes += fmt.Sprintf(" (asked by %s)", q.AskedBy)
			}
			notes += "\n"
		}
		notes += "\n"
	}

	return notes
}

// generateActionSummary creates a focused action items summary.
func (r *FacilitatorRole) generateActionSummary(session *MeetingSession) string {
	var summary string

	summary += fmt.Sprintf("# Action Items - %s\n\n", session.MeetingName)
	summary += fmt.Sprintf("Generated: %s\n\n", time.Now().Format("January 2, 2006 3:04 PM"))

	if len(session.Actions) == 0 {
		summary += "No action items recorded.\n"
		return summary
	}

	// Group by assignee
	byAssignee := make(map[string][]role.Action)
	for _, a := range session.Actions {
		assignee := a.Assignee
		if assignee == "" {
			assignee = "Unassigned"
		}
		byAssignee[assignee] = append(byAssignee[assignee], a)
	}

	for assignee, actions := range byAssignee {
		summary += fmt.Sprintf("## %s\n\n", assignee)
		for _, a := range actions {
			summary += fmt.Sprintf("- [ ] %s", a.Description)
			if a.DueDate != "" {
				summary += fmt.Sprintf(" (due: %s)", a.DueDate)
			}
			if a.Priority != "" {
				summary += fmt.Sprintf(" [%s]", a.Priority)
			}
			summary += "\n"
		}
		summary += "\n"
	}

	return summary
}

// Helper function to safely get string from map.
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
