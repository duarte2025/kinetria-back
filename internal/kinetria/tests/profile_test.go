package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileEndpoints(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup(t)

	var accessToken string

	t.Run("Setup: Register user", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Profile User",
			"email":    "profile@example.com",
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)

		resp, err := http.Post(ts.URL("/auth/register"), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		data := result["data"].(map[string]interface{})
		accessToken = data["accessToken"].(string)
		require.NotEmpty(t, accessToken)
	})

	t.Run("GET /profile", func(t *testing.T) {
		t.Run("with valid JWT", func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.URL("/profile"), nil)
			require.NoError(t, err)
			req.Header = ts.AuthHeader(accessToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			data := result["data"].(map[string]interface{})
			assert.NotEmpty(t, data["id"])
			assert.Equal(t, "Profile User", data["name"])
			assert.Equal(t, "profile@example.com", data["email"])

			// preferences should be present as a nested object
			prefs, ok := data["preferences"].(map[string]interface{})
			require.True(t, ok, "preferences should be a JSON object")
			_, hasTheme := prefs["theme"]
			_, hasLanguage := prefs["language"]
			assert.True(t, hasTheme, "preferences.theme should be present")
			assert.True(t, hasLanguage, "preferences.language should be present")
		})

		t.Run("without JWT", func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.URL("/profile"), nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})

		t.Run("with invalid JWT", func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.URL("/profile"), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer invalid.jwt.token")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("PATCH /profile", func(t *testing.T) {
		tests := []struct {
			name           string
			payload        interface{}
			expectedStatus int
			check          func(t *testing.T, data map[string]interface{})
		}{
			{
				name:           "update name",
				payload:        map[string]interface{}{"name": "Updated Name"},
				expectedStatus: http.StatusOK,
				check: func(t *testing.T, data map[string]interface{}) {
					assert.Equal(t, "Updated Name", data["name"])
				},
			},
			{
				name: "update preferences",
				payload: map[string]interface{}{
					"preferences": map[string]string{
						"theme":    "dark",
						"language": "en-US",
					},
				},
				expectedStatus: http.StatusOK,
				check: func(t *testing.T, data map[string]interface{}) {
					prefs := data["preferences"].(map[string]interface{})
					assert.Equal(t, "dark", prefs["theme"])
					assert.Equal(t, "en-US", prefs["language"])
				},
			},
			{
				name:           "update profileImageUrl",
				payload:        map[string]interface{}{"profileImageUrl": "https://example.com/avatar.jpg"},
				expectedStatus: http.StatusOK,
				check: func(t *testing.T, data map[string]interface{}) {
					assert.Equal(t, "https://example.com/avatar.jpg", data["profileImageUrl"])
				},
			},
			{
				name: "update multiple fields",
				payload: map[string]interface{}{
					"name":            "Multi Updated",
					"profileImageUrl": "https://example.com/new.jpg",
					"preferences": map[string]string{
						"theme":    "light",
						"language": "pt-BR",
					},
				},
				expectedStatus: http.StatusOK,
				check: func(t *testing.T, data map[string]interface{}) {
					assert.Equal(t, "Multi Updated", data["name"])
					assert.Equal(t, "https://example.com/new.jpg", data["profileImageUrl"])
					prefs := data["preferences"].(map[string]interface{})
					assert.Equal(t, "light", prefs["theme"])
					assert.Equal(t, "pt-BR", prefs["language"])
				},
			},
			{
				name:           "name too short",
				payload:        map[string]interface{}{"name": "A"},
				expectedStatus: http.StatusBadRequest,
				check:          nil,
			},
			{
				name:           "name with only spaces",
				payload:        map[string]interface{}{"name": "   "},
				expectedStatus: http.StatusBadRequest,
				check:          nil,
			},
			{
				name: "invalid preferences theme",
				payload: map[string]interface{}{
					"preferences": map[string]string{
						"theme":    "invalid-theme",
						"language": "pt-BR",
					},
				},
				expectedStatus: http.StatusBadRequest,
				check:          nil,
			},
			{
				name:           "no fields provided (empty body)",
				payload:        map[string]interface{}{},
				expectedStatus: http.StatusBadRequest,
				check:          nil,
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				body, err := json.Marshal(tc.payload)
				require.NoError(t, err)

				req, err := http.NewRequest("PATCH", ts.URL("/profile"), bytes.NewBuffer(body))
				require.NoError(t, err)
				req.Header = ts.AuthHeader(accessToken)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				if tc.check != nil {
					var result map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&result)
					require.NoError(t, err)
					data := result["data"].(map[string]interface{})
					tc.check(t, data)
				}
			})
		}

		t.Run("without JWT", func(t *testing.T) {
			payload := map[string]interface{}{"name": "No Auth User"}
			body, _ := json.Marshal(payload)

			req, err := http.NewRequest("PATCH", ts.URL("/profile"), bytes.NewBuffer(body))
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}
