# Offline objects only. Import loads the pure classifier,not a runtime helper.
. (Join-Path $PSScriptRoot 'output-channel-classifier.ps1')
function Assert-OfflineEvidenceInteger($Value,[long]$Minimum,[long]$Maximum){
    if(($Value -isnot [int] -and $Value -isnot [long]) -or $Value -lt $Minimum -or $Value -gt $Maximum){throw 'Bounded integer required'}
}
function Assert-OfflineEvidenceMetadata($ExitCode,$EncodingName,$OriginDeclaration){
    Assert-OfflineEvidenceInteger $ExitCode ([int]::MinValue) ([int]::MaxValue)
    if($EncodingName -isnot [string] -or $EncodingName -cnotin @('utf-8','utf-16le')){throw 'Explicit supported encoding required'}
    if($OriginDeclaration -isnot [string] -or $OriginDeclaration.Length -lt 1 -or $OriginDeclaration.Length -gt 1024){throw 'Bounded origin declaration required'}
}
function Assert-OfflineEvidenceShape($Value,[string[]]$Names){
    if($null -eq $Value -or $Value.GetType().FullName -cne 'System.Management.Automation.PSCustomObject'){throw 'Plain evidence object required'}
    $properties=@($Value.PSObject.Properties)
    if($properties.Count -ne $Names.Count){throw 'Evidence fields missing or unexpected'}
    foreach($property in $properties){
        if($property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or $property.Name -cnotin $Names){throw 'Plain exact evidence fields required'}
    }
}
function New-OfflineEvidenceStream([AllowEmptyCollection()][byte[]]$Bytes){
    $info=Get-OfflineOutputByteInfo $Bytes
    [pscustomobject]@{Base64=[Convert]::ToBase64String($Bytes);ByteLength=$info.ByteLength;SHA256=$info.SHA256}
}
function Get-OfflineEvidenceStreamBytes($Value){
    Assert-OfflineEvidenceShape $Value @('Base64','ByteLength','SHA256')
    Assert-OfflineEvidenceInteger $Value.ByteLength 0 1048576
    if($Value.Base64 -isnot [string] -or $Value.Base64.Length -gt 1398104 -or $Value.Base64 -cnotmatch '^[A-Za-z0-9+/]*={0,2}\z'){throw 'Bounded Base64 required'}
    if($Value.SHA256 -isnot [string] -or $Value.SHA256 -cnotmatch '^[A-F0-9]{64}\z'){throw 'Exact SHA256 required'}
    $bytes=[Convert]::FromBase64String($Value.Base64)
    if([Convert]::ToBase64String($bytes) -cne $Value.Base64){throw 'Canonical Base64 required'}
    if($bytes.Length -gt 1048576 -or $bytes.Length -ne $Value.ByteLength){throw 'Evidence byte length mismatch'}
    if((Get-OfflineOutputByteInfo $bytes).SHA256 -cne $Value.SHA256){throw 'Evidence byte hash mismatch'}
    return ,$bytes
}
function New-OfflineOutputEvidencePackage(
    [AllowNull()][AllowEmptyCollection()][byte[]]$StdoutBytes,
    [AllowNull()][AllowEmptyCollection()][byte[]]$StderrBytes,
    [object]$ExitCode,[object]$EncodingName,[object]$OriginDeclaration
){
    Assert-OfflineEvidenceMetadata $ExitCode $EncodingName $OriginDeclaration
    if($null -eq $StdoutBytes -or $null -eq $StderrBytes){throw 'Both raw byte arrays required'}
    if($StdoutBytes.Length -gt 1048576 -or $StderrBytes.Length -gt 1048576){throw 'Evidence stream exceeds 1MiB'}
    $stdout=New-OfflineEvidenceStream $StdoutBytes;$stderr=New-OfflineEvidenceStream $StderrBytes
    # Classify the stored snapshot,not caller arrays that may later be mutated.
    $classification=Get-OfflineOutputClassification -StdoutBytes ([Convert]::FromBase64String($stdout.Base64)) -StderrBytes ([Convert]::FromBase64String($stderr.Base64)) -ExitCode $ExitCode -EncodingName $EncodingName
    [pscustomobject]@{Schema=1;Policy='offline-output-evidence-v1';EncodingName=$EncodingName;ExitCode=$ExitCode;OriginDeclaration=$OriginDeclaration;OriginVerified=$false;Stdout=$stdout;Stderr=$stderr;Classification=$classification.Classification;ProgressCount=$classification.ProgressCount;Acceptance=$false;RuntimeApproved=$false}
}
function Test-OfflineOutputEvidencePackage($Package){
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Schema=1;Verified=$false;Reason='unverified';Classification='refused';ProgressCount=0;OriginVerified=$false;Acceptance=$false;RuntimeApproved=$false}
    try{
        Assert-OfflineEvidenceShape $Package @('Schema','Policy','EncodingName','ExitCode','OriginDeclaration','OriginVerified','Stdout','Stderr','Classification','ProgressCount','Acceptance','RuntimeApproved')
        Assert-OfflineEvidenceInteger $Package.Schema 1 1
        if($Package.Policy -isnot [string] -or $Package.Policy -cne 'offline-output-evidence-v1'){throw 'Unknown evidence policy'}
        Assert-OfflineEvidenceMetadata $Package.ExitCode $Package.EncodingName $Package.OriginDeclaration
        foreach($name in @('OriginVerified','Acceptance','RuntimeApproved')){
            if($Package.$name -isnot [bool] -or $Package.$name){throw 'Evidence cannot authenticate or approve'}
        }
        if($Package.Classification -isnot [string] -or $Package.Classification -cnotin @('empty','progress-only','refused')){throw 'Unknown stored classification'}
        Assert-OfflineEvidenceInteger $Package.ProgressCount 0 ([int]::MaxValue)
        $stdout=Get-OfflineEvidenceStreamBytes $Package.Stdout;$stderr=Get-OfflineEvidenceStreamBytes $Package.Stderr
        $classification=Get-OfflineOutputClassification -StdoutBytes $stdout -StderrBytes $stderr -ExitCode $Package.ExitCode -EncodingName $Package.EncodingName
        if($Package.Classification -cne $classification.Classification -or $Package.ProgressCount -ne $classification.ProgressCount){throw 'Stored classification disagrees with raw bytes'}
        $result.Verified=$true;$result.Reason='integrity-verified-only'
        $result.Classification=$classification.Classification;$result.ProgressCount=$classification.ProgressCount
    }catch{$result.Reason=$_.Exception.Message}
    return $result
}
