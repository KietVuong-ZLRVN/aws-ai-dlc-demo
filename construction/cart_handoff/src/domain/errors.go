package domain

import "errors"

var (
	ErrInvalidHandoffSource      = errors.New("invalid handoff source")
	ErrComboNotFound             = errors.New("combo not found")
	ErrComboAccessDenied         = errors.New("combo access denied")
	ErrComboPortfolioUnavailable = errors.New("combo portfolio service unavailable")
	ErrPlatformCartUnavailable   = errors.New("platform cart service unavailable")
)
