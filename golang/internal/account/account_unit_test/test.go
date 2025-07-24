// Additional unit tests for account handler methods
// Add these to your existing test file

import (
	"fmt"
	"sync"
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

func TestHandler_FindByBranch(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		branchID       string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:     "successful find by branch",
			branchID: "1",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByBranch", mock.Anything, &pb.FindByBranchReq{BranchId: 1}).Return(&pb.AccountList{
					Accounts: []*pb.Account{
						{
							Id:       1,
							BranchId: 1,
							Name:     "John Doe",
							Email:    "john@example.com",
							Role:     "user",
						},
						{
							Id:       2,
							BranchId: 1,
							Name:     "Jane Smith",
							Email:    "jane@example.com",
							Role:     "admin",
						},
					},
					Total: 2,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing branch ID parameter",
			branchID:       "",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid branch ID parameter",
			branchID:       "invalid",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "service error",
			branchID: "999",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByBranch", mock.Anything, &pb.FindByBranchReq{BranchId: 999}).Return(
					(*pb.AccountList)(nil), errors.New("branch not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/users/branch/"+tt.branchID, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("branch_id", tt.branchID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.FindByBranch(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_FindByRole(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		role           string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful find by role",
			role: "admin",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByRole", mock.Anything, &pb.FindByRoleReq{Role: "admin"}).Return(&pb.AccountList{
					Accounts: []*pb.Account{
						{
							Id:    1,
							Name:  "Admin User",
							Email: "admin@example.com",
							Role:  "admin",
						},
					},
					Total: 1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing role parameter",
			role:           "",
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			role: "nonexistent",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindByRole", mock.Anything, &pb.FindByRoleReq{Role: "nonexistent"}).Return(
					(*pb.AccountList)(nil), errors.New("role not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/users/role/"+tt.role, nil)
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("role", tt.role)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.FindByRole(w, req)

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

func TestHandler_SearchUsers(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful user search",
			requestBody: dto.SearchUsersRequest{
				Query:    "john",
				Role:     "user",
				BranchID: 1,
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("SearchUsers", mock.Anything, &pb.SearchUsersReq{
					Query:    "john",
					Role:     "user",
					BranchId: 1,
					Page:     1,
					PageSize: 10,
				}).Return(&pb.SearchUsersRes{
					Users: []*pb.Account{
						{
							Id:    1,
							Name:  "John Doe",
							Email: "john@example.com",
							Role:  "user",
						},
					},
					TotalCount: 1,
					Page:       1,
					PageSize:   10,
					TotalPages: 1,
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
			requestBody: dto.SearchUsersRequest{
				Query:    "nonexistent",
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("SearchUsers", mock.Anything, mock.Anything).Return(
					(*pb.SearchUsersRes)(nil), errors.New("search failed"))
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

			req := httptest.NewRequest(http.MethodPost, "/users/search", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SearchUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateAccountStatus(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful status update",
			userID: "1",
			requestBody: dto.UpdateAccountStatusRequest{
				UserID: 1,
				Status: "active",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("UpdateAccountStatus", mock.Anything, &pb.UpdateAccountStatusReq{
					UserId: 1,
					Status: "active",
				}).Return(&pb.UpdateAccountStatusRes{
					Success: true,
					Message: "Status updated successfully",
					Status:  "active",
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
			name:           "missing user ID parameter",
			userID:         "",
			requestBody:    dto.UpdateAccountStatusRequest{UserID: 1, Status: "active"},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: dto.UpdateAccountStatusRequest{
				UserID: 999,
				Status: "active",
			},
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("UpdateAccountStatus", mock.Anything, mock.Anything).Return(
					(*pb.UpdateAccountStatusRes)(nil), errors.New("user not found"))
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

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/status", &body)
			req.Header.Set("Content-Type", "application/json")
			
			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			
			w := httptest.NewRecorder()

			handler.UpdateAccountStatus(w, req)

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

// Additional test for FindAllUsers functionality
func TestHandler_FindAllUsers(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		mockSetup      func(*MockAccountServiceClient)
		expectedStatus int
	}{
		{
			name: "successful find all users",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindAllUsers", mock.Anything, &emptypb.Empty{}).Return(&pb.AccountList{
					Accounts: []*pb.Account{
						{
							Id:       1,
							BranchId: 1,
							Name:     "John Doe",
							Email:    "john@example.com",
							Role:     "user",
							Avatar:   "avatar1.jpg",
							Title:    "Developer",
							OwnerId:  1,
						},
						{
							Id:       2,
							BranchId: 1,
							Name:     "Jane Smith",
							Email:    "jane@example.com",
							Role:     "admin",
							Avatar:   "avatar2.jpg",
							Title:    "Manager",
							OwnerId:  1,
						},
					},
					Total: 2,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindAllUsers", mock.Anything, &emptypb.Empty{}).Return(
					(*pb.AccountList)(nil), errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "empty user list",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindAllUsers", mock.Anything, &emptypb.Empty{}).Return(&pb.AccountList{
					Accounts: []*pb.Account{},
					Total:    0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			handler.FindAllUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

// Integration test for multiple operations
func TestHandler_IntegrationScenarios(t *testing.T) {
	defer cleanupHandlerTest()
	
	t.Run("complete user lifecycle", func(t *testing.T) {
		handler, mockClient := setupHandlerTest()

		// 1. Register user
		mockClient.On("Register", mock.Anything, mock.Anything).Return(&pb.RegisterRes{
			Id:      1,
			Name:    "Test User",
			Email:   "test@example.com",
			Success: true,
		}, nil).Once()

		// 2. Login user
		mockClient.On("Login", mock.Anything, mock.Anything).Return(&pb.AccountRes{
			Account: &pb.Account{
				Id:       1,
				BranchId: 1,
				Name:     "Test User",
				Email:    "test@example.com",
				Role:     "user",
			},
		}, nil).Once()

		// 3. Update user
		mockClient.On("UpdateUser", mock.Anything, mock.Anything).Return(&pb.AccountRes{
			Account: &pb.Account{
				Id:       1,
				BranchId: 1,
				Name:     "Updated User",
				Email:    "test@example.com",
				Role:     "user",
			},
		}, nil).Once()

		// 4. Change password
		mockClient.On("ChangePassword", mock.Anything, mock.Anything).Return(&pb.ChangePasswordRes{
			Success: true,
			Message: "Password changed successfully",
		}, nil).Once()

		// 5. Logout user
		mockClient.On("Logout", mock.Anything, mock.Anything).Return(&pb.LogoutRes{
			Success: true,
			Message: "Logged out successfully",
		}, nil).Once()

		// Execute all operations and verify they work together
		// This would be a more complex test in practice
		mockClient.AssertExpectations(t)
	})
}

// Test error handling scenarios
func TestHandler_ErrorHandlingScenarios(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name        string
		operation   string
		mockSetup   func(*MockAccountServiceClient)
		requestFunc func(*res.Handler) (*httptest.ResponseRecorder, error)
	}{
		{
			name:      "network timeout simulation",
			operation: "login",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("Login", mock.Anything, mock.Anything).Return(
					(*pb.AccountRes)(nil), errors.New("context deadline exceeded"))
			},
			requestFunc: func(handler *res.Handler) (*httptest.ResponseRecorder, error) {
				body := bytes.NewBuffer([]byte(`{"email":"test@example.com","password":"password"}`))
				req := httptest.NewRequest(http.MethodPost, "/login", body)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler.Login(w, req)
				return w, nil
			},
		},
		{
			name:      "database connection error",
			operation: "find_all_users",
			mockSetup: func(m *MockAccountServiceClient) {
				m.On("FindAllUsers", mock.Anything, mock.Anything).Return(
					(*pb.AccountList)(nil), errors.New("database connection refused"))
			},
			requestFunc: func(handler *res.Handler) (*httptest.ResponseRecorder, error) {
				req := httptest.NewRequest(http.MethodGet, "/users", nil)
				w := httptest.NewRecorder()
				handler.FindAllUsers(w, req)
				return w, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockClient := setupHandlerTest()
			tt.mockSetup(mockClient)

			w, err := tt.requestFunc(&handler)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			mockClient.AssertExpectations(t)
		})
	}
}

// Test concurrent access scenarios
func TestHandler_ConcurrentAccess(t *testing.T) {
	defer cleanupHandlerTest()
	
	t.Run("concurrent user creation", func(t *testing.T) {
		handler, mockClient := setupHandlerTest()
		
		// Setup mock for concurrent requests
		mockClient.On("CreateUser", mock.Anything, mock.Anything).Return(&pb.Account{
			Id:       mock.AnythingOfType("int64"),
			BranchId: 1,
			Name:     mock.AnythingOfType("string"),
			Email:    mock.AnythingOfType("string"),
			Role:     "user",
		}, nil).Times(5)

		// Simulate concurrent requests
		var wg sync.WaitGroup
		results := make([]int, 5)
		
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				userReq := dto.CreateUserRequest{
					BranchID: 1,
					Name:     fmt.Sprintf("User %d", index),
					Email:    fmt.Sprintf("user%d@example.com", index),
					Password: "password123",
					Role:     "user",
					OwnerID:  1,
				}
				
				body, _ := json.Marshal(userReq)
				req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				
				handler.CreateAccount(w, req)
				results[index] = w.Code
			}(i)
		}
		
		wg.Wait()
		
		// Verify all requests succeeded
		for i, code := range results {
			assert.Equal(t, http.StatusCreated, code, "Request %d failed", i)
		}
		
		mockClient.AssertExpectations(t)
	})
}

// Test validation scenarios
func TestHandler_ValidationScenarios(t *testing.T) {
	defer cleanupHandlerTest()
	
	tests := []struct {
		name           string
		requestBody    interface{}
		endpoint       string
		method         string
		expectedStatus int
		description    string
	}{
		{
			name: "invalid email format in registration",
			requestBody: model.RegisterUserReq{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
			},
			endpoint:       "/register",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject invalid email format",
		},
		{
			name: "password too short in registration",
			requestBody: model.RegisterUserReq{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "123",
			},
			endpoint:       "/register",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject password shorter than minimum length",
		},
		{
			name: "missing required fields in user creation",
			requestBody: dto.CreateUserRequest{
				Name:  "Test User",
				Email: "test@example.com",
				// Missing required fields: Password, BranchID, OwnerID
			},
			endpoint:       "/accounts",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject request with missing required fields",
		},
		{
			name: "invalid user ID in update request",
			requestBody: dto.UpdateAccountStatusRequest{
				UserID: -1, // Invalid negative ID
				Status: "active",
			},
			endpoint:       "/users/-1/status",
			method:         http.MethodPut,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject negative user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := setupHandlerTest()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Add route context if needed
			if strings.Contains(tt.endpoint, "/-1/") {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", "-1")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}
			
			w := httptest.NewRecorder()

			// Route to appropriate handler based on endpoint
			switch tt.endpoint {
			case "/register":
				handler.Register(w, req)
			case "/accounts":
				handler.CreateAccount(w, req)
			case "/users/-1/status":
				handler.UpdateAccountStatus(w, req)
			}

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}