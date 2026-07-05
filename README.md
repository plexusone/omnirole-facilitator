# Omnirole Facilitator

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omnirole-facilitator/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omnirole-facilitator/actions/workflows/go-ci.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omnirole-facilitator
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omnirole-facilitator
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omnirole-facilitator
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omnirole-facilitator
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omnirole-facilitator/blob/main/LICENSE

Omnichannel Facilitator role for [omniskill](https://github.com/plexusone/omniskill)-compatible systems.

The Facilitator role enables AI agents to facilitate collaboration across channels (meetings, chat, phone) by preparing context, tracking discussions, and creating documentation.

## Features

- **Omnichannel** - Works across meetings (OmniMeet), chat (Slack), and phone (Twilio)
- **Preparation** - Gather pre-reads, create agendas, brief participants
- **Real-time Facilitation** - Track action items, decisions, and discussions
- **Documentation** - Generate notes, publish to Confluence, create follow-ups
- **Integration** - Connect with Confluence, Aha, Jira, GitHub, GitLab

## Installation

```bash
go get github.com/plexusone/omnirole-facilitator
```

## Quick Start

```go
import (
    "context"
    "github.com/plexusone/omnirole-facilitator"
)

// Create the role
role := facilitator.New(facilitator.Config{
    DefaultConfluenceSpace: "TEAM",
    DefaultAhaProduct:      "PRODUCT-1",
    EnableTranscription:    true,
    EnableActionTracking:   true,
})

// Initialize with skills
err := role.Init(ctx, map[string]skill.Skill{
    "meeting":    meetingSkill,
    "google":     googleSkill,
    "confluence": confluenceSkill,
})
```

## Required Skills

| Skill | Purpose |
|-------|---------|
| `meeting` | Join and participate in meetings via OmniMeet |
| `google` | Access Google Docs, Sheets, and Slides |
| `confluence` | Publish meeting notes to Confluence |

## Optional Skills

| Skill | Purpose |
|-------|---------|
| `chat` | Slack, Discord, Teams via OmniChat |
| `voice` | Phone calls via Twilio/OmniVoice |
| `aha` | Create features and initiatives in Aha |
| `jira` | Create and link Jira issues |
| `github` | Reference GitHub PRs and issues |
| `gitlab` | Reference GitLab MRs and issues |

## Workflows

The role provides three main workflows:

### prepare

Gather pre-reads and create agenda before meetings.

```go
result, err := workflow.Execute(ctx, map[string]any{
    "meeting_id":   "meeting-123",
    "meeting_name": "Sprint Planning",
})
```

### facilitate

Track action items, decisions, and discussions during meetings.

```go
result, err := workflow.Execute(ctx, map[string]any{
    "meeting_id": "meeting-123",
})
```

### wrapup

Generate notes and publish artifacts after meeting ends.

```go
result, err := workflow.Execute(ctx, map[string]any{
    "meeting_id":       "meeting-123",
    "confluence_space": "TEAM",
})
```

## Role Specification

The role implements the enhanced role system with a comprehensive `RoleSpec`:

```go
spec := role.Spec()
// Returns:
// - ID: "facilitator"
// - Responsibilities: prepare, facilitate, document
// - Behaviors: pre-meeting-prep, during-meeting-notes, post-meeting-wrapup
// - Artifacts: meeting-notes, action-items, decisions
// - Metrics: action-capture-rate, notes-published-time, meetings-facilitated
```

### Responsibilities

| Phase | Responsibility |
|-------|----------------|
| Pre-meeting | Gather pre-reads, create agenda, brief participants |
| Meeting | Track actions, decisions, and discussions |
| Post-meeting | Generate notes, publish to Confluence, create follow-ups |

### Behaviors

| ID | Trigger | Context |
|----|---------|---------|
| `pre-meeting-prep` | 15 minutes before meeting | Always |
| `during-meeting-notes` | Meeting joined | Meeting |
| `post-meeting-wrapup` | Meeting ended | Always |

### Metrics

| Metric | Type | Target |
|--------|------|--------|
| `action-capture-rate` | Gauge | >= 95% |
| `notes-published-time` | Histogram | <= 1 hour |
| `meetings-facilitated` | Counter | - |

## Session Management

The role maintains session state during meetings:

```go
// Start a session
session := role.StartSession("meeting-123", "Sprint Planning")

// Track items during meeting
session.AddAction(session.Action{
    Description: "Update API documentation",
    Assignee:    "john@example.com",
})

session.AddDecision(session.Decision{
    Description: "Use gRPC for internal APIs",
    MadeBy:      "team",
})

// End session
session = role.EndSession()
```

## Configuration

```go
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
    MeetingNotesTemplate string
}
```

## Package Structure

```
omnirole-facilitator/
├── role.go           # FacilitatorRole implementation
├── role_test.go      # Role tests
├── session.go        # Session management wrapper
├── workflows.go      # Workflow definitions
├── workflows_test.go # Workflow tests
├── prompts/
│   ├── system.md         # Main system prompt
│   ├── pre_meeting.md    # Pre-meeting phase prompt
│   ├── during_meeting.md # During-meeting phase prompt
│   └── post_meeting.md   # Post-meeting phase prompt
└── session/
    ├── session.go        # Core session tracking
    └── session_test.go   # Session tests
```

## Documentation

- [Role Interface Reference](https://plexusone.dev/omniskill/role-interface/)
- [API Reference](https://pkg.go.dev/github.com/plexusone/omnirole-facilitator)

## License

MIT License - see [LICENSE](LICENSE) for details.
