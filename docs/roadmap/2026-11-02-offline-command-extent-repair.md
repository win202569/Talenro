# Command extent local repair — executed 2026-11-03

User approved plan040b975b and two auxiliary runs. Native execution completed
RED16 and GREEN17 only; no additional probes, hosts, targets, compilations,
runtime snapshots or reviewers. Product/I1/I2 remains NOT_ACCEPTED.

## Actual results

Preparation1f2dc032 pushed before RED16. Actual RED0PASS/1FAIL/9NotRun/exit1:
SavedErrorAccepted failed with RuntimeException Message/FQID
OD: observation contract mismatch at unchanged OC-SourceError line33.
SetupError/RecordingError null.31source and76history pre/post hashes equal.
RED finalSHA44635E7FCE3CCFBCF176E4A02D707F79DC5BEB55BB713A761A269889699B33D1.

Only after expected RED, inserted one early return in Get-ODFailureExtent for
CommandAst with exact command name ConvertFrom-Json. It returns Node.Extent.
All other ancestor-walk logic and all other30operational sources unchanged.
Removing the insertion restores the original UTF8 bytes/SHA exactly, not merely
equivalent AST. Candidate parses; no hardcoded fixture coordinates in production.
Candidate12aa6b32 pushed before GREEN17.

GREEN10PASS/0FAIL/0NotRun/exit0, all errors null:
SavedErrorAccepted,CommandExtent,WrongColumnRejected,WrongLineRejected,
WrongOperationRejected,WrongFunctionRejected,WrongSourceRejected,
MethodsUnchanged,ThrowsUnchanged,ModelBoundaryUnchanged.
31source/76history pre/post hashes equal. All actual receipts/tool input/output
preserved; no replay, reclassification, copied baseline or patched test after RED.

This proves the saved row8 error is accepted with the command extent and the
specified local negative/model-invariance boundaries pass. It does not establish
that later rows or the complete contract/runtime chain will pass.

## Identities and recovery

- TestSHA75E2494CE8FC8DA2AE86AFC4EFFBA45CF8A5EE2DDD03155A60CFE35C92BDBFA5.
- InputSHAA96DB6DDDBCBEC6CCFDE3F5E1D5AAA5D95FE512F4144EEA64DCDCBDE345735A8.
- SourceSHA BB8FA693FC6CE09FE8CC54EA8F5FAD43CF483B015928C6631AA12C0218A1CA8B.
- GREEN finalSHA C25EC3F9A9041AC74BFD2B8D0CBE06D01650E6D3E222C87F0C3184E4E2DD0696.
- Preparation archive534files/135directories/31zero-byte, decoded equal,
  SHAAE05F62B7E81C4AA05591F71F578CDC0E6396A20D1679449327B99E8ECF86A98.
- Candidate archive538files/135directories/31zero-byte, decoded equal,
  SHA3DD3DC6D43B27A0C18E6599A8EDD040075609ADB9D2C2B622EE65C94259844C7.

Original repair manifest SHA remains
A9949DD46C7F928CFE7DB33F092AF0497E6F89556E8F359A8A55CB4484BB1556;
it intentionally still pins the OLD source helper. Do not rewrite/bypass its
checks or run the old driver. All nine old stops remain; no new stop was needed
because both runs matched expected outcomes. Core/lifecycle/continuation intents
remain absent. The old failed contract-green verdict is unchanged.

## Limits and next boundary

Final portable-command-extent.json contains541actual P/Q plus required H files,
135directories/31zero-byte entries; decoded bytes/lengths/hashes/directories
verified. SHA71861DA4FDBCA93A5D68AC255AE5B9FAFD709D9346219312A77DFA2345E364C1.
Embedded ledger records the moment before final archive creation; this report
records the subsequently verified preservation. Earlier archives untouched.

Consumed17auxiliary/4hosts/3compiles/186snapshots/0targets/0reviews. Auxiliary
allowance exhausted; original remaining3hosts/7targets/conditional1review frozen.
No new compile/snapshot budget. Further work needs a separately approved runtime
rebind/freeze and validation scope with new attempt identities, not old-slot reuse.

Ruling: used the approved PowerShell/P/Q two-run workflow rather than generic
Bash, extra tests/review or cleanup. This preserves bounded execution and raw
records; the tradeoff is that no whole-project suite or independent review ran.
The local regression is not a substitute for those deferred acceptance steps.
