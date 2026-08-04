# Talenro Project Foundation and Shared Contracts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立一个可重复构建、可测试、可观测的 Talenro Go 控制面单仓库基线，包含 PostgreSQL、Redis、NATS JetStream、本地开发环境、Protobuf/OpenAPI 共享契约及自动验证入口。

**Architecture:** 首个交付只搭建控制面基础和共享契约，不实现节点调度、计量支付或客户端隧道。PostgreSQL 是强一致事实来源；Redis 只承载可重建的短期状态；NATS JetStream 承载带幂等标识的领域事件。HTTP API 从仓库内 OpenAPI 生成 Go 接口，跨服务事件从仓库内 Protobuf 生成 Go 类型。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、Redis 8.8.1、NATS Server 2.14.3、NATS Go 1.52.0、pgx 5.10.0、go-redis 9.22.0、Buf 1.72.0、Protobuf Go 1.36.11、oapi-codegen 2.8.0、sqlc 1.31.1、goose 3.27.1、golangci-lint 2.12.2、Prometheus Go client 1.23.2、Docker Compose。

**Specifications:** [VPN architecture](../specs/2026-08-04-resilient-vpn-architecture-design.md)、[commercial client](../specs/2026-08-04-vpn-commercial-client-design.md)、[localization](../specs/2026-08-04-vpn-localization-design.md)、[brand naming](../specs/2026-08-04-talenro-brand-naming-design.md)。

## Global Constraints

- 产品主品牌固定为 `Talenro`，官方中文名为 `泰联诺`；消费者商店名为 `Talenro VPN`。
- MVP 目标为 10 万并发在线设备、5 个 POP；单设备产品带宽上限为 100 Mbps，但本计划不实现流量整形。
- 活跃连接故障恢复目标为 p95 小于 8 秒、p99 小于 10 秒；该指标由后续客户端与调度计划实现，本计划只保留无敏感信息的观测基础。
- Android、iOS、Windows、macOS 四端完全原生；共享的是契约和测试向量，不共享跨平台 UI 运行时代码。
- 客户端隧道使用 sing-box，服务端使用 Xray/sing-box 混合架构；本计划不得把两者链接进专有控制面二进制。
- sing-box 为 GPLv3+，Xray-core 为 MPL-2.0；许可证、完整对应源码和应用商店兼容性必须保留独立法律发布门。
- 不收集访问域名、目标 IP、DNS 内容或应用流量内容；日志、指标、追踪和测试夹具均不得包含这些字段。
- PostgreSQL 是设备、节点、凭证版本、最高配置版本及后续账本的事实来源；Redis 数据必须可丢弃并从事实来源重建。
- NATS JetStream 事件按“至少一次”交付设计；每个消费者必须用 `event_id` 或业务幂等键去重，不能假设恰好一次。
- 禁止使用浮动的 `latest` 容器标签；本计划固定 `postgres:18.4-alpine3.23`、`redis:8.8.1-alpine3.23`、`nats:2.14.3-alpine3.22`。
- Redis 8 使用三选一许可证；本地开发选择未修改官方镜像并按 AGPLv3 路径记录，生产上线前必须取得书面法律结论或商业授权，不得静默修改 Redis 服务端。
- Go 内部模块路径暂用 `talenro.local/platform`，仅供私有开发；取得正式代码托管命名空间或域名后，在首次公开模块发布前以单独迁移提交统一更换。
- 移动端 package ID、Apple bundle ID、Windows package identity 和 `com.talenro.*` 在商标与域名权属确认前不得冻结。
- 所有生成物由本地固定版本工具生成并提交；CI 必须验证重新生成后工作区无差异。

---

## Scope and delivery sequence

本计划是以下实施序列的第 1 个独立交付：

1. 项目基础、数据依赖、共享契约与健康服务（本计划）。
2. 设备注册、配置签名/HPKE、候选集调度和配置分发。
3. 节点代理、POP 期望状态、健康聚合与容量调度。
4. Tunnel Engine 状态机、自动/手动切换及四端平台适配。
5. 账号、设备、权益、流量账本、配额租约与硬停。
6. 商品目录、Apple/Google/Web/加密支付与退款生命周期。
7. 邀请返利、推广佣金和激励广告奖励。
8. 四端商业 UI、五语本地化、RTL、发布与商店合规。

本计划不创建业务表，不接入 sing-box/Xray，不实现认证、支付、计量、调度、加密配置或客户端 UI。结束时应得到一个能启动依赖、应用迁移、生成契约、运行健康 API、暴露基础指标并通过一条验证命令验收的仓库。

## Repository map

```text
.
├── api/
│   ├── openapi/control-api.v1.yaml       # 控制面 HTTP 契约源
│   └── proto/talenro/events/v1/          # 跨服务事件契约源
├── cmd/control-api/main.go               # 基础控制面进程入口
├── db/
│   ├── migrations/                       # goose 顺序 SQL 迁移
│   └── queries/                          # sqlc 查询源
├── deploy/dev/
│   ├── compose.yaml                      # 固定版本的本地依赖
│   └── nats.conf                         # JetStream 开发配置
├── gen/go/                               # 已提交的生成 Go 代码
├── internal/
│   ├── buildinfo/                        # 构建版本信息
│   ├── config/                           # 环境配置解析与校验
│   ├── contracts/events/                 # 事件信封验证
│   ├── controlapi/                       # OpenAPI 接口实现
│   ├── observability/                    # slog 与 Prometheus
│   ├── platform/                         # PostgreSQL/Redis/NATS 生命周期
│   ├── readiness/                        # 有界并发依赖检查
│   └── store/                            # sqlc 生成的数据访问代码
├── scripts/                              # Windows/Unix 验证和冒烟入口
├── docs/decisions/                       # 架构决策记录
├── docs/runbooks/                        # 本地开发与故障排查
├── buf.gen.yaml
├── buf.yaml
├── go.mod
├── go.sum
└── sqlc.yaml
```

### Task 1: Repository baseline and build identity

**Files:**
- Create: `.editorconfig`
- Create: `.gitattributes`
- Create: `.gitignore`
- Create: `.go-version`
- Create: `go.mod`
- Create: `README.md`
- Create: `internal/buildinfo/buildinfo_test.go`
- Create: `internal/buildinfo/buildinfo.go`

**Interfaces:**
- Consumes: 已确认的品牌和架构规格。
- Produces: `buildinfo.Info`、`buildinfo.Current() Info`，供所有可执行文件输出不含敏感信息的构建标识。

- [ ] **Step 1: Create repository metadata and the Go module**

`go.mod`：

```go
module talenro.local/platform

go 1.26.0

toolchain go1.26.5
```

`.go-version`：

```text
1.26.5
```

`.gitattributes`：

```gitattributes
* text=auto
*.go text eol=lf
*.proto text eol=lf
*.sql text eol=lf
*.yaml text eol=lf
*.yml text eol=lf
*.md text eol=lf
*.ps1 text eol=crlf
*.sh text eol=lf
```

`.gitignore`：

```gitignore
.env
.idea/
.vscode/
.DS_Store
Thumbs.db
bin/
coverage/
dist/
*.out
*.test
```

`.editorconfig`：

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab

[*.{yaml,yml,json,md,proto,sql}]
indent_style = space
indent_size = 2

[*.ps1]
end_of_line = crlf
indent_style = space
indent_size = 2
```

- [ ] **Step 2: Write the failing build-info test**

```go
package buildinfo

import "testing"

