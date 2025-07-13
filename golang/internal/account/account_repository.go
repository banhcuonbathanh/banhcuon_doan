package account

import (
	"context"
	"database/sql"
	"english-ai-full/internal/model"
	"time"

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

func NewAccountRepository(db *sql.DB) *Repository {
	return &Repository{
		db:     db,
		logger: logg.NewLogger(),
	}
}

func (r *Repository) CreateUser(ctx context.Context, user model.Account) (model.Account, error) {
	m := &orm.Account{
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Avatar:    null.String{String: user.Avatar},
		Title:     null.String{String: user.Title},
		Role:      string(user.Role),
		OwnerID:   null.Int64{Int64: user.OwnerID},
		CreatedAt: null.Time{Time: user.CreatedAt},
		UpdatedAt: null.Time{Time: user.UpdatedAt},
	}
	err := m.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return model.Account{}, err
	}

	return model.Account{
		ID:        m.ID,
		BranchID:  m.BranchID.Int64,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Avatar:    m.Avatar.String,
		Title:     m.Title.String,
		Role:      model.Role(m.Role),
		OwnerID:   m.OwnerID.Int64,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}, nil
}

func (r *Repository) Register(ctx context.Context, user model.Account) (model.Account, error) {
	m := &orm.Account{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}
	err := m.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return model.Account{}, err
	}

	return model.Account{
		ID:        m.ID,
		BranchID:  m.BranchID.Int64,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Avatar:    m.Avatar.String,
		Title:     m.Title.String,
		Role:      model.Role(m.Role),
		OwnerID:   m.OwnerID.Int64,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (model.Account, error) {
	user, err := orm.Accounts(qm.Where("email = ?", email)).One(ctx, r.db)
	if err != nil {
		return model.Account{}, pkgerrors.WithStack(ErrorUserNotFound)
	}

	return model.Account{
		ID:        user.ID,
		BranchID:  user.BranchID.Int64,
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Avatar:    user.Avatar.String,
		Title:     user.Title.String,
		Role:      model.Role(user.Role),
		OwnerID:   user.OwnerID.Int64,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (model.Account, error) {
	user, err := orm.Accounts(
		orm.AccountWhere.ID.EQ(id),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return model.Account{}, pkgerrors.WithStack(ErrorUserNotFound)
	}

	return model.Account{
		ID:        user.ID,
		BranchID:  user.BranchID.Int64,
		Name:      user.Name,
		Email:     user.Email,
		Avatar:    user.Avatar.String,
		Title:     user.Title.String,
		Role:      model.Role(user.Role),
		OwnerID:   user.OwnerID.Int64,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// DeleteUser performs a soft delete by setting deleted_at timestamp
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	// First check if user exists and is not already deleted
	user, err := orm.Accounts(
		orm.AccountWhere.ID.EQ(id),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return pkgerrors.WithStack(ErrorUserNotFound)
	}

	// Perform soft delete by setting deleted_at
	user.DeletedAt = null.Time{Time: time.Now(), Valid: true}
	user.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = user.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return nil
}

// UpdateUser updates an existing user account
func (r *Repository) UpdateUser(ctx context.Context, user model.Account) (model.Account, error) {
	existingUser, err := orm.Accounts(
		orm.AccountWhere.ID.EQ(user.ID),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return model.Account{}, pkgerrors.WithStack(ErrorUserNotFound)
	}

	// Update fields if provided
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Avatar != "" {
		existingUser.Avatar = null.String{String: user.Avatar, Valid: true}
	}
	if user.Title != "" {
		existingUser.Title = null.String{String: user.Title, Valid: true}
	}
	if user.Role != "" {
		existingUser.Role = string(user.Role)
	}
	if user.BranchID != 0 {
		existingUser.BranchID = null.Int64{Int64: user.BranchID, Valid: true}
	}
	if user.OwnerID != 0 {
		existingUser.OwnerID = null.Int64{Int64: user.OwnerID, Valid: true}
	}

	// Update timestamp
	existingUser.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = existingUser.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return model.Account{}, pkgerrors.WithStack(err)
	}

	return model.Account{
		ID:        existingUser.ID,
		BranchID:  existingUser.BranchID.Int64,
		Name:      existingUser.Name,
		Email:     existingUser.Email,
		Password:  existingUser.Password,
		Avatar:    existingUser.Avatar.String,
		Title:     existingUser.Title.String,
		Role:      model.Role(existingUser.Role),
		OwnerID:   existingUser.OwnerID.Int64,
		CreatedAt: existingUser.CreatedAt.Time,
		UpdatedAt: existingUser.UpdatedAt.Time,
	}, nil
}
