# Scalar receiver diagnostic — complete within the single-host scope

Date 2026-10-18. Base 325afd1e1abc4b8b0f8bbf71a885ed41e72efba2.
User approved one independent ordinary pinned PS7 host, <=60000ms, to compare
the old observation path and a fixed actual-wrapper handle. No product changes
or product API/import, compiled fixture, full regression or reviewer authorized.

## Result

DIAGNOSTIC_CAPTURED. PID48656,6333ms,exit0;no timeout/stderr;owned exit and
separate concurrent stream drains complete. 1121 mapped pins intact after
capture and read-only checks. New host allowance 1/1 SPENT;no retry.
All previous frozen scopes unchanged: latest repair1/3spent+2FROZEN,
review0/1ineligible. No old result reclassified or raw/receipt rewritten.

All6 cases (string/byte/byte-array x variable/exact note field) completed,
each with3 repeated observations:18 fixed-path checks and12 independent
key/table controls passed.130 actual-reference IDs recorded. IDs are scoped
to this process and do not restore live CLR references on clone.

| Kind | Old separate reads (both source forms,3repeats each) | Fixed handle |
| --- | --- | --- |
| string | different wrapper,key,table on each raw-versus-shim comparison | stable wrapper,key,table across reads |
| byte | distinct descriptor-owner wrappers,but same base key/table | stable wrapper,key,table across reads |
| byte-array | distinct descriptor-owner wrappers,but same base key/table | stable wrapper,key,table across reads |

String variable first repeat: old psextended owner/key ID3,table5; separately
acquired psobject owner/key ID7,table8. Raw retained only psextended;returned
raw ID matched original ID5. String note-field first repeat: owner/key34,
table36 versus owner/key38,table39. All6 string legacy observations disagree.
This controlled run directly observes the helper-boundary mismatch pattern;
it does NOT reconstruct the historical PID45752 instance.

All12 byte/array legacy observations share key/table despite different
descriptor-owner wrappers. First raw inventory has only psextended;later
inventories also include psobject. All6 string raw inventories contain only
psextended. Therefore taking raw inventory before creating the other reserved
view is another reason not to require both entries in that earlier snapshot.

## What changed in the diagnostic, not the product

A containing object[] retains input,actual wrapper and base. An explicit SDK
AsPSObject boundary obtains a wrapper once;later helpers receive the containing
array,not a newly adapted scalar. Fixed SDK factories materialize psobject/
psextended on that wrapper for fixture construction. Fixed-field snapshots
record wrapper/base/key/table and descriptor-owner references before judgment.
The same wrapper ID and table persist across repeated reads in all18 phases.
All36 after-generation descriptor observations have hidden/reserved/instance
flags, owner-key/table and own internal/constructor-owner consistency verified.

Independent controls deliberately obtain another wrapper:
- Same-base distinct string wrapper: different key/table.
- Same-base distinct byte/array wrapper: shared key/table.
- Independently constructed equal-content bases: different key/table for all.
All12 controls match those literal expected relationships,so content equality
cannot replace reference identity. Numeric sharing must not be mistaken for
a rule that every descriptor wrapper owner must be identical.

Important boundary: InputWasWrapper=false and CapturedInputRetained=false in
ALL6 cases. The explicit capture created a wrapper over the original base;
CapturedBaseRetained=true. This run did NOT exercise retaining an already
wrapped input, nor prove that hostile preexisting metadata survives capture.
Do not describe this as preserving arbitrary original wrapper provenance.

## Limits and next step

This resolves a controlled observation-boundary question,not M1 or product
validation. All product source stays B468B564;failed18DA24E1 unchanged.
No scalar/array public API,output-copy guard,hidden refusal matrix,full suite,
optional5-member history,local-storage mode or remaining numeric/bool types
were verified here. Factories/key method are test-only;not a product whitelist,
unforgeable SDK provenance,or a same-process reflection security boundary.
No formalP2-B/PS5/P2-C/P3/fullC12/live/security/identity/capability/runtime grant.

Next prepare a bounded test integration design: carry actual receivers without
re-adapting scalars,record operands before assertions,materialize reserved views
before inventory when that is the intended fixture history,and keep original
input/decoration retention separate from generated clean-wrapper proofs. Cover
already-wrapped/decorated inputs explicitly before relying on capture for
validation. Then establish the exact baseline defect gate before product repair.
This diagnostic neither unfreezes old slots nor grants further target runs.

## Evidence and portability

New leaf docs/diagnostics/isolated-fuzz/p2b-scalar-receiver-diagnostic contains
four immutable inputs and a postcapture derived portable-summary.json.
Ignored packet .superpowers/sdd/2026-10-18-p2b-scalar-receiver-diagnostic.
Portable summary records130 reference IDs,6cases/18phases/12controls and column
legends. Null means absent/not applicable,not success. It omits full call and
snapshot duplication,not stored in or claimed as the frozen input manifest.

| Artifact | SHA256 |
| --- | --- |
| frozen.json | 9E885991FAB1F6BFA71922893A080C38B8558CCAA3FA5CFF1C4E81B27BAB7ABD |
| receipt.json | CDE1BCF536D18498A3B4A3F09C178C1A5BC99C8DC1034DC7EB888C253F51AA43 |
| stdout.bin (229841bytes) | 9AB40CD97F648EEA2CB32AA2C9EFB944A968DF80131D938061E01A66BD9F3461 |
| stderr.bin (empty) | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |
| diagnostic.ps1 | 49A87509A138D91DC4CA9FBA5AE8B12E99C46B8888D0507196BD032DA3BDE3A1 |
| capture.ps1 | 45D8C9D01F23040296DCE650CACA2156D45A44C7F6B159D46F3290710ACFBBA3 |
| portable-summary.json (derived) | AF63C11BF00041E45F0494A708FF70C9BF16D4C834EAFD3A54091C74DF4D9AC8 |
| unchanged product | B468B56426A084E472E97F651E23BE83C985014E3B2A8972785DD0D10A76D039 |

1121pins=inherited1109+priorpacket5+priorreport/summary2+newinputs4+livebaseline1.
Pinned PS7 engine362A356C and SDK2D8AF026 unchanged. Execution/owned-exit/drain
bounds60000/2000/2000ms;ordinary parent/target;only owned target terminable.
No install,network/security setting change,product import/API/compile/edit,
extra target,reviewer or main merge. Static parse passed before launch.

Necessary portable sources/summary/report/status/recovery saved to existing
development branch. Fresh-computer recovery starts there,not automatically
main;clone excludes ignored raw,actual references,SDK/runtime/cache,sessions
and mappings. Neither clone nor this successful diagnostic resets budgets.
