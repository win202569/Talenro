# Versioned execution binding approval

User replied 是 after checkpoint0f07d4c0 to the explicit proposal: preserve old
records, authenticate their identities, create a versioned execution entry with
a distinct failure record, and continue within the original remaining budget.

Q/run-history-contract-runtime-v2.ps1 retains the original runner unchanged.
It requires exact prepared input/test/approval and historical parser-stop hashes,
verifies immutable history and original helpers, rejects preparation-v2 and
execution-v2 stops, and retains exclusive intent/receipt writes and25/26 limits.
New failures use P/history-contract-runtime-execution-v2.stop.json. The old stop
is historical evidence, not a successful preparation receipt. Its original
terminal exit/output remains unavailable; fresh output verification is documented.

No old attempt replay, new allowance, process-safety relaxation or main merge.
Versioned runner hash is recorded in actual auxiliary intents. Before RED25,
parse and archive the entry, checkpoint and verify remote equality. The fixed
68-pin inventory is unchanged: original helper stays pinned; new entry identity
is bound through the actual intent. Future gate must authenticate this identity.
First unexpected result stops remaining work with all actual records retained.
