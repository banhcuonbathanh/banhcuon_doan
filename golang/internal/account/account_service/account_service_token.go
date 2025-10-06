package account_service

import (
	"context"
	"english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RunDailyCleanup implements the gRPC method for daily token cleanup
func (s *AccountService) RunDailyCleanup(ctx context.Context, req *pb.RunDailyCleanupReq) (*pb.RunDailyCleanupRes, error) {
	s.logger.Info("Starting daily token cleanup job", map[string]interface{}{
		"job": "token_cleanup",
	})

	startTime := time.Now()

	// Run comprehensive cleanup
	err := s.userRepo.ComprehensiveTokenCleanup(ctx)
	
	if err != nil {
		s.logger.Error("Daily token cleanup failed", map[string]interface{}{
			"job":         "token_cleanup",
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		})
		
		return &pb.RunDailyCleanupRes{
			Success:    false,
			Message:    "Daily token cleanup failed: " + err.Error(),
			DurationMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	duration := time.Since(startTime).Milliseconds()
	
	s.logger.Info("Daily token cleanup completed successfully", map[string]interface{}{
		"job":         "token_cleanup",
		"duration_ms": duration,
	})

	return &pb.RunDailyCleanupRes{
		Success:    true,
		Message:    "Daily token cleanup completed successfully",
		DurationMs: duration,
	}, nil
}

// new asdvasdfsdfg
// services/account_service.go

func (s *AccountService) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	const operation = "refresh_token"
	startTime := time.Now()
	
	s.logger.SetOperation(operation)
	
	// Validate request
	if req.RefreshToken == "" {
		s.logger.Error("Refresh token is required", map[string]interface{}{
			"operation": operation,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.InvalidArgument, "Refresh token is required")
	}
	
	// Parse and validate refresh token JWT
	token, err := jwt.ParseWithClaims(req.RefreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	
	if err != nil {
		s.logger.Error("Failed to parse refresh token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}
	
	if !token.Valid {
		s.logger.Error("Invalid refresh token", map[string]interface{}{
			"operation": operation,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}
	
	// Extract claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		s.logger.Error("Failed to extract token claims", map[string]interface{}{
			"operation": operation,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Internal, "Failed to process token")
	}
	
	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		s.logger.Error("Refresh token has expired", map[string]interface{}{
			"operation":  operation,
			"expired_at": claims.ExpiresAt.Time,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Unauthenticated, "Refresh token has expired")
	}
	
	// Extract user ID from claims
	userIDStr := claims.Subject
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.logger.Error("Invalid user ID in token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userIDStr,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Internal, "Invalid token format")
	}
	
	// Verify refresh token exists in repository
	storedToken, err := s.userRepo.GetRefreshTokenByUserIDAndToken(ctx, userID, req.RefreshToken)
	if err != nil {
		s.logger.Error("Failed to retrieve stored refresh token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}
	
	// Check if token is revoked
	if storedToken.IsRevoked {
		s.logger.Error("Refresh token has been revoked", map[string]interface{}{
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Unauthenticated, "Refresh token has been revoked")
	}
	
	// Get user details from repository
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to retrieve user", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.NotFound, "User not found")
	}
	
	// Check if user account is active
	if user.Status != pb.AccountStatus_ACTIVE {
		s.logger.Error("User account is not active", map[string]interface{}{
			"operation": operation,
			"user_id":   userID,
			"status":    user.Status.String(),
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.PermissionDenied, "Account is not active")
	}
	    userAccount := account_dto.Account{
        ID:       user.ID,
        Email:    user.Email,
        Role:     user.Role,
        BranchID: user.BranchID,
    }
	// Generate new access token
	    accessToken, err := s.tokenMaker.CreateToken(userAccount)
	if err != nil {
		s.logger.Error("Failed to generate access token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Internal, "Failed to generate access token")
	}
	
	// Generate new refresh token (token rotation for security)
	newRefreshToken, err := s.tokenMaker.CreateRefreshToken(userAccount)
	if err != nil {
		s.logger.Error("Failed to generate new refresh token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Internal, "Failed to generate refresh token")
	}
	
	// Revoke old refresh token (token rotation)
	if err := s.userRepo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		s.logger.Error("Failed to revoke old refresh token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		// Don't fail the request if revocation fails, but log it
	}
	
	// Save new refresh token to repository
	if err := s.userRepo.SaveRefreshToken(ctx, userID, newRefreshToken, refreshExpiresAt); err != nil {
		s.logger.Error("Failed to save new refresh token", map[string]interface{}{
			"error":     err.Error(),
			"operation": operation,
			"user_id":   userID,
		})
		return &pb.RefreshTokenRes{
			Success: false,
		}, status.Error(codes.Internal, "Failed to save refresh token")
	}
	
	// Log successful token refresh
	s.logger.Info("Token refresh successful", map[string]interface{}{
		"operation":   operation,
		"user_id":     userID,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})
	
	// Convert time.Time to protobuf Timestamp
	expiresAtProto := timestamppb.New(accessExpiresAt)
	
	return &pb.RefreshTokenRes{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAtProto,
	}, nil
}

// Helper function to generate access token
func (s *AccountService) generateAccessToken(userID int64, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenTTL) // e.g., 15 minutes
	
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "english-ai-service",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	
	return tokenString, expiresAt, nil
}

// Helper function to generate refresh token
func (s *AccountService) generateRefreshToken(userID int64) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.refreshTokenTTL) // e.g., 7 days
	
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "english-ai-service",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	
	return tokenString, expiresAt, nil
}
// new adasdfsdfs