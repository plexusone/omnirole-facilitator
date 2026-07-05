// Package facilitator provides an omnichannel Facilitator role for omniskill-compatible systems.
//
// The Facilitator role enables AI agents to facilitate collaboration across channels:
//   - Prepares meetings (gathers pre-reads, creates agendas)
//   - Facilitates discussions (tracks actions, decisions, questions)
//   - Creates artifacts (meeting notes, summaries, action items)
//   - Integrates with external systems (Confluence, Aha, Jira, GitHub)
//
// # Required Skills
//
// The Meeting PM role requires the following skills:
//   - meeting: OmniMeet skill for joining/participating in meetings
//   - google: Google Docs/Sheets/Slides for document access
//   - confluence: Confluence for publishing meeting notes
//   - aha: Aha for product management integration (optional)
//
// # Workflows
//
// The role provides three main workflows:
//   - prepare: Gather pre-reads, create agenda, brief participants
//   - facilitate: Track action items, decisions, and discussions
//   - wrapup: Generate notes, publish to Confluence, create follow-ups
//
// # Usage
//
//	role := facilitator.New(facilitator.Config{
//	    DefaultConfluenceSpace: "TEAM",
//	    DefaultAhaProduct:      "PRODUCT-1",
//	})
//
//	// Initialize with required skills
//	err := role.Init(ctx, map[string]skill.Skill{
//	    "meeting":    meetingSkill,
//	    "google":     googleSkill,
//	    "confluence": confluenceSkill,
//	    "aha":        ahaSkill,
//	})
package facilitator

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/plexusone/omniskill/role"
	"github.com/plexusone/omniskill/skill"
)

//go:embed prompts/system.md
var systemPrompt string

//go:embed prompts/pre_meeting.md
var preMeetingPrompt string

//go:embed prompts/during_meeting.md
var duringMeetingPrompt string

//go:embed prompts/post_meeting.md
var postMeetingPrompt string

// Config configures the Meeting PM role.
type Config struct {
	// DefaultConfluenceSpace is the Confluence space for meeting notes.
	DefaultConfluenceSpace string

	// DefaultAhaProduct is the Aha product for feature/initiative creation.
	DefaultAhaProduct string

	// EnableTranscription enables real-time transcription.
	EnableTranscription bool

	// EnableActionTracking enables automatic action item detection.
	EnableActionTracking bool

	// MeetingNotesTemplate is a custom template for meeting notes.
	// If empty, uses the default template.
	MeetingNotesTemplate string
}

// FacilitatorRole implements the Meeting Program Manager role.
type FacilitatorRole struct {
	config    Config
	skills    map[string]skill.Skill
	workflows []role.Workflow
	session   *MeetingSession
}

// New creates a new Meeting PM role with the given configuration.
func New(cfg Config) *FacilitatorRole {
	r := &FacilitatorRole{
		config: cfg,
	}

	// Initialize workflows
	r.workflows = []role.Workflow{
		r.prepareWorkflow(),
		r.facilitateWorkflow(),
		r.wrapupWorkflow(),
	}

	return r
}

// Name returns the role identifier.
func (r *FacilitatorRole) Name() string {
	return "facilitator"
}

// Description returns a human-readable description.
func (r *FacilitatorRole) Description() string {
	return "Omnichannel Facilitator - facilitates collaboration across meetings, chat, and phone"
}

// SystemPrompt returns the role's system prompt.
func (r *FacilitatorRole) SystemPrompt(ctx context.Context) (string, error) {
	// Build dynamic prompt based on configuration
	prompt := systemPrompt

	// Add configuration context
	var configContext strings.Builder
	configContext.WriteString("\n\n## Current Configuration\n\n")

	if r.config.DefaultConfluenceSpace != "" {
		fmt.Fprintf(&configContext, "- Confluence Space: %s\n", r.config.DefaultConfluenceSpace)
	}
	if r.config.DefaultAhaProduct != "" {
		fmt.Fprintf(&configContext, "- Aha Product: %s\n", r.config.DefaultAhaProduct)
	}
	if r.config.EnableTranscription {
		configContext.WriteString("- Real-time transcription: ENABLED\n")
	}
	if r.config.EnableActionTracking {
		configContext.WriteString("- Automatic action tracking: ENABLED\n")
	}

	return prompt + configContext.String(), nil
}

// RequiredSkills returns the skills this role needs.
func (r *FacilitatorRole) RequiredSkills() []string {
	return []string{
		"meeting",    // OmniMeet for meeting participation
		"google",     // Google Docs/Sheets/Slides
		"confluence", // Confluence for notes
	}
}

