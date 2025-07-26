package account_unit_test


import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dto "english-ai-full/internal/account/account_dto"
	pb "english-ai-full/internal/proto_qr/account"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_VerifyEmail(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful email verification",
			requestBody: dto.VerifyEmailRequest{
				VerificationToken: "valid_verification_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("VerifyEmail", mock.Anything, &pb.VerifyEmailReq{
					VerificationToken: "valid_verification_token",
				}).Return(&pb.VerifyEmailRes{
					Success: true,
					Message: "Email verified successfully",
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
			name: "invalid verification token",
			requestBody: dto.VerifyEmailRequest{
				VerificationToken: "invalid_token",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("VerifyEmail", mock.Anything, mock.Anything).Return(
					(*pb.VerifyEmailRes)(nil), errors.New("invalid verification token"))
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

			req := httptest.NewRequest(http.MethodPost, "/verify-email", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.VerifyEmail(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_ResendVerification(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful resend verification",
			requestBody: dto.ResendVerificationRequest{
				Email: "john@example.com",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ResendVerification", mock.Anything, &pb.ResendVerificationReq{
					Email: "john@example.com",
				}).Return(&pb.ResendVerificationRes{
					Success: true,
					Message: "Verification email sent",
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
			requestBody: dto.ResendVerificationRequest{
				Email: "notfound@example.com",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("ResendVerification", mock.Anything, mock.Anything).Return(
					(*pb.ResendVerificationRes)(nil), errors.New("user not found"))
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

			req := httptest.NewRequest(http.MethodPost, "/resend-verification", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ResendVerification(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}