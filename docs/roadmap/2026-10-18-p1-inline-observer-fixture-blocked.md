# Inline observer offline batch — fixture setup blocked

Date: 2026-10-18. Base commit: 12e1aa1d9919ab2dc4cb72452b118ce2eb175e64.
Branch: codex/c12-b01-task9-coordinator. Status: UNFINISHED / NOT ACCEPTED.

## What changed

User authorized continuing the previously proposed minimal controller correction
and one bounded offline validation. The failed capture.ps1 remains byte-identical
(SHA256 79D41F8AC6F1EF1A0FE316183F31A0C38C349256FC40D60846E38F5648BA577D).

New capture-corrected.ps1 changes only the wrong inventory hash and the output
packet name (apart from trailing blank-line normalization). No observer, harness,
candidate, source inventory, timing, timeout or safety guard was changed.

Read-only checks before the correction found exactly one unknown fixed hash
literal in the old controller, matching the previous recorded failure. The
corrected script has 10 fixed SHA256 literals, all matching the independently
hashed source files, zero parser errors, and exactly the two intended substitutions.
All 615 historical file pins were independently rehashed without drift.
This static pin regression is NOT observer runtime equivalence evidence.

Corrected controller SHA256:
D07D39EF47E8BD82C1AB8EBA30CF43984F086C2AA18237A39BC284FDB141B2C8.

## Single offline run

One ordinary-user PS7 offline child host was launched by the corrected controller.
The allowance was consumed before Process.Start. Limits remained 30,000 ms execution,
2,000 ms owned-process exit, 2,000 ms stream drain. No candidate import or call,
actual ownership matrix, product test, previous-slot reuse, retry, or repair occurred.

Receipt:
- Start UTC: 2026-10-18T13:23:11.6807398Z; owned child PID 10460.
- Elapsed: 971 ms; exit 1; no timeout.
- OwnedExited / StreamsComplete / PinsIntact: all true.
- Frozen pins: 627; independent read-only rehash after failure: zero drift.
- Stderr: 362 raw bytes; ControllerError: null; Summary: null.
- ProductCalls / CandidateImports: 0; FormalAcceptance: false.

Verified paired cases (old and new each before/after, independent literal outcomes):
1. Int32/clean: PASS then PASS.
2. Int32/selected: PASS then PASS (selected decoration skipped).
3. Int32/descriptor: PASS then FIXTURE_INVALID|fixture|field-base-changed.

Case 4, Int32/storage, then failed during fixture construction in the OLD-side
run: offline.tests.ps1 line 116, "Storage fixture did not change type".
The harness assigns R-Wrapper output to the real note property's Value, then
checks that R-Type of that stored Value differs from its captured Stored type.
The guard observed equality and threw BEFORE the after-scan. The corresponding
new-side storage fixture was not reached. Thus only 3/45 paired cases completed;
the batch is neither 45-case success nor evidence of an observer disagreement.
The remaining cases were not completed, and no final RESULT line was emitted.

The fixture's assumption that this assignment changes actual stored type was
contradicted by its own guard. The retained record does not include actual type
names or identify exactly which constructor/binding/assignment step retained or
removed the wrapper. Do not claim an exact deeper cause without further evidence.
No fallback, weakened setup assertion, swapped expectation or post-launch edit
was applied. All raw files remain untouched.

## Preserved machine-local evidence

Packet: .superpowers/sdd/2026-10-18-p1-inline-ownership-observer-corrected

| File | SHA256 |
| --- | --- |
| frozen.json | B5CCFC9EFB1EACB8448FF5432ACBE100849A9288B23EF14CAD6B5F7CC74599C8 |
| invocation.json | 27817AF3873AEE1C6504686213C146DF79C71FFE4F633FD42EEB36832A65DB75 |
| stdout.txt | F23DD4DC15D2E579450A5B24D0D20E6C9D089017E855C4111DB565D6E21BC5FC |
| stderr.txt | AB59B4FD81CED45DE107FECE95B92B7DD3A9C20E89909286263641687F83ABE3 |
| fixture-inventory.json | A055241CF41A2D0D19C6A3DC7D1A1C8E00D0BE6E46F8D3A5D0168886137382E3 |

Raw/receipts stay ignored; the portable harness and its fixed literal 45-case
inventory remain committed for recovery. A fresh clone does not restore private
historical receipts or confer execution authorization. Read this report before
the previous preflight report; that earlier controller never launched a child.

## Next bounded work proposal

Preserve this failed fixture and result. Before proposing a revised equivalence
batch, characterize the storage setup using a separately scoped, synthetic-only
diagnostic: capture direct stored type before/after each wrapper-construction and
assignment boundary, with SDK reflection retained, no candidate imports/calls,
no ownership suite, and a fresh explicit small bound. Do not rerun this batch or
use old slots. Establish a real type-changing fixture before changing any tests;
do not remove storage-type coverage or weaken its guard to manufacture a pass.

After a justified fixture correction, full paired offline validation is still
required. The optimized observer has no complete equivalence or performance
validation. No main merge, I1 closure, security/runtime acceptance or C12 completion.

## Handoff

Only the new corrected controller and the latest result/status/recovery documents
belong to this checkpoint. Existing frozen source files and all historical
reports remain unchanged. The repository's standing preservation instruction
requires saving this clearly unfinished development checkpoint on the existing
development branch, not claiming acceptance.
