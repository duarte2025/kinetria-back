package service

import (
"encoding/json"
"errors"
"net/http"
"strconv"

"github.com/go-chi/chi/v5"
"github.com/google/uuid"
"github.com/kinetria/kinetria-back/internal/kinetria/domain/entities"
domainerrors "github.com/kinetria/kinetria-back/internal/kinetria/domain/errors"
domainexercises "github.com/kinetria/kinetria-back/internal/kinetria/domain/exercises"
"github.com/kinetria/kinetria-back/internal/kinetria/domain/ports"
gatewayauth "github.com/kinetria/kinetria-back/internal/kinetria/gateways/auth"
)

// --- DTOs ---

// LibraryExerciseDTO is the JSON representation of an exercise from the library.
type LibraryExerciseDTO struct {
ID           string   `json:"id"`
Name         string   `json:"name"`
Description  *string  `json:"description"`
Instructions *string  `json:"instructions"`
Tips         *string  `json:"tips"`
Difficulty   *string  `json:"difficulty"`
Equipment    *string  `json:"equipment"`
ThumbnailURL *string  `json:"thumbnailUrl"`
VideoURL     *string  `json:"videoUrl"`
Muscles      []string `json:"muscles"`
}

// UserStatsDTO is the JSON representation of a user's performance stats for an exercise.
type UserStatsDTO struct {
LastPerformed  *string  `json:"lastPerformed"`
BestWeight     *int     `json:"bestWeight"`
TimesPerformed int      `json:"timesPerformed"`
AverageWeight  *float64 `json:"averageWeight"`
}

// LibraryExerciseWithStatsDTO extends LibraryExerciseDTO with optional user stats.
type LibraryExerciseWithStatsDTO struct {
LibraryExerciseDTO
UserStats *UserStatsDTO `json:"userStats,omitempty"`
}

// SetDetailDTO is the JSON representation of a single recorded set.
type SetDetailDTO struct {
SetNumber int    `json:"setNumber"`
Reps      int    `json:"reps"`
Weight    *int   `json:"weight"`
Status    string `json:"status"`
}

// HistoryEntryDTO is the JSON representation of one session's exercise history.
type HistoryEntryDTO struct {
SessionID   string         `json:"sessionId"`
WorkoutName string         `json:"workoutName"`
PerformedAt string         `json:"performedAt"`
Sets        []SetDetailDTO `json:"sets"`
}

// ListExercisesResponse is the paginated response for GET /exercises.
type ListExercisesResponse struct {
Data []LibraryExerciseDTO `json:"data"`
Meta PaginationMetaDTO    `json:"meta"`
}

// ExerciseDetailResponse is the response for GET /exercises/:id.
type ExerciseDetailResponse struct {
Data LibraryExerciseWithStatsDTO `json:"data"`
}

// ExerciseHistoryResponse is the paginated response for GET /exercises/:id/history.
type ExerciseHistoryResponse struct {
Data []HistoryEntryDTO `json:"data"`
Meta PaginationMetaDTO `json:"meta"`
}

// --- Handler ---

// ExercisesHandler handles HTTP requests for the exercise library endpoints.
type ExercisesHandler struct {
listExercisesUC      *domainexercises.ListExercisesUC
getExerciseUC        *domainexercises.GetExerciseUC
getExerciseHistoryUC *domainexercises.GetExerciseHistoryUC
jwtManager           *gatewayauth.JWTManager
}

// NewExercisesHandler creates a new ExercisesHandler.
func NewExercisesHandler(
listExercisesUC *domainexercises.ListExercisesUC,
getExerciseUC *domainexercises.GetExerciseUC,
getExerciseHistoryUC *domainexercises.GetExerciseHistoryUC,
jwtManager *gatewayauth.JWTManager,
) *ExercisesHandler {
return &ExercisesHandler{
listExercisesUC:      listExercisesUC,
getExerciseUC:        getExerciseUC,
getExerciseHistoryUC: getExerciseHistoryUC,
jwtManager:           jwtManager,
}
}

// HandleListExercises handles GET /api/v1/exercises
// Returns a paginated list of exercises with optional filters.
func (h *ExercisesHandler) HandleListExercises(w http.ResponseWriter, r *http.Request) {
q := r.URL.Query()

// Parse pagination with defaults
page, err := parseLibraryIntParam(q.Get("page"), 1)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "page must be a valid integer")
return
}
if page < 1 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "page must be >= 1")
return
}

pageSize, err := parseLibraryIntParam(q.Get("pageSize"), 20)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be a valid integer")
return
}
if pageSize < 1 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be >= 1")
return
}
if pageSize > 100 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be <= 100")
return
}

filters := ports.ExerciseFilters{
MuscleGroup: nullableQueryParam(q.Get("muscleGroup")),
Equipment:   nullableQueryParam(q.Get("equipment")),
Difficulty:  nullableQueryParam(q.Get("difficulty")),
Search:      nullableQueryParam(q.Get("search")),
}

output, err := h.listExercisesUC.Execute(r.Context(), domainexercises.ListExercisesInput{
Filters:  filters,
Page:     page,
PageSize: pageSize,
})
if err != nil {
writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
return
}

dtos := make([]LibraryExerciseDTO, 0, len(output.Exercises))
for _, e := range output.Exercises {
dtos = append(dtos, mapExerciseToLibraryDTO(e))
}

