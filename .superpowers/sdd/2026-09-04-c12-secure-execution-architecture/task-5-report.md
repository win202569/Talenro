# Task 5: Prepared Artifact Direct-Leaf Ledger

Status: DONE

## Implementation

- Added `New-C12DirectLeafLedger`, `Register-C12DirectLeafIntent`, `Bind-C12DirectLeaf`, and `Converge-C12DirectLeafLedger`.
- Every ledger entry carries `Name`, `Kind`, `Expected`, `CreateAttempted`, `Identity`, `NumberOfLinks`, `RefCount`, `Lifecycle`, and `LastCleanupError`, plus retained ownership/cleanup handles used internally.
- Enforced the monotonic lifecycle `NeverAttempted -> CreateAttempted -> Bound -> CleanIntent -> Removed -> Absent`.
- Created `go-cache` and `go-tmp` atomically as individually owned non-reparse subtrees and retained their exact directory handles.
- Registered all validator source, temporary executable, final executable, initializer executable, and test2json direct leaves before creation; bound their volume serial, file index, reparse state, and link count immediately after successful creation/publication.
- Replaced prepared-artifact `Remove-C12BoundedDirectory` with a closed direct-child inventory, validate-all-before-delete behavior, exact handle-based per-leaf cleanup, and empty-root deletion only after every ledger entry is `Absent`.
- Added Windows `FileDispositionInfoEx` exact-handle deletion with POSIX semantics and a legacy fallback; link count and identity are reverified on the retained handle immediately before deletion.
- Preserved `CleanIntent`, exact identity, cleanup handle, and the last cleanup error across an expired deadline so a fresh-deadline finalizer retry uses the same state.
- Updated the native-failure regression to invoke the production `Invoke-C12Group(base)` path directly, avoiding the unrelated Task 4 focused-validator graph preparation while retaining the exact native-failure-before-cleanup assertions.

## TDD evidence

Initial existing RED:

```text
prepared direct-leaf ledger has no single reachable executable call path containing every term [...]
prepared direct-leaf cleanup contains forbidden executable AST term "Remove-C12BoundedDirectory"
production prepared artifact root lacks a direct-leaf ledger
FAIL
```

After adding the required mutation cases, the test remained RED for the same missing production ledger. Intermediate behavioral REDs exposed and corrected Windows delete-pending semantics and retained retry-state handling; assertions were not removed.

Final required GREEN, with `GOOS=windows`, `GOARCH=amd64`, `CGO_ENABLED=0`, and the authenticated overlay in `GOFLAGS`:

```text
--- PASS: TestC12PreparedArtifactDirectLeafLedgerIsClosed (47.69s)
--- PASS: TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup (46.41s)
--- PASS: TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline (19.50s)
PASS
ok talenro.local/platform/internal/testinfra 119.715s
```

The direct-ledger test covers an unknown direct sibling, hard-link splice, reparse leaf, replaced directory identity, retained foreign objects on rejection, and positive cleanup convergence.

Real prepared-artifact success-path regression:

```text
--- PASS: TestC12PreparedArtifactExecutionAndTamperConverge (49.20s)
    --- PASS: TestC12PreparedArtifactExecutionAndTamperConverge/success (49.20s)
PASS
ok talenro.local/platform/internal/testinfra 57.427s
```

PowerShell parsing and scoped `git diff --check` passed; the only output was the repository's normal LF-to-CRLF advisory.

## Scope and risks

- Only `scripts/run-c12-integration.ps1` and `internal/testinfra/c12_integration_manifest_test.go` are authorized for the commit.
- No name-, wildcard-, or recursive prepared-root cleanup is used. Recursive deletion remains confined inside a previously identity-bound `owned_ephemeral_subtree` handle.
- No external Docker volume or old one-shot/oracle artifact was inspected, mounted, relabelled, mutated, or deleted.
- A failed external compiler that materializes a malformed/zero-length temporary file before returning is deliberately retained rather than treated as owned without a successful identity bind. This is fail-closed leakage for later operator diagnosis, not name-based cleanup.

## Review correction 1 — retain exact handles through namespace convergence

The task review found that the initial file and directory deletion methods disposed their exact identity handles before proving namespace disappearance. On a legacy/delete-pending path or with another sharing handle, a timeout could therefore leave `CleanIntent` without the original retry authority.

