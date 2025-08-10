package account_handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dto "english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FindAccountByID handles finding user by ID with comprehensive logging
func (h *AccountHandler) FindAccountByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("find_account_by_id")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Find account by ID request started", baseContext)

	// Parse ID parameter
	id, apiErr := errorcustom.ParseIDParam(r, "id")
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid ID parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_id",
			context,
		)
		
		// Log API request with error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Add user ID to context
	baseContext["user_id"] = id
	handlerLog.Debug("User ID parsed successfully", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.FindByID(r.Context(), &pb.FindByIDReq{Id: id})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindByID",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		
		if strings.Contains(err.Error(), "not found") {
			httpStatus = http.StatusNotFound
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User with the specified ID was not found",
				httpStatus,
				"handler",
				"find_account_by_id",
				err,
			).WithDetail("user_id", id)
			
			logger.WarningWithCause(
				"User not found",
				"user_not_found",
				logger.LayerHandler,
				"find_account_by_id",
				baseContext,
			)
		} else {
			httpStatus = http.StatusInternalServerError
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Failed to retrieve user",
				httpStatus,
				"handler",
				"find_account_by_id",
				err,
			).WithDetail("user_id", id)
			
			logger.ErrorWithCause(
				"Service call failed",
				"service_error",
				logger.LayerExternal,
				"find_account_by_id",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindByID", false, err, serviceContext)
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, clientError, "find_account_by_id")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindByID", true, nil, serviceContext)

	// Log performance
	logger.LogPerformance("find_account_by_id", time.Since(start), baseContext)

	handlerLog.Info("Account found successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"account_email": res.Account.Email,
		"account_role":  res.Account.Role,
	}))

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), baseContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.FindAccountByIDResponse{
		ID:        id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	}, "find_account_by_id")
}

// GetUserProfile handles getting user profile with enhanced logging
func (h *AccountHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("get_user_profile")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Get user profile request started", baseContext)

	// Parse ID parameter or get from context
	idStr, _ := errorcustom.GetStringParam(r, "id", 0)
	var id int64
	var err error

	if idStr == "" {
		// Get user ID from JWT context
		userID, contextErr := h.getUserIDFromContext(ctx)
		if contextErr != nil {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"error": contextErr.Error(),
			})
			
			logger.ErrorWithCause(
				"Failed to get user ID from context",
				"context_extraction_error",
				logger.LayerHandler,
				"get_user_id_from_context",
				context,
			)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusUnauthorized, time.Since(start), context)
			
			errorcustom.HandleError(w, contextErr, "get_user_profile")
			return
		}
		id = userID
		baseContext["source"] = "jwt_context"
	} else {
		id, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"provided_id": idStr,
				"error":       err.Error(),
			})
			
			logger.ErrorWithCause(
				"Invalid user ID format",
				"invalid_parameter",
				logger.LayerHandler,
				"parse_user_id",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidInput,
				"Invalid user ID format",
				http.StatusBadRequest,
				"handler",
				"get_user_profile",
				err,
			).WithDetail("provided_id", idStr)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "get_user_profile")
			return
		}
		baseContext["source"] = "url_parameter"
	}

	// Add user ID to context
	baseContext["user_id"] = id
	handlerLog.Debug("User ID determined", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.FindByID(ctx, &pb.FindByIDReq{Id: id})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindByID",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		
		if strings.Contains(err.Error(), "user not found") {
			httpStatus = http.StatusNotFound
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User profile not found",
				httpStatus,
				"handler",
				"get_user_profile",
				err,
			).WithDetail("user_id", id)
			
			logger.WarningWithCause(
				"User profile not found",
				"user_not_found",
				logger.LayerHandler,
				"get_user_profile",
				baseContext,
			)
		} else {
			httpStatus = http.StatusInternalServerError
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"Failed to retrieve user profile",
				httpStatus,
				"handler",
				"get_user_profile",
				err,
			).WithDetail("user_id", id)
			
			logger.ErrorWithCause(
				"Service call failed",
				"service_error",
				logger.LayerExternal,
				"get_user_profile",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindByID", false, err, serviceContext)
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, clientError, "get_user_profile")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindByID", true, nil, serviceContext)

	// Log performance
	logger.LogPerformance("get_user_profile", time.Since(start), baseContext)

	userProfile := dto.UserProfile{
		ID:       res.Account.Id,
		BranchID: res.Account.BranchId,
		Name:     res.Account.Name,
		Email:    res.Account.Email,
		Avatar:   res.Account.Avatar,
		Title:    res.Account.Title,
		Role:     res.Account.Role,
		OwnerID:  res.Account.OwnerId,
	}

	handlerLog.Info("User profile retrieved successfully", utils.MergeContext(baseContext, map[string]interface{}{
		"profile_email": res.Account.Email,
		"profile_role":  res.Account.Role,
	}))

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), baseContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.UserProfileResponse{
		User: userProfile,
	}, "get_user_profile")
}

