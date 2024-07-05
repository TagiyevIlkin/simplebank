package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretyKeySize = 32

// JWTMaker is a JSON WEB Token maker
type JWTMaker struct {
	secretKey string
}

func (maker *JWTMaker) CreateToken(usename string, duraton time.Duration) (token string, err error) {

	payload, err := NewPayload(usename, duraton)
	if err != nil {
		return "", err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))

}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {

	keyFunct := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrinvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunct)

	if err != nil {

		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}

		return nil, ErrinvalidToken
	}

	paload, ok := jwtToken.Claims.(*Payload)

	if !ok {
		return nil, ErrinvalidToken
	}

	return paload, nil

}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretyKeySize {
		return nil, fmt.Errorf("Invalid key size: must be at least %d characters", minSecretyKeySize)
	}

	return &JWTMaker{
		secretKey,
	}, nil
}
