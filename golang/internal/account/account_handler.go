package account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator"
)

var ctx = context.Background()

type Handler struct {
	ctx  context.Context
	user pb.AccountServiceClient
	validator *validator.Validate
}

func New(user pb.AccountServiceClient) Handler {
	return Handler{
		  validator: validator.New(),
		ctx:  context.Background(),
		user: user,
	}
}

type FindByEmailResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h Handler) FindByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := h.user.FindByEmail(ctx, &pb.FindByEmailReq{Email: email})
	if err != nil {
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FindByEmailResponse{
		ID:        user.Account.Id,
		BranchID:  user.Account.BranchId,
		Name:      user.Account.Name,
		Email:     user.Account.Email,
		Avatar:    user.Account.Avatar,
		Title:     user.Account.Title,
		Role:      user.Account.Role,
		OwnerID:   user.Account.OwnerId,
		CreatedAt: user.Account.CreatedAt.AsTime(),
		UpdatedAt: user.Account.UpdatedAt.AsTime(),
	})
}

type RegisterResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status bool   `json:"status"`
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	req.Password = hashedPassword

	userRes, err := h.user.Register(ctx, &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Println("error registering user", err)
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	userRes, err := h.user.Login(ctx, &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("invalid email or password %v", err).Error(), http.StatusUnauthorized)
		return
	}

	user := mapping.ToPBUserRes(userRes)

	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		http.Error(w, "error creating refresh token", http.StatusInternalServerError)
		return
	}

	res := model.LoginUserRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: model.AccountLoginResponse{
			ID:       user.ID,
			BranchID: user.BranchID,
			Name:     user.Name,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Title:    user.Title,
			Role:     string(user.Role),
			OwnerID:  user.OwnerID,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logout successful"))
}

type CreateUserResponse struct {
	BranchID int64  `json:"branch_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	OwnerID  int64  `json:"owner_id"`
}

// func (h Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	var req model.CreateUserRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "error decoding request body", http.StatusBadRequest)
// 		return
// 	}

// 	hashedPassword, err := utils.HashPassword(req.Password)
// 	if err != nil {
// 		http.Error(w, "error hashing password", http.StatusInternalServerError)
// 		return
// 	}
// 	req.Password = hashedPassword

// 	// Call the user service to create the user
// 	userRes, err := h.user.CreateUser(ctx, &pb.AccountReq{
// 		BranchId: req.BranchID,
// 		Name:     req.Name,
// 		Email:    req.Email,
// 		Password: req.Password,
// 		Avatar:   req.Avatar,
// 		Title:    req.Title,
// 		Role:     req.Role,
// 		OwnerId:  req.OwnerID,
// 	})
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(CreateUserResponse{
// 		BranchID: userRes.BranchId,
// 		Name:     userRes.Name,
// 		Email:    userRes.Email,
// 		Avatar:   userRes.Avatar,
// 		Title:    userRes.Title,
// 		Role:     userRes.Role,
// 		OwnerID:  userRes.OwnerId,
// 	})
// }

type FindAccountByIDResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h Handler) FindAccountByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	res, err := h.user.FindByID(ctx, &pb.FindByIDReq{Id: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("error finding user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindAccountByIDResponse{
		ID:        id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	})
}

type UpdateUserResponse struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Title     string    `json:"title"`
	Role      string    `json:"role"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h Handler) UpdateUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, ErrMissingParameter.Error(), http.StatusBadRequest)
		return
	}

	// Parse user ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidParameter.Error(), http.StatusBadRequest)
		return
	}

	// Decode request body
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, ErrDecodeFailed.Error(), http.StatusBadRequest)
		return
	}

	// Set the ID from URL parameter
	req.ID = id

	// Call the service to update the user
	res, err := h.user.UpdateUser(ctx, &pb.UpdateUserReq{
		Id:       req.ID,
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		log.Printf("failed to update user: %v", err)
		http.Error(w, ErrUpdateUserFailed.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UpdateUserResponse{
		ID:        res.Account.Id,
		BranchID:  res.Account.BranchId,
		Name:      res.Account.Name,
		Email:     res.Account.Email,
		Avatar:    res.Account.Avatar,
		Title:     res.Account.Title,
		Role:      res.Account.Role,
		OwnerID:   res.Account.OwnerId,
		CreatedAt: res.Account.CreatedAt.AsTime(),
		UpdatedAt: res.Account.UpdatedAt.AsTime(),
	})
}

type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// Parse user ID
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	// Call the service to delete the user
	res, err := h.user.DeleteUser(ctx, &pb.DeleteAccountReq{UserID: id})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, ErrorUserNotFound.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteUserResponse{
		Success: res.Success,
		Message: "User deleted successfully",
	})
}
//---------------------- new craete account function ----------------------

func (h Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		// Handle validation errors
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", err.Field()))
			case "min":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param()))
			case "max":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param()))
			case "email":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid email address", err.Field()))
			case "gt":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param()))
			case "url":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be a valid URL", err.Field()))
			case "oneof":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param()))
			default:
				validationErrors = append(validationErrors, fmt.Sprintf("%s is invalid", err.Field()))
			}
		}
		
		response := map[string]interface{}{
			"error": "validation failed",
			"details": validationErrors,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	req.Password = hashedPassword

	// Call the user service to create the user
	userRes, err := h.user.CreateUser(ctx, &pb.AccountReq{
		BranchId: req.BranchID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
		Title:    req.Title,
		Role:     req.Role,
		OwnerId:  req.OwnerID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateUserResponse{
		BranchID: userRes.BranchId,
		Name:     userRes.Name,
		Email:    userRes.Email,
		Avatar:   userRes.Avatar,
		Title:    userRes.Title,
		Role:     userRes.Role,
		OwnerID:  userRes.OwnerId,
	})
}

//---------------