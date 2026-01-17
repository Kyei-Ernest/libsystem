package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type SearchService interface {
	Search(query string, page, pageSize int) (*SearchResult, error)
}

type searchService struct {
	client *elasticsearch.TypedClient
}

type SearchResult struct {
	Hits     []models.Document           `json:"hits"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"page_size"`
	Facets   map[string]map[string]int64 `json:"facets,omitempty"`
}

func NewSearchService(client *elasticsearch.TypedClient) SearchService {
	return &searchService{client: client}
}

func (s *searchService) Search(query string, page, pageSize int) (*SearchResult, error) {
	from := (page - 1) * pageSize

	// Default query if empty: match all
	var q *types.Query
	if query == "" {
		q = &types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	} else {
		// Multi-match query on title, description, and content
		q = &types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:     query,
				Fields:    []string{"title^2", "description", "content"}, // Boost title
				Fuzziness: "AUTO",
			},
		}
	}

	// Execute Search with Aggregations
	res, err := s.client.Search().
		Index("documents").
		Query(q).
		From(from).
		Size(pageSize).
		Aggregations(map[string]types.Aggregations{
			"file_types": {
				Terms: &types.TermsAggregation{
					Field: some("file_type.keyword"),
				},
			},
			"statuses": {
				Terms: &types.TermsAggregation{
					Field: some("status.keyword"),
				},
			},
			"collections": {
				Terms: &types.TermsAggregation{
					Field: some("collection_id.keyword"),
				},
			},
		}).
		Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Map results
	hits := []models.Document{}
	// Iterate through hits
	// Note: In Typed API, Hits is a struct with Hits field which is a slice of Hit
	for _, hit := range res.Hits.Hits {
		var doc models.Document
		if hit.Source_ != nil {
			if err := json.Unmarshal(hit.Source_, &doc); err == nil {
				hits = append(hits, doc)
			}
		}
	}

	// Parse aggregations
	facets := make(map[string]map[string]int64)
	if res.Aggregations != nil {
		facets["file_types"] = parseTermsAgg(res.Aggregations["file_types"])
		facets["statuses"] = parseTermsAgg(res.Aggregations["statuses"])
		facets["collections"] = parseTermsAgg(res.Aggregations["collections"])
	}

	return &SearchResult{
		Hits:     hits,
		Total:    res.Hits.Total.Value,
		Page:     page,
		PageSize: pageSize,
		Facets:   facets,
	}, nil
}

func some(s string) *string {
	return &s
}

func parseTermsAgg(agg types.Aggregate) map[string]int64 {
	result := make(map[string]int64)

	// Helper struct for JSON extraction
	type Bucket struct {
		Key      interface{} `json:"key"`
		DocCount int64       `json:"doc_count"`
	}
	type TermsAgg struct {
		Buckets []Bucket `json:"buckets"`
	}

	// Marshal back to JSON to unmarshal into our simple struct
	// This avoids fighting with the generated Typed API union types
	data, err := json.Marshal(agg)
	if err != nil {
		return result
	}

	var terms TermsAgg
	if err := json.Unmarshal(data, &terms); err != nil {
		return result
	}

	for _, bucket := range terms.Buckets {
		if keyStr, ok := bucket.Key.(string); ok {
			result[keyStr] = bucket.DocCount
		}
	}

	return result
}
