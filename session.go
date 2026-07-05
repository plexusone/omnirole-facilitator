package facilitator

import (
	"github.com/plexusone/omnirole-facilitator/session"
)

// MeetingSession is an alias for session.Session for convenience.
type MeetingSession = session.Session

// NewMeetingSession creates a new meeting session.
func NewMeetingSession(meetingID, meetingName string) *MeetingSession {
	return session.New(meetingID, meetingName)
}

// Re-export session types for convenience.
type (
	AgendaItem        = session.AgendaItem
	Document          = session.Document
	Participant       = session.Participant
	Decision          = session.Decision
	Question          = session.Question
	Note              = session.Note
	TranscriptSegment = session.TranscriptSegment
	Phase             = session.Phase
)

// Re-export phase constants.
const (
	PhasePreMeeting  = session.PhasePreMeeting
	PhaseMeeting     = session.PhaseMeeting
	PhasePostMeeting = session.PhasePostMeeting
)
