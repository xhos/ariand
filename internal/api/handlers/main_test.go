package handlers

import (
	"ariand/internal/api/middleware"
	"ariand/internal/config"
	"ariand/internal/db/postgres"
	"ariand/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var testDSN string

// TestMain sets up the test environment using a dockerized postgres database
func TestMain(m *testing.M) {
	setupLogger := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})

	pool, err := dockertest.NewPool("")
	if err != nil {
		setupLogger.Fatal("could not construct pool", "err", err)
	}
	if err := pool.Client.Ping(); err != nil {
		setupLogger.Fatal("could not connect to docker", "err", err)
	}

	setupLogger.Info("🐳 starting postgresql container")
	resource, err := pool.Run("postgres", "15-alpine", []string{
		"POSTGRES_USER=test",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})
	if err != nil {
		setupLogger.Fatal("could not start resource", "err", err)
	}
	setupLogger.Info("🐋 postgresql container is running")

	testDSN = fmt.Sprintf("postgres://test:secret@%s/testdb?sslmode=disable", resource.GetHostPort("5432/tcp"))

	if err := pool.Retry(func() error {
		db, err := sqlx.Open("pgx", testDSN)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		setupLogger.Fatal("could not connect to database", "err", err)
	}

	code := m.Run()

	setupLogger.Info("🐋 purging test container")
	if err := pool.Purge(resource); err != nil {
		setupLogger.Fatal("could not purge resource", "err", err)
	}
	setupLogger.Info("🐋 test container purged")

	os.Exit(code)
}

type testApp struct {
	server http.Handler
	db     *postgres.DB
	apiKey string
}

// do is a helper to execute http requests against the test server
func (app *testApp) do(t *testing.T, method, url string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := newAuthedRequest(t, method, url, app.apiKey, bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	app.server.ServeHTTP(rr, req)
	return rr
}

// colorfulLoggerAdapter adapts charm's logger to pgx's tracelog interface
type colorfulLoggerAdapter struct {
	logger *log.Logger
}

func (l *colorfulLoggerAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	var charmLevel log.Level
	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		charmLevel = log.DebugLevel
	case tracelog.LogLevelInfo:
		charmLevel = log.InfoLevel
	case tracelog.LogLevelWarn:
		charmLevel = log.WarnLevel
	case tracelog.LogLevelError:
		charmLevel = log.ErrorLevel
	default:
		charmLevel = log.DebugLevel
	}

	if sql, ok := data["sql"]; ok {
		data["query"] = sql
		delete(data, "sql")
	}

	keyvals := make([]any, 0, len(data)*2)
	for k, v := range data {
		keyvals = append(keyvals, k, v)
	}

	l.logger.Log(charmLevel, msg, keyvals...)
}

// newTestApp initializes a new application instance for testing
func newTestApp(t *testing.T) (*testApp, func()) {
	t.Helper()

	testLogger := log.NewWithOptions(os.Stderr, log.Options{
		Level: log.DebugLevel,
	})

	loggerAdapter := &colorfulLoggerAdapter{
		logger: testLogger.WithPrefix("🐘 pgx"),
	}

	tracer := &tracelog.TraceLog{
		Logger:   loggerAdapter,
		LogLevel: tracelog.LogLevelTrace,
	}

	pgxConfig, err := pgx.ParseConfig(testDSN)
	require.NoError(t, err, "failed to parse DSN")
	pgxConfig.Tracer = tracer

	dbConn := stdlib.OpenDB(*pgxConfig)
	sqlxDB := sqlx.NewDb(dbConn, "pgx")
	store, err := postgres.NewFromDB(sqlxDB)
	require.NoError(t, err, "failed to create store from DB")

	testCfg := &config.Config{
		ReceiptParserURL:     "http://localhost:9999",
		ReceiptParserTimeout: 5 * time.Second,
	}

	services := service.New(store, testLogger, testCfg)
	router := SetupRoutes(services)

	const testAPIKey = "test-api-key"
	stack := middleware.CreateStack(
		middleware.RequestID(),
		middleware.Logging(testLogger.WithPrefix("🌐 http")),
		middleware.CORS(),
		middleware.Auth(testLogger.WithPrefix("auth"), testAPIKey),
		middleware.RateLimit(),
	)

	app := &testApp{
		server: stack(router),
		db:     store,
		apiKey: testAPIKey,
	}

	cleanup := func() {
		store.Close()
	}

	return app, cleanup
}
