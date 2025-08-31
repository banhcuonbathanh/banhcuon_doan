// internal/error_custom/codes.go

package errorcustom

import "fmt"

// ============================================================================
// DOMAIN CONSTANTS
// ============================================================================

const (
	DomainAccount    = "account"

	DomainAuth    = "auth"
	DomainAdmin   = "admin"

	DomainSystem  = "system"
	// Add more domains as needed
)



const (
	// Generic error types
	ErrorTypeNotFound      = "NOT_FOUND"
	ErrorTypeValidation    = "VALIDATION_ERROR"
	ErrorTypeDuplicate     = "DUPLICATE"
	ErrorTypeAuthentication = "AUTHENTICATION_ERROR"
	ErrorTypeAuthorization = "AUTHORIZATION_ERROR"
	ErrorTypeBusinessLogic = "BUSINESS_LOGIC_ERROR"
	ErrorTypeExternalService = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrorTypeSystem        = "SYSTEM_ERROR"
	ErrorTypeInvalidInput  = "INVALID_INPUT"
	ErrorTypeRateLimit     = "RATE_LIMIT"
	ErrorTypeTimeout       = "TIMEOUT"
	  ErrorTypeConflict        = "conflict_error" // ✅ Add this

	   ErrorTypeDatabase = "DATABASE"
)


func GetServiceUnavailableCode(domain string) string {
	if domain == "" {
		return ErrorTypeServiceUnavailable
	}
	return fmt.Sprintf("%s_%s", domain, ErrorTypeServiceUnavailable)
}

// GetSystemErrorCode generates domain-specific system error codes
func GetSystemErrorCode(domain string) string {
	if domain == "" {
		return ErrorTypeSystem
	}
	return fmt.Sprintf("%s_%s", domain, ErrorTypeSystem)
}

func GetInvalidInputCode(domain string) string {
	if domain == "" {
		return ErrorTypeInvalidInput
	}
	return fmt.Sprintf("%s_%s", domain, ErrorTypeInvalidInput)
}


const (
	// Legacy/Specific codes - will be phased out in favor of domain-aware codes
	ErrCodeUserNotFound     = "user_NOT_FOUND"
	ErrCodeDuplicateEmail   = "user_DUPLICATE"
	ErrCodeWeakPassword     = "user_WEAK_PASSWORD"
	ErrCodeAuthFailed       = "auth_AUTHENTICATION_ERROR"
	ErrCodeAccessDenied     = "auth_AUTHORIZATION_ERROR"
	ErrCodeInvalidToken     = "auth_INVALID_TOKEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeValidationError  = "VALIDATION_ERROR"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeInternalError    = "SYSTEM_ERROR"
	ErrCodeServiceError     = "EXTERNAL_SERVICE_ERROR"
	ErrCodeRepositoryError  = "system_REPOSITORY_ERROR"
)



func GetDomainCode(baseCode, domain string) string {
	if domain == "" {
		return baseCode
	}
	return fmt.Sprintf("%s_%s", domain, baseCode)
}

// GetNotFoundCode returns domain-aware NOT_FOUND code
func GetNotFoundCode(domain string) string {
	return GetDomainCode(ErrorTypeNotFound, domain)
}

// GetValidationCode returns domain-aware VALIDATION_ERROR code
func GetValidationCode(domain string) string {
	return GetDomainCode(ErrorTypeValidation, domain)
}

// GetDuplicateCode returns domain-aware DUPLICATE code
func GetDuplicateCode(domain string) string {
	return GetDomainCode(ErrorTypeDuplicate, domain)
}

// GetAuthenticationCode returns domain-aware AUTHENTICATION_ERROR code
func GetAuthenticationCode(domain string) string {
	return GetDomainCode(ErrorTypeAuthentication, domain)
}

// GetAuthorizationCode returns domain-aware AUTHORIZATION_ERROR code
func GetAuthorizationCode(domain string) string {
	return GetDomainCode(ErrorTypeAuthorization, domain)
}

// GetBusinessLogicCode returns domain-aware BUSINESS_LOGIC_ERROR code
func GetBusinessLogicCode(domain string) string {
	return GetDomainCode(ErrorTypeBusinessLogic, domain)
}

// GetExternalServiceCode returns domain-aware EXTERNAL_SERVICE_ERROR code
func GetExternalServiceCode(domain string) string {
	return GetDomainCode(ErrorTypeExternalService, domain)
}

// GetSystemCode returns domain-aware SYSTEM_ERROR code
func GetSystemCode(domain string) string {
	return GetDomainCode(ErrorTypeSystem, domain)
}


func GetDatabaseCode(domain string) string {
	return GetDomainCode(ErrorTypeDatabase, domain)
} 


func GetRateLimitCode(domain string) string {
	return GetDomainCode(ErrorTypeRateLimit, domain)
}

// GetTimeoutCode returns domain-aware TIMEOUT_ERROR code
func GetTimeoutCode(domain string) string {
	return GetDomainCode(ErrorTypeTimeout, domain)
}

// ExtractDomainFromCode extracts domain from error code
func ExtractDomainFromCode(code string) string {
	parts := splitErrorCode(code)
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// GetBaseErrorType extracts base error type from domain-aware code
func GetBaseErrorType(code string) string {
	parts := splitErrorCode(code)
	if len(parts) == 2 {
		return parts[1]
	}
	return code
}

// IsErrorCodeForDomain checks if error code belongs to domain
func IsErrorCodeForDomain(code, domain string) bool {
	return ExtractDomainFromCode(code) == domain
}

// GetDomainSpecificCodes returns all error codes for a domain
func GetDomainSpecificCodes(domain string) map[string]string {
	baseTypes := []string{
		ErrorTypeNotFound,
		ErrorTypeValidation,
		ErrorTypeDuplicate,
		ErrorTypeAuthentication,
		ErrorTypeAuthorization,
		ErrorTypeBusinessLogic,
		ErrorTypeExternalService,
		ErrorTypeSystem,
	
		ErrorTypeRateLimit,
		ErrorTypeTimeout,
	}
	
	codes := make(map[string]string)
	for _, baseType := range baseTypes {
		codes[baseType] = GetDomainCode(baseType, domain)
	}
	
	return codes
}

// Helper function to split error code
func splitErrorCode(code string) []string {
	for i, r := range code {
		if r == '_' {
			return []string{code[:i], code[i+1:]}
		}
	}
	return []string{code}
}
// new 131313