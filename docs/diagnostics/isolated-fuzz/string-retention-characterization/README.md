# Four-case retention characterization — diagnostic only

Approved design:current-status.md at4b88ae0ab7721521188ca0aab1b7c08a52de7c58.
No product/candidate imports,helpers or decision calls. This diagnostic is not a
regression pass or I1 closure. Candidate1971F808 and every old source remain frozen.

Cases are held String-view control,raw Int32 field control,raw String field and
pre-stored explicit-wrapper String field. Every field is initialized before
measurement and never reassigned after attachment. Held view/member collection,
field descriptor and base references are retained. Raw String loss is a measured
capability,not a product failure or successful ownership isolation.

Identity comparisons use fixed public ReferenceEquals MethodInfo with object[]
arguments to avoid ordinary method-argument debasing of PSObject values. Public
GetType uses a fixed public MethodBase.Invoke trampoline for the same reason.
Two wrappers around one base calibrate same/different wrapper versus base identity
inside the held-view control. No private reflection,compiler or injected SDK proxy.
If that instrumentation is invalid,the sole attempt stops rather than guessing.

The six-line complete output is HOST,four FACT JSON records and terminal summary.
output-schema.json fixes order,exact fields/types,allowed measured capabilities,
nullable absence markers and control invariants before execution. Unknown booleans
are observations,not post-run expected values. Only facts are serialized.

capture.ps1 permits one ordinary fresh PS7 host,30s+2s owned exit+2s stream drain,
after old/new integrity checks and persisted consumption. No warmup or retry.
Code is preserved only as non-production diagnostic/recovery evidence;it is not
an approved reusable acceptance gate. Consult latest roadmap results before any
operation;Git/missing local raw state cannot renew consumed or frozen authority.
