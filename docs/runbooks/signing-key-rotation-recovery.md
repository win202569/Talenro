# Signing-key rotation and recovery

This runbook governs C1.1 trust-root metadata and configuration-signing keys. It does not authorize a production key ceremony by itself. Production release remains blocked until the organization has approved a witnessed root ceremony, separated roles, a non-exportable KMS/HSM/Vault signing implementation, and an immutable audit destination.

Local root/configuration seeds and the local signer exist only for tests. Never copy, escrow, restore, promote, or derive a production key from a local seed. A local seed is not a production backup or recovery mechanism.

## Roles and decision boundary

- The incident commander controls containment and declares entry and exit.
- The signer custodian changes KMS/HSM/Vault policy but cannot approve root metadata alone.
- Root-ceremony approvers witness and approve every new metadata version under separation of duties.
- The database operator performs read-only authority checks and does not edit signing metadata by hand.
- The client/release owner verifies the independent client result before enabling new issuance.

No single operator may both introduce a production signing key and authorize it in root metadata. Treat unexplained successful signing, unexpected metadata advancement, or signing outside an approved change window as suspected misuse, not as an ordinary outage.

## Detection signals

Use only these bounded signals. Compare rates with the deployment's approved baseline; do not attach raw logs, HTTP bodies, signer responses, or signing input.

| Signal | Meaning | Limitation |
| --- | --- | --- |
| `talenro_crypto_validations_total{operation="bundle_sign",result="failure",reason="key_unavailable"}` | A real bundle-signing call failed, timed out, or the observed provider became unavailable | It intentionally collapses provider detail and does not identify a key or device |
| `talenro_crypto_validations_total{operation="bundle_sign",result="success",reason="none"}` | Completed signing activity | A success is not proof that the change was authorized; compare only aggregate activity with the approved window |
| `talenro_security_events_total{operation="bundle_resolution",result="failure",reason="dependency"}` | A resolution request reached a dependency/signing failure | It can include a non-signer dependency and is corroborating evidence only |
| Stable HTTP code `signing_unavailable` | New issuance is failing closed | Never retain the response body or request context |
| Root-metadata version and signing-key state from the approved authority view | Current monotonic authorization state | Record the version and state only; do not export public-key bytes or key identifiers into the incident record |
| Byte/hash probe of an already published test bundle at primary and mirrors | Existing immutable distribution remains intact | Use a designated synthetic fixture; never record a production locator or ciphertext |

An ErrorReporter record is not authoritative. If present, only the finite tuple (`request_failure` or `dependency_failure`, `cryptography` or `dependency`, `trust`, `failure`, fixed fingerprint) may be used as corroboration.

## Immediate containment

1. Declare one of: planned rotation, signer availability incident, suspected active-key misuse, or suspected root compromise. If uncertain, choose the more severe class.
2. Stop or gate **new bundle issuance** without deleting existing bundle bytes or reducing read availability. Do not change client highest-version state.
3. For suspected misuse, remove the workload's ability to invoke the signer and freeze signer policy changes. Preserve KMS/HSM/Vault audit records in their native immutable system.
4. Keep primary and both mirrors serving exact already-published immutable bytes. Do not regenerate, re-encrypt, reserialize, purge, or replace them.
5. Freeze unrelated metadata edits and deployments. Do not retry a root ceremony until the previous attempt's commit/abort state is known.
6. If root authority may be compromised, stop all new issuance and all metadata publication. Escalate to the separate trust-root recovery process; an application operator cannot repair root compromise.

## Authority checks

Perform read-only checks in this order:

1. Confirm PostgreSQL is available and identify the highest committed trust-metadata version and its finite key states (`future`, `active`, `retiring`, or `revoked`, as applicable). PostgreSQL is the service authority; caches and mirrors are not.
2. Confirm the candidate metadata is root-signed, strictly higher than the previous accepted version, within its validity interval, and uses only the fixed Ed25519 algorithm.
3. Confirm the intended signing key is non-exportable in the approved production provider and that its activation policy matches the witnessed change record. Record only an approved ceremony/audit record reference, never provider output or key material.
4. Confirm at least one independent reference client accepts the new metadata from a clean previous highest-version state before any bundle is signed with the new key.
5. Confirm the old key remains authorized for the complete grace interval required by already issued, unexpired bundles.
6. Confirm existing primary/Mirror A/Mirror B bytes remain byte-identical and readable without invoking the signer.

Redis, NATS, a mirror, local seed files, and the signer's current alias are not authority for metadata version or key state.

## Planned rotation sequence

