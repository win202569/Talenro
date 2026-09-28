# Development-tool cache binding: approved fix and preparation

This continues the [real-verification checkpoint](2026-09-28-devtools-real-verification-checkpoint.md).
The user explicitly approved the Important cache-binding repair and its exact
147-entry cache-preparation list. No additional dependencies, push, merge or
restoration of the existing security gate is included.

## Implementation and observed regression

The three PowerShell consumers now select the verifier's fixed Go toolchain
directory and cache before any tool verification/execution. Inherited Go
configuration is replaced by the verifier's offline, readonly configuration;
module/overlay overrides still fail before work. The fixed directory also leads
the child PATH so Buf's local protoc plugin uses that Go. Original GO variables
and PATH are restored in finally. Root working directories and explicit tool
modfile arguments remain unchanged, as do legacy smoke/C12 source and time limits.

`TestDevtoolsConsumerCacheBinding` first failed for all three consumers:
42.140s, exit 1. Its actual child traces showed the foreign cache, foreign Go
executable and inherited proxy/workspace/toolchain settings after successful
verification of the fixed cache. The repaired test passed in 42.834s, exit 0,
including the nested plugin resolved through PATH. This is behavioral routing
evidence with fake external tools, not actual Buf/lint acceptance.

Additional environment-restoration coverage passed (7.83s); six missing-lock and
module-override consumer cases passed (28.31s). The original C11 privacy contract
needed a test-fixture adaptation because PATH substitution is intentionally no
longer honored. Its native fixed-location double forwards to the original batch
doubles with the existing C11 shell-metacharacter quoting boundary. An unwired
backend and two incorrect forwarding paths were preserved as failed harness
runs; the final unchanged assertions passed in 20.623s, exit 0. No production
fake-tool switch or verifier bypass was introduced.

The first combined run exited 1 in 395.166s: its only failing child was
`TestDevtoolsVersionLock/tools`, reporting sqlc package-map drift. The invocation
had inadvertently added `CGO_ENABLED=0`, unlike the original baseline/finalization
environment. Two bounded, strict offline metadata queries isolated the cause:
CGO=0 returned 573 packages, omitting `github.com/pganalyze/pg_query_go/v6/parser`
and `runtime/cgo`; the original default returned all 575 with zero differences.
No production or baseline change was needed. This failure is retained as a
test-invocation error, not attributed to cache preparation or the repair.

The complete original selector then passed in its original environment:
exit 0 / 379.127s, 1036 PASS markers, 0 FAIL, one existing inapplicable Skip,
under the unchanged 10-minute budget and without timing instrumentation. All
eight package maps and the version/evidence contract passed; 52 protected file
hashes were also checked unchanged. Four edited PowerShell files parsed and
the whitespace check passed. The Important cache-binding finding is repaired
with RED/GREEN and covering regression evidence. Real tool integrity, generation
and lint remain separate pending checks; no broad acceptance is claimed.

## Exact cache preparation

Preparation ran from 2026-09-28T13:10:01.4272126Z to
2026-09-28T13:11:03.7086853Z: exit 0, 62.288s, within one 900s deadline.
All 147 returned identities were compared against the approved committed list.

| Preparation mode | Count | Scope |
| --- | --- | --- |
| Existing archives, offline | 120 | Extract the selected version; no network fallback |
| Official proxy and checksum service | 27 | Prepare only the listed missing version |

Each result supplied the exact module/version, Sum, GoModSum and an existing
extracted directory. An isolated scratch module was used; downloaded contents
were not executed. Four production lock hashes remained unchanged. These are
preparation facts, not a successful whole-cache integrity result.

Full ordinary tests, real full C11/C12/bootstrap, 13 physical gates and I2/I3
remain blocked. The earlier 875.905s root integrity pass, compile-only pass,
historical readiness failure and prior review exclusions remain as recorded;
no cold-state stability or broad security acceptance is inferred.
