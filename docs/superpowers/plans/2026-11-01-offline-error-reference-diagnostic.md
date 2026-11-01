# Error Reference Propagation Diagnostic Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans natively after explicit approval. No agents or reviewers. Steps use checkbox syntax for tracking.

**Goal:** Measure where real ErrorRecord/Exception reference identity changes across collection and rethrow, without repairing or replaying the stopped integration.

**Architecture:** One small standalone diagnostic source, six fixed cases inside one host, actual missing-file errors, the exact frozen collector definition, and live pairwise reference comparisons before projection. The parent retains original streams and records one attempt; no new launch controller, runtime version or integration namespace.

**Tech Stack:** Existing Windows PowerShell7.6.5/Core/x64, .NET File/Process, SHA256 and AST parsing, no Add-Type or installs.

**Spec:** [Readonly causal analysis](../../roadmap/2026-11-01-offline-error-reference-readonly-analysis.md); original first-cause/reference/no-retry requirements in the [offline design](../specs/2026-10-19-offline-diagnostic-entry-migration-design.md). The diagnostic scope and six-case design below are proposed for approval, not a claim that the rethrow mechanism is already proven.

Status: DRAFT_PENDING_EXECUTION_APPROVAL. Baseline8a07dba54c802b17baad1ba8e377eb1482add8b5. User approved drafting only. No source, import, process, compilation, test, diagnostic or review in this turn.

## Global Constraints

- Request NEW1diagnostic host /0controlled targets /0compilations /0reviews. This is not an integration capture host, refund or release of frozen v3 remainder6/7/1/conditional1 or older limits.
- Preserve all existing source/tests/oracles/manifests/stops/attempts/raw/portable archives unchanged. No I1/I2 repair, integration RunOne/Prepare, original capture invocation, production module import or new runtime tree/controller.
- Use real File.ReadAllText errors, not constructed/thrown synthetic exceptions or deserialized errors. Do not invoke the C# fixture, production writer or profile entry.
- Compare actual live references before serialization. Equal text is not identity; object hash codes alone cannot prove equality. Never recover live references from JSON.
- One fixed attempt, no retries/new suffix/fix pass after any actual static, launch, record or validation failure. A lost/unknown result consumes the attempt and stops.
- Native/PowerShell finite scope replaces skill-default Bash/setup/baseline/TDD/full-suite/task-done reruns, review and deletion. No dynamic control-negative host beyond the one probe; diagnostic result is a measurement, not production acceptance.
- No PID/name lookup, tree kill, Job/watchdog, indefinite wait, readiness delay, installs/security changes, main merge or release. All5capability flags remain false; I1/I2 and C1/C2 not closed.

## Review Focus

1. Observer code must not replace/mutate an error before comparison: retain original objects and record pairwise ErrorRecord AND Exception equality, task1.
2. Collection before/after rethrow must be distinguishable: fixed event order and before/after primary/secondary counts, task1.
3. Same text from separate real faults must not merge: sixth independent-fault control, task1.
4. Unknown reference outcomes are data, not failed predictions or retrofitted PASS: validate structure/signatures, record either true/false result, task2.
5. A collector/recorder/transport failure must retain the real stop: original raw streams, one consumed intent, exact process/exit/drain and no fallback evidence, task2.

## Files, input freeze and attempt

Create only two new source files under `docs/diagnostics/isolated-fuzz/p2b-error-reference-diagnostic/`: `.gitattributes` (`* text eol=lf`) and `reference-probe.ps1`. No parent-controller script or general command/path interface. Probe has no caller-selected case/path/callback. Output/packet paths fixed internally.

Packet P: `.superpowers/sdd/2026-11-01-offline-error-reference-diagnostic`.
Create explicit approval.json, progress.md, static.tool-result.json, binding.json,
source/plan/approval snapshots, attempt.intent.json, stdout.bin, stderr.bin,
observations.jsonl, final.json and complete tool-result.json as actually produced.
Missing/partial outputs remain missing/partial. Export actual records at handoff
as new leaf/portable-final.json; no replacement of earlier archives.

