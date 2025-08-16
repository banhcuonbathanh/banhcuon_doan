// internal/error_custom/codes.go
// Domain-aware error code system
package errorcustom

import "fmt"

// ============================================================================
// DOMAIN CONSTANTS
// ============================================================================

const (
	DomainAccount    = "account"
	DomainCourse  = "course"
	DomainPayment = "payment"
	DomainAuth    = "auth"
	DomainAdmin   = "admin"
	DomainContent = "content"
	DomainSystem  = "system"
	// Add more domains as needed
)

// ============================================================================
// BASE ERROR CODES
// ============================================================================

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
)

// ============================================================================
// DOMAIN-SPECIFIC ERROR CODE GENERATORS
// ============================================================================

// GetNotFoundCode generates domain-specific not found error codes
// func GetNotFoundCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeNotFound
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeNotFound)
// }

// GetValidationCode generates domain-specific validation error codes
// func GetValidationCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeValidation
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeValidation)
// }

// // GetDuplicateCode generates domain-specific duplicate error codes
// func GetDuplicateCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeDuplicate
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeDuplicate)
// }

// // GetAuthenticationCode generates domain-specific authentication error codes
// func GetAuthenticationCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeAuthentication
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeAuthentication)
// }

// // GetAuthorizationCode generates domain-specific authorization error codes
// func GetAuthorizationCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeAuthorization
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeAuthorization)
// }

// // GetBusinessLogicCode generates domain-specific business logic error codes
// func GetBusinessLogicCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeBusinessLogic
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeBusinessLogic)
// }

// // GetExternalServiceCode generates domain-specific external service error codes
// func GetExternalServiceCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeExternalService
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeExternalService)
// }

// GetServiceUnavailableCode generates domain-specific service unavailable error codes
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

// GetInvalidInputCode generates domain-specific invalid input error codes
func GetInvalidInputCode(domain string) string {
	if domain == "" {
		return ErrorTypeInvalidInput
	}
	return fmt.Sprintf("%s_%s", domain, ErrorTypeInvalidInput)
}

// GetRateLimitCode generates domain-specific rate limit error codes
// func GetRateLimitCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeRateLimit
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeRateLimit)
// }

// GetTimeoutCode generates domain-specific timeout error codes
// func GetTimeoutCode(domain string) string {
// 	if domain == "" {
// 		return ErrorTypeTimeout
// 	}
// 	return fmt.Sprintf("%s_%s", domain, ErrorTypeTimeout)
// }

// ============================================================================
// SPECIFIC ERROR CODES (for backward compatibility and specific cases)
// ============================================================================

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

// ============================================================================
// DOMAIN-SPECIFIC CODE MAPS
// ============================================================================

// // GetDomainSpecificCodes returns all possible error codes for a domain
// func GetDomainSpecificCodes(domain string) map[string]string {
// 	codes := make(map[string]string)
	
// 	codes["NOT_FOUND"] = GetNotFoundCode(domain)
// 	codes["VALIDATION_ERROR"] = GetValidationCode(domain)
// 	codes["DUPLICATE"] = GetDuplicateCode(domain)
// 	codes["AUTHENTICATION_ERROR"] = GetAuthenticationCode(domain)
// 	codes["AUTHORIZATION_ERROR"] = GetAuthorizationCode(domain)
// 	codes["BUSINESS_LOGIC_ERROR"] = GetBusinessLogicCode(domain)
// 	codes["EXTERNAL_SERVICE_ERROR"] = GetExternalServiceCode(domain)
// 	codes["SERVICE_UNAVAILABLE"] = GetServiceUnavailableCode(domain)
// 	codes["SYSTEM_ERROR"] = GetSystemErrorCode(domain)
// 	codes["INVALID_INPUT"] = GetInvalidInputCode(domain)
// 	codes["RATE_LIMIT"] = GetRateLimitCode(domain)
// 	codes["TIMEOUT"] = GetTimeoutCode(domain)
	
// 	return codes
// }

// ============================================================================
// CODE UTILITIES
// ============================================================================

// IsErrorCodeForDomain checks if an error code belongs to a specific domain
// func IsErrorCodeForDomain(code, domain string) bool {
// 	if domain == "" {
// 		return true // Generic codes belong to all domains
// 	}
	
// 	domainPrefix := domain + "_"
// 	return code == domain || 
// 		   code[:len(domainPrefix)] == domainPrefix ||
// 		   !containsDomainPrefix(code) // Generic codes
// }

// containsDomainPrefix checks if a code contains any domain prefix
func containsDomainPrefix(code string) bool {
	domains := []string{DomainAccount, DomainCourse, DomainPayment, DomainAuth, DomainAdmin, DomainContent, DomainSystem}
	
	for _, domain := range domains {
		if len(code) > len(domain)+1 && code[:len(domain)+1] == domain+"_" {
			return true
		}
	}
	return false
}

// ExtractDomainFromCode extracts the domain from an error code
// func ExtractDomainFromCode(code string) string {
// 	domains := []string{DomainUser, DomainCourse, DomainPayment, DomainAuth, DomainAdmin, DomainContent, DomainSystem}
	
// 	for _, domain := range domains {
// 		if len(code) > len(domain)+1 && code[:len(domain)+1] == domain+"_" {
// 			return domain
// 		}
// 	}
// 	return "" // Generic/unknown domain
// }

// GetBaseErrorType extracts the base error type from a domain-specific code
// func GetBaseErrorType(code string) string {
// 	domain := ExtractDomainFromCode(code)
// 	if domain == "" {
// 		return code // Already a base type
// 	}
	
// 	return code[len(domain)+1:] // Remove domain prefix
// }

// new 12121

// GetDomainCode generates domain-aware error code
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

// GetDatabaseCode returns domain-aware DATABASE_ERROR code
// func GetDatabaseCode(domain string) string {
// 	return GetDomainCode(ErrorTypeDatabase, domain)
// }

// GetRateLimitCode returns domain-aware RATE_LIMIT_ERROR code
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