# Inline observer preparation — preflight blocked, NOT validated

Date: 2026-10-18. Base: ba1a68a4ceb4f1d5f10cd75fd65bda727793995f.
Branch: codex/c12-b01-task9-coordinator.

## Outcome

The single controller invocation exited 1 during preflight, before packet
creation, allowance consumption, or launch of its offline child host.
No offline fixture ran (0/45), no candidate was imported or called, and no
ownership matrix or speed measurement ran. This is an unfinished checkpoint.
The proposed packet .superpowers/sdd/2026-10-18-p1-inline-ownership-observer
does not exist. There is no new frozen manifest, invocation receipt or raw child
stdout/stderr. The controller's preflight failure remains visible in the session.

The guard reported Pin drift for stage-timing-ownership/output-inventory.json.
Read-only comparison established:
- Actual SHA256 and inherited frozen SHA256:
  7514F6C3348CF02EE59769F1EA895883E3A92F4C35373F7F7D963686156561D3.
- Mistaken new controller literal:
  7514F6C3348CF02EE61569F1EA895883E3A92F4C35373F7F7D963686156561D3.
- Root cause: global text replacement of inherited count 597 with 615 also
  altered those three characters inside this unrelated hash literal.
- All 615 inherited file pins were rehashed read-only after the failure:
  zero drift. The original inventory was NOT changed.

Per approved stop-on-inconsistency scope, no controller repair, retry, fixture
execution or candidate run followed. No old allowance was unfrozen. One controller
invocation occurred; zero offline child hosts launched. Do not describe this as
either a 45-case failure or success, or automatically reuse its unconsumed slot.

## Prepared files and checks

New independent leaf: docs/diagnostics/isolated-fuzz/inline-ownership-observer.
It contains the optimized whole-test COPY, offline-only 45-case equivalence
harness, one-shot capture controller, README, source pins and attributes.

Only V-Others differs in the optimized test. Small helper calls are expanded into
the loop, retaining public SDK reflection/trampoline calls, fresh object-array
arguments, all identity/type/value comparisons, owner-and-field skip logic and
held/fresh metadata checks. Original helper definitions and selected checks remain.

Static parser checks: all three new scripts have zero errors.
Whole-source check: identical outside V-Others after CRLF normalization, including
39-case matrix, literal expectations, two-phase scan sites and timing output.
This is source preservation, NOT runtime equivalence or performance evidence.

The harness was written before the optimized implementation. Its fixed expected
45 cases are described in README: five storage kinds, clean/selected skip and
descriptor/storage/base/type/held contamination, synthetic reread branch probes,
sharing controls, same-owner/different-field checks, late leaf and empty list.
Only ten allowlisted functions are extracted from source AST as data into fresh
modules. Whole test top-level code and candidate code are never invoked by it.
Setup guards and independent literal outcomes are required, not merely old==new.
Four decoy-collection probes are intentionally artificial branch tests, NOT
faithful production ownership fixtures. None have been runtime validated.

No product RED/GREEN cycle, full suite, independent review or actual speed
comparison was run within this narrowly approved offline allowance.

## Preserved new source SHA256

| File | SHA256 |
| --- | --- |
| .gitattributes | AC3C957075C3178B2FC9358435F8A01BD929D9ED5D7662226BD05D27FDB140D7 |
| capture.ps1 | 79D41F8AC6F1EF1A0FE316183F31A0C38C349256FC40D60846E38F5648BA577D |
| offline.tests.ps1 | 6FE3A856FAF7F46B507D42C0D3979FA15E6FDE485517BFD15D62B744D56FBA8C |
| owned-return.tests.ps1 | 8F41A9850474720F5AE2525E3E905018884DA2CB55A92E8C669F339965173E04 |
| README.md | 9C2BDB5F53B0CB1C7218BF4D1F4A58D3D2DCBA1571FD36BD0083F2FD812A17B8 |
| source-pins.json | 4D287D4024393FD9B916F44025A835045D20710A57DB10C7DC8C9AD94ECB4005 |

Candidate remains 1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8;
original instrumented test remains
986CDE7605B88500C35B70D5C4152B1D63370BAC3301B5BA514AC3B2C34AE5BA.
The previous 30-second timeout and all historical records remain unchanged.

## Next bounded proposal

Staging audit: nine explicit necessary files only. Checkout-filter bytes were
compared with every physical file: all nine are identical after CR removal.
Exact bytes match for controller, offline harness, README and source-pins JSON.
Git normalizes the attributes file to LF and the optimized test's final LF to
CRLF; their fresh-checkout SHA256 values (autocrlf=true) are respectively
9C8A6602667ED23ED583C803EB28CCC446EC285015FEEA235973A9B9E4B78132 and
0BCB3B95F2A215A8E143BAB05C70731A53D69F98CDFFE214393D8F3CBCEF9C73.
The three roadmap/recovery documents normalize to LF. Physical hashes above
describe the preserved preflight inputs, not a claim of fresh-checkout byte
identity for every file. No non-line-ending differences were found. Standard
diff whitespace check reports blank lines at EOF in six new files; they are
preserved with the failed preparation, not silently repaired after the stop.
The check excluding only blank-at-EOF warnings passes. No functional validation
is implied by these staging/content checks.

Preserve this failing controller. Prepare a separately named corrected controller
with the original inventory hash, using field-specific edits instead of global
numeric replacement. Check every new literal against its unchanged source before
launch. Obtain agreement to proceed after this mandated stop; do not treat this
checkpoint as permission to run either controller or the optimized whole test.
A future offline run remains one ordinary PS7 child, 30s + 2s owned exit + 2s drain,
no candidate import/call, no old-slot reuse, immediate stop on any inconsistency.
Only after equivalence passes should a separately approved bounded candidate
timing proposal be considered. No main/I1/security/runtime/C12 acceptance.

Fresh PC: use the development branch and this report. Portable source permits
inspection, but ignored historical pin receipts and engine paths are machine-local;
cloning does not recreate them or authorize execution.

