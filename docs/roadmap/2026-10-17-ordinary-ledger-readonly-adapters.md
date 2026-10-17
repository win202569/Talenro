# Real read-only adapter observations — diagnostic wrapper stopped

On development checkpoint `4b40f034815863cbf132360b734a8059d149bfda`,the user
authorized real read-only interface checks,not Stage B. Two fresh diagnostic
hosts ran sequentially with NoLogo/NoProfile/NonInteractive/EncodedCommand and
120s outer deadlines. Neither was a probe workload. No retries were performed.

The hosts imported the actual helper,which loads/compiles its existing C# type
definitions,then called only source/file/signature/binding query functions.
There were no fake OS functions. No planner,nonce generation,runner,process
workload method,Go/Docker operation or acceptance evaluator was invoked.
Compilation/module initialization may use normal runtime temporary/cache files;
this was not a promise of zero incidental OS writes or zero host processes.
No product source,system executable/link or security setting was modified.

## Host and interface evidence

| Diagnostic host | Query UTC | Child exit | Elapsed ms | Output disposition |
| --- | --- | --- | --- | --- |
| Core 7.6.5,x64 | 2026-10-17T11:51:11.2153467Z | 0 | 4912 | JSON,empty stderr |
| Desktop 5.1.26100.9711,x64 | 2026-10-17T11:51:16.6178048Z | 0 | 5033 | JSON plus CLIXML progress on stderr;wrapper stopped |

Both actual production read-only paths returned matching shell records:

- Policy:fixed-inbox-shell-binding-v1.
- Path:`C:\WINDOWS\System32\WindowsPowerShell\v1.0\powershell.exe`.
- SHA256:`60C9B29843624DD8AF6B6A7B26147753BE6AEEF65098B9A503CC50850AFE410F`.
- Length:495616 bytes.
- Held-handle identity:Volume3036474562,IndexHigh524288,IndexLow2141587,Links2.
- Authenticode Valid=True;subject
  `CN=Microsoft Windows, O=Microsoft Corporation, L=Redmond, S=Washington, C=US`.
- Signer certificate raw-byte SHA256:
  `E984867EEB6C36DF23AC5D84AF2C0CF8553D7E6BD5AF63F8CCB02DE97841F07A`.
- A second read binding was opened while the first remained held,and its whole
  record matched the expected first record;both disposed in finally.

Actual GetFileInformationByHandle Snapshot/Links and Authenticode adapters thus
returned usable values on both hosts in these sampled read-only calls. The
earlier single-link incompatibility did not reproduce in the new shell-specific
read-only binding path. This does not prove launch compatibility,process
containment,continuous mutation resistance,or all possible link configurations.
No write-access test was attempted against the system executable.

Both hosts ran Assert-OrdinaryLedgerProbeLoadedSources successfully and read all
11 source files through the real independent-file validator. Each had Links1;
names,hashes and lengths matched across the two captured outputs. Key updated
source hashes were:

- ordinary-ledger-native-probe.ps1:
  `B93AD786C47180EB22A0E6F7B2F339712E8D6B7FA1D1FA53BC256C5058BD59B3`.
- isolation.ps1:
  `1FD75426BBB1015664301108B1A2F5E3426874CB672EDFEE3FE7767BD1356E2D`.

All four actual ledger-native-root/child-001 run/control paths were absent in
both observations. No actual identity was created or consumed. No exact probe
plan/command hash or outer workload watchdog was prepared.

## Strict diagnostic failure and read-only diagnosis

The transient parent wrapper treats **any** stderr as failure,even with child
exit0. Desktop's captured stderr began `#< CLIXML` and contained four `Obj`
elements with `S="progress"`;their ProgressRecord type was Completed. Localized
activity text was encoding-damaged in the tool rendering. No child ErrorRecord
was visible in that payload. The parent then threw exactly:
`Diagnostic host failed or stderr nonempty;stop`,and its command exited1.

This demonstrates the failing layer:child metadata/signature queries returned
JSON/exit0;the parent's blanket stderr rule rejected progress serialization.
The distinction between Progress and Error is documented in Microsoft's
[output-stream reference](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_output_streams?view=powershell-7.6).
This reference supports distinct stream meanings,not a claim that every stderr
record is harmless. The transport behavior above is an observation of this run.

No output was silently filtered,no preference changed,no parser patched and
no second diagnostic run made. The overall strict diagnostic is **not** recorded
as clean/passed. A future wrapper must preserve and distinguish channels and
continue refusing Error/unknown/malformed/trailing data;that is not implemented
or authorized by this report. Do not suppress all stderr to obtain a pass.

Evidence source is captured tool output,not a byte-preserved raw-file archive;
no raw child stdout/stderr files were saved. This portable report records the
observed fields and parent failure,but cannot recreate original encoded bytes.

## Handoff and remaining gates

This was two read-only diagnostic-host executions,not two test-suite calls or
two native probe entries. Existing Stage A70/70 budgets remain exhausted;older
raw indices and the documented pending-entry anomaly were not modified.
No new full-suite result or security clearance is claimed. Defender state/events
were not refreshed here;the prior point-in-time status cannot clear historical
Defender findings. I2(all13 physical gates) and I3(historical validation cause)
remain OPEN acceptance blockers,as recorded in current status.

Next:prepare the remaining safety/evidence and exact-input decision checklist,
including the non-clean diagnostic channel disposition. Do not launch Stage B,
rerun consumed identities,change protections,merge main or activate acceptance.
Fresh-machine recovery must recheck host-specific file identities/signatures;
the recorded file index and signer/hash facts are not portable clearance.
