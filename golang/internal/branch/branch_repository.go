package branch

import (
	"context"
	"database/sql"
	"time"

	"english-ai-full/internal/model"
	logg "english-ai-full/logger"
	"english-ai-full/orm"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	pkgerrors "github.com/pkg/errors"
)

type Repository struct {
	db     *sql.DB
	logger *logg.Logger
}

func NewBranchRepository(db *sql.DB) *Repository {
	return &Repository{
		db:     db,
		logger: logg.NewLogger(),
	}
}

// CreateBranch creates a new branch in the database
func (r *Repository) CreateBranch(ctx context.Context, branch model.Branch) (model.Branch, error) {
	m := &orm.Branch{
		Name:      branch.Name,
		Address:   branch.Address,
		Phone:     null.String{String: branch.Phone, Valid: branch.Phone != ""},
		ManagerID: null.Int64{Int64: branch.ManagerID, Valid: branch.ManagerID != 0},
		CreatedAt: null.Time{Time: time.Now()},
		UpdatedAt: null.Time{Time: time.Now()},
	}

	err := m.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return model.Branch{}, pkgerrors.Wrap(err, "failed to insert branch")
	}

	return model.Branch{
		ID:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		Phone:     m.Phone.String,
		ManagerID: m.ManagerID.Int64,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}, nil
}

// GetBranchByID retrieves a branch by its ID
func (r *Repository) GetBranchByID(ctx context.Context, id int64) (model.Branch, error) {
	m, err := orm.FindBranch(ctx, r.db, id)
	if err != nil {
		return model.Branch{}, pkgerrors.Wrap(err, "failed to find branch")
	}

	return model.Branch{
		ID:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		Phone:     m.Phone.String,
		ManagerID: m.ManagerID.Int64,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}, nil
}

// GetAllBranches retrieves all branches with optional pagination
func (r *Repository) GetAllBranches(ctx context.Context, limit, offset int) ([]model.Branch, int64, error) {
	queryMods := []qm.QueryMod{
		qm.OrderBy("created_at DESC"),
	}

	if limit > 0 {
		queryMods = append(queryMods, qm.Limit(limit))
	}
	if offset > 0 {
		queryMods = append(queryMods, qm.Offset(offset))
	}

	branches, err := orm.Branches(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "failed to get branches")
	}

	total, err := orm.Branches().Count(ctx, r.db)
	if err != nil {
		return nil, 0, pkgerrors.Wrap(err, "failed to count branches")
	}

	var result []model.Branch
	for _, m := range branches {
		result = append(result, model.Branch{
			ID:        m.ID,
			Name:      m.Name,
			Address:   m.Address,
			Phone:     m.Phone.String,
			ManagerID: m.ManagerID.Int64,
			CreatedAt: m.CreatedAt.Time,
			UpdatedAt: m.UpdatedAt.Time,
		})
	}

	return result, total, nil
}

// UpdateBranch updates an existing branch
func (r *Repository) UpdateBranch(ctx context.Context, id int64, branch model.Branch) (model.Branch, error) {
	m, err := orm.FindBranch(ctx, r.db, id)
	if err != nil {
		return model.Branch{}, pkgerrors.Wrap(err, "failed to find branch")
	}

	// Update fields if provided
	if branch.Name != "" {
		m.Name = branch.Name
	}
	if branch.Address != "" {
		m.Address = branch.Address
	}
	if branch.Phone != "" {
		m.Phone = null.String{String: branch.Phone, Valid: true}
	}
	if branch.ManagerID != 0 {
		m.ManagerID = null.Int64{Int64: branch.ManagerID, Valid: true}
	}

	m.UpdatedAt = null.Time{Time: time.Now()}

	_, err = m.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return model.Branch{}, pkgerrors.Wrap(err, "failed to update branch")
	}

	return model.Branch{
		ID:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		Phone:     m.Phone.String,
		ManagerID: m.ManagerID.Int64,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}, nil
}

// DeleteBranch deletes a branch by its ID
func (r *Repository) DeleteBranch(ctx context.Context, id int64) error {
	m, err := orm.FindBranch(ctx, r.db, id)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to find branch")
	}

	_, err = m.Delete(ctx, r.db)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to delete branch")
	}

	return nil
}