resp := ListExercisesResponse{
Data: dtos,
Meta: PaginationMetaDTO{
Page:       output.Page,
PageSize:   output.PageSize,
Total:      output.Total,
TotalPages: output.TotalPages,
},
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(resp)
}

// HandleGetExercise handles GET /api/v1/exercises/{id}
// Returns exercise details. If authenticated, includes user stats.
func (h *ExercisesHandler) HandleGetExercise(w http.ResponseWriter, r *http.Request) {
exerciseIDStr := chi.URLParam(r, "id")
exerciseID, err := uuid.Parse(exerciseIDStr)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid exercise ID")
return
}

// Optional auth: try to extract userID without failing if absent/invalid
userID := tryExtractUserIDFromJWT(r, h.jwtManager)

result, err := h.getExerciseUC.Execute(r.Context(), exerciseID, userID)
if err != nil {
if errors.Is(err, domainerrors.ErrExerciseNotFound) {
writeError(w, http.StatusNotFound, "NOT_FOUND", "exercise not found")
return
}
writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
return
}

dto := LibraryExerciseWithStatsDTO{
LibraryExerciseDTO: mapExerciseToLibraryDTO(result.Exercise),
}
if result.UserStats != nil {
dto.UserStats = mapStatsToDTO(result.UserStats)
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(ExerciseDetailResponse{Data: dto})
}

// HandleGetExerciseHistory handles GET /api/v1/exercises/{id}/history
// Requires authentication. Returns the user's history of performing the exercise.
func (h *ExercisesHandler) HandleGetExerciseHistory(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

// Require authentication
userID, ok := ctx.Value(userIDKey).(uuid.UUID)
if !ok {
writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
return
}

exerciseIDStr := chi.URLParam(r, "id")
exerciseID, err := uuid.Parse(exerciseIDStr)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid exercise ID")
return
}

// Parse pagination with defaults
page, err := parseLibraryIntParam(r.URL.Query().Get("page"), 1)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "page must be a valid integer")
return
}
if page < 1 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "page must be >= 1")
return
}

pageSize, err := parseLibraryIntParam(r.URL.Query().Get("pageSize"), 20)
if err != nil {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be a valid integer")
return
}
if pageSize < 1 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be >= 1")
return
}
if pageSize > 100 {
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "pageSize must be <= 100")
return
}

output, err := h.getExerciseHistoryUC.Execute(ctx, domainexercises.GetExerciseHistoryInput{
ExerciseID: exerciseID,
UserID:     userID,
Page:       page,
PageSize:   pageSize,
})
if err != nil {
if errors.Is(err, domainerrors.ErrExerciseNotFound) {
writeError(w, http.StatusNotFound, "NOT_FOUND", "exercise not found")
return
}
writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
return
}

dtos := make([]HistoryEntryDTO, 0, len(output.Entries))
for _, entry := range output.Entries {
sets := make([]SetDetailDTO, 0, len(entry.Sets))
for _, s := range entry.Sets {
sets = append(sets, SetDetailDTO{
SetNumber: s.SetNumber,
Reps:      s.Reps,
Weight:    s.Weight,
Status:    s.Status,
})
}
dtos = append(dtos, HistoryEntryDTO{
SessionID:   entry.SessionID.String(),
WorkoutName: entry.WorkoutName,
PerformedAt: entry.PerformedAt.UTC().Format("2006-01-02T15:04:05Z"),
Sets:        sets,
})
}

resp := ExerciseHistoryResponse{
Data: dtos,
Meta: PaginationMetaDTO{
Page:       output.Page,
PageSize:   output.PageSize,
Total:      output.Total,
TotalPages: output.TotalPages,
},
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(resp)
}

// --- Helpers ---

// mapExerciseToLibraryDTO converts a domain Exercise entity to LibraryExerciseDTO.
func mapExerciseToLibraryDTO(e *entities.Exercise) LibraryExerciseDTO {
dto := LibraryExerciseDTO{
ID:      e.ID.String(),
Name:    e.Name,
Muscles: e.Muscles,
}
if e.ThumbnailURL != "" {
dto.ThumbnailURL = &e.ThumbnailURL
}
dto.Description = e.Description
dto.Instructions = e.Instructions
dto.Tips = e.Tips
dto.Difficulty = e.Difficulty
dto.Equipment = e.Equipment
dto.VideoURL = e.VideoURL
return dto
}

// parseLibraryIntParam parses an integer query param, returning defaultValue if the string is empty.
func parseLibraryIntParam(s string, defaultValue int) (int, error) {
if s == "" {
return defaultValue, nil
}
return strconv.Atoi(s)
}

// nullableQueryParam converts a query param string to *string, returning nil if empty.
func nullableQueryParam(s string) *string {
if s == "" {
return nil
}
return &s
}

// mapStatsToDTO converts ports.ExerciseUserStats to UserStatsDTO.
func mapStatsToDTO(stats *ports.ExerciseUserStats) *UserStatsDTO {
dto := &UserStatsDTO{
TimesPerformed: stats.TimesPerformed,
BestWeight:     stats.BestWeight,
AverageWeight:  stats.AverageWeight,
}
if stats.LastPerformed != nil {
s := stats.LastPerformed.UTC().Format("2006-01-02T15:04:05Z")
dto.LastPerformed = &s
}
return dto
}