// FindAllUsers handles getting all users with pagination and comprehensive logging
func (h *AccountHandler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("find_all_users")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Find all users request started", baseContext)

	// Parse pagination parameters
	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid pagination parameters",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_pagination",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Add pagination info to context
	paginationContext := utils.MergeContext(baseContext, map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	})

	handlerLog.Debug("Pagination parameters parsed", paginationContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.FindAllUsers(ctx, &emptypb.Empty{})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(paginationContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindAllUsers",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		logger.ErrorWithCause(
			"Service call failed",
			"service_error",
			logger.LayerExternal,
			"find_all_users",
			utils.MergeContext(paginationContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindAllUsers", false, err, serviceContext)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users",
			http.StatusInternalServerError,
			"handler",
			"find_all_users",
			err,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), paginationContext)
		
		errorcustom.HandleError(w, clientError, "find_all_users")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindAllUsers", true, nil, serviceContext)

	// Apply pagination manually since the proto doesn't support it
	users := res.Accounts
	totalCount := int64(len(users))
	
	start_idx := (page - 1) * pageSize
	end_idx := start_idx + pageSize
	
	if start_idx >= int32(len(users)) {
		users = []*pb.Account{}
	} else {
		if end_idx > int32(len(users)) {
			end_idx = int32(len(users))
		}
		users = users[start_idx:end_idx]
	}

	// Convert to response format
	var userResponses []dto.UserProfile
	for _, user := range users {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	// Final context with results
	resultContext := utils.MergeContext(paginationContext, map[string]interface{}{
		"total_count":    totalCount,
		"returned_count": len(userResponses),
		"total_pages":    (totalCount + int64(pageSize) - 1) / int64(pageSize),
	})

	// Log performance
	logger.LogPerformance("find_all_users", time.Since(start), resultContext)

	handlerLog.Info("All users retrieved successfully", resultContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), resultContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":       userResponses,
		"total_count": totalCount,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (totalCount + int64(pageSize) - 1) / int64(pageSize),
	}, "find_all_users")
}

// FindByRole handles finding users by role with comprehensive logging
func (h *AccountHandler) FindByRole(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("find_by_role")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Find by role request started", baseContext)

	// Parse role parameter
	role, apiErr := errorcustom.GetStringParam(r, "role", 1)
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid role parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_role",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Add role to context
	baseContext["role"] = role

	// Validate role
	var req struct {
		Role string `validate:"required,role"`
	}
	req.Role = role

	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Role validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_role",
				context,
			)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "find_by_role")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error",
				"validation_system_error",
				logger.LayerHandler,
				"validate_role",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeValidationError,
				"Invalid role",
				http.StatusBadRequest,
				"handler",
				"find_by_role",
				err,
			).WithDetail("role", role)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "find_by_role")
		}
		return
	}

	handlerLog.Debug("Role validation passed", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.FindByRole(ctx, &pb.FindByRoleReq{
		Role: role,
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindByRole",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		logger.ErrorWithCause(
			"Service call failed",
			"service_error",
			logger.LayerExternal,
			"find_by_role",
			utils.MergeContext(baseContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindByRole", false, err, serviceContext)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users by role",
			http.StatusInternalServerError,
			"handler",
			"find_by_role",
			err,
		).WithDetail("role", role)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, clientError, "find_by_role")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindByRole", true, nil, serviceContext)

	var userResponses []dto.UserProfile
	for _, user := range res.Accounts {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	// Result context
	resultContext := utils.MergeContext(baseContext, map[string]interface{}{
		"user_count": len(userResponses),
	})

	// Log performance
	logger.LogPerformance("find_by_role", time.Since(start), resultContext)

	handlerLog.Info("Users found by role", resultContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), resultContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users": userResponses,
		"count": len(userResponses),
	}, "find_by_role")
}

