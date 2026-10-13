# Main-state selected-issuer/serial binding — local execution record

Date: 2026-10-13. Workspace: `.worktrees/c12-b01-task9-coordinator`.
Branch: `codex/c12-b01-task9-coordinator`; unchanged base/HEAD:
`649cc0891011334dd032bcc4a9618110b0a9c5ce`.

Authority: approved [design](../superpowers/specs/2026-10-13-main-state-selected-issuer-serial-binding-design.md)
and [five-task Native plan](../superpowers/plans/2026-10-13-main-state-selected-issuer-serial-binding.md).
No staging, commit, push, merge, installation, cleanup or broad/physical tests
were authorized in this increment. Previous preservation approval does not
authorize saving this new increment.

## Local delivery scope

All five approved tasks are complete as a bounded local checkpoint. Same-reviewer
reinspection closed M1: new Critical0/Important0/Minor0, no outstanding finding.
Both complete25-command matrices pass on the identical frozen source. This is
not production readiness, merge authority or closure of historical blockers.

The new pure `internal/nodecontrol/nodeissuerbinding` package contains exactly
17 product files and 18 test files. Ten approved old consumer-test files add
only the exact new-package exception and its controls: 501 added lines, no
deleted lines. Old product/fixture/SQL/module files are unchanged.

The profile `main-selected-issuer-serial.ca-der-bound.v1` conditionally binds
Proposed identity references, not deployed issuer strings or storage mappings.
It owns the complete bounded input, evaluates real `nodeleafbinding.Evaluate`
once, protects immediate lower cleanup, retains full evidence and original
errors, and preserves the lower observation/expiry and inherited obligations.
It publishes only after the final context check.

The two 36-byte ASCII domains, including NUL, are:
`talenro.main-selected-issuer-ref.v1\x00` and
`talenro.main-selected-serial-ref.v1\x00`.
Issuer reference hashes domain + exact CA DER digest (68 bytes).
Serial reference hashes its domain + issuer digest + one-byte length + canonical
leaf serial magnitude (70–89 bytes). There is no hexadecimal/string/JSON
encoding, streaming hash, reparsing or caller-provided proof in the product.
Issuer mismatch is checked before serial mismatch; input is never repaired.

This local match does not remove any inherited responsibility, including the
older broader issuer/serial mapping obligation. Five explicit new responsibilities
remain: production store mapping/deployment provenance, key possession/handle,
history/lineage/row status, main-state/observation provenance, and full identity
composition/commit/runtime authorization. No runtime consumer is opened.

## Task trace and raw evidence

Execution scope:
`.superpowers/sdd/2026-10-13-main-state-selected-issuer-serial-binding/`.
Receipts preserve actual executable, offline process environment, arguments,
UTC times, exit code, stdout/stderr and Go/SQL/module hashes before and after.
Failed files remain failures. This ignored scope is local, not remotely saved.

1. Contract, budget and ten exact consumer branches: genuine missing-API RED001;
   missing-exception RED002–011, then corrected target-specific controls RED012–021.
   GREEN022–032. Nine unrelated-import expectations were removed because the
   existing helpers own only their exact target; no original rule was weakened.
   **M1 caveat:** 022 selects API/FixedErrors/Budget but excludes
   `TestIssuerBindingPreflightPriority` because its anchored alternative is
   `Preflight`. Priority coverage belongs to final full-package049, not022;
   the correct targeted selector also passed in075.
2. Full ownership/model lifecycle: RED033; GREEN034–035. Synthetic reflection
   traverses complete nested exported values, mutable backing and Old state.
   Tests cover two-way alias isolation, normalized times, two evidence slices,
   independent owners, shared shallow-model Clear, nil/zero and reachable clearing.
   Supplemental074 observes39 ownership context checkpoints and enumerates
   cancellation, deadline and ordinary panic without mutating the caller.
3. Fixed material, real lower and reference derivation: generator036 failed on
   an unused import;037 failed on a stale parent public key. Both corrected
   without old-product changes; these are generator failures, not product RED.
   Authoritative039 Go subprocess exit0 freezes39 packets. Independent fixture
   controls040 GREEN; missing lower/reference API041 RED;042 GREEN.
4. Exact binding/publication and complete lifecycle: missing Evaluate/projection
   RED043; selector044 GREEN; full new package045 GREEN. Formatting happened
   while045's wrapper was finishing, so045 is not a frozen final-matrix receipt.
   Final049 and074 verify the final source;074 dynamically observes710 context
   checkpoints in the complete composed evaluation. This is not710 new-layer-only
   checks and not a hard-real-time interruption guarantee.
