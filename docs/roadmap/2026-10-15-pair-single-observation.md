# Pair single observation — 2026-10-15

## Approved bounded scope

The user approved a Pair-only extension to the existing compiled-sample flow:
reuse the validated uninstrumented toolchain and Envelope005 package binary,
fresh `pair-observe-001`/caches, exactly one native invocation of
`FuzzIssuerPair`, fuzz5s/parallel1/test timeout30s/native60s. No copy, recompile,
version/listing/Envelope invocation, retry, full matrix, main merge, downloads or
security changes. Any failure stops. Acceptance remains false regardless of PASS.

`pair-sample.ps1` is a separate entry; existing Envelope/native helpers are
unchanged. Planning checks owned fresh paths, offline dependencies, the original
fuzz_test.go hash, full original11536-file copy binding, successful bound compile
receipt, exact compiler/environment/arguments, compiler log tokens, prior result
labels, binary/source receipt and current product manifest. Eight retained005
artifacts are hashed and revalidated after the invocation, together with the live
copied tree and current product inventory. Finalization precedes result publication.

Both fuzz targets occur in the pinned original fuzz_test.go. The copied fixed Go
source `src/cmd/go/internal/load/test.go:752–757,826–830` collects all valid Fuzz
functions and emits all entries into testing.InternalFuzzTarget. The prior
compile's Envelope selection does not remove Pair registration. This is source
evidence for reusing the same binary, not an extra executable listing operation.

## Tests and independent review

Initial RED failed because the Pair entry did not exist. New45 assertions PASS
on PowerShell7.6.5 and WindowsPowerShell5.1, including reruns after clarifying test
fixture comments. Existing109 scoped assertions also passed on each host:
154 assertions per host in total. Full project/matrix execution was not run,
as explicitly excluded by the approved scope. Tests generate portable receipts, binary placeholder,
tool/dependency placeholders and compile logs; no previous machine cache or raw
session is needed, and no unit executable is actually launched.

Full toolchain scanning and native execution are replaced in unit tests; product
enumeration is fixture-backed. Real JSON/files/hash checks, compile parsing,
manifest comparison, planner, completion gates and finalization are exercised.
Negatives cover bad prior labels/compile arguments/environment/binary/log,
consumed/wrong identities, exit/timeout/shutdown, missing PASS/worker evidence,
zero executions, diagnostic output, artifact/product/copied-source drift and a
native-wrapper exception. An exception with no returned receipt is not proof that
nothing launched: SamplingAttempted=true and the consumed identity prevent replay.

Independent read-only review found no Critical/Important issue and conditionally
approved the one invocation after tests. Its misleading fixture comment was
corrected. Remaining minor coverage limit: no dedicated result-write-failure
fixture; current ordering writes the result before emitting success. Review did
not independently verify native behavior, actual full manifests, host tests,
hostile concurrent filesystem mutations or broader acceptance. These helpers are
not transactional filesystem isolation. Operational audits remain necessary.

## Actual observation

The single authorized `pair-observe-001` invocation returned exit0. Before
planning, ten known prerequisite hashes matched the prior approved record:
reference/copy receipts plus the eight retained005 artifacts. All eight bound
artifacts remained unchanged after the invocation. No recompile or replacement
binary, version/listing command, Envelope execution or retry occurred.

Actual executable was the retained005 `nodeissuer.test.exe`, SHA256
`1C682F0021C23F19630B2603167403A215FC99DA4939D048B2BD1EAA801FDE76`.
Native UTC12:35:20.7750667Z–12:35:26.872603Z,6097.5363ms. Receipt: exit0,
TimedOut=false, RootExited=true, NaturalShutdownVerified/ShutdownEvidenceComplete/
CleanupVerified=true, TerminationRequested=false, final active0 with empty member
arrays. Natural observation/shutdown5ms. TotalProcesses3 is aggregate accounting,
not individual process identity proof; live member sampling was disabled.

Stdout ended PASS. Stderr recorded4/4 baseline inputs, one worker, only5 executions,
zero new interesting inputs (total4), and no diagnostic-hook line. The approved
positive-execution gate passed, but this is very limited coverage, not sufficient
fuzz validation or evidence of performance/correctness across the input space.
Do not infer why throughput was low or compare it causally with Envelope's74867.
No retry or longer run was used to replace this observation.

Saved result: Target=FuzzIssuerPair, Uninstrumented/SamplingExecuted/
ObservationValidationPassed=true, failure=null, Acceptance=false. No stopped.json
or new build directory exists. All701 before/after product Path/SHA256 pairs match;
full copied-tree/source/binary checks passed before publishing success.

After the attempt, an additional read-only original-toolchain full-manifest check
passed:11536 files/218649457 bytes,130335ms, completed12:39:16.7005362Z.
Original source is unchanged. Frozen reference/copy receipt hashes were rechecked
unchanged as well: reference
`AA86DADA2C7EF1BB591C78504FB8B5C168D35D836493A574ABC37DDA84880B55`,
copy `DD62FDC0704E8ADEE49F77AD7E2A91D3A1CDFC6E430ABD9AADBFC74D74F04D10`.
No historical PID was queried, and no original toolchain file was written.

| Local retained evidence (under pair-observe-001) | SHA256 |
| --- | --- |
| plan.json |7D74942A63FBD50C8BD9476BD0D9CD8E84A0C634907EC59DED7E82E3151B5861 |
| execute/receipt.json |65FDC0E8AC4F67DFB47DB618D293415078D67361666721E7D351FFEAA4EDC272 |
| execute/stdout.txt |C26DE83ABDC9496CD1301470918EC39ECCA1CF389EF0AE1C6504DA1800D1C431 |
| execute/stderr.txt |63F952A184115189F14861056A02665C5C61350AE52D2386E9039BA4918E5031 |
| sample-result.json |1855B2EA1D6F4F957F53640242EDCA9A0A0C2222B0B784485C093D40605C6A76 |
| before/after-product.json |5D711EE2E412812A09B4681E6287413B4F27DE842ED53C38496AABD71C18D2E2 |
| original-source-postcheck.json |6AF373C50DC74FE353F3B36F45363D5AFE050C7F55A75A16B5D0A9790E84EB44 |

The local identity is consumed; do not replay it.

## Recovery and remaining gates

Necessary helper/tests and progress are saved on the existing development branch,
not main. Read current-status.md and docs/runbooks/repository-recovery.md.
Raw receipts, toolchains, caches and binaries remain ignored and are not restored
by cloning. The new Pair unit tests are self-contained; old test suites retain
their documented independently prepared prerequisites.

To perform a future real validation on another computer, independently prepare
the fixed toolchain, offline dependencies and bound uninstrumented compiled
artifact; obtain an explicitly approved fresh identity. Do not replay001 or treat
this report as authority for another run, rebuild, full matrix or main merge.
Historical matrix3 remains24 pass/Envelope fail/Pair unrun in that original run.
A new separate Pair sample cannot rewrite it or establish the old deadline cause.
C11/C12/Task10/B01/production/main acceptance remains open.
