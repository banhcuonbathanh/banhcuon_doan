type AccountUseCaseInterface interface {
	// Core user operations
	RegisterUser(ctx context.Context, req account_dto.RegisterUserRequest) (account_dto.RegisterUserResponse, error)
	AuthenticateUser(ctx context.Context, req account_dto.LoginRequest) (account_dto.LoginResponse, error)
	GetUserProfile(ctx context.Context, userID int64) (account_dto.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, req account_dto.UpdateUserRequest) (account_dto.UpdateUserResponse, error)
	DeactivateUser(ctx context.Context, userID int64) error
	
	// Password management
	ChangeUserPassword(ctx context.Context, req account_dto.ChangePasswordRequest) error
	InitiatePasswordReset(ctx context.Context, email string) (string, error) // Returns reset token
	CompletePasswordReset(ctx context.Context, token, newPassword string) error
	
	// Account verification
	SendEmailVerification(ctx context.Context, email string) error
	VerifyUserEmail(ctx context.Context, token string) error
	ResendEmailVerification(ctx context.Context, email string) error
	
	// Enhanced search and filtering
	GetUsersByBranch(ctx context.Context, branchID int64) ([]account_dto.UserSummary, error)
	GetUsersByRole(ctx context.Context, role string) ([]account_dto.UserSummary, error)
	SearchUsers(ctx context.Context, req account_dto.SearchUsersRequest) (account_dto.SearchUsersResponse, error)
	
	// Token management
	RefreshUserToken(ctx context.Context, refreshToken string) (account_dto.TokenPair, error)
	ValidateUserToken(ctx context.Context, token string) (*account_dto.UserTokenInfo, error)
}

