# SDD ledger — plan: docs/superpowers/plans/2026-11-09-windows-c4-observer-entry.md

User approved inline execution after5a13d26b. Baseline5a13d26b, Task1 in progress.
Pre-flight: one integrated task; no shared interfaces between tasks. Existing linked
worktree verified, expected development branch, no superproject, clean baseline.
Fresh baseline: observe-c5=32, c5-observation-session=120, guard=48, managed=32,
native=78; all exit0 in separate PowerShell hosts (310 assertions).

Ruling: bundled workspace/task helpers require Bash; use this portable ledger and
read the task directly in PowerShell instead. Cost: manual task bookkeeping.
Ruling: verification stays within the approved synthetic-only boundary, overriding
generic full-project build/setup guidance. Cost: real Windows/product integration
remains unverified. No installs, .NET build/test or live observation authorized.

No scratch workspace created. No native launch allowance.

2026-11-10 implementation: default/policy/preflight RED missing-entry assertion,
then GREEN34. Receipt factory RED missing behavior, then GREEN105. Adapter/integration
RED missing backend, then GREEN152. All runs separate PowerShell7 hosts, exit1 on
each intended RED and exit0 on GREEN; no actual native/CIM/launch/sleep.
Fixture corrections: PowerShell array multiplication needed parentheses; shared
failure uses Kind=outcome, Outcome=observation-failed (not Kind=failure). Corrected
expectations to the unchanged shared contract. No production changes for these.

Manifest serialization and cleanup precede process acquisition; tests exercise
collision, invalid IDs/context, path/reparse checks, acquisition/write/flush/disposal
failures, exact manifest values and correlated shared synthetic evidence. Boundary
wrappers isolate filesystem writes; OS behavior and adversarial races not proven.
Final independent review completed below. Checkpoint is not live/UI acceptance.

Task1 implementation complete (base5a13d26b; final verification command pattern
`pwsh -NoProfile -File clients/windows/tests/<name>.tests.ps1`): observe-c4=152,
observe-c5=32, c5-observation-session=120, observation-session-guard=48,
process-observation=32, native-process-observation=78; total462, all exit0.
Both new scripts parse with zero errors; git diff --check passes; C5/shared runtime
and tests have empty diff against5a13d26b. Eight explicit necessary source/test/doc
files only in checkpoint. Independent review range is5a13d26b..checkpoint HEAD;
review this ledger's Ruling lines, full plan/spec, guide and status/recovery changes.

## Final independent review — 2026-11-10

Fresh reviewer c4_entry_final_review inspected5a13d26b..57e19be6, all changed files
and relevant unchanged modules. Fresh C4 rerun152 exit0; range diffcheck passes.
Verdict0 Critical/0 Important/1 Minor; scoped implementation checkpoint acceptable.
Executor regraded by user effect: coverage gap remains Minor, no demonstrated defect.
No fix pass required; no second review. No scratch directory to remove.

Final: minor (deferred): direct regression cases for serialization failure, underlying
stream disposal failure, and real New-C4TextWriter AutoFlush initialization failure
are absent. Inspected paths fail closed, but future edits have less regression protection.

Final: Ruling: actual Windows process/native/CIM interoperability stays UNVERIFIED —
live execution prohibited; cost: OS integration defects may surface later.
Final: Ruling: real token classification/CIM inventory availability stays UNVERIFIED —
only doubles authorized; cost: actual environment may reject or fail its preflight.
Final: Ruling: actual filesystem ACL/disk-full/device/reparse behavior stays UNVERIFIED —
memory doubles only; cost: real storage behavior may differ.
Final: Ruling: hostile filesystem replacement races remain outside trusted-workspace
scope — no security boundary claimed; cost: an attacker could defeat path checks.
Final: Ruling: no hard real-time interruption — synchronous calls cannot be forcibly
bounded here; cost: delayed OS/write/cleanup may overrun, never timely acceptance.
Final: Ruling: UI readiness and authorization remain operator prerequisites — switch
is not permission; cost: misuse could launch without a usable observation route.
Final: Ruling: English labels/offline notice/mirroring/clipping stay UNVERIFIED — no UI
run; cost: visible defects remain undetected by this checkpoint.
Final: Ruling: actual-email/technical-ID LTR readability stays UNVERIFIED — no real
control evidence; cost: readability defects may remain.
Final: Ruling: exact bidi codepoints stay UNVERIFIED — screenshots/stripping extraction
are insufficient; cost: hidden-control errors cannot be ruled out.
Final: Ruling: logout/re-entry profile retention/visible reset stays UNVERIFIED — no
UI execution; cost: profile/reset defects may remain.
Final: Ruling: exhaustive reset remains outside the selected route — visible state is
not full internal-state proof; cost: hidden reset defects need separate tests.
Final: Ruling: shaping/screen-reader/contrast/viewport/full-keyboard/fallback-language
acceptance remains outside scope — no evidence gathered; cost: usability gaps remain.
Final: Ruling: C6/overall C5/Task11/12 and wider gates/historical shutdown cause remain
unchanged — tooling is not product evidence; cost: existing blockers remain unresolved.
Final: Ruling: unchanged C5/shared modules receive integration review, not independent
requalification/refactoring — explicit no-edit scope; cost: preexisting defects may remain.
Final: Ruling: fresh-machine restoration of ignored artifacts/receipts/credentials is
not proven — only portable source/docs preserved; cost: a clone may lack live prerequisites.
Final: Ruling: remote equality is executor verification, not reviewer evidence; main
merge readiness not granted — follow AGENTS ordinary dev push only; cost: recovery
depends on successful remote verification and main remains behind this checkpoint.

Preservation: keep linked worktree and existing branch; no integration menu needed
because AGENTS and approved plan already select ordinary push, no PR/main merge.
Stage only reviewed necessary files. Verify final six suites and clean diff, commit
review/status docs, push verified origin, then compare remote/local SHA. Any failure
must be reported as local-only or incomplete rather than claiming remote recovery.
