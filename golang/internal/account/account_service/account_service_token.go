package account_service

import (
	"context"
	pb "english-ai-full/internal/proto_qr/account"
	"time"
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