package utils

import (
	"fmt"
	 "english-ai-full/utils/auth"
	"golang.org/x/crypto/bcrypt"
)


var (
	HashPassword = hashPassword
	GenerateJWTTokentest = auth.generateJWTToken
	GenerateRefreshTokentest = auth.generateRefreshToken
)

// Original implementation functions (unexported)
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashed), nil
}

// func CheckPassword(password string, hashedPassword string) error {
// 	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
// }



// func HashPassword(password string) (string, error) {
// 	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return "", fmt.Errorf("error hashing password: %w", err)
// 	}

// 	return string(hashed), nil
// }

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