func TestCurrentDefaultsAreStable(t *testing.T) {
	got := Current()
	if got.Version != "dev" || got.Commit != "unknown" || got.BuiltAt != "unknown" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}
```

- [ ] **Step 3: Run the test and verify it fails**

Run: `go test ./internal/buildinfo`

Expected: compilation fails because `Current` is undefined.

- [ ] **Step 4: Implement the minimal build-info package**

```go
package buildinfo

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"built_at"`
}

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

func Current() Info {
	return Info{Version: version, Commit: commit, BuiltAt: builtAt}
}
```

- [ ] **Step 5: Document the repository purpose**

Create `README.md` with the exact product name `Talenro`, links to `docs/superpowers/specs/`, the internal-only status of `talenro.local/platform`, and these initial commands:

```powershell
go version
go test ./...
```

- [ ] **Step 6: Verify and commit**

Run: `gofmt -w internal/buildinfo/*.go`

Run: `go test ./...`

Expected: PASS.

```bash
git add .editorconfig .gitattributes .gitignore .go-version go.mod README.md internal/buildinfo
git commit -m "build: establish Go repository baseline"
```

### Task 2: Pinned developer toolchain

**Files:**
- Modify: `go.mod`
- Create: `go.sum`
- Create: `scripts/check-tools.ps1`
- Create: `scripts/check-tools.sh`

**Interfaces:**
- Consumes: Go 1.26 tool directives.
- Produces: fixed local commands `go tool buf`, `go tool protoc-gen-go`, `go tool oapi-codegen`, `go tool sqlc`, `go tool goose`, and `go tool golangci-lint`.

- [ ] **Step 1: Pin every generator and validator with Go tool directives**

Run exactly:

```powershell
go get -tool github.com/bufbuild/buf/cmd/buf@v1.72.0
go get -tool google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
go get -tool github.com/pressly/goose/v3/cmd/goose@v3.27.1
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go mod tidy
```

- [ ] **Step 2: Add the Windows tool verification script**

```powershell
$ErrorActionPreference = 'Stop'

$expected = @{
  buf = '1.72.0'
  protoc = 'protoc-gen-go v1.36.11'
  oapi = 'v2.8.0'
  sqlc = 'v1.31.1'
  goose = 'v3.27.1'
  lint = '2.12.2'
}

$actual = @{
  buf = (go tool buf --version)
  protoc = (go tool protoc-gen-go --version)
  oapi = (go tool oapi-codegen --version)
  sqlc = (go tool sqlc version)
  goose = (go tool goose -version)
  lint = (go tool golangci-lint version)
}

foreach ($name in $expected.Keys) {
  if ($actual[$name] -notmatch [regex]::Escape($expected[$name])) {
    throw "$name version mismatch: $($actual[$name])"
  }
}
```

- [ ] **Step 3: Add the Unix tool verification script with the same checks**

```bash
#!/usr/bin/env bash
set -euo pipefail

go tool buf --version | grep -F '1.72.0'
go tool protoc-gen-go --version | grep -F 'v1.36.11'
go tool oapi-codegen --version | grep -F 'v2.8.0'
go tool sqlc version | grep -F 'v1.31.1'
go tool goose -version | grep -F 'v3.27.1'
go tool golangci-lint version | grep -F '2.12.2'
```

- [ ] **Step 4: Run both applicable checks and commit**

Run on Windows: `powershell -ExecutionPolicy Bypass -File scripts/check-tools.ps1`

Run on Unix CI: `bash scripts/check-tools.sh`

Expected: every pinned version is printed and the script exits 0.

```bash
git add go.mod go.sum scripts/check-tools.ps1 scripts/check-tools.sh
git commit -m "build: pin code generation toolchain"
```

### Task 3: Typed runtime configuration

**Files:**
- Create: `.env.example`
- Create: `internal/config/config_test.go`
- Create: `internal/config/config.go`

**Interfaces:**
- Consumes: environment lookup function `func(string) (string, bool)`.
- Produces: `config.Load(lookup) (Config, error)` and immutable `config.Config` fields used by process wiring.

- [ ] **Step 1: Write configuration tests**

```go
package config

import (
	"strings"
	"testing"
	"time"
)

func lookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := Load(lookup(map[string]string{}))
	if err == nil || !strings.Contains(err.Error(), "TALENRO_DATABASE_URL") {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestLoadAppliesSafeDefaults(t *testing.T) {
	got, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://talenro:talenro_dev@localhost:5432/talenro?sslmode=disable",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.HTTPAddress != "127.0.0.1:8080" || got.RedisAddress != "127.0.0.1:6379" || got.NATSURL != "nats://127.0.0.1:4222" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	if got.DependencyTimeout != 2*time.Second || got.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected timeouts: %+v", got)
	}
}

func TestLoadRejectsPublicHTTPWithoutExplicitOptIn(t *testing.T) {
	_, err := Load(lookup(map[string]string{
		"TALENRO_DATABASE_URL": "postgres://local",
		"TALENRO_HTTP_ADDRESS": "0.0.0.0:8080",
	}))
	if err == nil {
		t.Fatal("expected public bind rejection")
	}
}
```

- [ ] **Step 2: Verify the tests fail**

Run: `go test ./internal/config`

Expected: compilation fails because `Load` is undefined.

- [ ] **Step 3: Implement strict parsing**

```go
package config

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddress      string
	DatabaseURL     string
	RedisAddress    string
	NATSURL          string
	DependencyTimeout time.Duration
	ShutdownTimeout time.Duration
}

type Lookup func(string) (string, bool)

func Load(lookup Lookup) (Config, error) {
	cfg := Config{
		HTTPAddress: "127.0.0.1:8080",
		RedisAddress: "127.0.0.1:6379",
		NATSURL: "nats://127.0.0.1:4222",
		DependencyTimeout: 2 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	var ok bool
	if cfg.DatabaseURL, ok = lookup("TALENRO_DATABASE_URL"); !ok || cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("TALENRO_DATABASE_URL is required")
	}
	if value, exists := lookup("TALENRO_HTTP_ADDRESS"); exists && value != "" {
		cfg.HTTPAddress = value
	}
	if value, exists := lookup("TALENRO_REDIS_ADDRESS"); exists && value != "" {
		cfg.RedisAddress = value
	}
	if value, exists := lookup("TALENRO_NATS_URL"); exists && value != "" {
		cfg.NATSURL = value
	}
	allowPublic := false
	if value, exists := lookup("TALENRO_ALLOW_PUBLIC_HTTP"); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse TALENRO_ALLOW_PUBLIC_HTTP: %w", err)
		}
		allowPublic = parsed
	}
	host, _, err := net.SplitHostPort(cfg.HTTPAddress)
	if err != nil { return Config{}, fmt.Errorf("parse TALENRO_HTTP_ADDRESS: %w", err) }
	ip := net.ParseIP(host)
	publicBind := host == "" || (ip != nil && ip.IsUnspecified())
	if publicBind && !allowPublic {
		return Config{}, fmt.Errorf("public HTTP bind requires TALENRO_ALLOW_PUBLIC_HTTP=true")
	}
	return cfg, nil
}
```

Do not implement reflection-based automatic environment loading and do not expose a method that prints the complete configuration.

- [ ] **Step 4: Add the non-secret example environment file**

```dotenv
TALENRO_HTTP_ADDRESS=127.0.0.1:8080
TALENRO_DATABASE_URL=postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable
TALENRO_REDIS_ADDRESS=127.0.0.1:6379
TALENRO_NATS_URL=nats://127.0.0.1:4222
TALENRO_ALLOW_PUBLIC_HTTP=false
```

