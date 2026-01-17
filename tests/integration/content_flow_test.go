package integration

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/api-gateway/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestContentFlow(t *testing.T) {
	token, _ := getAuthToken(t)

	collectionClient := helpers.NewClient(cfg.CollectionAddr)
	collectionClient.SetToken(token)

	documentClient := helpers.NewClient(cfg.DocumentAddr)
	documentClient.SetToken(token)

	var collectionID string
	var documentID string

	// 1. Create Collection
	t.Run("Create Collection", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		name := fmt.Sprintf("Test Collection %d", rand.Intn(1000))
		req := map[string]interface{}{
			"name":        name,
			"description": "Integration test collection",
			"is_public":   true,
		}
		var respData struct {
			Success bool `json:"success"`
			Data    struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}
		resp, err := collectionClient.Request("POST", "/api/v1/collections", req, &respData)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		collectionID = respData.Data.ID
		assert.NotEmpty(t, collectionID)
	})

	// 2. Upload Document
	t.Run("Upload Document", func(t *testing.T) {
		// Create a dummy file
		filePath := filepath.Join(os.TempDir(), "test_doc.txt")
		err := os.WriteFile(filePath, []byte("This is a test document content for integration testing."), 0644)
		assert.NoError(t, err)
		defer os.Remove(filePath)

		metadata := map[string]string{
			"title":         "Integration Test Doc",
			"description":   "Uploaded via integration test",
			"collection_id": collectionID,
		}

		var respData struct {
			Success bool `json:"success"`
			Data    struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"data"`
		}

		resp, err := documentClient.UploadFile("/api/v1/documents", filePath, metadata, &respData)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		documentID = respData.Data.ID
		assert.NotEmpty(t, documentID)
	})

	// Export IDs for other tests if necessary (e.g. Search)
	// For simplicity, Search test will create its own data or we rely on this order if running sequentially
}
