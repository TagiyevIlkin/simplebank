package db

import (
	"context"
	"testing"

	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {

	account1 := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account1.ID,
		Amount:    util.RandomMoney(),
	}
	entry, err := testStore.CreateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, entry.AccountID, account1.ID)
	require.Equal(t, entry.Amount, arg.Amount)

}
