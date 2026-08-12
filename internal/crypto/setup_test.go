package crypto

import (
	"os"
	"testing"
)

// TestMain makes sure an ENCRYPTION_KEY exists for the crypto tests.
// The library now requires a real key and no longer falls back to a
// hardcoded default. Set one here if the environment does not provide it.
func TestMain(m *testing.M) {
	if os.Getenv("ENCRYPTION_KEY") == "" {
		os.Setenv("ENCRYPTION_KEY", "test-encryption-key")
	}
	os.Exit(m.Run())
}
