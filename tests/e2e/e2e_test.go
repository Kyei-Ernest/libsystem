package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8088/api/v1"
)

type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
}

type AuthResponse struct {
	Data struct {
		Token string `json:"token"`
		User  struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Document struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SearchResponse struct {
	Data struct {
		Hits   []Document                  `json:"hits"`
		Total  int64                       `json:"total"`
		Facets map[string]map[string]int64 `json:"facets"`
	} `json:"data"`
}

func TestEndToEndFlow(t *testing.T) {
	// Unique identifier for this test run
	runID := uuid.New().String()[:8]
	email := fmt.Sprintf("admin-%s@example.com", runID)
	username := fmt.Sprintf("admin-%s", runID)
	password := "SecurePass123!"

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Register User
	t.Logf("Registering user: %s", email)
	regBody, _ := json.Marshal(RegisterRequest{
		Email:     email,
		Username:  username,
		Password:  password,
		FirstName: "Admin",
		LastName:  "User",
	})
	resp, err := client.Post(baseURL+"/users/register", "application/json", bytes.NewBuffer(regBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Read response to get ID/Token if needed, or just login
	var regResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&regResp)
	resp.Body.Close()
	token := regResp.Data.Token
	userID := regResp.Data.User.ID
	require.NotEmpty(t, token, "Token should not be empty")

	// 2. Upgrade to Admin (Internal Hack or just use the user as is for now?)
	// The system defaults to "patron". To test Admin features, we need to be Admin.
	// Since we can't easily change database state from outside without an admin endpoint...
	// Wait, we implemented a way to update roles... but only an Admin can update roles!
	// Chicken and egg.
	// Make the prompt user aware: "End-to-end test assumes default user or pre-seeded admin".
	// However, for this test, let's proceed as Patron for Upload/Search/Download.
	// We can test Login/Register/Upload/Search/Download without Admin role.
	// Admin Deactivate will fail, verifying RBAC works!

	authHeader := "Bearer " + token

	// 3. Create Collection
	t.Log("Creating collection...")
	collName := fmt.Sprintf("Collection %s", runID)
	collBody, _ := json.Marshal(map[string]interface{}{
		"name":        collName,
		"description": "E2E Test Collection",
		"is_public":   true,
	})
	req, _ := http.NewRequest("POST", baseURL+"/collections", bytes.NewBuffer(collBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var collResp struct{ Data Collection }
	json.NewDecoder(resp.Body).Decode(&collResp)
	resp.Body.Close()
	collectionID := collResp.Data.ID
	require.NotEmpty(t, collectionID)

	// 4. Upload Document
	t.Log("Uploading document...")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a dummy TXT file with unique content
	uniqueContent := fmt.Sprintf("Hello World %s", runID)
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "test.txt")}
	h["Content-Type"] = []string{"text/plain"}
	part, _ := writer.CreatePart(h)
	part.Write([]byte(uniqueContent))

	writer.WriteField("title", fmt.Sprintf("Doc %s", runID))
	writer.WriteField("collection_id", collectionID)
	writer.Close()

	req, _ = http.NewRequest("POST", baseURL+"/documents", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	require.NoError(t, err)
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Upload Failed: %s", string(bodyBytes))
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var docResp struct{ Data Document }
	json.NewDecoder(resp.Body).Decode(&docResp)
	resp.Body.Close()
	docID := docResp.Data.ID
	require.NotEmpty(t, docID)

	// 5. Wait for Indexing (Poller)
	t.Log("Waiting for indexing...")
	var searchResp SearchResponse
	found := false
	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		req, _ = http.NewRequest("GET", baseURL+"/search?q="+runID, nil)
		// Search typically public, but we can send auth
		req.Header.Set("Authorization", authHeader)
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			json.Unmarshal(bodyBytes, &searchResp)
			if searchResp.Data.Total > 0 {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Document should be indexed and found")

	// 6. Verify Facets
	t.Log("Verifying facets...")
	// We expect facets in the search response now
	// require.NotNil(t, searchResp.Data.Facets) // This might fail if facets are empty but map exists?
	// Check if key exists
	_, hasFileTypes := searchResp.Data.Facets["file_types"]
	assert.True(t, hasFileTypes, "Facets should include file_types")

	// 7. Download Document
	t.Log("Downloading document...")
	req, _ = http.NewRequest("GET", baseURL+"/documents/"+docID+"/download", nil)
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify content match (first few bytes)
	header := make([]byte, 5)
	resp.Body.Read(header)
	resp.Body.Close()
	assert.Equal(t, "Hello", string(header))

	// 8. Admin Action (Should Fail as Patron)
	t.Log("Verifying RBAC (Patron cannot deactivate)...")
	req, _ = http.NewRequest("DELETE", baseURL+"/users/"+userID, nil) // Deactivate endpoint
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	require.NoError(t, err)
	// Should be Forbidden (403) or Unauthorized if logic differs, but definitely not 200
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	t.Log("E2E Test Completed Successfully")
}
