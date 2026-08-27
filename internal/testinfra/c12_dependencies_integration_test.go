//go:build integration

package testinfra

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

var c12DatabaseNamePattern = regexp.MustCompile(`^talenro_c12_([0-9a-f]{32})$`)

func TestC12DependenciesAreIsolatedAndBaseMigrated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	databaseURL, runSuffix := requireC12PostgresEndpoint(t, os.Getenv("TALENRO_DATABASE_URL"))
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal("open C12 PostgreSQL endpoint")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("ping C12 PostgreSQL endpoint")
	}
	var serverVersion string
	if err := pool.QueryRow(ctx, `SHOW server_version`).Scan(&serverVersion); err != nil {
		t.Fatal("read C12 PostgreSQL version")
	}
	if serverVersion != "18.4" {
		t.Fatalf("C12 PostgreSQL version = %q, want 18.4", serverVersion)
	}
	var migrationVersion int64
	var migrationApplied bool
	if err := pool.QueryRow(ctx, `SELECT version_id, is_applied FROM public.goose_db_version ORDER BY id DESC LIMIT 1`).Scan(&migrationVersion, &migrationApplied); err != nil {
		t.Fatal("read ordinary Goose migration high-water")
	}
	if migrationVersion != 6 || !migrationApplied {
		t.Fatalf("ordinary Goose migration high-water = (%d,%t), want (6,true)", migrationVersion, migrationApplied)
	}
	var tableCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='nodecontrol' AND table_type='BASE TABLE'`).Scan(&tableCount); err != nil {
		t.Fatal("count base nodecontrol catalog")
	}
	if tableCount != 25 {
		t.Fatalf("base nodecontrol table count = %d, want 25", tableCount)
	}

	redisAddress := requireC12HostPort(t, "TALENRO_REDIS_ADDRESS", os.Getenv("TALENRO_REDIS_ADDRESS"))
	rdb := redis.NewClient(&redis.Options{Addr: redisAddress})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatal("ping C12 Redis endpoint")
	}
	if size, err := rdb.DBSize(ctx).Result(); err != nil {
		t.Fatal("inspect C12 Redis isolation")
	} else if size != 0 {
		t.Fatalf("C12 Redis database is not isolated: initial key count = %d", size)
	}
	markerKey := "talenro:c12:run:" + runSuffix
	created, err := rdb.SetNX(ctx, markerKey, runSuffix, 5*time.Minute).Result()
	if err != nil || !created {
		t.Fatal("write unique C12 Redis run marker")
	}
	if marker, err := rdb.Get(ctx, markerKey).Result(); err != nil || marker != runSuffix {
		t.Fatal("read exact C12 Redis run marker")
	}

	natsURL := requireC12NATSEndpoint(t, os.Getenv("TALENRO_NATS_URL"))
	nc, err := nats.Connect(natsURL, nats.Timeout(3*time.Second), nats.Name("talenro-c12-dependency-gate"))
	if err != nil {
		t.Fatal("connect C12 NATS endpoint")
	}
	defer nc.Close()
	wantServerName := "talenro-c12-" + runSuffix + "-nats"
	if got := nc.ConnectedServerName(); got != wantServerName {
		t.Fatalf("C12 NATS run marker = %q, want %q", got, wantServerName)
	}
	jetStream, err := nc.JetStream()
	if err != nil {
		t.Fatal("create C12 JetStream context")
	}
	if _, err := jetStream.AccountInfo(nats.Context(ctx)); err != nil {
		t.Fatal("query C12 JetStream account")
	}
}

func requireC12PostgresEndpoint(t *testing.T, raw string) (string, string) {
	t.Helper()
	if raw == "" {
		t.Fatal("TALENRO_DATABASE_URL is missing")
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Hostname() != "127.0.0.1" {
		t.Fatal("TALENRO_DATABASE_URL is not an isolated loopback PostgreSQL endpoint")
	}
	if parsed.User == nil || parsed.User.Username() != "talenro" {
		t.Fatal("TALENRO_DATABASE_URL has the wrong run-owned user")
	}
	if _, present := parsed.User.Password(); !present {
		t.Fatal("TALENRO_DATABASE_URL lacks its run-owned credential")
	}
	if err := requireC12Port(parsed.Port()); err != nil {
		t.Fatal("TALENRO_DATABASE_URL has an invalid mapped port")
	}
	databaseName := strings.TrimPrefix(parsed.EscapedPath(), "/")
	match := c12DatabaseNamePattern.FindStringSubmatch(databaseName)
	if match == nil {
		t.Fatal("TALENRO_DATABASE_URL lacks the cryptographic C12 run marker")
	}
	return raw, match[1]
}

func requireC12HostPort(t *testing.T, variable, raw string) string {
	t.Helper()
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host != "127.0.0.1" || requireC12Port(port) != nil {
		t.Fatalf("%s is not an isolated loopback endpoint", variable)
	}
	return raw
}

func requireC12NATSEndpoint(t *testing.T, raw string) string {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "nats" || parsed.Hostname() != "127.0.0.1" || requireC12Port(parsed.Port()) != nil || parsed.User != nil || parsed.Path != "" {
		t.Fatal("TALENRO_NATS_URL is not an isolated loopback NATS endpoint")
	}
	return raw
}

func requireC12Port(raw string) error {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port")
	}
	return nil
}