// OptionalSkills returns skills that enhance the role but aren't required.
func (r *FacilitatorRole) OptionalSkills() []string {
	return []string{
		"aha",    // Aha for product management
		"jira",   // Jira for issue tracking
		"github", // GitHub for PR/issue references
		"gitlab", // GitLab for MR/issue references
	}
}

// Init initializes the role with its required skills.
func (r *FacilitatorRole) Init(ctx context.Context, skills map[string]skill.Skill) error {
	// Validate required skills
	for _, name := range r.RequiredSkills() {
		if _, ok := skills[name]; !ok {
			return fmt.Errorf("required skill not provided: %s", name)
		}
	}

	r.skills = skills
	return nil
}

// Close releases any resources held by the role.
func (r *FacilitatorRole) Close() error {
	if r.session != nil {
		return r.session.Close()
	}
	return nil
}

// Workflows returns the role's workflows.
func (r *FacilitatorRole) Workflows() []role.Workflow {
	return r.workflows
}

// GetSkill returns a skill by name, or nil if not available.
func (r *FacilitatorRole) GetSkill(name string) skill.Skill {
	if r.skills == nil {
		return nil
	}
	return r.skills[name]
}

// Session returns the current meeting session, if any.
func (r *FacilitatorRole) Session() *MeetingSession {
	return r.session
}

// StartSession begins a new meeting session.
func (r *FacilitatorRole) StartSession(meetingID, meetingName string) *MeetingSession {
	r.session = NewMeetingSession(meetingID, meetingName)
	return r.session
}

// EndSession ends the current meeting session.
func (r *FacilitatorRole) EndSession() *MeetingSession {
	session := r.session
	r.session = nil
	return session
}

// PhasePrompt returns the prompt for a specific meeting phase.
func (r *FacilitatorRole) PhasePrompt(phase Phase) string {
	switch phase {
	case PhasePreMeeting:
		return preMeetingPrompt
	case PhaseMeeting:
		return duringMeetingPrompt
	case PhasePostMeeting:
		return postMeetingPrompt
	default:
		return ""
	}
}

