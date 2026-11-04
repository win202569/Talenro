# Versioned preparation correction approval

User replied 是 after stopped checkpoint6e5c5f07 to retaining original drafts,
creating a corrected version and continuing within the unused single allowance.
Original prepare/common/parent/stop/archive remain byte-identical.

prepare-v2 fixes the malformed literal in WrapperPipelineUnchanged. Read-only
AST inspection also established that Finally.Extent is the brace-delimited
StatementBlock, not text including the finally keyword; v2 uses that exact
original AST extent and leaves the keyword in place. No diagnostic semantics,
limits, guard checks or input verification are relaxed.

parent-v2 authenticates original stop87702A83F5ECD81F87EE485DBBDDDE0BF93E5200283385D87A598A5A7CC1571D,
refuses new v2 stops and uses a distinct execution-v2.stop.json. Common helper
unchanged. Original drafts/stop/tool record and this appendix become pinned history.
No intent exists; counters27/6/3/217, one computation28/at most host7 remain.
First new unexpected check/execution result stops, no automatic repair or retry.
