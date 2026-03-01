package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthEndpoints(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	t.Run("Register new user", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["accessToken"])
		assert.NotEmpty(t, data["refreshToken"])
	})

	t.Run("Register with duplicate email", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Login with valid credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/login"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["accessToken"])
		assert.NotEmpty(t, data["refreshToken"])
	})

	t.Run("Login with invalid credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "WrongPassword!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/login"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Refresh token", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(loginPayload)

		resp, err := http.Post(ts.URL("/auth/login"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var loginResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResult)
		refreshToken := loginResult["data"].(map[string]interface{})["refreshToken"].(string)

		refreshPayload := map[string]string{
			"refreshToken": refreshToken,
		}
		body, _ = json.Marshal(refreshPayload)

		resp, err = http.Post(ts.URL("/auth/refresh"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].(map[string]interface{})
		assert.NotEmpty(t, data["accessToken"])
	})

	t.Run("Logout", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(loginPayload)

		resp, err := http.Post(ts.URL("/auth/login"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var loginResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&loginResult)
		data := loginResult["data"].(map[string]interface{})
		refreshToken := data["refreshToken"].(string)
		accessToken := data["accessToken"].(string)

		logoutPayload := map[string]string{
			"refreshToken": refreshToken,
		}
		body, _ = json.Marshal(logoutPayload)

		req, _ := http.NewRequest("POST", ts.URL("/auth/logout"), bytes.NewBuffer(body))
		req.Header = ts.AuthHeader(accessToken)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
