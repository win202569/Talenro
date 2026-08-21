# C1.1 dependency failure

This runbook covers PostgreSQL, Redis, NATS/JetStream, EmailSender, ConfigSigner, ErrorReporter, and primary/mirror distribution failures in C1.1. It refines the foundation runbook: `/livez` represents the process, while `/readyz` reports only the `postgres`, `redis`, and `outbox` components using the finite states `up`, `degraded`, and `down`. A component can be degraded without making the whole process unready when the documented authority semantics remain safe.

Use the signing-key and email-provider runbooks for the deeper provider-specific recovery. Use the token-family runbook if a failure reveals replay or theft.

## Safe detection signals

Use only the following finite surfaces. Query live health without archiving its body, and record only the named component and finite state.

- `/livez` HTTP status and `/readyz` HTTP status, plus only the `postgres`, `redis`, and `outbox` states from `{up,degraded,down}`. NATS has no separate readiness component.
- `talenro_outbox_backlog` and `talenro_outbox_oldest_seconds`.
- `talenro_security_events_total` with the implemented finite `operation`, `result`, and `reason` labels; dependency failures use `result="failure",reason="dependency"`.
- `talenro_crypto_validations_total`, especially `bundle_sign/failure/key_unavailable`; use only registered operations/results/reasons.
- `talenro_error_reports_total{component,result}` with registered components and `sent`, `dropped`, or `provider_failed`.
- ErrorReporter's finite event/category/component/outcome/fingerprint tuple, when available. Never include a raw error.
- Registered event type and aggregate count/age only; never inspect NATS payloads.
- A designated synthetic primary/Mirror A/Mirror B byte-hash and independent-verification pass/fail result. Never use or retain a production locator or envelope.

The HTTP request metric collapses application routes to bounded values and is useful for broad availability only. Do not add dependency name, URL, host, exception, identifier, or attacker-controlled text as a label.

## Failure matrix

| Dependency | Expected authority/readiness behavior | Immediate containment | Safe recovery proof |
| --- | --- | --- | --- |
| PostgreSQL | `/livez` remains 200; readiness is `down`; login, token validation/rotation, grants, registration, authorization, issuance, and acknowledgement fail closed | Stop security mutations and new issuance; keep the process for health only; preserve the database and backups | PostgreSQL is `up`; authority high-water marks are monotonic; an authenticated synthetic flow succeeds without undoing revocation |
| Redis | Existing short access tokens may still be validated against PostgreSQL; new login, challenge, ceremony, rotation, resend/rate-dependent work fails closed; initial failure is `degraded`, sustained failure reaches `down` | Do not replace challenges/nonces with process memory; let clients receive stable retry/dependency behavior | Redis is `up`; stale challenges cannot be consumed; a newly created challenge succeeds exactly once |
| NATS/JetStream | Committed domain/outbox facts can succeed and the publisher retries; readiness changes only through the `outbox` component when its configured age/backlog thresholds reach degraded/down, never through a synthetic `nats` component | Preserve PostgreSQL outbox and JetStream storage; pause nonessential producers only by approved policy; do not edit/republish payloads | Fixed `FileStorage` stream is healthy; durable PubAck resumes; backlog/age drain; duplicate event IDs cause no duplicate side effect |
| EmailSender | Registration/reset-delivery remains generic and committed; worker retries/deduplicates; verification authority is unchanged | Keep `required`/`grace`; never switch to local/disabled or expose pending material | External provider and durable consumer recover; one consumed-event fact and pending-field clear occur together; duplicate delivery is harmless |
| ConfigSigner | New issuance fails with `signing_unavailable`; existing unexpired immutable bytes remain readable; no local fallback | Gate new issuance; retain primary/mirrors; freeze key changes if misuse is possible | Active signer is authorized by higher/root-signed metadata; one higher-version synthetic bundle verifies at all three sources |
| ErrorReporter | Business paths and readiness are unchanged; queue can drop and finite drop/provider-failure metrics increase; no recursive report | Do not block requests or expand logs; preserve local bounded metrics | Provider recovers or queue drains within bounds; business/readiness behavior never changed |
| Primary bundle location | Resolution may return locations, but primary GET is unavailable; signer and authority are unaffected | Do not regenerate bytes or locator; direct clients to normal multi-source retry behavior | Mirror A/B retain identical bytes; restored primary serves the same stored hash |
| One mirror | Remaining two sources continue; no readiness/authority transition | Isolate only the failed mirror; retain only its read-only byte-store role and do not grant signer or account/device database authority | Restored mirror reads the byte store and matches the existing primary hash without regeneration |
| Both mirrors | Primary remains a source; redundancy is reduced | Freeze mirror deployment changes and preserve primary immutable storage | Each mirror independently restores and matches the primary/existing expected hash |

