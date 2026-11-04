param(
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$InputSHA256,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$IntentSHA256
)
Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
$script:PTClock=[Diagnostics.Stopwatch]::StartNew()
$script:PTRepo='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator'
$d=Join-Path $script:PTRepo '.superpowers/sdd/2026-11-04-runtime-prefix-timing-v1'
$l='docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2'
$inputPath=Join-Path $d 'input.json';$intentPath=Join-Path $d 'intent.json'
if((Get-FileHash $inputPath).Hash -cne $InputSHA256 -or (Get-FileHash $intentPath).Hash -cne $IntentSHA256){throw 'PT: bootstrap identity'}
$script:PTBinding=Get-Content $inputPath -Raw|ConvertFrom-Json -AsHashtable
foreach($path in $script:PTBinding.HelperSHA256.Keys){if((Get-FileHash (Join-Path $script:PTRepo $path)).Hash -cne $script:PTBinding.HelperSHA256[$path]){throw 'PT: helper identity'}}
. (Join-Path $PSScriptRoot 'runtime-prefix-timing-common.ps1')
$intent=Get-Content $intentPath -Raw|ConvertFrom-Json -AsHashtable
if($intent.Schema -ne 1 -or $intent.Computation -ne 27 -or $intent.Ceiling -ne 27 -or $intent.MaxProcesses -ne 1 -or $intent.InputSHA256 -cne $InputSHA256 -or $intent.AttemptConsumed -isnot [bool] -or -not $intent.AttemptConsumed -or $intent.Token -cnotmatch '^[0-9a-f-]{36}$'){throw 'PT: intent identity'}
if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'PT: runtime'}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent()
if($identity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or ([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'PT: fixed identity'}
Write-PTOnce (Join-Path $d 'child.claim.json') ([ordered]@{Schema=1;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Token=$intent.Token})
$script:PTJournal=[IO.File]::Open((Join-Path $d 'journal.jsonl'),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
$script:PTStack=[Collections.Generic.Stack[object]]::new()
$script:PTSeq=0L;$script:PTEventId=0L;$script:PTJournalBytes=0L;$script:PTJournalFailed=$false;$script:PTLastEvent=$null;$script:PTProfile=''
$before=$null;$after=$null;$failure=$null;$recording=$null;$code=2;$status='INCOMPLETE';$assertionEntered=$false;$assertionCompleted=$false
try{
    Write-PTEvent initialization begin '' ''
    $mapPath=Join-Path $d 'instrumentation.json'
    if((Get-PTHash $mapPath) -cne $script:PTBinding.InstrumentationSHA256){throw 'PT: instrumentation identity'}
    $instrumentation=Get-Content $mapPath -Raw|ConvertFrom-Json -AsHashtable
    foreach($entry in $instrumentation.Definitions){
        $original=@(Get-PTDefinitions (Join-Path $script:PTRepo ($l+'/'+$entry.File)) $instrumentation.Allowlist[$entry.File])|Where-Object {$_ -cmatch ('^function '+[regex]::Escape($entry.Name)+'\(')}
        if(@($original).Count -ne 1 -or [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes([string]$original))) -cne $entry.OriginalSHA256){throw 'PT: original definition drift'}
        if([Convert]::ToHexString([Security.Cryptography.SHA256]::HashData([Text.Encoding]::UTF8.GetBytes($entry.Text))) -cne $entry.TransformedSHA256){throw 'PT: transformed definition drift'}
        . ([scriptblock]::Create($entry.Text))
    }
    Write-PTEvent initialization end '' ''
    Write-PTEvent input.verify.before begin '' ''
    $before=Assert-PTInputs $script:PTBinding
    Write-PTEvent input.verify.before end '' ''
    $manifestPath=Join-Path $script:PTRepo $script:PTBinding.ManifestPath
    Write-PTEvent manifest.hash begin '' $manifestPath
    if((Get-PTHash $manifestPath) -cne $script:PTBinding.FrozenSHA256){throw 'PT: frozen identity'}
    Write-PTEvent manifest.hash end '' $manifestPath
    Write-PTEvent manifest.read begin '' $manifestPath
    $manifestText=Get-Content $manifestPath -Raw
    Write-PTEvent manifest.read end '' $manifestPath
    Write-PTEvent manifest.json begin '' $manifestPath
    $manifest=ConvertFrom-Json -AsHashtable -InputObject $manifestText
    Write-PTEvent manifest.json end '' $manifestPath
    $frames=[Collections.Generic.List[object]]::new();$contexts=[Collections.Generic.List[object]]::new()
    foreach($ref in $script:PTBinding.EvidenceOrder){
        $script:PTProfile=$ref.Profile;$path=Join-Path $script:PTRepo $ref.Path
        Write-PTEvent evidence.hash begin $script:PTProfile $path
        if((Get-PTHash $path) -cne $script:PTBinding.HistoricalEvidenceSHA256[$ref.Path]){throw 'PT: evidence identity'}
        Write-PTEvent evidence.hash end $script:PTProfile $path
        Write-PTEvent evidence.read begin $script:PTProfile $path
        $json=Get-Content $path -Raw
        Write-PTEvent evidence.read end $script:PTProfile $path
        Write-PTEvent evidence.json begin $script:PTProfile $path
        $e=ConvertFrom-Json -AsHashtable -InputObject $json
        Write-PTEvent evidence.json end $script:PTProfile $path
        Write-PTEvent evidence.extract begin $script:PTProfile $path
        $returns=@($e.Observations|Where-Object {$_.Name -ceq 'ProfileReturn'})
        if($returns.Count -ne 1 -or $returns[0].Value.Profile -cne $ref.Profile){throw 'PT: canonical profile'}
        $frames.Add($returns[0].Value.Observations)
        Write-PTEvent evidence.extract end $script:PTProfile $path
        Write-PTEvent context begin $script:PTProfile ''
        $contexts.Add((New-ODHistoricalContext $script:PTRepo $ref.Profile $manifest))
        Write-PTEvent context end $script:PTProfile ''
    }
    $script:PTProfile='entry-migration-core'
    Write-PTEvent baseline.copy begin $script:PTProfile ''
    $rows=ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $frames[0] -Depth 60)
    Write-PTEvent baseline.copy end $script:PTProfile ''
    Write-PTEvent baseline.assert begin $script:PTProfile ''
    $assertionEntered=$true
    $emitted=@(Assert-ODProfileObservations $rows $script:PTProfile $contexts[0])
    if($emitted.Count){throw 'PT: assertion emitted output'}
    $assertionCompleted=$true
    Write-PTEvent baseline.assert end $script:PTProfile ''
    $status='PREFIX_COMPLETED';$code=0
}catch{
    $failure=Get-PTError $_
    if($assertionEntered -and -not $assertionCompleted -and $failure.Type -ceq 'System.Management.Automation.RuntimeException' -and $failure.Message -ceq 'OD: observation contract mismatch' -and $failure.ErrorId -ceq $failure.Message){$status='ASSERTION_REJECTED';$code=1}
    try{if(-not $script:PTJournalFailed -and $script:PTStack.Count){$top=$script:PTStack.Peek();Write-PTEvent $top.Stage error $top.Profile $top.Source}}catch{$recording=Get-PTError $_}
}
try{
    Write-PTEvent input.verify.after begin '' ''
    $after=Assert-PTInputs $script:PTBinding
    Write-PTEvent input.verify.after end '' ''
}catch{$recording=Get-PTError $_;$code=2;$status='INCOMPLETE'}
try{$script:PTJournal.Dispose()}catch{$recording=Get-PTError $_;$code=2;$status='INCOMPLETE'}
Write-PTOnce (Join-Path $d 'result.json') ([ordered]@{Schema=1;Status=$status;ExitCode=$code;Error=$failure;RecordingError=$recording;InputSHA256=$InputSHA256;IntentSHA256=$IntentSHA256;Before=$before;After=$after;LastEvent=$script:PTLastEvent;UnclosedIntervals=$script:PTStack.Count;SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false})
exit $code
