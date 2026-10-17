# One authorized pure-memory type diagnostic;no real collection or OS queries.
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'output-capture-simulation.ps1')
$script:typeDiagnosticCalls=0
$plain=[pscustomobject]@{Kind='tick';ElapsedMs=0}
$extended=[pscustomobject]@{Kind='tick';ElapsedMs=0}
$extended|Add-Member -MemberType ScriptMethod -Name GetType -Force -Value {$script:typeDiagnosticCalls++;[System.Management.Automation.PSCustomObject]}
$method=[object].GetMethod('GetType')
foreach($fixture in @([pscustomobject]@{Name='plain';Value=$plain},[pscustomobject]@{Name='extended';Value=$extended})){
    $value=$fixture.Value;$events=[object[]]@($value)
    $direct=$method.Invoke($value,$null);$base=$method.Invoke($value.PSObject.BaseObject,$null)
    $arrayDirect=$method.Invoke($events,$null);$arrayBase=$method.Invoke($events.PSObject.BaseObject,$null)
    $validation='accepted';try{Assert-OfflineCaptureEvent $value}catch{$validation=$_.Exception.Message}
    $r=Get-OfflineCaptureSimulation -Events $events -EncodingName 'utf-8' -OriginDeclaration 'type-diagnostic-only'
    [pscustomobject]@{Fixture=$fixture.Name;ReflectionType=$direct.FullName;BaseReflectionType=$base.FullName;ArrayReflectionType=$arrayDirect.FullName;ArrayBaseReflectionType=$arrayBase.FullName;IsPSCustomObject=($value -is [System.Management.Automation.PSCustomObject]);EventValidation=$validation;Status=$r.Status;Reason=$r.Reason;Accepted=$r.AcceptedEventCount;CallbackCount=$script:typeDiagnosticCalls}|ConvertTo-Json -Compress
}
if($script:typeDiagnosticCalls -ne 0){throw 'Diagnostic unexpectedly invoked input callback'}
Write-Output 'LIMIT: synthetic memory type observations only;no source correction or full-suite verdict.'
