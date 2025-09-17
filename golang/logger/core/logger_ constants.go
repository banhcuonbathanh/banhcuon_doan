// logger/core/constants.go - Enhanced with better error tracking fields
package core

// Operation constants
const (
	// User operations
	OperationRegister     = "register"
	OperationLogin        = "login"
	OperationLogout       = "logout"
	OperationRefreshToken = "refresh_token"
	OperationUpdateUser   = "update_user"
	OperationDeleteUser   = "delete_user"
	OperationGetUser      = "get_user"
	OperationListUsers    = "list_users"
	OperationCreateUser   = "create_user"
	OperationVerifyEmail  = "verify_email"
	OperationResetPassword = "reset_password"
	OperationChangePassword = "change_password"
	
	// Database operations
	OperationInsert = "insert"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationSelect = "select"
	OperationQuery  = "query"
	
	// Validation operations
	OperationValidateStruct = "validate_struct"
	OperationValidateField  = "validate_field"
	OperationValidateEmail  = "validate_email"
	OperationValidatePassword = "validate_password"
	
	// External service operations
	OperationSendEmail    = "send_email"
	OperationUploadFile   = "upload_file"
	OperationDownloadFile = "download_file"
	OperationCallAPI      = "call_api"
)

// Field name constants for logging
const (
	// Request/Response fields
	FieldRequestID    = "request_id"
	FieldMethod       = "method"
	FieldEndpoint     = "endpoint" 
	FieldPath         = "path"
	FieldClientIP     = "client_ip"
	FieldUserAgent    = "user_agent"
	FieldStatusCode   = "status_code"
	FieldDurationMS   = "duration_ms"
	FieldResponseSize = "response_size"
	
	// User/Account fields
	FieldUserID      = "user_id"
	FieldEmail       = "email"
	FieldRole        = "role"
	FieldBranchID    = "branch_id"
	FieldOwnerID     = "owner_id"
	FieldName        = "name"
	FieldTitle       = "title"
	FieldAvatar      = "avatar"
	FieldCreatedAt   = "created_at"
	FieldUpdatedAt   = "updated_at"
	
	// Service/System fields
	FieldService       = "service"
	FieldServiceMethod = "service_method"
	FieldTargetService = "target_service"
	FieldCallType      = "call_type"
	FieldComponent     = "component"
	FieldOperation     = "operation"
	FieldLayer         = "layer"
	FieldDomain        = "domain"
	FieldVersion       = "version"
	FieldEnvironment   = "environment"
		FieldConfigPath   = "Config_Path"
	
	// Database fields
	FieldTable        = "table"
	FieldFunction     = "function"
	FieldRowsAffected = "rows_affected"
	FieldSuccess      = "success"
	FieldQuery        = "query"
	FieldQueryTime    = "query_time"
	
	// Enhanced Error tracking fields
	FieldError            = "error"
	FieldErrorCode        = "error_code"
	FieldErrorType        = "error_type"
	FieldCause            = "cause"
	FieldStackTrace       = "stack_trace"
	FieldSourceMethod     = "source_method"      // Which method the error originated from
	FieldErrorLocation    = "error_location"     // Specific location within method where error occurred
	FieldErrorTimestamp   = "error_timestamp"    // When the error occurred
	FieldErrorDurationMS  = "error_duration_ms"  // How long the operation took before failing
	FieldCallerFunction   = "caller_function"    // The calling function
	FieldCallerFile       = "caller_file"        // The calling file
	FieldCallerLine       = "caller_line"        // The calling line
	FieldCallerPackage    = "caller_package"     // The calling package
	
	// Validation fields
	FieldStructName     = "struct_name"
	FieldFieldName      = "field_name"
	FieldValidationTag  = "validation_tag"
	FieldValidationMsg  = "validation_message"
	FieldDecodeError    = "decode_error"
	FieldValidationStep = "validation_step"     // Which validation step failed
	
	// Authentication fields
	FieldTokenID       = "token_id"
	FieldTokenType     = "token_type"
	FieldSessionID     = "session_id"
	FieldTraceID       = "trace_id"
	FieldPermissions   = "permissions"
	FieldAuthMethod    = "auth_method"
	
	// External service fields
	FieldExternalID     = "external_id"
	FieldExternalStatus = "external_status"
	FieldRetryCount     = "retry_count"
	FieldTimeout        = "timeout"
	
	// File/Upload fields
	FieldFileName = "file_name"
	FieldFileSize = "file_size"
	FieldFileType = "file_type"
	FieldFilePath = "file_path"
	
	// Database-specific fields
	FieldSQLOperation     = "sql_operation"      // INSERT, UPDATE, DELETE, SELECT
	FieldDatabaseTable    = "database_table"     // Actual table name
	FieldBoilOperation    = "boil_operation"     // SQLBoiler operation name
	FieldConstraintType   = "constraint_type"    // Type of constraint violation
	
	// Enhanced context fields
	FieldInsertStep       = "insert_step"        // Which step of insert process
	FieldSuccessStep      = "success_step"       // Which step succeeded
	FieldValidationTarget = "validation_target"  // What was being validated
	FieldConversionStep   = "conversion_step"    // Which conversion step
	FieldAnalysisStep     = "analysis_step"      // Which analysis step
	FieldPerformanceDurationMS = "performance_duration_ms" // Performance timing
	FieldDatabaseSuccess  = "database_success"   // Database operation success flag
	FieldContextErrorType = "context_error_type" // Type of context error
	
	// Test/Debug fields
	FieldTest            = "test"
	FieldMessage         = "message"
	FieldTimestamp       = "timestamp"
	FieldAttemptedEmail  = "attempted_email"
	FieldAttemptedRole   = "attempted_role"
)

