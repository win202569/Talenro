Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$script:RIRepo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$script:RIRegistry=@{}
function New-RIPublisher(
    [ValidateSet('isolation-red','isolation-green')][string]$Grant,
    [ValidateSet('HostEvidence','Healthy','ExistingFailure')][string]$Case
){
    $key=$Grant+'/'+$Case
    if($script:RIRegistry.ContainsKey($key)){throw 'Publisher already declared'}
    $file=switch -CaseSensitive ($Case){'HostEvidence'{'evidence.json'}'Healthy'{'healthy.bin'}'ExistingFailure'{'existing-publisher.bin'}}
    $path=Join-Path $script:RIRepo ('.superpowers/sdd/2026-10-31-offline-record-isolation-repair/'+$Grant+'/'+$file)
    $value=[pscustomobject]@{Case=$Case;Path=$path;Status='Active';Consumed=$false;Open=0;Write=0;Flush=0;Dispose=0;Primary=$null;Secondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new();Overflow=$false}
    $script:RIRegistry[$key]=$value
    return $value
}
function Assert-RIPublisher($Publisher){
    $found=$false
    foreach($entry in $script:RIRegistry.Values){if([object]::ReferenceEquals($entry,$Publisher)){$found=$true;break}}
    if(-not $found){throw 'Unknown publisher reference'}
}
function Write-RIPublisherBytes($Publisher,[byte[]]$Bytes){
    Assert-RIPublisher $Publisher
    if($Publisher.Consumed -or $Publisher.Status -cne 'Active'){return [ordered]@{Ok=$false;Rejected=$true;Error=$null}}
    $Publisher.Consumed=$true
    $stream=$null
    try{
        if($null -eq $Bytes -or $Bytes.Length -gt 1048576){throw 'Publisher byte limit'}
        $Publisher.Open++
        $stream=[IO.File]::Open($Publisher.Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None) # RI-CALL
        $Publisher.Write++;$stream.Write($Bytes,0,$Bytes.Length)
        $Publisher.Flush++;$stream.Flush()
    }catch{$Publisher.Primary=$_}
    finally{
        if($null -ne $stream){
            $Publisher.Dispose++
            try{$stream.Dispose()}catch{if($null -eq $Publisher.Primary){$Publisher.Primary=$_}elseif($Publisher.Secondary.Count -lt 8){$Publisher.Secondary.Add($_)}else{$Publisher.Overflow=$true}}
        }
    }
    $Publisher.Status=if($null -eq $Publisher.Primary){'Completed'}else{'Failed'}
    return [ordered]@{Ok=($null -eq $Publisher.Primary);Rejected=$false;Error=$Publisher.Primary}
}
function Get-RIPublisherView($Publisher){
    Assert-RIPublisher $Publisher
    return [ordered]@{Case=$Publisher.Case;Path=$Publisher.Path;Status=$Publisher.Status;Consumed=$Publisher.Consumed;Open=$Publisher.Open;Write=$Publisher.Write;Flush=$Publisher.Flush;Dispose=$Publisher.Dispose;Primary=$Publisher.Primary;Secondary=$Publisher.Secondary.ToArray();Overflow=$Publisher.Overflow}
}
Export-ModuleMember
