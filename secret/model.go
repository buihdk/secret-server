package secret

import (
	"crypto/sha1"
	"encoding/base64"
	"time"
)

// Secret secret
// swagger:model Secret
type Secret struct {
	// The date and time of the creation
	CreatedAt time.Time `json:"createdAt,omitempty"`

	// The secret cannot be reached after this time
	ExpiresAt time.Time `json:"expiresAt,omitempty"`

	// Unique hash to identify the secrets
	Hash string `json:"hash,omitempty"`

	// How many times the secret can be viewed
	RemainingViews int32 `json:"remainingViews,omitempty"`

	// The secret itself
	SecretText string `json:"secretText,omitempty"`
}

// Hash generate hash field
func (m *Secret) DoHash() {
	h := sha1.New()
	h.Write([]byte(m.SecretText + m.CreatedAt.String()))
	bs := h.Sum(nil)

	m.Hash = base64.URLEncoding.EncodeToString(bs)
}
