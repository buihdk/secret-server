package crypto

import (
	"os"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	original := "my secret message"

	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == original {
		t.Fatal("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != original {
		t.Fatalf("expected %q, got %q", original, decrypted)
	}
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	enc1, _ := Encrypt("same input")
	enc2, _ := Encrypt("same input")
	if enc1 == enc2 {
		t.Fatal("two encryptions of the same plaintext should differ (random nonce)")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "key-one")
	encrypted, err := Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	os.Setenv("ENCRYPTION_KEY", "key-two")
	defer os.Unsetenv("ENCRYPTION_KEY")

	_, err = Decrypt(encrypted)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	encrypted, err := Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt failed for empty string: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}
