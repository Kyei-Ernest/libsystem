package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	Addresses []string
	Username  string
	Password  string
}

// NewClient creates a new Elasticsearch typed client
func NewClient(cfg Config) (*elasticsearch.TypedClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	client, err := elasticsearch.NewTypedClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %w", err)
	}

	return client, nil
}
