package account_unit_test
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
			requestBody: dto.CreateUserRequest{
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
			requestBody: dto.CreateUserRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			mockSetup:      func(m *MockAccountServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service creation error",
			requestBody: dto.CreateUserRequest{
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