Before invocation, binding.json pins exactly8inputs: two new source files,
existing integration capture.ps1, readonly analysis report, this approved plan,
fixed engine, SMA DLL, and actual approval.json. Snapshot the two sources and
approved plan/approval once separately. Record binding SHA in the consumed
intent and parent transcript; do not derive execution consent from a disk file.
Create P only after explicit execution approval and absent packet/outputs.

Existing capture SHA256 F8BC1C86C20A5C015C45F91312474CBC9048F6E481C6F0A4B281B969E429F8F2.
Extract exactly its Capture-Failure FunctionDefinitionAst text, check unique
definition and syntax, and dot-source only that pure definition into the new
probe's own scope. Never execute/import capture.ps1 itself. Reset its explicit
HostPrimary/typed HostSecondary/HostOverflow once per case; keep the complete
case state and observation references live until all comparisons finish.

## Task1 — six fixed observations and static preparation

**Files:** new two source files and P approval/static/binding/snapshots.
**Interfaces:** internal Invoke-ODReferenceFault(MissingPath) calls actual File.ReadAllText at one fixed source location; Observe-ODReference(Label,Record) retains the original ErrorRecord and Exception references and collector counts without collecting/mutating the error; no exported API.

- [ ] Record explicit approval of this whole plan and NEW1/0/0/0. Verify clean existing branch/origin, absent diagnostic packet/source/portable outputs, pinned capture and original failed archive. No historical replays or corrected rerun of the failed full-pin audit. Create only the explicit new diagnostic records.
- [ ] Define literal six-case order and structural/error/event contracts below BEFORE implementation. Do not predeclare which rethrow returns the same reference; that is the unknown being measured.
- [ ] Implement cases sequentially in one host, each with a fresh fixed P/cases/<Case>/ directory. Assert its absent.txt does not exist, then call ReadAllText on it. Creating the case directory ensures missing FILE, not a missing-directory error. No file is created at absent.txt. Exactly one fault operation in cases1–5, exactly two in case6.

| Case | Operation and observation points |
| --- | --- |
|DirectOuter|one real fault; outer catch observes and collects once|
|CatchBare|inner catch retains/observes/collects then bare throw; outer catch observes/collects|
|CatchRetained|identical inner catch including retained variable; use throw of that retained record instead of bare throw; outer observes/collects|
|FinallyBare|successful empty try; finally performs fault; inner catch inside finally observes/collects then bare throw; outer observes/collects|
|FinallyRetained|same finally control, but throw retained record; outer observes/collects|
|IndependentSameText|two separate real faults via same helper/path, each independently caught/observed/collected; no rethrow|

- [ ] In rethrow cases save exact inner/outer labels, original error projections and collector before/after counts. Retain actual objects separately from JSON. Pairwise compare every observed ErrorRecord and their Exception with ReferenceEquals; include full square matrices in final case data. Keep diagonals true/symmetry asserted, but off-diagonal true or false is measured data. Do not normalize wrappers, inner exceptions or fields to force equality. Case6 must retain both independent errors; compare their projections without requiring Stack text to be equal.
- [ ] Require operation errors Type=System.Management.Automation.MethodInvocationException, ErrorId=FileNotFoundException, message exactly `Exception calling "ReadAllText" with "1" argument(s): "Could not find file '{fixed-case-absent-path}'."`; bind source position/stack to the actual helper/call chain. Unexpected signatures stop; do not call unrelated errors valid measurements. DirectOuter collector total1 and no secondary; IndependentSameText collector total2 with primary first/secondary second by reference. Rethrow cases may measure total1 or2; preserve exact counts and references, not a predicted verdict. No case has Overflow or more than2entries.
- [ ] Use a separate retained direct JSONL journal opened once before risky pure-helper loading. It is diagnostic recording, not a substitute production publisher. Each case emits one complete row after live comparisons. Fixed root keys: Schema,Case,FaultCalls,Events,RecordSame,ExceptionSame,Collector,ErrorProjections. Events retain Label/BeforeCount/AfterCount and observation index; Collector contains PrimaryIndex,SecondaryIndices,Overflow. After six valid rows, write final row with exact keys Schema=1,Status="COMPLETE",Cases=6,FaultCalls=7 and the5capability flags=false, then flush/close the held stream once. That row alone is provisional: only parent-confirmed exit0 and healthy final flush/close make the packet complete. Never reopen after closing or log live ErrorRecord/Exception objects themselves.
- [ ] On first journal write/flush/close failure latch and stop, no reopen/retry/fallback JSON. Unexpected error to original stderr once if possible; nonzero exit, leave partial JSONL. No diagnostic success inferred from stdout. Bounds:6case rows+1final row, total1MiB, ≤2error observations/entries per case. Parent/static-only checks reject unbounded loops, dynamic paths/cases, imports/Add-Type or extra process creation. AST/LF/JSON and frozen input/source snapshots must pass before invocation; no correction pass after actual failure.