// Spec returns the complete role specification for Facilitator.
func (r *FacilitatorRole) Spec() *role.RoleSpec {
	return &role.RoleSpec{
		ID:          "facilitator",
		Name:        "Omnichannel Facilitator",
		Description: "Facilitates collaboration across meetings, chat, and phone",
		Version:     "1.0.0",
		Purpose:     "Transform meetings into actionable outcomes through structured facilitation, real-time tracking, and comprehensive documentation",
		Goals: []string{
			"Ensure all meetings have clear agendas and pre-reads",
			"Capture all action items, decisions, and key discussions",
			"Publish meeting notes within 1 hour of meeting end",
			"Track and follow up on action items",
		},
		Responsibilities: []role.Responsibility{
			{
				ID:          "prepare",
				Name:        "Meeting Preparation",
				Description: "Gather pre-reads, create agenda, brief participants before meetings",
				Phase:       "pre-meeting",
				Priority:    "high",
			},
			{
				ID:          "facilitate",
				Name:        "Meeting Facilitation",
				Description: "Track action items, decisions, and discussions during meetings",
				Phase:       "meeting",
				Priority:    "high",
			},
			{
				ID:          "document",
				Name:        "Documentation",
				Description: "Generate and publish meeting notes, summaries, and follow-ups",
				Phase:       "post-meeting",
				Priority:    "high",
			},
		},
		Skills: role.SkillRequirements{
			Required: []role.SkillRef{
				{Name: "meeting", Purpose: "Join and participate in meetings via OmniMeet"},
				{Name: "google", Purpose: "Access Google Docs, Sheets, and Slides"},
				{Name: "confluence", Purpose: "Publish meeting notes to Confluence"},
			},
			Optional: []role.SkillRef{
				{Name: "aha", Purpose: "Create features and initiatives in Aha"},
				{Name: "jira", Purpose: "Create and link Jira issues"},
				{Name: "github", Purpose: "Reference GitHub PRs and issues"},
				{Name: "gitlab", Purpose: "Reference GitLab MRs and issues"},
			},
		},
		Behaviors: []role.Behavior{
			{
				ID:          "pre-meeting-prep",
				Name:        "Pre-meeting Preparation",
				Description: "Automatically gather pre-reads and prepare agenda before meetings",
				Context:     role.BehaviorContextAlways,
				Trigger: role.BehaviorTrigger{
					Type:     role.TriggerTypeSchedule,
					Schedule: "15 minutes before meeting",
				},
				Actions: []role.BehaviorAction{
					{ID: "gather-prereads", Type: role.ActionTypeWorkflow, Workflow: "prepare"},
				},
				Enabled:  true,
				Priority: 100,
			},
			{
				ID:          "during-meeting-notes",
				Name:        "Real-time Note Taking",
				Description: "Track actions, decisions, and questions during the meeting",
				Context:     role.BehaviorContextMeeting,
				Trigger: role.BehaviorTrigger{
					Type:  role.TriggerTypeEvent,
					Event: role.EventMeetingJoined,
				},
				Actions: []role.BehaviorAction{
					{ID: "start-tracking", Type: role.ActionTypeWorkflow, Workflow: "facilitate"},
				},
				Enabled:  true,
				Priority: 100,
			},
			{
				ID:          "post-meeting-wrapup",
				Name:        "Post-meeting Wrap-up",
				Description: "Generate notes and publish artifacts after meeting ends",
				Context:     role.BehaviorContextAlways,
				Trigger: role.BehaviorTrigger{
					Type:  role.TriggerTypeEvent,
					Event: role.EventMeetingEnd,
				},
				Actions: []role.BehaviorAction{
					{ID: "generate-notes", Type: role.ActionTypeWorkflow, Workflow: "wrapup"},
				},
				Enabled:  true,
				Priority: 100,
			},
		},
		Artifacts: []role.ArtifactSpec{
			{
				ID:          "meeting-notes",
				Name:        "Meeting Notes",
				Description: "Comprehensive meeting notes with agenda, discussions, and outcomes",
				Type:        "document",
				Format:      "markdown",
				Required:    true,
				Trigger:     "post-meeting",
			},
			{
				ID:          "action-items",
				Name:        "Action Items",
				Description: "List of action items with assignees and due dates",
				Type:        "list",
				Format:      "markdown",
				Required:    true,
				Trigger:     "post-meeting",
			},
			{
				ID:          "decisions",
				Name:        "Decision Log",
				Description: "Record of decisions made during the meeting",
				Type:        "list",
				Format:      "markdown",
				Required:    false,
				Trigger:     "post-meeting",
			},
		},
		Metrics: []role.MetricDefinition{
			{
				ID:          "action-capture-rate",
				Name:        "Action Capture Rate",
				Description: "Percentage of mentioned action items that are captured",
				Type:        role.MetricTypeGauge,
				Unit:        role.UnitPercent,
				Target: &role.MetricTarget{
					Value:    95,
					Operator: role.OperatorGreaterThanOrEqual,
				},
			},
			{
				ID:          "notes-published-time",
				Name:        "Notes Publication Time",
				Description: "Time from meeting end to notes publication",
				Type:        role.MetricTypeHistogram,
				Unit:        role.UnitSeconds,
				Target: &role.MetricTarget{
					Value:    3600, // 1 hour in seconds
					Operator: role.OperatorLessThanOrEqual,
				},
				Buckets: []float64{300, 600, 1800, 3600, 7200}, // 5m, 10m, 30m, 1h, 2h
			},
			{
				ID:          "meetings-facilitated",
				Name:        "Meetings Facilitated",
				Description: "Total number of meetings facilitated",
				Type:        role.MetricTypeCounter,
				Unit:        role.UnitCount,
			},
		},
		Persona: &role.PersonaSpec{
			Tone:      "professional",
			Formality: "business-casual",
			Traits:    []string{"organized", "attentive", "concise", "action-oriented"},
		},
		Metadata: map[string]any{
			"confluence_space": r.config.DefaultConfluenceSpace,
			"aha_product":      r.config.DefaultAhaProduct,
		},
	}
}

// Behaviors returns the behaviors defined for this role.
func (r *FacilitatorRole) Behaviors() []role.Behavior {
	spec := r.Spec()
	return spec.Behaviors
}

// Metrics returns the metric definitions for this role.
func (r *FacilitatorRole) Metrics() []role.MetricDefinition {
	spec := r.Spec()
	return spec.Metrics
}

// Ensure FacilitatorRole implements Role and optional interfaces.
var _ role.Role = (*FacilitatorRole)(nil)
var _ role.SkillRequirer = (*FacilitatorRole)(nil)
var _ role.BehaviorProvider = (*FacilitatorRole)(nil)
var _ role.MetricsProvider = (*FacilitatorRole)(nil)
