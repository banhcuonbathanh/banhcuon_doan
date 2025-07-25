package account_unit_test




import (
	"bytes"

	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"


	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_Register(t *testing.T) {
	defer cleanupHandlerTest()
	
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

func TestHandler_Logout(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful logout",
			requestBody: dto.LogoutRequest{
				UserID: 1,
				Token:  "valid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Logout", mock.Anything, &pb.LogoutReq{
					UserId: 1,
					Token:  "valid_token",
				}).Return(&pb.LogoutRes{
					Success: true,
					Message: "Logged out successfully",
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
			name: "service error",
			requestBody: dto.LogoutRequest{
				UserID: 1,
				Token:  "invalid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Logout", mock.Anything, mock.Anything).Return(
					(*pb.LogoutRes)(nil), errors.New("invalid token"))
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

			req := httptest.NewRequest(http.MethodPost, "/logout", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_RefreshToken(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful token refresh",
			requestBody: dto.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("RefreshToken", mock.Anything, &pb.RefreshTokenReq{
					RefreshToken: "valid_refresh_token",
				}).Return(&pb.RefreshTokenRes{
					Success:      true,
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					ExpiresAt:    timestamppb.New(time.Now().Add(time.Hour)),
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
			name: "invalid refresh token",
			requestBody: dto.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("RefreshToken", mock.Anything, mock.Anything).Return(
					(*pb.RefreshTokenRes)(nil), errors.New("invalid refresh token"))
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

			req := httptest.NewRequest(http.MethodPost, "/refresh-token", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RefreshToken(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_ValidateToken(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful token validation",
			requestBody: dto.ValidateTokenRequest{
				Token: "valid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ValidateToken", mock.Anything, &pb.ValidateTokenReq{
					Token: "valid_token",
				}).Return(&pb.ValidateTokenRes{
					Valid:     true,
					UserId:    1,
					Message:   "Token is valid",
					ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
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
			name: "invalid token",
			requestBody: dto.ValidateTokenRequest{
				Token: "invalid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ValidateToken", mock.Anything, mock.Anything).Return(
					(*pb.ValidateTokenRes)(nil), errors.New("invalid token"))
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

			req := httptest.NewRequest(http.MethodPost, "/validate-token", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ValidateToken(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}