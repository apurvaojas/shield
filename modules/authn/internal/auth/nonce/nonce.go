package nonce

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// NonceValidator handles nonce generation and validation
// Used for CSRF protection and one-time use tokens
type NonceValidator interface {
	Generate(ctx context.Context) (string, error)
	Validate(ctx context.Context, nonce string) error
	Cleanup(ctx context.Context) error
}

// InMemoryNonceValidator is an in-memory implementation of NonceValidator
// For production, consider using Redis or database storage
type InMemoryNonceValidator struct {
	nonces map[string]time.Time
	mutex  sync.RWMutex
	ttl    time.Duration
}

// NewInMemoryNonceValidator creates a new in-memory nonce validator
func NewInMemoryNonceValidator(ttl time.Duration) *InMemoryNonceValidator {
	validator := &InMemoryNonceValidator{
		nonces: make(map[string]time.Time),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go validator.cleanupExpired()

	return validator
}

// Generate creates a new nonce
func (v *InMemoryNonceValidator) Generate(ctx context.Context) (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	nonce := hex.EncodeToString(bytes)

	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.nonces[nonce] = time.Now().Add(v.ttl)

	return nonce, nil
}

// Validate checks if a nonce is valid and removes it (one-time use)
func (v *InMemoryNonceValidator) Validate(ctx context.Context, nonce string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	expiry, exists := v.nonces[nonce]
	if !exists {
		return fmt.Errorf("invalid nonce")
	}

	// Remove nonce (one-time use)
	delete(v.nonces, nonce)

	if time.Now().After(expiry) {
		return fmt.Errorf("nonce expired")
	}

	return nil
}

// Cleanup removes expired nonces
func (v *InMemoryNonceValidator) Cleanup(ctx context.Context) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	now := time.Now()
	for nonce, expiry := range v.nonces {
		if now.After(expiry) {
			delete(v.nonces, nonce)
		}
	}

	return nil
}

// cleanupExpired runs periodically to clean up expired nonces
func (v *InMemoryNonceValidator) cleanupExpired() {
	ticker := time.NewTicker(v.ttl / 2) // Cleanup every half TTL
	defer ticker.Stop()

	for range ticker.C {
		_ = v.Cleanup(context.Background())
	}
}
