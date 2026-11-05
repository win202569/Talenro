# Isolated writer — approved single compiler correction

Parent plan: 2026-11-04-runtime-isolated-writer.md; design unchanged.
Approval: user 是 after6c167190, specifically one versioned correction and one
additional compilation/verification process. Original stopped records stay frozen.

## Exact change and scope

Saved compilation5 failed with CS1690 at tests.cs174:56, before unit execution.
The test helper inherits MemoryStream and exposes ReadOnlyMemory<byte> as a field.
In runtime-isolated-writer-v3.tests.cs change only that declaration to a public
property with a private setter. U01..U18 bodies/expectations remain byte-identical.
Do not suppress warnings, relax assertions, change runtime/protocol or compile
the old adapter. The prior compiler diagnostic is saved failing evidence; no
extra reproduction/mutation/RED compilation is authorized.

Fresh directory: .superpowers/sdd/2026-11-05-runtime-isolated-writer-v3.
New role-v3/driver-v3 use fresh claims and computation32/compilation6, authenticate
the prior stop, and accept Compile only. Inherited scenario code is unreachable
through parameter validation; no scenario gate or child is started this turn.
Current31/10/5/217; active correction ceiling32/11/6/217.
Earlier nine scenarios remain NOT_RUN; their future continuation/renumbering is
not performed under this one-worker correction.

## Ordered execution

1. Verify prior37-file archive and all original input pins. Check exact one-line
   test change and four PowerShell parse results, original adapter50176bytes/hash.
   Save new input/approval/static checks once. Unexpected preparation stops.
2. Archive and decode-check sources/evidence, update recovery/status, commit/push
   explicit files; require remote==HEAD before invocation.
3. Invoke driver-v3 Compile once with actual input SHA. One owned worker,
   compilation6 once, U01..U18 once, unchanged120000/2000/2000ms waits and output
   limits. No grandchildren, other tests, new dependencies or security changes.
4. Read receipts independently; check claims, actual attempts, exit, channels,
   map identity and18case results. Do not rerun to compensate for failure.
5. Archive/push result with binary metadata only; leave original failures intact.
   Stop after this correction's outcome. NOT_ACCEPTED, no business host/main merge.

First unexpected preparation/compiler/unit/recording result stops without retry.
No cleanup/deletion or reviewer agents; preserve the user's finite-scope exceptions.