- [ ] **Step 5: Verify and commit**

Run: `gofmt -w internal/config/*.go`

Run: `go test ./internal/config`

Expected: PASS.

```bash
git add .env.example internal/config
git commit -m "feat: add typed runtime configuration"
```

### Task 4: Reproducible local data dependencies

**Files:**
- Create: `deploy/dev/compose.yaml`
- Create: `deploy/dev/nats.conf`
- Create: `internal/testinfra/dependencies_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Consumes: `.env.example` endpoint values.
- Produces: PostgreSQL on `5432`, Redis on `6379`, NATS on `4222`, NATS monitoring on `8222`, and integration test tag `integration`.

- [ ] **Step 1: Add exact runtime client dependencies**

Run:

```powershell
go get github.com/jackc/pgx/v5@v5.10.0
go get github.com/redis/go-redis/v9@v9.22.0
go get github.com/nats-io/nats.go@v1.52.0
go mod tidy
```

- [ ] **Step 2: Define the NATS development server**

`deploy/dev/nats.conf`：

```conf
server_name: talenro-dev
http: 8222

jetstream {
  store_dir: "/data/jetstream"
  max_mem_store: 256MB
  max_file_store: 1GB
}
```

- [ ] **Step 3: Define fixed-version Docker services**

`deploy/dev/compose.yaml`：

```yaml
name: talenro-dev

services:
  postgres:
    image: postgres:18.4-alpine3.23
    environment:
      POSTGRES_DB: talenro
      POSTGRES_USER: talenro
      POSTGRES_PASSWORD: talenro_dev
    ports:
      - "127.0.0.1:5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U talenro -d talenro"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - postgres-data:/var/lib/postgresql/18/docker

  redis:
    image: redis:8.8.1-alpine3.23
    command: ["redis-server", "--appendonly", "yes", "--save", "60", "1"]
    ports:
      - "127.0.0.1:6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - redis-data:/data

  nats:
    image: nats:2.14.3-alpine3.22
    command: ["-c", "/etc/nats/nats.conf"]
    ports:
      - "127.0.0.1:4222:4222"
      - "127.0.0.1:8222:8222"
    healthcheck:
      test: ["CMD", "wget", "-q", "-O", "-", "http://127.0.0.1:8222/healthz?js-enabled-only=true"]
      interval: 2s
      timeout: 2s
      retries: 20
    volumes:
      - ./nats.conf:/etc/nats/nats.conf:ro
      - nats-data:/data

volumes:
  postgres-data:
  redis-data:
  nats-data:
```

- [ ] **Step 4: Write the dependency integration test**

```go
//go:build integration

package testinfra

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func TestDependenciesAreReachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, os.Getenv("TALENRO_DATABASE_URL"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if err := db.Ping(ctx); err != nil { t.Fatal(err) }

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil { t.Fatal(err) }

	nc, err := nats.Connect("nats://127.0.0.1:4222", nats.Timeout(2*time.Second))
	if err != nil { t.Fatal(err) }
	defer nc.Close()
	if _, err := nc.JetStream(); err != nil { t.Fatal(err) }
}
```

- [ ] **Step 5: Start, test, stop, and commit**

Run: `docker compose -f deploy/dev/compose.yaml config --quiet`

Run: `docker compose -f deploy/dev/compose.yaml up -d --wait`

Run:

```powershell
$env:TALENRO_DATABASE_URL='postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
go test -tags=integration ./internal/testinfra -v
```

Expected: PASS.

Run: `docker compose -f deploy/dev/compose.yaml down`

```bash
git add deploy/dev internal/testinfra go.mod go.sum
git commit -m "build: add local PostgreSQL Redis and NATS stack"
```

### Task 5: Migration and typed PostgreSQL access path

**Files:**
- Create: `db/migrations/00001_system_metadata.sql`
- Create: `db/queries/system_metadata.sql`
- Create: `sqlc.yaml`
- Create: `internal/store/system_metadata_integration_test.go`
- Generate: `internal/store/db.go`
- Generate: `internal/store/models.go`
- Generate: `internal/store/querier.go`
- Generate: `internal/store/system_metadata.sql.go`

**Interfaces:**
- Consumes: PostgreSQL connection from Task 4.
- Produces: `store.New(DBTX) *Queries`, `PutSystemMetadata(ctx, PutSystemMetadataParams) error`, and `GetSystemMetadata(ctx, string) (json.RawMessage, error)`.

- [ ] **Step 1: Write the first reversible migration**

```sql
-- +goose Up
CREATE TABLE system_metadata (
  key text PRIMARY KEY,
  value jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT system_metadata_key_format CHECK (key ~ '^[a-z][a-z0-9_.-]{0,127}$')
);

-- +goose Down
DROP TABLE system_metadata;
```

- [ ] **Step 2: Define typed queries**

```sql
-- name: PutSystemMetadata :exec
INSERT INTO system_metadata (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = now();

-- name: GetSystemMetadata :one
SELECT value
FROM system_metadata
WHERE key = $1;
```

- [ ] **Step 3: Configure sqlc**

```yaml
version: "2"
sql:
  - engine: postgresql
    schema: db/migrations
    queries: db/queries
    gen:
      go:
        package: store
        out: internal/store
        sql_package: pgx/v5
        emit_interface: true
        emit_json_tags: true
        emit_empty_slices: true
```

- [ ] **Step 4: Generate the store and write the integration test**

Run: `go tool sqlc generate`

```go
//go:build integration

package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSystemMetadataRoundTrip(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("TALENRO_DATABASE_URL"))
	if err != nil { t.Fatal(err) }
	defer pool.Close()

	queries := New(pool)
	want := []byte(`{"version":1}`)
	if err := queries.PutSystemMetadata(ctx, PutSystemMetadataParams{Key: "contracts.version", Value: want}); err != nil {
		t.Fatal(err)
	}
	got, err := queries.GetSystemMetadata(ctx, "contracts.version")
	if err != nil { t.Fatal(err) }
	if string(got) != string(want) { t.Fatalf("got %s want %s", got, want) }
}
```

- [ ] **Step 5: Apply, test, rollback-test, and reapply**

Run:

```powershell
docker compose -f deploy/dev/compose.yaml up -d --wait
$env:TALENRO_DATABASE_URL='postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
go test -tags=integration ./internal/store -v
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL down
go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
```

Expected: round trip passes; down and second up both succeed.

- [ ] **Step 6: Verify generated code is reproducible and commit**

Run: `go tool sqlc generate`

Run: `git diff --exit-code -- internal/store`

Expected: no diff.

```bash
git add db sqlc.yaml internal/store
git commit -m "feat: add migration and typed PostgreSQL path"
```

### Task 6: Versioned Protobuf event envelope

**Files:**
- Create: `buf.yaml`
- Create: `buf.gen.yaml`
- Create: `api/proto/talenro/events/v1/envelope.proto`
- Generate: `gen/go/talenro/events/v1/envelope.pb.go`
- Create: `internal/contracts/events/validate_test.go`
- Create: `internal/contracts/events/validate.go`

**Interfaces:**
- Consumes: Buf and protoc-gen-go pinned in Task 2.
- Produces: `eventsv1.EventEnvelope` and `events.ValidateEnvelope(*eventsv1.EventEnvelope) error`.

- [ ] **Step 1: Configure Buf lint, breaking checks, and generation**

`buf.yaml`：

```yaml
version: v2
modules:
  - path: api/proto
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

`buf.gen.yaml`：

