package account_handler

import (
	pb "english-ai-full/internal/proto_qr/account"
)

type AccountHandler struct {
	*BaseAccountHandler
}

func NewAccountHandler(userClient pb.AccountServiceClient) *AccountHandler {
	return &AccountHandler{
		BaseAccountHandler: NewBaseHandler(userClient),
	}
}