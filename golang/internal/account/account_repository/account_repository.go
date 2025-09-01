package account_repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	error_custom "english-ai-full/error_custom"
	"english-ai-full/internal/account"
	"english-ai-full/internal/account/account_dto"
	"english-ai-full/logger"
	"english-ai-full/orm"
	utils_config "english-ai-full/utils/config"

	"github.com/aarondl/null/v8"                 // Changed from volatiletech
	"github.com/aarondl/sqlboiler/v4/boil"       // Changed from volatiletech
	"github.com/aarondl/sqlboiler/v4/queries/qm" // Changed from volatiletech
)

// Ensure Repository implements the interface
var _ account.AccountRepositoryInterface = (*Repository)(nil)

// Repository handles account data persistence
type Repository struct {
	db                *sql.DB
	logger           *logger.SpecializedLogger
	errorHandler     *error_custom.RepositoryErrorManager
	config           *utils_config.Config
}

// NewAccountRepository creates a new account repository instance
func NewAccountRepository(db *sql.DB) *Repository {
	return &Repository{
		db:           db,
		logger:       logger.NewSpecializedDatabaseLogger(),
		errorHandler: error_custom.NewRepositoryErrorManager(),
		config:       utils_config.GetConfig(),
	}
}


func (r *Repository) CreateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
r.logger.LogDBOperation("create_user", "accounts", true, nil, nil)
	
m := &orm.Account{
    BranchID:  null.Int64{Int64: user.BranchID, Valid: user.BranchID > 0},
    Name:      user.Name,
    Email:     user.Email,
    Password:  user.Password,
    Avatar:    null.String{String: user.Avatar, Valid: user.Avatar != ""},
    Title:     null.String{String: user.Title, Valid: user.Title != ""},
    Role:      string(user.Role),
    OwnerID:   null.Int64{Int64: user.OwnerID, Valid: user.OwnerID > 0},
    Status:    null.String{String: string(user.Status), Valid: string(user.Status) != ""}, // Fixed this line
    CreatedAt: null.Time{Time: time.Now(), Valid: true},
    UpdatedAt: null.Time{Time: time.Now(), Valid: true},
}

	err := m.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		r.logger.LogDBOperation("create_user", "accounts", false, nil, nil)
		return account_dto.Account{}, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "create_user", map[string]interface{}{
				"email": user.Email,
				"name":  user.Name,
			})
	}

	return r.mapORMToDTO(m), nil
}
// Register creates a new user account (alias for CreateUser with additional validation)
func (r *Repository) Register(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
	r.logger.LogDBOperation("register_user", "accounts", 0, true, 1)

	// Check if email already exists
	exists, err := r.ExistsByEmail(ctx, user.Email)
	if err != nil {
		return account_dto.Account{}, err
	}
	if exists {
		return account_dto.Account{}, error_custom.NewDuplicateEmailError(user.Email)
	}

	return r.CreateUser(ctx, user)
}

// FindByEmail finds a user by email address
func (r *Repository) FindByEmail(ctx context.Context, email string) (account_dto.Account, error) {
	r.logger.LogDBOperation("find_by_email", "accounts", 0, true, 1)

	m, err := orm.Accounts(
		qm.Where("email = ?", email),
	).One(ctx, r.db)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.LogDBOperation("find_by_email", "accounts", 0, false, 0)
			return account_dto.Account{}, error_customer.NewUserNotFoundByEmail(email)
		}
		return account_dto.Account{}, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_email", map[string]interface{}{
				"email": email,
			})
	}

	r.logger.LogDBOperation("find_by_email", "accounts", 0, true, 1)
	return r.mapORMToDTO(m), nil
}

// FindByID finds a user by ID
func (r *Repository) FindByID(ctx context.Context, id int64) (account_dto.Account, error) {
	r.logger.LogDBOperation("find_by_id", "accounts", 0, true, 1)

	m, err := orm.FindAccount(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.LogDBOperation("find_by_id", "accounts", 0, false, 0)
			return account_dto.Account{}, error_customer.NewUserNotFoundByID(id)
		}
		return account_dto.Account{}, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_id", map[string]interface{}{
				"user_id": id,
			})
	}

	r.logger.LogDBOperation("find_by_id", "accounts", 0, true, 1)
	return r.mapORMToDTO(m), nil
}

