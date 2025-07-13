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

// TokenClaims interface defines the common behavior for all claim types
type TokenClaims interface {
    jwt.Claims
    GetIdentifier() string
}

// Add method to UserClaims to implement TokenClaims interface
func (uc *UserClaims) GetIdentifier() string {
    return fmt.Sprintf("%d:%s:%d", 
        uc.ID, 
        uc.Role,
        uc.ExpiresAt.Unix(),
    )
}

// Add method to TableClaims to implement TokenClaims interface
func (tc *TableClaims) GetIdentifier() string {
    return fmt.Sprintf("%d:%d:%s:%d",
        tc.Number,
        tc.Capacity,
        tc.Status,
        tc.ExpiresAt.Unix(),
    )
}

// JWTMaker is a JSON Web Token maker


// CreateTableToken creates a new token for a table
func (maker *JWTMaker) CreateTableToken(number int32, capacity int32, status string, duration time.Duration) (string, *TableClaims, error) {
    claims, err := NewTableClaims(number, capacity, status, duration)
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

// createShortTokenFromJWT creates a shortened version of a JWT token

// CreateShortTableToken creates a shortened version of a table token
func (maker *JWTMaker) CreateShortTableToken(number int32, capacity int32, status string, duration time.Duration) (string, error) {
    tokenString, claims, err := maker.CreateTableToken(number, capacity, status, duration)
    if err != nil {
        return "", fmt.Errorf("error creating table token: %w", err)
    }

    shortToken, err := maker.createShortTokenFromJWTTable(tokenString, claims)
    if err != nil {
        return "", fmt.Errorf("error creating short table token: %w", err)
    }

    return shortToken, nil
}

// VerifyShortTableToken verifies a shortened table token
func (maker *JWTMaker) VerifyShortTableToken(shortToken string) (*TableClaims, error) {
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
    if len(idParts) != 4 {
        return nil, fmt.Errorf("invalid identifier format")
    }

    // Parse the components
    var number, capacity int32
    fmt.Sscanf(idParts[0], "%d", &number)
    fmt.Sscanf(idParts[1], "%d", &capacity)
    status := idParts[2]
    var expireTime int64
    fmt.Sscanf(idParts[3], "%d", &expireTime)

    // Create new claims based on the identifier
    claims := &TableClaims{
        Number:   number,
        Capacity: capacity,
        Status:   status,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Unix(expireTime, 0)),
        },
    }

    // Verify expiration
    if time.Now().After(claims.ExpiresAt.Time) {
        return nil, fmt.Errorf("token has expired")
    }

    return claims, nil
}

// TableClaims represents the claims stored in the JWT token for tables
type TableClaims struct {
    Number   int32  `json:"number"`
    Capacity int32  `json:"capacity"`
    Status   string `json:"status"`
    jwt.RegisteredClaims
}

// NewTableClaims creates a new TableClaims
func NewTableClaims(number int32, capacity int32, status string, duration time.Duration) (*TableClaims, error) {
    tokenID, err := uuid.NewRandom()
    if err != nil {
        return nil, fmt.Errorf("error generating token ID: %w", err)
    }

    now := time.Now()
    claims := &TableClaims{
        Number:   number,
        Capacity: capacity,
        Status:   status,
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        tokenID.String(),
            Subject:   fmt.Sprintf("table_%d", number),
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
        },
    }

    return claims, nil
}

// Implement jwt.Claims interface methods for TableClaims
func (tc *TableClaims) GetExpirationTime() (*jwt.NumericDate, error) {
    return tc.ExpiresAt, nil
}

func (tc *TableClaims) GetIssuedAt() (*jwt.NumericDate, error) {
    return tc.IssuedAt, nil
}

func (tc *TableClaims) GetNotBefore() (*jwt.NumericDate, error) {
    return tc.NotBefore, nil
}

func (tc *TableClaims) GetIssuer() (string, error) {
    return tc.Issuer, nil
}

func (tc *TableClaims) GetSubject() (string, error) {
    return tc.Subject, nil
}

func (tc *TableClaims) GetAudience() (jwt.ClaimStrings, error) {
    return tc.Audience, nil
}

func (tc *TableClaims) GetID() (string, error) {
    return tc.ID, nil
}

func (maker *JWTMaker) createShortTokenFromJWTTable(originalToken string, claims TokenClaims) (string, error) {
    // Get identifier using the interface method
    identifier := claims.GetIdentifier()

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