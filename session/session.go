// Package session manages meeting session state for the Meeting PM role.
package session

import (
	"sync"
	"time"

	"github.com/plexusone/omniskill/role"
)

// Phase represents the current phase of a meeting.
type Phase string

const (
	PhasePreMeeting  Phase = "pre_meeting"
	PhaseMeeting     Phase = "meeting"
	PhasePostMeeting Phase = "post_meeting"
)

// Session tracks the state of an active meeting.
type Session struct {
	// Meeting identification
	MeetingID   string    `json:"meeting_id"`
	MeetingName string    `json:"meeting_name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`

	// Current phase
	Phase Phase `json:"phase"`

	// Agenda
	Agenda []AgendaItem `json:"agenda,omitempty"`

	// Pre-reads and reference documents
	PreReads []Document `json:"pre_reads,omitempty"`

	// Participants
	Participants []Participant `json:"participants,omitempty"`

	// Real-time tracking
	Actions   []role.Action `json:"actions,omitempty"`
	Decisions []Decision    `json:"decisions,omitempty"`
	Questions []Question    `json:"questions,omitempty"`
	Notes     []Note        `json:"notes,omitempty"`

	// Transcript segments (if transcription enabled)
	Transcript []TranscriptSegment `json:"transcript,omitempty"`

	// Generated artifacts
	Artifacts []role.Artifact `json:"artifacts,omitempty"`

	mu sync.RWMutex
}

// AgendaItem represents an item on the meeting agenda.
type AgendaItem struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Presenter   string        `json:"presenter,omitempty"`
	Status      string        `json:"status"` // pending, active, completed, skipped
	Notes       string        `json:"notes,omitempty"`
}

// Document represents a reference document or pre-read.
type Document struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Type     string         `json:"type"` // google_doc, confluence, url, file
	URL      string         `json:"url,omitempty"`
	Summary  string         `json:"summary,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Participant represents a meeting participant.
type Participant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"` // host, presenter, attendee, agent
	JoinedAt time.Time `json:"joined_at,omitempty"`
	LeftAt   time.Time `json:"left_at,omitempty"`
}

// Decision records a decision made during the meeting.
type Decision struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	MadeBy      string    `json:"made_by,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Context     string    `json:"context,omitempty"`
	AgendaItem  string    `json:"agenda_item,omitempty"`
}

// Question tracks an open or answered question.
type Question struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	AskedBy   string    `json:"asked_by,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Answer    string    `json:"answer,omitempty"`
	Answered  bool      `json:"answered"`
}

// Note is a general note or discussion point.
type Note struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	Speaker    string    `json:"speaker,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	AgendaItem string    `json:"agenda_item,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
}

// TranscriptSegment is a portion of the meeting transcript.
type TranscriptSegment struct {
	Speaker   string    `json:"speaker"`
	Text      string    `json:"text"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsFinal   bool      `json:"is_final"`
}

// New creates a new meeting session.
func New(meetingID, meetingName string) *Session {
	return &Session{
		MeetingID:   meetingID,
		MeetingName: meetingName,
		StartTime:   time.Now(),
		Phase:       PhasePreMeeting,
	}
}

// AddAgendaItem adds an item to the agenda.
func (s *Session) AddAgendaItem(item AgendaItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.Status == "" {
		item.Status = "pending"
	}
	s.Agenda = append(s.Agenda, item)
}

// AddPreRead adds a pre-read document.
func (s *Session) AddPreRead(doc Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PreReads = append(s.PreReads, doc)
}

// AddParticipant adds a participant.
func (s *Session) AddParticipant(p Participant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Participants = append(s.Participants, p)
}

// AddAction records an action item.
func (s *Session) AddAction(action role.Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Actions = append(s.Actions, action)
}

// AddDecision records a decision.
func (s *Session) AddDecision(decision Decision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Decisions = append(s.Decisions, decision)
}

// AddQuestion records a question.
func (s *Session) AddQuestion(q Question) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Questions = append(s.Questions, q)
}

// AddNote records a note.
func (s *Session) AddNote(note Note) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Notes = append(s.Notes, note)
}

// AddTranscript adds a transcript segment.
func (s *Session) AddTranscript(segment TranscriptSegment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transcript = append(s.Transcript, segment)
}

// AddArtifact adds a generated artifact.
func (s *Session) AddArtifact(artifact role.Artifact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Artifacts = append(s.Artifacts, artifact)
}

// SetPhase updates the meeting phase.
func (s *Session) SetPhase(phase Phase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Phase = phase
}

// End marks the session as ended.
func (s *Session) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EndTime = time.Now()
	s.Phase = PhasePostMeeting
}

// Duration returns the meeting duration.
func (s *Session) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

// OpenActions returns actions that haven't been linked to external systems.
func (s *Session) OpenActions() []role.Action {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var open []role.Action
	for _, a := range s.Actions {
		if len(a.Links) == 0 {
			open = append(open, a)
		}
	}
	return open
}

// OpenQuestions returns unanswered questions.
func (s *Session) OpenQuestions() []Question {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var open []Question
	for _, q := range s.Questions {
		if !q.Answered {
			open = append(open, q)
		}
	}
	return open
}

// Close releases any resources held by the session.
func (s *Session) Close() error {
	// Currently no resources to release, but this provides
	// a hook for future cleanup (e.g., closing connections,
	// flushing buffers, etc.)
	return nil
}
