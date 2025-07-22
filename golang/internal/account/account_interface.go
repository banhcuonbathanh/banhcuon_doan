// File: golang/internal/account/interfaces.go
package account

import (
	"context"
	account_dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/model"

	"english-ai-full/internal/proto_qr/account"
	pb "english-ai-full/internal/proto_qr/account"
	"net/http"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ===== REPOSITORY LAYER INTERFACE =====
// Domain layer defines what repository should do
type AccountRepositoryInterface interface {
	// User management
	CreateUser(ctx context.Context, user model.Account) (model.Account, error)
	Register(ctx context.Context, user model.Account) (model.Account, error)
	FindByEmail(ctx context.Context, email string) (model.Account, error)
	FindByID(ctx context.Context, id int64) (model.Account, error)
	FindAllUsers(ctx context.Context) ([]model.Account, error)
	UpdateUser(ctx context.Context, user model.Account) (model.Account, error)
	DeleteUser(ctx context.Context, id int64) error
	
	// Enhanced search and filtering
	FindByBranchID(ctx context.Context, branchID int64) ([]model.Account, error)
	FindByRole(ctx context.Context, role string) ([]model.Account, error)
	FindByOwnerID(ctx context.Context, ownerID int64) ([]model.Account, error)

	SearchUsers(ctx context.Context, query, role string, branchId int64, statusFilter []string, page, pageSize int32, sortBy, sortOrder string) (users []account.Account, totalCount int64, err error)
	// Account verification and status
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateAccountStatus(ctx context.Context, userID int64, status string) error
	
	// Password and token management
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error
	StoreResetToken(ctx context.Context, email, token string) error
	ValidateResetToken(ctx context.Context, token string) (string, error) // Returns email if valid
	StoreVerificationToken(ctx context.Context, email, token string) error
	ValidateVerificationToken(ctx context.Context, token string) (string, error) // Returns email if valid
	MarkEmailAsVerified(ctx context.Context, email string) error
}

// ===== SERVICE LAYER INTERFACE =====
// Application layer defines business logic interface
type AccountServiceInterface interface {
	// Basic gRPC service methods
	CreateUser(ctx context.Context, req *pb.AccountReq) (*pb.Account, error)
	UpdateUser(ctx context.Context, req *pb.UpdateUserReq) (*pb.AccountRes, error)
	DeleteUser(ctx context.Context, req *pb.DeleteAccountReq) (*pb.DeleteAccountRes, error)
	FindAllUsers(ctx context.Context, req *emptypb.Empty) (*pb.AccountList, error)
	FindByEmail(ctx context.Context, req *pb.FindByEmailReq) (*pb.AccountRes, error)
	Login(ctx context.Context, req *pb.LoginReq) (*pb.AccountRes, error)
	Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutRes, error)
	Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error)
	FindByID(ctx context.Context, req *pb.FindByIDReq) (*pb.FindByIDRes, error)
	
	// Password management
	ChangePassword(ctx context.Context, req *pb.ChangePasswordReq) (*pb.ChangePasswordRes, error)
	ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error)
	ForgotPassword(ctx context.Context, req *pb.ForgotPasswordReq) (*pb.ForgotPasswordRes, error)
	
	// Account verification and status
	VerifyEmail(ctx context.Context, req *pb.VerifyEmailReq) (*pb.VerifyEmailRes, error)
	ResendVerification(ctx context.Context, req *pb.ResendVerificationReq) (*pb.ResendVerificationRes, error)
	UpdateAccountStatus(ctx context.Context, req *pb.UpdateAccountStatusReq) (*pb.UpdateAccountStatusRes, error)
	
	// Enhanced search and filtering
	FindByRole(ctx context.Context, req *pb.FindByRoleReq) (*pb.AccountList, error)
	FindByBranch(ctx context.Context, req *pb.FindByBranchReq) (*pb.AccountList, error)
	SearchUsers(ctx context.Context, req *pb.SearchUsersReq) (*pb.AccountList, error)
	
	// Token/Session management
	RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error)
	ValidateToken(ctx context.Context, req *pb.ValidateTokenReq) (*pb.ValidateTokenRes, error)
	
	// Additional business logic methods
	ValidateUserCredentials(ctx context.Context, email, password string) (model.Account, error)
	DeactivateUser(ctx context.Context, userID int64) error
	GetUsersByBranch(ctx context.Context, branchID int64) ([]model.Account, error)
}

