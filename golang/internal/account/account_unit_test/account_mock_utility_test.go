package account_unit_test



import (

	"context"
	"errors"
	res "english-ai-full/internal/account/account_handler" 
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
	
	// Create handler using the New function
	handler := res.New(mockClient)
	
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
	
	return &handler, mockClient 
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