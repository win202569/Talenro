# Talenro Account, Device Identity, and Trust Root Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 Talenro Go 模块化单体中交付账号主体、设备 PoP 身份、隔离的 opaque token family、签名 HPKE 测试 bundle、三源不可变分发和隐私安全错误上报闭环。

**Architecture:** PostgreSQL 保存账号、设备、撤销、幂等、outbox 和最高可信版本等权威事实；Redis 只保存单次 challenge、WebAuthn ceremony 和限流窗口；NATS 按 transactional outbox 至少一次投递。`identity`、`deviceauth`、`trust` 通过应用接口和版本化事件协作，HTTP 仍由仓库 OpenAPI 生成，设备/配置信任由 Go 参考客户端做独立验证。

**Tech Stack:** Go 1.26.5、PostgreSQL 18.4、Redis 8.8.1、NATS 2.14.3、pgx 5.10.0、go-redis 9.22.0、NATS Go 1.52.0、go-webauthn 0.17.4、gowebpki/jcs 1.0.1、pquerna/otp 1.5.0、x/crypto 0.54.0、Go `crypto/hpke`、OpenAPI 3.0.3、oapi-codegen 2.8.0、sqlc 1.31.1、goose 3.27.1、Protobuf Go 1.36.11、Prometheus Go client 1.23.2、Docker Compose。

**Specification:** [Account, device identity, and trust design](../specs/2026-08-09-account-device-trust-design.md)

## Global Constraints

- 本计划只实现 C1.1；不得加入节点库存、POP agent、权益、流量账本、配额、调度、隧道凭据、Xray/sing-box 集成或四端 UI。
- Go toolchain 固定为 `go1.26.5`；所有新增模块版本必须写入 `go.mod`/`go.sum`，不得使用浮动版本。
- 新直接依赖固定为 `github.com/go-webauthn/webauthn v0.17.4`、`github.com/gowebpki/jcs v1.0.1`、`github.com/pquerna/otp v1.5.0`、`github.com/google/uuid v1.6.0`、`golang.org/x/crypto v0.54.0`。
- JCS 固定 RFC 8785；签名固定 Ed25519；HPKE 固定 Base mode + `DHKEM(X25519, HKDF-SHA256)` + `HKDF-SHA256` + `ChaCha20-Poly1305`；不提供运行时算法协商。
- Argon2id policy v1 固定 `memory=65536 KiB`、`time=3`、`parallelism=4`、16-byte salt、32-byte tag，符合 RFC 9106 第二推荐项；变更必须产生新 policy version。
- access token TTL 固定 10 分钟；account/device refresh idle TTL 固定 30 天、absolute TTL 固定 90 天；enrollment grant 10 分钟；PoP/WebAuthn challenge 2 分钟；email verification 24 小时；password reset 30 分钟；C1.1 测试 bundle 24 小时。
- idempotency 普通完成记录至少保留 24 小时；grant、rotation、revocation 等安全突变 tombstone 保留 90 天；consumed event ID 保留 30 天。
- `email_verification`：production 默认且只允许 `required` 或 `grace`；`grace` 固定一个 provisional 设备、24 小时和 `trial_restricted` policy marker；`disabled` 只允许 local/test。
- 认证请求体最大 64 KiB；加密 bundle HTTP 响应最大 1 MiB；JSON 必须拒绝未知字段、重复属性、非法 UTF-8 和 trailing data。
- 请求 deadline 默认 5s、范围 2–10s；Redis 250ms、范围 100–1000ms；signer 2s、范围 500ms–5s；error reporter 1s、范围 100ms–2s；clock skew 120s、范围 30–300s。
- Passkey 使用 discoverable credential、`attestation=none`、`residentKey=preferred`、`userVerification=required`；WebAuthn user handle 只能使用随机 principal ID，不得使用邮箱。
- TOTP 固定 6 位、30 秒 period、允许前后各一个 step；已接受 step 必须原子记录，不能重放；每账号最多 1 个 TOTP、10 个 Passkey、10 个一次性恢复码。
- 账号和设备 token 是两个 security scheme；任何 endpoint 不得接受错误域 token。
- PostgreSQL 是权威事实源；Redis 和 NATS 不能成为撤销、授权或最高 bundle version 的唯一来源。
- NATS 只按至少一次投递设计；consumer 必须使用 `event_id` 或业务幂等键去重。
- 日志、指标、HTTP 错误和 ErrorReporter 禁止邮箱、手机号、密码、token、nonce、principal/device ID、key、locator、bundle 明/密文、IP 原文、访问目标和 raw error。
- product metrics 的 method、route、result、reason、component 标签必须来自有限 allowlist；攻击者控制文本统一折叠。
- production profile 缺少外部 signer、外部 sensitive-field protector、外部 EmailSender、TLS HTTPS origin 任一项时必须拒绝启动；不得回退 local fixture。
- 所有生成物必须提交；`scripts/generate.*` 后生成目录无 diff；每个任务必须 TDD、运行 focused test，再运行受影响 package tests，并独立提交。
- Windows 验证继续使用显式 `GOOS=windows`、`GOARCH=amd64` 和任务专用 `GOCACHE`/`GOTMPDIR`；普通测试/构建显式 `CGO_ENABLED=0`，race 门显式 `CGO_ENABLED=1`；不得修改用户或全局 Go 配置。
- Redis 8.8.1 生产许可证门、sing-box GPLv3+ 门和 Xray MPL-2.0 门保持不变，本计划不能绕过。

---

## Scope check

规格已在 brainstorming 阶段拆成五个独立 C1 子规格。本计划只覆盖第一个串行安全基础：identity 建立 principal，deviceauth 依赖 principal 建立设备授权，trust 再依赖设备 HPKE 公钥签发测试 bundle；三者不能独立上线，因此保留为一份计划。其余四个 C1 规格各自另写设计和计划。

## Repository map

```text
api/
├── openapi/control-api.v1.yaml                 # 完整 C1.1 HTTP 契约源
└── proto/talenro/
    ├── events/v1/envelope.proto                # 现有至少一次事件信封
    ├── identity/v1/events.proto                # 账号/邮件安全事件
    ├── deviceauth/v1/events.proto              # 设备授权/撤销事件
    └── trust/v1/events.proto                   # bundle/trust metadata 事件
cmd/
├── control-api/main.go                         # 生产 composition root
└── trust-conformance/main.go                   # 独立 Go 参考验证器
db/
├── migrations/00002_identity.sql               # identity schema
├── migrations/00003_deviceauth.sql             # deviceauth schema
├── migrations/00004_trust_delivery.sql         # trust、outbox、idempotency schema
└── queries/
    ├── identity.sql
    ├── deviceauth.sql
    ├── trust.sql
    ├── idempotency.sql
    └── outbox.sql
internal/
├── apierrors/                                  # 有限 code/action 与安全 JSON
├── config/                                     # profile、provider 和预算护栏
├── controlapi/                                 # 生成接口的薄 HTTP adapters
├── deviceauth/                                 # grant、PoP、device token family
├── errorreport/                                # allowlist DTO 与有界 reporter
├── identity/                                   # account、password、WebAuthn、TOTP、session
├── idempotency/                                # request digest 与结果重放
├── outbox/                                     # publisher/consumer 与去重
├── secret/                                     # 不可误打印的 secret bytes
├── sensitive/                                  # lookup HMAC 与字段 AEAD
├── store/                                      # sqlc 生成代码
├── strictjson/                                 # bounded/duplicate-safe JSON
├── trust/                                      # JCS、signing metadata、HPKE envelope
└── trustclient/                                # 独立于服务端的 reference verifier
testdata/
├── crypto/rfc8785/                             # 上游官方 JCS vectors
├── crypto/rfc9180/                             # RFC HPKE vectors
└── privacy/                                    # secret canaries
```

## Interfaces frozen by this plan

```go
package identity

type PrincipalID string
type SessionID string

type Application interface {
	RegisterAccount(context.Context, RegisterAccountCommand) (RegisterAccountResult, error)
	CreateEmailVerificationDelivery(context.Context, CreateEmailVerificationDeliveryCommand) (RegisterAccountResult, error)
	VerifyEmail(context.Context, VerifyEmailCommand) error
	CreatePasswordResetDelivery(context.Context, CreatePasswordResetDeliveryCommand) (RegisterAccountResult, error)
	ResetPassword(context.Context, ResetPasswordCommand) (SessionTokens, error)
	ChangePassword(context.Context, ChangePasswordCommand) (SessionTokens, error)
	CreateSession(context.Context, CreateSessionCommand) (SessionTokens, error)
	CreateSessionChallenge(context.Context, CreateSessionChallengeCommand) (SessionChallenge, error)
	RotateSession(context.Context, RotateSessionCommand) (SessionTokens, error)
	RevokeSessions(context.Context, RevokeSessionsCommand) error
	CreateEnrollmentGrant(context.Context, CreateEnrollmentGrantCommand) (EnrollmentGrant, error)
}

// DeviceTransactionParticipant is implemented by identity's repository adapter
// and is called only inside a deviceauth-owned PostgreSQL transaction.
type DeviceTransactionParticipant interface {
	BindSessionToAuthorization(context.Context, store.DBTX, SessionID, uuid.UUID, time.Time) error
	RevokeAuthorizationSessions(context.Context, store.DBTX, uuid.UUID, time.Time) error
}

// DeviceAuthorizationParticipant is implemented by deviceauth and lets
// identity upgrade provisional authorizations without reading device tables.
type DeviceAuthorizationParticipant interface {
	ActivateVerifiedPrincipal(context.Context, store.DBTX, PrincipalID, time.Time) error
}

```

```go
package deviceauth

type Application interface {
	CreateChallenge(context.Context, CreateChallengeCommand) (Challenge, error)
	RegisterDevice(context.Context, RegisterDeviceCommand) (DeviceTokens, error)
	RotateDeviceToken(context.Context, RotateDeviceTokenCommand) (DeviceTokens, error)
	RevokeDevice(context.Context, RevokeDeviceCommand) error
	AuthorizeBundle(context.Context, AuthorizeBundleQuery) (BundleAuthority, error)
}

type TrustTransactionParticipant interface {
	AuthorizeBundleInTransaction(context.Context, store.DBTX, AuthorizeBundleQuery) (BundleAuthority, error)
}
```

```go
package trust

type ConfigSigner interface {
	KeyID() string
	Sign(context.Context, []byte) ([]byte, error)
}

type Application interface {
	Issue(context.Context, IssueCommand) (IssuedBundle, error)
	Resolve(context.Context, ResolveQuery) (Resolution, error)
	Acknowledge(context.Context, AcknowledgeCommand) error
}
```

```go
package sensitive

type EncryptedField struct {
	KeyVersion uint32
	Ciphertext []byte
}

type Protector interface {
	LookupDigest(domain string, canonical []byte) [32]byte
	Encrypt(domain string, plaintext []byte) (EncryptedField, error)
	Decrypt(domain string, value EncryptedField) ([]byte, error)
}
```

### Task 1: Pin security dependencies and enforce runtime profiles

**Files:**
- Create: `internal/secret/bytes_test.go`
- Create: `internal/secret/bytes.go`
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `internal/config/config_test.go`
- Modify: `internal/config/config.go`
- Modify: `.env.example`
- Modify: `scripts/smoke.ps1`
- Modify: `scripts/smoke.sh`
- Modify: `docs/licenses/dependency-policy.md`

**Interfaces:**
- Consumes: existing `config.Lookup` and smoke parser allowlist.
- Produces: `secret.Bytes`, `config.Profile`, `config.EmailVerificationMode`, `config.Provider`, and validated security/provider/timeout fields on `config.Config`.

- [ ] **Step 1: Write failing redaction and profile tests**

```go
package secret

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"
)

func TestBytesNeverFormatsOrMarshalsSecret(t *testing.T) {
	value := NewBytes([]byte("SECRET-CANARY"))
	if got := fmt.Sprintf("%v", value); got != "[REDACTED]" {
		t.Fatalf("formatted secret = %q", got)
	}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("secret unexpectedly marshaled")
	}
	copyValue := value.Copy()
	copyValue[0] = 'X'
	if string(value.Copy()) != "SECRET-CANARY" {
		t.Fatal("Copy exposed mutable backing storage")
	}
}
```

Add table tests to `internal/config/config_test.go` with these exact cases:

```go
func TestLoadRejectsUnsafeProductionSecurityProviders(t *testing.T) {
	base := map[string]string{
		"TALENRO_DATABASE_URL":           "database-fixture",
		"TALENRO_PROFILE":                "production",
		"TALENRO_PUBLIC_BASE_URL":        "https://api.example.invalid",
		"TALENRO_PRIMARY_BUNDLE_BASE_URL": "https://api.example.invalid",
		"TALENRO_MIRROR_A_BASE_URL":       "https://mirror-a.example.invalid",
		"TALENRO_MIRROR_B_BASE_URL":       "https://mirror-b.example.invalid",
		"TALENRO_WEBAUTHN_RP_ID":         "example.invalid",
		"TALENRO_WEBAUTHN_ORIGINS":       "https://app.example.invalid",
		"TALENRO_EMAIL_VERIFICATION_MODE": "required",
		"TALENRO_SIGNER_PROVIDER":         "external",
		"TALENRO_FIELD_PROTECTOR_PROVIDER": "external",
		"TALENRO_EMAIL_PROVIDER":          "external",
		"TALENRO_ERROR_REPORTER_PROVIDER": "external",
	}

	tests := []struct{ key, value string }{
		{key: "TALENRO_SIGNER_PROVIDER", value: "local"},
		{key: "TALENRO_FIELD_PROTECTOR_PROVIDER", value: "local"},
		{key: "TALENRO_EMAIL_PROVIDER", value: "local"},
		{key: "TALENRO_PUBLIC_BASE_URL", value: "http://example.invalid"},
		{key: "TALENRO_EMAIL_VERIFICATION_MODE", value: "disabled"},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			values := maps.Clone(base)
			values[test.key] = test.value
			_, err := Load(lookup(values))
			if err == nil || strings.Contains(err.Error(), test.value) {
				t.Fatalf("expected sanitized rejection, got %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests and verify RED**

Run on Windows with process-scoped `GOOS=windows`, `GOARCH=amd64`, `CGO_ENABLED=0`, and task-local `GOCACHE`/`GOTMPDIR`: `go test ./internal/secret ./internal/config -count=1`

Expected: compilation fails because `secret.NewBytes`, profile fields, and provider parsing are absent.

- [ ] **Step 3: Pin the reviewed dependencies**

Run exactly:

```powershell
go get github.com/go-webauthn/webauthn@v0.17.4
go get github.com/gowebpki/jcs@v1.0.1
go get github.com/google/uuid@v1.6.0
go get github.com/pquerna/otp@v1.5.0
go get golang.org/x/crypto@v0.54.0
go mod tidy
```

Record BSD-3-Clause/Apache-2.0 notices and upstream URLs in `docs/licenses/dependency-policy.md`; do not change the existing Redis/Xray/sing-box gates.

- [ ] **Step 4: Implement the non-printable secret wrapper**

```go
// Package secret prevents accidental formatting or JSON encoding of secret bytes.
package secret

import "errors"

// Bytes owns a private copy of sensitive bytes.
type Bytes struct{ value []byte }

// NewBytes copies value into a redacted container.
func NewBytes(value []byte) Bytes {
	return Bytes{value: append([]byte(nil), value...)}
}

// Copy returns an isolated copy for a cryptographic adapter.
func (b Bytes) Copy() []byte { return append([]byte(nil), b.value...) }

func (Bytes) String() string { return "[REDACTED]" }

// MarshalJSON always fails closed.
func (Bytes) MarshalJSON() ([]byte, error) {
	return nil, errors.New("secret: serialization forbidden")
}
```

- [ ] **Step 5: Extend typed configuration with exact defaults and guards**

Add these types and fields; parsing helpers must return value-free errors:

```go
type Profile string
type EmailVerificationMode string
type Provider string

const (
	ProfileLocal      Profile = "local"
	ProfileTest       Profile = "test"
	ProfileProduction Profile = "production"
	EmailRequired     EmailVerificationMode = "required"
	EmailGrace        EmailVerificationMode = "grace"
	EmailDisabled     EmailVerificationMode = "disabled"
	ProviderLocal     Provider = "local"
	ProviderExternal  Provider = "external"
	ProviderDiscard   Provider = "discard"
)

type SecurityConfig struct {
	Profile                Profile
	EmailVerification      EmailVerificationMode
	PublicBaseURL          string
	BundleBaseURLs         [3]string
	WebAuthnRPID           string
	WebAuthnOrigins        []string
	SignerProvider         Provider
	FieldProtectorProvider Provider
	EmailProvider          Provider
	ErrorReporterProvider  Provider
	RequestDeadline        time.Duration
	RedisTimeout           time.Duration
	SignerTimeout          time.Duration
	ErrorReportTimeout     time.Duration
	ClockSkew              time.Duration
	LoginRateLimit         RateLimitPolicy
	DeliveryRateLimit      RateLimitPolicy
	ChallengeRateLimit     RateLimitPolicy
	SensitiveLookupKey     secret.Bytes
	SensitiveEncryptionKey secret.Bytes
	LocalRootSigningSeed   secret.Bytes
	LocalConfigSigningSeed secret.Bytes
}
```

Local/test defaults are the Global Constraints values. Production requires HTTPS for `PublicBaseURL` and all three distinct bundle base URLs, non-local providers, `required|grace`, and no local key material. Bundle base URLs contain scheme/authority only, have no query/fragment/userinfo, and are normalized without a trailing slash. Decode each local key from unpadded base64url and require exactly 32 bytes.

- [ ] **Step 6: Update local environment and both smoke allowlists**

Add these development-only keys to `.env.example` and both exact TALENRO allowlists:

```dotenv
TALENRO_PROFILE=local
TALENRO_PUBLIC_BASE_URL=http://localhost:8080
TALENRO_PRIMARY_BUNDLE_BASE_URL=http://localhost:8080
TALENRO_MIRROR_A_BASE_URL=http://localhost:8081
TALENRO_MIRROR_B_BASE_URL=http://localhost:8082
TALENRO_WEBAUTHN_RP_ID=localhost
TALENRO_WEBAUTHN_ORIGINS=http://localhost:8080
TALENRO_EMAIL_VERIFICATION_MODE=disabled
TALENRO_SIGNER_PROVIDER=local
TALENRO_FIELD_PROTECTOR_PROVIDER=local
TALENRO_EMAIL_PROVIDER=local
TALENRO_ERROR_REPORTER_PROVIDER=discard
TALENRO_REQUEST_DEADLINE=5s
TALENRO_REDIS_TIMEOUT=250ms
TALENRO_SIGNER_TIMEOUT=2s
TALENRO_ERROR_REPORT_TIMEOUT=1s
TALENRO_CLOCK_SKEW=120s
TALENRO_LOGIN_RATE_LIMIT=10
TALENRO_LOGIN_RATE_WINDOW=15m
TALENRO_DELIVERY_RATE_LIMIT=5
TALENRO_DELIVERY_RATE_WINDOW=1h
TALENRO_CHALLENGE_RATE_LIMIT=20
TALENRO_CHALLENGE_RATE_WINDOW=5m
TALENRO_SENSITIVE_LOOKUP_KEY_B64=YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU
TALENRO_SENSITIVE_ENCRYPTION_KEY_B64=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY
TALENRO_LOCAL_ROOT_SIGNING_SEED_B64=cm9vdC1zaWduaW5nLXNlZWQtZm9yLWxvY2FsLXRlc3Q
TALENRO_LOCAL_CONFIG_SIGNING_SEED_B64=Y29uZmlnLXNpZ25pbmctc2VlZC1sb2NhbC10ZXN0LTE
```

The implementation must reject these known fixture values outside `local|test`. Preserve smoke behavior: only exact allowlisted keys load, ambient `TALENRO_*` values are removed, and output remains sanitized.

`RateLimitPolicy` is `{Limit uint32; Window time.Duration}`. Startup ranges are login `5..50` per `5m..1h`, delivery `1..20` per `10m..24h`, and challenge/rotation `5..100` per `1m..30m`; clients and commercial plans cannot override them.

- [ ] **Step 7: Verify and commit**

Run: `go test ./internal/secret ./internal/config -count=1`

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/smoke.ps1`

