package mapping

import (
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
)

func ToPBUserRes(u *pb.AccountRes) model.Account {
	return model.Account{
		ID:        u.Account.Id,
		BranchID:  u.Account.BranchId,
		Name:      u.Account.Name,
		Email:     u.Account.Email,
		Avatar:    u.Account.Avatar,
		Title:     u.Account.Title,
		Role:      model.Role(u.Account.Role),
		OwnerID:   u.Account.OwnerId,
		CreatedAt: u.Account.CreatedAt.AsTime(),
		UpdatedAt: u.Account.UpdatedAt.AsTime(),
	}
}
