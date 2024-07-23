package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	mockdb "github.com/TagiyevIlkin/simplebank/db/mock"
	db "github.com/TagiyevIlkin/simplebank/db/sqlc"
	"github.com/TagiyevIlkin/simplebank/pb"
	"github.com/TagiyevIlkin/simplebank/token"
	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestUpdateUserApi(t *testing.T) {
	user, _ := randomUser(t)
	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		// Happy case
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newEmail,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				// TODO Check Update User
				// arg := db.UpdateUserParams{
				// 	Username: user.Username,
				// 	FullName: sql.NullString{
				// 		String: newName,
				// 		Valid:  true,
				// 	},
				// 	Email: sql.NullString{
				// 		String: newEmail,
				// 		Valid:  true,
				// 	},
				// }

				updateUser := db.User{

					Username:        user.Username,
					FullName:        newName,
					Email:           newEmail,
					HashedPassword:  user.HashedPassword,
					PasswordChanged: user.PasswordChanged,
					CreatedAt:       user.CreatedAt,
					IsEmailVerified: user.IsEmailVerified,
				}
				store.EXPECT().
					// TODO Check Update User
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(updateUser, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {

				accessToken, _, err := tokenMaker.CreateToken(user.Username, user.Role, time.Minute)
				require.NoError(t, err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}

				return metadata.NewIncomingContext(context.Background(), md)

			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()

				require.Equal(t, user.Username, updatedUser.Username)
				require.Equal(t, newName, updatedUser.FullName)
				require.Equal(t, newEmail, updatedUser.Email)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)
			server := newTestServer(t, store, nil)

			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
