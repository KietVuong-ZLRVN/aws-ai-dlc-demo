package valueobjects

// ShopperSession carries the authenticated shopper's session credential.
// This unit has no knowledge of the shopper's profile — only their session token.
type ShopperSession struct {
	SessionToken string
}
