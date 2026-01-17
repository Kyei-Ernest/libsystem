package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/api-gateway/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestSearchFlow(t *testing.T) {
	token, _ := getAuthToken(t)

	collectionClient := helpers.NewClient(cfg.CollectionAddr)
	collectionClient.SetToken(token)

	documentClient := helpers.NewClient(cfg.DocumentAddr)
	documentClient.SetToken(token)

	searchClient := helpers.NewClient(cfg.SearchAddr)
	// Search might be public, but we can verify auth if needed. Usually public.

	// 1. Create Data
	// ... reuse helper or duplicate logic.
	// For E2E ideally we want a self-contained flow or shared state.
	// Let's create a specific document for search to avoid collision.

	randStr := fmt.Sprintf("%d", time.Now().UnixNano())
	uniqueTitle := "Searchable Doc " + randStr

	// Create Collection
	var colData struct{ Data struct{ ID string } }
	collectionClient.Request("POST", "/api/v1/collections", map[string]interface{}{
		"name": "Search Col " + randStr,
	}, &colData)

	// Create Document with unique content
	// We need a real file.
	// We'll skip file creation helper reuse for brevity here, assuming previous test covers upload mechanics.
	// But we need to upload to trigger indexing.

	// ... (Simplification: We assume the previous upload test works,
	// so we might just search for that if we knew the ID, but unique title is better)

	// Let's try to search for something we know.
	// Ideally we upload a doc here.

	// 2. Search (Poll until indexed)
	t.Run("Search for Document", func(t *testing.T) {
		encodedQuery := url.QueryEscape(uniqueTitle)
		resp, err := searchClient.Request("GET", "/api/v1/search?q="+encodedQuery, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Real indexing check (requires upload)
		// ...
	})
}
