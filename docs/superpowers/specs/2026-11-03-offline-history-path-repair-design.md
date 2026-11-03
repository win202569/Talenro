# Historical path verification — bounded local repair design

Status: DRAFT / PENDING_APPROVAL. User requested continuation after the stopped
extent contract attempt. This document proposes local repair only; no execution
allowance is consumed or inferred. Baseline0ed80d444037f0529068a395227dc761e608c3e0.

## Intent and evidence

The saved Prepare error points to Driver-ExtentContractGate's HistoricalFiles loop.
The map contains2307keys, including three pinned C:/ absolute paths. Preparation
and auxiliary regression read each key directly; the actual driver unconditionally
joins it to the D:/ worktree root. The saved error shows the resulting invalid
D:/.../C:/...pwsh.exe path. No new freeze/host was created. Ten stops, including
extent-contract.stop.json SHA
C26E23A5783CC4CC6062F9142000B2D68931D1C990AE30C0E5986800B222029F,
remain authoritative. See the separate extent-contract-rebind-result report.

## Approaches

Recommended small repair: one shared production historical-file verifier in
driver.ps1, called by the actual gate and exercised directly by new local tests.
This closes the observed difference between auxiliary checks and production use.

A one-line inline path substitution would be smaller, but a separate test-only
implementation could miss the same integration boundary again. A broader portable
path-schema migration would alter frozen identities and is outside this repair.
Reopening the stopped epoch now is excluded. The continuation request follows
the prior recommendation to fix path parsing and add corresponding checks; exact
local budget and written design still need approval.

## Production boundary

Only L/driver.ps1 may change, where
L=docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2.
Current driver SHA
A5A09F0C8D36DB0CBD56747C90E3D2DB8741A7018EF1B7FA7F21E775E94ECE81.

Extract a single history verification function consumed by the existing gate.
For a repository-relative key, canonicalize under the explicit worktree root and
require containment; reject rooted-relative, drive-relative, traversal escapes,
UNC/device forms and malformed input. For an absolute external key, accept only
the three exact original canonical identities from the fixed approved history:
bundled pwsh.exe, its System.Management.Automation.dll, and Windows PowerShell
powershell.exe. Preserve their expected SHA256 values; do not search PATH, replace
binaries, accept another executable or infer a new machine mapping.

Verify each actual file hash through the existing Driver-CheckHash/Driver-Hash
behavior. Require valid64hex expected hashes; missing files, wrong hashes,
unexpected external identities and conflicting canonical aliases reject.
Identical aliases may only share the same hash. No writes or process launch.

Gate wiring must call this production verifier on the authenticated original
HistoricalFiles map. All other gate checks, old stop rejection, approval/input
hashes, candidate bindings, process ownership, output/evidence predicates and
timeouts remain unchanged. Other30operational files stay byte-identical.
Do not alter frozen manifests, old tests/receipts or runtime installation.

## Proposed regression and budget

Add TWO local auxiliary computations, RED20/GREEN21 (new ceiling21), no host,
Prepare, snapshots, child targets, compilation or reviewer. Cumulative counters
would become21aux/4hosts/3compiles/186snapshots only if both run.

Use independent fixed names history-path-red and history-path-green for
intent/final/tool records, history-path.input.json and candidate input, and
history-path.stop.json. No extent-contract output slot reuse.

RED must exercise the actual current history-loop AST and existing hash helpers
on the saved absolute runtime key; expected failure is the saved invalid-path
behavior, not merely absence of a new function. Do not invoke driver top level
or bypass its stop. Freeze test/input identities before RED.

GREEN invokes the extracted production verifier on the actual full2307entry
history, and checks its path/hash branches with bounded in-memory mappings of
existing pinned files. No mocks of file hashing, no substitute test-only join,
no temporary file edits, no actual host start. Coverage includes all three
external identities, a repository-relative file, wrong hash, missing file,
unauthorized absolute location, drive-relative/rooted-relative path, UNC/device
form, traversal escape, conflicting and identical canonical aliases, malformed
hashes, unchanged source window and actual gate-to-verifier argument binding.

The implementation plan must fix group order, exact function signatures and
extraction allowlist, expected RED error fields, mutation fixtures and recorded
hashes before execution. Existing successful22+10groups are historical evidence,
not replayed or claimed as fresh coverage. Source-only checks do not replace
behavioral verification of real path resolution.

## Stops, preservation and later recovery

CreateNew intent before each computation; preserve full actual tool input/output.
First unexpected result/drift/recording failure stops without patch/retry. Even
GREEN success ends local scope: extent-contract.stop remains, old candidate source
hashes no longer match the edited driver, and no Prepare/host is authorized.

Persist portable preparation/candidate/final archives, result/status/recovery
documents and explicit source/test files; commit/push to the existing development
branch and verify remote equality. No secrets, tool binaries or raw ignored files
are staged. No main merge or product acceptance.

Re-freezing and the single unused host require a separate approved recovery route
with new identities anchored to this outcome and all old stops; do not silently
reactivate the unconsumed but frozen1host/31snapshots. Current actual counters
remain19aux/4hosts/3compiles/186snapshots. This document is planning only.
