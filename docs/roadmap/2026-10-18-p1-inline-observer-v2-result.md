# Observer storage v2 — offline equivalence 45/45

Date: 2026-10-18. Base b53ab5dfa422a4c6e25655be071fef513d2768c9.
Branch: codex/c12-b01-task9-coordinator.

## Verified result

The user continued the proposed single bounded v2 offline validation.
All45paired cases passed in1099ms, exit0, no timeout, stderr0, no controller
error. Owned child exited and both streams completed; all657frozen input pins
remained intact, including an independent read-only rehash after execution.
CandidateImports/ProductCalls0; FormalAcceptance false. No retry, post-launch
repair, candidate/whole-test import, actual ownership suite or old-slot reuse.

One ordinary-user PS7 child;30,000ms execution +2,000ms owned exit +2,000ms stream
drain. Allowance consumed before launch. StartUTC2026-10-18T13:31:29.6310977Z,
ownedPID14360. The unchanged fixed engine and source pin chain were checked first.
Eleven fixed controller hash literals matched actual files; parser errors0.

Saved stdout independently matched exactly45ordered paired PASS lines to the
fixed fixture inventory, followed by one exact RESULT line (46lines total).
Inventory SHA matches the original45case definition inventory unchanged.
Each pair uses separate fresh fixtures for old/new observers and both before/after
phases; outcomes must also match independently specified literal expectations.
Observers must emit zero pipeline objects. This is not merely old==new.

## Coverage and meaning

| Group | Cases | Verified outcome |
| --- | --- | --- |
| Five kinds x clean / selected-only decoration | 10 | Both phases PASS |
| Five kinds x descriptor / storage / base / expected-type change | 20 | Before PASS, after field-base-changed |
| Five kinds x held metadata contamination | 5 | Before PASS, after cross-owner gap |
| Four isolated fresh-reread branches | 4 | Before PASS, after cross-owner gap |
| Two deliberate-sharing controls | 2 | Before PASS, after cross-owner gap |
| Same-owner different-field clean/dirty | 2 | Clean PASS, dirty gap |
| Late dirty leaf / empty list | 2 | Late gap, empty PASS |
| Total | 45 | All paired outcomes match literals |

Five kinds: Int32, Int64, Boolean, raw String, pre-stored PSObject String.
The four fresh-reread probes intentionally substitute a separate real clean
Members collection in the captured context. They are synthetic branch-isolation
probes, not faithful product ownership fixtures. Sharing controls cover Int32 and
pre-stored PSObject String only. No exhaustive proof beyond the stated vectors.

This verifies the original and inlined V-Others implementations' observation
outcomes for this synthetic fixture matrix on the fixed engine. It does NOT
establish candidate correctness, product-mutant sensitivity, timing speedup,
completion of the actual39case ownership matrix, all runtime contexts,
security/runtime acceptance, I1 closure or C12 completion. No full-project suite,
independent review, actual ownership test or product mutation ran in this scope.

## Source preservation

The previously prepared v2 harness, setter helper, original/optimized tests,
candidate, original45case inventory and all old inputs/reports remain unchanged.
Only capture.ps1 and EXECUTION.md were added to the v2 leaf before freezing.
EXECUTION.md explains that README/source-pins reflect the earlier preparation
checkpoint, when no controller existed; the new manifest records this allowance.

647prior frozen inputs +4 prior raw receipt anchors +6v2files =657pins.
The harness still parses source as data, extracting ten allowlisted observer/
helper definitions and exactly the one validated test-only S-SetStored definition
into fresh modules. Whole original/optimized source equality outside V-Others
passed inside the completed batch, after newline normalization.
No source/helper implementation was edited after launch.

Historical failures are retained: original controller preflight hash error,
old harness case4 storage-setup failure, prior actual ownership timeouts, the
successful14record setup trace and10setter transitions. They were not rerun.

## Preserved evidence

Ignored local packet: .superpowers/sdd/2026-10-18-p1-inline-observer-storage-v2.

| File | SHA256 |
| --- | --- |
| frozen.json | 8A5FB6E48FC16867EA19FC9DC151AF7E4E4A303A4507BAF8BAE7F349D2427276 |
| invocation.json | 60B883B6AB22480F7B54ACEAEE5CB23BA7B3A623139E77C196D9598C258BDEBE |
| stdout.txt | AE3DAB2BC89BD7602A70DC25677A6F54ECB22FFC41B8A46B7A82CA9BDF1D4C3D |
| stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| fixture-inventory.json | A055241CF41A2D0D19C6A3DC7D1A1C8E00D0BE6E46F8D3A5D0168886137382E3 |
| capture.ps1 | 13F0F255059EB75F9DF5C155FD433626EE11420C3A8D1DE2F8E45B392A0084BA |
| EXECUTION.md | 7BA086750490EB3EF35F3032AD4E3657042A73A6C3DCA1E47F85BD5085C0F092 |

Unchanged v2 harness:
D6531D884313F22DBBC3A03F5D6909A7D8E58C51C0B216FF011663EE68DCFB33.
Setter91C107BF6D5A08D265F0837CB161D75C5C08FC28DB962745EC13DABEBE1A5B01.
Candidate1971F808DAB0C0A520FC708BF240AF4367D21EE811FE9AEB8664E605E7E8BCE8.
Optimized test physical8F41A9850474720F5AE2525E3E905018884DA2CB55A92E8C669F339965173E04;
as previously documented, Git normalizes its last LF to CRLF on a fresh checkout.
The original45case literal source and this report's complete group/outcome table
preserve portable progress. Raw receipts remain ignored and machine-local; cloning
does not recreate them, the engine cache or execution authority.

## Next proposal, not executed here

Now that scoped offline equivalence passed, propose one separately bounded
candidate stage-timing diagnostic using the optimized test copy. Preserve all
39cases,80decision calls,29direct scalar calls,1metadata call, both scan phases,
all history, literal expectations and strict saved-text checker. Keep30s execution
+2s owned exit +2s stream drain, ordinary token, all historical pins, consumed-before-
launch accounting and no retry/repair/later suites. No timeout increase or
assertion/history removal. Do not promise that the actual matrix will fit30s.

Require agreement to that candidate-executing scope before creating/launching
its controller. Use the unchanged strict timing checker and original inventory;
report either full matrix or validated prefix with limits. This completed offline
allowance authorizes no candidate run, main merge or wider acceptance.
