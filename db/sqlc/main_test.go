package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/TagiyevIlkin/simplebank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	conn, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
	testDB, err = sql.Open(conn.DBDriver, conn.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
	testQueries = New(testDB)

	os.Exit(m.Run())
}
