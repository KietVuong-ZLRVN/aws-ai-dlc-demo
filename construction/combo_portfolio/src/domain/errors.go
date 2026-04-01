package domain

import "errors"

var (
	ErrComboNotFound      = errors.New("combo not found")
	ErrComboAccessDenied  = errors.New("combo access denied")
	ErrInvalidItemCount   = errors.New("combo must contain between 2 and 10 items")
	ErrDuplicateItem      = errors.New("combo contains duplicate simpleSku")
	ErrInvalidComboName   = errors.New("combo name must be 1-100 characters")
	ErrShareTokenConflict = errors.New("share token collision, please retry")
)