## Immediate incident sequence

1. Confirm `/livez`, then read `/readyz` once without retaining a body. Record only component/state pairs and status code.
2. Name the affected dependency from the finite matrix. If multiple components fail, assume PostgreSQL authority is unavailable until proved otherwise.
3. Stop actions that require the failed authority dependency. Do not disable authentication, verification, proof, signature, or monotonic-version checks to improve availability.
4. Preserve PostgreSQL volumes/backups, JetStream FileStorage, immutable bundle stores, idempotency/security history, and provider audit records in place.
5. Keep unrelated safe behavior operating: liveness, existing immutable reads during signer failure, PostgreSQL-backed access validation during a short Redis outage, and committed domain writes during a NATS outage as specified above.
6. Use only stable public errors such as `dependency_unavailable` or `signing_unavailable`. Never expose or collect the underlying failure.
7. If compromise rather than outage is plausible, move to the stricter signer or token containment runbook before recovery.

## Authority checks before restart

1. **PostgreSQL:** verify availability, migration version, account/device authority states, token-family revocations, idempotency tombstones, trust-metadata highest version, bundle issuance highest versions, and outbox high-water state through approved bounded views. Record states/counts/versions only.
2. **Redis:** treat all remaining data as disposable coordination, not truth. Confirm no procedure relies on Redis to restore revocation, account state, authorization, or bundle version.
3. **NATS:** confirm the named C1.1 stream exists with the exact approved subjects, limits, retention, replicas, and `FileStorage`. Confirm the durable consumer identity rather than creating a replacement with different semantics.
4. **Signer:** confirm provider health and that the current key is authorized by the highest root-signed metadata. Provider alias/current state alone is insufficient.
5. **Email:** confirm delivery work is tied to committed outbox/event facts and consumption has not been manually advanced.
6. **Distribution:** compare only a designated synthetic expected hash across surviving sources. A mirror or CDN cache never determines the expected bytes.
7. **Reporter:** confirm its failure did not alter authority transactions, HTTP outcomes, readiness, or local bounded metrics.

## Recovery order

Recover in dependency order; prove each layer before enabling the next.

1. **PostgreSQL first.** Restore the existing authoritative instance or an approved backup. Before traffic, prove migrations and every security/trust high-water mark are not below a client-visible or audit-recorded value. If that cannot be proved, keep security mutations and issuance closed.
2. **Control API authority reads.** With Redis/NATS still considered unavailable, prove revoked/compromised tokens remain rejected and immutable existing bytes are unchanged.
3. **Redis.** Restore the approved instance. Permit only newly created challenges/ceremonies; never reconstruct or extend old entries. Prove a fresh challenge is consumed once and an old/unknown challenge fails.
4. **NATS/JetStream.** Restore the exact persistent stream and durable consumers. Resume publisher intake and require durable PubAck before outbox rows are marked published. Let backlog drain using original event UUIDs and prove duplicate delivery is idempotent.
5. **EmailSender.** Restore the approved external adapter and follow the email-provider runbook. Do not widen verification policy while queued work drains.
6. **ConfigSigner.** Restore the authorized external signer or complete publish-before-use rotation under the signing runbook. Issue only a new higher-version synthetic bundle.
7. **Primary and mirrors.** Restore byte-store readers independently. Require byte identity to the already committed expected hash; do not regenerate.
8. **ErrorReporter last.** Restore telemetry delivery without making it a readiness or business dependency. Observe bounded `sent`/`dropped`/`provider_failed` outcomes.
9. Reopen normal traffic gradually and watch only the finite health, outbox, security, crypto, and reporter aggregates.

