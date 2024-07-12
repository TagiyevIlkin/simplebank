package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/TagiyevIlkin/simplebank/api"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/gapi"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
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

	go runGatewayServer(config, store)

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

func runGatewayServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("Cannot create new grpc server:", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal("Cannot register  HandlerServer:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listerner, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("Cannot create  listerner:", err)
	}

	log.Printf("start HTTP gateway server at %v", listerner.Addr())

	err = http.Serve(listerner, mux)
	if err != nil {
		log.Fatal("Cannot start HTTP gateway server", err)
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