Added `TestC12PreparedArtifactCleanupRetainsDeletePendingHandle`. Its controllable cleanup handle issues a real cleanup intent, deliberately keeps the namespace present past the absolute deadline, and asserts that lifecycle remains `CleanIntent` with the same identity, same handle object, and exact error. After simulating namespace convergence, a fresh-deadline retry must release that same handle and advance through `Removed` to `Absent`.

RED:

```text
delete-pending failure did not retain CleanIntent, identity, original handle, and exact error
--- FAIL: TestC12PreparedArtifactCleanupRetainsDeletePendingHandle
```

The correction separates `RequestDeleteExact*` from `ReleaseDeletedExact`. Files and directories are reverified through their retained handles before deleting the exact namespace; those handles were created/opened with delete-sharing so the namespace can disappear while the identity handle remains live. `Converge-C12DirectLeafLedger` releases a file, subtree, or artifact-root handle only after the corresponding path is confirmed absent. A deadline or pending namespace preserves `CleanIntent`, the original handle, identity, and `LastCleanupError` for retry.

Focused GREEN results, all with the authenticated Windows/overlay environment:

```text
TestC12PreparedArtifactDirectLeafLedgerIsClosed                  PASS 72.29s
TestC12PreparedArtifactCleanupRetainsDeletePendingHandle        PASS 32.79s
ok talenro.local/platform/internal/testinfra 114.788s

TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline     PASS 21.83s
ok talenro.local/platform/internal/testinfra 28.459s

TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup            PASS 59.49s
ok talenro.local/platform/internal/testinfra 68.812s

TestC12PreparedArtifactExecutionAndTamperConverge/success       PASS 60.59s
ok talenro.local/platform/internal/testinfra 67.815s
```

PowerShell parsing remained clean. No production deadline or rejection assertion was relaxed, and no name/wildcard cleanup was introduced.

## Review correction 2 — handle-bound disposition and fail-closed namespace classification

The second review identified a TOCTOU in correction 1: after retained-handle verification, `File.Delete` / `Directory.Delete` still acted on a path name, and post-release `File.Exists` / `Directory.Exists` could mistake a type replacement for absence. The delete-pending regression also used only a fake cleanup handle.

RED evidence from the new real production C# harness:

```text
The term 'Request-C12DirectLeafDelete' is not recognized
--- FAIL: TestC12PreparedArtifactCleanupRejectsRealHandleNamespaceReplacement

real same-identity stage link-after-request: Access is denied.
--- FAIL: TestC12PreparedArtifactCleanupRejectsRealHandleNamespaceReplacement (22.57s)
```

The implementation now separates three states without any post-verification name deletion:

- request deletion by `SetFileInformationByHandle` on the retained exact file/directory handle, after immediately rechecking type, reparse state, identity, and (for files) `NumberOfLinks == 1`;
- release only that delete-pending handle;
- classify the exact namespace by atomically opening it with `OPEN_REPARSE_POINT | BACKUP_SEMANTICS`: missing converges, an open/classification failure preserves `CleanIntent`, identity, and the exact error, a foreign/type/reparse replacement is rejected without deletion, and a still-visible matching identity is safely reopened for another exact-handle request.

The production C# regression uses real `C12SealedExecutable` and `C12OwnedDirectory` request/release paths for file-to-directory and directory-to-file rename/replacement races. It also forces the supported legacy disposition compatibility branch and holds a second real shared handle, producing an actual Windows delete-pending namespace: classification is denied and retained as error state, then converges to absent after the blocker closes. The AppContext switch affects only whether the existing exact-handle legacy compatibility branch is selected; the default continues to prefer `FileDispositionInfoEx` and fall back when unsupported.

GREEN evidence with the authenticated Windows/overlay environment:

```text
TestC12PreparedArtifactCleanupRejectsRealHandleNamespaceReplacement PASS 22.06s
ok talenro.local/platform/internal/testinfra 33.108s

TestC12PreparedArtifactDirectLeafLedgerIsClosed                   PASS 75.99s
TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline      PASS 42.74s
TestC12PreparedArtifactCleanupRetainsDeletePendingHandle         PASS 36.32s
TestC12PreparedArtifactCleanupRejectsRealHandleNamespaceReplacement PASS 34.00s
```

The broad tamper aggregate additionally passed success, initializer-capabilities, byte-tamper, DACL, wrong-role, and replay. Its file-id-splice, reparse, and hard-link cases retained deliberately unregistered siblings (`*.original` / `receipt-hard-link.exe`) and therefore left the artifact root, which is the new fail-closed direct-leaf-ledger behavior rather than a safe target for name cleanup. The native-failure fixture was blocked before its intended assertion by `resolve postgres immutable image timed out after 8 seconds`; its earlier focused correction-1 run remains green.