```yaml
version: v2
clean: true
plugins:
  - local: ["go", "tool", "protoc-gen-go"]
    out: gen/go
    opt:
      - paths=source_relative
```

- [ ] **Step 2: Define the stable event envelope**

```proto
syntax = "proto3";

package talenro.events.v1;

import "google/protobuf/timestamp.proto";

option go_package = "talenro.local/platform/gen/go/talenro/events/v1;eventsv1";

message EventEnvelope {
  string event_id = 1;
  string event_type = 2;
  google.protobuf.Timestamp occurred_at = 3;
  string producer = 4;
  string aggregate_type = 5;
  string aggregate_id = 6;
  uint64 aggregate_version = 7;
  string idempotency_key = 8;
  string trace_parent = 9;
  bytes payload = 10;
}
```

Reserve removed field numbers and names in every future change; never recycle them.

- [ ] **Step 3: Generate code and write failing semantic tests**

Run: `go tool buf lint`

Run: `go tool buf generate`

```go
package events

import (
	"testing"

	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateEnvelopeRejectsMissingIdempotencyKey(t *testing.T) {
	e := &eventsv1.EventEnvelope{
		EventId: "01JTESTEVENT000000000000001",
		EventType: "talenro.system.ready.v1",
		OccurredAt: timestamppb.Now(),
		Producer: "control-api",
		AggregateType: "system",
		AggregateId: "foundation",
		AggregateVersion: 1,
		Payload: []byte(`{"ready":true}`),
	}
	if err := ValidateEnvelope(e); err == nil {
		t.Fatal("expected missing idempotency key error")
	}
}
```

Run: `go test ./internal/contracts/events`

Expected: compilation fails because `ValidateEnvelope` is undefined.

- [ ] **Step 4: Implement envelope validation**

```go
package events

import (
	"fmt"
	"strings"

	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
)

func ValidateEnvelope(e *eventsv1.EventEnvelope) error {
	if e == nil { return fmt.Errorf("event envelope is nil") }
	if e.EventId == "" { return fmt.Errorf("event_id is required") }
	if !strings.HasPrefix(e.EventType, "talenro.") { return fmt.Errorf("event_type must use talenro namespace") }
	if e.OccurredAt == nil || !e.OccurredAt.IsValid() { return fmt.Errorf("occurred_at is invalid") }
	if e.Producer == "" { return fmt.Errorf("producer is required") }
	if e.AggregateType == "" || e.AggregateId == "" || e.AggregateVersion == 0 { return fmt.Errorf("aggregate identity is incomplete") }
	if e.IdempotencyKey == "" { return fmt.Errorf("idempotency_key is required") }
	if len(e.Payload) == 0 { return fmt.Errorf("payload is required") }
	if len(e.Payload) > 256*1024 { return fmt.Errorf("payload exceeds 256 KiB") }
	return nil
}
```

- [ ] **Step 5: Add a valid round-trip test**

Add this test without email, device identifiers, target addresses, domains, credentials or traffic content in the fixture:

```go
func TestEnvelopeRoundTrip(t *testing.T) {
	want := &eventsv1.EventEnvelope{
		EventId: "01JTESTEVENT000000000000002",
		EventType: "talenro.system.ready.v1",
		OccurredAt: timestamppb.Now(),
		Producer: "control-api",
		AggregateType: "system",
		AggregateId: "foundation",
		AggregateVersion: 1,
		IdempotencyKey: "system:foundation:1",
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		Payload: []byte(`{"ready":true}`),
	}
	encoded, err := proto.Marshal(want)
	if err != nil { t.Fatal(err) }
	got := new(eventsv1.EventEnvelope)
	if err := proto.Unmarshal(encoded, got); err != nil { t.Fatal(err) }
	if err := ValidateEnvelope(got); err != nil { t.Fatal(err) }
	if !proto.Equal(want, got) { t.Fatalf("round trip mismatch: got %v want %v", got, want) }
}
```

Add `google.golang.org/protobuf/proto` to the test imports.

- [ ] **Step 6: Verify generation and commit**

Run: `gofmt -w internal/contracts/events/*.go`

Run: `go tool buf lint`

Run: `go tool buf generate`

Run: `go test ./internal/contracts/events`

Expected: PASS.

```bash
git add buf.yaml buf.gen.yaml api/proto gen/go internal/contracts/events go.mod go.sum
git commit -m "feat: define versioned event envelope"
```

### Task 7: OpenAPI control-plane health contract

**Files:**
- Create: `api/openapi/control-api.v1.yaml`
- Create: `api/openapi/oapi-codegen.yaml`
- Generate: `gen/go/talenro/controlapi/v1/server.gen.go`
- Create: `internal/controlapi/contract_test.go`

**Interfaces:**
- Consumes: oapi-codegen pinned in Task 2.
- Produces: `controlapiv1.ServerInterface`, `controlapiv1.HandlerFromMux`, and `controlapiv1.HealthResponse`.

- [ ] **Step 1: Define the minimal OpenAPI contract**

```yaml
openapi: 3.0.3
info:
  title: Talenro Control API
  version: 1.0.0
paths:
  /livez:
    get:
      operationId: getLiveness
      responses:
        "200":
          description: Process is alive
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/HealthResponse"
  /readyz:
    get:
      operationId: getReadiness
      responses:
        "200":
          description: Required dependencies are ready
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/HealthResponse"
        "503":
          description: At least one required dependency is unavailable
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/HealthResponse"
components:
  schemas:
    HealthResponse:
      type: object
      additionalProperties: false
      required: [status, checks]
      properties:
        status:
          type: string
          enum: [ok, unavailable]
        checks:
          type: object
          additionalProperties:
            type: string
```

- [ ] **Step 2: Configure standard-library server generation**

```yaml
package: controlapiv1
output: gen/go/talenro/controlapi/v1/server.gen.go
generate:
  models: true
  std-http-server: true
  embedded-spec: true
output-options:
  skip-prune: false
```

- [ ] **Step 3: Generate and test the embedded contract**

Run:

```powershell
go tool oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/control-api.v1.yaml
```

Create a test that calls `controlapiv1.GetSwagger()`, asserts no error, and verifies `/livez` and `/readyz` exist with GET operations.

```go
package controlapi

import (
	"testing"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
)

func TestEmbeddedContractHasHealthRoutes(t *testing.T) {
	spec, err := controlapiv1.GetSwagger()
	if err != nil { t.Fatal(err) }
	for _, path := range []string{"/livez", "/readyz"} {
		item := spec.Paths.Find(path)
		if item == nil || item.Get == nil {
			t.Fatalf("missing GET %s", path)
		}
	}
}
```

- [ ] **Step 4: Verify and commit**

Run: `go test ./internal/controlapi`

Run the generator again, then: `git diff --exit-code -- gen/go/talenro/controlapi/v1/server.gen.go`

Expected: PASS and no diff.

```bash
git add api/openapi gen/go/talenro/controlapi internal/controlapi
git commit -m "feat: define control API health contract"
```

### Task 8: Bounded readiness aggregation

**Files:**
- Create: `internal/readiness/checker_test.go`
- Create: `internal/readiness/checker.go`
- Create: `internal/platform/postgres/probe.go`
- Create: `internal/platform/redis/probe.go`
- Create: `internal/platform/nats/probe.go`

**Interfaces:**
- Consumes: `Ping(context.Context) error` capabilities from pgxpool, go-redis and NATS.
- Produces: `readiness.Probe`, `readiness.New(timeout, probes...) *Checker`, and `(*Checker).Check(context.Context) (bool, map[string]string)`.

