package token

import "time"

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(usename string, role string, duraton time.Duration) (string, *Payload, error)

	// VerifyToken checks if token or not
	VerifyToken(token string) (*Payload, error)
}
