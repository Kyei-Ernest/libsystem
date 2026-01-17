package integration

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/api-gateway/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestUserLifecycle(t *testing.T) {
	client := helpers.NewClient(cfg.UserAddr)

	// Generate unique user
	rand.Seed(time.Now().UnixNano())
	username := fmt.Sprintf("testuser_%d", rand.Intn(100000))
	email := fmt.Sprintf("%s@example.com", username)
	password := "Password123!"

	// 1. Register
	t.Run("Register User", func(t *testing.T) {
		req := map[string]string{
			"username":   username,
			"email":      email,
			"password":   password,
			"first_name": "Test",
			"last_name":  "User",
		}
		var respData map[string]interface{}
		resp, err := client.Request("POST", "/api/v1/auth/register", req, &respData)
		if err != nil {
			t.Logf("Register failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Logf("Register response status: %d", resp.StatusCode)
			t.Logf("Register response body: %v", respData)
		}
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, true, respData["success"])
	})

	// 2. Login
	t.Run("Login User", func(t *testing.T) {
		req := map[string]string{
			"email_or_username": email,
			"password":          password,
		}
		var respData struct {
			Success bool `json:"success"`
			Data    struct {
				Token string `json:"token"`
				User  struct {
					ID       string `json:"id"`
					Username string `json:"username"`
				} `json:"user"`
			} `json:"data"`
		}
		resp, err := client.Request("POST", "/api/v1/auth/login", req, &respData)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, respData.Data.Token)

		// Set token for future requests if needed, or return it
	})
}

// Helper to get a valid token for other tests
func getAuthToken(t *testing.T) (string, string) {
	client := helpers.NewClient(cfg.UserAddr)
	rand.Seed(time.Now().UnixNano())
	username := fmt.Sprintf("auth_user_%d", rand.Intn(100000))
	email := fmt.Sprintf("%s@example.com", username)
	password := "Password123!"

	// Register
	_, err := client.Request("POST", "/api/v1/auth/register", map[string]string{
		"username":   username,
		"email":      email,
		"password":   password,
		"first_name": "Auth",
		"last_name":  "User",
	}, nil)
	assert.NoError(t, err)

	// Login
	var respData struct {
		Data struct {
			Token string `json:"token"`
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	_, err = client.Request("POST", "/api/v1/auth/login", map[string]string{
		"email_or_username": email,
		"password":          password,
	}, &respData)
	assert.NoError(t, err)
	return respData.Data.Token, respData.Data.User.ID
}
