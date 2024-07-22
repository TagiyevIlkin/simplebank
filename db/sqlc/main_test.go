package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testStore Store

func TestMain(m *testing.M) {
	var err error
	conn, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
	connPool, err := pgxpool.New(context.Background(), conn.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}

	testStore = NewStore(connPool)
	os.Exit(m.Run())
}
