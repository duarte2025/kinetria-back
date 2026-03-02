// Package statistics provides use cases for workout statistics and analytics.
package statistics

import (
	"time"

	"github.com/google/uuid"
)

// OverviewStats holds aggregated workout statistics for a user over a period.
type OverviewStats struct {
	// Period
	StartDate time.Time
	EndDate   time.Time

	// Workouts
	TotalWorkouts  int
	AveragePerWeek float64

	// Time
	TotalTimeMinutes int

	// Sets/Reps/Volume
	TotalSets   int
	TotalReps   int
	TotalVolume int64 // gramas

	// Streak
	CurrentStreak int
	LongestStreak int
}

// ProgressionData holds the progression data for a user and optionally a specific exercise.
type ProgressionData struct {
	ExerciseID   *uuid.UUID
	ExerciseName string
	Points       []ProgressionPoint
}

// ProgressionPoint holds aggregated performance data for a single day.
type ProgressionPoint struct {
	Date        time.Time
	MaxWeight   int64   // gramas
	TotalVolume int64   // gramas * reps
	Change      float64 // percentual de mudança em relação ao ponto anterior
}

// PersonalRecord holds the best performance for a specific exercise.
type PersonalRecord struct {
	ExerciseID   uuid.UUID
	ExerciseName string
	Weight       int   // gramas
	Reps         int
	Volume       int64 // gramas * reps
	AchievedAt   time.Time
}

// FrequencyData holds the workout count for a specific date.
type FrequencyData struct {
	Date  time.Time
	Count int
}
