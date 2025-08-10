func (h *AccountHandler) UpdateAccountStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()
	
	// Create handler logger for this layer
	handlerLog := logger.NewHandlerLogger()
	handlerLog.SetOperation("update_account_status")
	
	// Base context for all logs
	baseContext := utils.CreateBaseContext(r, nil)
	
	// Get request ID for domain error handling
	requestID := errorcustom.GetRequestIDFromContext(ctx)
	domain := "account"
	
	handlerLog.Info("Account status update request started", baseContext)

	// Parse ID parameter
	id, apiErr := errorcustom.ParseIDParamWithDomain(r, "id", domain)

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
		
		// FIXED: Use HandleDomainError instead of RespondWithAPIError
		errorcustom.HandleDomainError(w, apiErr, domain, requestID)
		return
	}

	// Add user ID to context for subsequent logs
	baseContext["user_id"] = id

	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended pending"`
	}

	// FIXED: Decode request body using standard json.Decoder
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"error": err.Error(),
		})
		
		logger.ErrorWithCause(
			"Failed to decode request body",
			"json_decode_error",
			logger.LayerHandler,
			"decode_json",
			context,
		)
		
		// Log API request with error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		
		// Create a validation error and use domain error handling
		validationErr := errorcustom.NewValidationError(domain, "request_body", "Invalid JSON format", map[string]interface{}{
			"parse_error": err.Error(),
		})
		errorcustom.HandleDomainError(w, validationErr, domain, requestID)
		return
	}

	// Add requested status to context
	baseContext["requested_status"] = req.Status
	
	handlerLog.Debug("Request body decoded successfully", baseContext)

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		context := utils.MergeContext(baseContext, map[string]interface{}{
			"validation_error": err.Error(),
		})
		
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Create error collection for multiple validation errors
			errorCollection := errorcustom.NewErrorCollection(domain)
			
			// Log each validation error and add to collection
			for _, validationError := range validationErrors {
				logger.LogValidationError(
					validationError.Field(),
					validationError.Tag(),
					validationError.Value(),
				)
				
				// Add individual validation error to collection
				fieldErr := errorcustom.NewValidationError(domain, validationError.Field(), 
					fmt.Sprintf("Field validation failed: %s", validationError.Tag()), 
					map[string]interface{}{
						"field": validationError.Field(),
						"tag": validationError.Tag(),
						"value": validationError.Value(),
					})
				errorCollection.Add(fieldErr)
			}
			
			logger.WarningWithCause(
				"Request validation failed",
				"validation_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			// FIXED: Use HandleDomainError with error collection
			errorcustom.HandleDomainError(w, errorCollection.ToAPIError(), domain, requestID)
		} else {
			logger.ErrorWithCause(
				"Unexpected validation error",
				"validation_system_error",
				logger.LayerHandler,
				"validate_request",
				context,
			)
			
			// FIXED: Create validation error and use domain error handling
			validationErr := errorcustom.NewValidationError(domain, "system", "Validation failed", map[string]interface{}{
				"system_error": err.Error(),
			})
			errorcustom.HandleDomainError(w, validationErr, domain, requestID)
		}
		
		// Log API request with validation error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start), context)
		return
	}

	handlerLog.Debug("Request validation passed", baseContext)

	// Call user service
	serviceStart := time.Now()
	res, err := h.userClient.UpdateAccountStatus(ctx, &pb.UpdateAccountStatusReq{
		UserId: id,
		Status: req.Status,
	})
	serviceDuration := time.Since(serviceStart)
	
	// Log service call
	serviceContext := utils.MergeContext(baseContext, map[string]interface{}{
		"service":     "user-service",
		"method":      "UpdateAccountStatus",
		"duration_ms": serviceDuration.Milliseconds(),
	})
	
	if err != nil {
		// Determine error type and appropriate response
		if strings.Contains(err.Error(), "not found") {
			logger.LogServiceCall("user-service", "UpdateAccountStatus", false, err, serviceContext)
			
			logger.WarningWithCause(
				"User not found for status update",
				"user_not_found",
				logger.LayerHandler,
				"update_account_status",
				baseContext,
			)
			
			// Log API request with not found error
			logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusNotFound, time.Since(start), baseContext)
			
			// FIXED: Use domain-specific not found error
			notFoundErr := errorcustom.NewAccountNotFoundError(id) // or NewUserNotFoundByID if that's more appropriate
			errorcustom.HandleDomainError(w, notFoundErr, domain, requestID)
			return
		}
		
		// Log service call failure
		logger.LogServiceCall("user-service", "UpdateAccountStatus", false, err, serviceContext)
		
		logger.ErrorWithCause(
			"User service call failed",
			"service_error",
			logger.LayerExternal,
			"update_account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"error": err.Error(),
			}),
		)
		
		// Log API request with service error
		logger.LogAPIRequest(r.Method, r.URL.Path, http.StatusInternalServerError, time.Since(start), baseContext)
		
		// FIXED: Create external service error and use domain error handling
		serviceErr := errorcustom.NewExternalServiceError(domain, "user-service", "UpdateAccountStatus", err, true)
		errorcustom.HandleDomainError(w, serviceErr, domain, requestID)
		return
	}

	// Log successful service call
	logger.LogServiceCall("user-service", "UpdateAccountStatus", true, nil, serviceContext)
	
	// Determine response status
	status := http.StatusOK
	if !res.Success {
		status = http.StatusBadRequest
		
		logger.WarningWithCause(
			"Account status update failed at service level",
			"service_business_logic_error",
			logger.LayerService,
			"update_account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"service_message": res.Message,
				"service_success": res.Success,
			}),
		)
	} else {
		// Log successful user activity
		logger.LogUserActivity(
			fmt.Sprint(id),
			"", // email not available in this context
			"update",
			"account_status",
			utils.MergeContext(baseContext, map[string]interface{}{
				"old_status": "unknown", // Would need to be passed from service
				"new_status": req.Status,
			}),
		)
		
		handlerLog.Info("Account status updated successfully", utils.MergeContext(baseContext, map[string]interface{}{
			"service_message": res.Message,
		}))
	}

	// Log performance
	totalDuration := time.Since(start)
	logger.LogPerformance("update_account_status", totalDuration, baseContext)
	
	// Log API request completion
	logger.LogAPIRequest(r.Method, r.URL.Path, status, totalDuration, utils.MergeContext(baseContext, map[string]interface{}{
		"response_success": res.Success,
		"response_message": res.Message,
	}))

	// FIXED: Send response using domain success response
	responseData := map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	}
	
	if status == http.StatusOK {
		errorcustom.RespondWithDomainSuccess(w, responseData, domain, requestID)
	} else {
		// For business logic failures, create appropriate error
		businessErr := errorcustom.NewBusinessLogicError(domain, "status_update_failed", res.Message, map[string]interface{}{
			"requested_status": req.Status,
			"user_id": id,
		})
		errorcustom.HandleDomainError(w, businessErr, domain, requestID)
	}
}