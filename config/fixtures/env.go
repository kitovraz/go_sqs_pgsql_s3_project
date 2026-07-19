package fixtures

import (
	"database/sql"
	"fmt"
	"go_sqs_pqsql_s3_project/config"
	"go_sqs_pqsql_s3_project/store"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	Config *config.Config
	Db     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	os.Setenv("ENV", string(config.Env_Test))
	cfg, err := config.New()
	require.NoError(t, err)

	db, err := store.NewPostgresDb(cfg)
	require.NoError(t, err)

	return &TestEnv{
		Config: cfg,
		Db:     db,
	}
}

func (te *TestEnv) SetupDb(t *testing.T) func(t *testing.T) {
	fmt.Println("Set up")
	m, err := migrate.New(fmt.Sprintf("file://%s/db/migrations", te.Config.ProjectRoot), te.Config.DatabaseUrl())
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	return te.TeardownDb
}

func (te *TestEnv) TeardownDb(t *testing.T) {
	fmt.Println("TeardownDb - Truncate \"users\", \"refresh_tokens\", \"reports\"")
	_, err := te.Db.Exec(fmt.Sprintf("truncate table %s", strings.Join([]string{"users", "refresh_tokens", "reports"}, ", ")))
	require.NoError(t, err)
}
