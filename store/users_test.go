package store

import (
	"context"
	"fmt"
	"go_sqs_pqsql_s3_project/config"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	os.Setenv("ENV", string(config.Env_Test))
	cfg, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDb(cfg)
	require.NoError(t, err)
	defer db.Close()

	m, err := migrate.New(fmt.Sprintf("file://%s/db/migrations", cfg.ProjectRoot), cfg.DatabaseUrl())
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	email := "testing@test.com"
	testingPassword := "testingpassword"

	// Ensure a clean slate so re-runs don't collide on the unique email.
	_, err = db.Exec("DELETE FROM users WHERE email = $1", email)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM users WHERE email = $1", email)
	})

	userStore := NewUserStore(db)
	user, err := userStore.CreateUser(context.Background(), email, testingPassword)
	require.NoError(t, err)

	require.Equal(t, email, user.Email)
	require.NoError(t, user.ComparePassword(testingPassword))
}
