package main

import (
	"database/sql"
	"log"

	"github.com/TagiyevIlkin/simplebank/api"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot create new server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
}
