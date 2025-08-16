// ============================================================================
// USAGE EXAMPLES IN YOUR LAYERS
// ============================================================================

/*
// Example usage in handlers:
func (h *UserHandler) GetUser(w http.ResponseWriter, r \*http.Request) {
requestID := errorcustom.GetRequestIDFromContext(r.Context())
domain := errorcustom.GetDomainFromContext(r.Context())

    handlerErrorMgr := layer.NewHandlerErrorManager()

    userID, apiErr := handlerErrorMgr.ParseIDParameter(r, "user_id", domain, requestID)
    if apiErr != nil {
        handlerErrorMgr.RespondWithError(w, apiErr, domain, requestID)
        return
    }

    // Continue with business logic...

}

// Example usage in services:
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
serviceErrorMgr := layer.NewServiceErrorManager()

    // Validate business rules
    validations := map[string]func() error{
        "email_unique": func() error {
            if s.repo.EmailExists(user.Email) {
                return errorcustom.NewDuplicateEmailError(user.Email)
            }
            return nil
        },
        "password_strength": func() error {
            return validatePassword(user.Password)
        },
    }

    if apiErr := serviceErrorMgr.ValidateBusinessRules(errorcustom.DomainUser, validations); apiErr != nil {
        return apiErr
    }

    // Repository call with error wrapping
    if err := s.repo.CreateUser(user); err != nil {
        return serviceErrorMgr.WrapRepositoryError(err, errorcustom.DomainUser, "create_user", map[string]interface{}{
            "email": user.Email,
        })
    }

    return nil

}

// Example usage in repositories:
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
repoErrorMgr := layer.NewRepositoryErrorManager()

    user := &User{}
    err := r.db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).
        Scan(&user.ID, &user.Email, &user.Password)

    if err != nil {
        return nil, repoErrorMgr.HandleDatabaseError(err, errorcustom.DomainUser, "users", "get_by_email", map[string]interface{}{
            "email": email,
        })
    }

    return user, nil

}
\*/
