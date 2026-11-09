# Read-only deployed-artifact comparison

After27f93336, inspected the preserved old and trace deployment trees without
launching, rebuilding, loading application code, editing binaries or changing
services/settings. SHA256 comparison recursively covered regular files, excluding
only trace-generated close-stage-records/ and process-observations/ directories.
This is a current disk snapshot, not a historical loaded-module trace.

Paths relative to clients/windows/Talenro.Demo.App/bin/:
- Old: layout-diagnostics/Debug/net10.0-windows10.0.19041.0/win-x64/.
- Trace: close-stage-recording/x64/Debug/net10.0-windows10.0.19041.0/win-x64/.

## Results

Both trees have527 deployed files with identical relative-path sets.522 have
identical SHA256. Exactly five differ:

| File | Old SHA256 | Trace SHA256 |
| --- | --- | --- |
| Talenro.Demo.App.exe | 4180C9016386080AE442AEE091E9CBED4C638CF5E96DBD36316EC85E62CE3B95 | EC9E4375948F2E7AA36B34DF32F36A5374F68F61C6A1A4FE42A8108CD0255774 |
| Talenro.Demo.App.dll | 7F0A24D21A11084A57DEBD089BA2D5D258ADB923BC592A241D5997F4B77250BB | 7E14F4A771DB0F7024C95E95A4193A16448D7B444C8D11273717E8B15F46611B |
| Talenro.Demo.App.pdb | DF6BEE4476959FCED9DC0DC571FF035B5AECFD0F06DA0C0B922332BA7CEBF1EC | 829E316F48C2D4D480EADFD2F86CC9F4FD65E8C099BF40F1E6CC3BD6F4991925 |
| Talenro.Demo.Core.dll | 5794EA105B605B76A274FBC7C3348CBE5AAFC071C9DBD082D3F75F1689EE2746 | D801F9EE75D4E8EE7F3AF79A8F7667EDD73C18BABEBDBFC90DD80575C90B9F02 |
| Talenro.Demo.Core.pdb | 4A2C47BD930DC2A8FD8F4C5A18A395D38AEE51D5C8D85FE471F8F8C76EB7C5FA | E8638CA443FBF7C0338E18E53C9438C90E8F0FF274A5EDE125CFC4EF5D2F35C6 |

Thus all other deployed files, including dependencies, XBF/PRI resources,
deps.json and runtimeconfig.json, match byte-for-byte. Both runtime configs name
Microsoft.NETCore.App10.0.14; runtime target .NETCoreApp,Version=v10.0/win-x64 and
deps library-key sets match. This does not establish which external/system
components were loaded during historical runs or their state/version at the time.

## Core difference: static inspection only

PEReader inspected both Core DLLs without executing them. Each has377 method
entries; token/name/signature-blob keyed raw IL comparisons found377 identical,
zero changed/added entries. This is not a complete semantic equivalence proof:
other metadata, token targets, attributes, resources and native behavior were not
exhaustively compared. Core source was already unchanged in cf0f605d..e347d8cd.

Old Core PE Machine=I386, CorFlags=ILOnly; trace Core Machine=Amd64, CorFlags=ILOnly.
The old header does NOT by itself mean the running application was32-bit; no
Requires32Bit flag was reported. MVIDs differ (old91008d68-5848-4045-8df6-f94c5d878643,
trace4cd30eb9-5d2b-4d7f-9a2f-5f68acfc54ce). New builds explicitly used Platform=x64.
Platform/build identity is therefore another concrete difference; do not describe
the artifacts as differing only in recorder markers or infer causation from it.

## Remaining differences and next decision

- Recorder startup creates/open files, initializes a queue and background writer,
  waits up to2s for startup readiness; writer polls every10ms. Close path includes
  marker publication and prebound scope calls. These may perturb timing even
  without UI-close disk waits. No measured overhead or masking conclusion.
- Historical observer used about5s samples plus bounded outer process waits;
  trace runs use about2s sampling. Modules match source, session procedure differs.
- Program location, platform build identity, instrumentation and historical
  OS/service/IME/focus state still differ. Long-country and Alt+F4 were aligned
  in the latest trace run, but all remaining variables were not eliminated.

Evidence does not support a dependency-upgrade explanation within the compared
deployed files, nor justify changing shutdown logic or disabling services.
Before any new instrumented variants, a useful candidate is one separately
approved run of the PRESERVED old binary with the latest observer/input/long-country/
Alt+F4 procedure. That tests whether the old artifact still exhibits the symptom
under the present procedure/state. It cannot isolate recorder effect from build
identity and has no stage markers. Success would weaken a current deterministic
old-artifact explanation, not prove repair; failure would motivate tighter artifact
controls, not immediately identify cause. No new run is authorized by this report.

No speculative rebuild, runtime budget extension, binary replacement or main merge.
Preserve the original files. Compare this with the
[historical matrix](2026-11-09-windows-close-stage-history-comparison.md) and
[three recorded routes](2026-11-09-windows-close-stage-altf4-observation.md).
