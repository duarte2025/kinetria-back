package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkoutsEndpoints(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	var accessToken string
	var userID uuid.UUID

	t.Run("Setup: Register user", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Workout Test User",
			"email":    "workout@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		accessToken = data["accessToken"].(string)

		parsedUserID, err := ts.JWTManager.ParseToken(accessToken)
		require.NoError(t, err)
		userID = parsedUserID
	})

	var workoutIDs []uuid.UUID

	t.Run("Setup: Create workouts", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			workoutID := uuid.New()
			_, err := ts.DB.Exec(`
				INSERT INTO workouts (id, user_id, name, description, type, intensity, duration, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			`, workoutID, userID, fmt.Sprintf("Workout %d", i+1), "Test description", "FORÇA", "Alta", 60)
			require.NoError(t, err)
			workoutIDs = append(workoutIDs, workoutID)
		}
	})

	t.Run("List workouts successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/workouts"), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 5)

		meta := result["meta"].(map[string]interface{})
		assert.Equal(t, float64(1), meta["page"])
		assert.Equal(t, float64(20), meta["pageSize"])
		assert.Equal(t, float64(5), meta["total"])
		assert.Equal(t, float64(1), meta["totalPages"])
	})

	t.Run("List workouts with pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/workouts?page=1&pageSize=2"), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 2)

		meta := result["meta"].(map[string]interface{})
		assert.Equal(t, float64(1), meta["page"])
		assert.Equal(t, float64(2), meta["pageSize"])
		assert.Equal(t, float64(5), meta["total"])
		assert.Equal(t, float64(3), meta["totalPages"])
	})

	t.Run("List workouts without auth", func(t *testing.T) {
		resp, err := http.Get(ts.URL("/workouts"))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("List workouts with invalid page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/workouts?page=invalid"), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	var workoutWithExercises uuid.UUID

	t.Run("Setup: Create workout with exercises", func(t *testing.T) {
		workoutWithExercises = uuid.New()
		_, err := ts.DB.Exec(`
			INSERT INTO workouts (id, user_id, name, description, type, intensity, duration, image_url, created_at, updated_at)
			VALUES ($1, $2, 'Full Workout', 'Complete workout', 'HIPERTROFIA', 'Moderada', 90, 'https://example.com/image.jpg', NOW(), NOW())
		`, workoutWithExercises, userID)
		require.NoError(t, err)

		exerciseID1 := uuid.New()
		exerciseID2 := uuid.New()

		_, err = ts.DB.Exec(`
			INSERT INTO exercises (id, name, description, thumbnail_url, muscles, created_at, updated_at)
			VALUES 
				($1, 'Bench Press', 'Chest exercise', 'https://example.com/bench.jpg', '["Peito", "Tríceps"]'::jsonb, NOW(), NOW()),
				($2, 'Squat', 'Leg exercise', '', '["Pernas", "Glúteos"]'::jsonb, NOW(), NOW())
		`, exerciseID1, exerciseID2)
		require.NoError(t, err)

		_, err = ts.DB.Exec(`
			INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, weight, order_index)
			VALUES 
				($1, $2, 4, '8-12', 90, 80000, 1),
				($1, $3, 3, '10-15', 60, 0, 2)
		`, workoutWithExercises, exerciseID1, exerciseID2)
		require.NoError(t, err)
	})

	t.Run("Get workout by ID successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL(fmt.Sprintf("/workouts/%s", workoutWithExercises.String())), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.Equal(t, workoutWithExercises.String(), data["id"])
		assert.Equal(t, "Full Workout", data["name"])
		assert.Equal(t, "Complete workout", data["description"])
		assert.Equal(t, "HIPERTROFIA", data["type"])
		assert.Equal(t, "Moderada", data["intensity"])
		assert.Equal(t, float64(90), data["duration"])
		assert.Equal(t, "https://example.com/image.jpg", data["imageUrl"])

		exercises := data["exercises"].([]interface{})
		assert.Len(t, exercises, 2)

		ex1 := exercises[0].(map[string]interface{})
		assert.Equal(t, "Bench Press", ex1["name"])
		assert.Equal(t, float64(4), ex1["sets"])
		assert.Equal(t, "8-12", ex1["reps"])
		assert.Equal(t, float64(90), ex1["restTime"])
		assert.Equal(t, float64(80000), ex1["weight"])
		assert.Contains(t, ex1["muscles"], "Peito")
	})

	t.Run("Get workout by ID not found", func(t *testing.T) {
		fakeID := uuid.New()
		req, _ := http.NewRequest("GET", ts.URL(fmt.Sprintf("/workouts/%s", fakeID.String())), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Get workout by ID without auth", func(t *testing.T) {
		resp, err := http.Get(ts.URL(fmt.Sprintf("/workouts/%s", workoutWithExercises.String())))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Get workout with invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/workouts/invalid-uuid"), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Get workout from another user", func(t *testing.T) {
		otherUserID := uuid.New()
		otherWorkoutID := uuid.New()

		_, err := ts.DB.Exec(`
			INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
			VALUES ($1, 'Other User', 'other@example.com', 'hash', NOW(), NOW())
		`, otherUserID)
		require.NoError(t, err)

		_, err = ts.DB.Exec(`
			INSERT INTO workouts (id, user_id, name, type, intensity, duration, created_at, updated_at)
			VALUES ($1, $2, 'Other Workout', 'FORÇA', 'Alta', 60, NOW(), NOW())
		`, otherWorkoutID, otherUserID)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", ts.URL(fmt.Sprintf("/workouts/%s", otherWorkoutID.String())), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
