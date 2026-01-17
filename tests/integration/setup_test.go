package integration

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/api-gateway/tests/integration/helpers"
)

var (
	cfg *helpers.Config
)

func TestMain(m *testing.M) {
	// Setup
	log.Println("Starting Integration Tests...")
	cfg = helpers.LoadConfig()

	// Wait for services to be potentially ready (optional, but good practice in CI)
	// In local dev, we assume they are running.
	time.Sleep(1 * time.Second)

	// Run Tests
	code := m.Run()

	// Teardown
	log.Println("Integration Tests Finished.")
	os.Exit(code)
}
