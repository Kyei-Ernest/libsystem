package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Kyei-Ernest/libsystem/services/search-service/service"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService service.SearchService
}

func NewSearchHandler(s service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: s}
}

// Search performs a full-text search
// @Summary      Search documents
// @Description  Search for documents by query string
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        q          query     string  false "Query string"
// @Param        page       query     int     false "Page number" default(1)
// @Param        page_size  query     int     false "Page size" default(10)
// @Success      200  {object}  service.SearchResult "Search results"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       / [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	log.Printf("DEBUG: Search Handler called with query: '%s'", query)
	result, err := h.searchService.Search(query, page, pageSize)
	if err != nil {
		log.Printf("DEBUG: Search Service failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Search found results")

	// Wrap in standard response format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// AdvancedSearch performs an advanced search (placeholder)
// @Summary      Advanced search
// @Description  Advanced search options (currently alias for basic search)
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        q          query     string  false "Query string"
// @Success      200  {object}  service.SearchResult "Search results"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /advanced [get]
func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	// Placeholder for advanced search (filtering, faceting)
	// Currently reuses basic search
	h.Search(c)
}
