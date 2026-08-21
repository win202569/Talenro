# C1.1 threat model

This document covers only the C1.1 account principal, device identity, and configuration trust-root implementation. It describes the controls present in this repository and the production controls that remain release gates. A local signer seed, local field protector, local email sender, loopback HTTP origin, or local error reporter is a test fixture, not a production security control.

## Security objectives

- Bind account authority to an opaque principal rather than an email string.
- Admit a device only after account-authenticated grant consumption and proof of possession of its Ed25519 key.
- Keep account and device credential domains isolated, rotate refresh tokens once, and revoke a whole family on proved replay.
- Deliver one device-specific, signed and HPKE-encrypted bundle as byte-identical immutable content through three independent HTTP locations.
- Prevent a server, mirror, database reader, or network observer from silently lowering a client's highest trusted metadata or bundle version.
- Keep secrets, identifiers, raw dependency errors, request material, and attacker-controlled text out of HTTP errors, logs, metrics, and remote reports.
- Fail closed where a dependency participates in an authority decision while keeping already committed facts and immutable objects available where the design permits.

## Scope and explicit exclusions

C1.1 includes `identity`, `deviceauth`, `trust`, the control API, PostgreSQL authority, Redis single-use state, the transactional outbox and NATS delivery path, email delivery, the signer boundary, the primary bundle location, two mirror fixtures, and the independent Go verifier.

The following are explicitly outside this threat model and receive no security claim from C1.1:

- node inventory, POP agents, node health, capacity, desired-state reconciliation, or node quarantine;
- plans, entitlements, device-count policy, traffic ledgers, quota leases, or billing;
- scheduling, country or node-group selection, Rendezvous placement, or failure-domain diversity;
- VLESS, REALITY, Hysteria2, Xray, sing-box, tunnel credentials, tunnel traffic, DNS, destinations, or traffic metadata;
- native Android, iOS, Windows, or macOS key storage and lifecycle behavior;
- production root-key ceremony, production KMS/HSM/Vault adapters, production email-provider adapters, and downstream tunnel deployment.

The C1.1 bundle contains only the fixed test configuration. Treating it as a production tunnel configuration, or adding node or tunnel data to it without the later C1.2-C1.4 designs, invalidates this model.

## Assets and security properties

| Asset | Required property | Authoritative location or owner |
| --- | --- | --- |
| Principal, account state, credentials, sessions, and security history | Integrity, revocation, non-enumerability, bounded disclosure | PostgreSQL; passwords are Argon2id records and recoverable email values are application-encrypted |
| Enrollment grants and account/device access and refresh tokens | Confidentiality, single use or bounded lifetime, domain separation, replay detection | Plaintext exists only at issuance/client boundaries; PostgreSQL stores digests and family state |
| Device authorization and policy snapshot | Integrity, monotonic state, binding to the principal and device public keys | PostgreSQL |
| Device Ed25519 and X25519 private keys | Confidentiality and non-exportability | Device only; the service stores public keys, never private keys |
| Redis challenges, WebAuthn ceremonies, and rate-limit windows | Confidentiality, atomic single use, short lifetime | Redis; never authority for revocation or identity |
| Idempotency records and tombstones | Integrity, isolation by principal/operation/key, replay of only the original result | PostgreSQL; private replay bodies are protected, and new device-token replay bodies require the authoritative `family_id` |
| Trust root and configuration signing keys | Integrity, controlled use, non-exportability, auditable rotation | Root ceremony plus production KMS/HSM/Vault; business code receives only the signer interface and public metadata |
| Root-signed trust metadata | Authenticity, expiry, and monotonically increasing version | PostgreSQL and client-persisted highest trusted version |
| Signed HPKE bundle and locator | Authenticity, recipient confidentiality, immutable bytes, expiry, audience binding, monotonic version | PostgreSQL/object byte store and client-persisted highest trusted version; locator is high entropy but is not authorization |
| Transactional outbox and consumed event IDs | Atomic creation, at-least-once delivery, consumer deduplication | PostgreSQL; NATS transports events but is not authority |
| Email verification/reset delivery material | Confidentiality, bounded lifetime, one-time consumption | Protected pending fields in PostgreSQL; the provider receives only the minimum delivery DTO |
| HTTP errors, logs, metrics, and ErrorReporter DTOs | Bounded vocabulary, no secret or identifier leakage, bounded cardinality | Compile-time registries and sanitized adapters |
| Runtime profile and dependency endpoints | Integrity and fail-closed production configuration | Deployment authority and process startup validation |

