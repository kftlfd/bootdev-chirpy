package testutils

import (
	"chirpy/internal/app"
	"chirpy/internal/config"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func NewTestApp(t *testing.T) http.Handler {
	logger := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	))

	return app.BuildApp(
		Config(),
		SetupTestDB(t),
		logger,
	)
}

func Config() *config.Config {
	return &config.Config{
		Env: config.EnvVars{
			ServerSecret: "test-secret",
			PolkaKey:     "test-key",
		},
		IsDev: true,
	}
}

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Fatal("missing env: TEST_DB_URL")
	}

	db, err := sql.Open("postgres", dbURL)

	require.NoError(t, err)

	resetDB(t, db)

	t.Cleanup(func() {
		resetDB(t, db)
		db.Close()
	})

	return db
}

func resetDB(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		TRUNCATE TABLE users
		CASCADE;
	`)

	require.NoError(t, err)
}