// Response field constants for API responses
const (
	ResponseFieldID        = "id"
	ResponseFieldEmail     = "email"
	ResponseFieldName      = "name"
	ResponseFieldRole      = "role"
	ResponseFieldBranchID  = "branch_id"
	ResponseFieldOwnerID   = "owner_id"
	ResponseFieldTitle     = "title"
	ResponseFieldAvatar    = "avatar"
	ResponseFieldCreatedAt = "created_at"
	ResponseFieldUpdatedAt = "updated_at"
	ResponseFieldStatus    = "status"
	ResponseFieldMessage   = "message"
	ResponseFieldData      = "data"
	ResponseFieldError     = "error"
	ResponseFieldSuccess   = "success"
	ResponseFieldToken     = "token"
	ResponseFieldTokenType = "token_type"
	ResponseFieldExpiresAt = "expires_at"
)

// Enhanced Log message constants with better error context
const (
	// General messages
	MsgRequestStarted   = "Request started"
	MsgRequestCompleted = "Request completed"
	MsgOperationStarted = "Operation started"
	MsgOperationCompleted = "Operation completed"
	
	// Registration messages
	MsgRegisterStarted           = "=== REGISTER FUNCTION STARTED ==="
	MsgRegisterCompleted         = "User registration completed successfully"
	MsgRegisterValidationFailed  = "Registration validation failed"
	MsgRegisterUserExists        = "User already exists"
	MsgRegisterServiceCallFailed = "Service call failed"
	
	// Enhanced Database messages with more context
	MsgDatabaseOperationStarted  = "Starting database operation"
	MsgDatabaseOperationFailed   = "Database operation failed"
	MsgDatabaseOperationSuccess  = "Database operation completed successfully"
	MsgDatabaseInsertAttempt     = "Attempting to insert user"
	MsgDatabaseInsertSuccess     = "Successfully created user"
	MsgDatabaseInsertFailed      = "Failed to insert user"
	MsgDatabaseConstraintViolation = "Database constraint violation detected"
	MsgDatabaseConnectionError     = "Database connection error"
	MsgDatabaseTimeoutError        = "Database operation timeout"
	
	// Service messages
	MsgServiceCallStarted  = "Calling service"
	MsgServiceCallSuccess  = "Service call completed successfully"
	MsgServiceCallFailed   = "Service call failed"
	MsgServiceCallTimeout  = "Service call timeout"
	
	// Enhanced Validation messages
	MsgValidationStarted        = "Starting validation"
	MsgValidationFailed         = "Validation failed"
	MsgValidationError          = "Validation error"
	MsgStructValidationError    = "Struct validation error"
	MsgRequestValidationFailed  = "Request validation failed"
	MsgORMConversionFailed      = "ORM conversion failed"
	MsgDTOConversionFailed      = "DTO conversion failed"
	
	// Enhanced Context messages
	MsgContextCancelled = "Request context cancelled"
	MsgContextTimeout   = "Context timeout"
	MsgContextError     = "Context error"
	
	// Authentication messages
	MsgAuthStarted     = "Authentication started"
	MsgAuthSuccess     = "Authentication successful"
	MsgAuthFailed      = "Authentication failed"
	MsgTokenGenerated  = "Token generated"
	MsgTokenValidated  = "Token validated"
	MsgTokenExpired    = "Token expired"
	
	// External service messages
	MsgExternalCallStarted = "External service call started"
	MsgExternalCallSuccess = "External service call successful"
	MsgExternalCallFailed  = "External service call failed"
	
	// Email messages
	MsgEmailSendStarted = "Email send started"
	MsgEmailSendSuccess = "Email sent successfully"
	MsgEmailSendFailed  = "Email send failed"
	
	// Error wrapping messages
	MsgErrorWrappedWithContext = "Error wrapped with enhanced context"
	MsgErrorCauseAnalyzed      = "Error cause analyzed"
	MsgStackTraceCapture       = "Stack trace captured for error"
)