## Trust boundaries

1. **Untrusted client to control API.** TLS terminates before strict, size-bounded JSON parsing. Bearer domains are distinct. The client controls all request bytes, headers, idempotency keys, public keys, proofs, and timing.
2. **Application modules to PostgreSQL.** PostgreSQL is the sole authority for account/device state, token families, revocation, idempotency, issuance versions, acknowledgements, and outbox facts. Cross-domain changes use application participants inside one transaction, not direct reads of another module's tables.
3. **Application modules to Redis.** Redis owns disposable, short-lived, single-use coordination. Loss or ambiguity closes challenge-dependent operations; it cannot reactivate or revoke an authority record.
4. **PostgreSQL outbox to NATS and consumers.** Publication is at least once and is marked only after durable broker acknowledgement. Consumers must deduplicate by event ID. Message delivery never decides authentication or revocation.
5. **Identity worker to email provider.** Provider responses and raw errors are untrusted. Provider failure cannot roll back the committed account or expose whether an identity exists.
6. **Trust service to configuration signer.** The signer may fail, delay, panic, or be malicious. It receives domain-separated bytes through a bounded interface and does not expose private key material. Production key custody remains an external release gate.
7. **Trust service to primary and mirrors.** Locations receive an already generated immutable byte sequence. A mirror has byte-store capability only; it has no signing, decryption, account, or device authority.
8. **Distribution source to client verifier.** HTTP, locator knowledge, cache contents, and every envelope field are untrusted until strict parsing, HPKE open, JCS equality, Ed25519 verification, audience/expiry checks, and monotonic-version checks all succeed.
9. **Runtime to observability providers.** Logs, metrics, and reports cross an administrative or third-party boundary. Only finite events, labels, outcomes, fingerprints, build identity, and random trace correlation may cross it.
10. **Operators to root ceremony and KMS.** This is the highest-privilege boundary. Production rotation requires separated authority, witnessed approval, non-exportable keys, and an immutable audit record outside the application database.

## Attacker capabilities and assumptions

The model assumes an attacker can send arbitrary concurrent HTTP requests; observe response status, shape, size, and timing; steal one bearer token or locator; replay old requests, refresh tokens, events, metadata, or bundles; control a mirror or network cache; cause dependency timeouts and restarts; inject strings into provider errors; read a database snapshot; and obtain ordinary service-operator access. Tests also treat Redis, NATS, email, signer, and reporter behavior as Byzantine at their adapter boundaries for error, timeout, duplication, and panic containment.

The model does not assume resistance after simultaneous compromise of a target device private key and its active authorization, compromise of the production root authority, or arbitrary modification of both application binaries and their release controls. Those events require external incident response, key revocation, client update, or full trust-root recovery.

## Threat register