5. Closed boundaries: missing guards046 RED; actual empty-schema047 RED reports
   all17 product digests. Source/contract/lifecycle audit precedes fixed-schema
   acceptance;048 GREEN. Actual source mutations must change bytes before being
   rejected. Import aliases, bound-local renaming and benign formatting are positive
   controls. Exactly35 inventory entries and17 normalized product schemas are
   locked. These are finite revision guards, not formal or exhaustive safety proofs.

## Frozen material and independent controls

The four prior original/frozen constants were AST-extracted without editing
their sources. This package retains the original input/manifest constants and
prior frozen controls, alongside its new frozen input/manifest constants.
All34 previous packets change only Previous/Proposed IssuerID/Serial declarations;
all original raw bytes and manifests remain unchanged.

Five added real signed controls cover one-byte high-bit magnitude, twenty-byte
low-bit magnitude, same-key/name CA reissue, same leaf serial under another CA,
and the lower-rejected twenty-byte high-bit magnitude requiring a21-byte DER
INTEGER. Ordinary tests decode fixed constants and do not issue certificates
or read old fixture source.

CA reissue uses identical leaf DER, identical CA subject/SPKI and different
valid CA DER. The independent reviewer recomputed1,008 raw hashes, all39
reference preimages/digests, six frozen-constant equality checks and all34 allowed
deltas with zero mismatches, and independently verified the five leaf signatures.

The existing parser limits INTEGER content to20 bytes before removing a legal
leading sign byte. Full-pipeline positives are L1/high and L20/low; L20/high is
a private canonical-encoding helper positive only. The real21-byte INTEGER is
a lower rejection, not a claimed new-layer success or a parser relaxation.

The generator is byte-identically archived as `freeze.go.txt`, SHA256
`6C41C4BADBFE502D982355F1746271064DB22194081B5ECE7C12D84A77D32EB5`.
It is not active Go source and does not evade consumer scans. Fixed test keys
are not production keys. ECDSA signing randomness means replay gives different
signed bytes; frozen039 output is authoritative, not a byte-deterministic replay.

Generator038/039 saved actual Go exit0 receipts, but oversized console-tail
delivery stalled and their PowerShell wrappers were interrupted afterward
(exit1). These wrapper exits are not relabeled successful shell runs. Full raw
files remain; the output helper now suppresses oversized console lines.
Successful standalone040 fixture controls validate the preserved output.

## Final verification

45 exact paths (35 new Go files plus ten old tests) passed formatting with no
remaining listing output. The format manifest matches the frozen sources.

