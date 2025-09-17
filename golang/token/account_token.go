// golang/token/account_token.go - Enhanced version with improved logging and error handling
package token

import (
	"context"

	"english-ai-full/error_system"
	"english-ai-full/internal/account/account_dto"
	utils_config "english-ai-full/utils/config"

	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenMakerInterface defines the interface for token operations
type TokenMakerInterface interface {
	CreateToken(user account_dto.Account) (string, error)
	VerifyToken(tokenString string) (*account_dto.Account, error)
	CreateRefreshToken(user account_dto.Account) (string, error)
	ValidateRefreshToken(tokenString string) (*account_dto.Account, error)
	CreateResetToken(email string) (string, error)
	ValidateResetToken(tokenString string) (string, error)
	CreateVerificationToken(email string) (string, error)
	ValidateVerificationToken(tokenString string) (string, error)
}

// Logger interface to abstract logging functionality
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	SetOperation(operation string)
}

// LayerContext interface for request context management
type LayerContext interface {
	BuildOperationContext(operation string, requestID string) map[string]interface{}
	MergeWithContext(fields map[string]interface{}) map[string]interface{}
}

// Global token maker instance with mutex for thread safety
var (
	globalTokenMaker *JWTTokenMaker
	tokenMakerMutex  sync.RWMutex
	isInitialized    bool
	logger           Logger
	layerContext     LayerContext
)

// Function variables for easy mocking in tests
var (
	HashPassword         = hashPassword
	GenerateJWTToken     = generateJWTToken
	GenerateRefreshToken = generateRefreshToken
	ParseToken           = parseToken
)

// SetLogger sets the logger instance for the token package
func SetLogger(l Logger) {
	logger = l
}

// SetLayerContext sets the layer context instance for the token package
func SetLayerContext(lc LayerContext) {
	layerContext = lc
}

