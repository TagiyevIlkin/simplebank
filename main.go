package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

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

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot connect to the database")
	}

	if config.Enviroment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to the database")
	}

	// run db migration
	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistrubutor := worker.NewRedisTaskDistrubutor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGatewayServer(ctx, waitGroup, config, store, taskDistrubutor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistrubutor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Msg("error from wait group")
	}
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

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	redisOpt asynq.RedisClientOpt,
	store db.Store) {
	mailSender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskprocessor := worker.NewRedisTaskProcessor(redisOpt, store, mailSender)

	log.Info().Msg("start task processor")

	err := taskprocessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		taskprocessor.Shutdown()
		log.Info().Msg("task processor is stopped")
		return nil
	})
}

func runGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistrubutor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistrubutor)
	if err != nil {
		log.Fatal().Msg("cannot create new grpc server:")
	}

	rpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(rpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)

	// It allows to clients to explore what RPC are available on server and how to call them
	reflection.Register(grpcServer)

	listerner, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listerner:")
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server at %v", listerner.Addr())

		err = grpcServer.Serve(listerner)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}

			log.Error().Err(err).Msg("grpc server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown gRPC server")

		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is stopped")
		return nil
	})

}

func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistrubutor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistrubutor)
	if err != nil {
		log.Fatal().Msg("cannot create new grpc server:")
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

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal().Msg("cannot register  HandlerServer:")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// fs := http.FileServer(http.Dir("./doc/swagger/"))
	statikFs, err := fs.New()

	if err != nil {
		log.Fatal().Msg("cannot create  static fs:")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFs))
	mux.Handle("/swagger/", swaggerHandler)

	httpserver := &http.Server{
		Handler: gapi.HttpLogger(mux),
		Addr:    config.HTTPServerAddress,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %v", httpserver.Addr)
		err = httpserver.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			log.Error().Err(err).Msg("HTTP gateway server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown HTTP gateway server")

		err = httpserver.Shutdown(context.Background())
		if err != nil {
			log.Error().Msg("failed to shutdown HTTP gateway server")
			return err
		}
		log.Error().Err(err).Msg("HTTP gateway server is stopped")

		return nil
	})

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create new server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot connect to the database")
	}
}