// FindAllUsers retrieves all users
func (r *Repository) FindAllUsers(ctx context.Context) ([]account_dto.Account, error) {
	r.logger.LogDBOperation("find_all_users", "accounts", 0, true, 0)

	models, err := orm.Accounts().All(ctx, r.db)
	if err != nil {
		r.logger.LogDBOperation("find_all_users", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_all_users", nil)
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("find_all_users", "accounts", 0, true, int64(len(users)))
	return users, nil
}

// UpdateUser updates an existing user
func (r *Repository) UpdateUser(ctx context.Context, user account_dto.Account) (account_dto.Account, error) {
	r.logger.LogDBOperation("update_user", "accounts", 0, true, 1)

	m, err := orm.FindAccount(ctx, r.db, user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return account_dto.Account{}, error_customer.NewUserNotFoundByID(user.ID)
		}
		return account_dto.Account{}, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_user", map[string]interface{}{
				"user_id": user.ID,
			})
	}

	// Update fields
	m.BranchID = null.Int64{Int64: user.BranchID, Valid: user.BranchID > 0}
	m.Name = user.Name
	m.Email = user.Email
	if user.Password != "" {
		m.Password = user.Password
	}
	m.Avatar = null.String{String: user.Avatar, Valid: user.Avatar != ""}
	m.Title = null.String{String: user.Title, Valid: user.Title != ""}
	m.Role = string(user.Role)
	m.OwnerID = null.Int64{Int64: user.OwnerID, Valid: user.OwnerID > 0}
	m.Status = string(user.Status)
	m.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = m.Update(ctx, r.db, boil.Infer())
	if err != nil {
		r.logger.LogDBOperation("update_user", "accounts", 0, false, 0)
		return account_dto.Account{}, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_user", map[string]interface{}{
				"user_id": user.ID,
				"email":   user.Email,
			})
	}

	r.logger.LogDBOperation("update_user", "accounts", 0, true, 1)
	return r.mapORMToDTO(m), nil
}

// DeleteUser deletes a user by ID
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	r.logger.LogDBOperation("delete_user", "accounts", 0, true, 1)

	m, err := orm.FindAccount(ctx, r.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return error_customer.NewUserNotFoundByID(id)
		}
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "delete_user", map[string]interface{}{
				"user_id": id,
			})
	}

	_, err = m.Delete(ctx, r.db)
	if err != nil {
		r.logger.LogDBOperation("delete_user", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "delete_user", map[string]interface{}{
				"user_id": id,
			})
	}

	r.logger.LogDBOperation("delete_user", "accounts", 0, true, 1)
	return nil
}

// ===== ENHANCED SEARCH AND FILTERING =====

// FindByBranchID finds all users in a specific branch
func (r *Repository) FindByBranchID(ctx context.Context, branchID int64) ([]account_dto.Account, error) {
	r.logger.LogDBOperation("find_by_branch_id", "accounts", 0, true, 0)

	models, err := orm.Accounts(
		qm.Where("branch_id = ?", branchID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("find_by_branch_id", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_branch_id", map[string]interface{}{
				"branch_id": branchID,
			})
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("find_by_branch_id", "accounts", 0, true, int64(len(users)))
	return users, nil
}

// FindByBranchWithPagination finds users in a branch with pagination
func (r *Repository) FindByBranchWithPagination(ctx context.Context, branchID int64, offset, limit int) ([]account_dto.Account, int64, error) {
	r.logger.LogDBOperation("find_by_branch_paginated", "accounts", 0, true, 0)

	// Get total count
	totalCount, err := orm.Accounts(
		qm.Where("branch_id = ?", branchID),
	).Count(ctx, r.db)
	if err != nil {
		return nil, 0, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "count_by_branch", map[string]interface{}{
				"branch_id": branchID,
			})
	}

	// Get paginated results
	models, err := orm.Accounts(
		qm.Where("branch_id = ?", branchID),
		qm.OrderBy("created_at DESC"),
		qm.Offset(offset),
		qm.Limit(limit),
	).All(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("find_by_branch_paginated", "accounts", 0, false, 0)
		return nil, 0, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_branch_paginated", map[string]interface{}{
				"branch_id": branchID,
				"offset":    offset,
				"limit":     limit,
			})
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("find_by_branch_paginated", "accounts", 0, true, int64(len(users)))
	return users, totalCount, nil
}

