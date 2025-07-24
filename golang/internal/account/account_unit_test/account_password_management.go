package account_unit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dto "english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_ChangePassword(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful password change",
			userID: "1",
			requestBody: dto.ChangePasswordRequest{
				UserID:          1,
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ChangePassword", mock.Anything, &pb.ChangePasswordReq{
					UserId:          1,
					CurrentPassword: "oldpassword",
					NewPassword:     "hashed_newpassword123",
				}).Return(&pb.ChangePasswordRes{
					Success: true,
					Message: "Password changed successfully",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			userID:         "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "password hashing error",
			userID: "1",
			requestBody: dto.ChangePasswordRequest{
				UserID:          1,
				CurrentPassword: "oldpassword",
				NewPassword:     "error_password",
			},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "service error - wrong current password",
			userID: "1",
			requestBody: dto.ChangePasswordRequest{
				UserID:          1,
				CurrentPassword: "wrongpassword",
				NewPassword:     "newpassword123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ChangePassword", mock.Anything, mock.Anything).Return(
					(*pb.ChangePasswordRes)(nil), errors.New("current password is incorrect"))
			},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/change-password", &body)
			req.Header.Set("Content-Type", "application/json")
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.ChangePassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_ForgotPassword(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful forgot password",
			requestBody: dto.ForgotPasswordRequest{
				Email: "john@example.com",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ForgotPassword", mock.Anything, &pb.ForgotPasswordReq{
					Email: "john@example.com",
				}).Return(&pb.ForgotPasswordRes{
					Success: true,
					Message: "Password reset email sent",
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
			name: "user not found",
			requestBody: dto.ForgotPasswordRequest{
				Email: "notfound@example.com",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ForgotPassword", mock.Anything, mock.Anything).Return(
					(*pb.ForgotPasswordRes)(nil), errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
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

			req := httptest.NewRequest(http.MethodPost, "/forgot-password", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ForgotPassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_ResetPassword(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful password reset",
			requestBody: dto.ResetPasswordRequest{
				Token:       "valid_reset_token",
				NewPassword: "newpassword123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ResetPassword", mock.Anything, &pb.ResetPasswordReq{
					Token:       "valid_reset_token",
					NewPassword: "hashed_newpassword123",
				}).Return(&pb.ResetPasswordRes{
					Success: true,
					Message: "Password reset successfully",
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
			name: "password hashing error",
			requestBody: dto.ResetPasswordRequest{
				Token:       "valid_reset_token",
				NewPassword: "error_password",
			},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid reset token",
			requestBody: dto.ResetPasswordRequest{
				Token:       "invalid_token",
				NewPassword: "newpassword123",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ResetPassword", mock.Anything, mock.Anything).Return(
					(*pb.ResetPasswordRes)(nil), errors.New("invalid reset token"))
			},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPost, "/reset-password", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ResetPassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}