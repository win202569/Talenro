# Error reference propagation diagnostic — actual result

Status: STOPPED_PARENT_VALIDATION; NOT_ACCEPTED. Useful saved measurements
exist, but the original parent Exact=false must not be overwritten.

Executed the [approved minimal diagnostic plan](../superpowers/plans/2026-11-01-offline-error-reference-diagnostic.md)
after explicit execution approval “是”. Baseline
fdd3240f1b05b303834ce1220eb6b97c58a0d618; existing development branch.
NEW1diagnostic host/0targets/0compilations/0reviews, independent of the frozen
integration allowance. No original writer/capture/driver invocation or fix.

## Actual preparation and one host

Preflight verified clean branch/origin, absent new source/packet, original
collector SHA and failed continuation archive SHA. Created only new local LF
attributes and fixed reference-probe.ps1. AST/LF/JSON checks passed;8inputs
pinned,4source/plan/approval snapshots preserved. The actual collector function
was extracted as AST text; capture.ps1 was never executed or imported.

Plan SHA256:8BB3DD17AC0F86B412E1BB48DAFA6AA408385DC5E78804069EEA52AA7577F71C.
Probe SHA256:494A5945B278A0F8DC299EECA3786839BEF722E65EDC884D4871D881AE0E1496.
Binding SHA256:E4484D836278B97B99743F42A30B98F01384D2AD3DDBA5B29D3BCA228A8EC2BA.

One durable consumed intent preceded launch. Actual diagnostic PID10448 exited0;
owned exit and original stream drain true, stdout/stderr both0bytes, no timeout
or termination, post8input hashes intact. Six case rows and one provisional
COMPLETE row were written (15228bytes), with seven real missing-file operations.
No extra runtime/source patch, experiment, corrected checker or retry followed.

## Saved live-reference measurements

The matrices were computed from retained live objects inside the probe BEFORE
JSON projection. The following is a read-only reading of those actual rows,
not a new parent acceptance verdict:

| Case | Inner/outer ErrorRecord same | Exception same | Collector entries |
| --- | --- | --- | ---: |
| DirectOuter | one observation only | one observation only |1|
| CatchBare |false|true|2|
| CatchRetained |false|true|2|
| FinallyBare |false|true|2|
| FinallyRetained |false|true|2|
| IndependentSameText |false|false|2|

Thus the tested rethrow paths exposed different ErrorRecord objects referring
to the SAME underlying Exception, and the reference-only collector retained
both. The independent two-fault control exposed different records AND different
exceptions. In this simplified probe, changing bare throw to throwing a retained
record INSIDE catch/finally did not change that observation. This does not test
every throw placement or PowerShell environment.

This supports the earlier duplication hypothesis more specifically than equal
serialized messages could. It does not restore the missing first.Count/object
identities of historical IntentDisposeOnly, exercise compound Write/Dispose
faults, establish original writer correctness or fix I1/I2. It does not justify
deduplicating unrelated errors by message/ErrorId.

## Actual parent validation failure

The parent command exited1 with
`REF-PARENT: bounded ordered collection event` at its first case check. Its
final.json remains Exact=false and Measurements empty. Later parent validation
steps did not all run; the probe's own COMPLETE row does not supersede this.

Source-level trace of the preserved original parent command identifies a
singleton-label bug: `$labels=if(...){@('Outer')}` emits a single string from
the conditional pipeline; subsequent `$labels[0]` selects its first character,
while the actual first event label is `Outer`. The saved event counts/types are
0→1/System.Int64 and its observation index0, consistent with the label comparison
being the mismatch. This is an author-written parent checker defect, not a
platform/environment failure or a successful result hidden by display formatting.
The exact original parent command is preserved as tool-input JSON, not another
executable source. No fixed expression was substituted or original run regraded.

## Stop, recovery and recommended revision boundary

The new diagnostic allowance is consumed/exhausted:1host/0targets/0compiles/
0reviews. Original v3 remainder6/7/1/conditional1 stays FROZEN; older stopped
budgets/records and the earlier failed independent full-pin audit are unchanged.
All5capability flagsfalse; no I1/I2/C1/C2 closure or runtime acceptance.

Ruling: the approved finite Native/PowerShell diagnostic replaced skill-default
setup/Bash/TDD/full-suite/task-done reruns, review/fix/deletion. Cost: a limited
measurement, not complete production/integration validation. Once parent checking
failed, preservation took precedence over correcting it or launching again.

Portable-final.json preserves16actual files/9directories including root and
4snapshots, zero-byte original streams, complete measurements, parent failure,
actual tool objects and exact parent input. Decoded byte lengths/SHA256/equality
verified against originals. Portable SHA256:
A12F16F6F42A8365B0D7D34BE8D70771BBDF9FAD96F5FC43EBFDC0B8B37E74D6.

Next revision should address both the parent singleton-label handling and the
collector/reference contract. Merely changing throw syntax is not supported as
a sufficient repair by these rows. Consider a clearly defined single collection
owner and/or same-underlying-Exception identity handling while preserving the
original first record and distinct real faults; any altered production contract
and frozen input transition needs separately written approval. No source change,
regrade, revalidation, retry or integration budget release follows automatically.
