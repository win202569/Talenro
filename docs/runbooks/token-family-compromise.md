# Token-family compromise

Use this runbook when an account or device refresh token is proved to have been reused after rotation, when a bearer token is reported stolen, or when authority data indicates a refresh family is compromised. Account-session and device-token families are separate credential domains; investigate and recover each explicitly.

A token value is never diagnostic evidence. Do not request, receive, log, compare, hash, or paste a token, challenge, proof, idempotency key, request body, or authorization header during this procedure.

## Detection signals

The following finite records and aggregates are permitted. A generic authentication failure alone is not proof of compromise.

| Signal | Interpretation |
| --- | --- |
| PostgreSQL security category `account_refresh_replay` with fingerprint `identity.account_refresh_replay` | The account-session rotation path proved reuse and committed the fixed compromise transition |
| PostgreSQL security category `device_token_replay` with fingerprint `deviceauth.refresh_replay` | The cross-domain participant recorded a proved device refresh replay |
| Registered event type `talenro.deviceauth.token_family_compromised.v1` | A committed device-family compromise is awaiting or has undergone at-least-once publication; use the event type and count only, never its payload |
| `talenro_security_events_total{operation="device_authentication",result="failure",reason="invalid_credential"}` | Supporting aggregate for failed device rotation/authentication; it is not specific to replay |
| `talenro_security_events_total{operation="account_authentication",result="failure",reason="invalid_credential"}` | Supporting aggregate for failed account authentication; it is not specific to replay |
| `talenro_crypto_validations_total{operation="token_verify",result="failure",reason="invalid"}` | Supporting proof that token verification failed, without revealing token material |

Use the fixed database classification as the compromise decision when it exists. NATS delivery, Redis contents, a client's claim, and metrics are not authority. If the report is only “my token may have been copied,” treat it as a suspected theft and proactively revoke the relevant scope; do not manufacture a replay to prove it.

## Immediate containment

1. Establish whether the suspected credential is in the **account** or **device** domain. If unknown, contain both domains for the principal through approved high-risk revocation.
2. Confirm PostgreSQL availability. If it is unavailable, keep affected authentication/rotation and new grants closed; do not substitute Redis or cached state.
3. Revoke the affected family through the application command. A proved replay should already have atomically marked the family/session compromised and revoked active descendants; repeat only an idempotent authority check or approved broader revocation.
4. For an account-family incident, revoke other account sessions when theft scope is unknown or primary credentials may be exposed. Preserve the compromised state and security history.
5. For a device-family incident, revoke the device authorization when the device or its signing key may be controlled. This also prevents new device tokens and bundles.
6. Do not issue replacement refresh material into the affected family. Do not clear Redis to “unstick” rotation.
7. If aggregate failures indicate a campaign rather than a single family, activate bounded rate controls by operation, not identity, and escalate without collecting source IPs or account identifiers into this incident record.

## Authority checks

Use an approved read-only authority view and record only finite states/counts:

1. The family or account session is `compromised` or `revoked`, with a monotonically advanced state version.
2. Every active refresh descendant in that family is revoked or otherwise unusable.
3. Any affected access token is expired or rejected because the authoritative session/authorization is no longer active.
4. The associated account/device authorization state is consistent with the chosen containment scope.
5. Exactly one fixed security classification exists for the proved replay transition, subject to its application idempotency contract.
6. For a device family, the registered compromise event exists in the transactional outbox and publication/consumption state is consistent. A delayed event does not undo the committed revocation.
7. Idempotency tombstones remain present for the security mutation. A completed response may be replayed only to the original identical request; it must not repeat the mutation.

PostgreSQL is authoritative for all seven checks. Redis proves neither revocation nor safety. NATS is at-least-once transport and may legitimately contain a duplicate event.

## Scope decision

Choose the narrowest safe recovery based on authenticated evidence, not on convenience:

- **One account refresh family, device possession intact:** revoke that family; require fresh primary/high-risk authentication before creating a new account session.
- **Account password or strong factor may be exposed:** revoke all account sessions, rotate the affected primary credential/factor, then create a new session.
- **One device token family, device keys believed intact:** revoke the family and authorization, then require account-authenticated reenrollment. Do not attach a new family to the old authorization.
- **Device signing or HPKE private key may be exposed:** revoke the device authorization and reenroll as a new device identity with both keys freshly generated on the device.
- **Scope cannot be established:** revoke all account sessions and device authorizations reachable through approved application interfaces, then recover account authority before any device reenrollment.

