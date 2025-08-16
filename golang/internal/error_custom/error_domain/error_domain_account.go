package errorcustom 


// User domain error constructors
type AccountDomainErrors struct{}

// NewAccountDomainErrors creates a new user domain error helper
func NewAccountDomainErrors() *AccountDomainErrors {
	return &AccountDomainErrors{}
}
// 
