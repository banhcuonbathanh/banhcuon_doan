package account_unit_test

import (
	"context"
	"time"

	"english-ai-full/internal/account/account_handler"
	res "english-ai-full/internal/account/account_handler"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"
	"errors"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	utils_config "english-ai-full/utils/config"
)

// Mock for AccountServiceClient
type MockAccountServiceClient struct {
	mock.Mock
}

// Ensure MockAccountServiceClient implements pb.AccountServiceClient interface
var _ pb.AccountServiceClient = (*MockAccountServiceClient)(nil)

// Store original functions for restoration
var (
	originalHashPassword         func(string) (string, error)
	originalGenerateJWTToken     func(model.Account) (string, error)
	originalGenerateRefreshToken func(model.Account) (string, error)
)

// Mock utils functions
func mockHashPassword(password string) (string, error) {
	if password == "error_password" {
		return "", errors.New("hash error")
	}
	return "hashed_" + password, nil
}

func mockGenerateJWTToken(user model.Account) (string, error) {
	if user.Email == "jwt_error@example.com" {
		return "", errors.New("jwt error")
	}
	return "jwt_token_" + user.Email, nil
}

func mockGenerateRefreshToken(user model.Account) (string, error) {
	if user.Email == "refresh_error@example.com" {
		return "", errors.New("refresh token error")
	}
	return "refresh_token_" + user.Email, nil
}

// Setup function to create handler with mocks
func setupHandlerTest() (*res.AccountHandler, *MockAccountServiceClient) {
	mockClient := new(MockAccountServiceClient)
mockConfig := &utils_config.Config{
		Environment: utils_config.EnvTesting,
		AppName:     "Test App",
		Version:     "1.0.0",
		Debug:       true,
		
		Server: utils_config.ServerConfig{
			Address: "localhost",
			Port:    8080,
			GRPCAddress: "localhost",
			GRPCPort: 50051,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout: 120 * time.Second,
		},
		
		Database: utils_config.DatabaseConfig{
			URL: "sqlite:///:memory:",
			Host: "localhost",
			Port: 5432,
			Name: "test_db",
			User: "test",
			Password: "test",
			SSLMode: "disable",
			MaxConnections: 10,
			MaxIdleConns: 5,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		
		Security: utils_config.SecurityConfig{
			MaxLoginAttempts: 5,
			AccountLockoutMinutes: 15,
			SessionTimeout: 24 * time.Hour,
			AllowedOrigins: []string{"*"},
		},
		
		Password: utils_config.PasswordConfig{
			MinLength: 8,
			MaxLength: 128,
			SpecialChars: "!@#$%^&*()",
		},
		
		Pagination: utils_config.PaginationConfig{
			DefaultSize: 10,
			MaxSize: 100,
			Limit: 1000,
		},
		
		JWT: utils_config.JWTConfig{
			SecretKey: "test_secret_key_minimum_32_characters",
			ExpirationHours: 24,
			RefreshTokenExpirationDays: 30,
			Issuer: "test",
			Algorithm: "HS256",
			RefreshThreshold: 2 * time.Hour,
		},
		
		Email: utils_config.EmailConfig{
			VerificationExpiryHours: 24,
			FromAddress: "test@example.com",
			FromName: "Test",
			Templates: utils_config.EmailTemplates{},
		},
		
		RateLimit: utils_config.RateLimitConfig{
			Enabled: false,
			PerMinute: 60,
			PerHour: 3600,
			BurstSize: 10,
			WindowSize: 1 * time.Minute,
		},
		
		ExternalAPIs: utils_config.ExternalAPIConfig{
			Anthropic: utils_config.AnthropicConfig{
				APIKey: "test_key",
				APIURL: "https://api.anthropic.com",
				Timeout: 30 * time.Second,
				MaxRetries: 3,
			},
			QuanAn: utils_config.QuanAnConfig{
				Address: "localhost:8081",
				Timeout: 10 * time.Second,
				MaxRetries: 3,
			},
		},
		
		Logging: utils_config.LoggingConfig{
			Level: "error",
			Format: "json",
			Output: "stdout",
			MaxSize: 100,
			MaxBackups: 3,
			MaxAge: 28,
		},
		
		ValidRoles: []string{"user", "admin"},
		ValidAccountStatuses: []string{"active", "inactive"},
		
		Domains: utils_config.DomainConfig{
			Enabled: []string{"account"},
			Default: "system",
			ErrorTracking: utils_config.DomainErrorTrackingConfig{
				Enabled: true,
				LogLevel: "debug",
			},
			Account: utils_config.DomainAccountConfig{
				MaxLoginAttempts: 5,
				PasswordComplexity: false,
				EmailVerification: false,
			},
		},
		
		ErrorHandling: utils_config.ErrorHandlingConfig{
			IncludeStackTrace: true,
			SanitizeSensitiveData: false,
			RequestIDRequired: false,
		},
	}
	
	
	// Create handler using the New function
handler := account_handler.NewAccountHandler(mockClient, mockConfig)
	
	// Store original functions if not already stored
	if originalHashPassword == nil {
		originalHashPassword = utils.HashPassword
		originalGenerateJWTToken = utils.GenerateJWTToken
		originalGenerateRefreshToken = utils.GenerateRefreshToken
	}
	
	// Mock the utils functions
	utils.HashPassword = mockHashPassword
	utils.GenerateJWTToken = mockGenerateJWTToken
	utils.GenerateRefreshToken = mockGenerateRefreshToken
	
	return handler, mockClient 
}

// Cleanup function to restore original functions
func cleanupHandlerTest() {
	if originalHashPassword != nil {
		utils.HashPassword = originalHashPassword
		utils.GenerateJWTToken = originalGenerateJWTToken
		utils.GenerateRefreshToken = originalGenerateRefreshToken
	}
}

// Authentication-related mock methods
func (m *MockAccountServiceClient) Register(ctx context.Context, in *pb.RegisterReq, opts ...grpc.CallOption) (*pb.RegisterRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RegisterRes), args.Error(1)
}

