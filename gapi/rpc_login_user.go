package gapi

import (
	"context"
	"database/sql"

	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to find user")
	}

	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unauthorized user")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		req.GetUsername(),
		server.config.AccessTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create acces token user", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		req.GetUsername(),
		server.config.RefreshTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create refresh token user", err)
	}

	serssion, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    "", // TODO: fill from gRpc context metadata
		ClientIp:     "",
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create session", err)
	}

	rsp := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             serssion.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}

	return rsp, nil
}
