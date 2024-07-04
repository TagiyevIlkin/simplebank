package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomString(8)

	hashedPasswor, err := HashPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, hashedPasswor)

	err = CheckPassword(password, hashedPasswor)
	require.NoError(t, err)

	wrongPassword := RandomString(6)
	hashedPasswor, err = HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPasswor)

	err = CheckPassword(wrongPassword, hashedPasswor)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}
