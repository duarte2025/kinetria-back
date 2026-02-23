package vos

type WorkoutStatus string

const (
	WorkoutStatusDraft     WorkoutStatus = "draft"
	WorkoutStatusPublished WorkoutStatus = "published"
	WorkoutStatusArchived  WorkoutStatus = "archived"
)

func (s WorkoutStatus) String() string {
	return string(s)
}

func (s WorkoutStatus) IsValid() bool {
	switch s {
	case WorkoutStatusDraft, WorkoutStatusPublished, WorkoutStatusArchived:
		return true
	}
	return false
}
