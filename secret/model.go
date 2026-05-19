package secret

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Secret secret
// swagger:model Secret
type Secret struct {
	// The date and time of the creation
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt"`

	// The secret cannot be reached after this time
	ExpiresAt time.Time `json:"expiresAt,omitempty" bson:"expiresAt"`

	// Unique hash to identify the secrets
	Hash string `json:"hash,omitempty" bson:"hash"`

	// How many times the secret can be viewed
	RemainingViews int32 `json:"remainingViews,omitempty" bson:"remainingViews"`

	// The secret itself
	SecretText string `json:"secretText,omitempty" bson:"secretText"`
}

// DoHash sets Hash to a cryptographically random 128-bit hex string.
func (m *Secret) DoHash() {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand failure is unrecoverable
	m.Hash = hex.EncodeToString(b)
}
