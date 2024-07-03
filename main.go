package main

import (
	"database/sql"
	"log"

	"github.com/TagiyevIlkin/simplebank/api"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:Ilkin561@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080" // Localhost 8080
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
}
