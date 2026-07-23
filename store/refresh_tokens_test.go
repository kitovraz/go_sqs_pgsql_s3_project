package store_test

import (
	"context"
	"go_sqs_pqsql_s3_project/config/apiserver"
	"go_sqs_pqsql_s3_project/config/fixtures"
	"go_sqs_pqsql_s3_project/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefreshTokenStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	ctx := context.Background()

	refreshTokenStore := store.NewRefreshTokenStore(env.Db)

	userStore := store.NewUserStore(env.Db)
	user, err := userStore.CreateUser(ctx, "TST.@TEST.RU", "test")
	require.NoError(t, err)

	jwtManager := apiserver.NewJwtManger(env.Config)
	tokenPair, err := jwtManager.GenerateTokenPair(user.Id)
	require.NoError(t, err)

	refreshToken, err := refreshTokenStore.Create(ctx, user.Id, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, user.Id, refreshToken.UserId)

	refreshTokenExpirationTime, err := tokenPair.RefreshToken.Claims.GetExpirationTime()
	require.NoError(t, err)
	require.Equal(t, refreshTokenExpirationTime.Time.UnixMilli(), refreshToken.ExpiresAt.UnixMilli())

	refreshToken2, err := refreshTokenStore.ByPrimaryKey(ctx, user.Id, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, refreshToken2.UserId, refreshToken.UserId)
	require.Equal(t, refreshToken2.HashedToken, refreshToken.HashedToken)
	require.Equal(t, refreshToken2.CreatedAt, refreshToken.CreatedAt)
	require.Equal(t, refreshToken2.ExpiresAt, refreshToken.ExpiresAt)

	result, err := refreshTokenStore.Delete(ctx, user.Id)
	require.NoError(t, err)
	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rowsAffected)
}