// FindByRole finds all users with a specific role
func (r *Repository) FindByRole(ctx context.Context, role string) ([]account_dto.Account, error) {
	r.logger.LogDBOperation("find_by_role", "accounts", 0, true, 0)

	models, err := orm.Accounts(
		qm.Where("role = ?", role),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("find_by_role", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_role", map[string]interface{}{
				"role": role,
			})
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("find_by_role", "accounts", 0, true, int64(len(users)))
	return users, nil
}

// FindByOwnerID finds all users owned by a specific owner
func (r *Repository) FindByOwnerID(ctx context.Context, ownerID int64) ([]account_dto.Account, error) {
	r.logger.LogDBOperation("find_by_owner_id", "accounts", 0, true, 0)

	models, err := orm.Accounts(
		qm.Where("owner_id = ?", ownerID),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("find_by_owner_id", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "find_by_owner_id", map[string]interface{}{
				"owner_id": ownerID,
			})
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("find_by_owner_id", "accounts", 0, true, int64(len(users)))
	return users, nil
}

// SearchUsers performs advanced search with multiple filters
func (r *Repository) SearchUsers(ctx context.Context, query, role string, branchId int64, statusFilter []string, page, pageSize int32, sortBy, sortOrder string) (users []account.Account, totalCount int64, err error) {
	r.logger.LogDBOperation("search_users", "accounts", 0, true, 0)

	// Build query conditions
	var conditions []qm.QueryMod

	// Text search in name and email
	if query != "" {
		conditions = append(conditions, qm.Where("(name ILIKE ? OR email ILIKE ?)", "%"+query+"%", "%"+query+"%"))
	}

	// Role filter
	if role != "" {
		conditions = append(conditions, qm.Where("role = ?", role))
	}

	// Branch filter
	if branchId > 0 {
		conditions = append(conditions, qm.Where("branch_id = ?", branchId))
	}

	// Status filter
	if len(statusFilter) > 0 {
		placeholders := strings.Repeat("?,", len(statusFilter)-1) + "?"
		args := make([]interface{}, len(statusFilter))
		for i, status := range statusFilter {
			args[i] = status
		}
		conditions = append(conditions, qm.WhereIn("status IN ("+placeholders+")", args...))
	}

	// Get total count
	totalCount, err = orm.Accounts(conditions...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "search_users_count", map[string]interface{}{
				"query":        query,
				"role":         role,
				"branch_id":    branchId,
				"status_filter": statusFilter,
			})
	}

	// Add sorting
	orderClause := "created_at DESC" // default
	if sortBy != "" {
		direction := "ASC"
		if strings.ToUpper(sortOrder) == "DESC" {
			direction = "DESC"
		}
		orderClause = fmt.Sprintf("%s %s", sortBy, direction)
	}
	conditions = append(conditions, qm.OrderBy(orderClause))

	// Add pagination
	offset := (page - 1) * pageSize
	conditions = append(conditions, qm.Offset(int(offset)), qm.Limit(int(pageSize)))

	// Execute query
	models, err := orm.Accounts(conditions...).All(ctx, r.db)
	if err != nil {
		r.logger.LogDBOperation("search_users", "accounts", 0, false, 0)
		return nil, 0, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "search_users", map[string]interface{}{
				"query":        query,
				"role":         role,
				"branch_id":    branchId,
				"status_filter": statusFilter,
				"page":         page,
				"page_size":    pageSize,
			})
	}

	// Convert to proto format
	users = make([]account.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToProto(m)
	}

	r.logger.LogDBOperation("search_users", "accounts", 0, true, int64(len(users)))
	return users, totalCount, nil
}

