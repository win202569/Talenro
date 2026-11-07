# Exporter documented help spelling — failed

User explicitly approved one original-privilege `-h` query,maximum30s,after
581e2daf. No recording,settings changes,installation,elevation or automatic retry.
This new finite attempt is consumed. Observation remains BLOCKED;
44/34/6/217,NOT_ACCEPTED unchanged.

## Observed result

Tool76ea41 invoked pinned wpaexporter.exe with exactly one argument `-h` through
the unchanged bounded owner. Exporter version11.7.395.48728 and SHA256
FC603F0316027C024132C9571A02959AD90C89CBC246B49F08958A74E6CEED62 were claimed
before launch. Owner/common hashes checked;fresh exclusive receipt root used.

- Started=true,ExitConfirmed=true,ExitCode=-532462766,ElapsedMs=6388.5249.
- TimedOut=false,Failure=null,EndReason=EXITED. Owner completed but exporter failed.
- stdout0 bytes;stderr1886 bytes. All four module files unchanged.
- Raw stderr read4eb29d:missing deps.json diagnostic followed by the same nested
  WpaApplication/WpaEvents initialization failure and registry SecurityException.
  The stack again reaches KitsCeipOptedIn via RegistryKey.OpenSubKey.
- stderr SHA256 F448C931E495E3A45AC5663F69C4578DCDF067E9FB39DBF5728A5EDFAFE4BDBC
  exactly equals the earlier `-help` stderr. No usable help output returned.

Changing the help spelling alone did not resolve this startup failure in the
current execution context. This does not establish behavior in another context,
identify the exact registry key/policy,or prove that missing deps.json caused it.
No `/?` attempt,privilege change,registry inspection or exporter repair followed.

## Evidence and handoff

portable-c12-exporter-help-h.json preserves invocation/packaging recipes,claim,
original process/stage receipts and raw streams:8 entries,7641 decoded bytes,
every length/SHA checked against originals byaf7e41. Archive SHA256
C317CB50333D9E3DCF92C23DACEEF0F2FB1F31FCCC4653283A9E9FD2F7838A50.
Known local workspace/tool paths are retained;no credentials/private trace payload
observed. No binaries,caches or ETL included;previous evidence left unchanged.

Stop this automatic observation path here. Further work would require a separately
agreed environment-diagnostic scope or choosing another project task;do not repeat
help spelling changes,relax security,repair the installation or start capture as
an implicit continuation. Exporter availability remains one of several blockers:
collector ownership,save-time storage bounds and privacy/runtime approval are still
unresolved even if exporter startup is later fixed. Original native timeout cause
remains unknown. Author review only;no main merge or acceptance promotion.
