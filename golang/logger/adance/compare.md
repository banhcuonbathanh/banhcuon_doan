// BEFORE: Your current approach (lots of manual context filling)
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("=== REGISTER FUNCTION STARTED ===", map[string]interface{}{
		"test": "logging_verification",
		"endpoint": r.URL.Path,
		"method": r.Method,
		"operation": "register",
		"layer": "handler",
		"domain": h.domain,
		"request_id": requestID,
		"client_ip": clientIP,
	})
	// ... repeated context filling in every log call
}

// AFTER: Using the enhanced logger (minimal context needed)
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// One line to get fully contextualized logger
	log := h.domainLogger.WithContext(r.Context()).WithOperation("register")
	
	// All context (domain, layer, operation, request_id, etc.) automatically included
	log.Info("Register function started", map[string]interface{}{
		"endpoint": r.URL.Path,
		"method":   r.Method,
	})
	
	// Every subsequent log automatically has full context
	log.ErrorWithCause("Validation failed", "invalid_email") // Automatically includes domain, operation, etc.
}

// ==================================================
// IMPLEMENTATION GUIDE FOR YOUR 10+ DOMAINS
// ==================================================

// Step 1: Create domain loggers (one-time setup)
var (
	AccountLogger      = NewAccountLogger()      // domain=account, layer=handler, component=account
	ProductLogger      = NewProductLogger()      // domain=product, layer=handler, component=product
	OrderLogger        = NewOrderLogger()        // domain=order, layer=handler, component=order
	PaymentLogger      = NewPaymentLogger()      // domain=payment, layer=handler, component=payment
	InventoryLogger    = NewInventoryLogger()    // domain=inventory, layer=handler, component=inventory
	NotificationLogger = NewNotificationLogger() // domain=notification, layer=handler, component=notification
	UserLogger         = NewUserLogger()         // domain=user, layer=handler, component=user
	AuthLogger         = NewAuthLogger()         // domain=auth, layer=handler, component=auth
	ReportLogger       = NewReportLogger()       // domain=report, layer=handler, component=report
	AnalyticsLogger    = NewAnalyticsLogger()    // domain=analytics, layer=handler, component=analytics
)

// Step 2: Update your handlers to use domain loggers
type AccountHandler struct {
	logger       *DomainLogger  // Change from SpecializedLogger to DomainLogger
	// ... other dependencies remain the same
}

func NewAccountHandler(/* your existing dependencies */) *AccountHandler {
	return &AccountHandler{
		logger: AccountLogger, // Use the pre-configured domain logger
		// ... other dependencies
	}
}

// Step 3: Simplified handler methods
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Create contextual logger with operation (automatically includes domain, layer, component)
	log := h.logger.WithContext(r.Context()).WithOperation("register")
	
	log.Info("Register function started", map[string]interface{}{
		"endpoint": r.URL.Path,
		"method":   r.Method,
	})
	
	startTime := time.Now()
	requestID := h.getRequestID(r)
	
	// All logging methods now automatically include full context
	log.LogRequestStart(requestID, r.Method, r.URL.Path, "")
	
	if err := r.Context().Err(); err != nil {
		log.ErrorWithCause("Request context cancelled", "context_timeout")
		log.LogRequestEnd(requestID, http.StatusRequestTimeout, time.Since(startTime))
		return
	}
	
	var registerRequest account_dto.CreateUserRequest
	if err := h.handlerErrorMgr.DecodeJSONRequest(r, &registerRequest, h.domain, requestID); err != nil {
		log.ErrorWithCause("Failed to parse request body", "json_decode_error")
		log.LogRequestEnd(requestID, http.StatusBadRequest, time.Since(startTime))
		return
	}
	
	if err := h.validator.Struct(registerRequest); err != nil {
		log.LogStructValidationError("CreateUserRequest", registerRequest, "Request validation failed")
		log.LogRequestEnd(requestID, http.StatusBadRequest, time.Since(startTime))
		return
	}
	
	// Service call with automatic context
	log.Info("Calling CreateUser service", map[string]interface{}{
		"service": "UserService",
		"method":  "CreateUser",
		"email":   registerRequest.Email,
	})
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	pbRequest := &pb.AccountReq{
		BranchId: registerRequest.BranchID,
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
		Avatar:   registerRequest.Avatar,
		Title:    registerRequest.Title,
		Role:     registerRequest.Role,
		OwnerId:  registerRequest.OwnerID,
	}
	
	createdUser, err := h.userClient.CreateUser(ctx, pbRequest)
	if err != nil {
		log.ErrorWithCause("CreateUser service call failed", "service_error", map[string]interface{}{
			"error": err.Error(),
		})
		statusCode := h.getHTTPStatusFromError(err)
		log.LogRequestEnd(requestID, statusCode, time.Since(startTime))
		h.handlerErrorMgr.RespondWithError(w, err, h.domain, requestID)
		return
	}
	
	responseData := h.prepareUserResponse(createdUser)
	log.LogRequestEnd(requestID, http.StatusCreated, time.Since(startTime))
	h.handlerErrorMgr.RespondWithCreated(w, responseData, h.domain, requestID)
	
	log.Info("User registration completed successfully", map[string]interface{}{
		"user_id":     createdUser.Id,
		"email":       createdUser.Email,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})
}

