package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kyei-Ernest/libsystem/services/analytics-service/repository"
	"github.com/Kyei-Ernest/libsystem/shared/response"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsHandler(repo repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

// GetOverview returns total stats
// @Summary      Get overview stats
// @Description  Get total views and downloads
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  map[string]int64 "Overview stats"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /overview [get]
func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	stats, err := h.repo.GetTotalStats()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, stats, "Overview stats")
}

// GetTopDocuments returns top performing documents
// @Summary      Get top documents
// @Description  Get most viewed/downloaded documents
// @Tags         analytics
// @Produce      json
// @Param        limit      query     int     false  "Limit results" default(10)
// @Success      200  {object}  []repository.DocumentStats "Top documents"
// @Failure      400  {object}  response.Response "Invalid limit"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/popular [get]
func (h *AnalyticsHandler) GetTopDocuments(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(c, "Invalid limit")
		return
	}

	stats, err := h.repo.GetTopDocuments(limit)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, stats, "Top documents")
}

// GetActivity returns daily activity
// @Summary      Get daily activity
// @Description  Get views and downloads over time
// @Tags         analytics
// @Produce      json
// @Param        days       query     int     false  "Number of days" default(7)
// @Success      200  {object}  []repository.DailyActivity "Daily activity"
// @Failure      400  {object}  response.Response "Invalid days"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /activity [get]
func (h *AnalyticsHandler) GetActivity(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		response.BadRequest(c, "Invalid days")
		return
	}

	activity, err := h.repo.GetDailyActivity(days)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, activity, "Daily activity")
}

func handleError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}
