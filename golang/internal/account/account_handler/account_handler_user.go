package account_handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	dto "english-ai-full/internal/account/account_dto"
	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// CreateAccount handles user creation with comprehensive logging
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("create_account")
	
	// Extract client information
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Account creation request initiated", baseContext)

	// Decode request body
	var req dto.CreateUserRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "create_account", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode create account request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with decode error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "create_account")
		return
	}

	// Add user info to context
	userContext := utils.MergeContext(baseContext, map[string]interface{}{
		"email":     req.Email,
		"name":      req.Name,
		"role":      req.Role,
		"branch_id": req.BranchID,
		"owner_id":  req.OwnerID,
	})
	
	handlerLog.Info("Account creation request for user", userContext)

	// Validate request structure
	if err := h.validator.Struct(req); err != nil {
		context := utils.MergeContext(userContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error individually
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Create account request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "create_account")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during account creation",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
				"handler",
				"create_account",
				err,
			).WithDetail("email", req.Email)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "create_account")
		}
		return
	}

	handlerLog.Debug("Request validation passed", userContext)

	// Validate password with detailed logging
	if err := errorcustom.ValidatePasswordWithDetails(req.Password, "create_account"); err != nil {
		logger.LogValidationError("password", "Password validation failed", "***masked***")
		
		logger.WarningWithCause(
			"Password validation failed during account creation",
			"weak_password",
			logger.LayerHandler,
			"validate_password",
			userContext,
		)
		
		passwordErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeWeakPassword,
			"Password does not meet security requirements",
			http.StatusBadRequest,
			"handler",
			"create_account",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_validation")
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), userContext)
		
		errorcustom.HandleError(w, passwordErr, "create_account")
		return
	}

	handlerLog.Debug("Password validation passed", userContext)

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		context := utils.MergeContext(userContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Password hashing failed",
			"password_hashing_error",
			logger.LayerHandler,
			"hash_password",
			context,
		)
		
		hashErr := errorcustom.NewAPIErrorWithContext(
			errorcustom.ErrCodeInternalError,
			"Password processing failed during hashing",
			http.StatusInternalServerError,
			"handler",
			"create_account",
			err,
		).WithDetail("email", req.Email).
		  WithDetail("step", "password_hashing")
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), context)
		
		errorcustom.HandleError(w, hashErr, "create_account")
		return
	}

	// Call user creation service
	serviceStart := time.Now()
	handlerLog.Debug("Calling user creation service", utils.MergeContext(userContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "CreateUser",
	}))

	userRes, err := h.userClient.CreateUser(r.Context(), &pb.AccountReq{
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	serviceDuration := time.Since(serviceStart)

	// Log service call
	serviceContext := utils.MergeContext(userContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "CreateUser",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		var failureReason string
		
		if strings.Contains(err.Error(), "already exists") {
			failureReason = "duplicate_email"
			httpStatus = http.StatusConflict
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeDuplicateEmail,
				"An account with this email address already exists",
				httpStatus,
				"handler",
				"create_account",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "email_uniqueness_check").
			  WithDetail("suggestion", "Use a different email address or try logging in")
			
			logger.WarningWithCause(
				"Account creation failed - duplicate email",
				failureReason,
				logger.LayerHandler,
				"create_account",
				userContext,
			)
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"User creation service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"create_account",
				err,
			).WithDetail("email", req.Email).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"Account creation service call failed",
				failureReason,
				logger.LayerExternal,
				"create_account",
				utils.MergeContext(userContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "CreateUser", false, err, utils.MergeContext(serviceContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(userContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		errorcustom.HandleError(w, clientError, "create_account")
		return
	}

	// Account creation successful
	logger.LogServiceCall("user-service", "CreateUser", true, nil, serviceContext)

	// Add created user info to context
	createdUserContext := utils.MergeContext(userContext, map[string]interface{}{
		"created_user_id": userRes.Id,
	})

	// Log user activity
	logger.LogUserActivity(
		fmt.Sprint(userRes.Id),
		userRes.Email,
		"create",
		"user_account",
		createdUserContext,
	)

	// Log performance
	logger.LogPerformance("create_account", time.Since(start), createdUserContext)

	handlerLog.Info("Account created successfully", createdUserContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, time.Since(start), createdUserContext)

	errorcustom.RespondWithJSON(w, http.StatusCreated, dto.CreateUserResponse{
		BranchID: userRes.BranchId,
		Name:     userRes.Name,
		Email:    userRes.Email,
		Avatar:   userRes.Avatar,
		Title:    userRes.Title,
		Role:     userRes.Role,
		OwnerID:  userRes.OwnerId,
	}, "create_account")
}

// UpdateUserByID handles user updates with comprehensive logging
func (h *AccountHandler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create handler logger for this operation
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("update_user_by_id")
	
	// Create base context
	baseContext := utils.CreateBaseContext(r, nil)
	
	handlerLog.Info("Update user request initiated", baseContext)

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

	// Decode request body
	var req dto.UpdateUserRequest
	if err := errorcustom.DecodeJSON(r.Body, &req, "update_user", false); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode update user request",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with decode error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		errorcustom.HandleError(w, err, "update_user")
		return
	}

	// Add update info to context
	updateContext := utils.MergeContext(baseContext, map[string]interface{}{
		"email":     req.Email,
		"name":      req.Name,
		"role":      req.Role,
		"branch_id": req.BranchID,
		"owner_id":  req.OwnerID,
	})
	
	handlerLog.Info("Update user request for user", updateContext)

	// Validate request structure
	if err := h.validator.Struct(req); err != nil {
		context := utils.MergeContext(updateContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Log each validation error individually
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
			}
			
			logger.WarningWithCause(
				"Update user request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleValidationErrors(w, validationErrors, "update_user")
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error during user update",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			clientError := errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
				"handler",
				"update_user",
				err,
			).WithDetail("user_id", id)
			
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
			
			errorcustom.HandleError(w, clientError, "update_user")
		}
		return
	}

	handlerLog.Debug("Request validation passed", updateContext)

	// Call user update service
	serviceStart := time.Now()
	handlerLog.Debug("Calling user update service", utils.MergeContext(updateContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "UpdateUser",
	}))

	res, err := h.userClient.UpdateUser(r.Context(), &pb.UpdateUserReq{
		Id:       id,
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	serviceDuration := time.Since(serviceStart)

	// Log service call
	serviceContext := utils.MergeContext(updateContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "UpdateUser",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		var failureReason string
		
		if strings.Contains(err.Error(), "not found") {
			failureReason = "user_not_found"
			httpStatus = http.StatusNotFound
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User with the specified ID was not found",
				httpStatus,
				"handler",
				"update_user",
				err,
			).WithDetail("user_id", id)
			
			logger.WarningWithCause(
				"User update failed - user not found",
				failureReason,
				logger.LayerHandler,
				"update_user",
				updateContext,
			)
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
			
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"User update service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"update_user",
				err,
			).WithDetail("user_id", id).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"User update service call failed",
				failureReason,
				logger.LayerExternal,
				"update_user",
				utils.MergeContext(updateContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "UpdateUser", false, err, utils.MergeContext(serviceContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		// Log API request completion
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(updateContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		errorcustom.HandleError(w, clientError, "update_user")
		return
	}

	// User update successful
	logger.LogServiceCall("user-service", "UpdateUser", true, nil, serviceContext)

	// Add updated user info to context
	updatedUserContext := utils.MergeContext(updateContext, map[string]interface{}{
		"updated_email": res.Account.Email,
		"updated_role":  res.Account.Role,
	})

	// Log user activity
	logger.LogUserActivity(
		fmt.Sprint(id),
		res.Account.Email,
		"update",
		"user_profile",
		updatedUserContext,
	)

	// Log performance
	logger.LogPerformance("update_user_by_id", time.Since(start), updatedUserContext)

	handlerLog.Info("User updated successfully", updatedUserContext)

	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), updatedUserContext)

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.UpdateUserResponse{
		User: dto.UserProfile{
			ID:        res.Account.Id,
			BranchID:  res.Account.BranchId,
			Name:      res.Account.Name,
			Email:     res.Account.Email,
			Avatar:    res.Account.Avatar,
			Title:     res.Account.Title,
			Role:      res.Account.Role,
			OwnerID:   res.Account.OwnerId,
			CreatedAt: res.Account.CreatedAt.AsTime(),
			UpdatedAt: res.Account.UpdatedAt.AsTime(),
		},
		Success: true,
		Message: "User updated successfully",
	}, "update_user")
}
func (h *AccountHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("delete_user")
	
	clientIP := errorcustom.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	baseContext := utils.CreateBaseContext(r, map[string]interface{}{
		"ip":         clientIP,
		"user_agent": userAgent,
	})
	
	handlerLog.Info("Delete user request initiated", baseContext)

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
		
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		errorcustom.RespondWithAPIError(w, apiErr)
		return
	}

	baseContext["user_id"] = id
	handlerLog.Info("Delete user request for user ID", baseContext)

	logger.LogSecurityEvent(
		"user_deletion_attempt",
		"Attempt to delete user account",
		"high",
		baseContext,
	)

	serviceStart := time.Now()
	handlerLog.Debug("Calling user deletion service", utils.MergeContext(baseContext, map[string]interface{}{
		"service": "AccountService",
		"method":  "DeleteUser",
	}))

	res, err := h.userClient.DeleteUser(r.Context(), &pb.DeleteAccountReq{UserID: id})
	serviceDuration := time.Since(serviceStart)

	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "DeleteUser",
		"duration_ms": serviceDuration.Milliseconds(),
	})

	if err != nil {
		var httpStatus int
		var clientError *errorcustom.APIError
		var failureReason string
		
		if strings.Contains(err.Error(), "not found") {
			failureReason = "user_not_found"
			httpStatus = http.StatusNotFound
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeUserNotFound,
				"User with the specified ID was not found",
				httpStatus,
				"handler",
				"delete_user",
				err,
			).WithDetail("user_id", id)
			
			logger.WarningWithCause(
				"User deletion failed - user not found",
				failureReason,
				logger.LayerHandler,
				"delete_user",
				baseContext,
			)
		} else {
			failureReason = "service_error"
			httpStatus = http.StatusInternalServerError
			clientError = errorcustom.NewAPIErrorWithContext(
				errorcustom.ErrCodeServiceError,
				"User deletion service is temporarily unavailable. Please try again later.",
				httpStatus,
				"handler",
				"delete_user",
				err,
			).WithDetail("user_id", id).
			  WithDetail("step", "service_call").
			  WithDetail("retryable", true)
			
			logger.ErrorWithCause(
				"User deletion service call failed",
				failureReason,
				logger.LayerExternal,
				"delete_user",
				utils.MergeContext(baseContext, map[string]interface{}{
					"error": err.Error(),
				}),
			)
		}
		
		logger.LogServiceCall("user-service", "DeleteUser", false, err, utils.MergeContext(serviceContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		logger.LogAPIRequest(r.Method, r.URL.Path, httpStatus, time.Since(start), utils.MergeContext(baseContext, map[string]interface{}{
			"failure_reason": failureReason,
		}))
		
		errorcustom.HandleError(w, clientError, "delete_user")
		return
	}

	logger.LogServiceCall("user-service", "DeleteUser", true, nil, serviceContext)

	deletionContext := utils.MergeContext(baseContext, map[string]interface{}{
		"deletion_success": res.Success,
	})

	if res.Success {
		logger.LogUserActivity(
			fmt.Sprint(id),
			"",
			"delete",
			"user_account",
			deletionContext,
		)

		logger.LogSecurityEvent(
			"user_account_deleted",
			"User account successfully deleted",
			"high",
			deletionContext,
		)

		handlerLog.Info("User deleted successfully", deletionContext)
	} else {
		handlerLog.Warning("User deletion reported as unsuccessful", deletionContext)
	}

	logger.LogPerformance("delete_user", time.Since(start), deletionContext)
	logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), deletionContext)

	// Fixed response handling
	responseMessage := "User deleted successfully"
	if !res.Success {
		responseMessage = "User deletion was unsuccessful"
	}

	errorcustom.RespondWithJSON(w, http.StatusOK, dto.DeleteUserResponse{
		Success: res.Success,
		Message: responseMessage,
	}, "delete_user")
}