package account

import (
	"context"
	"errors"

	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"
	logg "english-ai-full/logger"
	"english-ai-full/utils"

	pkgerrors "github.com/pkg/errors"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceStruct struct {
	userRepo *Repository
	logger   *logg.Logger
	account.UnimplementedAccountServiceServer
}

func NewAccountService(userRepo *Repository) *ServiceStruct {
	return &ServiceStruct{
		userRepo: userRepo,
		logger:   logg.NewLogger(),
	}
}

func (s *ServiceStruct) CreateUser(ctx context.Context, req *account.AccountReq) (*account.Account, error) {
	user, err := s.userRepo.CreateUser(ctx, model.Account{
		BranchID: req.BranchId,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     model.Role(req.Role),
		OwnerID:  req.BranchId,
	})
	if err != nil {
		return nil, err
	}

	return &account.Account{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Role:      string(user.Role),
		Avatar:    user.Avatar,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *ServiceStruct) Register(ctx context.Context, req *account.RegisterReq) (*account.RegisterRes, error) {
	user, err := s.userRepo.Register(ctx, model.Account{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &account.RegisterRes{
		Success: true,
		Id:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
	}, nil
}

func (s *ServiceStruct) Login(ctx context.Context, req *account.LoginReq) (*account.AccountRes, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !utils.Compare(user.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	return &account.AccountRes{
		Account: &account.Account{
			Id:        user.ID,
			BranchId:  user.BranchID,
			Name:      user.Name,
			Email:     user.Email,
			Password:  user.Password,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      string(user.Role),
			OwnerId:   user.OwnerID,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *ServiceStruct) FindByEmail(ctx context.Context, req *account.FindByEmailReq) (*account.AccountRes, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, ErrorUserNotFound
		}

		return nil, err
	}

	return &account.AccountRes{Account: &account.Account{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Avatar:    user.Avatar,
		Title:     user.Title,
		Role:      string(user.Role),
		OwnerId:   user.OwnerID,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}}, nil
}

func (s *ServiceStruct) FindByID(ctx context.Context, req *account.FindByIDReq) (*account.FindByIDRes, error) {
	user, err := s.userRepo.FindByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}

		return nil, pkgerrors.WithStack(err)
	}

	return &account.FindByIDRes{Account: &account.Account{
		Id:        user.ID,
		BranchId:  user.BranchID,
		Name:      user.Name,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Title:     user.Title,
		Role:      string(user.Role),
		OwnerId:   user.OwnerID,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}}, nil
}

func (s *ServiceStruct) UpdateUser(ctx context.Context, req *account.UpdateUserReq) (*account.AccountRes, error) {
	user, err := s.userRepo.UpdateUser(ctx, model.Account{
		ID:       req.Id,
		BranchID: req.BranchId,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     model.Role(req.Role),
		OwnerID:  req.OwnerId,
	})
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
	}

	return &account.AccountRes{
		Account: &account.Account{
			Id:        user.ID,
			BranchId:  user.BranchID,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      string(user.Role),
			OwnerId:   user.OwnerID,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *ServiceStruct) DeleteUser(ctx context.Context, req *account.DeleteAccountReq) (*account.DeleteAccountRes, error) {
	err := s.userRepo.DeleteUser(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, ErrorUserNotFound) {
			return nil, pkgerrors.WithStack(ErrorUserNotFound)
		}
		return nil, pkgerrors.WithStack(err)
	}

	return &account.DeleteAccountRes{
		Success: true,
	}, nil
}
