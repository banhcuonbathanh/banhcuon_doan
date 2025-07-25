package account_handler


import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/mapping"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
)

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
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
	var req dto.LoginRequest
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

func (h Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to refresh token
	res, err := h.user.RefreshToken(ctx, &pb.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("error refreshing token: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
		"expires_at":    res.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Call the gRPC service to validate token
	res, err := h.user.ValidateToken(ctx, &pb.ValidateTokenReq{
		Token: token,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf("error validating token: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"valid":      res.Valid,
		"expires_at": res.ExpiresAt,
		"message":    res.Message,
		"id":         res.UserId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}