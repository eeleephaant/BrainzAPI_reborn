package entity

import "errors"

var (
	ErrNeed2FA            = errors.New("need 2FA")
	ErrWrongCredentials   = errors.New("wrong credentials")
	ErrUserBanned         = errors.New("user banned")
	ErrEmailAlreadyExists = errors.New("email already registred")
)