- [ ] **Step 1: Write readiness behavior tests**

```go
package readiness

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProbe struct { name string; err error }
func (p fakeProbe) Name() string { return p.name }
func (p fakeProbe) Ping(context.Context) error { return p.err }

func TestCheckReportsEveryProbeWithoutLeakingErrors(t *testing.T) {
	checker := New(50*time.Millisecond,
		fakeProbe{name: "postgres"},
		fakeProbe{name: "redis", err: errors.New("redis://user:secret@example")},
	)
	ready, checks := checker.Check(context.Background())
	if ready { t.Fatal("expected unavailable") }
	if checks["postgres"] != "ok" || checks["redis"] != "unavailable" {
		t.Fatalf("unexpected checks: %#v", checks)
	}
}
```

- [ ] **Step 2: Verify the test fails**

Run: `go test ./internal/readiness`

Expected: compilation fails because `New` is undefined.

- [ ] **Step 3: Implement bounded concurrent checks**

```go
package readiness

import (
	"context"
	"sync"
	"time"
)

type Probe interface {
	Name() string
	Ping(context.Context) error
}

type Checker struct {
	timeout time.Duration
	probes []Probe
}

func New(timeout time.Duration, probes ...Probe) *Checker {
	return &Checker{timeout: timeout, probes: append([]Probe(nil), probes...)}
}

func (c *Checker) Check(ctx context.Context) (bool, map[string]string) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	checks := make(map[string]string, len(c.probes))
	ready := true
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, probe := range c.probes {
		probe := probe
		wg.Add(1)
		go func() {
			defer wg.Done()
			status := "ok"
			if err := probe.Ping(ctx); err != nil { status = "unavailable" }
			mu.Lock()
			checks[probe.Name()] = status
			if status != "ok" { ready = false }
			mu.Unlock()
		}()
	}
	wg.Wait()
	return ready, checks
}
```

- [ ] **Step 4: Implement thin dependency probes**

Each adapter must have a stable public name and return the client error without logging it:

```go
type Probe struct { Pool *pgxpool.Pool }
func (Probe) Name() string { return "postgres" }
func (p Probe) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }
```

Implement equivalent `redis.Probe` using `Client.Ping(ctx).Err()` and `nats.Probe` using `Conn.FlushWithContext(ctx)`. Keep these packages free of HTTP and metrics concerns.

- [ ] **Step 5: Verify and commit**

Run: `gofmt -w internal/readiness internal/platform`

Run: `go test -race ./internal/readiness ./internal/platform/...`

Expected: PASS with no race report.

```bash
git add internal/readiness internal/platform
git commit -m "feat: add bounded dependency readiness checks"
```

### Task 9: Runnable control API with graceful shutdown

**Files:**
- Create: `internal/controlapi/handler_test.go`
- Create: `internal/controlapi/handler.go`
- Create: `internal/platform/dependencies.go`
- Create: `internal/platform/server.go`
- Create: `cmd/control-api/main.go`

**Interfaces:**
- Consumes: `config.Config`, `readiness.Checker`, generated `controlapiv1.ServerInterface`, pgxpool, Redis client and NATS connection.
- Produces: HTTP `GET /livez`, `GET /readyz`, `platform.Open(ctx, cfg)`, `(*Dependencies).Close()`, and a runnable `control-api` binary.

- [ ] **Step 1: Write handler tests against the generated router**

Use this table test; it does not require Docker:

```go
package controlapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/readiness"
)

type testProbe struct { name string; err error }
func (p testProbe) Name() string { return p.name }
func (p testProbe) Ping(context.Context) error { return p.err }

func TestHealthRoutes(t *testing.T) {
	checker := readiness.New(50*time.Millisecond,
		testProbe{name: "postgres"},
		testProbe{name: "redis", err: errors.New("redis://user:secret@example")},
	)
	router := controlapiv1.HandlerFromMux(NewHandler(checker), http.NewServeMux())

	tests := []struct{ path string; status int }{
		{path: "/livez", status: http.StatusOK},
		{path: "/readyz", status: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tt.status { t.Fatalf("%s got %d want %d", tt.path, rec.Code, tt.status) }
		body, err := io.ReadAll(rec.Result().Body)
		if err != nil { t.Fatal(err) }
		if strings.Contains(string(body), "secret") || strings.Contains(string(body), "example") {
			t.Fatalf("sensitive dependency error leaked: %s", body)
		}
	}
}
```

- [ ] **Step 2: Verify handler tests fail**

Run: `go test ./internal/controlapi`

Expected: compilation fails because `Handler` is undefined.

- [ ] **Step 3: Implement the generated interface**

```go
package controlapi

import (
	"encoding/json"
	"net/http"

	controlapiv1 "talenro.local/platform/gen/go/talenro/controlapi/v1"
	"talenro.local/platform/internal/readiness"
)

type Handler struct { checker *readiness.Checker }

func NewHandler(checker *readiness.Checker) *Handler { return &Handler{checker: checker} }

func (h *Handler) GetLiveness(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, http.StatusOK, controlapiv1.HealthResponse{Status: controlapiv1.HealthResponseStatusOk, Checks: map[string]string{}})
}

func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ready, checks := h.checker.Check(r.Context())
	if ready {
		writeHealth(w, http.StatusOK, controlapiv1.HealthResponse{Status: controlapiv1.HealthResponseStatusOk, Checks: checks})
		return
	}
	writeHealth(w, http.StatusServiceUnavailable, controlapiv1.HealthResponse{Status: controlapiv1.HealthResponseStatusUnavailable, Checks: checks})
}

func writeHealth(w http.ResponseWriter, status int, body controlapiv1.HealthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
```

- [ ] **Step 4: Implement dependency lifecycle**

`platform.Open(ctx, cfg)` must:

1. Create a pgx pool and ping it within `cfg.DependencyTimeout`.
2. Create a Redis client with explicit 2-second dial/read/write timeouts and ping it.
3. Connect NATS with name `talenro-control-api`, 2-second timeout, reconnect wait 500 ms, max reconnects 10, and no callback that logs message payloads.
4. Return a `Dependencies` struct only after every dependency succeeds.
5. Close already-opened clients in reverse order on partial failure.

`(*Dependencies).Close()` drains NATS with a 5-second deadline, closes Redis, then closes pgx. It must be idempotent through `sync.Once`.

Implement the lifecycle with this structure; keep the timeout values explicit:

```go
package platform

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"talenro.local/platform/internal/config"
)

type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis *redis.Client
	NATS *nats.Conn
	closeOnce sync.Once
}

func Open(ctx context.Context, cfg config.Config) (_ *Dependencies, err error) {
	deps := new(Dependencies)
	defer func() { if err != nil { deps.Close() } }()

	checkCtx, cancel := context.WithTimeout(ctx, cfg.DependencyTimeout)
	defer cancel()
	deps.Postgres, err = pgxpool.New(checkCtx, cfg.DatabaseURL)
	if err != nil { return nil, fmt.Errorf("create postgres pool: %w", err) }
	if err = deps.Postgres.Ping(checkCtx); err != nil { return nil, fmt.Errorf("ping postgres: %w", err) }

	deps.Redis = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
		DialTimeout: 2 * time.Second,
		ReadTimeout: 2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	if err = deps.Redis.Ping(checkCtx).Err(); err != nil { return nil, fmt.Errorf("ping redis: %w", err) }

	deps.NATS, err = nats.Connect(cfg.NATSURL,
		nats.Name("talenro-control-api"),
		nats.Timeout(2*time.Second),
		nats.ReconnectWait(500*time.Millisecond),
		nats.MaxReconnects(10),
		nats.DrainTimeout(5*time.Second),
	)
	if err != nil { return nil, fmt.Errorf("connect nats: %w", err) }
	return deps, nil
}

func (d *Dependencies) Close() {
	if d == nil { return }
	d.closeOnce.Do(func() {
		if d.NATS != nil { _ = d.NATS.Drain(); d.NATS.Close() }
		if d.Redis != nil { _ = d.Redis.Close() }
		if d.Postgres != nil { d.Postgres.Close() }
	})
}
```