| Threat | Implemented control | Residual risk and required operation |
| --- | --- | --- |
| Account enumeration through registration, login, reset, or delivery endpoints | Unknown and existing identities use generic public results and stable errors; strict request limits apply; password paths retain equivalent work; metrics use fixed labels | Network-level latency, rate, and provider-delivery side channels cannot be eliminated by response shaping alone. Rate limiting, provider privacy, and aggregate anomaly review remain required. |
| Password guessing or credential stuffing | Argon2id policy v1, bounded authentication errors, rate-limited categories, optional Passkey/TOTP/recovery controls, and high-risk reauthentication | A reused or weak user password can still be compromised. Operations must monitor aggregate authentication failures without collecting identities and provide user-facing recovery outside this repository. |
| Account or device token theft | High-entropy opaque tokens, digest-only lookup, short access lifetime, refresh-family state, isolated account/device schemes, and authoritative revocation | A stolen live access token can be used until expiry or revocation. TLS, endpoint hardening, client secure storage, and rapid family revocation remain essential. |
| Refresh-token replay | Atomic one-winner rotation; a proved used-token replay marks the whole family compromised/revoked, revokes descendants, writes a fixed security classification, and emits the registered device-family event | A theft that is never replayed may remain undetected. Concurrent legitimate retries can resemble attack traffic; authority state and idempotency records, not a single metric, determine the incident. |
| Grant, challenge, ceremony, or idempotency replay | High-entropy grants and challenges, Redis atomic consumption, PostgreSQL transaction/unique constraints, request digests, stable completed replay, and long-lived security tombstones | Redis loss makes new challenge-dependent operations unavailable. It must never be bypassed with process memory or a newly fabricated challenge. |
| Device-token replay-format deployment skew | This branch is not deployed. Its pre-release cutover strictly rejects a legacy completed replay blob that lacks a valid `family_id`; there is no rolling mixed-version compatibility. Deployment must quiesce writes, wait for legacy idempotency retention or perform an approved bounded cleanup, and upgrade every reader/writer atomically. | New and old binaries cannot safely overlap. After any new-format replay is written, rollback to a legacy reader/writer is prohibited; recovery is forward-only with preserved authority and security history. |
| Device impersonation or public-key substitution | Account-authenticated one-time grant; proof covers protocol, challenge, grant context, both public keys, operation, audience, and request nonce; registration is atomic | Compromise of the device signing private key defeats PoP for that device. Revoke and reenroll with newly generated keys; do not transfer old keys. |
| Malicious primary, mirror, CDN, or on-path cache | Three locations serve the exact stored bytes; locator is opaque; envelope hash is returned by authenticated resolution; client performs HPKE, exact JCS, signature, audience, locator, expiry, schema, and monotonic checks | A source can deny service, serve stale bytes that the client rejects, or correlate locator access. Multi-source retry improves availability, not confidentiality against endpoint traffic analysis. |
| Bundle or trust-metadata rollback | PostgreSQL issuance versions and client `highest_trusted` values only increase; metadata is root-signed; acknowledgement, expiry, mirror loss, offline state, and runtime rollback cannot lower the barrier | Loss of the client's durable highest-version state can re-enable an otherwise valid old bundle. Native-client secure persistence is deferred and must be specified before production use. |
| Database disclosure | Email and pending delivery values are application-protected; passwords and tokens are one-way records; device private keys and signer private keys are absent; bundles are HPKE ciphertext; raw locators are protected and indexed by digest | A snapshot still exposes public keys, state relationships, timestamps, ciphertext sizes, security history, and encrypted values vulnerable to compromise of the field-protection key. Rotate protection keys and minimize database access. |
| Database modification or rollback | Transactions, constraints, state versions, idempotency digests, signed metadata/bundles, and client monotonic checks constrain silent changes | A privileged writer can deny service, revoke users, alter operational history, or roll back unsigned account facts. Database access control, backups, audit, and restore validation are external operational controls. |
| NATS duplication, reordering, loss, or forgery | Transactional outbox is committed with the domain fact; JetStream durable acknowledgement precedes `published_at`; event UUID is the dedupe key; consumers record consumption; security decisions read PostgreSQL | An outage delays email and other side effects, and a compromised broker can cause denial or traffic analysis. Preserve outbox and consumer state; never repair authority by editing broker messages. |
| Email-provider outage or malicious response | Registration commits with a generic result; delivery is asynchronous, idempotent, and bounded; provider errors/bodies are collapsed; `required` and `grace` authority rules remain in PostgreSQL | Delivery delay can prevent verification or recovery. The provider necessarily learns delivery addressing, and its own data handling is a production vendor risk. Verification must not be disabled as an outage workaround. |
| Dependency outage | PostgreSQL-dependent authority decisions close immediately; Redis-dependent challenges close; NATS outage accumulates an observable outbox; signer outage blocks only new issuance; reporter failure is non-blocking; existing immutable bytes remain readable | Coordinated or prolonged outages can deny all new authentication and issuance. Recovery order must preserve PostgreSQL authority and monotonic versions; capacity planning and disaster recovery are external. |
| Insider or compromised configuration signer | Domain-separated signing input, provider-neutral bounded interface, root-signed key metadata, publish-before-use rotation, fixed algorithms, client key/time/version validation, and production KMS/root-ceremony gate | An authorized active signer can sign malicious C1.1 payloads. Separation of duties, KMS policy, witnessed root changes, change review, signing audit, and short activation windows are mandatory production controls. Root compromise is a separate trust-root incident. |
| Local seed used in production | Production profile rejects local providers and missing HTTPS origins; documentation treats fixtures as non-production | A deployment that bypasses startup/release gates loses the security claim. A local seed is never a backup or production recovery mechanism. |
| Log, metric, HTTP, panic, command-output, or reporter leakage | Stable HTTP DTOs; compile-time event/component/result registries; bounded metric labels; raw errors and arbitrary maps excluded; asynchronous bounded reporter; secret-canary acceptance across output surfaces | Host, profiler, crash-dump, third-party agent, or operator misconfiguration can bypass application-level controls. Do not collect raw process, request, provider, database, or broker artifacts. |
| Parser confusion, oversized input, or algorithm downgrade | 64 KiB auth request limit, 1 MiB envelope limit, duplicate/unknown/trailing/invalid UTF-8 rejection, fixed field bounds, strict JCS, fixed Ed25519/HPKE suite, and no runtime algorithm negotiation | Resource exhaustion below limits remains possible at scale. Edge rate limits and capacity controls are required; later protocol versions need a separate migration design. |
| ErrorReporter outage, panic, or queue exhaustion | Nonblocking validated DTO, bounded queue/batch/timeout, drop/provider-failure counters, panic containment, and no recursive reporting | Reports may be intentionally lost under stress. Local finite metrics are the fallback; absence of a remote report is not evidence of health. |

