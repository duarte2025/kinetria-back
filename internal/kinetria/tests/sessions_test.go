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

func TestSessionsEndpoints(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	var accessToken string
	var userID uuid.UUID
	var workoutID uuid.UUID

	t.Run("Setup: Register user and create workout", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Session Test User",
			"email":    "session@example.com",
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

		_, err = ts.DB.Exec(`
			INSERT INTO workouts (id, user_id, name, type, intensity, duration, created_at, updated_at)
			VALUES ($1, $2, 'Test Workout', 'FORÇA', 'Alta', 60, NOW(), NOW())
		`, uuid.New(), userID)
		require.NoError(t, err)

		err = ts.DB.QueryRow(`SELECT id FROM workouts WHERE user_id = $1`, userID).Scan(&workoutID)
		require.NoError(t, err)
	})

	t.Run("Start session successfully", func(t *testing.T) {
		payload := map[string]string{
			"workoutId": workoutID.String(),
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", ts.URL("/sessions"), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, workoutID.String(), data["workoutId"])
		assert.Equal(t, "active", data["status"])
		assert.NotEmpty(t, data["startedAt"])
	})

	t.Run("Start session with active session already exists", func(t *testing.T) {
		payload := map[string]string{
			"workoutId": workoutID.String(),
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", ts.URL("/sessions"), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Start session without auth", func(t *testing.T) {
		payload := map[string]string{
			"workoutId": workoutID.String(),
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/sessions"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Start session with invalid workout ID", func(t *testing.T) {
		payload := map[string]string{
			"workoutId": "invalid-uuid",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", ts.URL("/sessions"), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	var sessionID string
	var exerciseID uuid.UUID

	t.Run("Setup: Create exercise for set recording", func(t *testing.T) {
		exerciseID = uuid.New()
		_, err := ts.DB.Exec(`
			INSERT INTO exercises (id, name, created_at, updated_at)
			VALUES ($1, 'Test Exercise', NOW(), NOW())
		`, exerciseID)
		require.NoError(t, err)

		_, err = ts.DB.Exec(`
			INSERT INTO workout_exercises (workout_id, exercise_id, sets, reps, rest_time, order_index)
			VALUES ($1, $2, 3, '10', 60, 1)
		`, workoutID, exerciseID)
		require.NoError(t, err)

		err = ts.DB.QueryRow(`SELECT id FROM sessions WHERE user_id = $1 AND status = 'active'`, userID).Scan(&sessionID)
		require.NoError(t, err)
	})

	t.Run("Record set successfully", func(t *testing.T) {
		payload := map[string]interface{}{
			"exerciseId": exerciseID.String(),
			"setNumber":  1,
			"weight":     80000,
			"reps":       10,
			"status":     "completed",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", ts.URL(fmt.Sprintf("/sessions/%s/sets", sessionID)), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, sessionID, data["sessionId"])
		assert.Equal(t, float64(1), data["setNumber"])
		assert.Equal(t, "completed", data["status"])
	})

	t.Run("Record set with invalid session ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"exerciseId": exerciseID.String(),
			"setNumber":  1,
			"weight":     80000,
			"reps":       10,
			"status":     "completed",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", ts.URL("/sessions/invalid-uuid/sets"), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Finish session successfully", func(t *testing.T) {
		payload := map[string]string{
			"notes": "Great workout!",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PATCH", ts.URL(fmt.Sprintf("/sessions/%s/finish", sessionID)), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})

		assert.Equal(t, sessionID, data["id"])
		assert.Equal(t, "completed", data["status"])
		assert.NotEmpty(t, data["finishedAt"])
	})

	t.Run("Finish session already closed", func(t *testing.T) {
		payload := map[string]string{
			"notes": "Another note",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PATCH", ts.URL(fmt.Sprintf("/sessions/%s/finish", sessionID)), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Abandon session successfully", func(t *testing.T) {
		newSessionID := uuid.New()
		_, err := ts.DB.Exec(`
			INSERT INTO sessions (id, user_id, workout_id, started_at, status, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), 'active', NOW(), NOW())
		`, newSessionID, userID, workoutID)
		require.NoError(t, err)

		req, _ := http.NewRequest("PATCH", ts.URL(fmt.Sprintf("/sessions/%s/abandon", newSessionID.String())), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})

		assert.Equal(t, newSessionID.String(), data["id"])
		assert.Equal(t, "abandoned", data["status"])
	})

	t.Run("Abandon session not found", func(t *testing.T) {
		fakeSessionID := uuid.New()

		req, _ := http.NewRequest("PATCH", ts.URL(fmt.Sprintf("/sessions/%s/abandon", fakeSessionID.String())), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
