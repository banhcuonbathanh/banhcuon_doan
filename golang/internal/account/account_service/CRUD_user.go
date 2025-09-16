package account_service

import (
	"context"
	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/proto_qr/account"
	"english-ai-full/logger/core"
	"english-ai-full/token"
	"english-ai-full/utils"
	"fmt"
	"time"
)




func (s *AccountService) CreateUser(ctx context.Context, req *account.AccountReq) (*account.Account, error) {
    const operation = "create_user"
    startTime := time.Now()
    
    // Extract request ID from context or protobuf request
    requestID := s.getRequestIDFromContext(ctx, req)
    
    // Build operation context with request ID
    operationCtx := s.layerContext.BuildOperationContext(operation, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(req.Email),
        "role":       req.Role,
        "branch_id":  req.BranchId,
        "owner_id":   req.OwnerId,
    })

    // Set request ID in logger
 
    s.logger.SetOperation(operation)

    // Log operation start with request ID
    s.logger.Info(core.MsgOperationStarted, s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":     requestID,
        "operation":      operation,
        "target_user":    utils.MaskEmail(req.Email),
        "requested_role": req.Role,
        "branch_id":      req.BranchId,
    }))

    // Context cancellation check
    if err := ctx.Err(); err != nil {
        s.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled,
            core.LayerService, operation, s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id": requestID,
            }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return nil, appErr
    }

    // Input validation with request ID logging
    if err := s.validator.Struct(req); err != nil {
        s.logger.Error("Request validation failed", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "error":      err.Error(),
            "email":      utils.MaskEmail(req.Email),
        }))
        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return nil, appErr
    }

    // Auto-fill role logic...
    defaultRole := req.Role
    if defaultRole == "" {
        defaultRole = "guest"
        s.logger.Info("Auto-filling empty role with guest", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":     requestID,
            "operation":      operation,
            "email":          utils.MaskEmail(req.Email),
            "original_role":  req.Role,
            "default_role":   defaultRole,
        }))
    }

    // Convert protobuf request to DTO
    userDTO := account_dto.Account{
        BranchID: req.BranchId,
        Name:     req.Name,
        Email:    req.Email,
        Password: req.Password,
        Avatar:   req.Avatar,
        Title:    req.Title,
        Role:     account_dto.Role(defaultRole),
        OwnerID:  req.OwnerId,
        Status:   "active",
    }

    // Password hashing with request ID logging
    if s.passwordHash != nil {
        hashStartTime := time.Now()

        s.logger.Debug("Starting password hashing", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "operation":  operation,
            "email":      utils.MaskEmail(req.Email),
        }))

        hashedPassword, err := s.passwordHash.HashPassword(userDTO.Password)
        if err != nil {
            operationCtx["hash_error"] = "failed to hash password"

            s.logger.ErrorWithCause("Password hashing failed", "password_hash_error",
                core.LayerService, operation, s.layerContext.MergeWithContext(map[string]interface{}{
                    "request_id":        requestID,
                    "email":             utils.MaskEmail(req.Email),
                    "hash_duration_ms":  time.Since(hashStartTime).Milliseconds(),
                    "error":             err.Error(),
                }))

            appErr := s.errorHandler.Handle(err, operation, operationCtx)
            return nil, appErr
        }

        userDTO.Password = hashedPassword
    }

    // Add request ID to repository context
    ctx = context.WithValue(ctx, "request_id", requestID)

    // Log repository call with request ID
    repoStartTime := time.Now()
    s.logger.Info("Calling repository to create user", s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":      requestID,
        "operation":       operation,
        "target_layer":    core.LayerRepository,
        "target_function": core.FuncCreateUser,
        "email":           utils.MaskEmail(req.Email),
        "role":            defaultRole,
    }))

    // Call repository with request ID in context
    createdUser, err := s.userRepo.CreateUser(ctx, userDTO)
    if err != nil {
        operationCtx["repository_error"] = "failed to create user in repository"
        operationCtx["repository_duration_ms"] = time.Since(repoStartTime).Milliseconds()

        if appErr, ok := error_system.IsAppError(err); ok {
            s.logger.Error("Repository returned AppError", s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":             requestID,
                "error_code":             appErr.Code,
                "error_message":          appErr.Message,
                "email":                  utils.MaskEmail(req.Email),
                "repository_duration_ms": time.Since(repoStartTime).Milliseconds(),
            }))
            return nil, appErr
        }

        s.logger.ErrorWithDomainAndCause("Repository call failed", s.layerContext.Domain,
            core.CauseDatabaseError, core.LayerService, operation,
            s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":             requestID,
                "target_layer":           core.LayerRepository,
                "target_function":        core.FuncCreateUser,
                "email":                  utils.MaskEmail(req.Email),
                "repository_duration_ms": time.Since(repoStartTime).Milliseconds(),
                "error":                  err.Error(),
            }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return nil, appErr
    }

    // Log success with request ID
    totalDuration := time.Since(startTime)
    s.logger.Info(core.MsgOperationCompleted, s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":  requestID,
        "operation":   operation,
        "user_id":     createdUser.ID,
        "email":       utils.MaskEmail(createdUser.Email),
        "role":        createdUser.Role,
        "branch_id":   createdUser.BranchID,
        "success":     true,
        "duration_ms": totalDuration.Milliseconds(),
    }))

    // Email sending with request ID
    if s.emailService != nil {
        go func() {
            emailCtx := context.WithValue(context.Background(), "request_id", requestID)
            emailStartTime := time.Now()

            if err := s.emailService.SendWelcomeEmail(emailCtx, createdUser.Email, createdUser.Name); err != nil {
                s.logger.ErrorWithCause("Welcome email send failed", core.CauseEmailSendFailed,
                    core.LayerService, "send_welcome_email", s.layerContext.MergeWithContext(map[string]interface{}{
                        "request_id":         requestID,
                        "user_id":            createdUser.ID,
                        "email":              utils.MaskEmail(createdUser.Email),
                        "email_duration_ms":  time.Since(emailStartTime).Milliseconds(),
                        "error":              err.Error(),
                    }))
            } else {
                s.logger.Info("Welcome email sent successfully", s.layerContext.MergeWithContext(map[string]interface{}{
                    "request_id":         requestID,
                    "operation":          "send_welcome_email",
                    "user_id":            createdUser.ID,
                    "email":              utils.MaskEmail(createdUser.Email),
                    "email_duration_ms":  time.Since(emailStartTime).Milliseconds(),
                    "success":            true,
                }))
            }
        }()
    }

    return s.convertDTOToProto(&createdUser), nil
}
// ne