func (m *MockAccountServiceClient) Login(ctx context.Context, in *pb.LoginReq, opts ...grpc.CallOption) (*pb.AccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) Logout(ctx context.Context, in *pb.LogoutReq, opts ...grpc.CallOption) (*pb.LogoutRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.LogoutRes), args.Error(1)
}

func (m *MockAccountServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenReq, opts ...grpc.CallOption) (*pb.RefreshTokenRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RefreshTokenRes), args.Error(1)
}

func (m *MockAccountServiceClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenReq, opts ...grpc.CallOption) (*pb.ValidateTokenRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ValidateTokenRes), args.Error(1)
}

// User management mock methods
func (m *MockAccountServiceClient) CreateUser(ctx context.Context, in *pb.AccountReq, opts ...grpc.CallOption) (*pb.Account, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Account), args.Error(1)
}

func (m *MockAccountServiceClient) UpdateUser(ctx context.Context, in *pb.UpdateUserReq, opts ...grpc.CallOption) (*pb.AccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) DeleteUser(ctx context.Context, in *pb.DeleteAccountReq, opts ...grpc.CallOption) (*pb.DeleteAccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteAccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) FindAllUsers(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.AccountList, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountList), args.Error(1)
}

func (m *MockAccountServiceClient) FindByID(ctx context.Context, in *pb.FindByIDReq, opts ...grpc.CallOption) (*pb.FindByIDRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.FindByIDRes), args.Error(1)
}

func (m *MockAccountServiceClient) FindByEmail(ctx context.Context, in *pb.FindByEmailReq, opts ...grpc.CallOption) (*pb.AccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) FindByBranch(ctx context.Context, in *pb.FindByBranchReq, opts ...grpc.CallOption) (*pb.AccountList, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountList), args.Error(1)
}

func (m *MockAccountServiceClient) FindByRole(ctx context.Context, in *pb.FindByRoleReq, opts ...grpc.CallOption) (*pb.AccountList, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountList), args.Error(1)
}

// Password and security mock methods
func (m *MockAccountServiceClient) ChangePassword(ctx context.Context, in *pb.ChangePasswordReq, opts ...grpc.CallOption) (*pb.ChangePasswordRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ChangePasswordRes), args.Error(1)
}

func (m *MockAccountServiceClient) ForgotPassword(ctx context.Context, in *pb.ForgotPasswordReq, opts ...grpc.CallOption) (*pb.ForgotPasswordRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ForgotPasswordRes), args.Error(1)
}

func (m *MockAccountServiceClient) ResetPassword(ctx context.Context, in *pb.ResetPasswordReq, opts ...grpc.CallOption) (*pb.ResetPasswordRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ResetPasswordRes), args.Error(1)
}

// Email verification mock methods
func (m *MockAccountServiceClient) VerifyEmail(ctx context.Context, in *pb.VerifyEmailReq, opts ...grpc.CallOption) (*pb.VerifyEmailRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.VerifyEmailRes), args.Error(1)
}

func (m *MockAccountServiceClient) ResendVerification(ctx context.Context, in *pb.ResendVerificationReq, opts ...grpc.CallOption) (*pb.ResendVerificationRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ResendVerificationRes), args.Error(1)
}

// Search and status mock methods
func (m *MockAccountServiceClient) SearchUsers(ctx context.Context, in *pb.SearchUsersReq, opts ...grpc.CallOption) (*pb.SearchUsersRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SearchUsersRes), args.Error(1)
}

func (m *MockAccountServiceClient) UpdateAccountStatus(ctx context.Context, in *pb.UpdateAccountStatusReq, opts ...grpc.CallOption) (*pb.UpdateAccountStatusRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateAccountStatusRes), args.Error(1)
}