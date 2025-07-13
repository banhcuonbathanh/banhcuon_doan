package branch

import (
	"context"
	"strconv"

	"english-ai-full/internal/model"
	branchpb "english-ai-full/internal/proto_qr/branch"
	logg "english-ai-full/logger"

	pkgerrors "github.com/pkg/errors"
)

type Service struct {
	repo   *Repository
	logger *logg.Logger
	branchpb.UnimplementedBranchServiceServer
}

func NewBranchService(repo *Repository) *Service {
	return &Service{
		repo:   repo,
		logger: logg.NewLogger(),
	}
}

// CreateBranch creates a new branch
func (s *Service) CreateBranch(ctx context.Context, req *branchpb.CreateBranchRequest) (*branchpb.CreateBranchResponse, error) {
	s.logger.Info("Creating new branch: " + req.Name)

	// Validate required fields
	if req.Name == "" {
		return &branchpb.CreateBranchResponse{}, pkgerrors.New("branch name is required")
	}
	if req.Address == "" {
		return &branchpb.CreateBranchResponse{}, pkgerrors.New("branch address is required")
	}

	branch := model.Branch{
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		ManagerID: req.ManagerId,
	}

	createdBranch, err := s.repo.CreateBranch(ctx, branch)
	if err != nil {
		s.logger.Error("Failed to create branch: " + err.Error())
		return &branchpb.CreateBranchResponse{}, pkgerrors.Wrap(err, "failed to create branch")
	}

	s.logger.Info("Successfully created branch with ID: " + strconv.FormatInt(createdBranch.ID, 10))

	return &branchpb.CreateBranchResponse{
		Id:        createdBranch.ID,
		Name:      createdBranch.Name,
		Address:   createdBranch.Address,
		Phone:     createdBranch.Phone,
		ManagerId: createdBranch.ManagerID,
		CreatedAt: createdBranch.CreatedAt.String(),
		UpdatedAt: createdBranch.UpdatedAt.String(),
	}, nil
}

// GetBranchByID retrieves a branch by its ID
func (s *Service) GetBranchByID(ctx context.Context, req *branchpb.GetBranchByIDRequest) (*branchpb.GetBranchResponse, error) {
	s.logger.Info("Getting branch by ID: " + strconv.FormatInt(req.Id, 10))

	branch, err := s.repo.GetBranchByID(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get branch: " + err.Error())
		return &branchpb.GetBranchResponse{}, pkgerrors.Wrap(err, "failed to get branch")
	}

	return &branchpb.GetBranchResponse{
		Id:        branch.ID,
		Name:      branch.Name,
		Address:   branch.Address,
		Phone:     branch.Phone,
		CreatedAt: branch.CreatedAt.String(),
		UpdatedAt: branch.UpdatedAt.String(),
	}, nil
}

// UpdateBranch updates an existing branch
func (s *Service) UpdateBranch(ctx context.Context, req *branchpb.UpdateBranchRequest) (*branchpb.UpdateBranchResponse, error) {
	s.logger.Info("Updating branch with ID: " + strconv.FormatInt(req.Id, 10))

	branch, err := s.repo.UpdateBranch(ctx, req.Id, model.Branch{
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		ManagerID: req.ManagerId,
	})
	if err != nil {
		s.logger.Error("Failed to update branch: " + err.Error())
		return &branchpb.UpdateBranchResponse{}, pkgerrors.Wrap(err, "failed to update branch")
	}

	return &branchpb.UpdateBranchResponse{
		Id:        branch.ID,
		Name:      branch.Name,
		Address:   branch.Address,
		Phone:     branch.Phone,
		CreatedAt: branch.CreatedAt.String(),
		UpdatedAt: branch.UpdatedAt.String(),
	}, nil
}
