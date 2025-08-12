// Add this to your account_handler_main.go file to ensure interface compliance

package account_handler

import (
	"english-ai-full/internal/account"
	pb "english-ai-full/internal/proto_qr/account"
	utils_config "english-ai-full/utils/config"
)

// Compile-time interface compliance check
var _ account.AccountHandlerInterface = (*AccountHandler)(nil)

type AccountHandler struct {
	*BaseAccountHandler
}

func NewAccountHandler(userClient pb.AccountServiceClient, cfg *utils_config.Config) *AccountHandler {
	return &AccountHandler{
		BaseAccountHandler: NewBaseHandler(userClient, cfg),
	}
}

