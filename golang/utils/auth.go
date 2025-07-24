package utils

import (
	"english-ai-full/internal/model"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)
var (

	GenerateJWTToken     = generateJWTToken
	GenerateRefreshToken = generateRefreshToken
)

func Compare(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func generateJWTToken(user model.Account) (string, error) {
	claims := jwt.MapClaims{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	value, err := token.SignedString([]byte(AppConfig.JwtSecret))
	if err != nil {
		return "", err
	}

	return value, err
}

func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(AppConfig.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

func generateRefreshToken(user model.Account) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"exp":        time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 ngày
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	value, err := token.SignedString([]byte(AppConfig.JwtSecret))
	if err != nil {
		return "", err
	}
	return value, err
}


// new 2112121212121212



type JWTTokenMaker struct {
	secretKey                string
	accessTokenDuration      time.Duration
	refreshTokenDuration     time.Duration
	resetTokenDuration       time.Duration
	verificationTokenDuration time.Duration
}

type JWTClaims struct {
	UserID   int64       `json:"user_id"`
	Email    string      `json:"email"`
	Role     model.Role  `json:"role"`
	BranchID int64       `json:"branch_id"`
	jwt.RegisteredClaims
}

func NewJWTTokenMaker(secretKey string) *JWTTokenMaker {
	return &JWTTokenMaker{
		secretKey:                secretKey,
		accessTokenDuration:      15 * time.Minute,    // 15 minutes
		refreshTokenDuration:     7 * 24 * time.Hour,  // 7 days
		resetTokenDuration:       1 * time.Hour,       // 1 hour
		verificationTokenDuration: 24 * time.Hour,     // 24 hours
	}
}

func (maker *JWTTokenMaker) CreateToken(user model.Account) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		BranchID: user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) VerifyToken(tokenString string) (*model.Account, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &model.Account{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		BranchID: claims.BranchID,
	}, nil
}

func (maker *JWTTokenMaker) CreateRefreshToken(user model.Account) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		BranchID: user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) ValidateRefreshToken(tokenString string) (*model.Account, error) {
	return maker.VerifyToken(tokenString) // Same validation logic
}

func (maker *JWTTokenMaker) CreateResetToken(email string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.resetTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) ValidateResetToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.Subject, nil
}

func (maker *JWTTokenMaker) CreateVerificationToken(email string) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.verificationTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) ValidateVerificationToken(tokenString string) (string, error) {
	return maker.ValidateResetToken(tokenString) // Same validation logic
}
// new done 121212121212