package token

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserClaims represents the claims stored in the JWT token
type UserClaims struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey}
}

// NewUserClaims creates a new UserClaims
func NewUserClaims(id int64, email string, role string, duration time.Duration) (*UserClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("error generating token ID: %w", err)
	}

	now := time.Now()
	claims := &UserClaims{
		ID:    id,
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	return claims, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(id int64, email string, role string, duration time.Duration) (string, *UserClaims, error) {
	claims, err := NewUserClaims(id, email, role, duration)
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("error signing token: %w", err)
	}

	return tokenStr, claims, nil
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token signing method")
		}
		return []byte(maker.secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// CreateShortToken creates a shortened version of the JWT token
func (maker *JWTMaker) CreateShortToken(id int64, email string, role string, duration time.Duration) (string, error) {
	// Generate the original JWT token
	tokenString, claims, err := maker.CreateToken(id, email, role, duration)
	if err != nil {
		return "", fmt.Errorf("error creating token: %w", err)
	}

	// Create the shortened version
	shortToken, err := maker.createShortTokenFromJWT(tokenString, claims)
	if err != nil {
		return "", fmt.Errorf("error creating short token: %w", err)
	}

	return shortToken, nil
}

// createShortTokenFromJWT creates a shortened version of a JWT token
func (maker *JWTMaker) createShortTokenFromJWT(originalToken string, claims *UserClaims) (string, error) {
	// Create a unique identifier using claims
	identifier := fmt.Sprintf("%d:%s:%d",
		claims.ID,
		claims.Role,
		claims.ExpiresAt.Unix(),
	)

	// Hash the original token using the secret key for added security
	hasher := sha256.New()
	hasher.Write([]byte(originalToken))
	hasher.Write([]byte(maker.secretKey))
	hash := hasher.Sum(nil)

	// Take first 8 bytes of the hash
	shortHash := hash[:8]

	// Encode both parts using base64url
	encodedID := base64.RawURLEncoding.EncodeToString([]byte(identifier))
	encodedHash := base64.RawURLEncoding.EncodeToString(shortHash)

	// Combine them
	return fmt.Sprintf("%s.%s", encodedID, encodedHash), nil
}

// VerifyShortToken verifies a shortened token
func (maker *JWTMaker) VerifyShortToken(shortToken string) (*UserClaims, error) {
	parts := strings.Split(shortToken, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid short token format")
	}

	// Decode the identifier part
	identifierBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid identifier encoding: %w", err)
	}

	// Parse the identifier
	idParts := strings.Split(string(identifierBytes), ":")
	if len(idParts) != 3 {
		return nil, fmt.Errorf("invalid identifier format")
	}

	// Create new claims based on the identifier
	// You would typically validate this against your database
	claims := &UserClaims{
		ID:   parseInt64(idParts[0]),
		Role: idParts[1],
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(parseInt64(idParts[2]), 0)),
		},
	}

	// Verify expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// Helper function to parse int64
func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
