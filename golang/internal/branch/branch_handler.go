package branch

import (
	"context"
	"encoding/json"
	branchpb "english-ai-full/internal/proto_qr/branch"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Handler struct {
	ctx     context.Context
	service branchpb.BranchServiceClient
}

func NewBranchHandler(service branchpb.BranchServiceClient) *Handler {
	return &Handler{
		ctx:     context.Background(),
		service: service,
	}
}

// CreateBranchRequest represents the request to create a new branch
type CreateBranchRequest struct {
	Name      string `json:"name" validate:"required"`
	Address   string `json:"address" validate:"required"`
	Phone     string `json:"phone,omitempty"`
	ManagerID int64  `json:"manager_id,omitempty"`
}

// CreateBranch handles POST requests to create a new branch
func (h *Handler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	response, err := h.service.CreateBranch(r.Context(), &branchpb.CreateBranchRequest{
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		ManagerId: req.ManagerID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetBranchByID handles GET requests to retrieve a branch by ID
func (h *Handler) GetBranchByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid branch ID", http.StatusBadRequest)
		return
	}

	response, err := h.service.GetBranchByID(r.Context(), &branchpb.GetBranchByIDRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