- [ ] **Step 5: Implement HTTP server construction**

`server.go` must construct an `http.Server` with:

```go
ReadHeaderTimeout: 5 * time.Second
ReadTimeout:       10 * time.Second
WriteTimeout:      10 * time.Second
IdleTimeout:       60 * time.Second
MaxHeaderBytes:    16 << 10
```

Register only generated control API routes in this task. Do not add debug endpoints or `net/http/pprof` to the public listener.

```go
func NewHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: address,
		Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
		MaxHeaderBytes: 16 << 10,
	}
}
```

- [ ] **Step 6: Wire process startup and signal handling**

`main.go` must use `signal.NotifyContext` for `os.Interrupt` and `syscall.SIGTERM`, load configuration from `os.LookupEnv`, open dependencies, build probes, start the HTTP server, and call `Shutdown` with `cfg.ShutdownTimeout`. Log only stable event names, build info, listener address and sanitized error categories through `log/slog`; never log full database URLs or NATS/Redis credentials.

Keep `main` thin and put cancellation behavior in `run` so it can be tested:

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.LookupEnv); err != nil {
		slog.Error("control_api_stopped", "category", "startup_or_runtime")
		os.Exit(1)
	}
}

func run(ctx context.Context, lookup config.Lookup) error {
	cfg, err := config.Load(lookup)
	if err != nil { return fmt.Errorf("load config: %w", err) }
	deps, err := platform.Open(ctx, cfg)
	if err != nil { return fmt.Errorf("open dependencies: %w", err) }
	defer deps.Close()

	checker := readiness.New(cfg.DependencyTimeout,
		postgresprobe.Probe{Pool: deps.Postgres},
		redisprobe.Probe{Client: deps.Redis},
		natsprobe.Probe{Conn: deps.NATS},
	)
	mux := http.NewServeMux()
	handler := controlapiv1.HandlerFromMux(controlapi.NewHandler(checker), mux)
	server := platform.NewHTTPServer(cfg.HTTPAddress, handler)
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) { return nil }
		return fmt.Errorf("serve HTTP: %w", err)
	}
}
```

Import the three probe packages with aliases `postgresprobe`, `redisprobe`, and `natsprobe`. Graceful cancellation is exercised by the Unix smoke test in Task 12 after real dependencies are ready.

Build with:

```powershell
go build -trimpath -ldflags "-X talenro.local/platform/internal/buildinfo.version=0.0.0-dev -X talenro.local/platform/internal/buildinfo.commit=local" -o bin/control-api.exe ./cmd/control-api
```

- [ ] **Step 7: Run unit and integration smoke tests**

Run: `go test -race ./internal/controlapi ./internal/platform/...`

Run dependencies and migrations, then start the binary and request:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/livez
Invoke-RestMethod http://127.0.0.1:8080/readyz
```

Expected: both return `status = ok`; stopping PostgreSQL makes `/readyz` return HTTP 503 while `/livez` remains HTTP 200.

- [ ] **Step 8: Commit**

```bash
git add cmd/control-api internal/controlapi internal/platform
git commit -m "feat: add runnable control API health service"
```

### Task 10: Privacy-safe Prometheus observability

**Files:**
- Create: `internal/observability/metrics_test.go`
- Create: `internal/observability/metrics.go`
- Create: `internal/observability/http.go`
- Modify: `.env.example`
- Modify: `internal/config/config_test.go`
- Modify: `internal/config/config.go`
- Modify: `internal/platform/server.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Consumes: generated route patterns and HTTP status codes.
- Produces: private-listener `/metrics`, `observability.Registry`, and middleware with bounded labels `method`, `route`, `status_class`.

- [ ] **Step 1: Add the Prometheus client**

Run: `go get github.com/prometheus/client_golang@v1.23.2`

Run: `go mod tidy`

- [ ] **Step 2: Write label-cardinality tests**

Create a test registry and prove that raw request data does not become labels:

```go
package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareUsesOnlyBoundedRouteLabel(t *testing.T) {
	registry := NewRegistry()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := registry.Middleware("/readyz", next)
	for _, target := range []string{
		"/readyz?email=user@example.invalid",
		"/random-device-123?target=198.51.100.1",
	} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
	}
	families, err := registry.Gatherer.Gather()
	if err != nil { t.Fatal(err) }
	for _, family := range families {
		encoded := family.String()
		for _, forbidden := range []string{"example.invalid", "198.51.100.1", "random-device-123"} {
			if strings.Contains(encoded, forbidden) { t.Fatalf("unbounded label leaked: %s", forbidden) }
		}
		if strings.Contains(family.GetName(), "http_requests") &&
			(!strings.Contains(encoded, `name:"route"`) || !strings.Contains(encoded, `value:"/readyz"`)) {
			t.Fatalf("expected normalized route label: %s", encoded)
		}
	}
}
```

- [ ] **Step 3: Implement a private registry and middleware**

Implement `Registry`, `NewRegistry`, `Middleware`, and `Handler` as follows:

```go
type Registry struct {
	prometheus.Registerer
	Gatherer prometheus.Gatherer
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
	BuildInfo *prometheus.GaugeVec
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "talenro_control_http_requests_total",
		Help: "Completed control API HTTP requests.",
	}, []string{"method", "route", "status_class"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "talenro_control_http_request_duration_seconds",
		Help: "Control API HTTP request duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status_class"})
	build := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "talenro_control_build_info",
		Help: "Talenro control API build identity.",
	}, []string{"version", "commit"})
	info := buildinfo.Current()
	build.WithLabelValues(info.Version, info.Commit).Set(1)
	registry.MustRegister(requests, duration, build)
	return &Registry{Registerer: registry, Gatherer: registry, Requests: requests, Duration: duration, BuildInfo: build}
}

func (r *Registry) Middleware(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, req)
		statusClass := strconv.Itoa(recorder.status/100) + "xx"
		labels := []string{req.Method, route, statusClass}
		r.Requests.WithLabelValues(labels...).Inc()
		r.Duration.WithLabelValues(labels...).Observe(time.Since(started).Seconds())
	})
}

func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.Gatherer, promhttp.HandlerOpts{})
}
```

Use metric names:

```text
talenro_control_http_requests_total
talenro_control_http_request_duration_seconds
talenro_control_build_info
```

Allowed labels are exactly `method`, `route`, and `status_class`; `build_info` may label `version` and `commit`. Do not label account, device, node, POP, email, country, IP, domain, protocol credential or error string.

- [ ] **Step 4: Separate public health and private metrics listeners**

Add `TALENRO_METRICS_ADDRESS` with default `127.0.0.1:9090` to `config.Config`. Start a second `http.Server` that registers only `/metrics`; keep it loopback-only unless explicit deployment configuration puts it behind authenticated infrastructure. Both servers must share the same cancellation and graceful-shutdown path.

Extend `TestLoadAppliesSafeDefaults` with:

```go
if got.MetricsAddress != "127.0.0.1:9090" {
	t.Fatalf("unexpected metrics address: %s", got.MetricsAddress)
}
```

Add the field and parsing with the same `net.SplitHostPort` and unspecified-address rejection used for `HTTPAddress`:

```go
MetricsAddress string
```

Wire the private server without mounting public API routes:

```go
metrics := observability.NewRegistry()
metricsMux := http.NewServeMux()
metricsMux.Handle("/metrics", metrics.Handler())
metricsServer := platform.NewHTTPServer(cfg.MetricsAddress, metricsMux)

