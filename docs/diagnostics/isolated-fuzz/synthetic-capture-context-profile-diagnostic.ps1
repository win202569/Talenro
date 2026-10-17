# Exploratory source-copy instrumentation only,not product or suite acceptance.
param([Parameter(Mandatory=$true)][ValidateSet('ordinary','instrumented')][string]$Mode)
$ErrorActionPreference='Stop'
$script:ctxActive=$false;$script:ctxFault=$false;$script:ctxMaxDepth=0
$script:ctxFrames=New-Object 'System.Collections.Generic.List[object]'
$script:ctxEntries=New-Object long[] 7;$script:ctxExits=New-Object long[] 7
$script:ctxInclusive=New-Object long[] 7;$script:ctxSubtracted=New-Object long[] 7
$names=@('Get-OfflineCaptureValueType','Get-OfflineEvidenceValueType','Assert-ContractPassiveProperties','Assert-OfflineEvidencePassiveValue','Assert-ContractShape','Get-ContractPacket','Assert-ContractEquivalent')
function Write-ContextRow($Row){[Console]::Out.WriteLine('CONTEXT '+(ConvertTo-Json -InputObject $Row -Compress));[Console]::Out.Flush()}
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess);[Console]::Out.Flush()
Write-ContextRow ([ordered]@{Event='start';Stage='setup';Mode=$Mode})
$setup=[Diagnostics.Stopwatch]::StartNew()
$source=Join-Path $PSScriptRoot 'capture-adapter-contract.ps1'
if($Mode -ceq 'instrumented'){$source=Join-Path ([IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))) '.superpowers/sdd/2026-10-17-synthetic-capture-context-profile/copies/capture-adapter-contract.ps1'}
. $source
$events=New-Object object[] 4096
for($i=0;$i -lt 4096;$i++){$events[$i]=[pscustomobject]@{Kind='tick';ElapsedMs=[long]0}}
$packet=[pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=$events}
$record=New-SyntheticCaptureRecord $packet
if(@($record.PSObject.Properties).Count -ne 20 -or $record.Events.Count -ne 4096){throw 'Fixture construction failed'}
Write-ContextRow ([ordered]@{Event='end';Stage='setup';Mode=$Mode;ElapsedMs=[Math]::Round($setup.Elapsed.TotalMilliseconds,3)})
Write-ContextRow ([ordered]@{Event='start';Stage='verify';Mode=$Mode})
$script:ctxActive=($Mode -ceq 'instrumented');$watch=[Diagnostics.Stopwatch]::StartNew()
try{$verdict=Test-SyntheticCaptureRecord $record}finally{$watch.Stop();$script:ctxActive=$false}
if(-not $verdict.Verified -or @($verdict.PSObject.Properties).Count -ne 5 -or $verdict.Reason -cne 'consistent-synthetic-only' -or $verdict.OriginVerified -or $verdict.Acceptance -or $verdict.RuntimeApproved){throw 'Public verification did not return expected verdict'}
if($record.Events.Count -ne 4096 -or $record.Lifecycle.AcceptedEventCount -ne 4096 -or $record.Lifecycle.Status -cne 'incomplete' -or $null -ne $record.Lifecycle.ExitCode -or $null -ne $record.Lifecycle.Package -or $record.LiveActionsPerformed -or $record.OriginVerified -or $record.Acceptance -or $record.RuntimeApproved){throw 'Unexpected fixture facts'}
if($script:ctxFault -or $script:ctxFrames.Count -ne 0){throw 'Unbalanced diagnostic stack'}
Write-ContextRow ([ordered]@{Event='end';Stage='verify';Mode=$Mode;ElapsedMs=[Math]::Round($watch.Elapsed.TotalMilliseconds,3);Verified=$true;BindingSHA256=$record.BindingSHA256})
$totalEntries=[long]0;$totalExits=[long]0
for($i=0;$i -lt 7;$i++){
    if($script:ctxEntries[$i] -ne $script:ctxExits[$i] -or $script:ctxSubtracted[$i] -lt 0 -or $script:ctxInclusive[$i] -lt $script:ctxSubtracted[$i]){throw 'Diagnostic count/timing inconsistency'}
    $totalEntries+=$script:ctxEntries[$i];$totalExits+=$script:ctxExits[$i]
    if($Mode -ceq 'instrumented'){
        if($script:ctxEntries[$i] -le 0){throw 'Missing instrumented target calls'}
        Write-ContextRow ([ordered]@{Event='metric';Function=$names[$i];Calls=$script:ctxEntries[$i];InclusiveMs=[Math]::Round(1000.0*$script:ctxInclusive[$i]/[Diagnostics.Stopwatch]::Frequency,3);SubtractedMs=[Math]::Round(1000.0*$script:ctxSubtracted[$i]/[Diagnostics.Stopwatch]::Frequency,3)})
    }
}
if($Mode -ceq 'ordinary' -and ($totalEntries -ne 0 -or $totalExits -ne 0)){throw 'Ordinary run recorded instrumentation'}
Write-ContextRow ([ordered]@{Event='checks';Mode=$Mode;Balanced=$true;Functions=$(if($Mode -ceq 'instrumented'){7}else{0});Entries=$totalEntries;Exits=$totalExits;MaxDepth=$script:ctxMaxDepth})
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: in-context public verifier timing;not suite acceptance.');[Console]::Out.Flush()
