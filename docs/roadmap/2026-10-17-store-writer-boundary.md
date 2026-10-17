# Store writer inherited-admission boundary coverage

## Approved bounded scope

User approved the short test-only design after
`6cb7851b07b9b124a3ce2a1964237e8610cc742c`:extend the existing store suite for
non-exact labels/fields,executable properties,fake types,length shadows and
normal Unicode/raw-byte persistence. Two NEW hosts only,PS7 thenPS5,120000ms
each,stop on failure,no retry,no production repair. Tests and preservation docs
only;no store format change,real collection,security change or main merge.

Existing54 assertions remain unchanged. Added26 boundary cases and one terminal
completion check,plus explicit host banners. This is characterization coverage
of existing behavior,not a new implementation fix or newly demonstrated
RED/GREEN cycle. No intentionally weakened validator,forced mutant run or extra
baseline/full-project test hosts. All prior/frozen allowances remain unchanged.

## Observable coverage

- Eight label/field mutations:policy and classification embedded NUL/soft hyphen,
  encoding NUL/soft hyphen/case,and malformed Policy field name.
- Three wrong-container GetType spoofing cases:package/stdout/stderr.
- Three positive unused GetType method cases:publication succeeds without
  invoking the method and retains literal stdout/stderr bytes.
- Three executable object properties and two executable scalar properties:
  actual callback counters stay unchanged and no destination is created.
- Four bound-shadow cases:1025-character origin with Length1,Base64 Length,
  SHA256 Length64,and Base64 Rank1;all rejected before directory creation.
- One object-valued encoding with a ToString callback:refused without coercion.
- Two Unicode origin/raw-byte roundtrips under utf-8/utf-16le declarations:
  literal abc and00FF0D0A snapshots,exit-1 and refused classification retained,
  exact composed/decomposed/CJK origin retained without normalization,and all
  package/store origin/acceptance/runtime authority flags remainfalse.

The refusal helper calls the real public writer with a fresh unique direct-child
destination,requires the known package-preflight refusal,checks destination
absence and unchanged callback count. This excludes incidental later I/O failure
as a passing rejection. No mocked validator or source-text behavior assertions.
Scalar test annotations are removed in finally blocks to avoid fixture residue.
Positive tests follow negative cases in the same host. Expectations for byte
snapshots are independent literal Base64 values,not computed by store helpers.

## Verification —2/2 consumed

|Host|Observed banner|Exit|Elapsed ms|PASS|FAIL|Stderr characters|
|---|---|---|---|---|---|---|
|PS7|7.6.5Core x64|0|10782|81|0|0|
|PS5|5.1.26100.9711Desktop x64|0|13905|81|0|0|

All162 assertions passed in the two complete expanded store suites;no retry,
timeout or code/test change between/after runs. Static parser check before
launch found0 errors;it is not counted as dynamic verification. Existing
hardlink/junction/exclusive-writer refusal tests also passed. Existing
[three-consumer regression](2026-10-17-evidence-dependent-regression.md) remains
separate evidence;consumer production sources have not changed since it ran.

SHA256 rechecked after both hosts:

- Store tests:`80DCB954361F6ACE71FC9649D25A35748ACB3D578491A30C3D3F9862ACBDC147`.
- Store source:`7710A1149B42CFD0C2505988CEE6748F399C3FB5E8C34D795B35FBCB8558BB81`.
- Package:`857543A9DDA9B9C849B2507C6EE419C1F73AA0D0C8599CD9D88738841D021198`.
- Classifier:`F323C9C46190DE06F25BFA55A2B7983707B1FF5877548F9703870F9BB771A0BF`.
- Two-entry invocation index:`8734CC038681AF35F12F1519BEACE7CAB3EB2723216A0CAC97D6267876C85B1E`.

Ignored local index/runner/text stdout/stderr receipts retained under
`.superpowers/sdd/2026-10-17-store-writer-boundary`;synthetic store files/links
remain in the dedicated ignored fixture root. No artifact deletion. Text host
receipts are not live raw capture/provenance evidence. Portable tracked tests
require no private cache/session data;new computers prepare their own hosts.
Consumed2/2 is not automatic replay authority.

## Conclusion and remaining boundaries

No production correction needed for these tested writer-admission paths.
No store binary layout/schema/interface migration. No new independent review
was included in the approved packet;do not expand previous review verdicts to
this test change. Author verification only,not merge/full-project certification.
These selected fixtures do not exhaust CodeProperty,arbitrary prebinding
coercion,internal helper misuse,concurrent/process-wide mutation or path races.
They do not make descriptor validation a sandbox or authenticate declarations.

Package-only persistence still omits the full synthetic capture manifest,
arguments,ordered events,binding and lifecycle/control facts. No atomicity,
fsync/power-loss promise,live collector/watchdog/OS cleanup guarantee,command/
clock/origin provenance,security/vendor clearance,Stage B,I2/I3/full C12 or main
acceptance follows. Existing safety blockers and raw failed histories remain.

Next safe proposal:read-only inventory of full capture-record persistence fields
and integrity/recovery requirements,then present a separately approved design.
Do not infer permission to rewrite storage or start real capture from passing
package-only writer tests.
