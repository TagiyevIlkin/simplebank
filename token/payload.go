package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrinvalidToken = errors.New("Token is invalid")
	ErrExpiredToken = errors.New("Token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// Valid checks if the tokern payload is valied or not
func (payload *Payload) Valid() error {

	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}

// NewPayload creates a new token payload with a specific userrname and duration
func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {

	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := Payload{
		ID:        tokenId,
		Username:  username,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return &payload, nil
}