// Test/Debug constants
const (
	TestLoggingVerification = "logging_verification"
	TestDatabaseConnection  = "database_connection"
	TestServiceConnection   = "service_connection"
	TestValidation         = "validation_test"
)

// Enhanced Error cause constants with more specific causes
const (
	// Context-related causes
	CauseContextCancelled       = "context_cancelled"
	CauseContextTimeout         = "context_timeout"
	
	// Validation-related causes
	CauseValidationFailed       = "validation_failed"
	CauseStructValidationFailed = "struct_validation_failed"
	CauseDecodeRequestFailed    = "decode_request_failed"
	CauseORMConversionFailed    = "orm_conversion_failed"
	CauseDTOConversionFailed    = "dto_conversion_failed"
	
	// Database-related causes
	CauseDuplicateEmail         = "duplicate_email"
	CauseDuplicateEntry         = "duplicate_entry"
	CauseDatabaseError          = "database_error"
	CauseConstraintViolation    = "constraint_violation"
	CauseForeignKeyViolation    = "foreign_key_violation"
	CauseUniqueConstraintViolation = "unique_constraint_violation"
	CauseDatabaseConnectionFailed  = "database_connection_failed"
	CauseDatabaseTimeout          = "database_timeout"
	CauseTransactionFailed        = "transaction_failed"
	
	// Service-related causes
	CauseServiceError           = "service_error"
	CauseExternalServiceError   = "external_service_error"
	CauseServiceTimeout         = "service_timeout"
	CauseServiceUnavailable     = "service_unavailable"
	
	// Authentication-related causes
	CauseAuthenticationFailed   = "authentication_failed"
	CausePermissionDenied       = "permission_denied"
	CauseTokenExpired           = "token_expired"
	CauseInvalidToken           = "invalid_token"
	CauseInvalidCredentials     = "invalid_credentials"
	
	// Network-related causes
	CauseNetworkError           = "network_error"
	CauseTimeout                = "timeout"
	CauseConnectionRefused      = "connection_refused"
	CauseConnectionLost         = "connection_lost"
	
	// File/Resource-related causes
	CauseEmailSendFailed        = "email_send_failed"
	CauseFileUploadFailed       = "file_upload_failed"
	CauseFileNotFound           = "file_not_found"
	CauseResourceNotFound       = "resource_not_found"
	
	// Business logic causes
	CauseBusinessRuleViolation  = "business_rule_violation"
	CauseInvalidState           = "invalid_state"
	CauseDataInconsistency      = "data_inconsistency"
)

// Table name constants (for repository layer)
const (
	TableAccounts     = "accounts"
	TableUsers        = "users"
	TableSessions     = "sessions"
	TableTokens       = "tokens"
	TablePermissions  = "permissions"
	TableRoles        = "roles"
	TableBranches     = "branches"
	TableAuditLog     = "audit_log"
)

// Function name constants (for repository layer)
const (
	FuncCreateUser      = "CreateUser"
		FuncLogin      = "Login"
	FuncGetUser         = "GetUser"
	FuncUpdateUser      = "UpdateUser"
	FuncDeleteUser      = "DeleteUser"
	FuncGetUserByEmail  = "GetUserByEmail"
	FuncListUsers       = "ListUsers"
	FuncCreateSession   = "CreateSession"
	FuncGetSession      = "GetSession"
	FuncDeleteSession   = "DeleteSession"
	FuncCreateToken     = "CreateToken"
	FuncValidateToken   = "ValidateToken"
	FuncRevokeToken     = "RevokeToken"
	
	// Helper function names
	FuncBuildORMAccount         = "buildORMAccount"
	FuncBuildORMAccountWithContext = "buildORMAccountWithContext"
	FuncConvertORMToDTO         = "convertORMToDTO"
	FuncWrapErrorWithContext    = "wrapErrorWithContext"
	FuncHandleContextError      = "handleContextError"
	FuncHandleInsertSuccess     = "handleInsertSuccess"
	FuncDetermineCause          = "determineCause"
	FuncLogDatabaseOperation    = "logDatabaseOperation"
)

// Error location constants for better error tracking
const (
	LocationContextCheck      = "context_check"
	LocationBuildORMAccount   = "build_orm_account"
	LocationDatabaseInsert    = "database_insert"
	LocationORMConversion     = "orm_conversion"
	LocationDTOConversion     = "dto_conversion"
	LocationValidation        = "validation"
	LocationErrorWrapping     = "error_wrapping"
	LocationCauseAnalysis     = "cause_analysis"
)