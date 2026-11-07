# Non-administrator comparison — not performed

User approved a normal non-administrator PowerShell comparison after36a9967f:
open only HKCU\Software\Microsoft\Windows Kits\CEIP read-only,no value reads or
permission changes. The approved outside-restriction tool call6c56d5 first checked
the effective Windows principal's Administrator role. It returned true,so the
guard stopped before either registry-open call:

`RESULT=STOP_ADMINISTRATOR_CONTEXT; REGISTRY_NOT_ACCESSED`

This is an environment precondition failure,not a new registry access result.
Do not infer that ordinary users can or cannot open the key. No alternate token,
scheduled task,service,Explorer launch or privilege manipulation was attempted.
No exporter/collector/Go process launched. Earlier denial evidence remains valid
only for its original execution context.

Next handoff:the user may manually open PowerShell without Run as administrator
and perform the approved exact-key read-only check with an administrator guard.
Return only OPENED_READ_ONLY,NOT_FOUND,or the exception type/message;no values,
account identifiers or unrelated registry information are needed. If the guard
again detects administrator membership,stop. This comparison does not authorize
repairs,elevated registry access or a new exporter/capture run.

Normal-user comparison remains pending;observation BLOCKED;44/34/6/217,
NOT_ACCEPTED unchanged. No permissions/settings changed. Author review only.
