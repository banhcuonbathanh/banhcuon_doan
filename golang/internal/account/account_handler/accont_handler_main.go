// Add this to your account_handler_main.go file to ensure interface compliance

package account_handler

import (
	"english-ai-full/internal/account"
	pb "english-ai-full/internal/proto_qr/account"
)

// Compile-time interface compliance check
var _ account.AccountHandlerInterface = (*AccountHandler)(nil)

type AccountHandler struct {
	*BaseAccountHandler
}

func NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler {
	return &AccountHandler{
		BaseAccountHandler: NewBaseHandler(userClient),
	}
}

