package secret

import (
	"testing"
)

func TestDoHashPopulatesHash(t *testing.T) {
	s := &Secret{SecretText: "test secret"}
	s.DoHash()

	if s.Hash == "" {
		t.Fatal("DoHash should set Hash to a non-empty string")
	}
	// 16 random bytes hex-encoded = 32 characters
	if len(s.Hash) != 32 {
		t.Fatalf("expected hash length 32, got %d", len(s.Hash))
	}
}

func TestDoHashProducesUniqueValues(t *testing.T) {
	s1 := &Secret{SecretText: "same text"}
	s2 := &Secret{SecretText: "same text"}

	s1.DoHash()
	s2.DoHash()

	if s1.Hash == s2.Hash {
		t.Fatal("hashes for identical inputs should differ (crypto/rand)")
	}
}

func TestDoHashIsHex(t *testing.T) {
	s := &Secret{}
	s.DoHash()

	for _, c := range s.Hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("hash contains non-hex character: %q", c)
		}
	}
}