func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	log := h.logger.WithContext(r.Context()).WithOperation("login")
	
	log.Info("Login attempt started")
	// All your login logic with simplified logging...
}

func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	log := h.logger.WithContext(r.Context()).WithOperation("update_profile")
	
	log.Info("Profile update started")
	// All your update logic with simplified logging...
}

// For all your other 14 functions, just change the operation name:
// log := h.logger.WithContext(r.Context()).WithOperation("delete_account")
// log := h.logger.WithContext(r.Context()).WithOperation("reset_password")
// log := h.logger.WithContext(r.Context()).WithOperation("verify_email")
// etc.

// Step 4: Apply same pattern to all other domains
type ProductHandler struct {
	logger *DomainLogger
	// ... other dependencies
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		logger: ProductLogger, // Pre-configured for product domain
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	log := h.logger.WithContext(r.Context()).WithOperation("create_product")
	
	log.Info("Product creation started")
	// ... your product creation logic
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	log := h.logger.WithContext(r.Context()).WithOperation("get_products")
	
	log.Info("Fetching products")
	// ... your product fetching logic
}

// ==================================================
// EVEN SIMPLER: MIDDLEWARE APPROACH
// ==================================================

// Step 5: Use middleware to automatically inject logger into context
func LoggerMiddleware(domain string) func(http.Handler) http.Handler {
	domainLogger := NewDomainLogger(domain)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := generateRequestID()
			
			// Create context with logger
			ctx := context.WithValue(r.Context(), "logger", domainLogger.WithContext(r.Context()))
			ctx = context.WithValue(ctx, RequestIDKey, requestID)
			
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Then in your handlers, just extract from context:
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Extract pre-configured logger from context
	log := r.Context().Value("logger").(*ContextualLogger).WithOperation("register")
	
	// Now all logging is contextual with just one line!
	log.Info("Register function started")
	// ... rest of your logic
}

// ==================================================
// CONFIGURATION-BASED APPROACH (YAML)
// ==================================================

// Step 6: Create logging.yaml configuration file
/*
domains:
  account:
    layer: "handler"
    component: "account"
    level: "info"
    operations:
      - "register"
      - "login"
      - "update_profile"
      - "delete_account"
      # ... all 15 operations
  
  product:
    layer: "handler" 
    component: "product"
    level: "info"
    operations:
      - "create_product"
      - "get_products"
      - "update_product"
      # ... all product operations
  
  # ... all 10+ domains
*/

// Load configuration and create loggers automatically
var DomainLoggers = make(map[string]*DomainLogger)

func init() {
	config := loadLoggingConfig("config/logging.yaml")
	
	for domain, domainConfig := range config.Domains {
		DomainLoggers[domain] = NewDomainLogger(domain)
	}
}

// Usage becomes even simpler:
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := DomainLoggers["account"].WithContext(r.Context()).WithOperation("register")
	
	log.Info("Register function started")
	// All context automatically included!
}

// ==================================================
// SUMMARY OF BENEFITS
// ==================================================

/*
BEFORE (your current approach):
- 15 functions × 10+ domains = 150+ places to manually fill context
- Every log call needs: domain, layer, operation, request_id, etc.
- High maintenance overhead
- Easy to forget or mistype context fields
- Inconsistent logging across domains

AFTER (with enhanced logger):
- One line per function: log := h.logger.WithContext(r.Context()).WithOperation("operation_name")
- All context automatically included in every subsequent log call
- Consistent logging across all domains
- Easy to add new fields globally
- Type-safe operation names
- Automatic operation detection possible
- Configuration-driven setup

MAINTENANCE REDUCTION:
- From ~1000 lines of logging context code
- To ~150 lines of simple operation setup
- 85%+ reduction in logging-related code
- Much easier to maintain and modify
*/