# Combined boundary subphase timing — timeout during4096 evaluation, no completed duration

Approved one fresh ordinary-token PS7 diagnostic child after813a8c76, configured
30000ms, same combined order, unchanged product/assertions. No warmup, retry,
fallback, repair, reviewer, old-slot use or limit increase. This is throwaway
diagnostic tooling; it does not replace formal acceptance.

## Observed partial result

PS7.6.5 Core x64/admin=False;TimedOut=true,exit-1,total30080ms,owned-host-exited,
stderr empty,CaptureValid=true,Complete=false,FormalAcceptance=false,
Disposition=BudgetExhausted. Configured deadline30000; actual30080 includes
termination/collection overhead. Valid capture means internally consistent saved
prefix, NOT successful verification. No second process was launched.

Raw147 lines:95 timing events,52 ordinary lines,42 PASS markers,45 completed
pairs and5 open nested scopes. PASS markers cover new3 reflection checks,
fixture8 and first31 original35 groups through PassiveDescriptors. The original
ArrayAndIntegerBounds group, remaining3 original35 groups and all10 ownership
regression groups have no completed PASS. Final wrapper PASS/LIMIT absent.
No assertion failure or stderr error was saved before termination.

| Boundary subphase | Start ms | End ms | Elapsed ms |
| --- | ---: | ---: | ---: |
| InvalidInputs | 18229.301 | 18584.631 | 355.331 |
| Build4096 | 18586.391 | 19875.077 | 1288.686 |
| Check4096 | 19875.750 | Missing | Unknown |
| Evaluate4096 | 19878.997 | Missing | Unknown |
| Assert4096 | Not reached | Missing | Unknown |
| Check4097/Evaluate4097/Assert4097 | Not reached | Missing | Unknown |

Offsets relative to the first Run marker, not process launch. Check4096 includes
argument construction and its nested evaluation/assertion scopes; elapsed values
would overlap, so should not be summed. Build is only the original array allocation
and4096 observation construction. InvalidInputs includes original numeric/array/
shadow-Length refusal checks. No completed4096 API result or original equality
assertion exists this run;4097 was not reached. Do not substitute a deadline
difference for a completed evaluation time, extrapolate completion, or infer a
permanent hang.

## Important comparison with previous diagnostic

| Milestone from first Run marker | Previous group-only diagnostic ms | Current subphase diagnostic ms |
| --- | ---: | ---: |
| Start original35 | 1587.987 | 2269.330 |
| End original35/PassiveDescriptors | 5822.881 | 18180.642 |
| Start ArrayAndIntegerBounds | 5823.606 | 18181.547 |

Much of the difference appears BEFORE the4096 boundary work. Current original35/
ProgressMonotonic alone took2709.436ms versus previous201.804ms. These are two
single observations with differing instrumentation, not a controlled experiment.
Host contention, JIT/runtime variability and new timing-branch overhead have
not been distinguished or measured independently. No historical CPU/load trace
was captured. It is unjustified to attribute this timeout solely to4096 work,
claim a product regression, or infer that fixed metadata reuse caused the change.

The previous complete29.783s diagnostic remains its own observation. This new
incomplete run does not overwrite it, nor any old uninstrumented S3/R3 timeout.

## Fidelity and instrumentation boundaries

New folder retains previous combined diagnostic copies, not edits in place.
Ownership and reflection copies are byte-identical to their previous versions.
Wrapper differs only in diagnostic LIMIT wording. Original35 copy adds literal
subphase mode initialization/assignments, boundary markers, and marker branches
around the existing Get-OfflineSupervisionDecision and P1-Need equality statement
inside P1-Check. Original P1-Check/P1-Bad boundary call expressions, argument
evaluation order, helper scope, test bodies/assertions and combined order remain.