// ===== ACCOUNT VERIFICATION AND STATUS =====

// ExistsByEmail checks if a user exists with the given email
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	r.logger.LogDBOperation("exists_by_email", "accounts", 0, true, 1)

	exists, err := orm.Accounts(
		qm.Where("email = ?", email),
	).Exists(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("exists_by_email", "accounts", 0, false, 0)
		return false, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "exists_by_email", map[string]interface{}{
				"email": email,
			})
	}

	r.logger.LogDBOperation("exists_by_email", "accounts", 0, true, 1)
	return exists, nil
}

// UpdateAccountStatus updates the status of a user account
func (r *Repository) UpdateAccountStatus(ctx context.Context, userID int64, status string) error {
	r.logger.LogDBOperation("update_account_status", "accounts", 0, true, 1)

	m, err := orm.FindAccount(ctx, r.db, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return error_customer.NewUserNotFoundByID(userID)
		}
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_account_status", map[string]interface{}{
				"user_id": userID,
				"status":  status,
			})
	}

	m.Status = status
	m.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = m.Update(ctx, r.db, boil.Whitelist("status", "updated_at"))
	if err != nil {
		r.logger.LogDBOperation("update_account_status", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_account_status", map[string]interface{}{
				"user_id": userID,
				"status":  status,
			})
	}

	r.logger.LogDBOperation("update_account_status", "accounts", 0, true, 1)
	return nil
}

// ===== PASSWORD AND TOKEN MANAGEMENT =====

// UpdatePassword updates a user's password
func (r *Repository) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	r.logger.LogDBOperation("update_password", "accounts", 0, true, 1)

	m, err := orm.FindAccount(ctx, r.db, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return error_customer.NewUserNotFoundByID(userID)
		}
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_password", map[string]interface{}{
				"user_id": userID,
			})
	}

	m.Password = hashedPassword
	m.UpdatedAt = null.Time{Time: time.Now(), Valid: true}

	_, err = m.Update(ctx, r.db, boil.Whitelist("password", "updated_at"))
	if err != nil {
		r.logger.LogDBOperation("update_password", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_password", map[string]interface{}{
				"user_id": userID,
			})
	}

	r.logger.LogDBOperation("update_password", "accounts", 0, true, 1)
	return nil
}

// StoreResetToken stores a password reset token
func (r *Repository) StoreResetToken(ctx context.Context, email, token string) error {
	r.logger.LogDBOperation("store_reset_token", "password_reset_tokens", 0, true, 1)

	// Note: This assumes you have a password_reset_tokens table
	// You might need to create an ORM model for this table
	query := `
		INSERT INTO password_reset_tokens (email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) 
		DO UPDATE SET token = $2, expires_at = $3, created_at = $4`

	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	_, err := r.db.ExecContext(ctx, query, email, token, expiresAt, time.Now())

	if err != nil {
		r.logger.LogDBOperation("store_reset_token", "password_reset_tokens", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "password_reset_tokens", "store_reset_token", map[string]interface{}{
				"email": email,
			})
	}

	r.logger.LogDBOperation("store_reset_token", "password_reset_tokens", 0, true, 1)
	return nil
}

