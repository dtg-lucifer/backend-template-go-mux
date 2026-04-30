package pkg

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost // 10 rounds — increase to 12 for higher security

// HashPassword hashes a plaintext password using bcrypt.
// The returned string is safe to store directly in the database.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hashed), nil
}

// ComparePassword returns nil if plain matches the stored bcrypt hash, or an error otherwise.
func ComparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