func (s *AccountService) Login(ctx context.Context, loginReq *account.LoginReq) (*account.AccountRes, error) {
    const operation = "login"
    startTime := time.Now()
    
    // Generate request ID for this operation
    requestID := fmt.Sprintf("service_req_%d", time.Now().UnixNano())
    
    // Build operation context with request ID
    operationCtx := s.layerContext.BuildOperationContext(operation, map[string]interface{}{
        "request_id": requestID,
        "email":      utils.MaskEmail(loginReq.Email),
    })

    // Set request ID in logger
    s.logger.SetOperation(operation)

    // Log operation start with request ID
    s.logger.Info(core.MsgOperationStarted, s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id": requestID,
        "operation":  operation,
        "email":      utils.MaskEmail(loginReq.Email),
    }))

    // Context cancellation check
    if err := ctx.Err(); err != nil {
        s.logger.ErrorWithCause(core.MsgContextError, core.CauseContextCancelled,
            core.LayerService, operation, s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id": requestID,
            }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    // Input validation with request ID logging
    if err := s.validator.Struct(loginReq); err != nil {
        s.logger.Error("Login request validation failed", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "error":      err.Error(),
            "email":      utils.MaskEmail(loginReq.Email),
        }))
        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    // Add request ID to repository context
    ctx = context.WithValue(ctx, "request_id", requestID)

    // Log repository call with request ID
    repoStartTime := time.Now()
    s.logger.Info("Calling repository to find user by email", s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":      requestID,
        "operation":       operation,
        "target_layer":    core.LayerRepository,
        "target_function": "FindByEmail",
        "email":           utils.MaskEmail(loginReq.Email),
    }))

    // Call repository to find user by email
    user, err := s.userRepo.FindByEmail(ctx, loginReq.Email)
    if err != nil {
        operationCtx["repository_error"] = "failed to find user by email"
        operationCtx["repository_duration_ms"] = time.Since(repoStartTime).Milliseconds()

        if appErr, ok := error_system.IsAppError(err); ok {
            s.logger.Error("Repository returned AppError", s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":             requestID,
                "error_code":             appErr.Code,
                "error_message":          appErr.Message,
                "email":                  utils.MaskEmail(loginReq.Email),
                "repository_duration_ms": time.Since(repoStartTime).Milliseconds(),
            }))
            return &account.AccountRes{}, appErr
        }

        s.logger.ErrorWithDomainAndCause("Repository call failed", s.layerContext.Domain,
            core.CauseDatabaseError, core.LayerService, operation,
            s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":             requestID,
                "target_layer":           core.LayerRepository,
                "target_function":        "FindByEmail",
                "email":                  utils.MaskEmail(loginReq.Email),
                "repository_duration_ms": time.Since(repoStartTime).Milliseconds(),
                "error":                  err.Error(),
            }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    // Check if user account is active
    if user.Status != "active" {
        operationCtx["account_status"] = user.Status
        
        s.logger.Error("Login attempt for inactive account", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":     requestID,
            "operation":      operation,
            "email":          utils.MaskEmail(loginReq.Email),
            "account_status": user.Status,
            "user_id":        user.ID,
        }))

        // Create a custom error for inactive account
        inactiveErr := fmt.Errorf("account is %s", user.Status)
        appErr := s.errorHandler.Handle(inactiveErr, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    // Password verification with request ID logging
    if s.passwordHash != nil {
        hashStartTime := time.Now()

        s.logger.Debug("Starting password verification", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id": requestID,
            "operation":  operation,
            "email":      utils.MaskEmail(loginReq.Email),
            "user_id":    user.ID,
        }))

        // Use ComparePassword to verify the password
        isValid := s.passwordHash.ComparePassword(user.Password, loginReq.Password)

        if !isValid {
            operationCtx["invalid_credentials"] = "password does not match"

            s.logger.Error("Invalid credentials provided", s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":               requestID,
                "operation":                operation,
                "email":                    utils.MaskEmail(loginReq.Email),
                "user_id":                  user.ID,
                "verification_duration_ms": time.Since(hashStartTime).Milliseconds(),
            }))

            // Create a custom error for invalid credentials
            invalidCredErr := fmt.Errorf("invalid email or password")
            appErr := s.errorHandler.Handle(invalidCredErr, operation, operationCtx)
            return &account.AccountRes{}, appErr
        }

        s.logger.Debug("Password verification successful", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":               requestID,
            "operation":                operation,
            "email":                    utils.MaskEmail(loginReq.Email),
            "user_id":                  user.ID,
            "verification_duration_ms": time.Since(hashStartTime).Milliseconds(),
        }))
    }

    // Generate JWT tokens after successful authentication
    tokenStartTime := time.Now()
    
    // Convert user DTO to account DTO for token generation
    userAccount := account_dto.Account{
        ID:       user.ID,
        Email:    user.Email,
        Role:     user.Role, // Make sure this matches the Role type in account_dto
        BranchID: user.BranchID,
    }

    // Generate access token
    accessToken, err := token.GenerateJWTToken(userAccount)
    if err != nil {
        operationCtx["token_generation_error"] = "failed to generate access token"
        
        s.logger.Error("Failed to generate access token", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":              requestID,
            "operation":               operation,
            "user_id":                 user.ID,
            "email":                   utils.MaskEmail(user.Email),
            "token_generation_duration_ms": time.Since(tokenStartTime).Milliseconds(),
            "error":                   err.Error(),
        }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    // Generate refresh token
    refreshToken, err := token.GenerateRefreshToken(userAccount)
    if err != nil {
        operationCtx["token_generation_error"] = "failed to generate refresh token"
        
        s.logger.Error("Failed to generate refresh token", s.layerContext.MergeWithContext(map[string]interface{}{
            "request_id":              requestID,
            "operation":               operation,
            "user_id":                 user.ID,
            "email":                   utils.MaskEmail(user.Email),
            "token_generation_duration_ms": time.Since(tokenStartTime).Milliseconds(),
            "error":                   err.Error(),
        }))

        appErr := s.errorHandler.Handle(err, operation, operationCtx)
        return &account.AccountRes{}, appErr
    }

    s.logger.Debug("JWT tokens generated successfully", s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":              requestID,
        "operation":               operation,
        "user_id":                 user.ID,
        "email":                   utils.MaskEmail(user.Email),
        "token_generation_duration_ms": time.Since(tokenStartTime).Milliseconds(),
    }))

    // Log successful login with request ID
    totalDuration := time.Since(startTime)
    s.logger.Info(core.MsgOperationCompleted, s.layerContext.MergeWithContext(map[string]interface{}{
        "request_id":  requestID,
        "operation":   operation,
        "user_id":     user.ID,
        "email":       utils.MaskEmail(user.Email),
        "role":        user.Role,
        "branch_id":   user.BranchID,
        "success":     true,
        "duration_ms": totalDuration.Milliseconds(),
    }))

    // Optional: Log login activity (if you have an activity logging service)
    if s.emailService != nil {
        go func() {
            activityStartTime := time.Now()

            // You could add login activity logging here
            s.logger.Info("User login activity recorded", s.layerContext.MergeWithContext(map[string]interface{}{
                "request_id":            requestID,
                "operation":             "record_login_activity",
                "user_id":               user.ID,
                "email":                 utils.MaskEmail(user.Email),
                "activity_duration_ms":  time.Since(activityStartTime).Milliseconds(),
                "success":               true,
            }))
        }()
    }

    // Convert DTO to Proto and include tokens
    accountProto := s.convertDTOToProto(&user)
    
    return &account.AccountRes{
        Account:      accountProto,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}