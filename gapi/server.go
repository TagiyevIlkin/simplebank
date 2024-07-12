package gapi

import (
	"fmt"

	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/token"
	"github.com/TagiyevIlkin/simplebank/util"
)

// Server serves all gRPC requests for our bankimg service
type Server struct {
	pb.UnimplementedSimpleBankServer
	store      db.Store
	config     util.Config
	tokenMaker token.Maker
}

// NewServer creates  new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
