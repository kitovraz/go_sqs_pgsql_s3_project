package store_test

import (
	"context"
	"go_sqs_pqsql_s3_project/config/fixtures"
	"go_sqs_pqsql_s3_project/store"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	email := "testing@test.com"
	testingPassword := "testingpassword"

	testEnv := fixtures.NewTestEnv(t)
	teardownDb := testEnv.SetupDb(t)
	defer t.Cleanup(func() {
		teardownDb(t)
	})

	userStore := store.NewUserStore(testEnv.Db)
	user, err := userStore.CreateUser(ctx, email, testingPassword)
	require.NoError(t, err)
	require.Less(t, now, user.CreatedAt)

	require.Equal(t, email, user.Email)
	require.NoError(t, user.ComparePassword(testingPassword))

	user2, err := userStore.ById(ctx, user.Id)
	require.NoError(t, err)
	require.Equal(t, user.Email, user2.Email)
	require.Equal(t, user.Id, user2.Id)
	require.Equal(t, user.HashedPasswordBase64, user2.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt, user2.CreatedAt)

	user3, err := userStore.ByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.Equal(t, user.Email, user3.Email)
	require.Equal(t, user.Id, user3.Id)
	require.Equal(t, user.HashedPasswordBase64, user3.HashedPasswordBase64)
	require.Equal(t, user.CreatedAt, user3.CreatedAt)
}
