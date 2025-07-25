package account_handler




import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	pb "english-ai-full/internal/proto_qr/account"
	"english-ai-full/utils"

	"github.com/go-chi/chi"
)

func (h Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Call the gRPC service to initiate forgot password
	res, err := h.user.ForgotPassword(ctx, &pb.ForgotPasswordReq{
		Email: req.Email,
	})
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			// For security, don't reveal whether email exists or not
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "If the email exists, a password reset link has been sent",
			})
			return
		}
		http.Error(w, fmt.Sprintf("error processing forgot password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
		"token":   res.ResetToken, // Only for testing - remove in production
	})
}

func (h Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "error processing password", http.StatusInternalServerError)
		return
	}

	// Call the gRPC service to reset password
	res, err := h.user.ResetPassword(ctx, &pb.ResetPasswordReq{
		Token:       req.Token,
		NewPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			http.Error(w, "invalid or expired reset token", http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("error resetting password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from URL parameter
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

	// Define request structure for password change
	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
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

	// Hash new password
	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "error hashing new password", http.StatusInternalServerError)
		return
	}

	// Call the gRPC service to change password
	res, err := h.user.ChangePassword(ctx, &pb.ChangePasswordReq{
		UserId:          id,
		CurrentPassword: req.OldPassword,
		NewPassword:     hashedNewPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "wrong") {
			http.Error(w, "invalid old password", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error changing password: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": res.Success,
		"message": res.Message,
	})
}