// InitializeTokenMaker initializes the global token maker with configuration
func InitializeTokenMaker() error {
	const operation = "initialize_token_maker"
	startTime := time.Now()
	
	tokenMakerMutex.Lock()
	defer tokenMakerMutex.Unlock()

	// Set operation context
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Info("Starting token maker initialization", operationCtx)
	}

	config := utils_config.GetConfig()
	if config == nil {
		err := error_system.NewError(error_system.ErrSystemError, "Configuration not initialized - GetConfig() returned nil")
		
		if logger != nil {
			logger.Error("Configuration is nil during token maker initialization", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return err
	}

	if logger != nil {
		logger.Debug("Configuration found during token maker initialization", operationCtx)
	}

	// Check if JWT config exists
	if config.JWT.SecretKey == "" {
		err := error_system.NewError(error_system.ErrSystemError, "JWT secret key not configured or empty")
		
		if logger != nil {
			logger.Error("JWT secret key is empty in configuration", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return err
	}

	if logger != nil {
		logger.Debug("JWT secret key found in configuration", 
			mergeContext(operationCtx, map[string]interface{}{
				"secret_key_length": len(config.JWT.SecretKey),
			}))
	}

	globalTokenMaker = NewJWTTokenMaker(config.JWT.SecretKey)
	isInitialized = true
	
	if logger != nil {
		logger.Info("Token maker initialized successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}
	
	return nil
}

// IsInitialized returns whether the token maker has been initialized
func IsInitialized() bool {
	tokenMakerMutex.RLock()
	defer tokenMakerMutex.RUnlock()
	return isInitialized
}

// getTokenMaker returns the global token maker with proper error handling
func getTokenMaker() (*JWTTokenMaker, error) {
	const operation = "get_token_maker"
	startTime := time.Now()
	
	tokenMakerMutex.RLock()
	if globalTokenMaker != nil && isInitialized {
		defer tokenMakerMutex.RUnlock()
		return globalTokenMaker, nil
	}
	tokenMakerMutex.RUnlock()

	// Try to initialize if not already done (with write lock)
	tokenMakerMutex.Lock()
	defer tokenMakerMutex.Unlock()

	// Double-check after acquiring write lock
	if globalTokenMaker != nil && isInitialized {
		return globalTokenMaker, nil
	}

	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Info("Token maker not initialized, attempting auto-initialization", operationCtx)
	}

config := utils_config.GetConfig()


// logger.Info("sadfasdfasdfasdfsdfdsafasdfasdfsdsadkajsdhfadhsfkhsdjfhklsaj", map[string]interface{}{
//     "secret_key": config.JWT.SecretKey,
// })
	if config == nil {
		err := error_system.NewError(error_system.ErrSystemError, 
			"Configuration not initialized - please ensure config is loaded before using token functions")
		
		if logger != nil {
			logger.Error("Configuration is nil during auto-initialization", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
					"suggestion":  "ensure utils_config.InitializeConfig() is called before using token functions",
				}))
		}
		return nil, err
	}

	if config.JWT.SecretKey == "" {
		err := error_system.NewError(error_system.ErrSystemError, "JWT secret key not configured")
		
		if logger != nil {
			logger.Error("JWT secret key is empty during auto-initialization", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, err
	}

	globalTokenMaker = NewJWTTokenMaker(config.JWT.SecretKey)
	isInitialized = true
	
	if logger != nil {
		logger.Info("Token maker auto-initialized successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"secret_key_prefix": maskSecretKey(config.JWT.SecretKey),
				"duration_ms":       time.Since(startTime).Milliseconds(),
			}))
	}
	
	return globalTokenMaker, nil
}

// Compare compares a hashed password with a plain password
func Compare(hashedPassword, password string) bool {
	const operation = "compare_password"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	result := err == nil

	if logger != nil {
		if result {
			logger.Debug("Password comparison successful", 
				mergeContext(operationCtx, map[string]interface{}{
					"result":      "match",
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		} else {
			logger.Debug("Password comparison failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"result":      "no_match",
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
	}

	return result
}

// Actual implementation functions
func hashPassword(password string) (string, error) {
	const operation = "hash_password"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Starting password hashing", operationCtx)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to hash password")
		
		if logger != nil {
			logger.Error("Password hashing failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Password hashing successful", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return string(hashedBytes), nil
}

func generateJWTToken(user account_dto.Account) (string, error) {
	const operation = "generate_jwt_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	// Add user context (without sensitive info)
	operationCtx["user_id"] = user.ID
	operationCtx["user_email"] = maskEmail(user.Email)
	operationCtx["user_role"] = user.Role
	operationCtx["branch_id"] = user.BranchID

	if logger != nil {
		logger.SetOperation(operation)
		logger.Info("Generating JWT token", operationCtx)
	}
	
	tokenMaker, err := getTokenMaker()
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to get token maker for JWT generation")
		
		if logger != nil {
			logger.Error("Failed to get token maker for JWT generation", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}
	
	token, err := tokenMaker.CreateToken(user)
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to create JWT token")
		
		if logger != nil {
			logger.Error("Failed to create JWT token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}
	
	if logger != nil {
		logger.Info("JWT token generated successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}
	
	return token, nil
}

func generateRefreshToken(user account_dto.Account) (string, error) {
	const operation = "generate_refresh_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	// Add user context (without sensitive info)
	operationCtx["user_id"] = user.ID
	operationCtx["user_email"] = maskEmail(user.Email)
	operationCtx["user_role"] = user.Role
	operationCtx["branch_id"] = user.BranchID

	if logger != nil {
		logger.SetOperation(operation)
		logger.Info("Generating refresh token", operationCtx)
	}
	
	tokenMaker, err := getTokenMaker()
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to get token maker for refresh token generation")
		
		if logger != nil {
			logger.Error("Failed to get token maker for refresh token generation", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}
	
	token, err := tokenMaker.CreateRefreshToken(user)
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to create refresh token")
		
		if logger != nil {
			logger.Error("Failed to create refresh token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}
	
	if logger != nil {
		logger.Info("Refresh token generated successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}
	
	return token, nil
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	const operation = "parse_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Parsing token", operationCtx)
	}

	tokenMaker, err := getTokenMaker()
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to get token maker for parsing")
		
		if logger != nil {
			logger.Error("Failed to get token maker for token parsing", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}
	
	// Use the secret key from the token maker
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenMaker.secretKey), nil
	})
	
	if err != nil {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Failed to parse token")
		
		if logger != nil {
			logger.Error("Token parsing failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Invalid token claims")
		
		if logger != nil {
			logger.Error("Invalid token claims", 
				mergeContext(operationCtx, map[string]interface{}{
					"claims_ok":   ok,
					"token_valid": token.Valid,
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}

	if logger != nil {
		logger.Debug("Token parsed successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return claims, nil
}

// JWTTokenMaker handles JWT token operations
type JWTTokenMaker struct {
	secretKey                 string
	accessTokenDuration       time.Duration
	refreshTokenDuration      time.Duration
	resetTokenDuration        time.Duration
	verificationTokenDuration time.Duration
}

type JWTClaims struct {
	UserID   int64            `json:"user_id"`
	Email    string           `json:"email"`
	Role     account_dto.Role `json:"role"`
	BranchID int64            `json:"branch_id"`
	jwt.RegisteredClaims
}

func NewJWTTokenMaker(secretKey string) *JWTTokenMaker {
	return &JWTTokenMaker{
		secretKey:                 secretKey,
		accessTokenDuration:       15 * time.Minute,   // 15 minutes
		refreshTokenDuration:      7 * 24 * time.Hour, // 7 days
		resetTokenDuration:        1 * time.Hour,      // 1 hour
		verificationTokenDuration: 24 * time.Hour,     // 24 hours
	}
}

func (maker *JWTTokenMaker) CreateToken(user account_dto.Account) (string, error) {
	const operation = "create_access_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	operationCtx["user_id"] = user.ID
	operationCtx["user_email"] = maskEmail(user.Email)
	operationCtx["token_type"] = "access"
	operationCtx["duration"] = maker.accessTokenDuration.String()

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Creating access token", operationCtx)
	}

	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		BranchID: user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to sign access token")
		
		if logger != nil {
			logger.Error("Failed to sign access token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Access token created successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return tokenString, nil
}

func (maker *JWTTokenMaker) VerifyToken(tokenString string) (*account_dto.Account, error) {
	const operation = "verify_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Verifying token", operationCtx)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	})

	if err != nil {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Failed to parse token for verification")
		
		if logger != nil {
			logger.Error("Token verification parsing failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Invalid token or claims")
		
		if logger != nil {
			logger.Error("Invalid token or claims during verification", 
				mergeContext(operationCtx, map[string]interface{}{
					"claims_ok":   ok,
					"token_valid": token.Valid,
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}

	account := &account_dto.Account{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		BranchID: claims.BranchID,
	}

	if logger != nil {
		logger.Debug("Token verified successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"user_id":     claims.UserID,
				"user_email":  maskEmail(claims.Email),
				"user_role":   claims.Role,
				"branch_id":   claims.BranchID,
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return account, nil
}

func (maker *JWTTokenMaker) CreateRefreshToken(user account_dto.Account) (string, error) {
	const operation = "create_refresh_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	operationCtx["user_id"] = user.ID
	operationCtx["user_email"] = maskEmail(user.Email)
	operationCtx["token_type"] = "refresh"
	operationCtx["duration"] = maker.refreshTokenDuration.String()

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Creating refresh token", operationCtx)
	}

	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		BranchID: user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to sign refresh token")
		
		if logger != nil {
			logger.Error("Failed to sign refresh token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Refresh token created successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return tokenString, nil
}

func (maker *JWTTokenMaker) ValidateRefreshToken(tokenString string) (*account_dto.Account, error) {
	return maker.VerifyToken(tokenString) // Same validation logic
}

func (maker *JWTTokenMaker) CreateResetToken(email string) (string, error) {
	const operation = "create_reset_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	operationCtx["email"] = maskEmail(email)
	operationCtx["token_type"] = "reset"
	operationCtx["duration"] = maker.resetTokenDuration.String()

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Creating reset token", operationCtx)
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.resetTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to sign reset token")
		
		if logger != nil {
			logger.Error("Failed to sign reset token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Reset token created successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return tokenString, nil
}

func (maker *JWTTokenMaker) ValidateResetToken(tokenString string) (string, error) {
	const operation = "validate_reset_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Validating reset token", operationCtx)
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	})

	if err != nil {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Failed to parse reset token")
		
		if logger != nil {
			logger.Error("Reset token parsing failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		appErr := error_system.NewError(error_system.ErrInvalidToken, "Invalid reset token or claims")
		
		if logger != nil {
			logger.Error("Invalid reset token or claims", 
				mergeContext(operationCtx, map[string]interface{}{
					"claims_ok":   ok,
					"token_valid": token.Valid,
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Reset token validated successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"email":       maskEmail(claims.Subject),
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return claims.Subject, nil
}

func (maker *JWTTokenMaker) CreateVerificationToken(email string) (string, error) {
	const operation = "create_verification_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	operationCtx["email"] = maskEmail(email)
	operationCtx["token_type"] = "verification"
	operationCtx["duration"] = maker.verificationTokenDuration.String()

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Creating verification token", operationCtx)
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(maker.verificationTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		appErr := error_system.NewError(error_system.ErrSystemError, "Failed to sign verification token")
		
		if logger != nil {
			logger.Error("Failed to sign verification token", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	if logger != nil {
		logger.Debug("Verification token created successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"expires_at":  claims.ExpiresAt.Time,
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return tokenString, nil
}

func (maker *JWTTokenMaker) ValidateVerificationToken(tokenString string) (string, error) {
	const operation = "validate_verification_token"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Validating verification token", operationCtx)
	}

	// Use the same validation logic as reset token
	email, err := maker.ValidateResetToken(tokenString)
	if err != nil {
		if logger != nil {
			logger.Error("Verification token validation failed", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", err
	}

	if logger != nil {
		logger.Debug("Verification token validated successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"email":       maskEmail(email),
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return email, nil
}

// Utility functions for logging and masking sensitive data

// maskEmail masks the email address for logging purposes
func maskEmail(email string) string {
	if email == "" {
		return "empty"
	}
	
	// Find the @ symbol
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			atIndex = i
			break
		}
	}
	
	if atIndex <= 0 {
		return "invalid_email"
	}
	
	// Show first 2 characters and last part after @
	if atIndex <= 2 {
		return email[:1] + "***" + email[atIndex:]
	}
	
	return email[:2] + "***" + email[atIndex:]
}

// maskSecretKey masks the secret key for logging purposes
func maskSecretKey(secretKey string) string {
	if len(secretKey) < 4 {
		return "***"
	}
	return secretKey[:4] + "..."
}

// mergeContext merges two context maps
func mergeContext(base, additional map[string]interface{}) map[string]interface{} {
	if layerContext != nil {
		return layerContext.MergeWithContext(additional)
	}
	
	// Fallback manual merge if layerContext is not available
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range additional {
		result[k] = v
	}
	return result
}

// Context-aware token operations (for use with request contexts)

// CreateTokenWithContext creates a JWT token with request context
func CreateTokenWithContext(ctx context.Context, user account_dto.Account, requestID string) (string, error) {
	const operation = "create_token_with_context"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, requestID)
	} else {
		operationCtx = map[string]interface{}{
			"operation":  operation,
			"request_id": requestID,
			"timestamp":  startTime,
		}
	}

	// Check context timeout
	if err := ctx.Err(); err != nil {
		appErr := error_system.NewError(error_system.ErrTimeout, "Request context cancelled or timed out")
		
		if logger != nil {
			logger.Error("Context error during token creation", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"user_id":     user.ID,
					"user_email":  maskEmail(user.Email),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", appErr
	}

	return generateJWTToken(user)
}

// VerifyTokenWithContext verifies a token with request context
func VerifyTokenWithContext(ctx context.Context, tokenString, requestID string) (*account_dto.Account, error) {
	const operation = "verify_token_with_context"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, requestID)
	} else {
		operationCtx = map[string]interface{}{
			"operation":  operation,
			"request_id": requestID,
			"timestamp":  startTime,
		}
	}

	// Check context timeout
	if err := ctx.Err(); err != nil {
		appErr := error_system.NewError(error_system.ErrTimeout, "Request context cancelled or timed out")
		
		if logger != nil {
			logger.Error("Context error during token verification", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, appErr
	}

	tokenMaker, err := getTokenMaker()
	if err != nil {
		if logger != nil {
			logger.Error("Failed to get token maker for verification with context", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return nil, err
	}

	return tokenMaker.VerifyToken(tokenString)
}

// Batch token operations for performance optimization

// CreateTokenPair creates both access and refresh tokens
func CreateTokenPair(user account_dto.Account) (accessToken, refreshToken string, err error) {
	const operation = "create_token_pair"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	operationCtx["user_id"] = user.ID
	operationCtx["user_email"] = maskEmail(user.Email)
	operationCtx["user_role"] = user.Role

	if logger != nil {
		logger.SetOperation(operation)
		logger.Info("Creating token pair", operationCtx)
	}

	// Create access token
	accessToken, err = generateJWTToken(user)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to create access token in pair", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", "", err
	}

	// Create refresh token
	refreshToken, err = generateRefreshToken(user)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to create refresh token in pair", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return "", "", err
	}

	if logger != nil {
		logger.Info("Token pair created successfully", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return accessToken, refreshToken, nil
}

// Health check function for token maker
func HealthCheck() error {
	const operation = "token_health_check"
	startTime := time.Now()
	
	var operationCtx map[string]interface{}
	if layerContext != nil {
		operationCtx = layerContext.BuildOperationContext(operation, "system")
	} else {
		operationCtx = map[string]interface{}{
			"operation": operation,
			"timestamp": startTime,
		}
	}

	if logger != nil {
		logger.SetOperation(operation)
		logger.Debug("Performing token maker health check", operationCtx)
	}

	// Check if token maker is initialized
	if !IsInitialized() {
		err := error_system.NewError(error_system.ErrSystemError, "Token maker not initialized")
		
		if logger != nil {
			logger.Error("Token maker health check failed - not initialized", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return err
	}

	// Try to get token maker
	_, err := getTokenMaker()
	if err != nil {
		if logger != nil {
			logger.Error("Token maker health check failed - unable to get instance", 
				mergeContext(operationCtx, map[string]interface{}{
					"error":       err.Error(),
					"duration_ms": time.Since(startTime).Milliseconds(),
				}))
		}
		return err
	}

	if logger != nil {
		logger.Debug("Token maker health check passed", 
			mergeContext(operationCtx, map[string]interface{}{
				"duration_ms": time.Since(startTime).Milliseconds(),
			}))
	}

	return nil
}


