package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/TagiyevIlkin/simplebank/api"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/gapi"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/util"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	runGrpcServer(config, store)
}

func runGrpcServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot create new grpc server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)

	// It allows to clients to explore what RPC are available on server and how to call them
	reflection.Register(grpcServer)

	listerner, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("Cannot create listerner:", err)
	}

	log.Printf("start gRPC server at %v", listerner.Addr())

	err = grpcServer.Serve(listerner)
	if err != nil {
		log.Fatal("Cannot start gRPC server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot create new server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}
}
