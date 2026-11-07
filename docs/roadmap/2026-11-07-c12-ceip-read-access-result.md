# Exact-key read access — denied

User explicitly approved one read-only open of
HKCU\Software\Microsoft\Windows Kits\CEIP after acc4c34f,without reading values,
enumerating other locations,changing permissions or elevating. Stop on denial.

Tool69fb96 used RegistryKey.OpenBaseKey(CurrentUser,Default),then
OpenSubKey("Software\\Microsoft\\Windows Kits\\CEIP",false),with disposal in
finally. No GetValue/GetValueNames/GetSubKeyNames or write operation was called.
The combined operation returned:

```json
{"Result":"OPEN_FAILED","ExceptionType":"System.Security.SecurityException","Message":"Requested registry access is not allowed.","HResult":-2146233078}
```

The diagnostic command completed in0.658148s. Its outer exit0 means the exception
was captured,not that registry access succeeded. No retry or elevated check followed.
No exporter,WPR,WPA,Go,collector or installation operation was launched.

This independently reproduces registry read-access denial in the current tool
execution context without launching the exporter. It supports an access-context
blocker rather than a help-spelling remedy. It does not identify the policy source,
key ACL,existence/value of the key,or whether a normal interactive user would see
the same result. Because both open calls share one catch,the receipt alone does
not distinguish base-hive opening from subkey opening;do not overstate that detail.
The earlier exporter stack specifically named OpenSubKey,which is separate evidence.

Do not infer a corrupt installation or permission repair requirement. No permission,
telemetry preference,compatibility variable or security setting was changed. The
approved read-access attempt is consumed. Stop automatic diagnosis of this key;
further execution-context comparison requires a separately explicit scope through
the normal approval mechanism,not an alternate route around denied access.

Observation route remains BLOCKED. Native300s timeout attribution,collector
ownership,save-time resource limits and privacy approval remain unresolved.
44/34/6/217,NOT_ACCEPTED unchanged. Original evidence retained;author review only.
