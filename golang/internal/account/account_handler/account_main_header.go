package account_handler



import (
	"context"

	"fmt"




account_main "english-ai-full/internal/account"
	pb "english-ai-full/internal/proto_qr/account"

	"github.com/go-playground/validator"
)

var ctx = context.Background()

type Handler struct {
	ctx       context.Context
	user      pb.AccountServiceClient
	validator *validator.Validate
}

// Ensure Handler implements AccountHandlerInterface
var _ account_main.AccountHandlerInterface = (*Handler)(nil)

func New(user pb.AccountServiceClient) Handler {
	return Handler{
		validator: validator.New(),
		ctx:       context.Background(),
		user:      user,
	}
}

// Helper function to extract user ID from JWT token context
func (h Handler) getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}