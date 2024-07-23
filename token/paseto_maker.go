package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// Pasetomaker is a PASETO token maker
type Pasetomaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func (maker *Pasetomaker) CreateToken(usename string, role string, duraton time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(usename, role, duraton)
	if err != nil {
		return "", payload, err
	}
	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	if err != nil {
		return "", payload, err
	}

	return token, payload, err

}

func (maker *Pasetomaker) VerifyToken(token string) (*Payload, error) {

	payload := &Payload{}
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrinvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}
	return payload, nil
}

// NewPasetoMaker creates a new PasetorMaker
func NewPasetoMaker(symmetricKey string) (Maker, error) {

	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("Invalid key size: must %d characters", minSecretyKeySize)
	}

	maker := &Pasetomaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}
	return maker, nil
}
