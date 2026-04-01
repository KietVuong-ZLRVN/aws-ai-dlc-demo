package apperrors

import "errors"

// ErrWishlistUnavailable is returned when the wishlist service cannot be reached.
var ErrWishlistUnavailable = errors.New("wishlist service unavailable")

// ErrAIUnavailable is returned when the AI inference service fails after retries.
var ErrAIUnavailable = errors.New("AI inference service unavailable")
