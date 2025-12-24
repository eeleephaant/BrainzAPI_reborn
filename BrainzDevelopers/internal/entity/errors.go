package entity

import "errors"

var (
	ErrNeed2FA              = errors.New("need 2FA")
	ErrWrongCredentials     = errors.New("wrong credentials")
	ErrUserBanned           = errors.New("user banned")
	ErrEmailAlreadyExists   = errors.New("email already registred")
	ErrEmailNotConfirmed    = errors.New("email is not confirmed")
	ErrEmailCodeExpired     = errors.New("email code expired")
	ErrEmailCodeAlreadyUsed = errors.New("email code already used")
)
