# Single uninstrumented Envelope validation — 2026-10-15

## Approved scope and implementation

The user approved one independent unpatched fixed Go1.26.5 copy and one fresh
original FuzzIssuerEnvelope validation. Original toolchain remains read-only.
Version/origin limits60s each, compile600s, fuzz5s/parallel1/test timeout30s,
native execution60s. No retry, Pair, full matrix, main merge, downloads or security
changes. These limits and existing native ownership/cleanup gates are unchanged.

`docs/diagnostics/isolated-fuzz/envelope-uninstrumented.ps1` adds a separate entry:
complete11536-file manifests, unique safe paths, three original source/executable
pins, independent copy receipt, live copied-tree and original project binding.
Diagnostic environment settings and GODEBUG are cleared in the child environment.
The compiler log must identify the copied context/fuzz compiler sources and the
original Envelope callback/parser package. Fresh build/fuzz caches are used.
Success requires natural verified completion, worker evidence, positive execution
count, terminal PASS, no diagnostic hook output, unchanged product/source/binary.
Finalization precedes success publication. Acceptance always remains false.

## Tests and independent review

Initial RED confirmed the new entry was absent. A positive-path test caught a
PowerShell automatic `$Matches` variable collision; renaming the local regex
collection fixed that new driver bug before any actual Go execution.
New28 plus existing sample22/compile19/preparation19/planning21 assertions passed
on PowerShell7.6.5 and WindowsPowerShell5.1:109 scoped assertions per host.
These are synthetic boundary tests, not additional Go/fuzz runs.

Independent read-only review found no Critical/Important issue. Limits: branch
tests replace the expensive combined toolchain validator; direct receipt/link/live
integration relies on actual validation. Missing terminal PASS and zero-execution
negatives are not independently exercised. A wrapper exception without returned
receipt does not prove no process launched: SamplingAttempted and consumed identity
must prevent replay. No aggregate process count proves individual process identity.

## Actual attempt

Reserved copy identity: `uninstrumented-001`; run identity: `envelope-observe-005`,
under `.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator/`.
The reference is `framework-001/copy-receipt.json`, SHA256
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55`.
Only its frozen BeforeManifest and Source are used; the instrumented framework
tree is not the copy source. The single copy completed successfully:11536 files,
218649457 bytes. Both original-before and copied manifests match the frozen
reference; the copy routine also compared original-after against original-before.
Copy receipt SHA256:
`DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.

Copy preparation was unusually slow. Read-only diagnosis found all11536 paths/
single-link checks readable (21334ms), and an additional original full-manifest
comparison passed (45138ms). Current-command-only runtime statistics advanced;
the copy subsequently returned exit0. A default-sandbox CIM diagnostic was denied;
the scoped elevated read succeeded without changing process/system settings.
No exact latency cause, security-software attribution, or test failure is inferred.
Systematic debugging guidance kept these checks read-only and prevented guessed
fixes or a second copy.

### One actual validation — PASS, not product acceptance

The one authorized005 invocation returned exit0. Version/origin checks verified
Go1.26.5 windows/amd64 and the independent copy's GOROOT/GOTOOLDIR plus this
worktree's go.mod. One build ran from12:15:24.4202368Z to12:17:02.561417Z
(98141.1802ms), exit0, no timeout, natural verified shutdown and final0/empty.
TotalProcesses517 is aggregate accounting, not individual identity evidence.
Compiler logs passed original context/fuzz and Envelope callback/parser checks.

One execution ran12:18:05.2070554Z–12:18:12.2524212Z (7045.3658ms), exit0,
TimedOut=false, RootExited=true, NaturalShutdownVerified/CleanupVerified=true,
TerminationRequested=false, final active0 with empty member arrays, shutdown0ms.
TotalProcesses3 is aggregate only; live member sampling was not enabled.
Stdout ended PASS; stderr recorded5/5 baseline inputs, one worker and74867
executions,6 new interesting inputs (total11), not6 defects. No diagnostic hook
record appeared. No retry occurred. Result Uninstrumented/SamplingExecuted/
ObservationValidationPassed=true, failure=null, Acceptance=false; no stopped.json.

The binary is15323136 bytes. Full live copy/source/binary validation passed
before success publication; all701 before/after product Path/SHA256 pairs match,
including the original fuzz callback. The unmodified short run did not reproduce
the historical deadline. Different caches/input sequence/entry prevent causal,
performance or product-repair conclusions, including comparison with004's73921.

After the invocation, one final read-only original-source manifest check passed:
11536 files/218649457 bytes,42135ms, completed12:20:37.3011060Z. The reference and
copy receipt hashes above were rechecked unchanged. This establishes original
toolchain immutability across this attempt without querying any historical PID.

| Local retained evidence (relative to005 unless noted) | SHA256 |
| --- | --- |
| plan.json |6ED4D9A0CB8FC391FE8B49FFF898CD95A8ABE9ADE5DD8CCE71C75790B6D855CA |
| binary-receipt.json |77182A10B7CBBE81D36D256C3A45447A6FAF40F0E853BA6468BFEAC751D98FC3 |
| build/receipt.json |1B32A547B86E4E3F4C9CA1EC960A6B6D9695A6CAF5760120D8FC0577E8DFB1CD |
| build/stderr.txt |172B4FC93B8BAD35D9681DE796E63550F09867B595EE640452C6D74FD33DA2F7 |
| execute/receipt.json |7F916C38C6337502C9F256A4957A4FD3BB1F5F1C4DB3B347C90588DA8FCD347F |
| execute/stdout.txt |C26DE83ABDC9496CD1301470918EC39ECCA1CF389EF0AE1C6504DA1800D1C431 |
| execute/stderr.txt |CAC7A53DF6F860533B1EC4D6C6EF488D6D1A8326AB959E347318CD44F15DB381 |
| sample-result.json |0D620877AC156427B576103D44B8C80185AD61FD375BC302F86DBB7CFACCE6CF |
| original-source-postcheck.json |1FE2DFF7C31A8A98D389A4E945522D76553EA0066B80DE5AF9FD1AF99DC94020 |
| before/after-product.json |5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |
| nodeissuer.test.exe |1C682F0021C23F19630B2603167403A215FC99DA4939D048B2BD1EAA801FDE76 |

All times above are UTC. Both actual identities are consumed, not reusable.

## Recovery and remaining gates

Recover source/tests and this progress record from the existing development
branch, not automatically main. Read current-status.md and
docs/runbooks/repository-recovery.md. Raw receipts, toolchains, binaries and caches
remain ignored and are not restored by cloning. Independently prepare the pinned
toolchain and offline module cache on another computer; unit tests accept explicit
Toolchain/ModuleCache parameters. Do not replay consumed local identities or infer
permission for another actual invocation from test commands.

One short uninstrumented result cannot establish the historical deadline cause or
clear matrix3, origin001/87, Pair, C11/C12/Task10/B01/production/main acceptance.
Historical evidence and the earlier instrumented004 sample remain distinct.
