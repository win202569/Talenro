# Exporter startup — read-only analysis

Source5560b94f;analysis only,not a fix. No exporter/WPA/WPR/Go relaunch,
registry reads or changes,elevation,installation,trace capture or ETL processing.
Systematic debugging used to distinguish the observed fatal path from hypotheses.
44/34/6/217,NOT_ACCEPTED unchanged.

## Evidence checked

Read-only8ecda0 verified all109 entry lengths/SHA256 in the committed preflight
archive;outer SHA remains A5E81FC3D7F97E1BA0A2B557417F225998DFA536BC6201E93DB4452CA93EBF64.
Decoded ExporterHelp stderr contains the missing deps.json message followed by
managed type-initializer failures and the inner registry SecurityException.
stdout is empty. Previous exit-532462766 remains a failed help invocation.

Local inspection5f18f0/18a374/ce961c found these installed files under
C:/Program Files (x86)/Windows Kits/10/Windows Performance Toolkit:

| File | Version / finding | SHA256 |
| --- | --- | --- |
| wpaexporter.exe |11.7.395.48728;unchanged from failed launch | FC603F0316027C024132C9571A02959AD90C89CBC246B49F08958A74E6CEED62 |
| wpaexporter.dll |11.7.395.48728;present | 65BFE55CEB823C43C129E1FC44DFF8522A558EF0722014EE54259067A747DA9A |
| wpaexporter.dll.config |146 bytes;declares .NET Framework4.5.2 startup | 18331AE06E5B3823DE5A7D219CDFE8A427951975953C8C0D5116AC621B62D7A5 |

coreclr/clrjit/hostfxr/hostpolicy and System.Private.CoreLib are present with6.0
version metadata. Shell DLL is11.7.395.48728. Its installed config contains UI/table
preferences;no documented switch bypassing startup registry access was found in
the inspected configs. The adjacent addins list names Shell/NT/NT.Shell components.
File presence/config text does not establish the exact loader configuration or
installation integrity. Do not infer required repair from mixed config/version labels.
No matching deps.json/runtimeconfig file was returned by the targeted directory
search;the previously reported exact deps path was independently absent40b9c0.

## What follows, and what does not

The fatal exception chain names RegistryKey.OpenSubKey and
WpaEventSource.GetValue/KitsCeipOptedIn during WpaEvents/WpaApplication initialization.
This identifies the immediate observed failure boundary:startup registry access,
not parsing an ETL or running the Go verifier. The stack does not expose the exact
key,access policy origin or whether another execution context would succeed.
Do not claim a specific registry ACL,administrator requirement or policy defect.

The missing deps.json message precedes that failure,while managed application
initialization was reached. It is therefore insufficient to establish a fatal
dependency-resolution failure or a causal link to the registry exception. Equally,
the available evidence does not prove the missing file is harmless. No repair,
synthetic deps file,runtime substitution or permission change is justified here.

## Documented export route and command caveat

Microsoft's [WPAExporter explanation](https://devblogs.microsoft.com/oldnewthing/20201209-00/?p=104530)
describes supplying an ETL and a saved WPA view profile to obtain CSV. It lists
help forms `-h` and `/?`,and notes option-category ordering. The historical failed
invocation used `-help`;do not describe that spelling as verified supported syntax.
The article is from2020,not a compatibility guarantee for installed11.7.395.48728.
Changing help spelling is not a demonstrated fix for the earlier initializer failure.
No alternative spelling was tried in this turn.

The documented route establishes that profile-based CSV export exists,not that
this installation can initialize offline or that an appropriate file-event view,
PID/creation-time binding,timestamp mapping and loss indicators are already verified.
No suitable working export pipeline has been demonstrated. A fresh export profile
would still need its own schema/column validation and privacy review.

## Decision and remaining authority

Read-only investigation complete;runtime readiness remains BLOCKED. Do not repeat
the same file checks indefinitely or start collecting traces without an analysis path.
The original native300s timeout is a separate unresolved issue.

Any next executable experiment is a new scope requiring explicit approval;the
prior stopped help budget is not permission to retry. A future proposal should
first decide how to establish a supported exporter startup path without weakening
security,then specify one bounded experiment and its interpretation. No automatic
elevated launch,registry inspection,install/repair or recording is authorized by
this report. If that scope cannot be approved,retain the blocker and pause this
observation route rather than presenting more sample checks as root-cause progress.

Even exporter startup success would leave collector ownership,save-time storage,
privacy approval and cleanup bounds unresolved. Original evidence remains intact;
no source/test behavior changed. Author read-only review only.