// FindByBranch handles finding users by branch with comprehensive logging
func (h *AccountHandler) FindByBranch(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("find_by_branch")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Find by branch request started", baseContext)

	// Parse branch ID parameter
	branchID, apiErr := errorcustom.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid branch ID parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_branch_id",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Add branch ID to context
	baseContext["branch_id"] = branchID
	handlerLog.Debug("Branch ID parsed successfully", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindByBranch",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		logger.ErrorWithCause(
			"Service call failed",
			"service_error",
			logger.LayerExternal,
			"find_by_branch",
			utils.MergeContext(baseContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindByBranch", false, err, serviceContext)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeServiceError,
			"Failed to retrieve users by branch",
			http.StatusInternalServerError,
			"handler",
			"find_by_branch",
			err,
		).WithDetail("branch_id", branchID)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), baseContext)
		
		errorcustom.HandleError(w, clientError, "find_by_branch")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindByBranch", true, nil, serviceContext)

	var userResponses []dto.UserProfile
	for _, user := range res.Accounts {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	// Result context
	resultContext := utils.MergeContext(baseContext, map[string]interface{}{
		"user_count": len(userResponses),
	})

	// Log performance
	logger.LogPerformance("find_by_branch", time.Since(start), resultContext)

	handlerLog.Info("Users found by branch", resultContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), resultContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users": userResponses,
		"count": len(userResponses),
	}, "find_by_branch")
}

// SearchUsers handles advanced user search with comprehensive logging
func (h *AccountHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("search_users")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("User search request started", baseContext)

	// Parse query parameters
	query := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")
	branchIDStr := r.URL.Query().Get("branch_id")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	// Parse branch ID if provided
	var branchID int64
	if branchIDStr != "" {
		var err error
		branchID, err = strconv.ParseInt(branchIDStr, 10, 64)
		if err != nil {
			context := utils.MergeContext(baseContext, map[string]interface{}{
				"branch_id_str": branchIDStr,
				"error":         err.Error(),
			})
			
			logger.ErrorWithCause(
				"Invalid branch_id parameter",
				"invalid_parameter",
				logger.LayerHandler,
				"parse_branch_id",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeInvalidInput,
				"Invalid branch_id parameter",
				http.StatusBadRequest,
				"handler",
				"search_users",
				err,
			).WithDetail("branch_id", branchIDStr)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "search_users")
			return
		}
	}

	// Parse pagination parameters
	page, pageSize, apiErr := h.getPaginationParams(r)
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid pagination parameters",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_pagination",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Set defaults
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	statusFilter := []string{}
	if status != "" {
		statusFilter = strings.Split(status, ",")
	}

	// Create search context
	searchContext := utils.MergeContext(baseContext, map[string]interface{}{
		"query":         query,
		"role":          role,
		"branch_id":     branchID,
		"status_filter": statusFilter,
		"sort_by":       sortBy,
		"sort_order":    sortOrder,
		"page":          page,
		"page_size":     pageSize,
	})

	handlerLog.Debug("Search parameters parsed", searchContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.SearchUsers(ctx, &pb.SearchUsersReq{
		Query:    query,
		Role:     role,
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
		Sort: &pb.SortInfo{
			SortBy:    sortBy,
			SortOrder: sortOrder,
		},
		StatusFilter: statusFilter,
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(searchContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "SearchUsers",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		logger.ErrorWithCause(
			"User search service call failed",
			"service_error",
			logger.LayerExternal,
			"search_users",
			utils.MergeContext(searchContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log service call failure
		logger.LogServiceCall("user-service", "SearchUsers", false, err, serviceContext)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeServiceError,
			"Search failed",
			http.StatusInternalServerError,
			"handler",
			"search_users",
			err,
		).WithDetail("query", query).
		  WithDetail("role", role).
		  WithDetail("branch_id", branchID)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), searchContext)
		
		errorcustom.HandleError(w, clientError, "search_users")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "SearchUsers", true, nil, serviceContext)

	// Convert response according to your protobuf definition
	var userResponses []dto.UserProfile
	for _, user := range res.Accounts {
		userResponses = append(userResponses, dto.UserProfile{
			ID:        user.Id,
			BranchID:  user.BranchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Title:     user.Title,
			Role:      user.Role,
			OwnerID:   user.OwnerId,
			CreatedAt: user.CreatedAt.AsTime(),
			UpdatedAt: user.UpdatedAt.AsTime(),
		})
	}

	// Calculate total pages
	totalPages := int32(0)
	if res.Pagination != nil && res.Pagination.PageSize > 0 {
		totalPages = (res.Total + res.Pagination.PageSize - 1) / res.Pagination.PageSize
	}

	// Final result context
	resultContext := utils.MergeContext(searchContext, map[string]interface{}{
		"total_found":    res.Total,
		"returned_count": len(userResponses),
		"total_pages":    totalPages,
		"has_next":       res.Pagination != nil && res.Pagination.HasNext,
		"has_prev":       res.Pagination != nil && res.Pagination.HasPrev,
	})

	// Log performance
	logger.LogPerformance("search_users", time.Since(start), resultContext)

	handlerLog.Info("User search completed", resultContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), resultContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":       userResponses,
		"total_count": res.Total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"pagination": map[string]interface{}{
			"has_next": res.Pagination != nil && res.Pagination.HasNext,
			"has_prev": res.Pagination != nil && res.Pagination.HasPrev,
		},
	}, "search_users")
}

