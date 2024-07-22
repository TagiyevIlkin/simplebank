package gapi

import (
	"context"
	"errors"

	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/TagiyevIlkin/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	if violations := validateLoginUserRequest(req); len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
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
		return nil, status.Errorf(codes.Internal, "Failed to create acces token user:%s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		req.GetUsername(),
		server.config.RefreshTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create refresh token user:%s", err)
	}

	mtdt := server.extrachMetaData(ctx)
	serssion, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIp,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create session:%s", err)
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

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateFullName(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return violations
}