// ===== HANDLER LAYER INTERFACE =====
// Presentation layer defines HTTP handler interface
type AccountHandlerInterface interface {
	// Authentication endpoints
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	ValidateToken(w http.ResponseWriter, r *http.Request)
	
	// User management endpoints
	CreateAccount(w http.ResponseWriter, r *http.Request)
	FindAccountByID(w http.ResponseWriter, r *http.Request)
	FindByEmail(w http.ResponseWriter, r *http.Request)
	FindAllUsers(w http.ResponseWriter, r *http.Request)
	UpdateUserByID(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	
	// Password management endpoints
	ChangePassword(w http.ResponseWriter, r *http.Request)
	ForgotPassword(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	
	// Account verification endpoints
	VerifyEmail(w http.ResponseWriter, r *http.Request)
	ResendVerification(w http.ResponseWriter, r *http.Request)
	UpdateAccountStatus(w http.ResponseWriter, r *http.Request)
	
	// Enhanced search and filtering endpoints
	FindByRole(w http.ResponseWriter, r *http.Request)
	FindByBranch(w http.ResponseWriter, r *http.Request)
	SearchUsers(w http.ResponseWriter, r *http.Request)
	
	// Additional endpoints
	GetUserProfile(w http.ResponseWriter, r *http.Request)
	GetUsersByBranch(w http.ResponseWriter, r *http.Request)
}

// ===== USE CASE INTERFACE (Optional - for complex business logic) =====
type AccountUseCaseInterface interface {
	// Core user operations
	RegisterUser(ctx context.Context, req account_dto.RegisterUserRequest) (account_dto.RegisterUserResponse, error)
	AuthenticateUser(ctx context.Context, req account_dto.LoginRequest) (account_dto.LoginResponse, error)
	GetUserProfile(ctx context.Context, userID int64) (account_dto.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, req account_dto.UpdateUserRequest) (account_dto.UpdateUserResponse, error)
	DeactivateUser(ctx context.Context, userID int64) error
	
	// Password management
	ChangeUserPassword(ctx context.Context, req account_dto.ChangePasswordRequest) error
	InitiatePasswordReset(ctx context.Context, email string) (string, error) // Returns reset token
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
	
	// Account verification
	SendEmailVerification(ctx context.Context, email string) error
	VerifyUserEmail(ctx context.Context, token string) error
	ResendEmailVerification(ctx context.Context, email string) error
	
	// Enhanced search and filtering
	GetUsersByBranch(ctx context.Context, branchID int64) ([]account_dto.UserSummary, error)
	GetUsersByRole(ctx context.Context, role string) ([]account_dto.UserSummary, error)
	SearchUsers(ctx context.Context, req account_dto.SearchUsersRequest) (account_dto.SearchUsersResponse, error)
	
	// Token management
	RefreshUserToken(ctx context.Context, refreshToken string) (account_dto.TokenPair, error)
	ValidateUserToken(ctx context.Context, token string) (*account_dto.UserTokenInfo, error)
}

// ===== EXTERNAL DEPENDENCIES INTERFACES =====
// For dependency injection and testing

// Database interface
type DatabaseInterface interface {
	Ping(ctx context.Context) error
	Close() error
	BeginTx(ctx context.Context) (TransactionInterface, error)
}

// Transaction interface
type TransactionInterface interface {
	Commit() error
	Rollback() error
}

// Password hasher interface
type PasswordHasherInterface interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) bool
}

// JWT token interface
type TokenMakerInterface interface {
	CreateToken(user model.Account) (string, error)
	VerifyToken(token string) (*model.Account, error)
	CreateRefreshToken(user model.Account) (string, error)
	ValidateRefreshToken(token string) (*model.Account, error)
	CreateResetToken(email string) (string, error)
	ValidateResetToken(token string) (string, error) // Returns email if valid
	CreateVerificationToken(email string) (string, error)
	ValidateVerificationToken(token string) (string, error) // Returns email if valid
}

// Logger interface
type LoggerInterface interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// ===== NOTE =====
// Request models are defined in: requests.go
// Response models are defined in: responses.go

// ===== VALIDATION INTERFACE =====
type ValidatorInterface interface {
	Validate(req interface{}) error
	ValidateStruct(s interface{}) error
}

// ===== CACHE INTERFACE (if needed) =====
type CacheInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// ===== EMAIL SERVICE INTERFACE (Enhanced) =====
type EmailServiceInterface interface {
	SendWelcomeEmail(ctx context.Context, email, name string) error
	SendPasswordResetEmail(ctx context.Context, email, resetToken string) error
	SendAccountDeactivationEmail(ctx context.Context, email, name string) error
	SendVerificationEmail(ctx context.Context, email, verificationToken string) error
	SendPasswordChangedNotification(ctx context.Context, email, name string) error
}

// ===== NOTIFICATION SERVICE INTERFACE (if needed) =====
type NotificationServiceInterface interface {
	SendNotification(ctx context.Context, userID int64, message string) error
	SendBulkNotification(ctx context.Context, userIDs []int64, message string) error
}

// ===== CONFIGURATION INTERFACE =====
type ConfigInterface interface {
	GetDatabaseURL() string
	GetJWTSecret() string
	GetServerAddress() string
	GetGRPCAddress() string
	GetTokenExpiration() int
	GetRefreshTokenExpiration() int
	GetResetTokenExpiration() int
	GetVerificationTokenExpiration() int
}

// ===== METRICS INTERFACE (for monitoring) =====
type MetricsInterface interface {
	IncrementCounter(name string, tags map[string]string)
	RecordDuration(name string, duration int64, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
}