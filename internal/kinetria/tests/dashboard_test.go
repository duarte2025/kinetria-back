package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kinetria/kinetria-back/internal/kinetria/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardEndpoint(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	var accessToken string
	var userID uuid.UUID
	var workoutID uuid.UUID

	t.Run("Setup: Register user and create data", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Dashboard User",
			"email":    "dashboard@example.com",
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

		// Get user ID from database
		ctx := context.Background()
		err = ts.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", "dashboard@example.com").Scan(&userID)
		require.NoError(t, err)

		workoutID = uuid.New()

		workoutType := string(vos.WorkoutTypeForca)
		intensity := string(vos.WorkoutIntensityAlta)

		_, err = ts.DB.ExecContext(ctx, `
			INSERT INTO workouts (id, user_id, name, type, intensity, duration, image_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, workoutID, userID, "Today's Workout", workoutType, intensity, 60, "https://example.com/image.jpg")
		require.NoError(t, err)

		now := time.Now()
		sessionID := uuid.New()
		_, err = ts.DB.ExecContext(ctx, `
			INSERT INTO sessions (id, user_id, workout_id, status, started_at, finished_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, sessionID, userID, workoutID, vos.SessionStatusCompleted, now.Add(-2*time.Hour), now)
		require.NoError(t, err)
	})

	t.Run("Get dashboard", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/dashboard"), nil)
		req.Header = ts.AuthHeader(accessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})

		user := data["user"].(map[string]interface{})
		assert.Equal(t, userID.String(), user["id"])
		assert.Equal(t, "Dashboard User", user["name"])
		assert.Equal(t, "dashboard@example.com", user["email"])

		todayWorkout := data["todayWorkout"].(map[string]interface{})
		assert.Equal(t, workoutID.String(), todayWorkout["id"])
		assert.Equal(t, "Today's Workout", todayWorkout["name"])

		weekProgress := data["weekProgress"].([]interface{})
		assert.Len(t, weekProgress, 7)

		stats := data["stats"].(map[string]interface{})
		assert.NotNil(t, stats["calories"])
		assert.NotNil(t, stats["totalTimeMinutes"])
	})

	t.Run("Get dashboard without auth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL("/dashboard"), nil)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Dashboard with no workout today", func(t *testing.T) {
		payload := map[string]string{
			"name":     "No Workout User",
			"email":    "noworkout@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		token := data["accessToken"].(string)

		req, _ := http.NewRequest("GET", ts.URL("/dashboard"), nil)
		req.Header = ts.AuthHeader(token)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		json.NewDecoder(resp.Body).Decode(&result)
		data = result["data"].(map[string]interface{})

		assert.Nil(t, data["todayWorkout"])

		weekProgress := data["weekProgress"].([]interface{})
		assert.Len(t, weekProgress, 7)

		stats := data["stats"].(map[string]interface{})
		assert.Equal(t, float64(0), stats["calories"])
		assert.Equal(t, float64(0), stats["totalTimeMinutes"])
	})
}
