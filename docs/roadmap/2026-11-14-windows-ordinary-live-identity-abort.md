# Ordinary live observation — strict identity abort

Date2026-11-14 Asia/Tbilisi; source checkpoint868899b4. User authorized one
host-context launch, zero retries, max360000ms. Original entry unchanged.

Run29a966a471034fb5bb838febb4a2f981, PID12632. Preflight passed; one target acquired.
Start291.6522ms; native identity922.6956ms, IdentityVerified=false. Persisted
outcome1161.1549ms: identity-failed, callback-or-evidence-error, ExitVerified=false,
Finalized=false. Console abort1171.6143ms, session-failed, ExitVerified=false,
OverrunMs0; child PowerShell exit1. No rerun, no fallback or identity relaxation.

Read-only CIM check of original PID: actual CommandLine183 characters versus
quoted executable expected182; exactly expected plus one U+0020 trailing space.
This violates the explicit fixed no-trailing-space contract. No product patch
or assumption that trimming command lines is safe was made.

Initially no targetable window; later exactly one ordinary candidate window
title Talenro appeared. Screenshot/accessibility show Chinese offline demo Login,
read-only demo@example.com and masked demo password, Enter demo button, no
LAYOUT TEST marker. This was cleanup observation, NOT a completed UI acceptance
route. Did not enter demo, navigate Account, resize, connect, activate Delete,
or perform keyboard acceptance. Exact viewport/RTL/other coverage NOT_RUN.

Clicked observed titlebar Close normally, well before5min. Successful subsequent
CIM inventory returned0 target/prototype processes; subsequent window list empty.
This establishes post-close absence only, NOT a verified original-handle exit:
the observer had already aborted/disposed its resources. ExitVerified remains
false and no target exit code is claimed. No forced termination or system change.

Original ignored receipts retained under candidate ordinary-account-observations/
29a966a471034fb5bb838febb4a2f981/:
- manifest.json SHA256 C3111D927A825EC9A4238BFCBE37A1832AAC83AEFE092864FC07D0780E2F355E
- samples.jsonl SHA256 CEB39C97913AE3E5E1145C19A7232CEFB7D58C4643EE9F4A36397709BE3028CB

Launch allowance spent. Next proposed scope: read-only inspect host no-argument
serialization and prepare a precise binding change with synthetic tests, not
implement or relaunch without approval. Keep strict identity; do not automatically
Trim, accept arbitrary suffixes, or use another artifact. Runtime/C4/security/
Release acceptance unchanged. Clone restores sanitized record, not raw receipts.