// ValidateResetToken validates a password reset token and returns the email
func (r *Repository) ValidateResetToken(ctx context.Context, token string) (string, error) {
	r.logger.LogDBOperation("validate_reset_token", "password_reset_tokens", 0, true, 1)

	var email string
	var expiresAt time.Time

	query := `SELECT email, expires_at FROM password_reset_tokens WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).Scan(&email, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.LogDBOperation("validate_reset_token", "password_reset_tokens", 0, false, 0)
			return "", error_customer.NewAuthenticationError("account", "invalid_reset_token")
		}
		return "", r.errorHandler.HandleDatabaseError(
			err, "account", "password_reset_tokens", "validate_reset_token", map[string]interface{}{
				"token": token,
			})
	}

	// Check if token has expired
	if time.Now().After(expiresAt) {
		r.logger.LogDBOperation("validate_reset_token", "password_reset_tokens", 0, false, 0)
		return "", error_customer.NewAuthenticationError("account", "expired_reset_token")
	}

	r.logger.LogDBOperation("validate_reset_token", "password_reset_tokens", 0, true, 1)
	return email, nil
}

// StoreVerificationToken stores an email verification token
func (r *Repository) StoreVerificationToken(ctx context.Context, email, token string) error {
	r.logger.LogDBOperation("store_verification_token", "email_verification_tokens", 0, true, 1)

	query := `
		INSERT INTO email_verification_tokens (email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) 
		DO UPDATE SET token = $2, expires_at = $3, created_at = $4`

	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	_, err := r.db.ExecContext(ctx, query, email, token, expiresAt, time.Now())

	if err != nil {
		r.logger.LogDBOperation("store_verification_token", "email_verification_tokens", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "email_verification_tokens", "store_verification_token", map[string]interface{}{
				"email": email,
			})
	}

	r.logger.LogDBOperation("store_verification_token", "email_verification_tokens", 0, true, 1)
	return nil
}

// ValidateVerificationToken validates an email verification token
func (r *Repository) ValidateVerificationToken(ctx context.Context, token string) (string, error) {
	r.logger.LogDBOperation("validate_verification_token", "email_verification_tokens", 0, true, 1)

	var email string
	var expiresAt time.Time

	query := `SELECT email, expires_at FROM email_verification_tokens WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).Scan(&email, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.LogDBOperation("validate_verification_token", "email_verification_tokens", 0, false, 0)
			return "", error_customer.NewAuthenticationError("account", "invalid_verification_token")
		}
		return "", r.errorHandler.HandleDatabaseError(
			err, "account", "email_verification_tokens", "validate_verification_token", map[string]interface{}{
				"token": token,
			})
	}

	// Check if token has expired
	if time.Now().After(expiresAt) {
		r.logger.LogDBOperation("validate_verification_token", "email_verification_tokens", 0, false, 0)
		return "", error_customer.NewAuthenticationError("account", "expired_verification_token")
	}

	r.logger.LogDBOperation("validate_verification_token", "email_verification_tokens", 0, true, 1)
	return email, nil
}

// MarkEmailAsVerified marks an email as verified
func (r *Repository) MarkEmailAsVerified(ctx context.Context, email string) error {
	r.logger.LogDBOperation("mark_email_verified", "accounts", 0, true, 1)

	// Update the user's email verification status
	_, err := r.db.ExecContext(ctx, 
		"UPDATE accounts SET email_verified = true, updated_at = $1 WHERE email = $2",
		time.Now(), email)

	if err != nil {
		r.logger.LogDBOperation("mark_email_verified", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "mark_email_verified", map[string]interface{}{
				"email": email,
			})
	}

	// Delete the verification token
	_, err = r.db.ExecContext(ctx, 
		"DELETE FROM email_verification_tokens WHERE email = $1", email)
	if err != nil {
		// Log but don't fail - the main operation succeeded
		r.logger.Warn("Failed to delete verification token", map[string]interface{}{
			"email": email,
			"error": err.Error(),
		})
	}

	r.logger.LogDBOperation("mark_email_verified", "accounts", 0, true, 1)
	return nil
}

// ===== TOKEN CLEANUP =====