Mode is set only by literal4096/4097 diagnostic assignments, not input-property
inspection. The diagnostic P1-Check now evaluates marker conditions on OTHER
calls too (mode None): that overhead is real but unquantified. This is not
instrumentation-free evidence. Markers use Console output, not pipeline objects,
so product result capture is not intentionally changed. Product6CF0D354 is
imported unchanged from the parent directory; original test scripts are not
edited or directly executed.

All four saved files reverse exactly to previous combined copies before launch;
that previous layer had already reversed to originals. Static parsing passed.
147 immutable pins matched before/after:138 inherited +five previous packet
anchors +four new copies. New parser inherits corrected ApiRequired/HOST order
and freezes134 possible events/70 ordinary lines; observed95/52 are ordered
prefixes, with no truncated tail. Saved raw timestamps independently reproduce
all45 completed receipt durations and the five-scope stack. No raw/index/parser
repair or verdict upgrade, no retry. Success/error paths not exercised here
are not claimed validated.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `B0071B33581B40376A6E57A4DE1EE3409BBD6B4DAC20A458210A1CA635D43BA3` |
| run-once.ps1 | `0AD6C92E486547D4562F3287D29425FA4A38061BBA5BC5211770E3E49BE4A169` |
| invocations.json | `A511B9D091441459D3E52B5B8DE6843AF8B7142DAEE5806A0F44BB857E82AEFA` |
| D1.stdout.txt | `7160EFACB553667B288573F43E66F706B6A07920A7490376BCB19EA419BCA692` |
| D1.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |
| probe.ps1 (portable copy) | `6960F0B2B39FDF1BCE6D2415EE4368164F237E2F061DB13F8193378F25861466` |
| offline-supervision-policy.tests.ps1 (portable copy) | `47615E6D77246E9B38EB8998D53FD1233AD2993A519FA484E52FFF38C6BD8B8F` |
| offline-supervision-policy.ownership.tests.ps1 (portable copy) | `03A8B7A39F15BA2AF234DC3B0D2DF638DBDD16C7287BE7854357E437D1C519DE` |
| offline-supervision-policy.reflection-reuse.tests.ps1 (portable copy) | `BED41B65AB4B785F808559FE1487DE9F57D55EE4E440F931A027F4514ECEA626` |

Ignored packet
`.superpowers/sdd/2026-10-17-p1-combined-boundary-timing/`
retains scope/controller/manifest/index/raw; all older evidence retained. Portable
copies/report/status/recovery are saved on the existing development branch,
not main. Cloning does not restore ignored evidence/runtime or authorize replay.
Hashes identify measured local bytes, not independent attestation; checkout
line conversion can change them.

## Handoff

New diagnostic1/1 spent. Product/original suites unchanged. Previous reflection
reuse3/4/one frozen, extended-view3/7/four frozen, old timeouts and review0/1
ineligible remain. Candidate unaccepted,I1/P1 unresolved; no PS5/main/C12/project/
live/runtime/release acceptance.

Next prioritize a read-only audit of instrumentation differences and measurement
conditions, then scope any controlled paired measurement explicitly. Do not
continue speculative product optimization, retry this run, spend frozen slots
or raise the30s gate from these partial data. Neither large-input evaluation
nor instrumentation/host effects has a unique causal attribution here.

## All completed pairs

Relative rounded elapsed milliseconds; nested Stage durations overlap children.

