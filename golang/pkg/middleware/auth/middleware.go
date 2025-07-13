package auth

import (
	"context"
	httpserve "english-ai-full/pkg/pkg/http"
	"english-ai-full/utils"
	"fmt"
	"log"
	"net/http"
	"strings"

	"english-ai-full/token"
)

type AuthKey struct{}

type TableAuthKey struct{}

type Role string

const (
	RoleGuest    Role = "guest"
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleOwner    Role = "owner"
)

func isRoleAllowed(userRole Role, allowedRoles []Role) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func verifyClaimsFromAuthHeader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header is missing")
	}

	fields := strings.Fields(authHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header")
	}

	token := fields[1]
	log.Printf("Token received: %s", token) // Log the token (be careful with this in production)
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		log.Printf("Error verifying token: %v", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("Claims verified: %+v", claims)
	return claims, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			httpserve.ErrorHandler(w, http.StatusUnauthorized, ErrUnauthorized, "Unauthorized")
			return
		}

		bearToken := strings.Split(authHeader, "Bearer ")
		if len(bearToken) != 2 {
			httpserve.ErrorHandler(w, http.StatusUnauthorized, ErrInvalidToken, "Invalid token")
			return
		}

		token := bearToken[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			httpserve.ErrorHandler(w, http.StatusUnauthorized, ErrInvalidToken, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