// CleanupExpiredTokens removes expired tokens from both reset and verification tables
func (r *Repository) CleanupExpiredTokens(ctx context.Context) error {
	r.logger.LogDBOperation("cleanup_expired_tokens", "tokens", 0, true, 0)

	// Clean up expired password reset tokens
	_, err := r.db.ExecContext(ctx, 
		"DELETE FROM password_reset_tokens WHERE expires_at < $1", time.Now())
	if err != nil {
		r.logger.LogDBOperation("cleanup_expired_tokens", "password_reset_tokens", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "password_reset_tokens", "cleanup_expired_tokens", nil)
	}

	// Clean up expired email verification tokens
	_, err = r.db.ExecContext(ctx, 
		"DELETE FROM email_verification_tokens WHERE expires_at < $1", time.Now())
	if err != nil {
		r.logger.LogDBOperation("cleanup_expired_tokens", "email_verification_tokens", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "email_verification_tokens", "cleanup_expired_tokens", nil)
	}

	r.logger.LogDBOperation("cleanup_expired_tokens", "tokens", 0, true, 0)
	return nil
}

// DeleteResetToken removes a specific reset token (typically after password reset)
func (r *Repository) DeleteResetToken(ctx context.Context, email string) error {
	r.logger.LogDBOperation("delete_reset_token", "password_reset_tokens", 0, true, 1)

	_, err := r.db.ExecContext(ctx, 
		"DELETE FROM password_reset_tokens WHERE email = $1", email)
	if err != nil {
		r.logger.LogDBOperation("delete_reset_token", "password_reset_tokens", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "password_reset_tokens", "delete_reset_token", map[string]interface{}{
				"email": email,
			})
	}

	r.logger.LogDBOperation("delete_reset_token", "password_reset_tokens", 0, true, 1)
	return nil
}

// ===== BATCH OPERATIONS =====

// UpdateMultipleUserStatus updates status for multiple users
func (r *Repository) UpdateMultipleUserStatus(ctx context.Context, userIDs []int64, status string) error {
	r.logger.LogDBOperation("update_multiple_user_status", "accounts", 0, true, int64(len(userIDs)))

	if len(userIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := strings.Repeat("?,", len(userIDs)-1) + "?"
	args := make([]interface{}, len(userIDs)+2)
	args[0] = status
	args[1] = time.Now()
	for i, id := range userIDs {
		args[i+2] = id
	}

	query := fmt.Sprintf("UPDATE accounts SET status = $1, updated_at = $2 WHERE id IN (%s)", placeholders)
	_, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		r.logger.LogDBOperation("update_multiple_user_status", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "update_multiple_user_status", map[string]interface{}{
				"user_ids": userIDs,
				"status":   status,
			})
	}

	r.logger.LogDBOperation("update_multiple_user_status", "accounts", 0, true, int64(len(userIDs)))
	return nil
}

// DeleteMultipleUsers deletes multiple users by IDs
func (r *Repository) DeleteMultipleUsers(ctx context.Context, userIDs []int64) error {
	r.logger.LogDBOperation("delete_multiple_users", "accounts", 0, true, int64(len(userIDs)))

	if len(userIDs) == 0 {
		return nil
	}

	// Build placeholders for IN clause
	placeholders := strings.Repeat("?,", len(userIDs)-1) + "?"
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM accounts WHERE id IN (%s)", placeholders)
	result, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		r.logger.LogDBOperation("delete_multiple_users", "accounts", 0, false, 0)
		return r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "delete_multiple_users", map[string]interface{}{
				"user_ids": userIDs,
			})
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Warn("Could not get rows affected for delete_multiple_users", map[string]interface{}{
			"error": err.Error(),
		})
	}

	r.logger.LogDBOperation("delete_multiple_users", "accounts", 0, true, rowsAffected)
	return nil
}

// ===== STATISTICS AND ANALYTICS =====

// GetUserCountByRole returns count of users grouped by role
func (r *Repository) GetUserCountByRole(ctx context.Context) (map[string]int64, error) {
	r.logger.LogDBOperation("get_user_count_by_role", "accounts", 0, true, 0)

	query := `SELECT role, COUNT(*) as count FROM accounts GROUP BY role ORDER BY role`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.LogDBOperation("get_user_count_by_role", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "get_user_count_by_role", nil)
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		result[role] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	r.logger.LogDBOperation("get_user_count_by_role", "accounts", 0, true, int64(len(result)))
	return result, nil
}