## Account recovery sequence

1. Complete high-risk reauthentication with an uncompromised factor. If no factor remains trustworthy, use the separately approved account-recovery channel.
2. Revoke the affected session/family, or all sessions when scope requires it. Verify the authority transition before issuing anything new.
3. If password theft is possible, change the password and invalidate prior session authority. If a Passkey, TOTP seed, or recovery-code set is suspect, revoke/rotate that factor independently.
4. Create a new account session and refresh family. Return secret material only to the authenticated client; operators must never handle it.
5. Confirm the old family remains compromised/revoked and cannot rotate, even after Redis or process restart.
6. If devices were also revoked, proceed with the device sequence only after account recovery is complete.

## Device reenrollment sequence

1. Revoke the old device authorization and its token family. Preserve the device and compromise history; do not reactivate it.
2. From a newly authenticated account session, request a new one-time enrollment grant.
3. Generate fresh Ed25519 and X25519 key pairs on the target device. Never copy the old device private keys or generate them on the server/operator workstation.
4. Obtain a new single-use challenge, complete proof of possession, and atomically consume the grant to create a new device authorization and family.
5. Resolve and obtain a new test bundle for the new authorization. The bundle must use a higher server issuance version for that new authority where applicable; no old locator or envelope is reused.
6. Independently verify the bundle and persist its highest trusted version before acknowledgement.
7. Confirm the old authorization, family, tokens, bundle audience, and keys cannot be used by the new device or restored by retry.

## Pre-release `family_id` replay-format cutover

This branch has not been deployed. Every newly encoded device-token idempotency replay requires the authoritative `family_id`, and the new decoder strictly rejects a legacy blob with a missing, malformed, or nil value. There is intentionally no rolling mixed-version compatibility.

Before deployment, quiesce device registration and rotation writes, then either wait for every legacy completed idempotency replay to leave its approved retention window or execute a separately approved bounded cleanup that targets only those legacy replay bodies and preserves security history and tombstones. Upgrade every replay reader and writer atomically, verify new-format replay and family-compromise behavior, and reopen traffic only after the authority checks pass. Once a new-format replay exists, do not roll any API or worker back to the legacy format; recover forward with the same or a newer compatible binary.

## Rollback prohibitions

Never:

- change `compromised`, `revoked`, or consumed authority back to `active`;
- delete the replay security record, outbox event, consumed-event record, or idempotency tombstone to permit retry;
- issue a child token in the compromised family or copy a completed replay body into a new family;
- roll an API or worker back to the legacy replay format after a required-`family_id` record can exist;
- extend an old refresh absolute deadline;
- accept the wrong token domain as a recovery shortcut;
- bypass the single-use challenge because Redis is unavailable;
- reuse a device private key, grant, challenge, authorization, locator, or immutable envelope during reenrollment;
- lower a client's highest trusted metadata or bundle version after compromise.

If containment was broader than necessary, recovery still proceeds by creating new authority. It does not undo the revocation.

## Evidence handling

Retain only incident time, credential domain (`account` or `device`), finite authority states, aggregate affected-family counts, fixed security category/fingerprint, registered event type/count, bounded metric deltas, and pass/fail recovery checks. Use a separately controlled case reference to connect support workflow to the incident; do not place a principal, device, family, session, event, or trace identifier in shared evidence.

Do not collect tokens, token digests, passwords, factors, recovery codes, grants, nonces, proofs, device or principal identifiers, public/private keys, bundle locators or bytes, email addresses, IP addresses, request/response bodies or headers, raw database rows, NATS payloads, Redis values, environment values, connection strings, or raw errors.

## Exit criteria

Close the incident only when:

- every affected family is authoritatively compromised/revoked and has no usable descendants;
- the chosen account sessions and device authorizations are revoked with monotonic state versions;
- old access/refresh material fails through the correct token-domain boundary after process and Redis recovery;
- required account credentials/factors have been recovered or rotated in the correct order;
- every affected device is either intentionally revoked or reenrolled with fresh device-generated Ed25519/X25519 keys and a new authorization/family;
- a reenrolled device independently verifies a newly resolved immutable bundle and no highest-version barrier was lowered;
- at-least-once duplicate compromise events cause no duplicate side effect;
- bounded security/crypto signals have returned to the approved baseline; and
- no secret or stable identifier was collected as evidence.
