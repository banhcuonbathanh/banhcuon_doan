package account

import "errors"

var (
	ErrorUserNotFound   = errors.New("user not found")
	ErrUpdateUserFailed = errors.New("update user failed")
	ErrMissingParameter = errors.New("missing parameter")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrDecodeFailed     = errors.New("decode failed")
)
