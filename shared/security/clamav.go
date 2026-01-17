package security

import (
	"fmt"
	"io"

	"github.com/dutchcoders/go-clamd"
)

// VirusScanner provides virus scanning functionality using ClamAV
type VirusScanner struct {
	client *clamd.Clamd
}

// NewVirusScanner creates a new virus scanner instance
// addr should be in format "tcp://localhost:3310" or "unix:///var/run/clamav/clamd.ctl"
func NewVirusScanner(addr string) (*VirusScanner, error) {
	client := clamd.NewClamd(addr)

	// Test connection
	err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClamAV daemon: %w", err)
	}

	return &VirusScanner{
		client: client,
	}, nil
}

// ScanStream scans a file stream for viruses
// Returns (infected bool, virusName string, error)
func (s *VirusScanner) ScanStream(reader io.Reader) (bool, string, error) {
	response, err := s.client.ScanStream(reader, make(chan bool))
	if err != nil {
		return false, "", fmt.Errorf("scan failed: %w", err)
	}

	for result := range response {
		// ClamAV statuses: OK, FOUND, ERROR
		if result.Status == clamd.RES_FOUND {
			return true, result.Description, nil // Virus found
		}
		if result.Status == clamd.RES_ERROR {
			return false, "", fmt.Errorf("scan error: %s", result.Description)
		}
	}

	return false, "", nil // Clean
}

// ScanFile is a convenience wrapper for scanning files
func (s *VirusScanner) ScanFile(reader io.Reader, filename string) error {
	infected, virus, err := s.ScanStream(reader)
	if err != nil {
		return fmt.Errorf("virus scan failed for %s: %w", filename, err)
	}

	if infected {
		return fmt.Errorf("file %s is infected with: %s", filename, virus)
	}

	return nil
}

// Ping checks if ClamAV daemon is responsive
func (s *VirusScanner) Ping() error {
	return s.client.Ping()
}