| Kind/name | Start ms | End ms | Elapsed ms |
| --- | ---: | ---: | ---: |
| Reflection/InputMutation | 321.864 | 537.874 | 216.009 |
| Reflection/DecorationAfterClean | 538.286 | 589.485 | 51.199 |
| Reflection/DescriptorAfterClean | 589.736 | 624.336 | 34.600 |
| Stage/ReflectionReuse | 279.592 | 624.859 | 345.268 |
| Ownership/FixtureConstructor | 652.872 | 672.620 | 19.748 |
| Ownership/FixturePrimitiveMatrix | 672.933 | 753.150 | 80.217 |
| Ownership/FixtureDescriptors | 753.437 | 766.115 | 12.678 |
| Ownership/FixtureInputIdentity | 766.385 | 780.923 | 14.538 |
| Ownership/FixtureIsolation | 781.206 | 791.225 | 10.018 |
| Ownership/FixtureWrappedInputIdentity | 791.500 | 795.314 | 3.814 |
| Ownership/FixtureExtendedView | 795.577 | 808.840 | 13.263 |
| Ownership/FixtureComplete | 809.362 | 2089.476 | 1280.115 |
| Original35/ApiRequired | 2269.330 | 2273.358 | 4.029 |
| Original35/EmptyPrepared | 2273.998 | 2307.600 | 33.602 |
| Original35/NormalStages | 2307.988 | 2430.772 | 122.784 |
| Original35/AbsoluteDeadline | 2431.155 | 2486.803 | 55.648 |
| Original35/StopBoundary | 2487.230 | 2941.874 | 454.644 |
| Original35/TotalDeadline | 2948.597 | 4005.304 | 1056.706 |
| Original35/StartupConsumesBudget | 4006.437 | 4389.831 | 383.394 |
| Original35/EqualTimeSequence | 4391.503 | 4739.296 | 347.793 |
| Original35/BackwardTime | 4741.007 | 4955.218 | 214.211 |
| Original35/SequenceRefusal | 4957.397 | 5171.853 | 214.456 |
| Original35/SessionClockRefusal | 5173.073 | 5615.824 | 442.752 |
| Original35/ActorDirection | 5617.086 | 5788.707 | 171.621 |
| Original35/ArmPrerequisites | 5789.410 | 6643.777 | 854.367 |
| Original35/ReleaseOnce | 6645.728 | 7481.646 | 835.918 |
| Original35/ProgressMonotonic | 7483.035 | 10192.471 | 2709.436 |
| Original35/ExitBeforeEOF | 10193.784 | 10514.973 | 321.189 |
| Original35/UnknownExit | 10516.274 | 10790.468 | 274.194 |
| Original35/NonzeroExit | 10791.592 | 11105.720 | 314.128 |
| Original35/DataBounds | 11106.330 | 12162.314 | 1055.983 |
| Original35/ControlBounds | 12162.953 | 13518.611 | 1355.658 |
| Original35/ReadFault | 13519.731 | 13780.652 | 260.921 |
| Original35/SupervisorLost | 13781.896 | 13970.803 | 188.907 |
| Original35/LauncherLost | 13972.038 | 14172.412 | 200.373 |
| Original35/FailureSticky | 14173.191 | 14427.513 | 254.322 |
| Original35/PhaseBudgets | 14428.590 | 15887.911 | 1459.321 |
| Original35/CleanupUnknown | 15889.634 | 16476.279 | 586.644 |
| Original35/PersistRefused | 16477.221 | 17044.600 | 567.379 |
| Original35/TerminalTrailing | 17045.539 | 17289.225 | 243.687 |
| Original35/InvalidScenario | 17289.689 | 17655.785 | 366.096 |
| Original35/InvalidObservation | 17656.329 | 17995.473 | 339.144 |
| Original35/PassiveDescriptors | 17996.355 | 18180.642 | 184.288 |
| Subphase/InvalidInputs | 18229.301 | 18584.631 | 355.331 |
| Subphase/Build4096 | 18586.391 | 19875.077 | 1288.686 |

## Open scopes at final saved event

These starts are observations, not completed durations.

| Kind/name | Start ms | Status |
| --- | ---: | --- |
| Run/CombinedRegression | 0.000 | No end marker |
| Stage/ExistingAll | 625.090 | No end marker |
| Original35/ArrayAndIntegerBounds | 18181.547 | No end marker |
| Subphase/Check4096 | 19875.750 | No end marker |
| Subphase/Evaluate4096 | 19878.997 | No end marker |