// GetUsersByBranch handles getting users by branch with pagination and comprehensive logging
func (h *AccountHandler) GetUsersByBranch(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("get_users_by_branch")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Get users by branch request started", baseContext)

	// Get branch ID from URL parameter using utility function
	branchID, apiErr := errorcustom.ParseIDParam(r, "branch_id")
	if apiErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": apiErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid branch ID parameter",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_branch_id",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	// Add branch ID to context
	baseContext["branch_id"] = branchID

	// Parse pagination parameters
	page, pageSize, paginationErr := h.getPaginationParams(r)
	if paginationErr != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": paginationErr.Error(),
		})
		
		logger.ErrorWithCause(
			"Invalid pagination parameters",
			"invalid_parameter",
			logger.LayerHandler,
			"parse_pagination",
			context,
		)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.RespondWithAPIError(w, paginationErr)
		return
	}

	// Create request context with pagination
	requestContext := utils.MergeContext(baseContext, map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	})

	handlerLog.Debug("Parameters parsed successfully", requestContext)

	// Call the gRPC service to get users by branch
	serviceStart := time.Now()
	res, err := h.userClient.FindByBranch(ctx, &pb.FindByBranchReq{
		BranchId: branchID,
		Pagination: &pb.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
		},
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(requestContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "FindByBranch",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		logger.ErrorWithCause(
			"Service call failed",
			"service_error",
			logger.LayerExternal,
			"get_users_by_branch",
			utils.MergeContext(requestContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log service call failure
		logger.LogServiceCall("user-service", "FindByBranch", false, err, serviceContext)
		
		clientError := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeServiceError,
			"Failed to get users by branch",
			http.StatusInternalServerError,
			"handler",
			"get_users_by_branch",
			err,
		).WithDetail("branch_id", branchID).
		  WithDetail("page", page).
		  WithDetail("page_size", pageSize)
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), requestContext)
		
		errorcustom.HandleError(w, clientError, "get_users_by_branch")
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "FindByBranch", true, nil, serviceContext)

	// Result context with response data
	resultContext := utils.MergeContext(requestContext, map[string]interface{}{
		"total_found":    res.Total,
		"returned_count": len(res.Accounts),
		"has_next":       res.Pagination != nil && res.Pagination.HasNext,
		"has_prev":       res.Pagination != nil && res.Pagination.HasPrev,
	})

	// Log performance
	logger.LogPerformance("get_users_by_branch", time.Since(start), resultContext)

	handlerLog.Info("Users by branch retrieved successfully", resultContext)

	// Log user activity - viewing branch users
	if userID, err := h.getUserIDFromContext(ctx); err == nil {
		logger.LogUserActivity(
			fmt.Sprint(userID),
			"", // email not available in this context
			"view",
			"branch_users",
			resultContext,
		)
	}

	// Prepare and send response
	response := map[string]interface{}{
		"accounts":   res.Accounts,
		"total":      res.Total,
		"pagination": res.Pagination,
	}

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), resultContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, response, "get_users_by_branch")
}