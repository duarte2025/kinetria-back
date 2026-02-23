package vos

type AuditAction string

const (
	AuditActionUserCreated AuditAction = "user_created"
	AuditActionUserUpdated AuditAction = "user_updated"
	AuditActionUserDeleted AuditAction = "user_deleted"

	AuditActionWorkoutCreated AuditAction = "workout_created"
	AuditActionWorkoutUpdated AuditAction = "workout_updated"
	AuditActionWorkoutDeleted AuditAction = "workout_deleted"

	AuditActionSessionStarted   AuditAction = "session_started"
	AuditActionSessionCompleted AuditAction = "session_completed"
	AuditActionSessionCancelled AuditAction = "session_cancelled"

	AuditActionSetRecorded AuditAction = "set_recorded"
	AuditActionSetUpdated  AuditAction = "set_updated"
	AuditActionSetDeleted  AuditAction = "set_deleted"

	AuditActionLogin          AuditAction = "login"
	AuditActionLogout         AuditAction = "logout"
	AuditActionTokenRefreshed AuditAction = "token_refreshed"
)

func (a AuditAction) String() string {
	return string(a)
}

func (a AuditAction) IsValid() bool {
	switch a {
	case AuditActionUserCreated, AuditActionUserUpdated, AuditActionUserDeleted,
		AuditActionWorkoutCreated, AuditActionWorkoutUpdated, AuditActionWorkoutDeleted,
		AuditActionSessionStarted, AuditActionSessionCompleted, AuditActionSessionCancelled,
		AuditActionSetRecorded, AuditActionSetUpdated, AuditActionSetDeleted,
		AuditActionLogin, AuditActionLogout, AuditActionTokenRefreshed:
		return true
	}
	return false
}