### Review correction 2 follow-up — fail-closed tamper expectations

The three legacy tamper cases now treat the first cleanup rejection as the expected result and assert all of the security state, rather than expecting the root to disappear:

- the exact unknown sibling remains present;
- `RootLastCleanupError` preserves the exact unknown-sibling rejection;
- artifact-root identity ownership and `RootLifecycle=Bound` remain intact;
- the executable ledger entry keeps its original identity and `Lifecycle=Bound`, without a fabricated per-entry error;
- production cleanup never empties the directory after encountering the unknown sibling.

Only after these assertions, the test fixture cleans objects that it explicitly created and identity-bound. Relocated originals are deleted through fixture-owned `C12SealedExecutable` handles. The replacement file is independently handle-bound. The junction is re-inspected as the same reparse-directory identity before exact non-recursive removal. The hard-link path is re-inspected as the held identity with link count two before removal, and the surviving bound path is then re-inspected as the same identity with link count one. Production cleanup is invoked again only after the fixture-owned foreign objects are gone.

The initial fixture-cleanup RED was:

```text
file-id-splice: You cannot call a method on a null-valued expression
reparse:        You cannot call a method on a null-valued expression
hard-link:      PASS 50.78s
```

The cause was a test-only assumption that identity binding populated `CleanupHandle`; the corrected fixture opens and retains its own exact handle immediately after relocating the original.

Final isolated GREEN:

```text
TestC12PreparedArtifactExecutionAndTamperConverge/file-id-splice PASS 48.68s
TestC12PreparedArtifactExecutionAndTamperConverge/reparse        PASS 28.35s
TestC12PreparedArtifactExecutionAndTamperConverge/hard-link      PASS 32.76s
ok talenro.local/platform/internal/testinfra 139.988s

TestC12BaseRunnerCapturesNativeFailuresBeforeCleanup             PASS 71.27s
ok talenro.local/platform/internal/testinfra 79.038s
```

The native fixture passed when rerun without concurrent harness load; no production timeout was changed.

## Review correction 3 — handle-relative owned-subtree traversal

The retained root handle previously protected only the initial and final root check. Recursive enumeration and descendant opens still used `RootPath` strings, leaving rename-and-replacement windows at the root and at every nested directory.

The Windows implementation now grants retained directory handles `FILE_LIST_DIRECTORY` and enumerates them with `NtQueryDirectoryFile(FileIdBothDirectoryInformation)`. It opens each enumerated child with `NtCreateFile` relative to `OBJECT_ATTRIBUTES.RootDirectory=parentHandle`, always using `FILE_OPEN_REPARSE_POINT`. Before recursion or disposition, the opened handle's file ID, type, and reparse state must equal the enumeration snapshot. Recursion passes only the bound child handle; deletion remains `SetFileInformationByHandle`.

Enumeration uses a 64 KiB native buffer and treats only `STATUS_NO_MORE_FILES` as completion. Each deletion restarts the scan, while `.` and `..` advance with `RestartScan=false`; this avoids repeatedly returning the first pseudo-entry. All other failing NTSTATUS values are converted with `RtlNtStatusToDosError` and rejected.

The controlled regression performs both root rename+replacement and nested rename+replacement after enumeration but before relative open. Separate markers prove both windows fired. In each case the replacement remains present through the production operation and the external sentinel survives. The original owned object is cleaned only through its retained handle-relative traversal; the nested replacement is rejected on the enumeration/open identity mismatch.

RED included a real non-converging enumeration caused by restarting on every `.` query, followed by the precise fix above. Final GREEN:

```text
TestC12OwnedSubtreeTraversalRejectsRelativeReplacement PASS 3.18s
ok talenro.local/platform/internal/testinfra 4.626s

TestC12PreparedArtifactDirectLeafLedgerIsClosed                         PASS 3.80s
TestC12PreparedArtifactCleanupRetainsOwnershipAfterDeadline            PASS 2.02s
TestC12PreparedArtifactCleanupRetainsDeletePendingHandle               PASS 6.36s
TestC12PreparedArtifactCleanupRejectsRealHandleNamespaceReplacement    PASS 2.30s
TestC12OwnedSubtreeTraversalRejectsRelativeReplacement                 PASS 1.29s
ok talenro.local/platform/internal/testinfra 17.410s
```