Expected: focused tests PASS; smoke either passes with Docker or fails only at a dependency stage with a sanitized status and no secret value.

```bash
git add go.mod go.sum .env.example internal/secret internal/config scripts/smoke.ps1 scripts/smoke.sh docs/licenses/dependency-policy.md
git commit -m "build: pin identity and trust dependencies"
```

### Task 2: Freeze C1.1 HTTP and domain-event contracts

**Files:**
- Modify: `api/openapi/control-api.v1.yaml`
- Create: `api/proto/talenro/identity/v1/events.proto`
- Create: `api/proto/talenro/deviceauth/v1/events.proto`
- Create: `api/proto/talenro/trust/v1/events.proto`
- Modify: `internal/controlapi/contract_test.go`
- Modify: `internal/contracts/events/validate_test.go`
- Modify: `internal/contracts/events/validate.go`
- Modify: `internal/controlapi/handler.go`
- Create: `internal/controlapi/c1_unavailable.go`
- Modify generated files under: `gen/go/talenro/controlapi/v1/`, `gen/go/talenro/identity/v1/`, `gen/go/talenro/deviceauth/v1/`, `gen/go/talenro/trust/v1/`

**Interfaces:**
- Consumes: `api/proto/talenro/events/v1.EventEnvelope` and existing oapi/buf generation.
- Produces: generated HTTP request/response models, two opaque token security schemes, and PII-free versioned event payloads.

- [ ] **Step 1: Add a failing exact-surface contract test**

Replace the path expectation map with all existing health paths plus these exact operation IDs:

```go
var c11Operations = map[string]string{
	"POST /v1/accounts":                         "createAccount",
	"POST /v1/email-verification-deliveries":    "createEmailVerificationDelivery",
	"POST /v1/email-verifications":              "verifyEmail",
	"POST /v1/password-reset-deliveries":       "createPasswordResetDelivery",
	"POST /v1/password-resets":                  "resetPassword",
	"POST /v1/password-changes":                 "changePassword",
	"POST /v1/account-sessions":                 "createAccountSession",
	"POST /v1/account-auth-challenges":          "createAccountAuthChallenge",
	"POST /v1/account-token-rotations":          "rotateAccountToken",
	"POST /v1/account-session-revocations":      "revokeAccountSessions",
	"POST /v1/passkey-registration-options":     "createPasskeyRegistrationOptions",
	"POST /v1/passkey-credentials":              "createPasskeyCredential",
	"POST /v1/passkey-authentication-options":   "createPasskeyAuthenticationOptions",
	"POST /v1/passkey-revocations":              "revokePasskey",
	"POST /v1/totp-enrollments":                 "createTOTPEnrollment",
	"POST /v1/totp-verifications":               "verifyTOTPEnrollment",
	"POST /v1/totp-revocations":                 "revokeTOTP",
	"POST /v1/recovery-code-rotations":           "rotateRecoveryCodes",
	"POST /v1/recovery-code-consumptions":        "consumeRecoveryCode",
	"POST /v1/device-enrollment-grants":          "createDeviceEnrollmentGrant",
	"POST /v1/device-auth-challenges":            "createDeviceAuthChallenge",
	"POST /v1/devices":                           "registerDevice",
	"POST /v1/device-token-rotations":            "rotateDeviceToken",
	"POST /v1/device-revocations":                "revokeDevice",
	"POST /v1/config-bundle-resolutions":         "resolveConfigBundle",
	"POST /v1/config-bundle-acknowledgements":    "acknowledgeConfigBundle",
	"GET /b/{bundle_locator}":                    "getImmutableBundle",
}
```

The same test must assert `additionalProperties: false`, request body limit metadata `x-talenro-max-bytes: 65536`, bundle response media type `application/vnd.talenro.bundle+json`, and both security schemes `AccountOpaqueToken` and `DeviceOpaqueToken`.

- [ ] **Step 2: Run the contract test and verify RED**

Run: `go test ./internal/controlapi -run TestOpenAPIContract -count=1`

Expected: FAIL listing the first missing C1.1 path.

- [ ] **Step 3: Add exact shared API schemas**

Define these bounded schemas in OpenAPI; every object sets `additionalProperties: false`:

```yaml
    PublicError:
      type: object
      additionalProperties: false
      required: [code, action, trace_id]
      properties:
        code:
          type: string
          enum: [malformed_request, authentication_failed, action_not_allowed, idempotency_conflict, state_conflict, version_rollback, request_too_large, unsupported_schema, rate_limited, dependency_unavailable, signing_unavailable]
        action:
          type: string
          enum: [retry, reauthenticate, reenroll, upgrade_client, contact_support]
        retry_after_ms:
          type: integer
          format: int64
          minimum: 1
          maximum: 10000
        trace_id:
          type: string
          pattern: '^[A-Za-z0-9_-]{16,64}$'
    OpaqueToken:
      type: string
      minLength: 43
      maxLength: 86
      pattern: '^[A-Za-z0-9_-]+$'
    IdempotencyKey:
      type: string
      minLength: 22
      maxLength: 86
      pattern: '^[A-Za-z0-9_-]+$'
    TimestampString:
      type: string
      format: date-time
      maxLength: 35
```

Use body schemas with fixed string/array limits; public IDs use UUID strings; emails max 254 bytes; device display name max 64 Unicode scalar values; key fields use unpadded base64url with exact decoded length validated in handlers.

The OpenAPI operation matrix is exact; every POST requires `Idempotency-Key`, every listed bearer scheme is the only accepted scheme, and all operations reference the stable error responses:

| Operation | Security | Required request fields | Success |
| --- | --- | --- | --- |
| `createAccount` | public | `email,password,locale` | 202 generic accepted |
| `createEmailVerificationDelivery` | public | `email,locale` | 202 generic accepted |
| `verifyEmail` | public | `token` | 204 |
| `createPasswordResetDelivery` | public | `email,locale` | 202 generic accepted |
| `resetPassword` | public | `email,token,new_password,client_signing_public_key` | 200 account tokens |
| `changePassword` | `AccountOpaqueToken` | `current_password,new_password,reauthentication,client_signing_public_key` | 200 account tokens |
| `createAccountSession` | public | `method,client_signing_public_key` plus exactly one password or WebAuthn proof variant | 200 account tokens |
| `createAccountAuthChallenge` | public refresh proof | `refresh_token,request_nonce` | 201 challenge |
| `rotateAccountToken` | public refresh proof | `refresh_token,challenge_id,request_nonce,signature` | 200 account tokens |
| `revokeAccountSessions` | `AccountOpaqueToken` | `scope,reauthentication`; `session_id` only for scope `one` | 204 |
| `createPasskeyRegistrationOptions` | `AccountOpaqueToken` | `reauthentication` | 200 bounded WebAuthn options |
| `createPasskeyCredential` | `AccountOpaqueToken` | `ceremony_id,response,reauthentication` | 204 |
| `createPasskeyAuthenticationOptions` | public | no body fields | 200 bounded WebAuthn options |
| `revokePasskey` | `AccountOpaqueToken` | `credential_id,reauthentication` | 204 |
| `createTOTPEnrollment` | `AccountOpaqueToken` | `reauthentication` | 201 one-time secret/URI |
| `verifyTOTPEnrollment` | `AccountOpaqueToken` | `code,reauthentication` | 204 |
| `revokeTOTP` | `AccountOpaqueToken` | `reauthentication` | 204 |
| `rotateRecoveryCodes` | `AccountOpaqueToken` | `reauthentication` | 201 exactly ten one-time codes |
| `consumeRecoveryCode` | public recovery proof | `email,code,client_signing_public_key` | 200 account tokens |
| `createDeviceEnrollmentGrant` | `AccountOpaqueToken` | `reauthentication` | 201 one-time grant |
| `createDeviceAuthChallenge` | public grant/refresh proof | exactly one of registration (`enrollment_grant,request_nonce,signing_public_key,hpke_public_key`) or rotation (`refresh_token,request_nonce`) | 201 challenge |
| `registerDevice` | public grant+challenge proof | `enrollment_grant,challenge_id,request_nonce,signing_public_key,hpke_public_key,display_name,signature` | 201 device tokens |
| `rotateDeviceToken` | public refresh proof | `refresh_token,challenge_id,request_nonce,signature` | 200 device tokens |
| `revokeDevice` | `AccountOpaqueToken` | `device_id,reauthentication` | 204 |
| `resolveConfigBundle` | `DeviceOpaqueToken` | no body fields; test sequence is server-controlled | 200 resolution |
| `acknowledgeConfigBundle` | `DeviceOpaqueToken` | `bundle_id,bundle_version` | 204 |
| `getImmutableBundle` | public locator | path `bundle_locator` | 200 immutable bytes |

Reusable exact scalar schemas are: password 12–1024 UTF-8 bytes; locale 2–35 characters matching the database pattern; opaque token/idempotency values 22–86 base64url characters except 32-byte tokens normally encode to 43; recovery code exactly 39 characters matching `^[A-Z2-7]{4}(-[A-Z2-7]{4}){7}$`; UUID canonical lowercase text; challenge/nonce/public key 43-character base64url decoded to 32 bytes; Ed25519 signature 86-character base64url decoded to 64 bytes; TOTP exactly six ASCII digits; WebAuthn credential ID 22–1366 base64url characters decoded to 16–1024 bytes; WebAuthn response at most 64 KiB; display name at most 64 Unicode scalar values/256 UTF-8 bytes. Request objects express conditional variants with `oneOf` plus `additionalProperties:false` on each variant.

- [ ] **Step 4: Add exact event messages**

`identity/v1/events.proto`:

```proto
syntax = "proto3";
package talenro.identity.v1;
option go_package = "talenro.local/platform/gen/go/talenro/identity/v1;identityv1";

message AccountStateChanged {
  string principal_id = 1;
  string state = 2;
  uint64 version = 3;
}

message EmailDeliveryRequested {
  string delivery_id = 1;
  string principal_id = 2;
  string template_id = 3;
  string locale = 4;
}
```

`deviceauth/v1/events.proto`:

```proto
syntax = "proto3";
package talenro.deviceauth.v1;
option go_package = "talenro.local/platform/gen/go/talenro/deviceauth/v1;deviceauthv1";

message DeviceAuthorizationChanged {
  string principal_id = 1;
  string device_id = 2;
  string authorization_id = 3;
  string state = 4;
  uint64 version = 5;
}

message DeviceTokenFamilyCompromised {
  string authorization_id = 1;
  string family_id = 2;
  uint64 version = 3;
}
```

`trust/v1/events.proto`:

```proto
syntax = "proto3";
package talenro.trust.v1;
option go_package = "talenro.local/platform/gen/go/talenro/trust/v1;trustv1";

message BundleIssued {
  string authorization_id = 1;
  string bundle_id = 2;
  string bundle_version = 3;
  string envelope_sha256 = 4;
}

message BundleAcknowledged {
  string authorization_id = 1;
  string bundle_id = 2;
  string bundle_version = 3;
}
```

No event field may contain email, token, nonce, public/private key, locator, URL, ciphertext or raw error.

- [ ] **Step 5: Generate and validate determinism**

Before compiling the generated interface, embed a fail-closed adapter in `Handler`:

```go
type Handler struct {
	C1Unavailable
	checker *readiness.Checker
}

type C1Unavailable struct{}

func writeC1Unavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(controlapiv1.PublicError{
		Code:    controlapiv1.PublicErrorCode("dependency_unavailable"),
		Action:  controlapiv1.PublicErrorAction("retry"),
		TraceId: "c1-not-wired-0001",
	})
}
```

`c1_unavailable.go` implements the generated surface explicitly. Every POST operation has the exact generated signature `(http.ResponseWriter, *http.Request)` because request bodies remain in the request; the immutable GET additionally receives `bundleLocator string`:

```go
func (C1Unavailable) CreateAccount(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateEmailVerificationDelivery(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) VerifyEmail(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreatePasswordResetDelivery(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) ResetPassword(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) ChangePassword(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateAccountSession(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateAccountAuthChallenge(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RotateAccountToken(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RevokeAccountSessions(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreatePasskeyRegistrationOptions(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreatePasskeyCredential(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreatePasskeyAuthenticationOptions(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RevokePasskey(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateTOTPEnrollment(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) VerifyTOTPEnrollment(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RevokeTOTP(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RotateRecoveryCodes(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) ConsumeRecoveryCode(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateDeviceEnrollmentGrant(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) CreateDeviceAuthChallenge(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RegisterDevice(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RotateDeviceToken(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) RevokeDevice(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) ResolveConfigBundle(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) AcknowledgeConfigBundle(w http.ResponseWriter, _ *http.Request) { writeC1Unavailable(w) }
func (C1Unavailable) GetImmutableBundle(w http.ResponseWriter, _ *http.Request, _ string) { writeC1Unavailable(w) }
```

This is a real fail-closed incremental state, not a success stub. Task 17 deletes `C1Unavailable` only after every operation has a production adapter. The OpenAPI contract keeps all POST parameters in JSON bodies and the idempotency/auth values in headers, so these signatures are the generated interface contract.

Run: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/generate.ps1`

Run: `go test ./internal/controlapi ./internal/contracts/events -count=1`

Run the generator a second time, then: `git diff --exit-code -- api gen internal/store`

Expected: tests PASS and second generation produces no diff.

- [ ] **Step 6: Commit**

```bash
git add api/openapi api/proto gen/go internal/controlapi/contract_test.go internal/contracts/events
git commit -m "feat: define identity device and trust contracts"
```

### Task 3: Add reversible identity persistence and generated queries

**Files:**
- Create: `db/migrations/00002_identity.sql`
- Create: `db/queries/identity.sql`
- Create: `internal/store/identity_integration_test.go`
- Create: `internal/testinfra/postgres.go`
- Modify: `sqlc.yaml`
- Modify generated files under: `internal/store/`

**Interfaces:**
- Consumes: application-generated UUID values, `json.RawMessage`, and existing pgx/sqlc configuration.
- Produces: sqlc methods for atomic account/email/password/session/strong-credential state; later identity repositories must use these methods rather than handwritten SQL.

- [ ] **Step 1: Write the failing migration/query integration test**

```go
//go:build integration

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"talenro.local/platform/internal/store"
)