publicMux := http.NewServeMux()
publicHandler := controlapiv1.HandlerFromMux(controlapi.NewHandler(checker), publicMux)
publicServer := platform.NewHTTPServer(cfg.HTTPAddress, publicHandler)
```

Start both servers under one `errCh`. On cancellation call `Shutdown` for the metrics server first and the public server second using the same bounded shutdown context. A non-`http.ErrServerClosed` error from either listener cancels the sibling listener and is returned from `run`.

- [ ] **Step 5: Verify and commit**

Run: `gofmt -w internal/observability internal/platform internal/config`

Run: `go test -race ./internal/observability ./internal/platform/... ./internal/config`

Expected: PASS and no unbounded label appears.

```bash
git add internal/observability internal/platform internal/config go.mod go.sum
git commit -m "feat: add privacy-safe service metrics"
```

### Task 11: One-command verification and supply-chain gates

**Files:**
- Create: `.golangci.yml`
- Create: `scripts/generate.ps1`
- Create: `scripts/generate.sh`
- Create: `scripts/verify.ps1`
- Create: `scripts/verify.sh`
- Create: `docs/decisions/0001-control-plane-foundation.md`
- Create: `docs/licenses/dependency-policy.md`

**Interfaces:**
- Consumes: every source, generated file and unit test created in Tasks 1–10.
- Produces: deterministic `generate` and `verify` entrypoints used locally and by any future CI provider.

- [ ] **Step 1: Configure a focused linter set**

`.golangci.yml`：

```yaml
version: "2"
run:
  timeout: 5m
linters:
  default: none
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - bodyclose
    - contextcheck
    - errorlint
    - exhaustive
    - gosec
    - misspell
    - nilerr
    - noctx
    - revive
formatters:
  enable:
    - gofmt
    - goimports
issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

- [ ] **Step 2: Add deterministic generation scripts**

`scripts/generate.ps1`：

```powershell
$ErrorActionPreference = 'Stop'

function Invoke-Checked([scriptblock]$Command) {
  & $Command
  if ($LASTEXITCODE -ne 0) { throw "Command failed: $Command" }
}

Invoke-Checked { go tool buf lint }
Invoke-Checked { go tool buf generate }
Invoke-Checked { go tool oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/control-api.v1.yaml }
Invoke-Checked { go tool sqlc generate }
Invoke-Checked { gofmt -w gen/go internal/store }
```

`scripts/generate.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail

go tool buf lint
go tool buf generate
go tool oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/control-api.v1.yaml
go tool sqlc generate
gofmt -w gen/go internal/store
```

- [ ] **Step 3: Add deterministic verification scripts**

`scripts/verify.ps1`：

```powershell
$ErrorActionPreference = 'Stop'

function Invoke-Checked([scriptblock]$Command) {
  & $Command
  if ($LASTEXITCODE -ne 0) { throw "Command failed: $Command" }
}

& "$PSScriptRoot/check-tools.ps1"
& "$PSScriptRoot/generate.ps1"
Invoke-Checked { git diff --exit-code -- api gen internal/store }
Invoke-Checked { go vet ./... }
Invoke-Checked { go tool golangci-lint run ./... }
Invoke-Checked { go test -race -count=1 ./... }
Invoke-Checked { docker compose -f deploy/dev/compose.yaml config --quiet }
```

`scripts/verify.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"${script_dir}/check-tools.sh"
"${script_dir}/generate.sh"
git diff --exit-code -- api gen internal/store
go vet ./...
go tool golangci-lint run ./...
go test -race -count=1 ./...
docker compose -f deploy/dev/compose.yaml config --quiet
```

The integration suite remains an explicit command because it starts containers. On Windows use:

```powershell
$ErrorActionPreference = 'Stop'
$env:TALENRO_DATABASE_URL='postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
try {
  docker compose -f deploy/dev/compose.yaml up -d --wait
  if ($LASTEXITCODE -ne 0) { throw 'compose up failed' }
  go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
  if ($LASTEXITCODE -ne 0) { throw 'migration failed' }
  go test -tags=integration -count=1 ./internal/testinfra ./internal/store
  if ($LASTEXITCODE -ne 0) { throw 'integration tests failed' }
} finally {
  docker compose -f deploy/dev/compose.yaml down
}
```

On Unix use:

```bash
set -euo pipefail
export TALENRO_DATABASE_URL='postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
trap 'docker compose -f deploy/dev/compose.yaml down' EXIT
docker compose -f deploy/dev/compose.yaml up -d --wait
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" up
go test -tags=integration -count=1 ./internal/testinfra ./internal/store
```

- [ ] **Step 4: Record the stack decision and license boundary**

The ADR must state:

- Go is selected for control-plane services and node agents.
- PostgreSQL is authoritative; Redis is rebuildable; NATS is at-least-once.
- OpenAPI governs client-facing HTTP and Protobuf governs service events.
- This plan intentionally uses a modular monolith process first; service extraction requires measured scaling or failure-isolation evidence.
- Redis 8.8.1 is consumed unmodified under the selected AGPLv3 path for development; production use is blocked until legal review or commercial licensing is recorded.
- sing-box and Xray remain process boundaries outside the proprietary control-plane binary.

- [ ] **Step 5: Run the complete local gate and commit**

Run on Windows: `powershell -ExecutionPolicy Bypass -File scripts/verify.ps1`

Expected: all checks pass and `git status --short` is empty before the commit except for the files in this task.

```bash
git add .golangci.yml scripts docs/decisions docs/licenses
git commit -m "build: add deterministic verification gates"
```

### Task 12: End-to-end smoke runbook and foundation acceptance

**Files:**
- Create: `scripts/smoke.ps1`
- Create: `scripts/smoke.sh`
- Create: `docs/runbooks/local-development.md`
- Create: `docs/runbooks/foundation-failures.md`
- Create: `docs/roadmap/implementation-sequence.md`

**Interfaces:**
- Consumes: control-api binary, dependency compose stack, migrations and verification scripts.
- Produces: repeatable foundation acceptance with failure injection and the ordered list of subsequent plans.

- [ ] **Step 1: Implement the smoke scripts**

The scripts must:

1. Start `deploy/dev/compose.yaml` with `--wait`.
2. Export the exact local environment values from `.env.example` without evaluating arbitrary shell text.
3. Apply migrations with goose.
4. Start `go run ./cmd/control-api` as a child process.
5. Poll `/livez` for at most 10 seconds.
6. Assert `/livez` and `/readyz` return 200.
7. Stop the PostgreSQL container only.
8. Assert `/readyz` returns 503 within 5 seconds while `/livez` remains 200.
9. Restart PostgreSQL and assert `/readyz` returns 200 within 10 seconds.
10. Terminate the child process gracefully and stop the compose stack in a finally/trap block.

Never print `.env`, process environment, database URL, headers, response bodies from future authenticated endpoints, or Docker inspect output.

`scripts/smoke.ps1`：

