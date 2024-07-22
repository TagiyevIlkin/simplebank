package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/TagiyevIlkin/simplebank/api"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	_ "github.com/TagiyevIlkin/simplebank/doc/statik"
	"github.com/TagiyevIlkin/simplebank/gapi"
	"github.com/TagiyevIlkin/simplebank/mail"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/TagiyevIlkin/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("Cannot connect to the database")
	}

	if config.Enviroment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Msg("Cannot connect to the database")
	}

	// run db migration
	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistrubutor := worker.NewRedisTaskDistrubutor(redisOpt)

	go runTaskProcessor(config, redisOpt, store)
	go runGatewayServer(config, store, taskDistrubutor)

	runGrpcServer(config, store, taskDistrubutor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance:")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up:")
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailSender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskprocessor := worker.NewRedisTaskProcessor(redisOpt, store, mailSender)

	log.Info().Msg("start task processor")

	err := taskprocessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runGrpcServer(config util.Config, store db.Store, taskDistrubutor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistrubutor)
	if err != nil {
		log.Fatal().Msg("Cannot create new grpc server:")
	}

	rpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(rpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)

	// It allows to clients to explore what RPC are available on server and how to call them
	reflection.Register(grpcServer)

	listerner, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("Cannot create listerner:")
	}

	log.Info().Msgf("start gRPC server at %v", listerner.Addr())

	err = grpcServer.Serve(listerner)
	if err != nil {
		log.Fatal().Msg("Cannot start gRPC server")
	}
}

func runGatewayServer(config util.Config, store db.Store, taskDistrubutor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistrubutor)
	if err != nil {
		log.Fatal().Msg("Cannot create new grpc server:")
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
		log.Fatal().Msg("Cannot register  HandlerServer:")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// fs := http.FileServer(http.Dir("./doc/swagger/"))
	statikFs, err := fs.New()

	if err != nil {
		log.Fatal().Msg("Cannot create  static fs:")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	listerner, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("Cannot create  listerner:")
	}

	log.Info().Msgf("start HTTP gateway server at %v", listerner.Addr().String())

	handler := gapi.HttpLogger(mux)
	err = http.Serve(listerner, handler)
	if err != nil {
		log.Fatal().Msg("Cannot start HTTP gateway server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("Cannot create new server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("Cannot connect to the database")
	}
}
