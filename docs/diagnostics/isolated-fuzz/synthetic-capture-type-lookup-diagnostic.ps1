# Exploratory metadata-only comparison;not product code or acceptance.
# Every launch requires separate finite authorization. No product functions
# loaded/replaced,caller result caching,archive I/O or public suite replay.
$ErrorActionPreference='Stop'
$script:probeCallbacks=0
$script:probeFixedMethodInfo=[object].GetMethod('GetType')
function Get-ProbeBaselineType([object]$Value){
    if($null -eq $Value){return $null}
    return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)
}
function Get-ProbeCandidateType([object]$Value){
    if($null -eq $Value){return $null}
    return $script:probeFixedMethodInfo.Invoke($Value.PSObject.BaseObject,$null)
}
function Write-ProbeRow($Row){[Console]::Out.WriteLine('LOOKUP '+(ConvertTo-Json -InputObject $Row -Compress));[Console]::Out.Flush()}
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess);[Console]::Out.Flush()
if($PSVersionTable.PSVersion.Major -notin @(5,7)){throw 'Unexpected diagnostic host'}
if($script:probeFixedMethodInfo.DeclaringType -ne [object] -or $script:probeFixedMethodInfo.Name -cne 'GetType' -or $script:probeFixedMethodInfo.IsStatic){throw 'Unexpected fixed metadata'}
$forged=[pscustomobject]@{Kind='tick';ElapsedMs=[long]0}
Add-Member -InputObject $forged -MemberType ScriptMethod -Name GetType -Force -Value {$script:probeCallbacks++;[string]}
$cases=@(
    [pscustomobject]@{Name='null';Value=$null;Expected=$null},
    [pscustomobject]@{Name='string';Value='tick';Expected=[string]},
    [pscustomobject]@{Name='int32';Value=[int]0;Expected=[int]},
    [pscustomobject]@{Name='int64';Value=[long]0;Expected=[long]},
    [pscustomobject]@{Name='boolean';Value=$false;Expected=[bool]},
    [pscustomobject]@{Name='objects';Value=[object[]]@(1,'x');Expected=[object[]]},
    [pscustomobject]@{Name='bytes';Value=[byte[]]@(0,255);Expected=[byte[]]},
    [pscustomobject]@{Name='forged-gettype';Value=$forged;Expected=[System.Management.Automation.PSCustomObject]}
)
foreach($helper in @('Get-ProbeBaselineType','Get-ProbeCandidateType')){
    foreach($case in $cases){$observed=& $helper -Value $case.Value;if($observed -ne $case.Expected){throw ('Fixture type mismatch: '+$case.Name)}}
}
if($script:probeCallbacks -ne 0){throw 'Caller GetType dispatched'}
Write-ProbeRow ([ordered]@{Event='fixtures';CasesPerVariant=8;Variants=2;Callbacks=$script:probeCallbacks})
# Five object,three string,one int64 calls per nine approximates the static
# per-tick type-query mix of the two audited paths. No null in timed groups.
$values=[object[]]@($forged,$forged,$forged,'tick',$forged,$forged,'tick','tick',[long]0)
$expectedTypes=[object[]]@([System.Management.Automation.PSCustomObject],[System.Management.Automation.PSCustomObject],[System.Management.Automation.PSCustomObject],[string],[System.Management.Automation.PSCustomObject],[System.Management.Automation.PSCustomObject],[string],[string],[long])
$order=if($PSVersionTable.PSVersion.Major -eq 7){@('baseline','candidate')}else{@('candidate','baseline')}
foreach($variant in $order){
    $helper=if($variant -ceq 'baseline'){'Get-ProbeBaselineType'}else{'Get-ProbeCandidateType'}
    Write-ProbeRow ([ordered]@{Event='start';Variant=$variant;Calls=36864})
    $watch=[Diagnostics.Stopwatch]::StartNew()
    for($i=0;$i -lt 36864;$i++){
        $slot=$i%9;$observed=& $helper -Value $values[$slot]
        if($observed -ne $expectedTypes[$slot]){throw 'Timed type mismatch'}
    }
    $watch.Stop()
    if($script:probeCallbacks -ne 0){throw 'Timed caller GetType dispatched'}
    Write-ProbeRow ([ordered]@{Event='end';Variant=$variant;Calls=36864;ElapsedMs=[Math]::Round($watch.Elapsed.TotalMilliseconds,3);Callbacks=$script:probeCallbacks})
}
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: metadata lookup comparison only;not public verifier or suite acceptance.');[Console]::Out.Flush()
