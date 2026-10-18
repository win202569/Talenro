# Fresh original PS5 complete entry — timeout, NOT PASS

Following checkpoint5a6270cf and the user's approval to follow its recommendation,
one NEW independent ordinary PS5 host ran the unchanged original complete entry.
This was a sole1/1 allowance, cap30000ms, no warmup, retry, diagnostic copy,
product/test edit, reviewer or use of old/frozen slots. The allowance is spent.

## Result

- Started UTC:2026-10-18T08:39:52.5834652Z (local date2026-10-18).
- PS5:5.1.26100.9711 Desktop x64; parent/child admin=False.
- Timeout:true; actual wall30070ms, including70ms termination/collection overhead.
- Cumulative child process CPU22984.375ms; exit-1; owned-host-exited.
- Error stream empty;65 ordered output lines,48 PASS markers.
- Last PASS:P1PassiveDescriptors; next expected PASS:P1ArrayAndIntegerBounds.
- CaptureValid=true (valid prefix only), ExpectedResult=false,
  FormalProjectAcceptance=false. Controller intentionally reported STOP after
  saving the incomplete result. No child assertion error was captured.

Order remained shape6,reflection3,fixture8,original35,ownership10. The completed
prefix contains shape6+reflection3+fixture8+first31 original groups. The remaining
original groups, original LIMIT, ownership10 and final wrappers are absent.
Expected complete output is83 lines/64 PASS (62 groups plus two wrapper markers).
Last PASS does not identify the exact instruction active at termination; there
were no internal timing markers. Do not attribute the whole30s to one call.

## Read-only verification and interpretation

Independently matched every saved line to the frozen original output inventory,
checked48 PASS/65 lines and next expected marker, both raw hashes, receipt fields,
controller/engine hashes and all270 frozen pins. Output is byte-identical to the
historical original PS5 timeout prefix. The source/entry/tests were unchanged.
Owned-handle termination only, ordinary token enforcement, separate UTF-8
streams and the existing bounded wait/drain checks were retained.

This fresh original-entry failure means the [successful reverse B/A diagnostics](2026-10-18-p1-ps5-reverse-pair.md)
do not establish original PS5 acceptance. CPU differs from the historical run;
neither that difference nor identical output uniquely identifies a cause.
No claim of stable performance, product repair or instrumentation-only cause.
Both historical and current failed verdicts remain intact.

Original candidate SHA:
`953CC2A8AED3A0EC392683D91175262A33F4841EF45F1D12B8134CEB29E622CA`.
Original entry SHA:
`12D6F5460F97B9E09453B0387C30E3D20CE19B2EDF1CDF4DBF4E77DA31561422`.
Engine SHA:
`60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`.

## Provenance and handoff

Ignored local packet `.superpowers/sdd/2026-10-18-p1-original-ps5-fresh/`
retains scope/controller/frozen/index/raw, not committed or portable prerequisites.

| Artifact | SHA256 |
|---|---|
| frozen.json | D3F1E7B3EE96E9E4BA5FF8E259FE7F8452FC943FB8249537BE32F7132BDE9494 |
| run.ps1 | 65671959024DB482C99E5E21815538832017DA334BD9AFBD3ECE478C47FF36EE |
| invocations.json | F4A0C82C8338FD4463723CE6FDDFAAB9A105309E29E08C2FE4B31AE59D53714F |
| PS5.stdout.txt | CC6DA49921E2CA5A07B5A2DC1752EBAC3C058950834FD85F5990B23B71E44DB6 |
| PS5.stderr.txt | E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855 |

Pins inherit262 from the paired diagnostic, add its seven immutable packet files
and the original expected-output manifest. Prior records were not rewritten.
All old budgets/frozen slots remain unchanged, including reverse pair2/2 spent;
review0/1 unused/ineligible. Candidate remains UNACCEPTED; PS5/I1/P1 and all
main/C12/project/live/runtime/security/release gates remain unresolved.

Next recommendation: a ZERO-process read-only comparison of original versus
successful diagnostic product-loading paths, scope and invocation differences,
using existing source/results only. Establish a specific hypothesis before
requesting any further bounded experiment. Do not repeat unchanged full runs,
raise the timeout, remove checks or infer a fix from the diagnostic copies.

Recover this report and existing committed source/tests on
`codex/c12-b01-task9-coordinator` along with current status and recovery runbook.
Cloning does not restore ignored raw, cached runtimes or authority for new runs.