func TestIdentityAccountAndSessionRoundTrip(t *testing.T) {
	pool := openMigratedPool(t)
	queries := store.New(pool)
	principalID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := queries.CreateAccount(context.Background(), store.CreateAccountParams{
		ID: principalID, State: "pending_email", StateVersion: 1,
		Locale: "en", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := queries.GetAccountForUpdate(context.Background(), principalID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != "pending_email" || got.StateVersion != 1 {
		t.Fatalf("unexpected account: %+v", got)
	}
}
```

Create `internal/testinfra/postgres.go` with build tag `integration`. `OpenMigratedPostgres(t testing.TB)` reads only `TALENRO_DATABASE_URL`, opens a pgx pool, registers `t.Cleanup(pool.Close)`, and fails with the fixed message `testinfra: postgres unavailable` without echoing the URL. Migration setup remains in the verification script; this helper does not shell out.

- [ ] **Step 2: Run generation/test and verify RED**

Run: `go tool sqlc generate`

Expected: FAIL because `db/migrations/00002_identity.sql` and named identity queries do not exist.

Before GREEN, add these exact sqlc mappings so application IDs and timestamps do not leak pgx wrapper types into domain interfaces:

```yaml
          - db_type: uuid
            go_type:
              import: github.com/google/uuid
              type: UUID
          - db_type: uuid
            nullable: true
            go_type:
              import: github.com/google/uuid
              type: NullUUID
          - db_type: timestamptz
            go_type:
              import: time
              type: Time
          - db_type: timestamptz
            nullable: true
            go_type:
              import: database/sql
              type: NullTime
```

- [ ] **Step 3: Create the reversible identity migration**

Create schema `identity` and these exact tables/constraints:

```sql
-- +goose Up
CREATE SCHEMA identity;

CREATE TABLE identity.accounts (
  id uuid PRIMARY KEY,
  state text NOT NULL CHECK (state IN ('pending_email','active','suspended','deletion_pending','deleted')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  locale text NOT NULL CHECK (locale ~ '^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$'),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CHECK (updated_at >= created_at)
);

CREATE TABLE identity.email_identities (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL UNIQUE REFERENCES identity.accounts(id),
  lookup_key_version integer NOT NULL CHECK (lookup_key_version > 0),
  lookup_digest bytea NOT NULL UNIQUE CHECK (octet_length(lookup_digest) = 32),
  ciphertext bytea NOT NULL CHECK (octet_length(ciphertext) BETWEEN 29 AND 2048),
  encryption_key_version integer NOT NULL CHECK (encryption_key_version > 0),
  verification_token_hash bytea CHECK (verification_token_hash IS NULL OR octet_length(verification_token_hash) = 32),
  verification_expires_at timestamptz,
  verification_consumed_at timestamptz,
  verification_delivery_id uuid UNIQUE,
  verification_delivery_ciphertext bytea CHECK (verification_delivery_ciphertext IS NULL OR octet_length(verification_delivery_ciphertext) BETWEEN 29 AND 4096),
  verification_delivery_key_version integer CHECK (verification_delivery_key_version IS NULL OR verification_delivery_key_version > 0),
  verified_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CHECK ((verification_delivery_id IS NULL) = (verification_delivery_ciphertext IS NULL)),
  CHECK ((verification_delivery_id IS NULL) = (verification_delivery_key_version IS NULL))
);

CREATE TABLE identity.password_credentials (
  principal_id uuid PRIMARY KEY REFERENCES identity.accounts(id),
  policy_version integer NOT NULL CHECK (policy_version > 0),
  memory_kib integer NOT NULL CHECK (memory_kib >= 65536),
  time_cost integer NOT NULL CHECK (time_cost >= 3),
  parallelism integer NOT NULL CHECK (parallelism >= 1 AND parallelism <= 16),
  salt bytea NOT NULL CHECK (octet_length(salt) = 16),
  password_hash bytea NOT NULL CHECK (octet_length(password_hash) = 32),
  reset_token_hash bytea CHECK (reset_token_hash IS NULL OR octet_length(reset_token_hash) = 32),
  reset_expires_at timestamptz,
  reset_consumed_at timestamptz,
  reset_delivery_id uuid UNIQUE,
  reset_delivery_ciphertext bytea CHECK (reset_delivery_ciphertext IS NULL OR octet_length(reset_delivery_ciphertext) BETWEEN 29 AND 4096),
  reset_delivery_key_version integer CHECK (reset_delivery_key_version IS NULL OR reset_delivery_key_version > 0),
  updated_at timestamptz NOT NULL,
  CHECK ((reset_delivery_id IS NULL) = (reset_delivery_ciphertext IS NULL)),
  CHECK ((reset_delivery_id IS NULL) = (reset_delivery_key_version IS NULL))
);

CREATE TABLE identity.account_sessions (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  state text NOT NULL CHECK (state IN ('active','review_required','revoked','compromised')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  client_signing_public_key bytea NOT NULL CHECK (octet_length(client_signing_public_key) = 32),
  access_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(access_token_hash) = 32),
  access_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE identity.account_refresh_tokens (
  token_hash bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  session_id uuid NOT NULL REFERENCES identity.account_sessions(id),
  previous_token_hash bytea CHECK (previous_token_hash IS NULL OR octet_length(previous_token_hash) = 32),
  state text NOT NULL CHECK (state IN ('active','used','revoked')),
  issued_at timestamptz NOT NULL,
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  used_at timestamptz,
  revoked_at timestamptz
);
CREATE INDEX account_refresh_tokens_session_idx ON identity.account_refresh_tokens(session_id);

CREATE TABLE identity.passkey_credentials (
  credential_id bytea PRIMARY KEY CHECK (octet_length(credential_id) BETWEEN 16 AND 1024),
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  public_key bytea NOT NULL CHECK (octet_length(public_key) BETWEEN 32 AND 4096),
  attestation_format text NOT NULL CHECK (length(attestation_format) BETWEEN 1 AND 64),
  transports text[] NOT NULL DEFAULT '{}',
  protocol_flags smallint NOT NULL CHECK (protocol_flags BETWEEN 0 AND 255),
  sign_count bigint NOT NULL CHECK (sign_count >= 0),
  state text NOT NULL CHECK (state IN ('active','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  revoked_at timestamptz
);
CREATE INDEX passkey_credentials_principal_idx ON identity.passkey_credentials(principal_id);

CREATE TABLE identity.totp_credentials (
  principal_id uuid PRIMARY KEY REFERENCES identity.accounts(id),
  ciphertext bytea NOT NULL CHECK (octet_length(ciphertext) BETWEEN 29 AND 1024),
  encryption_key_version integer NOT NULL CHECK (encryption_key_version > 0),
  state text NOT NULL CHECK (state IN ('pending','active','revoked')),
  last_accepted_step bigint,
  created_at timestamptz NOT NULL,
  verified_at timestamptz,
  revoked_at timestamptz
);

CREATE TABLE identity.recovery_code_sets (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  generation integer NOT NULL CHECK (generation > 0),
  code_hashes bytea[] NOT NULL CHECK (cardinality(code_hashes) BETWEEN 1 AND 10),
  state text NOT NULL CHECK (state IN ('active','superseded','exhausted','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (principal_id, generation)
);

CREATE TABLE identity.security_events (
  id uuid PRIMARY KEY,
  principal_id uuid REFERENCES identity.accounts(id),
  category text NOT NULL CHECK (category ~ '^[a-z][a-z0-9_]{0,63}$'),
  fingerprint text NOT NULL CHECK (fingerprint ~ '^[a-z][a-z0-9_.-]{0,127}$'),
  aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
  occurred_at timestamptz NOT NULL
);
CREATE INDEX security_events_principal_time_idx ON identity.security_events(principal_id, occurred_at DESC);

-- +goose Down
DROP SCHEMA identity CASCADE;
```

- [ ] **Step 4: Add named queries with atomic transitions**

`db/queries/identity.sql` must define these names and semantics:

```sql
-- name: CreateAccount :exec
INSERT INTO identity.accounts (id, state, state_version, locale, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAccountForUpdate :one
SELECT * FROM identity.accounts WHERE id = $1 FOR UPDATE;

-- name: FindIdentityByLookupDigest :one
SELECT * FROM identity.email_identities WHERE lookup_digest = $1;

-- name: ActivateVerifiedAccount :one
UPDATE identity.accounts
SET state = 'active', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state = 'pending_email'
RETURNING *;

-- name: ConsumeEmailVerification :one
UPDATE identity.email_identities
SET verification_consumed_at = $3, verified_at = $3, updated_at = $3
WHERE principal_id = $1
  AND verification_token_hash = $2
  AND verification_consumed_at IS NULL
  AND verification_expires_at >= $3
RETURNING *;

-- name: GetPasswordCredential :one
SELECT * FROM identity.password_credentials WHERE principal_id = $1;

-- name: CreateAccountSession :exec
INSERT INTO identity.account_sessions
  (id, principal_id, state, state_version, client_signing_public_key, access_token_hash,
   access_expires_at, absolute_expires_at, created_at, updated_at)
VALUES ($1, $2, 'active', 1, $3, $4, $5, $6, $7, $7);

-- name: GetRefreshTokenForUpdate :one
SELECT r.token_hash, r.session_id, r.previous_token_hash, r.state AS refresh_state,
       r.issued_at, r.idle_expires_at, r.absolute_expires_at, r.used_at, r.revoked_at,
       s.state AS session_state, s.state_version AS session_state_version,
       s.client_signing_public_key, s.principal_id
FROM identity.account_refresh_tokens r
JOIN identity.account_sessions s ON s.id=r.session_id
WHERE r.token_hash = $1 FOR UPDATE OF r, s;

-- name: MarkAccountRefreshUsed :one
UPDATE identity.account_refresh_tokens
SET state = 'used', used_at = $2
WHERE token_hash = $1 AND state = 'active'
RETURNING *;

-- name: RevokeAccountSession :execrows
UPDATE identity.account_sessions
SET state = 'revoked', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state IN ('active','review_required');

-- name: AcceptTOTPStep :execrows
UPDATE identity.totp_credentials
SET last_accepted_step = $2
WHERE principal_id = $1 AND state = 'active'
  AND (last_accepted_step IS NULL OR last_accepted_step < $2);
```

Complete the same file with these named queries; there are no handwritten identity SQL calls outside sqlc:

```sql
-- name: CreateEmailIdentity :exec
INSERT INTO identity.email_identities
  (id, principal_id, lookup_key_version, lookup_digest, ciphertext, encryption_key_version,
   verification_token_hash, verification_expires_at, verification_delivery_id,
   verification_delivery_ciphertext, verification_delivery_key_version, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12);

-- name: ResetEmailVerification :one
UPDATE identity.email_identities
SET verification_token_hash = $2, verification_expires_at = $3,
    verification_consumed_at = NULL, verification_delivery_id=$4,
    verification_delivery_ciphertext=$5, verification_delivery_key_version=$6, updated_at=$7
WHERE principal_id = $1 AND verified_at IS NULL
RETURNING *;

-- name: GetPendingEmailDelivery :one
SELECT verification_delivery_id, verification_delivery_ciphertext,
       verification_delivery_key_version, verification_expires_at
FROM identity.email_identities
WHERE verification_delivery_id=$1 AND verification_consumed_at IS NULL;

-- name: ClearPendingEmailDelivery :execrows
UPDATE identity.email_identities
SET verification_delivery_id=NULL, verification_delivery_ciphertext=NULL,
    verification_delivery_key_version=NULL, updated_at=$2
WHERE verification_delivery_id=$1;

-- name: CreatePasswordCredential :exec
INSERT INTO identity.password_credentials
  (principal_id, policy_version, memory_kib, time_cost, parallelism, salt, password_hash, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);

-- name: UpdatePasswordCredential :execrows
UPDATE identity.password_credentials
SET policy_version=$2, memory_kib=$3, time_cost=$4, parallelism=$5,
    salt=$6, password_hash=$7, updated_at=$8
WHERE principal_id=$1;

-- name: SetPasswordReset :one
UPDATE identity.password_credentials
SET reset_token_hash=$2, reset_expires_at=$3, reset_consumed_at=NULL,
    reset_delivery_id=$4, reset_delivery_ciphertext=$5,
    reset_delivery_key_version=$6, updated_at=$7
WHERE principal_id=$1
RETURNING *;

-- name: GetPendingPasswordResetDelivery :one
SELECT reset_delivery_id, reset_delivery_ciphertext, reset_delivery_key_version, reset_expires_at
FROM identity.password_credentials
WHERE reset_delivery_id=$1 AND reset_consumed_at IS NULL;

-- name: ConsumePasswordReset :one
UPDATE identity.password_credentials
SET policy_version=$3, memory_kib=$4, time_cost=$5, parallelism=$6,
    salt=$7, password_hash=$8, reset_consumed_at=$9,
    reset_delivery_id=NULL, reset_delivery_ciphertext=NULL,
    reset_delivery_key_version=NULL, updated_at=$9
WHERE principal_id=$1 AND reset_token_hash=$2
  AND reset_consumed_at IS NULL AND reset_expires_at >= $9
RETURNING *;

-- name: FindAccountAccessToken :one
SELECT s.id, s.principal_id, s.state, s.state_version, s.access_expires_at,
       s.absolute_expires_at, a.state AS account_state
FROM identity.account_sessions s
JOIN identity.accounts a ON a.id=s.principal_id
WHERE s.access_token_hash=$1;

-- name: InsertAccountRefreshToken :exec
INSERT INTO identity.account_refresh_tokens
  (token_hash, session_id, previous_token_hash, state, issued_at, idle_expires_at, absolute_expires_at)
VALUES ($1,$2,$3,'active',$4,$5,$6);

-- name: RotateAccountSessionAccess :one
UPDATE identity.account_sessions
SET access_token_hash=$2, access_expires_at=$3, updated_at=$4,
    state_version=state_version+1
WHERE id=$1 AND state='active' AND absolute_expires_at >= $4
RETURNING *;

-- name: RevokeAccountRefreshTokens :execrows
UPDATE identity.account_refresh_tokens
SET state='revoked', revoked_at=$2
WHERE session_id=$1 AND state='active';

-- name: MarkAccountSessionCompromised :execrows
UPDATE identity.account_sessions
SET state='compromised', state_version=state_version+1, updated_at=$2
WHERE id=$1 AND state IN ('active','review_required');

-- name: MarkPrincipalSessionsReviewRequired :execrows
UPDATE identity.account_sessions
SET state='review_required', state_version=state_version+1, updated_at=$2
WHERE principal_id=$1 AND state='active';

-- name: CreatePasskeyCredential :exec
INSERT INTO identity.passkey_credentials
  (credential_id, principal_id, public_key, attestation_format, transports,
   protocol_flags, sign_count, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8);

-- name: ListActivePasskeys :many
SELECT * FROM identity.passkey_credentials
WHERE principal_id=$1 AND state='active'
ORDER BY created_at LIMIT 10;

-- name: UpdatePasskeyCounter :execrows
UPDATE identity.passkey_credentials
SET sign_count=$3, protocol_flags=$4, updated_at=$5
WHERE credential_id=$1 AND principal_id=$2 AND state='active' AND sign_count <= $3;

-- name: RevokePasskey :execrows
UPDATE identity.passkey_credentials
SET state='revoked', revoked_at=$3, updated_at=$3
WHERE credential_id=$1 AND principal_id=$2 AND state='active';

-- name: CreateTOTPEnrollment :exec
INSERT INTO identity.totp_credentials
  (principal_id, ciphertext, encryption_key_version, state, created_at)
VALUES ($1,$2,$3,'pending',$4);

-- name: GetTOTPForUpdate :one
SELECT * FROM identity.totp_credentials WHERE principal_id=$1 FOR UPDATE;

-- name: ActivateTOTP :execrows
UPDATE identity.totp_credentials
SET state='active', last_accepted_step=$2, verified_at=$3
WHERE principal_id=$1 AND state='pending';

-- name: RevokeTOTP :execrows
UPDATE identity.totp_credentials
SET state='revoked', revoked_at=$2
WHERE principal_id=$1 AND state IN ('pending','active');

-- name: CreateRecoveryCodeSet :exec
INSERT INTO identity.recovery_code_sets
  (id, principal_id, generation, code_hashes, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,'active',$5,$5);

-- name: GetActiveRecoveryCodeSetForUpdate :one
SELECT * FROM identity.recovery_code_sets
WHERE principal_id=$1 AND state='active' FOR UPDATE;

-- name: ConsumeRecoveryCode :one
UPDATE identity.recovery_code_sets
SET code_hashes=array_remove(code_hashes, $2::bytea),
    state=CASE WHEN cardinality(array_remove(code_hashes, $2::bytea))=0 THEN 'exhausted' ELSE state END,
    updated_at=$3
WHERE id=$1 AND state='active' AND $2::bytea=ANY(code_hashes)
RETURNING *;

-- name: RevokeRecoveryCodeSets :execrows
UPDATE identity.recovery_code_sets
SET state=$2, updated_at=$3
WHERE principal_id=$1 AND state='active' AND $2 IN ('superseded','revoked');

-- name: InsertSecurityEvent :exec
INSERT INTO identity.security_events
  (id, principal_id, category, fingerprint, aggregate_version, occurred_at)
VALUES ($1,$2,$3,$4,$5,$6);
```

The application validates every recovery digest is exactly 32 bytes before insertion; `ConsumeRecoveryCode` is the only path that removes a digest and therefore provides atomic single use.

- [ ] **Step 5: Generate, test reversal, and commit**

Run: `go tool sqlc generate`

Run: `go test -tags=integration ./internal/store -run TestIdentity -count=1`

Run exactly against the dev database: `go tool goose -dir db/migrations postgres "$env:TALENRO_DATABASE_URL" up`, `down-to 1`, then `up`.

Expected: sqlc succeeds, integration test PASS, and `00002_identity.sql` reverses/reapplies without residue.

```bash
git add sqlc.yaml db/migrations/00002_identity.sql db/queries/identity.sql internal/store
git commit -m "feat: add authoritative identity persistence"
```

### Task 4: Add reversible device identity persistence

**Files:**
- Create: `db/migrations/00003_deviceauth.sql`
- Create: `db/queries/deviceauth.sql`
- Create: `internal/store/deviceauth_integration_test.go`
- Modify generated files under: `internal/store/`

**Interfaces:**
- Consumes: principal UUID from Task 3; application-generated device/family UUIDs; 32-byte Ed25519/X25519 public keys.
- Produces: sqlc methods for grant consumption, device creation, authorization transitions, token-family compromise and provisional policy snapshots.

- [ ] **Step 1: Write the failing concurrent-grant persistence test**

The integration test creates one grant, runs two transactions calling `ConsumeEnrollmentGrant`, and asserts exactly one transaction returns a row while the other returns `pgx.ErrNoRows`; then it asserts one device row exists for `consumed_device_id`.

```go
func consumeGrant(ctx context.Context, pool *pgxpool.Pool, tokenHash []byte, deviceID uuid.UUID, now time.Time) error {
	tx, err := pool.Begin(ctx)
	if err != nil { return err }
	defer func() { _ = tx.Rollback(ctx) }()
	q := store.New(tx)
	if err = q.CreateDevice(ctx, store.CreateDeviceParams{
		ID: deviceID, PrincipalID: testPrincipalID, SigningPublicKey: bytes.Repeat([]byte{1}, 32),
		HpkePublicKey: bytes.Repeat([]byte{2}, 32), KeyVersion: 1, CreatedAt: now, UpdatedAt: now,
	}); err != nil { return err }
	if _, err = q.ConsumeEnrollmentGrant(ctx, store.ConsumeEnrollmentGrantParams{
		TokenHash: tokenHash, ConsumedDeviceID: uuid.NullUUID{UUID: deviceID, Valid: true}, ConsumedAt: now,
	}); err != nil { return err }
	return tx.Commit(ctx)
}
```

`testPrincipalID` is the account created by the fixture. Each competing transaction inserts a distinct device before consuming the grant; the losing transaction must roll back that insert, proving the foreign-key-safe transaction order and the one-device invariant together.

- [ ] **Step 2: Run generation/test and verify RED**

Run: `go tool sqlc generate`

Expected: FAIL because deviceauth schema and queries are absent.

- [ ] **Step 3: Create the exact deviceauth schema**

`00003_deviceauth.sql` creates schema `deviceauth` and these tables:

```sql
-- +goose Up
CREATE SCHEMA deviceauth;

CREATE TABLE deviceauth.enrollment_grants (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  account_session_id uuid NOT NULL REFERENCES identity.account_sessions(id),
  token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
  policy_marker text NOT NULL CHECK (policy_marker IN ('standard','trial_restricted')),
  provisional_until timestamptz,
  state text NOT NULL CHECK (state IN ('unused','consumed','expired')),
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  consumed_at timestamptz,
  consumed_device_id uuid,
  CHECK ((policy_marker='trial_restricted') = (provisional_until IS NOT NULL))
);

CREATE TABLE deviceauth.devices (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  display_name_ciphertext bytea CHECK (display_name_ciphertext IS NULL OR octet_length(display_name_ciphertext) BETWEEN 29 AND 1024),
  display_name_key_version integer,
  signing_public_key bytea NOT NULL CHECK (octet_length(signing_public_key) = 32),
  hpke_public_key bytea NOT NULL CHECK (octet_length(hpke_public_key) = 32),
  key_version integer NOT NULL CHECK (key_version > 0),
  state text NOT NULL CHECK (state IN ('active','suspended','revoked')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (principal_id, signing_public_key),
  UNIQUE (principal_id, hpke_public_key)
);
ALTER TABLE deviceauth.enrollment_grants
  ADD CONSTRAINT enrollment_grants_consumed_device_fk
  FOREIGN KEY (consumed_device_id) REFERENCES deviceauth.devices(id);

CREATE TABLE deviceauth.device_authorizations (
  id uuid PRIMARY KEY,
  principal_id uuid NOT NULL REFERENCES identity.accounts(id),
  device_id uuid NOT NULL UNIQUE REFERENCES deviceauth.devices(id),
  state text NOT NULL CHECK (state IN ('provisional','active','suspended','revoked')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  provisional_until timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX one_provisional_authorization_per_principal
  ON deviceauth.device_authorizations(principal_id) WHERE state='provisional';

ALTER TABLE identity.account_sessions
  ADD COLUMN device_authorization_id uuid REFERENCES deviceauth.device_authorizations(id);
CREATE INDEX account_sessions_device_authorization_idx
  ON identity.account_sessions(device_authorization_id);

CREATE TABLE deviceauth.device_token_families (
  id uuid PRIMARY KEY,
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  state text NOT NULL CHECK (state IN ('active','compromised','revoked')),
  state_version bigint NOT NULL CHECK (state_version > 0),
  access_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(access_token_hash) = 32),
  access_expires_at timestamptz NOT NULL,
  idle_expires_at timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE deviceauth.device_refresh_tokens (
  token_hash bytea PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  family_id uuid NOT NULL REFERENCES deviceauth.device_token_families(id),
  previous_token_hash bytea CHECK (previous_token_hash IS NULL OR octet_length(previous_token_hash) = 32),
  state text NOT NULL CHECK (state IN ('active','used','revoked')),
  issued_at timestamptz NOT NULL,
  used_at timestamptz,
  revoked_at timestamptz
);
CREATE INDEX device_refresh_tokens_family_idx ON deviceauth.device_refresh_tokens(family_id);

CREATE TABLE deviceauth.device_policy_snapshots (
  authorization_id uuid PRIMARY KEY REFERENCES deviceauth.device_authorizations(id),
  schema_version text NOT NULL CHECK (schema_version = 'device-policy-v1'),
  policy jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  CHECK (octet_length(policy::text) <= 4096)
);

-- +goose Down
ALTER TABLE identity.account_sessions DROP COLUMN device_authorization_id;
DROP SCHEMA deviceauth CASCADE;
```

- [ ] **Step 4: Add exact atomic queries**

```sql
-- name: ConsumeEnrollmentGrant :one
UPDATE deviceauth.enrollment_grants
SET state = 'consumed', consumed_at = $3, consumed_device_id = $2
WHERE token_hash = $1 AND state = 'unused' AND expires_at >= $3
RETURNING *;

-- name: GetGrantForChallenge :one
SELECT id, principal_id, account_session_id, token_hash, policy_marker, provisional_until, expires_at
FROM deviceauth.enrollment_grants
WHERE token_hash = $1 AND state = 'unused' AND expires_at >= $2;

-- name: CreateDevice :exec
INSERT INTO deviceauth.devices
  (id, principal_id, display_name_ciphertext, display_name_key_version, signing_public_key, hpke_public_key, key_version, state, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8);

-- name: GetAuthorizationForUpdate :one
SELECT * FROM deviceauth.device_authorizations WHERE id = $1 FOR UPDATE;

-- name: FindDeviceAccessToken :one
SELECT f.id AS family_id, f.authorization_id, f.state AS family_state,
       f.state_version AS family_state_version, f.access_expires_at,
       f.idle_expires_at, f.absolute_expires_at,
       a.principal_id, a.device_id, a.state AS authorization_state,
       a.state_version AS authorization_state_version,
       ac.state AS account_state
FROM deviceauth.device_token_families f
JOIN deviceauth.device_authorizations a ON a.id = f.authorization_id
JOIN identity.accounts ac ON ac.id = a.principal_id
WHERE f.access_token_hash = $1;

-- name: GetDeviceRefreshForUpdate :one
SELECT r.token_hash, r.family_id, r.previous_token_hash, r.state AS refresh_state,
       r.issued_at, r.used_at, r.revoked_at,
       f.authorization_id, f.state AS family_state, f.state_version AS family_state_version,
       f.idle_expires_at, f.absolute_expires_at
FROM deviceauth.device_refresh_tokens r
JOIN deviceauth.device_token_families f ON f.id = r.family_id
WHERE r.token_hash = $1 FOR UPDATE OF r, f;

-- name: CompromiseDeviceTokenFamily :one
UPDATE deviceauth.device_token_families
SET state = 'compromised', state_version = state_version + 1, updated_at = $2
WHERE id = $1 AND state = 'active'
RETURNING *;
```

Complete `deviceauth.sql` with the following exact query set:

```sql
-- name: CreateEnrollmentGrant :exec
INSERT INTO deviceauth.enrollment_grants
  (id, principal_id, account_session_id, token_hash, policy_marker, provisional_until, state, expires_at, created_at)
VALUES ($1,$2,$3,$4,$5,$6,'unused',$7,$8);

-- name: CreateDeviceAuthorization :exec
INSERT INTO deviceauth.device_authorizations
  (id, principal_id, device_id, state, state_version, provisional_until, created_at, updated_at)
VALUES ($1,$2,$3,$4,1,$5,$6,$6);

-- name: CreateDeviceTokenFamily :exec
INSERT INTO deviceauth.device_token_families
  (id, authorization_id, state, state_version, access_token_hash, access_expires_at,
   idle_expires_at, absolute_expires_at, created_at, updated_at)
VALUES ($1,$2,'active',1,$3,$4,$5,$6,$7,$7);

-- name: InsertDeviceRefreshToken :exec
INSERT INTO deviceauth.device_refresh_tokens
  (token_hash, family_id, previous_token_hash, state, issued_at)
VALUES ($1,$2,$3,'active',$4);

-- name: CreateDevicePolicySnapshot :exec
INSERT INTO deviceauth.device_policy_snapshots
  (authorization_id, schema_version, policy, created_at)
VALUES ($1,'device-policy-v1',$2,$3);

-- name: BindAccountSessionToAuthorization :execrows
UPDATE identity.account_sessions
SET device_authorization_id=$2, state_version=state_version+1, updated_at=$3
WHERE id=$1 AND state='active' AND device_authorization_id IS NULL;

-- name: RevokeDeviceBoundAccountSessions :execrows
UPDATE identity.account_sessions
SET state='revoked', state_version=state_version+1, updated_at=$2
WHERE device_authorization_id=$1 AND state IN ('active','review_required');

-- name: MarkDeviceRefreshUsed :one
UPDATE deviceauth.device_refresh_tokens
SET state='used', used_at=$2
WHERE token_hash=$1 AND state='active'
RETURNING *;

-- name: RotateDeviceFamilyAccess :one
UPDATE deviceauth.device_token_families
SET access_token_hash=$2, access_expires_at=$3, idle_expires_at=$4,
    state_version=state_version+1, updated_at=$5
WHERE id=$1 AND state='active' AND idle_expires_at >= $5 AND absolute_expires_at >= $5
RETURNING *;

-- name: RevokeDeviceRefreshTokens :execrows
UPDATE deviceauth.device_refresh_tokens
SET state='revoked', revoked_at=$2
WHERE family_id=$1 AND state='active';

-- name: ActivateProvisionalAuthorization :execrows
UPDATE deviceauth.device_authorizations
SET state='active', state_version=state_version+1, provisional_until=NULL, updated_at=$2
WHERE principal_id=$1 AND state='provisional' AND provisional_until >= $2;

-- name: RevokeDeviceAuthorization :execrows
UPDATE deviceauth.device_authorizations
SET state='revoked', state_version=state_version+1, updated_at=$2
WHERE id=$1 AND state IN ('provisional','active','suspended');

-- name: RevokeDeviceRecord :execrows
UPDATE deviceauth.devices
SET state='revoked', updated_at=$2
WHERE id=$1 AND state IN ('active','suspended');

-- name: GetActiveBundleAuthority :one
SELECT a.id AS authorization_id, a.principal_id, a.device_id,
       a.state AS authorization_state, a.state_version,
       d.hpke_public_key, d.key_version, p.schema_version, p.policy,
       ac.state AS account_state
FROM deviceauth.device_authorizations a
JOIN deviceauth.devices d ON d.id=a.device_id
JOIN deviceauth.device_policy_snapshots p ON p.authorization_id=a.id
JOIN identity.accounts ac ON ac.id=a.principal_id
WHERE a.id=$1 AND d.state='active'
  AND (a.state='active' OR (a.state='provisional' AND a.provisional_until >= $2));
```

Callers pass the transaction clock as `$2`. The adapter then requires account `active`, or `pending_email` only when the returned immutable policy marker is `trial_restricted`.

All grant/device/authorization/family/refresh/snapshot creation runs in one pgx transaction owned by the repository adapter in this exact order: insert device, consume grant, insert authorization, bind the grant's account session through `identity.DeviceTransactionParticipant`, insert policy snapshot, family and refresh token. The service validates the grant principal/session before any insert; every later error rolls the entire transaction back.

- [ ] **Step 5: Generate, verify reversal/concurrency, and commit**

Run: `go tool sqlc generate`

Run: `go test -tags=integration ./internal/store -run TestDevice -count=1`

Run migrations `up`, `down-to 2`, `up` and rerun the test.

Expected: exactly one concurrent grant consume succeeds; migration reversal is clean.

```bash
git add db/migrations/00003_deviceauth.sql db/queries/deviceauth.sql internal/store
git commit -m "feat: add authoritative device identity persistence"
```

### Task 5: Add trust, idempotency, and transactional-outbox persistence

**Files:**
- Create: `db/migrations/00004_trust_delivery.sql`
- Create: `db/queries/trust.sql`
- Create: `db/queries/idempotency.sql`
- Create: `db/queries/outbox.sql`
- Create: `internal/store/trust_delivery_integration_test.go`
- Modify generated files under: `internal/store/`

**Interfaces:**
- Consumes: active device authorization UUID and application-generated IDs.
- Produces: monotonic issuance, locator hash lookup, immutable envelope storage, ack, idempotency replay, outbox claiming and consumer dedupe.

- [ ] **Step 1: Write failing monotonic/idempotency integration tests**

Start from no version row and call `NextBundleVersion` concurrently twice. The returned set must be exactly `{1,2}` and the stored value must be `2`; a third call must return `3`. For idempotency, race two `TryBeginIdempotency` inserts for the same scope/operation/key: exactly one returns one affected row. Lock the winner with `GetIdempotencyForUpdate`, store a response, and assert the same request digest replays that response while a different digest is classified as conflict by the repository adapter.

- [ ] **Step 2: Run sqlc and verify RED**

Run: `go tool sqlc generate`

Expected: FAIL because trust/shared schema and queries are absent.

- [ ] **Step 3: Create the exact migration**

```sql
-- +goose Up
CREATE SCHEMA trust;

CREATE TABLE trust.trust_root_metadata (
  version bigint PRIMARY KEY CHECK (version > 0),
  canonical_payload bytea NOT NULL CHECK (octet_length(canonical_payload) BETWEEN 2 AND 65536),
  signature bytea NOT NULL CHECK (octet_length(signature) = 64),
  valid_from timestamptz NOT NULL,
  valid_until timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  CHECK (valid_until > valid_from)
);

CREATE TABLE trust.signing_key_metadata (
  key_id text PRIMARY KEY CHECK (key_id ~ '^[A-Za-z0-9_-]{16,64}$'),
  root_metadata_version bigint NOT NULL REFERENCES trust.trust_root_metadata(version),
  algorithm text NOT NULL CHECK (algorithm = 'Ed25519'),
  public_key bytea NOT NULL CHECK (octet_length(public_key) = 32),
  state text NOT NULL CHECK (state IN ('future','active','retiring','revoked')),
  not_before timestamptz NOT NULL,
  not_after timestamptz NOT NULL,
  CHECK (not_after > not_before)
);

CREATE TABLE trust.highest_bundle_versions (
  authorization_id uuid PRIMARY KEY REFERENCES deviceauth.device_authorizations(id),
  highest_issued bigint NOT NULL CHECK (highest_issued > 0),
  updated_at timestamptz NOT NULL
);

CREATE TABLE trust.bundle_issuances (
  id uuid PRIMARY KEY,
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  bundle_version bigint NOT NULL CHECK (bundle_version > 0),
  locator_hash bytea NOT NULL UNIQUE CHECK (octet_length(locator_hash) = 32),
  locator_ciphertext bytea NOT NULL CHECK (octet_length(locator_ciphertext) BETWEEN 29 AND 512),
  locator_key_version integer NOT NULL CHECK (locator_key_version > 0),
  envelope bytea NOT NULL CHECK (octet_length(envelope) BETWEEN 64 AND 1048576),
  envelope_sha256 bytea NOT NULL CHECK (octet_length(envelope_sha256) = 32),
  signer_key_id text NOT NULL REFERENCES trust.signing_key_metadata(key_id),
  issued_at timestamptz NOT NULL,
  not_before timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  UNIQUE (authorization_id, bundle_version),
  CHECK (expires_at > not_before)
);

CREATE TABLE trust.bundle_acknowledgements (
  bundle_id uuid NOT NULL REFERENCES trust.bundle_issuances(id),
  authorization_id uuid NOT NULL REFERENCES deviceauth.device_authorizations(id),
  bundle_version bigint NOT NULL CHECK (bundle_version > 0),
  acknowledged_at timestamptz NOT NULL,
  PRIMARY KEY (bundle_id, authorization_id)
);

CREATE TABLE idempotency_records (
  principal_scope text NOT NULL CHECK (principal_scope ~ '^[A-Za-z0-9:_-]{1,128}$'),
  operation text NOT NULL CHECK (operation ~ '^[a-z][a-z0-9_]{0,63}$'),
  idempotency_key_hash bytea NOT NULL CHECK (octet_length(idempotency_key_hash) = 32),
  request_digest bytea NOT NULL CHECK (octet_length(request_digest) = 32),
  state text NOT NULL CHECK (state IN ('in_progress','completed','failed')),
  response_status integer,
  response_ciphertext bytea CHECK (response_ciphertext IS NULL OR octet_length(response_ciphertext) <= 1048608),
  response_key_version integer CHECK (response_key_version IS NULL OR response_key_version > 0),
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  CHECK ((response_ciphertext IS NULL) = (response_key_version IS NULL)),
  PRIMARY KEY (principal_scope, operation, idempotency_key_hash)
);

CREATE TABLE transactional_outbox (
  event_id uuid PRIMARY KEY,
  event_type text NOT NULL CHECK (event_type ~ '^talenro[.][a-z0-9_.-]{1,127}$'),
  aggregate_type text NOT NULL CHECK (aggregate_type ~ '^[a-z][a-z0-9_]{0,63}$'),
  aggregate_id uuid NOT NULL,
  aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
  idempotency_key text NOT NULL CHECK (idempotency_key ~ '^[A-Za-z0-9:_-]{1,128}$'),
  payload bytea NOT NULL CHECK (octet_length(payload) BETWEEN 1 AND 262144),
  occurred_at timestamptz NOT NULL,
  available_at timestamptz NOT NULL,
  claimed_until timestamptz,
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  published_at timestamptz
);
CREATE INDEX outbox_available_idx ON transactional_outbox(available_at, occurred_at) WHERE published_at IS NULL;

CREATE TABLE consumed_event_ids (
  consumer text NOT NULL CHECK (consumer ~ '^[a-z][a-z0-9_.-]{0,127}$'),
  event_id uuid NOT NULL,
  consumed_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (consumer, event_id)
);

-- +goose Down
DROP TABLE consumed_event_ids;
DROP TABLE transactional_outbox;
DROP TABLE idempotency_records;
DROP SCHEMA trust CASCADE;
```

- [ ] **Step 4: Add exact monotonic and queue queries**

```sql
-- name: NextBundleVersion :one
INSERT INTO trust.highest_bundle_versions (authorization_id, highest_issued, updated_at)
VALUES ($1, 1, $2)
ON CONFLICT (authorization_id) DO UPDATE
SET highest_issued = trust.highest_bundle_versions.highest_issued + 1,
    updated_at = EXCLUDED.updated_at
RETURNING highest_issued;

-- name: GetHighestBundleVersion :one
SELECT highest_issued FROM trust.highest_bundle_versions WHERE authorization_id=$1;

-- name: GetBundleByLocatorHash :one
SELECT * FROM trust.bundle_issuances
WHERE locator_hash = $1 AND expires_at >= $2;

-- name: InsertBundleAcknowledgement :exec
INSERT INTO trust.bundle_acknowledgements
  (bundle_id, authorization_id, bundle_version, acknowledged_at)
VALUES ($1,$2,$3,$4)
ON CONFLICT (bundle_id, authorization_id) DO NOTHING;

-- name: TryBeginIdempotency :execrows
INSERT INTO idempotency_records
  (principal_scope, operation, idempotency_key_hash, request_digest, state, created_at, expires_at)
VALUES ($1,$2,$3,$4,'in_progress',$5,$6)
ON CONFLICT (principal_scope, operation, idempotency_key_hash) DO NOTHING;

-- name: GetIdempotencyForUpdate :one
SELECT * FROM idempotency_records
WHERE principal_scope=$1 AND operation=$2 AND idempotency_key_hash=$3
FOR UPDATE;

-- name: CompleteIdempotency :one
UPDATE idempotency_records
SET state = 'completed', response_status = $5, response_ciphertext = $6, response_key_version=$7
WHERE principal_scope = $1 AND operation = $2 AND idempotency_key_hash = $3
  AND request_digest=$4 AND state = 'in_progress'
RETURNING *;

-- name: ClaimOutboxBatch :many
WITH candidates AS (
  SELECT event_id FROM transactional_outbox
  WHERE published_at IS NULL AND available_at <= $1
    AND (claimed_until IS NULL OR claimed_until < $1)
  ORDER BY occurred_at
  LIMIT $2
  FOR UPDATE SKIP LOCKED
)
UPDATE transactional_outbox o
SET claimed_until = $3, attempts = attempts + 1
FROM candidates c
WHERE o.event_id = c.event_id
RETURNING o.*;

-- name: MarkOutboxPublished :execrows
UPDATE transactional_outbox SET published_at = $2, claimed_until = NULL
WHERE event_id = $1 AND published_at IS NULL;

-- name: RecordConsumedEvent :execrows
INSERT INTO consumed_event_ids (consumer, event_id, consumed_at, expires_at)
VALUES ($1,$2,$3,$4)
ON CONFLICT DO NOTHING;
```

Complete the files with this exact repository surface:

```sql
-- name: InsertTrustRootMetadata :exec
INSERT INTO trust.trust_root_metadata
  (version, canonical_payload, signature, valid_from, valid_until, created_at)
VALUES ($1,$2,$3,$4,$5,$6);

-- name: ListTrustRootMetadata :many
SELECT * FROM trust.trust_root_metadata ORDER BY version;

-- name: InsertSigningKeyMetadata :exec
INSERT INTO trust.signing_key_metadata
  (key_id, root_metadata_version, algorithm, public_key, state, not_before, not_after)
VALUES ($1,$2,'Ed25519',$3,$4,$5,$6);

-- name: ListSigningKeysForMetadata :many
SELECT * FROM trust.signing_key_metadata
WHERE root_metadata_version=$1 ORDER BY key_id;

-- name: InsertBundleIssuance :exec
INSERT INTO trust.bundle_issuances
  (id, authorization_id, bundle_version, locator_hash, locator_ciphertext,
   locator_key_version, envelope, envelope_sha256, signer_key_id,
   issued_at, not_before, expires_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);

-- name: GetBundleIssuance :one
SELECT * FROM trust.bundle_issuances WHERE id=$1;

-- name: InsertOutboxEvent :exec
INSERT INTO transactional_outbox
  (event_id,event_type,aggregate_type,aggregate_id,aggregate_version,
   idempotency_key,payload,occurred_at,available_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9);

-- name: ReleaseOutboxClaim :execrows
UPDATE transactional_outbox
SET claimed_until=NULL, available_at=$2
WHERE event_id=$1 AND published_at IS NULL;

-- name: GetOutboxHealth :one
SELECT count(*)::bigint AS backlog,
       COALESCE(EXTRACT(EPOCH FROM ($1-min(occurred_at))),0)::double precision AS oldest_age_seconds
FROM transactional_outbox WHERE published_at IS NULL;

-- name: PruneConsumedEvents :execrows
DELETE FROM consumed_event_ids WHERE expires_at < $1;

-- name: PruneIdempotencyRecords :execrows
DELETE FROM idempotency_records WHERE expires_at < $1 AND state IN ('completed','failed');
```

Claim timeout is exactly 30 seconds; the repository rejects a batch outside `1..100` before calling sqlc. Security-mutation idempotency callers choose the 90-day tombstone expiry, all other callers choose at least 24 hours. `GetIdempotencyForUpdate` is called inside the same transaction after a lost `TryBeginIdempotency` race; the adapter never polls an `in_progress` record outside a caller deadline.

- [ ] **Step 5: Generate, verify, and commit**

Run: `go tool sqlc generate`

Run: `go test -tags=integration ./internal/store -run 'TestTrust|TestIdempotency|TestOutbox' -count=1`

Run migrations `up`, `down-to 3`, `up` and rerun tests.

Expected: monotonic and idempotency invariants PASS; reversal leaves identity/device schemas intact.

```bash
git add db/migrations/00004_trust_delivery.sql db/queries/trust.sql db/queries/idempotency.sql db/queries/outbox.sql internal/store
git commit -m "feat: add trust delivery persistence"
```

### Task 6: Build strict input, stable-error, clock, random, and opaque-token primitives

**Files:**
- Create: `internal/strictjson/decode_test.go`
- Create: `internal/strictjson/decode.go`
- Create: `internal/apierrors/errors_test.go`
- Create: `internal/apierrors/errors.go`
- Create: `internal/securitykit/clock.go`
- Create: `internal/securitykit/random.go`
- Create: `internal/securitykit/token_test.go`
- Create: `internal/securitykit/token.go`
- Create: `internal/strictjson/fuzz_test.go`
- Create: `internal/ratelimit/limiter_test.go`
- Create: `internal/ratelimit/limiter.go`

**Interfaces:**
- Produces: `strictjson.Decode(io.Reader, int64, any) error`, finite `apierrors.Error`, `securitykit.Clock`, `securitykit.RandomSource`, domain-separated opaque token generation/digest, and a privacy-safe rate-limit contract.
- Consumes: no database or transport package; domain packages may depend on these primitives.

- [ ] **Step 1: Write failing parser, error, and token tests**

The table test must reject an unknown member, duplicate member at the root or nested level, invalid UTF-8, two JSON values, a body larger than 64 KiB, and nesting deeper than 16. It must accept one valid object exactly at the field limits. Add this public-error privacy test:

```go
func TestPublicErrorContainsOnlyFiniteFields(t *testing.T) {
	err := apierrors.New(apierrors.AuthenticationFailed, apierrors.Reauthenticate)
	body, marshalErr := json.Marshal(err.Public("trace-safe-000001"))
	if marshalErr != nil { t.Fatal(marshalErr) }
	for _, forbidden := range []string{"SECRET-CANARY", "cause", "message", "stack"} {
		if bytes.Contains(body, []byte(forbidden)) { t.Fatalf("disclosed %q", forbidden) }
	}
}
```

The token test uses a deterministic random source, asserts 32 random bytes encode to 43 unpadded base64url characters, and asserts account-access, account-refresh, device-access, device-refresh, enrollment-grant, email-verification and recovery-code domains produce different SHA-256 digests for identical raw bytes. Rate-limit tests require a finite operation enum, a 32-byte rotating subject digest, allowed/denied results only, and a backend error that never contains its key/value.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./internal/strictjson ./internal/apierrors ./internal/securitykit ./internal/ratelimit -count=1`

Expected: compilation fails because all three packages are absent.

- [ ] **Step 3: Implement the exact finite interfaces**

```go
package securitykit

type Clock interface { Now() time.Time }
type RandomSource interface { Read([]byte) (int, error) }

type TokenDomain string
const (
	AccountAccessToken TokenDomain = "talenro/account-access/v1"
	AccountRefreshToken TokenDomain = "talenro/account-refresh/v1"
	DeviceAccessToken TokenDomain = "talenro/device-access/v1"
	DeviceRefreshToken TokenDomain = "talenro/device-refresh/v1"
	EnrollmentGrantToken TokenDomain = "talenro/enrollment-grant/v1"
	EmailVerificationToken TokenDomain = "talenro/email-verification/v1"
	RecoveryCodeToken TokenDomain = "talenro/recovery-code/v1"
)

func NewOpaqueToken(random RandomSource) (secret.Bytes, error)
func EncodeOpaqueToken(secret.Bytes) string
func DigestToken(domain TokenDomain, raw secret.Bytes) [32]byte
func DecodeOpaqueToken(value string) (secret.Bytes, error)
```

```go
package ratelimit

type Operation string
const (
	Login Operation = "login"
	Delivery Operation = "delivery"
	Challenge Operation = "challenge"
)
type Limiter interface {
	Allow(context.Context, Operation, [32]byte, config.RateLimitPolicy) (bool, error)
}
```

Subject digests are HMAC-SHA256 over `"TALENRO-RATE-LIMIT-V1\x00" || operation || UTC-window-start || canonical-subject` using the sensitive lookup key; only the digest reaches Redis. The digest rotates each window and cannot become a cross-window user tracker.

`apierrors.Code` contains exactly the OpenAPI code enum; `Action` contains exactly the five public actions. `Error` stores only code/action/retry class and an internal category enum, implements `error` with a fixed value-free string, and never wraps provider/database/crypto errors. `Public(traceID)` is the only JSON-facing conversion.

`strictjson.Decode` first reads through `io.LimitReader(max+1)`, validates UTF-8, tokenizes once to detect duplicates/depth, then decodes with `DisallowUnknownFields`, and requires EOF. All returned errors are package sentinels whose text contains no input bytes.

- [ ] **Step 4: Add fuzz invariants and verify GREEN**

Seed the fuzz target with every invalid case and assert it never panics, never accepts trailing bytes, and never includes input bytes in an error. Run:

```powershell
go test ./internal/strictjson ./internal/apierrors ./internal/securitykit ./internal/ratelimit -count=1
go test ./internal/strictjson -run '^$' -fuzz FuzzDecode -fuzztime 10s
```

Expected: PASS with bounded memory and no secret/input disclosure.

- [ ] **Step 5: Commit**

```bash
git add internal/strictjson internal/apierrors internal/securitykit internal/ratelimit
git commit -m "feat: add bounded security primitives"
```

### Task 7: Implement sensitive-field protection and Argon2id credentials

**Files:**
- Create: `internal/sensitive/local_test.go`
- Create: `internal/sensitive/local.go`
- Create: `internal/identity/email_test.go`
- Create: `internal/identity/email.go`
- Create: `internal/identity/password_test.go`
- Create: `internal/identity/password.go`
- Create: `internal/identity/password_benchmark_test.go`
- Create: `testdata/crypto/rfc9106/argon2id-v1.json`

**Interfaces:**
- Implements the frozen `sensitive.Protector` interface with local AES-256-GCM/HMAC-SHA256 fixture keys.
- Produces `identity.CanonicalEmail`, `identity.PasswordPolicy`, `HashPassword`, and constant-work `VerifyPassword`.

- [ ] **Step 1: Write failing known-vector, isolation, and enumeration tests**

Tests cover: lowercasing only the domain; preserving local-part dots and `+tag`; rejecting invalid UTF-8/NUL/over-254-byte values; distinct HMAC domains; AEAD domain substitution failure; key-version mismatch; tamper failure; defensive copies; the RFC 9106-derived checked-in vector; and equal Argon2 invocation count for a real credential and the dummy credential.

```go
func TestVerifyPasswordUsesOneArgon2InvocationForUnknownIdentity(t *testing.T) {
	calls := 0
	derive := func(password, salt []byte, p PasswordPolicy) []byte {
		calls++
		return argon2.IDKey(password, salt, p.Time, p.MemoryKiB, p.Parallelism, p.TagBytes)
	}
	_ = VerifyPasswordWithDeriver([]byte("wrong"), DummyCredential(), CurrentPasswordPolicy(), derive)
	if calls != 1 { t.Fatalf("argon2 calls=%d", calls) }
}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/sensitive ./internal/identity -run 'Test(Local|Email|Password|Verify)' -count=1`

Expected: compilation fails because the protector and credentials are absent.

- [ ] **Step 3: Implement exact crypto policies**

```go
type PasswordPolicy struct {
	Version uint32
	MemoryKiB uint32
	Time uint32
	Parallelism uint8
	SaltBytes uint32
	TagBytes uint32
}

func CurrentPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{Version:1, MemoryKiB:65536, Time:3, Parallelism:4, SaltBytes:16, TagBytes:32}
}

func CanonicalizeEmail(string) (CanonicalEmail, error)
func HashPassword(RandomSource, []byte, PasswordPolicy) (PasswordCredential, error)
func VerifyPassword([]byte, PasswordCredential, PasswordPolicy) (match, needsUpgrade bool)
func DummyCredential() PasswordCredential
```

Passwords are accepted only as 12–1024 bytes of valid UTF-8, copied into local byte slices, and zeroed after derivation where practical. Comparison uses `subtle.ConstantTimeCompare`. `DummyCredential` is deterministic process data with the current cost and is invoked on unknown email, suspended account, malformed stored state, and wrong password paths.

`sensitive.Local` is constructed only from `secret.Bytes` through `NewLocal(lookupKey, encryptionKey, keyVersion)`. Lookup digest is `HMAC-SHA256(lookupKey, "TALENRO-LOOKUP-V1\x00" || domain || "\x00" || canonical)`; AEAD AAD is `"TALENRO-FIELD-V1\x00" || domain || "\x00" || keyVersion`. Nonce is 12 secure random bytes and ciphertext storage is `nonce || sealed`.

- [ ] **Step 4: Benchmark and enforce the startup policy**

Run once outside short unit timeouts:

```powershell
go test ./internal/identity -run '^$' -bench BenchmarkArgon2idV1 -benchtime 3x
go test ./internal/sensitive ./internal/identity -count=1
```

Record the local duration and memory in `password_benchmark_test.go` comments. Do not weaken parameters based on the workstation; a future policy version and capacity review are required for change.

- [ ] **Step 5: Commit**

```bash
git add internal/sensitive internal/identity testdata/crypto/rfc9106
git commit -m "feat: add protected identity credentials"
```

### Task 8: Wrap idempotency and transactional outbox behind application repositories

**Files:**
- Create: `internal/idempotency/service_test.go`
- Create: `internal/idempotency/service.go`
- Create: `internal/outbox/repository_test.go`
- Create: `internal/outbox/repository.go`
- Create: `internal/outbox/consumer_test.go`
- Create: `internal/outbox/consumer.go`
- Create: `internal/contracts/events/marshal.go`

**Interfaces:**
- Produces transaction-scoped `idempotency.Repository` and `outbox.Repository`; both accept a sqlc `DBTX` supplied by the caller transaction.
- Consumes generated store methods and the validated Protobuf event envelope.

- [ ] **Step 1: Write failing replay, conflict, and duplicate-consumer tests**

Use a fake store to assert these exact outcomes:

```go
type Outcome string
const (
	Started Outcome = "started"
	Replay Outcome = "replay"
	Conflict Outcome = "conflict"
	InProgress Outcome = "in_progress"
)

func TestBeginClassifiesExistingRecord(t *testing.T) { /* same digest=>Replay; different=>Conflict */ }
func TestConsumerRunsSideEffectOnceForDuplicateEventID(t *testing.T) { /* two deliveries, one effect */ }
func TestOutboxPayloadRejectsPIIFields(t *testing.T) { /* descriptor guard */ }
```

The idempotency request digest is SHA-256 over `"TALENRO-IDEMPOTENCY-REQUEST-V1\x00" || operation || "\x00" || strict canonical request bytes`; the key digest uses a different domain. Neither plaintext key nor body is stored.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/idempotency ./internal/outbox ./internal/contracts/events -count=1`

Expected: compilation fails because repository wrappers and envelope marshal are absent.

- [ ] **Step 3: Implement the transaction contracts**

```go
package idempotency

type Repository interface {
	Begin(context.Context, Scope, string, []byte, time.Time, time.Time) (Record, Outcome, error)
	Complete(context.Context, Record, int, []byte) (Record, error)
}

type Scope struct { Principal, Operation string }
```

`Begin` rejects keys outside 22–86 base64url characters, calls `TryBeginIdempotency`, and on conflict calls `GetIdempotencyForUpdate` on the same transaction. A completed matching record decrypts through `sensitive.Protector` domain `idempotency/response/v1` and returns copied replay bytes; a mismatched digest returns `Conflict`; an unfinished matching record returns `InProgress`. Before `Complete`, response bodies are limited to 1 MiB and encrypted with key version; plaintext token-bearing responses are never stored. The fixed unauthenticated registration scope is `anonymous_registration`; every authenticated operation uses its opaque principal UUID plus credential domain.

```go
package outbox

type Repository interface {
	Append(context.Context, *eventsv1.EventEnvelope) error
}
type Handler interface { HandleInTransaction(context.Context, store.DBTX, *eventsv1.EventEnvelope) error }
```

`Append` calls `ValidateEnvelope`, deterministic protobuf marshal, and `InsertOutboxEvent`. For database side effects, consumer transaction order is: `RecordConsumedEvent`; if zero rows, commit without side effect; otherwise run `HandleInTransaction` against the same transaction and commit. Do not acknowledge NATS before commit. External-provider effects are not passed to this generic handler; the Task 18 email consumer uses delivery-ID idempotency because a remote call cannot share the PostgreSQL transaction.

- [ ] **Step 4: Verify package and integration behavior**

Run:

```powershell
go test ./internal/idempotency ./internal/outbox ./internal/contracts/events -count=1
go test -tags=integration ./internal/store -run 'TestIdempotency|TestOutbox' -count=1
```

Expected: matching replay is byte-identical, conflicting body is stable, and duplicate event IDs create one side effect.

- [ ] **Step 5: Commit**

```bash
git add internal/idempotency internal/outbox internal/contracts/events
git commit -m "feat: add idempotent transactional delivery"
```

### Task 9: Implement account registration, email verification, and delivery events

**Files:**
- Create: `internal/identity/repository.go`
- Create: `internal/identity/register_test.go`
- Create: `internal/identity/register.go`
- Create: `internal/identity/email_verification_test.go`
- Create: `internal/identity/email_verification.go`
- Create: `internal/identity/local_email_sender.go`
- Create: `internal/identity/postgres_repository.go`
- Create: `internal/ratelimit/redis_test.go`
- Create: `internal/ratelimit/redis.go`

**Interfaces:**
- Implements `identity.Application.RegisterAccount`, email verification, password-reset delivery/reset, and `CreateEnrollmentGrant` from the frozen interface.
- Produces `EmailSender.Send(context.Context, Delivery) error` and a local capture-only fixture.

- [ ] **Step 1: Write failing registration/enumeration/policy tests**

Cover `required`, `grace`, and `disabled`; duplicate email and new email return the same public 202 result shape; the duplicate path invokes exactly one dummy Argon2 derivation; the new path atomically creates account/email/password/security-event/outbox; a delivery failure cannot roll back registration; verification and password-reset tokens are distinct/single-use; password reset expires at 30 minutes and marks prior sessions for review; verification activates `pending_email` and provisional authorizations; required mode rejects grant before verification; grace returns only the fixed `trial_restricted` marker and 24-hour provisional boundary.

```go
type EmailSender interface { Send(context.Context, Delivery) error }
type Delivery struct {
	DeliveryID uuid.UUID
	TemplateID TemplateID
	Locale string
	Recipient string
	OneTimeToken secret.Bytes
}
```

`Delivery` exists only after the email consumer decrypts the protected pending-delivery record. `Send` is required to use `DeliveryID` as the provider idempotency key and to treat an already-accepted delivery as success. No sender method accepts an arbitrary template, map, provider options, principal ID, raw token string, or raw error; `OneTimeToken` cannot format or marshal.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/identity -run 'Test(Register|VerifyEmail|PasswordReset|EnrollmentPolicy)' -count=1`

Expected: compilation fails because the application/repository implementation is absent.

- [ ] **Step 3: Implement the exact command/results and transaction order**

```go
type RegisterAccountCommand struct {
	Email string
	Password secret.Bytes
	Locale string
	IdempotencyKey string
}
type RegisterAccountResult struct { Accepted bool }

type VerifyEmailCommand struct {
	Token secret.Bytes
	IdempotencyKey string
}

type CreateEmailVerificationDeliveryCommand struct {
	Email string
	Locale string
	IdempotencyKey string
}

type CreatePasswordResetDeliveryCommand struct {
	Email string
	Locale string
	IdempotencyKey string
}

type ResetPasswordCommand struct {
	Email string
	Token secret.Bytes
	NewPassword secret.Bytes
	ClientSigningPublicKey [32]byte
	IdempotencyKey string
}

type EnrollmentGrant struct {
	Token secret.Bytes
	ExpiresAt time.Time
	PolicyMarker string
}

type CreateEnrollmentGrantCommand struct {
	PrincipalID PrincipalID
	SessionID SessionID
	Reauthentication Reauthentication
	IdempotencyKey string
}
```

Registration transaction order is: begin idempotency; canonicalize/protect email; perform lookup; for a new identity generate/hash password and email token; encrypt one fixed-schema pending-delivery record containing only recipient, verification token, template and locale; insert account, email, credential and bounded security event; append `EmailDeliveryRequested` containing only delivery ID/template/locale; complete idempotency with the generic accepted response; commit. Existing identities perform current-cost dummy hash and complete with the same response. The token plaintext exists only before encryption and inside the `EmailSender` adapter call; it is never written to an event, log, metric, report or public response.

Email-delivery retry and password-reset delivery always return the same generic accepted shape. For an eligible known identity, each rate-limits through Redis, rotates the correct token hash and encrypted pending-delivery record, and appends one delivery event in one transaction; ineligible/unknown identities take the same limiter/dummy-work path without exposing existence. Verification transaction order is: digest token; lock/consume email; clear pending delivery; activate account; call injected `DeviceAuthorizationParticipant.ActivateVerifiedPrincipal` on the same `store.DBTX`; append state events; complete idempotency; commit. Identity never selects a deviceauth table. Password-reset transaction locks the identity/credential, consumes a nonexpired 30-minute token, hashes the new password, clears delivery ciphertext, marks every account session `review_required`, creates one new session bound to the supplied client signing key, appends a security event, completes idempotency and commits. Unknown/used/expired tokens all map to the same `authentication_failed` category.

`ratelimit.Redis.Allow` uses one Lua script: `INCR` the digest key, set `PEXPIRE` only when count becomes one, return count and `PTTL`. The Go adapter allows only when `count<=policy.Limit`, wraps the configured Redis timeout, and maps every Redis/script/parse ambiguity to a value-free dependency error. It never logs the key or subject.

- [ ] **Step 4: Verify concurrency and privacy**

Run:

```powershell
go test ./internal/identity ./internal/ratelimit -count=1
go test -race ./internal/identity -run 'Test(Register|VerifyEmail)' -count=1
go test -tags=integration ./internal/store -run TestIdentity -count=1
```

Run the duplicate/new registration timing harness 50 times and require the same status, response fields, Argon2 call count and configured deadline class; do not assert nanosecond equality.

- [ ] **Step 5: Commit**

```bash
git add internal/identity internal/ratelimit
git commit -m "feat: add private account registration"
```

### Task 10: Implement account authentication, isolated sessions, rotation, and revocation

**Files:**
- Create: `internal/identity/session_test.go`
- Create: `internal/identity/session.go`
- Create: `internal/identity/challenge_store_test.go`
- Create: `internal/identity/challenge_store.go`
- Create: `internal/identity/redis_challenge_store.go`
- Modify: `internal/identity/postgres_repository.go`

**Interfaces:**
- Completes `identity.Application.CreateSession`, `RotateSession`, and `RevokeSessions`.
- Produces `identity.AccountAuthenticator` for HTTP account-token middleware and a Redis-backed single-use challenge store.

- [ ] **Step 1: Write failing auth-domain and replay tests**

Tests require: unknown email, wrong password, suspended account and unknown access token all map to identical `authentication_failed`; account access cannot decode as device access; access TTL is 10 minutes; refresh idle/absolute are 30/90 days; rotation consumes one Redis challenge and one refresh token; two concurrent rotations yield one success; used-token replay marks the session compromised and revokes all active refresh tokens; Redis outage closes login/rotation but existing account access still validates through PostgreSQL; password policy upgrade happens only after successful login; authenticated password change requires recent reauthentication and marks every existing session `review_required` before creating a new bound session.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/identity -run 'Test(Session|AccountToken|Refresh|Redis)' -count=1`

Expected: compilation fails on missing session/challenge types.

- [ ] **Step 3: Implement exact session and challenge contracts**

```go
type CreateSessionCommand struct {
	Method SessionMethod
	Email string
	Password secret.Bytes
	WebAuthnCeremonyID string
	WebAuthnResponse json.RawMessage
	ClientSigningPublicKey [32]byte
	IdempotencyKey string
}
type SessionTokens struct {
	AccessToken secret.Bytes
	AccessExpiresAt time.Time
	RefreshToken secret.Bytes
	RefreshIdleExpiresAt time.Time
	RefreshAbsoluteExpiresAt time.Time
}
type RotateSessionCommand struct {
	RefreshToken secret.Bytes
	ChallengeID string
	RequestNonce [32]byte
	Signature [64]byte
	IdempotencyKey string
}
type CreateSessionChallengeCommand struct {
	RefreshToken secret.Bytes
	RequestNonce [32]byte
	IdempotencyKey string
}
type SessionChallenge struct {
	ChallengeID string
	Challenge [32]byte
	ExpiresAt time.Time
}
type RevokeSessionsCommand struct {
	PrincipalID PrincipalID
	Scope SessionRevokeScope
	SessionID SessionID
	IdempotencyKey string
}

type ChangePasswordCommand struct {
	PrincipalID PrincipalID
	CurrentPassword secret.Bytes
	NewPassword secret.Bytes
	Reauthentication Reauthentication
	ClientSigningPublicKey [32]byte
	IdempotencyKey string
}

type ChallengeStore interface {
	Create(context.Context, ChallengeRecord, time.Duration) error
	Consume(context.Context, string, [32]byte) (ChallengeRecord, error)
}
```

`CreateSessionChallenge` authenticates the refresh-token digest against PostgreSQL authority before writing Redis. Redis keys are random challenge IDs, never token/principal/device values. Values contain protocol version, fixed operation, random nonce, context digest and expiry only. Consume uses one Lua script implementing atomic `GETDEL` and a 2-minute key TTL; any Redis ambiguity fails closed.

`SessionMethod` is a closed enum `password|passkey`; exactly one matching proof variant is allowed. Task 10 implements the password branch and Task 11 implements the passkey branch using the same session transaction. Login always calls the Redis limiter using the rotating HMAC of canonical email before the real/dummy Argon2 path; limiter denial returns finite 429, and Redis ambiguity returns sanitized 503. At login the client supplies a 32-byte Ed25519 session public key, which is stored on the account session. Refresh proof bytes begin `"TALENRO-ACCOUNT-ROTATION-V1\x00"` and bind protocol `account-rotation-v1`, challenge, session UUID, operation `rotate_account_token`, configured audience and request nonce. The server loads the signing key from the locked session row; a request cannot substitute it. This key is a per-session client key and is not the device identity key introduced in Task 12.

Account rotation transaction locks the refresh row, distinguishes `active` from `used`, validates session/account state and absolute/idle expiry, marks old used, rotates access, inserts new refresh, completes idempotency and commits. On `used`, the same transaction marks session compromised, revokes refresh tokens, writes a fixed security event and returns public authentication failure.

- [ ] **Step 4: Verify race, Redis fault, and integration behavior**

Run:

```powershell
go test ./internal/identity -count=1
go test -race ./internal/identity -run 'TestConcurrentAccountRefresh' -count=1
go test -count=50 ./internal/identity -run 'TestConcurrentAccountRefresh|TestUsedRefreshReplay'
```

With Docker, stop only Redis and assert new session/rotation fail with a sanitized 503 while a non-expired access-token repository lookup still succeeds; start Redis and assert recovery.

- [ ] **Step 5: Commit**

```bash
git add internal/identity
git commit -m "feat: add isolated account sessions"
```

### Task 11: Add Passkey, TOTP, and one-time recovery-code controls

**Files:**
- Create: `internal/identity/webauthn_test.go`
- Create: `internal/identity/webauthn.go`
- Create: `internal/identity/totp_test.go`
- Create: `internal/identity/totp.go`
- Create: `internal/identity/recovery_test.go`
- Create: `internal/identity/recovery.go`
- Create: `internal/identity/reauth.go`
- Modify: `internal/identity/postgres_repository.go`

**Interfaces:**
- Produces explicit strong-auth methods used by the Task 17 HTTP adapter.
- Consumes pinned go-webauthn and pquerna/otp only through identity-owned adapters; Redis ceremony state uses the Task 10 challenge store with separate domains.

- [ ] **Step 1: Write failing ceremony and independent-revocation tests**

Cover: random opaque WebAuthn user handle rather than email; exact RP ID/origin allowlists; `attestation=none`, `residentKey=preferred`, `userVerification=required`; ceremony expiry/single-use; maximum 10 active passkeys; counter rollback rejection; one pending/active TOTP only; six digits, SHA-1 compatibility for RFC 6238, 30-second step and ±1 window; atomic last-step replay rejection; exactly 10 new recovery codes; plaintext displayed once; atomic single-code removal; rotation supersedes the old set; each factor can be revoked without revoking the other factors; every mutation requires a recently reauthenticated account session.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/identity -run 'Test(Passkey|WebAuthn|TOTP|Recovery|Reauth)' -count=1`

Expected: compilation fails on the missing factor APIs.

- [ ] **Step 3: Implement the exact application surface**

```go
type StrongAuthApplication interface {
	BeginPasskeyRegistration(context.Context, BeginPasskeyRegistrationCommand) (json.RawMessage, error)
	FinishPasskeyRegistration(context.Context, FinishPasskeyRegistrationCommand) error
	BeginPasskeyAuthentication(context.Context, BeginPasskeyAuthenticationCommand) (json.RawMessage, error)
	FinishPasskeyAuthentication(context.Context, FinishPasskeyAuthenticationCommand) (SessionTokens, error)
	RevokePasskey(context.Context, RevokePasskeyCommand) error
	BeginTOTPEnrollment(context.Context, BeginTOTPEnrollmentCommand) (TOTPEnrollment, error)
	VerifyTOTPEnrollment(context.Context, VerifyTOTPEnrollmentCommand) error
	RevokeTOTP(context.Context, RevokeTOTPCommand) error
	RotateRecoveryCodes(context.Context, RotateRecoveryCodesCommand) (RecoveryCodes, error)
	ConsumeRecoveryCode(context.Context, ConsumeRecoveryCodeCommand) (SessionTokens, error)
}

type Reauthentication struct {
	SessionID SessionID
	Method ReauthMethod
	Proof secret.Bytes
}
```

WebAuthn ceremony JSON is bounded at 64 KiB before entering the dependency. Store only the library session data plus a fixed operation and expiry under a random Redis ID; finish consumes it once. Persist credential ID/public key/flags/sign count only after the library verifies RP ID, origin, challenge and UV. Dependency errors map to fixed internal fingerprints.

TOTP secrets are 20 random bytes, encrypted with domain `identity/totp/v1`, and rendered as an `otpauth` URI only in the one enrollment response. Verification locks the row, checks steps `now-1..now+1`, and calls `AcceptTOTPStep` in the same transaction; zero affected rows means replay.

Recovery codes are 20 random bytes encoded as grouped base32 for display. Store only domain-separated SHA-256 digests; validate all ten 32-byte entries before `CreateRecoveryCodeSet`. Successful consume creates a session in the same transaction and returns the code plaintext nowhere.

- [ ] **Step 4: Verify protocol fixtures, concurrency, and dependency boundaries**

Run:

```powershell
go test ./internal/identity -count=1
go test -race ./internal/identity -run 'TestConcurrent(TOTP|Recovery)' -count=1
go test -count=50 ./internal/identity -run 'TestConcurrent(TOTP|Recovery)'
```

Construct protocol-minimal registration/authentication fixtures directly in `_test.go` files from the W3C WebAuthn Level 3 data model, and record `https://www.w3.org/TR/webauthn-3/` in test comments. Do not copy the external W3C conformance corpus into this repository.

- [ ] **Step 5: Commit**

```bash
git add internal/identity
git commit -m "feat: add independently revocable account factors"
```

### Task 12: Implement one-time device enrollment and proof of possession

**Files:**
- Create: `internal/deviceauth/types.go`
- Create: `internal/deviceauth/pop_test.go`
- Create: `internal/deviceauth/pop.go`
- Create: `internal/deviceauth/challenge_store_test.go`
- Create: `internal/deviceauth/challenge_store.go`
- Create: `internal/deviceauth/redis_challenge_store.go`
- Create: `internal/deviceauth/register_test.go`
- Create: `internal/deviceauth/register.go`
- Create: `internal/deviceauth/postgres_repository.go`

**Interfaces:**
- Implements `CreateChallenge` and `RegisterDevice` on the frozen `deviceauth.Application`.
- Consumes an account-authenticated enrollment grant from identity, Redis for nonce state, and Task 4 transaction queries.

- [ ] **Step 1: Write failing PoP substitution/replay/concurrency tests**

Generate client Ed25519 and X25519 keypairs in the test. The valid proof succeeds; changing protocol version, challenge, grant digest, either public key, operation, audience, request nonce or signature fails. An anonymous challenge fails. A challenge is single-use and two concurrent registrations using one grant create exactly one device/authorization/family. A grace account produces `provisional` plus the exact immutable `device-policy-v1` JSON; required unverified account fails. Verify no private-key bytes enter repository calls, logs, metrics or events.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/deviceauth -run 'Test(PoP|Challenge|Register|Concurrent)' -count=1`

Expected: compilation fails because deviceauth implementation is absent.

- [ ] **Step 3: Freeze and implement the signed proof bytes**

```go
type ProofInput struct {
	ProtocolVersion string
	Challenge [32]byte
	GrantDigest [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey [32]byte
	Operation string
	Audience string
	RequestNonce [32]byte
}

func ProofBytes(input ProofInput) []byte
```

`ProofBytes` is the concatenation of fixed-width fields and length-prefixed UTF-8 strings, beginning with `"TALENRO-DEVICE-POP-V1\x00"`. The only accepted constants are protocol `device-pop-v1`, operation `register_device`, and the configured HTTPS public origin. Verification is `ed25519.Verify(signingPublicKey, ProofBytes(input), signature)` after exact length checks.

The canonical `device-policy-v1` JSON is one of exactly two shapes: standard `{"mode":"standard"}` or grace `{"expires_at":"<UTC-RFC3339-seconds>","max_devices":"1","mode":"trial_restricted"}`. The grace expiry is copied from the grant's `provisional_until`; neither shape contains a node-group, quota, plan or client-controlled field.

```go
type RegisterDeviceCommand struct {
	EnrollmentGrant secret.Bytes
	ChallengeID string
	RequestNonce [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey [32]byte
	DisplayName string
	Signature [64]byte
	IdempotencyKey string
}

type CreateChallengeCommand struct {
	Kind ChallengeKind
	EnrollmentGrant secret.Bytes
	RefreshToken secret.Bytes
	RequestNonce [32]byte
	SigningPublicKey [32]byte
	HPKEPublicKey [32]byte
	IdempotencyKey string
}

type Challenge struct {
	ChallengeID string
	Challenge [32]byte
	ExpiresAt time.Time
}
```

`ChallengeKind` is the closed enum `registration|rotation`; Task 12 implements registration and Task 13 adds rotation. Exactly the matching credential/key fields are allowed. Registration challenge first applies the configured challenge limiter to a rotating HMAC of the grant digest, then stores a random nonce and digest of grant + exact public request context in Redis for 2 minutes. Registration first begins the PostgreSQL transaction and classifies idempotency: a completed replay returns the decrypted original response without touching Redis; a new request atomically consumes Redis state, then follows the exact Task 4 insert order and completes idempotency before commit. A Redis-consumed but database-uncommitted attempt must obtain a new challenge while keeping the same idempotency key and request digest; the existing `in_progress` record is removed by rollback, so no second device can survive.

- [ ] **Step 4: Verify race and PostgreSQL integration**

Run:

```powershell
go test ./internal/deviceauth -count=1
go test -race ./internal/deviceauth -run TestConcurrentRegistration -count=1
go test -count=50 ./internal/deviceauth -run 'TestConcurrentRegistration|TestChallengeSingleUse'
go test -tags=integration ./internal/store -run TestDevice -count=1
```

Expected: every field substitution fails; exactly one device survives the transaction race.

- [ ] **Step 5: Commit**

```bash
git add internal/deviceauth
git commit -m "feat: add proof-bound device enrollment"
```

### Task 13: Implement isolated device tokens, rotation, authorization, and revocation

**Files:**
- Create: `internal/deviceauth/token_test.go`
- Create: `internal/deviceauth/token.go`
- Create: `internal/deviceauth/rotate_test.go`
- Create: `internal/deviceauth/rotate.go`
- Create: `internal/deviceauth/revoke_test.go`
- Create: `internal/deviceauth/revoke.go`
- Modify: `internal/deviceauth/postgres_repository.go`

**Interfaces:**
- Completes `RotateDeviceToken`, `RevokeDevice`, and `AuthorizeBundle` on the frozen interface.
- Produces `DeviceAuthenticator.Authenticate(context.Context, secret.Bytes) (BundleAuthority, error)` for device-only HTTP routes and implements `TrustTransactionParticipant` for trust-owned issuance transactions.

- [ ] **Step 1: Write failing domain-isolation, replay, and revocation tests**

Tests require device access never authenticates account routes and account access never authenticates device routes; TTLs are 10 minutes/30-day idle/90-day absolute; rotation proof binds challenge, family, operation, audience and request nonce; two concurrent rotations yield one success; old-token replay marks family compromised and revokes its active refresh tokens; suspended/revoked device or account cannot rotate or resolve; device revocation atomically revokes authorization, device, every family/refresh and future bundle issuance; other devices remain valid.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/deviceauth -run 'Test(DeviceToken|Rotate|Replay|Revoke|Authorize)' -count=1`

Expected: compilation fails on missing device token operations.

- [ ] **Step 3: Implement the exact token proof and transitions**

```go
type RotateDeviceTokenCommand struct {
	RefreshToken secret.Bytes
	ChallengeID string
	RequestNonce [32]byte
	Signature [64]byte
	IdempotencyKey string
}

type RevokeDeviceCommand struct {
	AccountPrincipal identity.PrincipalID
	DeviceID uuid.UUID
	Reauthentication identity.Reauthentication
	IdempotencyKey string
}

type BundleAuthority struct {
	AuthorizationID uuid.UUID
	PrincipalID uuid.UUID
	DeviceID uuid.UUID
	HPKEPublicKey [32]byte
	DeviceKeyVersion uint32
	PolicySchema string
	Policy json.RawMessage
}
```

The rotation branch of `CreateChallenge` authenticates refresh-token digest and active family/authorization/account state in PostgreSQL, applies the challenge limiter to a rotating family digest, then places the 2-minute record in Redis. Rotation proof begins `"TALENRO-DEVICE-ROTATION-V1\x00"` and contains fixed protocol, challenge, family UUID bytes, operation `rotate_device_token`, configured audience and request nonce. The device signing public key is loaded through authorization; the request cannot choose it. The PostgreSQL transition locks refresh/family, marks used, rotates access/idle expiry, inserts new refresh and completes idempotency. Used-token replay commits compromise/revocation/security-event/outbox before returning authentication failure.

Device revocation locks authorization, revokes all related device rows, calls injected `identity.DeviceTransactionParticipant.RevokeAuthorizationSessions` on the same transaction, and emits `DeviceAuthorizationChanged`; no cross-module table is selected from deviceauth application code—account state and reauthentication arrive through identity application interfaces.

- [ ] **Step 4: Verify concurrency and fault behavior**

Run:

```powershell
go test ./internal/deviceauth -count=1
go test -race ./internal/deviceauth -run 'TestConcurrentDeviceRefresh|TestDeviceRevocationRace' -count=1
go test -count=50 ./internal/deviceauth -run 'TestConcurrentDeviceRefresh|TestUsedDeviceRefreshReplay|TestDeviceRevocationRace'
```

Expected: one rotation wins, replay always compromises the family, and revocation cannot be bypassed by a concurrent resolve.

- [ ] **Step 5: Commit**

```bash
git add internal/deviceauth
git commit -m "feat: add isolated device authorization tokens"
```

### Task 14: Implement canonical trust metadata and the local Ed25519 signers

**Files:**
- Create: `internal/trust/types.go`
- Create: `internal/trust/validate_test.go`
- Create: `internal/trust/validate.go`
- Create: `internal/trust/jcs_test.go`
- Create: `internal/trust/jcs.go`
- Create: `internal/trust/signer_test.go`
- Create: `internal/trust/signer.go`
- Create: `internal/trust/metadata_test.go`
- Create: `internal/trust/metadata.go`
- Create: `internal/trust/postgres_repository.go`
- Create: `testdata/crypto/rfc8785/README.md`
- Create upstream-permitted vector files under: `testdata/crypto/rfc8785/`

**Interfaces:**
- Implements frozen `trust.ConfigSigner`; adds a separate local root signer fixture and root-metadata verifier.
- Produces strict typed payload/signed-bundle/metadata values and RFC 8785 canonicalization.

- [ ] **Step 1: Write failing JCS, schema, signature, and rotation tests**

Tests run the RFC 8785 official vectors, reject duplicate keys/non-I-JSON/non-finite numbers/invalid Unicode, require long integers and timestamps to be strings, enforce every object/collection/byte bound, and reject unknown fields. Sign/verify must fail after changing payload byte, payload digest, key ID, algorithm or signature. Metadata tests require monotonically increasing version, root signature, active-key validity, “publish new key first, sign later,” retirement grace, and rollback/unknown-root/unknown-algorithm rejection.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./internal/trust -run 'Test(JCS|Validate|Sign|Metadata)' -count=1`

Expected: compilation fails because trust types and validators are absent.

- [ ] **Step 3: Implement the exact versioned JSON types**

```go
type PayloadV1 struct {
	SchemaVersion string `json:"schema_version"`
	BundleID string `json:"bundle_id"`
	BundleLocator string `json:"bundle_locator"`
	BundleVersion string `json:"bundle_version"`
	Audience string `json:"audience"`
	IssuedAt string `json:"issued_at"`
	NotBefore string `json:"not_before"`
	ExpiresAt string `json:"expires_at"`
	PolicySnapshot json.RawMessage `json:"policy_snapshot"`
	TestConfig TestConfigV1 `json:"test_config"`
}

type TestConfigV1 struct {
	Message string `json:"message"`
	Sequence string `json:"sequence"`
}

type SignedBundleV1 struct {
	PayloadJCS string `json:"payload_jcs"`
	PayloadSHA256 string `json:"payload_sha256"`
	SignerKeyID string `json:"signer_key_id"`
	Algorithm string `json:"algorithm"`
	Signature string `json:"signature"`
	Padding string `json:"padding"`
}
```

Fixed constants are schema `talenro-config-bundle/v1`, signature algorithm `Ed25519`, signing domain `"TALENRO-CONFIG-BUNDLE-SIGNATURE-V1\x00"`, metadata domain `"TALENRO-TRUST-METADATA-V1\x00"`, UTC RFC3339 timestamps without fractional seconds, and decimal unsigned integers without leading zero. `PolicySnapshot` must strictly decode as `device-policy-v1`; `TestConfigV1.Message` is at most 256 bytes and `Sequence` matches the integer-string rule.

Canonicalization calls `jcs.Transform` only after `strictjson` validation and then re-decodes the output into the same exact type. The checked-in README records upstream URL, commit/tag, file hashes and Apache-2.0 notice.

- [ ] **Step 4: Implement signer fixtures and metadata repository flow**

`NewLocalConfigSigner(seed secret.Bytes)` and `NewLocalRootSigner(seed secret.Bytes)` require exactly 32 bytes, derive Ed25519 keys, and use distinct key-ID domains. Neither exposes private-key bytes. Production composition never constructs them. The signer call is wrapped in the configured 500ms–5s timeout.

Metadata publication transaction validates root signature, inserts the root metadata and its complete signing-key set, then commits. `EnsureLocalMetadata` exists only in local/test composition and idempotently publishes version 1 from the two configured fixture seeds when no metadata exists. Production never auto-generates or auto-publishes a root/signing key; it requires externally provisioned metadata plus an external signer whose key ID is already active. Bundle issue selects one `active` key whose validity covers the full bundle interval; a `future`, `retiring` outside grace, `revoked`, absent, or not-yet-published key fails closed.

- [ ] **Step 5: Verify vectors and deterministic bytes**

Run twice:

```powershell
go test ./internal/trust -run 'Test(JCS|Validate|Sign|Metadata)' -count=1
```

Hash golden JCS and metadata outputs before/after; expected hashes are identical.

- [ ] **Step 6: Commit**

```bash
git add internal/trust testdata/crypto/rfc8785 docs/licenses/dependency-policy.md
git commit -m "feat: add canonical trust metadata"
```

### Task 15: Build the fixed HPKE envelope and an independent reference verifier

**Files:**
- Create: `internal/trust/envelope_test.go`
- Create: `internal/trust/envelope.go`
- Create: `internal/trust/fuzz_test.go`
- Create: `internal/trustclient/types.go`
- Create: `internal/trustclient/verify_test.go`
- Create: `internal/trustclient/verify.go`
- Create: `internal/trustclient/store_test.go`
- Create: `internal/trustclient/store.go`
- Create: `internal/trustclient/fuzz_test.go`
- Create: `cmd/trust-conformance/main.go`
- Create: `cmd/trust-conformance/main_test.go`
- Create: `testdata/crypto/rfc9180/README.md`
- Create: `testdata/crypto/rfc9180/base-x25519-hkdf-sha256-chacha20poly1305.json`

**Interfaces:**
- Produces `trust.Seal`, `trustclient.VerifyAndStage`, and a standalone conformance executable.
- Consumes Go 1.26 `crypto/hpke`, but server and client validation implementations may share constants/types only—not verification control flow.

- [ ] **Step 1: Write failing KAT, tamper, rollback, and crash-safety tests**

Cover the RFC 9180 Appendix A Base-mode test vector whose suite is X25519/HKDF-SHA256/ChaCha20-Poly1305; copy only that case into `base-x25519-hkdf-sha256-chacha20poly1305.json` and record the RFC section, source URL and SHA-256 in the README. Use its private recipient key, `enc`, AAD and ciphertext to test deterministic `Recipient.Open`; the public Go API intentionally chooses fresh sender randomness and is not forced to reproduce the vector `enc`. Also test successful cross-package server seal/client open; wrong recipient key; and individual mutation of envelope version, suite IDs, recipient selector, locator, `enc`, ciphertext, AAD, payload, digest, signature, audience, key ID and every time/version field. Reject truncation, input over 1 MiB, plaintext over its bucket limit, expired/future outside 120-second skew, lower/equal conflicting version and trust-metadata rollback. Simulate a crash between durable stage and active-config swap: restart must retain the higher trusted version and never reactivate a lower bundle.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/trust ./internal/trustclient ./cmd/trust-conformance -run 'Test(HPKE|Envelope|Verify|Rollback|Crash)' -count=1`

Expected: compilation fails on missing envelope/verifier functions.

- [ ] **Step 3: Implement exact envelope formation**

```go
type OuterEnvelopeV1 struct {
	EnvelopeVersion string `json:"envelope_version"`
	KEM string `json:"kem"`
	KDF string `json:"kdf"`
	AEAD string `json:"aead"`
	RecipientKeyID string `json:"recipient_key_id"`
	BundleLocator string `json:"bundle_locator"`
	Enc string `json:"enc"`
	Ciphertext string `json:"ciphertext"`
}

func Seal(random securitykit.RandomSource, recipient [32]byte, locator [32]byte, signed SignedBundleV1) ([]byte, [32]byte, error)
```

Exact public suite strings are `DHKEM-X25519-HKDF-SHA256`, `HKDF-SHA256`, and `CHACHA20-POLY1305`; `info` is `talenro-config-bundle/v1`. Use the fixed API sequence:

```go
publicKey, err := ecdh.X25519().NewPublicKey(recipient[:])
if err != nil { return nil, [32]byte{}, ErrRecipientKey }
hpkeKey, err := hpke.NewDHKEMPublicKey(publicKey)
if err != nil { return nil, [32]byte{}, ErrRecipientKey }
enc, sender, err := hpke.NewSender(
	hpkeKey,
	hpke.HKDFSHA256(),
	hpke.ChaCha20Poly1305(),
	[]byte("talenro-config-bundle/v1"),
)
if err != nil { return nil, [32]byte{}, ErrEnvelopeSeal }
```

After `enc` exists, build an outer-header type excluding ciphertext, JCS-canonicalize it as AAD, and call `sender.Seal(aad, paddedSignedBundle)`. The per-bundle selector is the first 16 bytes of SHA-256 over `"TALENRO-RECIPIENT-SELECTOR-V1\x00" || recipient || locator`. All binary fields are unpadded base64url.

Padding uses server randomness and the smallest one of fixed total plaintext buckets `4096, 8192, 16384, 32768, 65536`; inputs larger than 64 KiB fail before HPKE. Final envelope JCS must be at most 1 MiB and its SHA-256 is the distribution digest.

- [ ] **Step 4: Implement the independent client sequence and atomic store**

```go
type Store interface {
	LoadHighest(context.Context, string) (uint64, error)
	StageAndAdvance(context.Context, string, uint64, []byte) error
	Activate(context.Context, string, uint64) error
}

func VerifyAndStage(context.Context, []byte, ecdh.KeyExchanger, TrustMetadata, Expected, Store, time.Time, time.Duration) (VerifiedBundle, error)
```

Client order is fixed: bound bytes; strict outer parse; suite/length checks; recompute selector/header JCS; `hpke.NewRecipient` and `Open`; bound plaintext; remove validated padding; strict signed parse; re-canonicalize exact payload; verify digest and root-approved Ed25519 key; validate audience/locator/schema/time/semantics; compare persisted highest; atomically stage bytes and advance version; activate; return. An ack occurs only outside this function after durable success.

The CLI accepts file paths for envelope, metadata, private key and state directory; it writes no secret values, emits only `verified` or a finite category, and uses exit 0/2. The client store writes a temp file, fsyncs file, atomically renames, fsyncs directory where supported, and never lowers the highest-version file.

- [ ] **Step 5: Run vectors, fuzz, cross-process, and race gates**

Run:

```powershell
go test ./internal/trust ./internal/trustclient ./cmd/trust-conformance -count=1
go test ./internal/trust ./internal/trustclient -run '^$' -fuzz Fuzz -fuzztime 10s
go test -race ./internal/trust ./internal/trustclient -count=1
go build -o $env:TEMP\talenro-trust-conformance.exe ./cmd/trust-conformance
```

Have a test server process create the fixture and invoke the built CLI as a separate process. Expected: valid fixture exits 0; every tamper exits 2 without disclosing input.

- [ ] **Step 6: Commit**

```bash
git add internal/trust internal/trustclient cmd/trust-conformance testdata/crypto/rfc9180
git commit -m "feat: add signed HPKE trust envelope"
```

### Task 16: Issue immutable bundles and expose primary plus two byte-identical mirrors

**Files:**
- Create: `internal/trust/application_test.go`
- Create: `internal/trust/application.go`
- Create: `internal/trust/distribution_test.go`
- Create: `internal/trust/distribution.go`
- Create: `internal/trust/mirror.go`
- Create: `internal/trust/postgres_repository_test.go`
- Modify: `internal/trust/postgres_repository.go`

**Interfaces:**
- Implements the frozen `trust.Application`.
- Produces an immutable byte store used identically by primary `/b/{locator}`, Mirror A and Mirror B.

- [ ] **Step 1: Write failing issuance/distribution/idempotency tests**

Tests require: authorization before issue; one DB-allocated monotonic version; exact generation order; signer timeout closes only new issuance; no plaintext bundle persisted; same idempotency key/retry returns identical locator/envelope/URLs; different body conflicts; locator is 32 random bytes and only its domain-separated SHA-256 lookup digest plus encrypted recoverable value are stored; primary and both mirrors return exactly identical bytes/SHA-256/content type/cache headers; one mirror failure does not alter the other sources; ack is idempotent and cannot lower any version.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/trust -run 'Test(Issue|Distribution|Mirror|Acknowledge)' -count=1`

Expected: compilation fails because the application/distribution handlers are absent.

- [ ] **Step 3: Implement issue and resolution transactions**

```go
type IssueCommand struct {
	AuthorizationID uuid.UUID
	TestConfig TestConfigV1
	IdempotencyKey string
}
type IssuedBundle struct {
	BundleID uuid.UUID
	BundleVersion uint64
	Locator secret.Bytes
	EnvelopeSHA256 [32]byte
}
type Resolution struct {
	Locator string
	EnvelopeSHA256 string
	Locations [3]string
	CacheControl string
}
```

Within one transaction: begin idempotency; call injected `deviceauth.TrustTransactionParticipant.AuthorizeBundleInTransaction` with the same `store.DBTX`; call `NextBundleVersion`; generate IDs/locator; construct/validate/JCS/sign/pad/HPKE; protect locator; insert issuance bytes/digest; append `BundleIssued`; complete idempotency with the exact response bytes; commit. A signer or HPKE failure rolls back the version increment. Trust never selects deviceauth tables. The test configuration is only `{message,sequence}` and contains no node, endpoint, tunnel credential, quota or plan data.

Resolve authenticates the device, finds or issues its current test bundle according to a fixed server-side test sequence, then returns three configured HTTPS base URLs ending in the same opaque locator. The locator GET performs a 32-byte decode, computes `SHA-256("TALENRO-BUNDLE-LOCATOR-V1\x00" || locator)`, and performs a PostgreSQL immutable-byte lookup; it does not require auth headers and reveals only the same not-found shape for unknown/expired values. The digest needs no secret because locators have 256 bits of server randomness; this lets mirror processes query bytes without receiving any protector/signing keys.

- [ ] **Step 4: Implement immutable HTTP semantics**

All three handlers set exactly:

```text
Content-Type: application/vnd.talenro.bundle+json
Cache-Control: public, max-age=86400, immutable
ETag: "<lowercase-envelope-sha256-hex>"
X-Content-Type-Options: nosniff
```

They write stored bytes directly without JSON decoding/re-encoding. Range requests are rejected with 416; request bodies are ignored/closed; unknown, malformed and expired locators all return the same body-free 404. Mirror fixtures receive only a `ByteStore` and have no signer, HPKE, identity, deviceauth or sensitive protector dependency.

- [ ] **Step 5: Verify identity and outage behavior**

Run:

```powershell
go test ./internal/trust -count=1
go test -race ./internal/trust -run 'TestConcurrentIssue|TestMirror' -count=1
go test -count=50 ./internal/trust -run 'TestConcurrentIssue|TestAcknowledge'
```

Hash all three response bodies in tests; require one unique hash. Stop the signer fixture after publication: resolve of a new version fails 503, while all already-published unexpired locator GETs continue.

- [ ] **Step 6: Commit**

```bash
git add internal/trust
git commit -m "feat: publish immutable test bundles"
```

### Task 17: Replace the fail-closed placeholders with bounded HTTP adapters

**Files:**
- Create: `internal/controlapi/c1_handler_test.go`
- Create: `internal/controlapi/c1_handler.go`
- Create: `internal/controlapi/auth_test.go`
- Create: `internal/controlapi/auth.go`
- Create: `internal/controlapi/errors_test.go`
- Create: `internal/controlapi/errors.go`
- Create: `internal/controlapi/trace.go`
- Modify: `internal/controlapi/handler.go`
- Delete: `internal/controlapi/c1_unavailable.go`

**Interfaces:**
- Implements every Task 2 generated `controlapiv1.ServerInterface` method with identity/deviceauth/trust applications.
- Enforces the two opaque-token domains, strict JSON, idempotency headers, deadlines and stable public errors at one transport boundary.

- [ ] **Step 1: Write failing generated-handler conformance tests**

Create a compile-time assertion and a table with all Task 2 operation IDs:

```go
var _ controlapiv1.ServerInterface = (*Handler)(nil)

type operationCase struct {
	Method, Path, AuthDomain string
	RequiresIdempotency bool
	WantStatus int
}
```

For each route assert: correct account/device/public domain; wrong bearer domain yields the same 401 shape; missing/malformed/high-entropy-invalid idempotency key yields bounded 400; body at 65536 bytes is accepted and 65537 yields 413 without calling the app; duplicate/unknown/trailing JSON yields 400; request deadline cancellation reaches the app; raw errors and secret canaries never enter body/header/log/metrics; error code/action/status match the Task 2 contract. Immutable GET is public, body-free, 1 MiB bounded, and not instrumented with raw locator labels.

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./internal/controlapi -run 'Test(C1|Auth|Error|Body|Deadline)' -count=1`

Expected: tests fail because `C1Unavailable` returns 503 for every operation.

- [ ] **Step 3: Implement shared transport guards**

```go
type Applications struct {
	Identity identity.Application
	StrongAuth identity.StrongAuthApplication
	AccountAuth identity.AccountAuthenticator
	Device deviceauth.Application
	DeviceAuth deviceauth.DeviceAuthenticator
	Trust trust.Application
}

func NewHandler(checker *readiness.Checker, apps Applications, deadline time.Duration, random securitykit.RandomSource) *Handler
```

`auth.go` parses exactly `Authorization: Bearer <opaque>` with one space, decodes through the correct token domain and puts a typed principal in request context. It never accepts query/cookie tokens. Account routes are all strong-factor/revocation/grant routes; device routes are challenge after grant context, device rotation, bundle resolution/ack; `/v1/devices` is grant+challenge authenticated; account/device login and verification endpoints are public as specified in OpenAPI.

`errors.go` is the only writer for `apierrors.Error`. It obtains a 16-byte random trace ID, maps the finite code table to HTTP status, emits generated `PublicError`, and logs/reports only event/category/component/outcome/fingerprint/trace ID. It never calls `err.Error()` on an infrastructure or crypto error.

Every body method calls `http.MaxBytesReader`, then `strictjson.Decode`, then explicit decoded-length/Unicode/count validators before the application. Every write route passes the exact `Idempotency-Key`; no adapter generates one. A derived context uses the configured request deadline and is canceled before returning.

- [ ] **Step 4: Implement every operation mapping and remove placeholders**

Map operation IDs to application calls one-for-one:

| Operation IDs | Application |
| --- | --- |
| `createAccount`, `createEmailVerificationDelivery`, `verifyEmail`, `createPasswordResetDelivery`, `resetPassword`, `changePassword`, password branch of `createAccountSession`, `createAccountAuthChallenge`, `rotateAccountToken`, `revokeAccountSessions` | `identity.Application` |
| passkey branch of `createAccountSession` | `identity.StrongAuthApplication.FinishPasskeyAuthentication` |
| all `Passkey`, `TOTP`, and `RecoveryCode` operation IDs | `identity.StrongAuthApplication` |
| `createDeviceEnrollmentGrant` | `identity.Application.CreateEnrollmentGrant` |
| `createDeviceAuthChallenge`, `registerDevice`, `rotateDeviceToken`, `revokeDevice` | `deviceauth.Application` |
| `resolveConfigBundle`, `acknowledgeConfigBundle`, `getImmutableBundle` | `trust.Application` / immutable byte handler |

Delete `C1Unavailable` only after the compile-time assertion passes and the table proves no operation returns the placeholder trace ID. Keep `/livez` and `/readyz` behavior unchanged except Task 18 readiness policy.

- [ ] **Step 5: Verify route and race behavior**

Run:

```powershell
go test ./internal/controlapi -count=1
go test -race ./internal/controlapi -count=1
go test -count=20 ./internal/controlapi -run 'Test(C1|Auth|Deadline)'
```

Expected: all generated methods are reachable, bounded, domain-correct and privacy safe.

- [ ] **Step 6: Commit**

```bash
git add internal/controlapi
git commit -m "feat: expose bounded account device trust APIs"
```

### Task 18: Add asynchronous delivery, privacy-safe reporting, bounded metrics, and runtime composition

**Files:**
- Create: `internal/errorreport/reporter_test.go`
- Create: `internal/errorreport/reporter.go`
- Create: `internal/errorreport/local.go`
- Create: `internal/outbox/publisher_test.go`
- Create: `internal/outbox/publisher.go`
- Create: `internal/outbox/nats.go`
- Create: `internal/identity/email_consumer_test.go`
- Create: `internal/identity/email_consumer.go`
- Create: `internal/readiness/policy_test.go`
- Create: `internal/readiness/policy.go`
- Modify: `internal/readiness/checker.go`
- Modify: `internal/readiness/checker_test.go`
- Modify: `internal/observability/metrics.go`
- Modify: `internal/observability/metrics_test.go`
- Modify: `internal/platform/dependencies.go`
- Modify: `internal/platform/server.go`
- Modify: `cmd/control-api/main.go`
- Create: `cmd/bundle-mirror/main.go`
- Create: `cmd/bundle-mirror/main_test.go`
- Modify: `deploy/dev/compose.yaml`
- Modify: `.env.example`
- Modify: `scripts/smoke.ps1`
- Modify: `scripts/smoke.sh`

**Interfaces:**
- Produces `errorreport.Reporter`, outbox publisher/email consumer workers, policy-based readiness, finite security metrics, and local composition.
- Preserves PostgreSQL as authority; NATS/email/reporter failures cannot roll back committed domain transactions.

- [ ] **Step 1: Write failing worker, readiness, reporter, and composition tests**

Reporter tests require a fixed DTO (no map/error/body fields), queue capacity 100, batch max 20, timeout from config, nonblocking enqueue, drop-on-full counter, no recursive self-report, cancellation and provider-panic containment. Publisher tests require claim limit 100, 30-second lease, deterministic subject allowlist, publish then mark, failed publish release with bounded exponential backoff (1s, 2s, 4s, 8s, max 30s, ±20% crypto-random jitter), duplicate delivery idempotence and no raw NATS error. Email consumer tests require delivery-ID dedupe and provider failure without account rollback.

Readiness tests require:

- PostgreSQL first failure: `down` immediately.
- Redis first/second consecutive failure: `degraded`; third: `down`; two consecutive successes recover to `up`.
- NATS socket state alone never determines readiness.
- Outbox backlog `>=1000` or oldest `>=60s`: `degraded`; backlog `>=10000` or oldest `>=300s`: `down`; lower values: `up`.
- Signer, EmailSender and ErrorReporter are not readiness probes; signer blocks issuance, the others are async.

Composition tests require production profile to reject local/discard providers, missing HTTPS origin or unavailable configured external adapter. Local profile constructs fixtures. Mirror process has database/HTTP dependencies only and cannot obtain signer or field keys.

- [ ] **Step 2: Run tests and verify RED**

Run:

```powershell
go test ./internal/errorreport ./internal/outbox ./internal/identity ./internal/readiness ./internal/observability ./internal/platform ./cmd/control-api ./cmd/bundle-mirror -count=1
```

Expected: compilation/behavior fails because workers/policy/composition are absent and readiness still directly mirrors all sockets.

- [ ] **Step 3: Implement fixed DTOs and worker lifecycles**

```go
type Report struct {
	Event Event
	Category Category
	Component Component
	Outcome Outcome
	Fingerprint Fingerprint
	BuildVersion string
	TraceID string
}

type Reporter interface { TryReport(Report) bool }
```

All fields except build/trace are enum types. The reporter owns one bounded channel and one worker; provider calls use independent timeout contexts. `Close` rejects new reports, drains only until shutdown deadline, then drops remaining items with a metric.

Outbox publisher owns one goroutine and uses subjects derived only from the finite event type registry. NATS message ID is event UUID; raw payload is the validated Protobuf envelope. Email consumer first checks `consumed_event_ids`, retrieves the encrypted one-time delivery by delivery ID, decrypts inside the adapter boundary, calls `EmailSender` with that delivery ID as mandatory provider idempotency key, then records the consumed event and clears the pending ciphertext when the provider reports acceptance. A crash after provider acceptance may call `Send` again, but the same delivery ID must not create a second provider effect. Provider response/error bytes are discarded.

- [ ] **Step 4: Implement configurable bounded readiness and metrics**

Add these local defaults and production-safe ranges:

```dotenv
TALENRO_REDIS_DOWN_AFTER_FAILURES=3
TALENRO_REDIS_RECOVER_AFTER_SUCCESSES=2
TALENRO_OUTBOX_DEGRADED_BACKLOG=1000
TALENRO_OUTBOX_DOWN_BACKLOG=10000
TALENRO_OUTBOX_DEGRADED_AGE=60s
TALENRO_OUTBOX_DOWN_AGE=300s
TALENRO_ERROR_REPORT_QUEUE=100
TALENRO_ERROR_REPORT_BATCH=20
```

Allowed ranges are failures/successes `1..10`, degraded backlog `100..10000`, down backlog `degraded+1..100000`, degraded age `10s..10m`, down age `degraded+1s..1h`, queue `10..1000`, batch `1..100` and `batch<=queue`. Update both smoke-script exact allowlists.

New Prometheus families use only fixed labels:

- `talenro_security_events_total{operation,result,reason}`
- `talenro_outbox_backlog` and `talenro_outbox_oldest_seconds` with no labels
- `talenro_error_reports_total{component,result}`
- `talenro_crypto_validations_total{operation,result,reason}`

Tests structurally gather descriptors and send attacker-controlled strings; the number of series must stay bounded and no canary may appear in a label/help text.

- [ ] **Step 5: Compose bounded runtimes and mirror fixture**

`openRuntime` constructs pools/clients, protectors, local or external providers, repositories, applications, handler, publisher, consumers and reporter in dependency order. On any failure it closes only successfully opened dependencies in reverse order using fixed error categories. Shutdown order is: stop public/mirror listeners; stop intake; stop publisher/consumers; drain reporter to its deadline; close NATS, Redis, PostgreSQL. A shared cancellation cause never includes raw errors.

The reusable `cmd/bundle-mirror` accepts only database URL, listen address and HTTP budgets. Development Compose runs it twice as `mirror-a` and `mirror-b` on distinct ports; neither container receives signer, lookup, encryption, email or reporter key variables.

- [ ] **Step 6: Verify worker fault boundaries and commit**

Run:

```powershell
go test ./internal/errorreport ./internal/outbox ./internal/identity ./internal/readiness ./internal/observability ./internal/platform ./cmd/control-api ./cmd/bundle-mirror -count=1
go test -race ./internal/errorreport ./internal/outbox ./internal/readiness ./cmd/control-api ./cmd/bundle-mirror -count=1
go test -count=50 ./internal/outbox ./internal/readiness -run 'Test(Publisher|Consumer|Policy)'
```

Expected: async failures do not alter committed authority; readiness follows policy; all goroutines terminate under the configured deadline.

```bash
git add internal/errorreport internal/outbox internal/identity internal/readiness internal/observability internal/platform cmd/control-api cmd/bundle-mirror deploy/dev/compose.yaml .env.example scripts/smoke.ps1 scripts/smoke.sh
git commit -m "feat: compose resilient trust services"
```

### Task 19: Close C1.1 with Docker E2E, fault injection, privacy gates, and runbooks

**Files:**
- Create: `internal/e2e/c11_test.go`
- Create: `internal/e2e/faults_test.go`
- Create: `internal/e2e/privacy_test.go`
- Create: `scripts/verify-c11.ps1`
- Create: `scripts/verify-c11.sh`
- Modify: `scripts/verify.ps1`
- Modify: `scripts/verify.sh`
- Modify: `scripts/smoke.ps1`
- Modify: `scripts/smoke.sh`
- Create: `docs/security/c11-threat-model.md`
- Create: `docs/runbooks/signing-key-rotation-recovery.md`
- Create: `docs/runbooks/token-family-compromise.md`
- Create: `docs/runbooks/email-provider-outage.md`
- Create: `docs/runbooks/c11-dependency-failure.md`
- Modify: `docs/roadmap/implementation-sequence.md`

**Interfaces:**
- Verifies the complete public contract through real HTTP, Docker PostgreSQL/Redis/NATS, primary API, two mirror processes and the independent reference client.
- Produces deterministic one-command C1.1 acceptance for PowerShell and Bash.

- [ ] **Step 1: Write the failing E2E acceptance test before expanding smoke**

The integration-tagged test must execute this exact flow through HTTP only:

1. Register accounts under `required`, `grace`, and test-only `disabled` profiles.
2. Prove generic duplicate/new registration responses and verify required/grace behavior.
3. Log in with password, exercise generic password reset and authenticated password change, create/revoke Passkey, TOTP and recovery codes independently, and prove account-session review/revocation.
4. Obtain one grant, generate Ed25519/X25519 keys only in the reference client, complete PoP, and prove grant/challenge replay and anonymous registration fail.
5. Prove account/device token-domain isolation, one successful concurrent rotation, used-token family compromise and recovery by reenrollment.
6. Resolve one test bundle, fetch primary/Mirror A/Mirror B, require byte-identical SHA-256, independently decrypt/verify/stage it, ack it, and reject every tamper/rollback case.

Run: `go test -tags=e2e ./internal/e2e -run TestC11HappyPath -count=1`

Expected: RED because the smoke environment and fixtures are not yet orchestrated.

- [ ] **Step 2: Add bounded fault and privacy matrices**

Each matrix row records precondition, injected failure, expected HTTP/readiness/authority result, recovery check and maximum wall-clock bound:

| Failure | Required assertion |
| --- | --- |
| PostgreSQL stop | readiness down; auth/grant/rotation/resolve/ack closed; liveness remains |
| Redis stop | existing access validation continues; login/challenge/rotation closed; degraded then down threshold |
| NATS stop | committed registration succeeds; outbox grows; readiness follows age/backlog; recovery publishes once |
| EmailSender reject | generic registration remains committed; delivery retries/dedupes; no provider detail |
| ConfigSigner timeout | new issue fails; existing immutable bytes remain available |
| ErrorReporter reject/panic/full | business/readiness unchanged; finite drop metric increments |

Also stop/restart each mirror independently and primary independently, prove remaining sources serve the same bytes, then recover without regenerating. Run all grant/refresh/idempotency races 100 times.

Privacy tests inject distinct canaries in email, password, tokens, nonce, principal/device IDs, locator, key bytes, ciphertext, URL credentials, provider errors and HTTP bodies. Capture HTTP, structured logs, Prometheus gather, reporter DTOs, panic recovery and command output; zero canary matches are allowed. Inspect metric descriptors to prove bounded label cardinality.

- [ ] **Step 3: Implement two equivalent verification entrypoints**

`verify-c11.ps1/.sh` must resolve repo root from the script path, preserve exact nonzero child status, sanitize output, and run in this order:

1. `scripts/check-tools.*`
2. clean `scripts/generate.*`, then generated-tree diff check
3. `go test ./... -count=1`
4. focused fuzz targets, each 10 seconds
5. `go test -race ./... -count=1`
6. `go vet ./...`
7. pinned `go tool golangci-lint run ./...`
8. migration `up`, `down-to 1`, `up`
9. `go test -tags=integration ./... -count=1`
10. `docker compose -f deploy/dev/compose.yaml config --quiet`
11. `go test -tags=e2e ./internal/e2e -count=1`
12. both-shell bounded smoke when that shell exists.

Any unavailable required gate is a failed acceptance, not a pass. Scripts must identify only the fixed stage and exit code; no raw stderr, command line, environment value, URL or secret is echoed. They must stop only exact child/container targets they started and always run bounded cleanup while preserving the primary status.

- [ ] **Step 4: Expand smoke without weakening its existing ownership guarantees**

Build exact binaries before start, launch actual PIDs, and retain the current PowerShell exact-PID and Bash absolute TERM/KILL deadline guarantees. The smoke flow waits for primary and both mirrors, runs migrations, performs the happy path with the conformance binary, injects PostgreSQL/Redis/NATS failures one service at a time, restores each service, scans captured sanitized artifacts, and calls Compose down. Do not use `docker system prune`, volume deletion, wildcard process termination, shell evaluation, or ambient `TALENRO_*` variables.

- [ ] **Step 5: Write operational security documentation**

The threat model must enumerate assets, trust boundaries, attacker capabilities, account enumeration, token theft/replay, malicious mirror, database disclosure, dependency outage, insider signer misuse, log/metric leakage and excluded downstream node/tunnel threats, with the implemented control and residual risk for each.

Runbooks must contain detection using only finite events/metrics, immediate containment, authority checks, recovery, rotation/reenrollment sequencing, rollback prohibition, evidence handling without secret collection and explicit exit criteria. The signing runbook must enforce root ceremony/KMS as a production release gate; local seeds are never a production recovery method.

- [ ] **Step 6: Run the complete verification from a clean tree**

On Windows, set only process-scoped `GOOS=windows`, `GOARCH=amd64`, task-specific `GOCACHE`, `GOTMPDIR` and `GOLANGCI_LINT_CACHE`, then run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-c11.ps1
git status --short
```

If Git Bash is installed, run `bash scripts/verify-c11.sh` from a directory outside the repository. Expected: every stage PASS, no generated diff, Docker is clean after smoke, and the only status entries are the intended Task 19 files before commit.

- [ ] **Step 7: Mark C1.1 complete and commit**

Only after Step 6 passes, change roadmap item C1.1 from `current design` to `complete` and C1.2 to `current design` without claiming C1.2 implementation.

```bash
git add internal/e2e scripts docs/security docs/runbooks docs/roadmap/implementation-sequence.md
git commit -m "test: close account device trust acceptance"
```

## Specification coverage audit

| Design requirement | Plan tasks and executable evidence |
| --- | --- |
| PostgreSQL authority, reversible schemas, concurrency invariants | Tasks 3–5, 8–13; `up/down/up`, integration and 50/100-run races |
| Email privacy, password policy, enumeration resistance | Tasks 7, 9, 10, 18, 19; Argon2 vector/benchmark, dummy-work and canary tests |
| Passkey, TOTP, recovery codes, independent revocation | Task 11 plus HTTP/E2E coverage in Tasks 17 and 19 |
| One-time grant, Redis nonce, device PoP and atomic registration | Tasks 4, 12 and 19 substitution/replay/concurrency matrix |
| Account/device opaque-token isolation, rotation and family compromise | Tasks 3–4, 6, 10, 13, 17 and 19 |
| RFC 8785, Ed25519, root metadata and rotation order | Task 14, official vector and tamper/rollback tests |
| RFC 9180 fixed HPKE suite, strict envelope and client crash safety | Task 15, official recipient-open vector, cross-process/fuzz/race tests |
| Monotonic bundle issue/accept, primary plus two immutable mirrors | Tasks 5, 15, 16, 18 and 19; byte-identical three-source hashes |
| Idempotency, transactional outbox and at-least-once dedupe | Tasks 5, 8, 9, 12–13, 16, 18 and 19 |
| PostgreSQL/Redis/NATS/Email/Signer/Reporter fault matrix | Tasks 10, 16, 18 and Docker matrix in Task 19 |
| Stable errors, strict 64 KiB input, 1 MiB envelope and finite metrics | Tasks 2, 6, 15, 17–19 |
| Logs, metrics, HTTP and reporter privacy | Tasks 6–19, consolidated canary scan in Task 19 |
| Production provider/TLS gates and local-only fixtures | Tasks 1, 14, 18 and release checks in Task 19 |
| Deterministic OpenAPI/Protobuf/sqlc generation and full release gates | Tasks 2–5 and `verify-c11` in Task 19 |
| Threat model and four operational runbooks | Task 19 |
| Explicit exclusion of nodes, plans, quota, scheduling and tunnel cores | Global Constraints, Task 16 minimal `TestConfigV1`, and final scope scan |

## Reviewed dependency sources

- Go 1.26 `crypto/hpke`: `https://pkg.go.dev/crypto/hpke` and RFC 9180.
- go-webauthn v0.17.4: `https://github.com/go-webauthn/webauthn/releases/tag/v0.17.4` (BSD-3-Clause).
- gowebpki/jcs v1.0.1: `https://github.com/gowebpki/jcs/tree/v1.0.1` (Apache-2.0); official RFC vectors remain required because this adapter has a small maintenance surface.
- pquerna/otp v1.5.0: `https://github.com/pquerna/otp/releases/tag/v1.5.0`.
- Argon2id policy: RFC 9106 section 4 second recommended option, 64 MiB/3 iterations/4 lanes.
- WebAuthn behavior: W3C WebAuthn Level 3 Candidate Recommendation Snapshot, `https://www.w3.org/TR/webauthn-3/`.

Do not update these pins during implementation merely because a newer version appears. Any pin change requires a separate compatibility/license/security review and regeneration pass.
