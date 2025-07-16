// File: golang/internal/account/interfaces.go
package account

import (
	"context"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"net/http"
)

// ===== REPOSITORY LAYER INTERFACE =====
// Domain layer defines what repository should do
type AccountRepositoryInterface interface {
	// User management
	CreateUser(ctx context.Context, user model.Account) (model.Account, error)
	Register(ctx context.Context, user model.Account) (model.Account, error)
	FindByEmail(ctx context.Context, email string) (model.Account, error)
	FindByID(ctx context.Context, id int64) (model.Account, error)
	UpdateUser(ctx context.Context, user model.Account) (model.Account, error)
	DeleteUser(ctx context.Context, id int64) error
	
	// Additional methods you might need
	FindByBranchID(ctx context.Context, branchID int64) ([]model.Account, error)
	FindByOwnerID(ctx context.Context, ownerID int64) ([]model.Account, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// ===== SERVICE LAYER INTERFACE =====
// Application layer defines business logic interface
type AccountServiceInterface interface {
	// gRPC service methods
	CreateUser(ctx context.Context, req *pb.AccountReq) (*pb.Account, error)
	Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error)
	Login(ctx context.Context, req *pb.LoginReq) (*pb.AccountRes, error)
	FindByEmail(ctx context.Context, req *pb.FindByEmailReq) (*pb.AccountRes, error)
	FindByID(ctx context.Context, req *pb.FindByIDReq) (*pb.FindByIDRes, error)
	UpdateUser(ctx context.Context, req *pb.UpdateUserReq) (*pb.AccountRes, error)
	DeleteUser(ctx context.Context, req *pb.DeleteAccountReq) (*pb.DeleteAccountRes, error)
	
	// Business logic methods
	ValidateUserCredentials(ctx context.Context, email, password string) (model.Account, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
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
	
	// User management endpoints
	CreateAccount(w http.ResponseWriter, r *http.Request)
	FindAccountByID(w http.ResponseWriter, r *http.Request)
	FindByEmail(w http.ResponseWriter, r *http.Request)
	UpdateUserByID(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	
	// Additional endpoints
	GetUserProfile(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
	GetUsersByBranch(w http.ResponseWriter, r *http.Request)
}

// ===== USE CASE INTERFACE (Optional - for complex business logic) =====
type AccountUseCaseInterface interface {
	RegisterUser(ctx context.Context, req RegisterUserRequest) (RegisterUserResponse, error)
	AuthenticateUser(ctx context.Context, req LoginRequest) (LoginResponse, error)
	GetUserProfile(ctx context.Context, userID int64) (UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error)
	DeactivateUser(ctx context.Context, userID int64) error
	GetUsersByBranch(ctx context.Context, branchID int64) ([]UserSummary, error)
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
}

// Logger interface
type LoggerInterface interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// ===== REQUEST/RESPONSE MODELS FOR USE CASES =====
type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	BranchID int64  `json:"branch_id" validate:"required,gt=0"`
}

type RegisterUserResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Success bool   `json:"success"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserProfile `json:"user"`
}

type UserProfile struct {
	ID       int64  `json:"id"`
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	OwnerID  int64  `json:"owner_id"`
}

type UserProfileResponse struct {
	User UserProfile `json:"user"`
}

type UpdateUserRequest struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	BranchID int64  `json:"branch_id"`
}

// type UpdateUserResponse struct {
// 	User UserProfile `json:"user"`
// }

type UserSummary struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

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

// ===== EMAIL SERVICE INTERFACE (if needed) =====
type EmailServiceInterface interface {
	SendWelcomeEmail(ctx context.Context, email, name string) error
	SendPasswordResetEmail(ctx context.Context, email, resetToken string) error
	SendAccountDeactivationEmail(ctx context.Context, email, name string) error
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
}

// ===== METRICS INTERFACE (for monitoring) =====
type MetricsInterface interface {
	IncrementCounter(name string, tags map[string]string)
	RecordDuration(name string, duration int64, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
}