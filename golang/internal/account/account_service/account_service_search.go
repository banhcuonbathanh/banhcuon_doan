// internal/account/account_service_search.go
package account_service

import (
	"context"
	"fmt"

	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/model"
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


// GetUserProfile handles user profile requests (current user or specific user)
func (s *ServiceStruct) GetUserProfile(ctx context.Context, req *account.FindByIDReq) (*account.FindByIDRes, error) {
	s.logServiceCall("GetUserProfile", map[string]interface{}{
		"user_id": req.Id,
	})

	// If no user ID provided, get current user from context
	var targetUserID int64
	if req.Id != 0 {
		targetUserID = req.Id
	} else {
		// Get current user ID from context (JWT token, session, etc.)
		currentUserID, err := s.getCurrentUserFromContext(ctx)
		if err != nil {
			return nil, s.handleServiceError("GetUserProfile", "Failed to get current user from context", err, false)
		}
		targetUserID = currentUserID
	}

	// Find user by ID
	user, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, s.handleRepositoryError(err, "find_user_profile", "users", map[string]interface{}{
			"user_id": targetUserID,
		})
	}

	return &account.FindByIDRes{
		Account: s.modelToProto(user),
	}, nil
}

// Alternative: GetCurrentUserProfile for getting current user only
func (s *ServiceStruct) GetCurrentUserProfile(ctx context.Context, req *emptypb.Empty) (*account.FindByIDRes, error) {
	s.logServiceCall("GetCurrentUserProfile", map[string]interface{}{})

	// Get current user ID from context (JWT token, session, etc.)
	currentUserID, err := s.getCurrentUserFromContext(ctx)
	if err != nil {
		return nil, s.handleServiceError("GetCurrentUserProfile", "Failed to get current user from context", err, false)
	}

	// Find user by ID
	user, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil {
		return nil, s.handleRepositoryError(err, "find_current_user_profile", "users", map[string]interface{}{
			"user_id": currentUserID,
		})
	}

	return &account.FindByIDRes{
		Account: s.modelToProto(user),
	}, nil
}



func (s *ServiceStruct) GetUsersByBranch(ctx context.Context, req *account.FindByBranchReq) (*account.AccountList, error) {
	s.logServiceCall("GetUsersByBranch", map[string]interface{}{
		"branch_id": req.BranchId,
		"page":      req.Pagination.Page,
		"page_size": req.Pagination.PageSize,
	})

	// Set default pagination if not provided
	page := int32(1)
	pageSize := int32(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}

	// Calculate offset for database query
	offset := (page - 1) * pageSize

	// Get users by branch ID with pagination
	users, total, err := s.userRepo.FindByBranchWithPagination(ctx, req.BranchId, int(offset), int(pageSize))
	if err != nil {
		return nil, s.handleRepositoryError(err, "find_users_by_branch", "users", map[string]interface{}{
			"branch_id": req.BranchId,
			"page":      page,
			"page_size": pageSize,
		})
	}

	// Convert model accounts to protobuf accounts
	var protoAccounts []*account.Account
	for _, user := range users {
		protoAccounts = append(protoAccounts, s.modelToProto(user))
	}

	// Calculate pagination info
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	hasNext := page < totalPages
	hasPrev := page > 1

	return &account.AccountList{
		Accounts: protoAccounts,
		Total:    int32(total),
		Pagination: &account.PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}



// getCurrentUserFromContext extracts current user ID from context
// This implementation depends on your authentication middleware
func (s *ServiceStruct) getCurrentUserFromContext(ctx context.Context) (int64, error) {
	// Option 1: If you store user ID directly in context
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(int64); ok {
			return id, nil
		}
		if id, ok := userID.(int); ok {
			return int64(id), nil
		}
	}

	// Option 2: If you store user claims/token in context
	if claims := ctx.Value("user_claims"); claims != nil {
		// Extract user ID from JWT claims
		// This depends on your JWT token structure
		// Example implementation:
		/*
		if userClaims, ok := claims.(*TokenClaims); ok {
			return userClaims.UserID, nil
		}
		*/
	}

	// Option 3: If you store user object in context
	if user := ctx.Value("user"); user != nil {
		if userObj, ok := user.(*model.Account); ok {
			return userObj.ID, nil
		}
	}

	return 0, errorcustom.NewAuthenticationError("No user found in context")
}

// ========================================
// Update account_service_main.go - Enhanced GetUsersByBranch
// ========================================

// Replace existing GetUsersByBranch business logic method with:
// GetUsersByBranchSimple - simplified version for internal use
func (s *ServiceStruct) GetUsersByBranchSimple(ctx context.Context, branchID int64) ([]model.Account, error) {
	return s.userRepo.FindByBranchID(ctx, branchID)
}


