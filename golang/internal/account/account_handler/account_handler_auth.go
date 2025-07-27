package account_handler

import (
	"log"
	"net/http"
	"strings"

	errorcustom "english-ai-full/internal/error_custom"
	"english-ai-full/internal/mapping"
	dto "english-ai-full/internal/account/account_dto"
	"english-ai-full/internal/model"
	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-playground/validator/v10"
)

// Login handles user authentication
func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	userRes, err := h.userClient.Login(r.Context(), &pb.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Printf("Login error for email %s: %v", req.Email, err)
		authErr := errorcustom.NewAuthenticationError("invalid credentials")
		utils.HandleError(w, authErr)
		return
	}

	user := mapping.ToPBUserRes(userRes)
	accessToken, err := utils.GenerateJWTToken(user)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		log.Printf("Refresh token error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Authentication processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, model.LoginUserRes{
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
	})
}

// Register handles user registration
func (h *AccountHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
	if err := utils.DecodeJSON(r.Body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			utils.HandleValidationErrors(w, validationErrors)
		} else {
			utils.HandleError(w, errorcustom.NewAPIError(
				errorcustom.ErrCodeValidationError,
				"Validation failed",
				http.StatusBadRequest,
			))
		}
		return
	}

	if err := utils.ValidatePasswordWithDetails(req.Password); err != nil {
		utils.HandleError(w, err)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeInternalError,
			"Password processing failed",
			http.StatusInternalServerError,
		))
		return
	}

	userRes, err := h.userClient.Register(r.Context(), &pb.RegisterReq{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	})
	if err != nil {
		log.Printf("Registration error: %v", err)

		if strings.Contains(err.Error(), "already exists") {
			utils.HandleError(w, errorcustom.NewDuplicateEmailError(req.Email))
			return
		}

		utils.HandleError(w, errorcustom.NewAPIError(
			errorcustom.ErrCodeServiceError,
			"Registration failed",
			http.StatusInternalServerError,
		))
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, dto.RegisterResponse{
		ID:     userRes.Id,
		Name:   userRes.Name,
		Email:  userRes.Email,
		Status: userRes.Success,
	})
}

// Logout handles user logout
func (h *AccountHandler) Logout(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}