package vos

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusCompleted  SessionStatus = "completed"
	SessionStatusCancelled  SessionStatus = "cancelled"
)

func (s SessionStatus) String() string {
	return string(s)
}

func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusInProgress, SessionStatusCompleted, SessionStatusCancelled:
		return true
	}
	return false
}