## Security invariants during failure and recovery

- PostgreSQL remains the authority even when Redis, NATS, email, signer, reporter, primary, or mirrors are unavailable.
- No recovery may change a compromised, revoked, consumed, or higher-version fact back to an earlier state.
- New login, challenge, rotation, authorization, or issuance fails when its security dependency is unavailable; no in-process substitute is permitted.
- Existing, unexpired immutable bundle bytes may remain available during signer failure, but no source may regenerate or reserialize them.
- Root metadata for a replacement signing key must be published and accepted before that key signs a bundle.
- Evidence collection uses only the bounded signals listed in the runbooks and never collects request bodies, credentials, identifiers, locators, key material, ciphertext, connection strings, provider text, or raw errors.

## Validation and remaining release gates

Unit, race, fuzz, deterministic-generation, integration, privacy-canary, and Docker/E2E gates together substantiate this model; one class cannot substitute for another. C1.1 is not complete until the clean-tree `verify-c11` acceptance runs against real Docker PostgreSQL, Redis, and NATS, actual API/mirror processes, and the independent verifier, with every required stage available and passing. Release approval must also record completion of the atomic `family_id` replay-format cutover; this undeployed branch provides no rolling compatibility with legacy replay blobs.

Production release additionally requires a witnessed root ceremony, a real non-exportable external signer, an external sensitive-field protector, an external email provider, HTTPS origins, dependency backup/restore drills, and the relevant legal gates. Passing local-fixture acceptance does not satisfy those production controls.
