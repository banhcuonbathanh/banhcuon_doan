package account

import (
	"bytes"
	"context"
	"encoding/json"
	res "english-ai-full/internal/account"
	"errors"

	// "fmt"
	"net/http"
	"net/http/httptest"

	// "strings"
	"testing"
	"time"

	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock for AccountServiceClient
type MockAccountServiceClient struct {
	mock.Mock
}

// Ensure MockAccountServiceClient implements pb.AccountServiceClient interface
var _ pb.AccountServiceClient = (*MockAccountServiceClient)(nil)

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

func (m *MockAccountServiceClient) FindByEmail(ctx context.Context, in *pb.FindByEmailReq, opts ...grpc.CallOption) (*pb.AccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) Login(ctx context.Context, in *pb.LoginReq, opts ...grpc.CallOption) (*pb.AccountRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AccountRes), args.Error(1)
}

func (m *MockAccountServiceClient) Register(ctx context.Context, in *pb.RegisterReq, opts ...grpc.CallOption) (*pb.RegisterRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RegisterRes), args.Error(1)
}

func (m *MockAccountServiceClient) FindByID(ctx context.Context, in *pb.FindByIDReq, opts ...grpc.CallOption) (*pb.FindByIDRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.FindByIDRes), args.Error(1)
}

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
func setupHandlerTest() (*res.Handler, *MockAccountServiceClient) {
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

func TestHandler_Register(t *testing.T) {
	defer cleanupHandlerTest() // Ensure cleanup after all tests
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			requestBody: model.RegisterUserReq{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Register", mock.Anything, &pb.RegisterReq{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "hashed_password123",
				}).Return(&pb.RegisterRes{
					Id:      1,
					Name:    "John Doe",
					Email:   "john@example.com",
					Success: true,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password hashing error",
			requestBody: model.RegisterUserReq{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "error_password",
			},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "service registration error",
			requestBody: model.RegisterUserReq{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Register", mock.Anything, mock.Anything).Return(
					(*pb.RegisterRes)(nil), errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: model.LoginUserReq{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Login", mock.Anything, &pb.LoginReq{
					Email:    "john@example.com",
					Password: "password123",
				}).Return(&pb.AccountRes{
					Account: &pb.Account{
						Id:       1,
						BranchId: 1,
						Name:     "John Doe",
						Email:    "john@example.com",
						Avatar:   "avatar.jpg",
						Title:    "Developer",
						Role:     "user",
						OwnerId:  1,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "login service error",
			requestBody: model.LoginUserReq{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Login", mock.Anything, mock.Anything).Return(
					(*pb.AccountRes)(nil), errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_CreateAccount(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful account creation",
			requestBody: res.CreateUserRequest{
				BranchID: 1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Avatar:   "https://example.com/avatar.jpg",
				Title:    "Developer",
				Role:     "user",
				OwnerID:  1,
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("CreateUser", mock.Anything, mock.Anything).Return(&pb.Account{
					Id:       1,
					BranchId: 1,
					Name:     "John Doe",
					Email:    "john@example.com",
					Avatar:   "https://example.com/avatar.jpg",
					Title:    "Developer",
					Role:     "user",
					OwnerId:  1,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error - missing required fields",
			requestBody: res.CreateUserRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service creation error",
			requestBody: res.CreateUserRequest{
				BranchID: 1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Avatar:   "https://example.com/avatar.jpg",
				Title:    "Developer",
				Role:     "user",
				OwnerID:  1,
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("CreateUser", mock.Anything, mock.Anything).Return(
					(*pb.Account)(nil), errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/accounts", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateAccount(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_FindAccountByID(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful find by ID",
			userID: "1",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByID", mock.Anything, &pb.FindByIDReq{Id: 1}).Return(&pb.FindByIDRes{
					Account: &pb.Account{
						Id:        1,
						BranchId:  1,
						Name:      "John Doe",
						Email:     "john@example.com",
						Avatar:    "avatar.jpg",
						Title:     "Developer",
						Role:      "user",
						OwnerId:   1,
						CreatedAt: timestamppb.New(time.Now()),
						UpdatedAt: timestamppb.New(time.Now()),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing ID parameter",
			userID:         "",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid ID parameter",
			userID:         "invalid",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByID", mock.Anything, &pb.FindByIDReq{Id: 999}).Return(
					(*pb.FindByIDRes)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/accounts/"+tt.userID, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.FindAccountByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_FindByEmail(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:  "successful find by email",
			email: "john@example.com",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByEmail", mock.Anything, &pb.FindByEmailReq{Email: "john@example.com"}).Return(&pb.AccountRes{
					Account: &pb.Account{
						Id:        1,
						BranchId:  1,
						Name:      "John Doe",
						Email:     "john@example.com",
						Avatar:    "avatar.jpg",
						Title:     "Developer",
						Role:      "user",
						OwnerId:   1,
						CreatedAt: timestamppb.New(time.Now()),
						UpdatedAt: timestamppb.New(time.Now()),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "service error",
			email: "notfound@example.com",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByEmail", mock.Anything, &pb.FindByEmailReq{Email: "notfound@example.com"}).Return(
					(*pb.AccountRes)(nil), errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/accounts/email/"+tt.email, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("email", tt.email)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.FindByEmail(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateUserByID(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "1",
			requestBody: model.UpdateUserRequest{
				Name:     "John Updated",
				Email:    "john.updated@example.com",
				BranchID: 1,
				Avatar:   "new_avatar.jpg",
				Title:    "Senior Developer",
				Role:     "admin",
				OwnerID:  1,
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("UpdateUser", mock.Anything, mock.Anything).Return(&pb.AccountRes{
					Account: &pb.Account{
						Id:        1,
						BranchId:  1,
						Name:      "John Updated",
						Email:     "john.updated@example.com",
						Avatar:    "new_avatar.jpg",
						Title:     "Senior Developer",
						Role:      "admin",
						OwnerId:   1,
						CreatedAt: timestamppb.New(time.Now()),
						UpdatedAt: timestamppb.New(time.Now()),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing ID parameter",
			userID:         "",
			requestBody:    model.UpdateUserRequest{},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid ID parameter",
			userID:         "invalid",
			requestBody:    model.UpdateUserRequest{},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: model.UpdateUserRequest{
				Name: "John Updated",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("UpdateUser", mock.Anything, mock.Anything).Return(
					(*pb.AccountRes)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			var body bytes.Buffer
			json.NewEncoder(&body).Encode(tt.requestBody)

			req := httptest.NewRequest(http.MethodPut, "/accounts/"+tt.userID, &body)
			req.Header.Set("Content-Type", "application/json")
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.UpdateUserByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteUser(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful deletion",
			userID: "1",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("DeleteUser", mock.Anything, &pb.DeleteAccountReq{UserID: 1}).Return(&pb.DeleteAccountRes{
					Success: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing ID parameter",
			userID:         "",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid ID parameter",
			userID:         "invalid",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("DeleteUser", mock.Anything, &pb.DeleteAccountReq{UserID: 999}).Return(
					(*pb.DeleteAccountRes)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodDelete, "/accounts/"+tt.userID, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUserProfile(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful profile retrieval",
			userID: "1",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByID", mock.Anything, &pb.FindByIDReq{Id: 1}).Return(&pb.FindByIDRes{
					Account: &pb.Account{
						Id:       1,
						BranchId: 1,
						Name:     "John Doe",
						Email:    "john@example.com",
						Avatar:   "avatar.jpg",
						Title:    "Developer",
						Role:     "user",
						OwnerId:  1,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing ID parameter",
			userID:         "",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByID", mock.Anything, &pb.FindByIDReq{Id: 999}).Return(
					(*pb.FindByIDRes)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/profile/"+tt.userID, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.GetUserProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}
