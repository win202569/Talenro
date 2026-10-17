# Offline objects only. Import loads the pure classifier,not a runtime helper.
. (Join-Path $PSScriptRoot 'output-channel-classifier.ps1')
function Get-OfflineEvidenceValueType([object]$Value){
    if($null -eq $Value){return $null}
    return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)
}
function Assert-OfflineEvidencePassiveValue($Value){
    if($null -eq $Value){return}
    foreach($property in $Value.PSObject.Properties){
        if($property.MemberType -in @([Management.Automation.PSMemberTypes]::ScriptProperty,[Management.Automation.PSMemberTypes]::CodeProperty)){throw 'Executable evidence property rejected'}
        if(([StringComparer]::OrdinalIgnoreCase.Equals($property.Name,'Length') -or [StringComparer]::OrdinalIgnoreCase.Equals($property.Name,'Rank')) -and $property.MemberType -ne [Management.Automation.PSMemberTypes]::Property){throw 'Evidence bound shadow rejected'}
    }
}
function Assert-OfflineEvidenceString($Value,[int]$Minimum,[int]$Maximum){
    if((Get-OfflineEvidenceValueType $Value) -ne [string]){throw 'Bounded evidence string required'}
    Assert-OfflineEvidencePassiveValue $Value
    if($Value.Length -lt $Minimum -or $Value.Length -gt $Maximum){throw 'Bounded evidence string required'}
}
function Test-OfflineEvidenceExact($Left,$Right){
    if($Left -isnot [string] -or $Right -isnot [string]){return $false}
    return [string]::Equals($Left,$Right,[StringComparison]::Ordinal)
}
function Test-OfflineEvidenceExactMember($Value,[string[]]$Names){
    if($Value -isnot [string]){return $false}
    foreach($name in $Names){if([string]::Equals($Value,$name,[StringComparison]::Ordinal)){return $true}}
    return $false
}
function Assert-OfflineEvidenceBytes($Value){
    if((Get-OfflineEvidenceValueType $Value) -ne [byte[]]){throw 'Both raw byte arrays required'}
    Assert-OfflineEvidencePassiveValue $Value
    if($Value.Length -gt 1048576){throw 'Evidence stream exceeds 1MiB'}
}
function Assert-OfflineEvidenceInteger($Value,[long]$Minimum,[long]$Maximum){
    if($Value -isnot [int] -and $Value -isnot [long]){throw 'Bounded integer required'}
    Assert-OfflineEvidencePassiveValue $Value
    if($Value -lt $Minimum -or $Value -gt $Maximum){throw 'Bounded integer required'}
}
function Assert-OfflineEvidenceMetadata($ExitCode,$EncodingName,$OriginDeclaration){
    Assert-OfflineEvidenceInteger $ExitCode ([int]::MinValue) ([int]::MaxValue)
    try{Assert-OfflineEvidenceString $EncodingName 1 8}catch{throw 'Explicit supported encoding required'}
    if(-not (Test-OfflineEvidenceExactMember $EncodingName @('utf-8','utf-16le'))){throw 'Explicit supported encoding required'}
    try{Assert-OfflineEvidenceString $OriginDeclaration 1 1024}catch{throw 'Bounded origin declaration required'}
}
function Assert-OfflineEvidenceShape($Value,[string[]]$Names){
    $type=Get-OfflineEvidenceValueType $Value
    if($null -eq $type -or -not (Test-OfflineEvidenceExact $type.FullName 'System.Management.Automation.PSCustomObject')){throw 'Plain evidence object required'}
    $properties=@($Value.PSObject.Properties)
    if($properties.Count -ne $Names.Count){throw 'Evidence fields missing or unexpected'}
    foreach($property in $properties){
        if($property.MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty -or -not (Test-OfflineEvidenceExactMember $property.Name $Names)){throw 'Plain exact evidence fields required'}
    }
}
function New-OfflineEvidenceStream([AllowEmptyCollection()][byte[]]$Bytes){
    Assert-OfflineEvidenceBytes $Bytes
    $info=Get-OfflineOutputByteInfo $Bytes
    [pscustomobject]@{Base64=[Convert]::ToBase64String($Bytes);ByteLength=$info.ByteLength;SHA256=$info.SHA256}
}
function Get-OfflineEvidenceStreamBytes($Value){
    Assert-OfflineEvidenceShape $Value @('Base64','ByteLength','SHA256')
    Assert-OfflineEvidenceInteger $Value.ByteLength 0 1048576
    try{Assert-OfflineEvidenceString $Value.Base64 0 1398104}catch{throw 'Bounded Base64 required'}
    if($Value.Base64 -cnotmatch '^[A-Za-z0-9+/]*={0,2}\z'){throw 'Bounded Base64 required'}
    try{Assert-OfflineEvidenceString $Value.SHA256 64 64}catch{throw 'Exact SHA256 required'}
    if($Value.SHA256 -cnotmatch '^[A-F0-9]{64}\z'){throw 'Exact SHA256 required'}
    $bytes=[Convert]::FromBase64String($Value.Base64)
    if(-not (Test-OfflineEvidenceExact ([Convert]::ToBase64String($bytes)) $Value.Base64)){throw 'Canonical Base64 required'}
    if($bytes.Length -gt 1048576 -or $bytes.Length -ne $Value.ByteLength){throw 'Evidence byte length mismatch'}
    if(-not (Test-OfflineEvidenceExact (Get-OfflineOutputByteInfo $bytes).SHA256 $Value.SHA256)){throw 'Evidence byte hash mismatch'}
    return ,$bytes
}
function New-OfflineOutputEvidencePackage(
    [AllowNull()][AllowEmptyCollection()][byte[]]$StdoutBytes,
    [AllowNull()][AllowEmptyCollection()][byte[]]$StderrBytes,
    [object]$ExitCode,[object]$EncodingName,[object]$OriginDeclaration
){
    Assert-OfflineEvidenceMetadata $ExitCode $EncodingName $OriginDeclaration
    if($null -eq $StdoutBytes -or $null -eq $StderrBytes){throw 'Both raw byte arrays required'}
    Assert-OfflineEvidenceBytes $StdoutBytes;Assert-OfflineEvidenceBytes $StderrBytes
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
        try{Assert-OfflineEvidenceString $Package.Policy 26 26}catch{throw 'Unknown evidence policy'}
        if(-not (Test-OfflineEvidenceExact $Package.Policy 'offline-output-evidence-v1')){throw 'Unknown evidence policy'}
        Assert-OfflineEvidenceMetadata $Package.ExitCode $Package.EncodingName $Package.OriginDeclaration
        foreach($name in @('OriginVerified','Acceptance','RuntimeApproved')){
            Assert-OfflineEvidencePassiveValue $Package.$name
            if($Package.$name -isnot [bool] -or $Package.$name){throw 'Evidence cannot authenticate or approve'}
        }
        try{Assert-OfflineEvidenceString $Package.Classification 1 13}catch{throw 'Unknown stored classification'}
        if(-not (Test-OfflineEvidenceExactMember $Package.Classification @('empty','progress-only','refused'))){throw 'Unknown stored classification'}
        Assert-OfflineEvidenceInteger $Package.ProgressCount 0 ([int]::MaxValue)
        $stdout=Get-OfflineEvidenceStreamBytes $Package.Stdout;$stderr=Get-OfflineEvidenceStreamBytes $Package.Stderr
        $classification=Get-OfflineOutputClassification -StdoutBytes $stdout -StderrBytes $stderr -ExitCode $Package.ExitCode -EncodingName $Package.EncodingName
        if(-not (Test-OfflineEvidenceExact $Package.Classification $classification.Classification) -or $Package.ProgressCount -ne $classification.ProgressCount){throw 'Stored classification disagrees with raw bytes'}
        $result.Verified=$true;$result.Reason='integrity-verified-only'
        $result.Classification=$classification.Classification;$result.ProgressCount=$classification.ProgressCount
    }catch{$result.Reason=$_.Exception.Message}
    return $result
}