For a single isolated dependency, earlier healthy layers need not be restarted. The order still describes what must be authoritative before that dependency is trusted again.

## Dependency-specific cautions

### PostgreSQL

Do not delete a database volume, use `docker compose down --volumes`, perform an unbounded restore, or write corrective SQL during incident triage. Database restore is a separately approved destructive/recovery action. A restore that lowers trust metadata, bundle issuance, revocation, or idempotency high-water marks is not acceptable even if readiness becomes green.

### Redis

Do not promote Redis data to authority, replay a saved Redis dump to revive challenges, or use in-process maps as fallback. If Redis data must be cleared, use the environment-specific approved procedure and expect all outstanding challenges/ceremonies to fail and be recreated normally.

### NATS/JetStream

Do not recreate the stream with memory storage, altered subjects, weaker acknowledgement, or a new durable name to bypass pending work. Do not manually mark PostgreSQL events published. At-least-once duplicates are repaired by consumer deduplication, not by deleting transport history.

### Signer and immutable distribution

A signer outage does not justify replacing already published bytes. A mirror may receive only its listener setting and a dedicated least-privilege byte-store database credential; it must never receive a signer, identity/device authority, provider, Redis/NATS, or key credential. Recovery is valid only if a restored source serves the exact previously committed bytes.

### Reporter

Do not turn on raw exception, request, body, header, environment, or provider logging because reports are unavailable. Reporter drops are an explicit bounded failure mode.

## Rollback prohibitions

Never:

- reactivate revoked/compromised authority, unconsume a grant/challenge/event, or delete an idempotency tombstone;
- lower a trust-metadata, bundle issuance, client highest-trusted, account/device state, or migration version;
- restore an old database merely because it starts faster;
- Ack or delete failed events, mark outbox rows published without durable PubAck, or change event UUIDs;
- weaken production profile, TLS, email verification, signer, field-protection, or reporter-provider gates;
- switch to local signer/seed, local field protector, local email sender, memory JetStream, or in-process Redis substitutes;
- regenerate or reserialize a bundle under an existing locator/version/ETag;
- use wildcard process termination, volume deletion, `docker system prune`, or unbounded cleanup.

Fix forward with monotonic state and newly authorized events/versions.

## Evidence handling

Retain only incident times, process/build identity, `/livez` and `/readyz` status codes, finite component/state pairs, bounded metric deltas, registered event types with aggregate counts/age, authority state/count/version summaries, approved provider/backup audit references, and synthetic pass/fail results.

Never collect environment variables, connection URLs, service credentials, Docker inspect output, volume contents, database rows, Redis values, NATS payloads, email/provider bodies, raw errors, stack traces, process dumps, HTTP bodies/headers, account/principal/device/session/event/trace identifiers, IPs, tokens, nonces, keys, locators, bundle plaintext/ciphertext, DNS, destinations, or traffic data. Keep native KMS/database/provider audit evidence in place rather than exporting it.

## Exit criteria

Close the incident only when:

- `/livez` is healthy and every applicable readiness component has the expected finite state, with no unsafe dependency silently omitted;
- PostgreSQL authority and all migration/security/trust high-water marks are verified nondecreasing;
- Redis-dependent operations use fresh single-use state and revoked tokens remain rejected independently of Redis;
- NATS uses the exact approved persistent stream, durable PubAck precedes publication marking, backlog/age return below approved thresholds, and duplicate delivery is harmless;
- email, signer, reporter, and distribution satisfy their row in the failure matrix and their specific runbook where applicable;
- one synthetic end-to-end flow authenticates, enrolls, resolves, independently verifies, and acknowledges a higher-version test bundle;
- primary and both mirrors serve byte-identical immutable content after independent restart;
- stable HTTP errors and bounded telemetry contain no raw dependency detail or secret;
- no prohibited rollback, manual message/database repair, local-provider fallback, or destructive cleanup occurred; and
- any production release gate that was not actually exercised remains blocked rather than being declared passed.
