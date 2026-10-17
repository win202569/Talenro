# Independent full verification — PS7 timed out; PS5 not launched

Aftera7017a47 the user approved two fresh ordinary-token children: PS7 then PS5,
configured30000ms each, unchanged complete relevant test entry, original product/
assertions untouched, CPU/wall/hash receipts. Stop on any failure/timeout/capture
problem; no retry, warmup, fallback, diagnostic copies, old-slot use or main merge.

## Actual result

Only PS7 was launched. PS7.6.5 Core x64/admin=False,exit-1,TimedOut=true,total
30055ms,process CPU37500.000ms,owned-host-exited,stderr empty,CaptureValid=true,
ExpectedResult=false,FormalProjectAcceptance=false. Configured cap30000; actual
total includes55ms termination/collection overhead. Capture validity is consistency
of a saved prefix, NOT a passing verification.

Target was the EXISTING unchanged
`docs/diagnostics/isolated-fuzz/offline-supervision-policy.reflection-reuse-regression.ps1`.
No added helper conditions, subphase markers or diagnostic test copies. Wrapper
runs new reflection3, ownership fixture8, original35 and ownership regression10
in the existing order. Product6CF0D354 unchanged. Target SHA256:
`0B7692FE426407AA9431D54BDB1AA964216FB695CD5A202045ABDDCEB5C61069`.

Saved63 stdout lines contain52 PASS markers:
- All3 reflection-reuse checks.
- All8 fixture groups.
- All35 original groups, including ArrayAndIntegerBounds and Complete.
- Ownership ApiRequired,DecoratedInteger,DecoratedBoolean,DecoratedString,
  RootScalars and ObservationScalars (6/10).

Ownership PassiveDescriptors,SequentialIsolation,PlainControls and Complete have
no completed PASS. Ownership final LIMIT and wrapper final PASS also absent;
the complete inventory would have69 lines/57 PASS markers. No assertion failure
or stderr error was saved before termination, but missing groups cannot be
assumed passing. Last completed PASS does not identify exact executing statement.

PS5 has NO invocation row, child launch or stream pair. The second slot is frozen
under the approved failure-stop rule, not spent or treated as cancelled success.

CPU is cumulative process/thread time, including runtime/JIT/setup/logging, so
it can exceed wall elapsed. It is not per-group profiling, CPU percentage or a
historical host-load trace. These totals cannot uniquely attribute the timeout.
No internal per-test timings or unique cause are inferred from this run.

## Meaning for acceptance

The original relevant combined flow still cannot demonstrate completion inside
its unchanged30s gate in this observation. Previous prefix PASS and the separate
29.783s instrumented full-copy diagnostic do not override this original-flow
failure. This does not prove a particular security assertion is false, an exact
deadlock, or a unique timing cause. It DOES leave full verification blocked.

No PS5 proof, independent review, mutant/final frozen runs, main integration or
I1/P1 closure was produced. Candidate6CF0D354 remains unaccepted. No broader C12/
project/live/runtime/release acceptance follows from the completed groups.

## Integrity and preservation

Controller statically parsed, froze166 immutable pins before launch:158 inherited
+seven previous paired packet anchors +existing PS5 executable. Both engines,
product and original sources are pinned; expected output inventory follows actual
Context/ApiRequired/HOST order. Slot row consumed before launch. Hidden ordinary
no-profile/no-shell child, separate UTF8 streams, owned-handle deadline cleanup.
All166 pins/controller/raw hashes matched afterward; saved63 lines independently
matched expected prefix/52 PASS count. No raw, source, parser or verdict repair.

PS7 engine SHA256:
`362A356CE7F0940EC74F73A8FC2C990A2CC24A38A11C90BBD8ECA947110AD139`.
PS5 executable pinned but NOT executed:
`60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`.
Start2026-10-17T19:05:12.1230307Z. Successful full-completion and PS5 parser paths
were not exercised or claimed validated. Timeout owned-exit/stream capture path
did execute here; other error cases are unexercised.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | `3B921E58B790DCB73DD00C98EFA26558300B8C82FD38A2656C9B29B62213736B` |
| run.ps1 | `C11BCBB60D02F7F10C83BC4C50D281FFBFE4F034E0F30084BB2B64663A0567AA` |
| invocations.json | `B7E087CB03B9AFE4EB4FBC709F898BD8E43487689FA269713613D67DEAF0F4EF` |
| PS7.stdout.txt | `6BB32585B37138A40457375DFDEE6720DD44265AB98526A2851E0C0BA8F27A13` |
| PS7.stderr.txt | `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855` |

Ignored packet
`.superpowers/sdd/2026-10-17-p1-independent-full-verification/`
keeps scope/controller/frozen manifest/index/raw. All older packets and their
verdicts preserved. Existing portable entry/source/tests plus this report/status/
recovery restore inspectable unfinished progress from the development branch;
ignored evidence/runtime and replay authority do not clone. Hashes identify local
bytes, not independent attestation; checkout line conversion may change them.

## Handoff

New1/2 spent,PS5 remaining1 frozen. Previous reflection-reuse3/4/one frozen and
extended-view3/7/four frozen stay separate; all diagnostics retain their original
spent counts,review0/1 ineligible,I1/P1 unresolved. No automatic remaining-slot
resume,limit increase,security-check removal,whole-suite split-to-pass or retry.

Next prepare a narrowly scoped semantic-preserving optimization design for the
already measured validation hotspot, retaining all per-value defenses and a
separately approved complete verification budget. This run does not justify
another unmodified replay or a claimed root cause. Product implementation needs
the design/approval step; no optimization was attempted in this verification.