```powershell
$ErrorActionPreference = 'Stop'
$compose = @('-f', 'deploy/dev/compose.yaml')
$env:TALENRO_HTTP_ADDRESS = '127.0.0.1:8080'
$env:TALENRO_DATABASE_URL = 'postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
$env:TALENRO_REDIS_ADDRESS = '127.0.0.1:6379'
$env:TALENRO_NATS_URL = 'nats://127.0.0.1:4222'
$client = [System.Net.Http.HttpClient]::new()
$process = $null

function Get-Status([string]$Url) {
  $response = $client.GetAsync($Url).GetAwaiter().GetResult()
  return [int]$response.StatusCode
}

function Wait-Status([string]$Url, [int]$Expected, [int]$Seconds) {
  $deadline = [DateTimeOffset]::UtcNow.AddSeconds($Seconds)
  do {
    try {
      if ((Get-Status $Url) -eq $Expected) { return }
    } catch { }
    Start-Sleep -Milliseconds 250
  } while ([DateTimeOffset]::UtcNow -lt $deadline)
  throw "$Url did not reach HTTP $Expected within $Seconds seconds"
}

try {
  docker compose @compose up -d --wait
  if ($LASTEXITCODE -ne 0) { throw 'compose up failed' }
  go tool goose -dir db/migrations postgres $env:TALENRO_DATABASE_URL up
  if ($LASTEXITCODE -ne 0) { throw 'migration failed' }
  $process = Start-Process -FilePath 'go' -ArgumentList @('run', './cmd/control-api') -PassThru -NoNewWindow
  Wait-Status 'http://127.0.0.1:8080/livez' 200 10
  Wait-Status 'http://127.0.0.1:8080/readyz' 200 10
  docker compose @compose stop postgres
  if ($LASTEXITCODE -ne 0) { throw 'postgres stop failed' }
  Wait-Status 'http://127.0.0.1:8080/readyz' 503 5
  Wait-Status 'http://127.0.0.1:8080/livez' 200 2
  docker compose @compose start postgres
  if ($LASTEXITCODE -ne 0) { throw 'postgres start failed' }
  Wait-Status 'http://127.0.0.1:8080/readyz' 200 10
} finally {
  if ($null -ne $process -and -not $process.HasExited) {
    Stop-Process -Id $process.Id -Force
    $process.WaitForExit()
  }
  $client.Dispose()
  docker compose @compose down
}
```

The Windows script uses forced child cleanup because PowerShell cannot portably deliver SIGTERM to another console process. The graceful signal path is verified by the Unix smoke script and by HTTP server shutdown tests.

`scripts/smoke.sh`：

```bash
#!/usr/bin/env bash
set -euo pipefail

compose=(docker compose -f deploy/dev/compose.yaml)
export TALENRO_HTTP_ADDRESS='127.0.0.1:8080'
export TALENRO_DATABASE_URL='postgres://talenro:talenro_dev@127.0.0.1:5432/talenro?sslmode=disable'
export TALENRO_REDIS_ADDRESS='127.0.0.1:6379'
export TALENRO_NATS_URL='nats://127.0.0.1:4222'
api_pid=''

cleanup() {
  if [[ -n "${api_pid}" ]] && kill -0 "${api_pid}" 2>/dev/null; then
    kill -TERM "${api_pid}"
    wait "${api_pid}"
  fi
  "${compose[@]}" down
}
trap cleanup EXIT

wait_status() {
  local url="$1" expected="$2" seconds="$3" deadline
  deadline=$((SECONDS + seconds))
  while (( SECONDS < deadline )); do
    if [[ "$(curl --silent --output /dev/null --write-out '%{http_code}' "${url}" || true)" == "${expected}" ]]; then
      return 0
    fi
    sleep 0.25
  done
  echo "${url} did not reach HTTP ${expected}" >&2
  return 1
}

"${compose[@]}" up -d --wait
go tool goose -dir db/migrations postgres "${TALENRO_DATABASE_URL}" up
go run ./cmd/control-api &
api_pid=$!
wait_status 'http://127.0.0.1:8080/livez' 200 10
wait_status 'http://127.0.0.1:8080/readyz' 200 10
"${compose[@]}" stop postgres
wait_status 'http://127.0.0.1:8080/readyz' 503 5
wait_status 'http://127.0.0.1:8080/livez' 200 2
"${compose[@]}" start postgres
wait_status 'http://127.0.0.1:8080/readyz' 200 10
```

- [ ] **Step 2: Write the local development runbook**

Document prerequisites, pinned versions, startup, migration, generation, unit tests, integration tests, smoke test, safe log collection and normal shutdown. Include Windows PowerShell commands first and Bash equivalents second.

- [ ] **Step 3: Write the failure runbook**

For PostgreSQL, Redis and NATS, document:

- expected `/readyz` value;
- expected `/livez` value;
- local inspection command that does not reveal secrets;
- recovery command;
- the rule that clearing Redis is recoverable while deleting PostgreSQL volumes is destructive and requires explicit approval.

- [ ] **Step 4: Write the implementation sequence**

Create one paragraph for each of the eight plans listed in “Scope and delivery sequence,” with its independently testable exit condition. Mark only this foundation plan as current; do not assign dates or invent staffing estimates.

- [ ] **Step 5: Run acceptance and commit**

Run: `powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/verify.ps1`

Expected: smoke failure injection recovers, all validation passes, and `git status --short` shows only Task 12 files.

```bash
git add scripts/smoke.ps1 scripts/smoke.sh docs/runbooks docs/roadmap
git commit -m "docs: add foundation smoke and operations runbooks"
```

## Final acceptance checklist

- [ ] `go version` reports Go 1.26.5.
- [ ] No container or generator uses a floating `latest` tag.
- [ ] `scripts/verify.ps1` and `scripts/verify.sh` pass on their supported platforms.
- [ ] Unit tests pass with the race detector.
- [ ] Protobuf, OpenAPI and sqlc regeneration leaves no Git diff.
- [ ] All migrations apply, roll back one step, and reapply successfully.
- [ ] Integration tests reach PostgreSQL, Redis and JetStream.
- [ ] `/livez` stays healthy during dependency failure.
- [ ] `/readyz` reports only bounded dependency names and sanitized states.
- [ ] `/metrics` is loopback-only by default and has no user, device, node, IP, domain or error-string labels.
- [ ] Graceful shutdown stops HTTP listeners, drains NATS and closes Redis/PostgreSQL clients.
- [ ] Redis and tunnel-core license boundaries are recorded as release gates.
- [ ] The repository contains no account PII, traffic destination, real credential or secret in source, fixtures, logs or generated artifacts.
- [ ] `git status --short` is empty after the final commit.

## Official version references

- [Go release history](https://go.dev/doc/devel/release)
- [PostgreSQL versioning policy](https://www.postgresql.org/support/versioning/)
- [PostgreSQL official container](https://hub.docker.com/_/postgres)
- [Redis releases and license options](https://github.com/redis/redis/releases)
- [Redis official container](https://hub.docker.com/_/redis)
- [NATS server releases](https://github.com/nats-io/nats-server/releases)
- [NATS official container](https://hub.docker.com/_/nats/)
- [Buf releases](https://github.com/bufbuild/buf/releases)
- [Buf local plugin configuration](https://buf.build/docs/configuration/v2/buf-gen-yaml/)
- [oapi-codegen releases](https://github.com/oapi-codegen/oapi-codegen/releases)
- [sqlc releases](https://github.com/sqlc-dev/sqlc/releases)
- [goose releases](https://github.com/pressly/goose/releases)
