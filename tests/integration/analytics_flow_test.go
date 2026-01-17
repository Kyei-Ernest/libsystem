package integration

import (
	"net/http"
	"testing"

	"github.com/Kyei-Ernest/libsystem/services/api-gateway/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsFlow(t *testing.T) {
	client := helpers.NewClient(cfg.AnalyticsAddr)

	// 1. Get Overview
	t.Run("Get Analytics Overview", func(t *testing.T) {
		var respData struct {
			Success bool `json:"success"`
			Data    struct {
				TotalViews     int64 `json:"total_views"`
				TotalDownloads int64 `json:"total_downloads"`
			} `json:"data"`
		}
		resp, err := client.Request("GET", "/api/v1/analytics/overview", nil, &respData)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, respData.Success)
		// We can't assert exact numbers without knowing initial state, but response structure is valid.
	})
}
