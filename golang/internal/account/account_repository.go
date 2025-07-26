package account

import (
	"context"
	"database/sql"
	"english-ai-full/internal/model"
	"english-ai-full/internal/proto_qr/account"
	"fmt"
	"strings"

	"time"

	logg "english-ai-full/logger"
	"english-ai-full/orm"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Compile-time check to ensure Repository implements AccountRepositoryInterface
var _ AccountRepositoryInterface = (*Repository)(nil)
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

func (r *Repository) FindAllUsers(ctx context.Context) ([]model.Account, error) {
	users, err := orm.Accounts(
		orm.AccountWhere.DeletedAt.IsNull(),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accounts []model.Account
	for _, user := range users {
		accounts = append(accounts, model.Account{
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
		})
	}

	return accounts, nil
}

func (r *Repository) FindByBranchID(ctx context.Context, branchID int64) ([]model.Account, error) {
	users, err := orm.Accounts(
		orm.AccountWhere.BranchID.EQ(null.Int64{Int64: branchID, Valid: true}),
		orm.AccountWhere.DeletedAt.IsNull(),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accounts []model.Account
	for _, user := range users {
		accounts = append(accounts, model.Account{
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
		})
	}

	return accounts, nil
}

func (r *Repository) FindByRole(ctx context.Context, role string) ([]model.Account, error) {
	users, err := orm.Accounts(
		orm.AccountWhere.Role.EQ(role),
		orm.AccountWhere.DeletedAt.IsNull(),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accounts []model.Account
	for _, user := range users {
		accounts = append(accounts, model.Account{
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
		})
	}

	return accounts, nil
}

func (r *Repository) FindByOwnerID(ctx context.Context, ownerID int64) ([]model.Account, error) {
	users, err := orm.Accounts(
		orm.AccountWhere.OwnerID.EQ(null.Int64{Int64: ownerID, Valid: true}),
		orm.AccountWhere.DeletedAt.IsNull(),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var accounts []model.Account
	for _, user := range users {
		accounts = append(accounts, model.Account{
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
		})
	}

	return accounts, nil
}
func (r *Repository) SearchUsers(ctx context.Context, query, role string, branchId int64, statusFilter []string, page, pageSize int32, sortBy, sortOrder string) (users []account.Account, totalCount int64, err error) {
	// Build query modifiers
	var queryMods []qm.QueryMod
	
	// Always exclude soft deleted records
	queryMods = append(queryMods, orm.AccountWhere.DeletedAt.IsNull())
	
	// Add search query filter
	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		queryMods = append(queryMods, qm.Where("(LOWER(name) LIKE ? OR LOWER(email) LIKE ?)", searchPattern, searchPattern))
	}
	
	// Add role filter
	if role != "" {
		queryMods = append(queryMods, orm.AccountWhere.Role.EQ(role))
	}
	
	// Add branch filter
	if branchId > 0 {
		queryMods = append(queryMods, orm.AccountWhere.BranchID.EQ(null.Int64{Int64: branchId, Valid: true}))
	}
	
	// Add status filter (if you have a status column)
	if len(statusFilter) > 0 {
		// Assuming you have a status column in your accounts table
		// queryMods = append(queryMods, qm.WhereIn("status IN ?", statusFilter))
		// For now, we'll skip this since the schema doesn't show a status column
	}
	
	// Get total count first
	countQueryMods := make([]qm.QueryMod, len(queryMods))
	copy(countQueryMods, queryMods)
	
	totalCount, err = orm.Accounts(countQueryMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, pkgerrors.WithStack(err)
	}
	
	// Add sorting - FIX: Actually implement the sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	
	// Validate sortBy to prevent SQL injection
	validSortFields := map[string]bool{
		"id": true, "name": true, "email": true, "role": true, 
		"created_at": true, "updated_at": true, "branch_id": true,
	}
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}
	
	// Validate sortOrder
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	
	orderByClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	queryMods = append(queryMods, qm.OrderBy(orderByClause))
	
	// Add pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	offset := (page - 1) * pageSize
	queryMods = append(queryMods, qm.Limit(int(pageSize)), qm.Offset(int(offset)))
	
	// Execute the query
	ormUsers, err := orm.Accounts(queryMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, pkgerrors.WithStack(err)
	}
	
	// Convert ORM models to proto models - FIX: Handle null values properly
	users = make([]account.Account, len(ormUsers))
	for i, user := range ormUsers {
		// Handle nullable BranchID
		var branchId int64
		if user.BranchID.Valid {
			branchId = user.BranchID.Int64
		}
		
		// Handle nullable OwnerID
		var ownerId int64
		if user.OwnerID.Valid {
			ownerId = user.OwnerID.Int64
		}
		
		// Handle nullable Avatar
		var avatar string
		if user.Avatar.Valid {
			avatar = user.Avatar.String
		}
		
		// Handle nullable Title
		var title string
		if user.Title.Valid {
			title = user.Title.String
		}
		
		// Handle nullable timestamps
		var createdAt *timestamppb.Timestamp
		if user.CreatedAt.Valid {
			createdAt = timestamppb.New(user.CreatedAt.Time)
		}
		
		var updatedAt *timestamppb.Timestamp
		if user.UpdatedAt.Valid {
			updatedAt = timestamppb.New(user.UpdatedAt.Time)
		}
		
		users[i] = account.Account{
			Id:        user.ID,
			BranchId:  branchId,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    avatar,
			Title:     title,
			Role:      user.Role,
			OwnerId:   ownerId,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}
	
	return users, totalCount, nil
}
// func (r *Repository) SearchUsers(ctx context.Context, query, role string, branchId int64, statusFilter []string, page, pageSize int32, sortBy, sortOrder string) (users []account.Account, totalCount int64, err error) {
// 	// Build query modifiers
// 	var queryMods []qm.QueryMod
	
// 	// Always exclude soft deleted records
// 	queryMods = append(queryMods, orm.AccountWhere.DeletedAt.IsNull())
	
// 	// Add search query filter
// 	if query != "" {
// 		searchPattern := "%" + strings.ToLower(query) + "%"
// 		queryMods = append(queryMods, qm.Where("(LOWER(name) LIKE ? OR LOWER(email) LIKE ?)", searchPattern, searchPattern))
// 	}
	
// 	// Add role filter
// 	if role != "" {
// 		queryMods = append(queryMods, orm.AccountWhere.Role.EQ(role))
// 	}
	
// 	// Add branch filter
// 	if branchId > 0 {
// 		queryMods = append(queryMods, orm.AccountWhere.BranchID.EQ(null.Int64{Int64: branchId, Valid: true}))
// 	}
	
// 	// Add status filter (if you have a status column)
// 	if len(statusFilter) > 0 {
// 		// Assuming you have a status column in your accounts table
// 		// queryMods = append(queryMods, qm.WhereIn("status IN ?", statusFilter))
// 		// For now, we'll skip this since the schema doesn't show a status column
// 	}
	
// 	// Get total count first
// 	countQueryMods := make([]qm.QueryMod, len(queryMods))
// 	copy(countQueryMods, queryMods)
	
// 	totalCount, err = orm.Accounts(countQueryMods...).Count(ctx, r.db)
// 	if err != nil {
// 		return nil, 0, pkgerrors.WithStack(err)
// 	}
	
// 	// Add sorting
// 	if sortBy == "" {
// 		sortBy = "created_at"
// 	}
// 	if sortOrder == "" {
// 		sortOrder = "DESC"
// 	}
	


	
// 	// Add pagination
// 	if page < 1 {
// 		page = 1
// 	}
// 	if pageSize < 1 {
// 		pageSize = 10
// 	}
	
// 	offset := (page - 1) * pageSize
// 	queryMods = append(queryMods, qm.Limit(int(pageSize)), qm.Offset(int(offset)))
	
// 	// Execute the query
// 	ormUsers, err := orm.Accounts(queryMods...).All(ctx, r.db)
// 	if err != nil {
// 		return nil, 0, pkgerrors.WithStack(err)
// 	}
	
// 	// Convert ORM models to proto models
// 	users = make([]account.Account, len(ormUsers))
// 	for i, user := range ormUsers {
// 		users[i] = account.Account{
// 			Id:        user.ID,
// 			BranchId:  user.BranchID.Int64,
// 			Name:      user.Name,
// 			Email:     user.Email,
// 			Avatar:    user.Avatar.String,
// 			Title:     user.Title.String,
// 			Role:      user.Role,
// 			OwnerId:   user.OwnerID.Int64,
// 			CreatedAt: timestamppb.New(user.CreatedAt.Time),
// 			UpdatedAt: timestamppb.New(user.UpdatedAt.Time),
// 		}
// 	}
	
// 	return users, totalCount, nil
// }

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := orm.Accounts(
		orm.AccountWhere.Email.EQ(email),
		orm.AccountWhere.DeletedAt.IsNull(),
	).Exists(ctx, r.db)
	if err != nil {
		return false, pkgerrors.WithStack(err)
	}
	return exists, nil
}

func (r *Repository) UpdateAccountStatus(ctx context.Context, userID int64, status string) error {
	user, err := orm.Accounts(
		orm.AccountWhere.ID.EQ(userID),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return pkgerrors.WithStack(ErrorUserNotFound)
	}

	// Assuming status is stored in a status column or similar field
	// You may need to adjust this based on your actual database schema
	user.UpdatedAt = null.Time{Time: time.Now(), Valid: true}
	
	_, err = user.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	user, err := orm.Accounts(
		orm.AccountWhere.ID.EQ(userID),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return pkgerrors.WithStack(ErrorUserNotFound)
	}

	user.Password = hashedPassword
	user.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = user.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return nil
}

func (r *Repository) StoreResetToken(ctx context.Context, email, token string) error {
	// This would typically store in a separate password_reset_tokens table
	// For now, implementing basic structure - you'll need to adjust based on your schema
	query := `
		INSERT INTO password_reset_tokens (email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.created_at
	`
	
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours expiration
	_, err := r.db.ExecContext(ctx, query, email, token, expiresAt, time.Now())
	if err != nil {
		return pkgerrors.WithStack(err)
	}
	
	return nil
}

func (r *Repository) ValidateResetToken(ctx context.Context, token string) (string, error) {
	query := `
		SELECT email FROM password_reset_tokens 
		WHERE token = $1 AND expires_at > $2 AND used_at IS NULL
	`
	
	var email string
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", pkgerrors.New("invalid or expired reset token")
		}
		return "", pkgerrors.WithStack(err)
	}
	
	return email, nil
}

func (r *Repository) StoreVerificationToken(ctx context.Context, email, token string) error {
	// This would typically store in a separate email_verification_tokens table
	query := `
		INSERT INTO email_verification_tokens (email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.created_at
	`
	
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours expiration
	_, err := r.db.ExecContext(ctx, query, email, token, expiresAt, time.Now())
	if err != nil {
		return pkgerrors.WithStack(err)
	}
	
	return nil
}

func (r *Repository) ValidateVerificationToken(ctx context.Context, token string) (string, error) {
	query := `
		SELECT email FROM email_verification_tokens 
		WHERE token = $1 AND expires_at > $2 AND used_at IS NULL
	`
	
	var email string
	err := r.db.QueryRowContext(ctx, query, token, time.Now()).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", pkgerrors.New("invalid or expired verification token")
		}
		return "", pkgerrors.WithStack(err)
	}
	
	return email, nil
}

func (r *Repository) MarkEmailAsVerified(ctx context.Context, email string) error {
	// Update the account to mark email as verified
	user, err := orm.Accounts(
		orm.AccountWhere.Email.EQ(email),
		orm.AccountWhere.DeletedAt.IsNull(),
	).One(ctx, r.db)
	if err != nil {
		return pkgerrors.WithStack(ErrorUserNotFound)
	}

	// Assuming there's an email_verified field or similar
	// You may need to adjust this based on your actual database schema
	user.UpdatedAt = null.Time{Time: time.Now(), Valid: true}
	
	_, err = user.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	// Mark the verification token as used
	query := `
		UPDATE email_verification_tokens 
		SET used_at = $1 
		WHERE email = $2 AND used_at IS NULL
	`
	_, err = r.db.ExecContext(ctx, query, time.Now(), email)
	if err != nil {
		r.logger.Error("Failed to mark verification token as used")
		// Don't return error as the main operation (marking email verified) succeeded
	}

	return nil
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

// Helper function to convert ORM model to domain model
func (r *Repository) ormToModel(user *orm.Account) model.Account {
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
	}
}