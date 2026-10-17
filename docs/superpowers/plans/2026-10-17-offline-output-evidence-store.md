# Offline output evidence store implementation plan

> **For agentic workers:** Use superpowers:executing-plans inline; no implementer delegation.

**Goal:** Save and reread bounded offline evidence without treating incomplete files as completed evidence.
**Architecture:** Primitive binary payload, closed and reread before a last-written length/SHA256 completion marker. Reconstruct the existing package; trust no stored classification or approval.
**Tech stack:** Windows PowerShell 5.1 / PowerShell 7, .NET streams, existing isolation file metadata helpers.
**Spec:** Approved persistence design at `docs/roadmap/current-status.md`, commit 6ce0bbd6, confirmed by the user on 2026-10-17.

## Global constraints

- Fixed worktree-owned `.superpowers/diagnostics/offline-output-evidence-store` scope; caller prepares root, writer creates only a fresh direct child.
- CreateNew files; no overwrite, deletion, automatic retry or real process collection.
- Reject UNC, ADS, traversal, reparse ancestry and multiply-linked evidence files. Hold read bindings through verification.
- Payload magic `TLROEV01`, little-endian Int32 exit/encoding(1=utf8,2=utf16le), Int32 origin byte length/strict UTF8 bytes, Int32 stdout length/bytes, Int32 stderr length/bytes; exact EOF.
- Origin 1..1024 characters/at most4096 bytes; each stream at most1048576 bytes; payload at most3145728 bytes.
- Marker magic `TLROEC01`, Int64 payload length, SHA256 raw32 bytes; exactly48 bytes.
- Stop record on best-effort failure; preserve incomplete files. No authenticity, hostile-admin/concurrent namespace resistance, fsync/power-loss guarantee or acceptance/runtime authority.
- Exactly four allowed fresh test hosts: PS7 RED, PS5 RED, PS7 GREEN, PS5 GREEN; sequential,120s each,no retries. Old suite budgets frozen; no full-project suite or live diagnostic replay authorized.

## Review focus

Namespace races and hardlink timing (no hostile namespace guarantee); write failures between payload and marker; strict byte bounds before allocation; malformed length/encoding/UTF8/EOF with recomputed hash; PS5/PS7 binary compatibility. Tests cover malformed fixtures, incomplete states, hardlink/reparse rejection and boundaries; mid-write OS failure injection remains a review limitation.

## Task 1: Store and reader

Files: create `docs/diagnostics/isolated-fuzz/output-evidence-store.ps1` and `.tests.ps1`; preserve package/classifier/isolation unchanged.

Interfaces: `Write-OfflineOutputEvidenceStore -Directory <absolute direct child> -Package <existing package>` throws on failure, otherwise returns reader result. `Read-OfflineOutputEvidenceStore -Directory <path>` returns `{Verified,Reason,Package,OriginVerified=false,Acceptance=false,RuntimeApproved=false}`. No caller-selected root.

- [x] Write tests first: literal abc/raw binary roundtrip, empty/UTF16/nonzero exit/progress reconstruction, missing/partial/stopped/hash mismatch, valid-hash malformed binary, limits, existing destination preservation, path/link refusal.
- [x] RED: approved runner with new Store allowance, host7 then5; expected missing store API assertion, not syntax/dependency failure.
- [x] Implement strict bounded parser and marker protocol; use isolation metadata only, no runtime helper.
- [x] GREEN: achieved only in separately approved elevated continuation below;earlier failed GREEN records preserved,not silently replayed.
- [x] One scoped fresh source review;findings/limits preserved. Necessary files/checkpoints already committed/pushed;final verification documentation follows standing preservation rule.

## Execution ledger

Base:6ce0bbd620eee9332a7fb7173a685fef56be6571. Existing linked worktree verified; no shared task interfaces beyond unchanged package and isolation helpers.
Ruling: reuse the approved short design as binding spec; this plan transcribes that already-approved execution, not a new design gate. Cost if wrong: interface naming needs revision, no wider execution authority.
Ruling: use native PowerShell bookkeeping instead of Bash skill helper scripts; no extra test run via task-done, no dependency installation/full suite, no scratch deletion. User's four-host cap and preservation instructions take precedence. Cost: no fresh whole-project regression claim; raw evidence remains local/ignored.

Task1: INCOMPLETE. Tests written first;RED hosts1/2 matched missing API. Implementation written;GREEN3 failed before store entry on fixture negative exit string. Readonly AST confirmed StringConstantExpressionAst;no code correction or rerun. Fourth call unstarted/frozen. Source/test parse clean only;this is not behavioral verification. Evidence/next authority: `docs/roadmap/2026-10-17-offline-output-evidence-store.md`.

Final checkpoint review: one fresh readonly reviewer;0Critical,1Important(same unresolved fixture type failure),2Minor. No fix pass claimed;execution stopped before correction. Deferred minors: empty-origin mutation isolation and nonemptyUTF16 persistence coverage. Ruling on review exclusions: hostile namespace/concurrent mutation/ACL authenticity/mid-write faults/crash durability/real collection/runtime/wider correctness/acceptance remain unsupported or unverified,as detailed in checkpoint;cost if relied upon is false trust/unsafe runtime evidence. No stronger completion guarantee inferred from best-effort stopped marker. No second review or additional run.

## Approved fixture correction continuation

After f2c562a4,user approved the explicitly numeric negative fixture argument
and at most2 new sequential hosts,PS7 thenPS5,120s each,no retry. Original3
records preserved,original unused fourth frozen;no additional production fix
or minor-coverage changes in this continuation. Separate ignored StoreRepair
index/budget2. Apply only `-ExitCode ([int]-1)` in New-StorePackage,then run
suite once per host,stopping if the first fails. Prior failed GREEN plus AST
provides the fixture defect evidence;no new RED replay or full-suite authority.
Expected result:50 PASS/host,exit0,empty stderr;actual result must be recorded.

Continuation result:INCOMPLETE. Only the approved numeric fixture cast applied;
unchanged store source. New call1 PS7 exit1/7373ms after45 PASS at first
HardLink setup(access denied). No retries orPS5;remaining slot frozen. Old
index SHA unchanged;repair index SHA514F1512188C77376657E4FFB518ACA22B32B9C3BBD23E7C3D5A0B072890725F.
Prior fixture issue no longer occurs;hardlink/reparse/lock/PS5 unverified. No
second review or minor fixes. Exact audit and next elevation proposal in report.

## Approved elevated verification continuation

User approved at most2 new scoped elevated verification hosts after fa4f2094,
PS7 thenPS5,120s each,no retry/skip/security-setting change. Separate ignored
StoreElevated index/budget2;unchanged source/tests. Prior indexes and unused
slots remain frozen. Run sequentially and check actual child exit before the
second host. Preserve failures and fixture files;no new implementation scope.

Task1 scoped verification:COMPLETE under latest elevated authority only.
PS7 exit0/3433ms/50PASS/empty stderr;PS5 exit0/1853ms/50PASS/empty stderr.
No source/test changes since fa4f2094. Elevated2/2 consumed;index
SHAC9D7AE456A7DF8F8FD685B6C6EC7F5DE9C57F1B492CA17A223CCBF0F6AFE4701;
older index hashes unchanged. Numeric fixture Important fixed;2 minor coverage
gaps remain deferred,no second reviewer/fix loop. No wider acceptance or merge,
no scratch deletion. Prior rulings/limits remain;recorded history is not erased.