First complete final matrix049–073:25 actual exit-zero commands —21 full
packages, specified authority tests, actual store compile-only, and two existing
nodeissuer fuzz targets. All before/after676-file source inventories match:
`F2E97B69041781E5E4AECD488C569BBBFE522E37F0AF99459A535E83088D7C35`
(SHA256 of the audit's canonical source-list JSON).

Envelope/Pair each ran8 executions, baselines5/4, zero new interesting inputs.
Actual Go durations6.832/8.253s; whole subprocess durations64.22/10.20s.
Configured5s fuzz time is not a whole-process wall-clock guarantee. These low
counts are limited evidence, not adequate exhaustive fuzz coverage.

Supplement075 explicitly selects and passes PreflightPriority. No product
regression was found or changed for M1; an invented failing Go test is not
claimed. Required full post-correction re-matrix076–100 passed all25 commands;
all before/after sources match the same676-file freeze. Envelope/Pair repeat
counts14/6, baselines5/4, zero new interesting inputs. Actual Go durations
8.077/8.110s and subprocess durations10.27/10.15s remain finite evidence.
The repeat formatting listing is empty; raw101 records45 paths, actual exit0,
zero-byte stdout/stderr and matching before/after hashes. Same-reviewer
reinspection has accepted the bounded local checkpoint with no outstanding finding.

## Independent review

Fresh reviewer `/root/selected_issuer_serial_final_review`, strongest available
`gpt-6-astra` at supported `xhigh`, reviewed read-only after source freeze.
Initial result: Critical0 / Important0 / Minor1. M1 is the early selector
coverage-attribution gap above, not a missing final behavior or product defect.

The reviewer independently checked all35 files, ten old diffs, spec/plan,
recorded commands/raw outputs,45 format hashes, the676-file freeze, branch,
HEAD and index, and fixture relationships. It did not rerun Go.
Same-reviewer documentation/scope reinspection closed M1, verified075 and all
076–100 raw results,101 formatting, frozen sources and unchanged integrity.
New findings: Critical0/Important0/Minor0. It confirmed all16 author rulings,
append-only current-status and checkbox-only plan edits (approved plan hash
reproduced by in-memory normalization). Reinspection was entirely read-only:
no file/Git edits, Go reruns, agents or chats. No scope drift was found.

## Author rulings on every declined domain

| Domain set aside by reviewer | Author ruling and cost |
| --- | --- |
| Production issuer/store mapping and migrations | Remain unverified; local digests are not production-string compatibility or migration evidence. |
| Deployment/CA trust/registry provenance | Supplied valid signatures do not establish deployed trust; operational acceptance remains open. |
| Private-key possession/key handles | No evidence produced; these responsibilities remain unresolved. |
| Previous certificate/history/lineage/row status | Proposed-only binding; no Previous proof or historical validity is claimed. |
| Full composition/commit/Ready/runtime | No such consumer or authorization added; future integration and acceptance are required. |
| Trusted clock/observation provenance | Conditional lower times retained; no independent clock trust proof. |
| Issuance uniqueness/hash collision freedom | Not established by SHA256 vectors; no uniqueness or mathematical collision-free claim. |
| Comprehensive baseline correctness | Relevant integration/scoped regressions only; old stack is not re-certified. |
| Whole repo/Linux/race/concurrency | Excluded; Windows scoped tests only and concurrent Model access unsupported. |
| Real DB/Docker/TLS/providers | Not run; compile-only store and installed Docker do not prove interoperability. |
| Hardware/reboot/PITR/physical recovery | No physical evidence; existing gates remain open. |
| Real-time/OOM/crash/hidden-copy erasure | Cooperative checkpoints and reachable cleanup only; no fatal-failure or physical-erasure promise. |
| Exhaustive fuzz/formal AST or memory safety | Finite tests/guards and low fuzz counts are limitations, not proof. |
| Remote recovery/CI/Git/deployment | New increment remains local/uncommitted; no remote recoverability or deployment readiness. |
| C12/B01/Task10/main/I2/I3/historical closure | No closure inferred; all historical blockers and unknown failure causes remain. |
| Final-document accuracy | Same-reviewer reinspection checked delivery evidence and rulings; these closure notes record its actual accepted verdict, not an author-substituted review. |

Execution adaptations: R01 PowerShell adaptation of Bash-only task helpers
costs manual identity/hash audits; R02 scoped/no-Git/no-install/no-cleanup plan
overrides generic broad setup/finish defaults, leaving physical/remote gates
open; R03 approved correction/re-matrix/same-reviewer reinspection is followed;
R04 approved sequential Native work has no intermediate independent reviewer,
so independent feedback arrives only at the final frozen increment.

## Integrity and handoff

Design SHA256 remains
`A4F48076299E2C13C0C37C03276860262AACB0C6420D7B534CA18345CD2E3BF9`.
Git index SHA256 remains
`CF1CB7A683FF65B05070635FDDB1ADC813081E1E3F2872697F84E8D4DC2ABBF5`.
The641-file old baseline differs only at the ten approved tests. HEAD remains
`649cc0891011334dd032bcc4a9618110b0a9c5ce`; no Git writes occurred.

This is a bounded local component checkpoint, not C12 completion or physical
acceptance. Commit/push requires separate authorization. Source, fixed test
material and documentation are local; ignored receipts/caches/archive are not
claimed GitHub-saved or recoverable on another computer.

## 2026-10-13 preservation handoff

After the explicit commit/push next-step offer, the user requested
"continue the next step". This authorizes saving the49 necessary files on
the existing `codex/c12-b01-task9-coordinator` branch to
`https://github.com/win202569/Talenro.git`, not a PR, main merge or cleanup.

All49 preflight files matched the completed handoff. Fresh pre-commit102–126
passed all25 scoped commands with the same676-file before/after source freeze
and clean45-file formatting. Envelope/Pair fuzz counts37,020/8, baselines5/4,
new interesting3/0; Go8.319/8.803s, subprocess10.9672579/14.913876s.
These remain finite evidence; all earlier failures, review rulings and wider
acceptance blockers remain recorded. Only this execution record and appended
current-status gain preservation notes; reviewed Go/spec/plan stay unchanged.

The preceding local-only paragraphs describe implementation delivery.
Commit/push success must be verified by the resulting Git commit and matching
GitHub branch tip; recovery uses this development branch, not main.
Ignored receipts, caches and generator archive are not among the49 saved files.
