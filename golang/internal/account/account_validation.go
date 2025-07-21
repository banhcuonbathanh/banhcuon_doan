package account
type CreateUserRequest struct {
	BranchID int64  `json:"branch_id" validate:"required,gt=0"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
	Title    string `json:"title" validate:"required,min=2,max=100"`
	Role     string `json:"role" validate:"required,oneof=admin user manager"`
	OwnerID  int64  `json:"owner_id" validate:"required,gt=0"`
}