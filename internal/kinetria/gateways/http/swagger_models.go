package service

// DTOs for Swagger documentation

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required" example:"Bruno Costa"`
	Email    string `json:"email" validate:"required,email" example:"bruno@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"Password123!"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"bruno@example.com"`
	Password string `json:"password" validate:"required" example:"Password123!"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required" example:"abc123def456..."`
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required" example:"abc123def456..."`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"abc123def456..."`
	ExpiresIn    int    `json:"expiresIn" example:"3600"`
}

// RefreshResponse represents the token refresh response
type RefreshResponse struct {
	AccessToken string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn   int    `json:"expiresIn" example:"3600"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Logged out successfully"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request body"`
}

// SuccessResponse represents a generic success response wrapper
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// DashboardUser represents user information in dashboard
type DashboardUser struct {
	ID              string `json:"id" example:"b67a784a-e54b-4330-9b6b-8dd930c4a746"`
	Name            string `json:"name" example:"Bruno Costa"`
	Email           string `json:"email" example:"bruno@example.com"`
	ProfileImageURL string `json:"profileImageUrl" example:"/assets/avatars/default.png"`
}

// TodayWorkout represents today's workout information
type TodayWorkout struct {
	ID        string `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Name      string `json:"name" example:"Treino A - Peito e Tríceps"`
	Type      string `json:"type" example:"HIPERTROFIA"`
	Intensity string `json:"intensity" example:"ALTA"`
}

// DayProgress represents progress for a single day
type DayProgress struct {
	Day    string `json:"day" example:"Q"`
	Date   string `json:"date" example:"2026-02-25"`
	Status string `json:"status" example:"completed" enums:"completed,missed,future"`
}

// WeekStats represents weekly statistics
type WeekStats struct {
	Calories         int `json:"calories" example:"420"`
	TotalTimeMinutes int `json:"totalTimeMinutes" example:"60"`
}

// DashboardResponse represents the complete dashboard data
type DashboardResponse struct {
	User         DashboardUser  `json:"user"`
	TodayWorkout *TodayWorkout  `json:"todayWorkout"`
	WeekProgress []DayProgress  `json:"weekProgress"`
	Stats        WeekStats      `json:"stats"`
}

// Workout represents a workout plan
type Workout struct {
	ID        string `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Name      string `json:"name" example:"Treino A - Peito e Tríceps"`
	Type      string `json:"type" example:"HIPERTROFIA"`
	Intensity string `json:"intensity" example:"ALTA"`
}

// WorkoutListResponse represents the workout list response
type WorkoutListResponse struct {
	Workouts   []Workout      `json:"workouts"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"25"`
	TotalPages int `json:"totalPages" example:"3"`
}

// StartSessionRequest represents the request to start a workout session
type StartSessionRequest struct {
	WorkoutID string `json:"workoutId" validate:"required" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Notes     string `json:"notes" example:"Treino de peito e tríceps"`
}

// StartSessionResponse represents the response after starting a session
type StartSessionResponse struct {
	SessionID string `json:"sessionId" example:"f1e2d3c4-b5a6-7890-1234-567890abcdef"`
	WorkoutID string `json:"workoutId" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Status    string `json:"status" example:"active"`
	StartedAt string `json:"startedAt" example:"2026-02-25T15:30:00Z"`
}

// RecordSetRequest represents the request to record a set
type RecordSetRequest struct {
	ExerciseID string  `json:"exerciseId" validate:"required" example:"e1f2g3h4-i5j6-7890-abcd-ef1234567890"`
	SetNumber  int     `json:"setNumber" validate:"required,min=1" example:"1"`
	Reps       int     `json:"reps" validate:"required,min=0" example:"12"`
	Weight     float64 `json:"weight" validate:"min=0" example:"80.5"`
	Status     string  `json:"status" validate:"required,oneof=completed skipped" example:"completed"`
}

// RecordSetResponse represents the response after recording a set
type RecordSetResponse struct {
	SetRecordID string  `json:"setRecordId" example:"s1t2u3v4-w5x6-7890-abcd-ef1234567890"`
	ExerciseID  string  `json:"exerciseId" example:"e1f2g3h4-i5j6-7890-abcd-ef1234567890"`
	SetNumber   int     `json:"setNumber" example:"1"`
	Reps        int     `json:"reps" example:"12"`
	Weight      float64 `json:"weight" example:"80.5"`
	Status      string  `json:"status" example:"completed"`
}

// FinishSessionRequest represents the request to finish a session
type FinishSessionRequest struct {
	Notes string `json:"notes" example:"Treino completo! Ótima performance."`
}

// AbandonSessionRequest represents the request to abandon a session
type AbandonSessionRequest struct {
	Notes string `json:"notes" example:"Precisei sair mais cedo"`
}

// SessionStatusResponse represents the response after finishing/abandoning a session
type SessionStatusResponse struct {
	SessionID  string `json:"sessionId" example:"f1e2d3c4-b5a6-7890-1234-567890abcdef"`
	Status     string `json:"status" example:"completed"`
	FinishedAt string `json:"finishedAt" example:"2026-02-25T16:15:00Z"`
}

// UserPreferencesSwagger represents user preferences in profile request/response
type UserPreferencesSwagger struct {
	// Theme is the UI theme; valid values: "dark", "light"
	Theme string `json:"theme" example:"dark" enums:"dark,light"`
	// Language is the display language; valid values: "pt-BR", "en-US"
	Language string `json:"language" example:"pt-BR" enums:"pt-BR,en-US"`
}

// ProfileResponse represents the user profile in API responses
type ProfileResponse struct {
	ID              string                 `json:"id" example:"b67a784a-e54b-4330-9b6b-8dd930c4a746"`
	Name            string                 `json:"name" example:"Bruno Costa"`
	Email           string                 `json:"email" example:"bruno@example.com"`
	ProfileImageURL *string                `json:"profileImageUrl" example:"https://cdn.kinetria.app/avatars/bruno.jpg"`
	Preferences     UserPreferencesSwagger `json:"preferences"`
}

// UpdateProfileRequestSwagger represents the request body for PATCH /profile
type UpdateProfileRequestSwagger struct {
	// Name, when provided, must have 2–100 characters
	Name *string `json:"name" example:"Bruno Costa"`
	// ProfileImageURL, when provided, replaces the current profile image URL
	ProfileImageURL *string `json:"profileImageUrl" example:"https://cdn.kinetria.app/avatars/bruno.jpg"`
	// Preferences, when provided, replaces the user's full preferences object
	Preferences *UserPreferencesSwagger `json:"preferences"`
}
