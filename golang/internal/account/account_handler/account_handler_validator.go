package account_handler

import (
	"context"
	"log"

	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidatePassword validates password using utils function
func ValidatePassword(fl validator.FieldLevel) bool {
	return utils.ValidatePassword(fl.Field().String()) == nil
}

// ValidateRole validates if the role is valid
func ValidateRole(fl validator.FieldLevel) bool {
	validRoles := map[string]bool{
		"admin":   true,
		"user":    true,
		"manager": true,
	}
	return validRoles[fl.Field().String()]
}

// ValidateEmailUnique checks if email is unique using the gRPC client
func ValidateEmailUnique(userClient pb.AccountServiceClient) validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, err := userClient.FindByEmail(context.Background(), &pb.FindByEmailReq{
			Email: fl.Field().String(),
		})

		if err == nil {
			return false // Email exists
		}

		if status.Code(err) == codes.NotFound {
			return true // Email is unique
		}

		log.Printf("Email uniqueness check error: %v", err)
		return true
	}
}