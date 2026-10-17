# Focused memory-only diagnosis of normal verification before/after ETS fixtures.
# Run only under separately approved finite diagnostic authority;not a live probe.
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'capture-adapter-contract.ps1')
Write-Output ('HOST: '+$PSVersionTable.PSVersion+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
function Diagnostic-Packet{
    [pscustomobject]@{Schema=1;SourceMode='synthetic';SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z';Events=[object[]]@()}
}
function Show-DiagnosticMembers($Value,[string]$Label){
    $members=@($Value.PSObject.Properties|Where-Object{$_.Name -in @('Length','Rank')})
    foreach($member in $members){Write-Output ('MEMBER '+$Label+' '+$member.Name+' '+$member.MemberType)}
}
function Show-DiagnosticRecord([string]$Label){
    $record=New-SyntheticCaptureRecord (Diagnostic-Packet)
    $verdict=Test-SyntheticCaptureRecord $record
    Write-Output ('NORMAL '+$Label+' verified='+$verdict.Verified+' reason='+$verdict.Reason+' status='+$record.Lifecycle.Status)
    Show-DiagnosticMembers $record.Arguments.Base64 'arguments-base64'
    Show-DiagnosticMembers $record.Lifecycle.Status 'lifecycle-status'
    Show-DiagnosticMembers $record.Lifecycle.Reason 'lifecycle-reason'
    Show-DiagnosticMembers $record.Lifecycle.Stdout.Base64 'stdout-base64'
    Show-DiagnosticMembers $record.Events 'events'
    try{$null=Read-ContractSnapshot $record.Arguments 65536;Write-Output 'STEP arguments passed'}catch{Write-Output ('STEP arguments failed: '+$_.Exception.Message)}
    try{
        $clean=Get-ContractPacket (Diagnostic-Packet);$expected=New-ContractValidatedRecord $clean
        foreach($property in $expected.PSObject.Properties){
            try{Assert-ContractEquivalent $record.($property.Name) $property.Value;Write-Output ('FIELD '+$property.Name+' passed')}
            catch{Write-Output ('FIELD '+$property.Name+' failed: '+$_.Exception.Message);Write-Output $_.ScriptStackTrace}
        }
    }catch{Write-Output ('STEP expected reconstruction failed: '+$_.Exception.Message)}
}
Show-DiagnosticRecord 'before-shadow'
foreach($target in @('arguments','stdout','status')){
    $record=New-SyntheticCaptureRecord (Diagnostic-Packet)
    $value=if($target -eq 'arguments'){$record.Arguments.Base64}elseif($target -eq 'stdout'){$record.Lifecycle.Stdout.Base64}else{$record.Lifecycle.Status}
    $length=if($target -eq 'status'){10}else{0}
    $wrapped=Add-Member -InputObject $value -MemberType NoteProperty -Name Length -Value $length -Force -PassThru
    if($target -eq 'arguments'){$record.Arguments.Base64=$wrapped}elseif($target -eq 'stdout'){$record.Lifecycle.Stdout.Base64=$wrapped}else{$record.Lifecycle.Status=$wrapped}
    Write-Output ('SHADOW '+$target+' verified='+(Test-SyntheticCaptureRecord $record).Verified)
    if($target -eq 'arguments'){$record.Arguments.Base64.PSObject.Properties.Remove('Length')}
    elseif($target -eq 'stdout'){$record.Lifecycle.Stdout.Base64.PSObject.Properties.Remove('Length')}
    else{$record.Lifecycle.Status.PSObject.Properties.Remove('Length')}
    Show-DiagnosticRecord ('after-'+$target)
}
Write-Output 'DIAGNOSTIC COMPLETE: memory-only;no acceptance.'
