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