## Task2 — one host, exact transport, then stop

**Files:** actual P attempt/raw/observations/final/tool receipts only; no source edits.
**Interfaces:** consumes actual binding identity and explicit approval; produces measurement/partial/unknown, never integration authority.

- [ ] Parent executor creates attempt.intent.json via CreateNew, writes actual binding SHA, approved1host/0targets/0compiles/0reviews and consumed=true, flushes/closes before launch. Any existing intent or uncertain launch state forbids another attempt. Probe verifies fixed binding inputs and intent before its cases.
- [ ] In a single finite PowerShell orchestration, retain exactly one .NET Process; launch only the pinned engine with -NoLogo -NoProfile -NonInteractive -File fixed reference-probe.ps1. Use one original stdout/stderr stream consumer each. Overall host wait≤120000ms, exact-owned-object exit wait2000ms and shared stream drain2000ms. On startup throw/timeout, determine association only from that same object; terminate only it if associated/still running, drain/release once. No process search/tree kill or extra readiness wait. Retain actual lifecycle tuple and original exception in final.json.
- [ ] Preserve initial tool object and every actual wait chunk; wait only returned session ID until terminal completion. Save the complete transcript only afterward. Tool yield is not command completion. Lost/unknown/nonzero execution or recording failure consumes the single host attempt and stops, without edits/retry.
- [ ] On exit0 require owned exit/drain true, no timeout/termination/rescue, empty stderr, exact6ordered case rows+1COMPLETE final and all input hashes unchanged. Validate row keys/types, fault counts, actual fixed error signatures, reference-matrix diagonal/symmetry, collector identity/count invariants and projections. Do NOT require a preselected same/different outcome for the rethrow cases. Report measured identities; inconclusive behavior is a valid diagnostic conclusion if records are complete, not proof of I1 repair.
- [ ] Stop after this one host even if complete. New diagnostic1host allowance exhausted, integration6/7/1/conditional1 remains FROZEN. No fixing writer/collector/test assertions, integration retries or independent review is included.

## Task3 — exact recovery and limited conclusion

**Files:** new docs/roadmap/2026-11-01-offline-error-reference-diagnostic.md, new leaf/portable-final.json; update current-status/repository-recovery.
**Interfaces:** actual records→source-bound diagnostic explanation, not reconstructed original aborted-case references.

- [ ] Distinguish measured reference propagation in this simplified probe from the missing internal state of historical IntentDisposeOnly and from compound Write/Dispose production behavior. Never fill in historical first.Count or claim I1/I2 fixed. Recommend a later source/contract adjustment only if the actual mechanism supports it; no follow-on code or budget release.
- [ ] On success or failure, archive every actual file/directory/source snapshot/raw/partial/zero-byte/error/tool object as base64 with lengths/hashes; decode and byte-compare originals. Preserve all earlier source/records/archives unchanged. No fabricated terminal rows. Explicit necessary source/doc/archive staging, byte-identity checks, commit/push existing branch and verify remote SHA. No main merge/deletion/raw-directory/cache/credential commit.

## Self-review and execution gate

Six cases contrast outer-only, matching catch/finally rethrow pairs and separate
same-text faults; live matrices prevent JSON-equality substitution. Bounded
collector/events, original signatures, separate journal and exact process/stream
checks cover five listed risks. This is a simplified7real-I/O-error diagnostic
inside1host,0targets/0compiles/0reviews, not the ten original writer vectors.
No runtime controls or errors from the old capture are executed or patched.

Please explicitly approve Native implementation and the NEW1diagnostic host /
0targets /0compilations /0reviews, then stop after the single result. Existing
integration remainder stays frozen regardless of outcome. Planning approval
alone does not authorize creating/running the probe or changing frozen source.