1. Create a new non-exportable signing key inside the approved KMS/HSM/Vault boundary. Apply least-privilege invocation policy and keep application invocation disabled.
2. In a witnessed root ceremony, build metadata version `N+1` that adds the new key before use and retains the old key for the full overlap window. Validate schema, fixed algorithm, validity, and monotonic version independently.
3. Root-sign and publish the exact metadata bytes. Commit the authoritative record once; do not hand-edit a committed record.
4. Use the independent client to verify version `N+1` from the currently accepted version and prove rollback to `N` is rejected.
5. Enable application invocation of the new signer, switch the configured signer, and issue one designated synthetic test bundle at a higher bundle version.
6. Fetch the synthetic envelope from primary and both mirrors, require identical SHA-256, and independently decrypt, verify, and stage it. Do not collect the envelope as incident evidence.
7. Observe bounded signing-success and distribution signals through the approved soak window. Keep the old key available for verification, not new signing.
8. After every old-key bundle and supported client overlap window has ended, use a second witnessed ceremony to publish a strictly higher metadata version that marks the old key revoked. Verify the new version independently before disabling the old provider key.
9. Disable, then schedule destruction of, the old provider key according to the organization's retention policy. Destruction is a separately approved irreversible action.

## Availability recovery

For an outage with no indication of misuse:

1. Keep new issuance closed and existing immutable distribution online.
2. Restore connectivity, policy, quota, or provider health without changing metadata or switching to a local signer.
3. Re-run the authority checks above. A provider recovery does not authorize an unlisted key.
4. Perform one synthetic higher-version issuance and three-source/reference-client verification.
5. Reopen normal issuance only after failure counters stop increasing and the success path is demonstrated within the configured signer timeout.

If provider restoration requires a replacement key, follow the planned rotation sequence. Do not import an exported private key or reuse a key from another environment.

## Suspected signing-key compromise

1. Keep issuance disabled and revoke application invocation of the suspected key at the provider boundary.
2. Determine the last root-metadata version and last approved change window using authority and ceremony audit records, not application logs.
3. Create a fresh non-exportable key. Through an emergency witnessed root ceremony, publish a strictly higher metadata version that authorizes the replacement and marks the suspected key revoked.
4. Ensure clients receive and accept the replacement metadata before any replacement-key bundle is issued.
5. Reissue the current valid test configuration as a **new higher bundle version** for each affected authorization. Never overwrite an issuance or reuse a locator/envelope.
6. Require independent verification and byte identity across all three sources before reopening issuance.
7. Keep incident scope open for any bundle signed during the suspected window. Revocation prevents future trust; it cannot make already accepted malicious content disappear.

If the root key, ceremony system, or metadata publication authority may be compromised, this sequence is insufficient. Keep issuance stopped and execute the organization's out-of-band root-replacement/client-update procedure.

## Rollback prohibitions

Never:

- decrement or overwrite a trust-metadata version or bundle version;
- reactivate a revoked key by editing its row or republishing an old metadata version;
- sign with a new key before clients can trust metadata that authorizes it;
- use local seeds, an ephemeral process key, or another environment's key to restore production;
- restore a database snapshot without proving that metadata and issuance high-water marks are at least as high as every client-visible version;
- regenerate an immutable bundle under the same locator, version, or ETag;
- remove the old verification key before the defined overlap and expiry window ends.

If a new key behaves incorrectly, stop new issuance, fix forward with a higher metadata version and, where needed, a higher bundle version.

## Evidence handling

Retain only:

- incident times rounded to the approved operational precision;
- aggregate deltas for the bounded metrics listed above;
- metadata version numbers and finite key states;
- approved ceremony/change/audit record references;
- pass/fail results for synthetic three-source identity and independent verification;
- build/release identity already exposed by the bounded build metric.

Do not collect or paste root/configuration seeds, private or public key bytes, key identifiers, signing input, signatures, bundle locator or bytes, request/response bodies or headers, environment values, connection URLs, provider errors, stack traces, database rows, or process dumps. Preserve provider audit evidence in place and grant reviewers read-only access rather than exporting it.

## Exit criteria

Close the incident only when all applicable statements are true:

- the highest authoritative metadata version is known, root-verified, and no lower version is being accepted;
- the active production signer is non-exportable, explicitly authorized by that metadata, and permitted by a witnessed change record;
- unauthorized or suspected keys cannot sign, while required old verification keys remain available only for their approved overlap;
- one new synthetic higher-version bundle verifies independently and has byte-identical primary/Mirror A/Mirror B content;
- existing immutable content remained unchanged throughout recovery;
- bounded failure signals have returned to the approved baseline and successful signing is observed;
- no secret-bearing evidence was collected;
- any suspected-key window and client remediation have an assigned owner; and
- the production root-ceremony/KMS release gate remains satisfied. Local-fixture success never satisfies this criterion.
