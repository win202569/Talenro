# Exploratory name-extraction comparison only,not product code or acceptance.
# Separate finite launch authority required. No product helper replacement,
# value/descriptor cache,archive I/O or public verifier/suite replay.
$ErrorActionPreference='Stop'
$script:probeGetterCalls=0
function Get-ProbePipelineNames([object]$Value){
    $names=@($Value.PSObject.Properties|ForEach-Object{$_.Name})
    return ,$names
}
function Get-ProbeForeachNames([object]$Value){
    $names=@(foreach($property in $Value.PSObject.Properties){$property.Name})
    return ,$names
}
function Assert-ProbeNames($Names,[object[]]$Expected){
    if($null -eq $Names -or [object].GetMethod('GetType').Invoke($Names.PSObject.BaseObject,$null) -ne [object[]] -or $Names.Rank -ne 1 -or $Names.Length -ne $Expected.Length){throw 'Name array shape mismatch'}
    for($j=0;$j -lt $Expected.Length;$j++){
        if($Names[$j] -isnot [string] -or -not [string]::Equals($Names[$j],$Expected[$j],[StringComparison]::Ordinal)){throw 'Name order/text mismatch'}
    }
}
function Write-ProbeRow($Row){[Console]::Out.WriteLine('NAMES '+(ConvertTo-Json -InputObject $Row -Compress));[Console]::Out.Flush()}
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess);[Console]::Out.Flush()
if($PSVersionTable.PSVersion.Major -notin @(5,7)){throw 'Unexpected diagnostic host'}
$unicodeName=([string][char]0x00E9)+'e'+[char]0x0301
$unicodeFields=[ordered]@{};$unicodeFields.Add($unicodeName,1);$unicodeFields.Add('Plain',2)
$getterFixture=[pscustomobject][ordered]@{First=1}
Add-Member -InputObject $getterFixture -MemberType ScriptProperty -Name Hidden -Value {$script:probeGetterCalls++;throw 'Getter must not run'}
$cases=@(
    [pscustomobject]@{Name='empty';Value=[pscustomobject]@{};Expected=[object[]]@()},
    [pscustomobject]@{Name='single';Value=[pscustomobject][ordered]@{Only=1};Expected=[object[]]@('Only')},
    [pscustomobject]@{Name='multiple';Value=[pscustomobject][ordered]@{Zeta=1;Alpha=2;Middle=3};Expected=[object[]]@('Zeta','Alpha','Middle')},
    [pscustomobject]@{Name='unicode';Value=[pscustomobject]$unicodeFields;Expected=[object[]]@($unicodeName,'Plain')},
    [pscustomobject]@{Name='getter-descriptor';Value=$getterFixture;Expected=[object[]]@('First','Hidden')}
)
foreach($helper in @('Get-ProbePipelineNames','Get-ProbeForeachNames')){
    foreach($case in $cases){$names=& $helper -Value $case.Value;Assert-ProbeNames $names $case.Expected}
}
if($script:probeGetterCalls -ne 0){throw 'Fixture getter evaluated'}
Write-ProbeRow ([ordered]@{Event='fixtures';CasesPerVariant=5;Variants=2;GetterCalls=$script:probeGetterCalls})
$events=New-Object object[] 4096
for($i=0;$i -lt 4096;$i++){$events[$i]=[pscustomobject][ordered]@{Kind='tick';ElapsedMs=[long]0}}
$expectedNames=[object[]]@('Kind','ElapsedMs')
$order=if($PSVersionTable.PSVersion.Major -eq 7){@('pipeline','foreach')}else{@('foreach','pipeline')}
foreach($variant in $order){
    $helper=if($variant -ceq 'pipeline'){'Get-ProbePipelineNames'}else{'Get-ProbeForeachNames'}
    Write-ProbeRow ([ordered]@{Event='start';Variant=$variant;Objects=4096})
    $watch=[Diagnostics.Stopwatch]::StartNew()
    for($i=0;$i -lt 4096;$i++){$names=& $helper -Value $events[$i];Assert-ProbeNames $names $expectedNames}
    $watch.Stop()
    if($script:probeGetterCalls -ne 0){throw 'Timed getter evaluated'}
    Write-ProbeRow ([ordered]@{Event='end';Variant=$variant;Objects=4096;ElapsedMs=[Math]::Round($watch.Elapsed.TotalMilliseconds,3);GetterCalls=$script:probeGetterCalls})
}
[Console]::Out.WriteLine('DIAGNOSTIC COMPLETE: ordered property names only;not public verifier or suite acceptance.');[Console]::Out.Flush()