// GetUserCountByBranch returns count of users grouped by branch
func (r *Repository) GetUserCountByBranch(ctx context.Context) (map[int64]int64, error) {
	r.logger.LogDBOperation("get_user_count_by_branch", "accounts", 0, true, 0)

	query := `SELECT branch_id, COUNT(*) as count FROM accounts WHERE branch_id IS NOT NULL GROUP BY branch_id ORDER BY branch_id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.LogDBOperation("get_user_count_by_branch", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "get_user_count_by_branch", nil)
	}
	defer rows.Close()

	result := make(map[int64]int64)
	for rows.Next() {
		var branchID, count int64
		if err := rows.Scan(&branchID, &count); err != nil {
			return nil, err
		}
		result[branchID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	r.logger.LogDBOperation("get_user_count_by_branch", "accounts", 0, true, int64(len(result)))
	return result, nil
}

// GetRecentUsers returns recently created users
func (r *Repository) GetRecentUsers(ctx context.Context, limit int) ([]account_dto.Account, error) {
	r.logger.LogDBOperation("get_recent_users", "accounts", 0, true, 0)

	models, err := orm.Accounts(
		qm.OrderBy("created_at DESC"),
		qm.Limit(limit),
	).All(ctx, r.db)

	if err != nil {
		r.logger.LogDBOperation("get_recent_users", "accounts", 0, false, 0)
		return nil, r.errorHandler.HandleDatabaseError(
			err, "account", "accounts", "get_recent_users", map[string]interface{}{
				"limit": limit,
			})
	}

	users := make([]account_dto.Account, len(models))
	for i, m := range models {
		users[i] = r.mapORMToDTO(m)
	}

	r.logger.LogDBOperation("get_recent_users", "accounts", 0, true, int64(len(users)))
	return users, nil
}

// ===== HELPER METHODS =====

// mapORMToDTO converts ORM model to DTO
func (r *Repository) mapORMToDTO(m *orm.Account) account_dto.Account {
	return account_dto.Account{
		ID:        m.ID,
		BranchID:  m.BranchID.Int64,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Avatar:    m.Avatar.String,
		Title:     m.Title.String,
		Role:      account_dto.Role(m.Role),
		OwnerID:   m.OwnerID.Int64,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Time,
		UpdatedAt: m.UpdatedAt.Time,
	}
}

// mapORMToProto converts ORM model to Proto format
func (r *Repository) mapORMToProto(m *orm.Account) account.Account {
	return account.Account{
		Id:        m.ID,
		BranchId:  m.BranchID.Int64,
		Name:      m.Name,
		Email:     m.Email,
		Avatar:    m.Avatar.String,
		Title:     m.Title.String,
		Role:      m.Role,
		OwnerId:   m.OwnerID.Int64,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Time.Unix(),
		UpdatedAt: m.UpdatedAt.Time.Unix(),
	}
}

// mapDTOToORM converts DTO to ORM model (helper for complex operations)
func (r *Repository) mapDTOToORM(dto account_dto.Account) *orm.Account {
	return &orm.Account{
		ID:        dto.ID,
		BranchID:  null.Int64{Int64: dto.BranchID, Valid: dto.BranchID > 0},
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  dto.Password,
		Avatar:    null.String{String: dto.Avatar, Valid: dto.Avatar != ""},
		Title:     null.String{String: dto.Title, Valid: dto.Title != ""},
		Role:      string(dto.Role),
		OwnerID:   null.Int64{Int64: dto.OwnerID, Valid: dto.OwnerID > 0},
		Status:    dto.Status,
		CreatedAt: null.Time{Time: dto.CreatedAt, Valid: !dto.CreatedAt.IsZero()},
		UpdatedAt: null.Time{Time: dto.UpdatedAt, Valid: !dto.UpdatedAt.IsZero()},
	}
}