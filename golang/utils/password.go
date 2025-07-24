package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)


var (
	HashPassword = hashPassword
	// GenerateJWTTokentest = auth.generateJWTToken
	// GenerateRefreshTokentest = auth.generateRefreshToken
)

// Original implementation functions (unexported)
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashed), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// new 12121212
// File: golang/internal/auth/bcrypt_password_hasher.go




type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost, // Cost of 10
	}
}

func NewBcryptPasswordHasherWithCost(cost int) *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		cost: cost,
	}
}

func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (h *BcryptPasswordHasher) ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
// new 1212121
