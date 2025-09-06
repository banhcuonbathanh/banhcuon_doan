// logger/core/constants.go
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
	
	// Database fields
	FieldTable        = "table"
	FieldFunction     = "function"
	FieldRowsAffected = "rows_affected"
	FieldSuccess      = "success"
	FieldQuery        = "query"
	FieldQueryTime    = "query_time"
	
	// Error fields
	FieldError      = "error"
	FieldErrorCode  = "error_code"
	FieldErrorType  = "error_type"
	FieldCause      = "cause"
	FieldStackTrace = "stack_trace"
	
	// Validation fields
	FieldStructName     = "struct_name"
	FieldFieldName      = "field_name"
	FieldValidationTag  = "validation_tag"
	FieldValidationMsg  = "validation_message"
	FieldDecodeError    = "decode_error"
	
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

// Log message constants
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
	
	// Database messages
	MsgDatabaseOperationStarted  = "Starting database operation"
	MsgDatabaseOperationFailed   = "Database operation failed"
	MsgDatabaseOperationSuccess  = "Database operation completed successfully"
	MsgDatabaseInsertAttempt     = "Attempting to insert user"
	MsgDatabaseInsertSuccess     = "Successfully created user"
	MsgDatabaseInsertFailed      = "Failed to insert user"
	
	// Service messages
	MsgServiceCallStarted  = "Calling service"
	MsgServiceCallSuccess  = "Service call completed successfully"
	MsgServiceCallFailed   = "Service call failed"
	MsgServiceCallTimeout  = "Service call timeout"
	
	// Validation messages
	MsgValidationStarted    = "Starting validation"
	MsgValidationFailed     = "Validation failed"
	MsgValidationError      = "Validation error"
	MsgStructValidationError = "Struct validation error"
	MsgRequestValidationFailed = "Request validation failed"
	
	// Context messages
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
)

// Test/Debug constants
const (
	TestLoggingVerification = "logging_verification"
	TestDatabaseConnection  = "database_connection"
	TestServiceConnection   = "service_connection"
	TestValidation         = "validation_test"
)

// Error cause constants
const (
	CauseContextCancelled       = "context_cancelled"
	CauseContextTimeout         = "context_timeout"
	CauseValidationFailed       = "validation_failed"
	CauseStructValidationFailed = "struct_validation_failed"
	CauseDecodeRequestFailed    = "decode_request_failed"
	CauseDuplicateEmail         = "duplicate_email"
	CauseDatabaseError          = "database_error"
	CauseServiceError           = "service_error"
	CauseExternalServiceError   = "external_service_error"
	CauseAuthenticationFailed   = "authentication_failed"
	CausePermissionDenied       = "permission_denied"
	CauseTokenExpired           = "token_expired"
	CauseInvalidToken           = "invalid_token"
	CauseEmailSendFailed        = "email_send_failed"
	CauseFileUploadFailed       = "file_upload_failed"
	CauseNetworkError           = "network_error"
	CauseTimeout                = "timeout"
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
)