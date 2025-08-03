// internal/account/account_service_search.go
package account_service

import (
	"context"
	"fmt"

	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/proto_qr/account"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FindByEmail finds user by email address
func (s *ServiceStruct) FindByEmail(ctx context.Context, req *account.FindByEmailReq) (*account.AccountRes, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			// Return the custom error that your handler can detect
			return nil, errorcustom.NewUserNotFoundByEmail(req.Email)
		}
		// For other repository errors, wrap with context
		return nil, pkgerrors.WithStack(err)
	}

	return &account.AccountRes{Account: &account.Account{
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
// FindByID finds user by ID
func (s *ServiceStruct) FindByID(ctx context.Context, req *account.FindByIDReq) (*account.FindByIDRes, error) {
	user, err := s.userRepo.FindByID(ctx, req.Id)
	if err != nil {
		// Check if it's a user not found error using string matching
		if strings.Contains(err.Error(), "not found") {
			// Return the custom error that your handler can detect
			return nil, errorcustom.NewUserNotFoundByID(req.Id)
		}
		// For other repository errors, wrap with context
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

// FindByEmail finds user by email


// FindAllUsers retrieves all users
func (s *ServiceStruct) FindAllUsers(ctx context.Context, req *emptypb.Empty) (*account.AccountList, error) {
	users, err := s.userRepo.FindAllUsers(ctx)
	if err != nil {
		// Create a service error for repository failures
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"FindAllUsers",
			"Failed to retrieve all users",
			err,
			false, // Not retryable for general repository errors
		)
		return nil, serviceErr
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

// FindByRole finds users by role
func (s *ServiceStruct) FindByRole(ctx context.Context, req *account.FindByRoleReq) (*account.AccountList, error) {
	users, err := s.userRepo.FindByRole(ctx, req.Role)
	if err != nil {
		// Create a service error for repository failures
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"FindByRole",
			fmt.Sprintf("Failed to retrieve users by role: %s", req.Role),
			err,
			false, // Not retryable for general repository errors
		)
		return nil, serviceErr
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

// FindByBranch finds users by branch ID
func (s *ServiceStruct) FindByBranch(ctx context.Context, req *account.FindByBranchReq) (*account.AccountList, error) {
	users, err := s.userRepo.FindByBranchID(ctx, req.BranchId)
	if err != nil {
		// Create a service error for repository failures
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"FindByBranch",
			fmt.Sprintf("Failed to retrieve users by branch ID: %d", req.BranchId),
			err,
			false, // Not retryable for general repository errors
		)
		return nil, serviceErr
	}

	var accountList []*account.Account
	for _, user := range users {
		accountList = append(accountList, &account.Account{
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
		})
	}

	return &account.AccountList{
		Accounts: accountList,
		Total:    int32(len(accountList)),
	}, nil
}

// SearchUsers performs advanced user search with filtering and pagination
func (s *ServiceStruct) SearchUsers(ctx context.Context, req *account.SearchUsersReq) (*account.SearchUsersRes, error) {
	// Extract pagination info
	var page, pageSize int32
	if req.Pagination != nil {
		page = req.Pagination.Page
		pageSize = req.Pagination.PageSize
	} else {
		// Default values if pagination is not provided
		page = 1
		pageSize = 10
	}

	// Extract sort info
	var sortBy, sortOrder string
	if req.Sort != nil {
		sortBy = req.Sort.SortBy
		sortOrder = req.Sort.SortOrder
	}

	users, totalCount, err := s.userRepo.SearchUsers(ctx, req.Query, req.Role, req.BranchId, req.StatusFilter, page, pageSize, sortBy, sortOrder)
	if err != nil {
		// Check if it's a connection/timeout issue that might be retryable
		isRetryable := strings.Contains(err.Error(), "connection") || 
					 strings.Contains(err.Error(), "timeout") ||
					 strings.Contains(err.Error(), "unavailable")
		
		// Create a service error for repository failures
		serviceErr := errorcustom.NewServiceError(
			"AccountService",
			"SearchUsers",
			"Failed to search users",
			err,
			isRetryable,
		)
		return nil, serviceErr
	}

	var accounts []*account.Account
	// Use index-based loop or range over pointers to avoid copying the struct
	for i := range users {
		user := &users[i] // Get pointer to avoid copying
		accounts = append(accounts, &account.Account{
			Id:        user.Id,
			BranchId:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerId:   user.OwnerId,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// Calculate pagination info
	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize)) // Ceiling division
	hasNext := page < totalPages
	hasPrev := page > 1

	// Create pagination info
	paginationInfo := &account.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	// Return SearchUsersRes with the accounts slice directly
	return &account.SearchUsersRes{
		Accounts:   accounts,
		Total:      int32(totalCount),
		Pagination: paginationInfo,
	}, nil
}