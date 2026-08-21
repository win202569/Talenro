# Email-provider outage

This runbook covers failure, timeout, rejection, or sustained throttling of the production `EmailSender` used for verification and password-reset deliveries. Email is an asynchronous side effect: PostgreSQL account and delivery facts remain authoritative, and provider failure must not roll back a registration or disclose whether an identity exists.

Never work around an outage by setting production email verification to `disabled`, switching to the local sender, manually marking an address verified, or extracting a pending verification/reset token.

## Detection signals

Use only bounded event names, finite report tuples, and aggregate metrics. Do not inspect provider response bodies, NATS messages, email addresses, or pending delivery ciphertext.

| Signal | Interpretation |
| --- | --- |
| Registered event type `talenro.identity.email_delivery_requested.v1` and aggregate durable-consumer pending/redelivery counts for consumer `identity-email-delivery-v1` | Delivery work exists and is not completing normally; use counts/age only, never message payloads |
| ErrorReporter tuple `worker_failure`, `dependency`, `email`, `failure`, `dependency` | The email intake/worker encountered a sanitized dependency failure |
| `talenro_error_reports_total{component="email",result="sent"}` | The bounded report about an email-worker failure reached the reporter provider; this counts telemetry delivery, not successful email |
| `talenro_error_reports_total{component="email",result="dropped"}` or `result="provider_failed"` | Email-worker telemetry was dropped or its reporter provider failed; do not infer email-provider health from this metric alone |
| `talenro_outbox_backlog` and `talenro_outbox_oldest_seconds` | Distinguishes an upstream outbox/NATS publication problem from a downstream email-consumer/provider problem |
| Generic registration/reset-delivery HTTP result and account-registration/recovery aggregate rates | Confirms the public non-enumerating contract remains stable; it does not prove delivery success |

Alert thresholds are deployment policy. Keep alert dimensions to the exact finite component/result/event registries above; never add provider text, recipient, template, event ID, or route text as a label.

## Immediate containment

1. Declare the incident and classify it as provider rejection, provider timeout/unreachable, provider credential/policy failure, or unknown. Do not copy the provider error into the case.
2. Confirm that PostgreSQL-backed registration/reset-delivery requests still return the expected generic result and that account authority is not being rolled back.
3. Keep `required` and `grace` policies unchanged. Under `required`, do not issue enrollment grants before verification. Under `grace`, retain the fixed provisional single-device/24-hour restriction.
4. Keep the durable email consumer and its retry/Nak behavior intact unless retries threaten a documented provider quota. If intake must be paused, cancel it cleanly; do not Ack messages after cancellation.
5. Do not delete or replay outbox rows, JetStream messages, consumed-event IDs, idempotency records, or protected pending-delivery values.
6. Restrict provider credential changes to the secret manager/adapter owner. Rotate credentials if compromise is suspected, but never print or validate them through a shell transcript.
7. If the reporter is also failing, continue the business path and use local aggregate metrics. Reporter failure is not a reason to stop or alter authority transactions.

## Authority checks

Perform read-only aggregate checks in this order:

1. PostgreSQL is available and account/email states match the configured verification policy. An account is verified only by the committed one-time verification transition.
2. Pending verification/reset deliveries have nonzero bounded lifetime and protected recoverable material; collect only counts and oldest age by finite delivery kind.
3. The corresponding transactional outbox facts exist. A committed registration remains valid even if publication has not occurred.
4. If NATS is healthy, the durable email consumer retains pending/redelivered work. Event delivery is at least once and event ID is the dedupe key.
5. For completed events, the same short PostgreSQL transaction records consumption and clears the corresponding pending protected delivery. Neither half may be considered success alone.
6. Public HTTP responses remain generic for existing and nonexistent identities. Do not use delivery existence or timing to answer a support request about account existence.

The email provider is not authority for account state, verification, token consumption, or deduplication. Provider dashboards and delivery receipts may be used by the provider team under their own privacy process, but are not copied into this incident.

## Recovery sequence

1. Restore the approved external provider adapter, network path, quota, policy, or secret-manager credential. Do not switch production to a local fixture.
2. Run a provider-owned synthetic delivery that contains no real account data and record only pass/fail. This checks the provider boundary, not application authority.
3. Resume the durable consumer. Let it redeliver using the original event IDs; do not publish replacement messages by hand.
4. For each successful delivery, allow the consumer's normal short transaction to atomically record the consumed event ID and clear the matching protected pending value.
5. Confirm duplicate delivery of the same event ID produces no second provider side effect. A Nak/retry before success is expected; an Ack occurs only after successful processing.
6. Watch aggregate consumer pending/redelivery counts fall and verify the finite email-worker failure tuple stops increasing.
7. Confirm new registration, verification-delivery, password-reset-delivery, and subsequent verification/reset flows complete with generic public behavior.
8. Allow expired pending deliveries to remain expired. A user requests a new delivery through the normal idempotent API; operators do not extend or reuse an old token.

If the provider was compromised, finish credential rotation and provider-side revocation before step 2. If delivery content or addressing may have been exposed, follow the privacy incident process in parallel without extracting application data for this runbook.

## Rollback prohibitions

Never:

- mark an email verified, complete a password reset, or issue an enrollment grant based on a provider receipt;
- change production verification to `disabled`, widen `grace`, or extend a verification/reset token to restore availability;
- roll back a committed account because its email was not sent;
- delete an outbox row, durable message, consumed-event ID, pending-delivery record, or idempotency tombstone to force retry;
- Ack failed/canceled work, publish a copied payload, or change an event ID;
- decrypt or copy pending delivery material into a support tool;
- paste a provider error, response, request, credential, recipient, or message body into logs or evidence.

Recovery always moves forward through the existing event and one-time-token contracts. Expired work is replaced by a newly authorized delivery, never revived.

## Evidence handling

Retain only incident start/end times, finite incident class, aggregate event pending/redelivery counts, aggregate pending-delivery counts/oldest age by delivery kind, the bounded email-worker report tuple, bounded metric deltas, build/release identity, and synthetic/recovery pass/fail results.

Do not collect email addresses, template parameters, verification/reset tokens or digests, pending ciphertext, event IDs or payloads, NATS messages, account/principal/session identifiers, request/response bodies or headers, provider bodies/errors/message IDs, credentials, URLs, environment values, database rows, logs, traces, stack dumps, or IP data.

## Exit criteria

Close the incident only when:

- the approved external provider passes a data-free synthetic check;
- the durable consumer is running, pending/redelivery aggregates return to the approved baseline, and no work was manually deleted or acknowledged;
- successfully handled events have one consumed-event fact and no corresponding pending protected delivery;
- duplicate event delivery produces no duplicate provider side effect;
- new registration and reset-delivery requests remain generic, while real verification/reset completes through the normal one-time flow;
- `required` and `grace` authority rules were never weakened and no expired token was revived;
- bounded worker-failure signals stop increasing and reporter failure, if any, did not affect business/readiness behavior;
- all credential/provider changes are captured in the approved secret-management audit trail; and
- no secret, recipient, stable identifier, message content, or raw provider artifact was collected.
