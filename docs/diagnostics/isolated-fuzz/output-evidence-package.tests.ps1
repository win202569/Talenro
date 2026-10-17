$ErrorActionPreference='Stop'
function Check-Evidence([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'output-evidence-package.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Evidence ($null -ne (Get-Command New-OfflineOutputEvidencePackage -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineEvidencePackageRequired'
Write-Output ('HOST: '+$PSVersionTable.PSVersion+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
# Breaks caught:non-exact labels,caller GetType dispatch,executable/length shadows.
$script:evidenceBoundaryFailures=0;$script:evidenceBoundaryCallbacks=0
function Check-EvidenceBoundary([bool]$Value,[string]$Name){
    if($Value){Write-Output ('PASS: '+$Name)}else{$script:evidenceBoundaryFailures++;Write-Output ('FAIL: '+$Name)}
}
function New-BoundaryEvidencePackage{New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@(97,98,99)) -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'boundary-fixture'}
function Check-BoundaryEvidenceRefusal($Value,[string]$Name){
    $actual=Test-OfflineOutputEvidencePackage $Value
    Check-EvidenceBoundary (-not $actual.Verified -and -not $actual.OriginVerified -and -not $actual.Acceptance -and -not $actual.RuntimeApproved) $Name
}
function Check-BoundaryConstructionRefusal($OutBytes,$Encoding,$Origin,[string]$Name){
    $rejected=$false
    try{$null=New-OfflineOutputEvidencePackage -StdoutBytes $OutBytes -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName $Encoding -OriginDeclaration $Origin}catch{$rejected=$true}
    Check-EvidenceBoundary $rejected $Name
}
function Add-BoundaryLength($Value,$Length){
    $wrapped=Add-Member -InputObject $Value -MemberType NoteProperty -Name Length -Value $Length -Force -PassThru
    if($wrapped.PSObject.Properties['Length'].MemberType -ne [Management.Automation.PSMemberTypes]::NoteProperty){throw 'Boundary length fixture not attached'}
    return ,$wrapped
}
$boundaryNul=[string][char]0;$boundaryIgnorable=[string][char]0x00AD
foreach($mutation in @('policy-null','policy-ignorable','classification-null','encoding-null-refused')){
    $bad=New-BoundaryEvidencePackage
    switch($mutation){
        policy-null{$bad.Policy+=''+$boundaryNul}
        policy-ignorable{$bad.Policy='offline-output-'+$boundaryIgnorable+'evidence-v1'}
        classification-null{$bad.Classification='em'+$boundaryNul+'pty'}
        encoding-null-refused{$bad.EncodingName='utf'+$boundaryNul+'-8';$bad.Classification='refused'}
    }
    Check-BoundaryEvidenceRefusal $bad ('BoundaryExactLabel-'+$mutation)
}
$bad=New-BoundaryEvidencePackage;$bad.PSObject.Properties.Remove('Policy');Add-Member -InputObject $bad -NotePropertyName ('Policy'+$boundaryNul) -NotePropertyValue 'offline-output-evidence-v1'
Check-BoundaryEvidenceRefusal $bad 'BoundaryExactFieldName'
foreach($token in @(('utf'+$boundaryNul+'-8'),('utf'+$boundaryIgnorable+'-8'),'UTF-8')){
    Check-BoundaryConstructionRefusal ([byte[]]@()) $token 'fixture' 'BoundaryConstructorExactEncoding'
}
foreach($target in @('package','stdout','stderr')){
    $bad=New-BoundaryEvidencePackage;$script:evidenceBoundaryCallbacks=0
    $object=if($target -eq 'package'){$bad}elseif($target -eq 'stdout'){$bad.Stdout}else{$bad.Stderr}
    Add-Member -InputObject $object -MemberType ScriptMethod -Name GetType -Force -Value {$script:evidenceBoundaryCallbacks++;[Management.Automation.PSCustomObject]}
    $actual=Test-OfflineOutputEvidencePackage $bad
    Check-EvidenceBoundary ($actual.Verified -and $script:evidenceBoundaryCallbacks -eq 0) ('BoundaryUnusedGetTypeNeverCalled-'+$target)
}
$script:evidenceBoundaryCallbacks=0;$fake=@{}
Add-Member -InputObject $fake -MemberType ScriptMethod -Name GetType -Force -Value {$script:evidenceBoundaryCallbacks++;[Management.Automation.PSCustomObject]}
$actual=Test-OfflineOutputEvidencePackage $fake
Check-EvidenceBoundary (-not $actual.Verified -and $script:evidenceBoundaryCallbacks -eq 0) 'BoundaryWrongContainerCannotDispatchGetType'
foreach($target in @('package','stdout','stderr')){
    $bad=New-BoundaryEvidencePackage;$script:evidenceBoundaryCallbacks=0
    if($target -eq 'package'){$bad.PSObject.Properties.Remove('Policy');Add-Member -InputObject $bad -MemberType ScriptProperty -Name Policy -Value {$script:evidenceBoundaryCallbacks++;'offline-output-evidence-v1'}}
    elseif($target -eq 'stdout'){$bad.Stdout.PSObject.Properties.Remove('Base64');Add-Member -InputObject $bad.Stdout -MemberType ScriptProperty -Name Base64 -Value {$script:evidenceBoundaryCallbacks++;'YWJj'}}
    else{$bad.Stderr.PSObject.Properties.Remove('SHA256');Add-Member -InputObject $bad.Stderr -MemberType ScriptProperty -Name SHA256 -Value {$script:evidenceBoundaryCallbacks++;'unused'}}
    $actual=Test-OfflineOutputEvidencePackage $bad
    Check-EvidenceBoundary (-not $actual.Verified -and $script:evidenceBoundaryCallbacks -eq 0) ('BoundaryExecutableObjectPropertyRefused-'+$target)
}
$bad=New-BoundaryEvidencePackage;$script:evidenceBoundaryCallbacks=0
$bad.Stdout.Base64=Add-Member -InputObject $bad.Stdout.Base64 -MemberType ScriptProperty -Name Length -Force -PassThru -Value {$script:evidenceBoundaryCallbacks++;4}
$actual=Test-OfflineOutputEvidencePackage $bad
Check-EvidenceBoundary (-not $actual.Verified -and $script:evidenceBoundaryCallbacks -eq 0) 'BoundaryScalarLengthGetterNeverCalled'
$bad.Stdout.Base64.PSObject.Properties.Remove('Length')
$bad=New-BoundaryEvidencePackage;$script:evidenceBoundaryCallbacks=0
$bad.Policy=Add-Member -InputObject $bad.Policy -MemberType ScriptProperty -Name Extra -PassThru -Value {$script:evidenceBoundaryCallbacks++;'unused'}
$actual=Test-OfflineOutputEvidencePackage $bad
Check-EvidenceBoundary (-not $actual.Verified -and $script:evidenceBoundaryCallbacks -eq 0) 'BoundaryUnusedExecutableScalarPropertyRefused'
$bad.Policy.PSObject.Properties.Remove('Extra')
Check-BoundaryConstructionRefusal (Add-BoundaryLength ([byte[]]@(97,98,99)) 0) 'utf-8' 'fixture' 'BoundaryByteLengthShadowRefused-small'
Check-BoundaryConstructionRefusal (Add-BoundaryLength (New-Object byte[] 1048577) 0) 'utf-8' 'fixture' 'BoundaryByteLengthShadowRefused-oversized'
Check-BoundaryConstructionRefusal ([byte[]]@()) 'utf-8' (Add-BoundaryLength ('x'*1025) 1) 'BoundaryOriginLengthShadowRefused'
foreach($target in @('base64','hash')){
    $bad=New-BoundaryEvidencePackage
    if($target -eq 'base64'){$bad.Stdout.Base64=Add-BoundaryLength $bad.Stdout.Base64 1}
    else{$bad.Stdout.SHA256=Add-BoundaryLength $bad.Stdout.SHA256 64}
    Check-BoundaryEvidenceRefusal $bad ('BoundarySnapshotLengthShadowRefused-'+$target)
    if($target -eq 'base64'){$bad.Stdout.Base64.PSObject.Properties.Remove('Length')}else{$bad.Stdout.SHA256.PSObject.Properties.Remove('Length')}
}
$script:evidenceBoundaryCallbacks=0
$scriptedBytes=Add-Member -InputObject ([byte[]]@(97,98,99)) -MemberType ScriptProperty -Name Length -Force -PassThru -Value {$script:evidenceBoundaryCallbacks++;0}
$rejected=$false;try{$null=New-OfflineOutputEvidencePackage -StdoutBytes $scriptedBytes -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'fixture'}catch{$rejected=$true}
Check-EvidenceBoundary ($rejected -and $script:evidenceBoundaryCallbacks -eq 0) 'BoundaryByteLengthGetterNeverCalled'
$scriptedBytes.PSObject.Properties.Remove('Length')
$unicodeOrigin=([string][char]0x00E9)+'e'+[char]0x0301+[char]0x6A21
$good=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@(97,98,99)) -StderrBytes ([byte[]]@(0,255,13,10)) -ExitCode ([int]-1) -EncodingName 'utf-8' -OriginDeclaration $unicodeOrigin
Check-EvidenceBoundary ((Test-OfflineOutputEvidencePackage $good).Verified -and [string]::Equals($good.OriginDeclaration,$unicodeOrigin,[StringComparison]::Ordinal) -and [string]::Equals($good.Stdout.Base64,'YWJj',[StringComparison]::Ordinal) -and [string]::Equals($good.Stderr.Base64,'AP8NCg==',[StringComparison]::Ordinal)) 'BoundaryUnicodeOriginAndOriginalBytesPreserved'
Check-Evidence ($script:evidenceBoundaryFailures -eq 0) 'EvidenceBoundaryRequired'
$source=[byte[]]@(97,98,99);$empty=[byte[]]@()
$package=New-OfflineOutputEvidencePackage -StdoutBytes $source -StderrBytes $empty -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'synthetic-only'
Check-Evidence ($package.Stdout.Base64 -ceq 'YWJj' -and $package.Stdout.ByteLength -eq 3 -and $package.Stdout.SHA256 -ceq 'BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD' -and $package.Stderr.Base64 -ceq '' -and $package.Stderr.ByteLength -eq 0) 'LosslessIndependentKnownBytes'
Check-Evidence (-not $package.OriginVerified -and -not $package.Acceptance -and -not $package.RuntimeApproved -and $package.Classification -ceq 'empty') 'PackageNeverAuthenticatesOrApproves'
$verified=Test-OfflineOutputEvidencePackage $package
Check-Evidence ($verified.Verified -and $verified.Classification -ceq 'empty' -and -not $verified.OriginVerified -and -not $verified.Acceptance -and -not $verified.RuntimeApproved) 'ReverificationIsIntegrityOnly'
$source[0]=122
Check-Evidence ($package.Stdout.Base64 -ceq 'YWJj' -and (Test-OfflineOutputEvidencePackage $package).Verified) 'PackageIndependentOfLaterCallerMutation'
$roundtrip=$package|ConvertTo-Json -Depth 8|ConvertFrom-Json
Check-Evidence ((Test-OfflineOutputEvidencePackage $roundtrip).Verified) 'PlainObjectJsonRoundtripWithoutFileIO'
foreach($mutation in @('schema','policy','encoding','exit-string','exit-bool','exit-range','origin-verified','acceptance','runtime','classification','count','missing','unknown','leaf-unknown','base64-data','base64-invalid','base64-whitespace','hash','hash-newline','length','length-bool','length-float','base64-oversized','origin-empty','origin-oversized')){
    $bad=$package|ConvertTo-Json -Depth 8|ConvertFrom-Json
    switch($mutation){
        schema{$bad.Schema=2} policy{$bad.Policy='other'} encoding{$bad.EncodingName='guess'}
        exit-string{$bad.ExitCode='0'} exit-bool{$bad.ExitCode=$false} exit-range{$bad.ExitCode=[long]2147483648}
        origin-verified{$bad.OriginVerified=$true} acceptance{$bad.Acceptance=$true} runtime{$bad.RuntimeApproved=$true}
        classification{$bad.Classification='progress-only'} count{$bad.ProgressCount=99}
        missing{$bad.PSObject.Properties.Remove('Stdout')} unknown{$bad|Add-Member Extra 'unexpected'}
        leaf-unknown{$bad.Stdout|Add-Member Extra 'unexpected'}
        base64-data{$bad.Stdout.Base64='YWJk'} base64-invalid{$bad.Stdout.Base64='!'}
        base64-whitespace{$bad.Stdout.Base64="YWJj`n"} hash{$bad.Stdout.SHA256='A'*64}
        hash-newline{$bad.Stdout.SHA256+="`n"} length{$bad.Stdout.ByteLength=2}
        length-bool{$bad.Stdout.ByteLength=$true} length-float{$bad.Stdout.ByteLength=[double]3}
        base64-oversized{$bad.Stdout.Base64='A'*1398108} origin-empty{$bad.OriginDeclaration=''}
        origin-oversized{$bad.OriginDeclaration='x'*1025}
    }
    $result=Test-OfflineOutputEvidencePackage $bad
    Check-Evidence (-not $result.Verified -and $result.Classification -ceq 'refused' -and -not $result.Acceptance -and -not $result.RuntimeApproved) ('InconsistentPackageRefused-'+$mutation)
}
$noncanonical=$package|ConvertTo-Json -Depth 8|ConvertFrom-Json
$noncanonical.Stdout.Base64='Zh==';$noncanonical.Stdout.ByteLength=1
$result=Test-OfflineOutputEvidencePackage $noncanonical
Check-Evidence (-not $result.Verified -and $result.Reason -ceq 'Canonical Base64 required') 'NonzeroBase64PadBitsRefusedBeforeHash'
$claim=$package|ConvertTo-Json -Depth 8|ConvertFrom-Json;$claim.OriginDeclaration='another unverified claim'
$result=Test-OfflineOutputEvidencePackage $claim
Check-Evidence ($result.Verified -and -not $result.OriginVerified) 'OriginClaimIsNotAuthenticated'
foreach($encoding in @('utf-8','utf-16le')){
    $p=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName $encoding -OriginDeclaration 'synthetic'
    Check-Evidence ((Test-OfflineOutputEvidencePackage $p).Verified) ('EncodingMetadataRetained-'+$encoding)
}
foreach($code in @(1,-1)){
    $p=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes ([byte[]]@()) -ExitCode $code -EncodingName 'utf-8' -OriginDeclaration 'failed-child-fixture'
    $result=Test-OfflineOutputEvidencePackage $p
    Check-Evidence ($result.Verified -and $result.Classification -ceq 'refused' -and -not $result.Acceptance) ('FailedChildEvidenceCanBeRetained-'+$code)
}
$binary=[byte[]]@(0,255,13,10)
$p=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes $binary -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'invalid-text-fixture'
$result=Test-OfflineOutputEvidencePackage $p
Check-Evidence ($p.Stderr.Base64 -ceq 'AP8NCg==' -and $result.Verified -and $result.Classification -ceq 'refused' -and $binary[1] -eq 255) 'RefusedTextStillRetainsExactRawBytes'
$record='<Obj S="progress" RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN><MS><I64 N="SourceId">1</I64><PR N="Record"><AV>fixture</AV><AI>0</AI><Nil /><PI>-1</PI><PC>-1</PC><T>Completed</T><SR>-1</SR><SD> </SD></PR></MS></Obj>'
$xml='#< CLIXML'+"`n"+'<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04">'+$record+'</Objs>'
$utf8=New-Object Text.UTF8Encoding($false,$true)
$p=New-OfflineOutputEvidencePackage -StdoutBytes ([byte[]]@()) -StderrBytes ($utf8.GetBytes($xml)) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'progress-fixture'
$result=Test-OfflineOutputEvidencePackage $p
Check-Evidence ($result.Verified -and $result.Classification -ceq 'progress-only' -and $result.ProgressCount -eq 1) 'ClassificationRecomputedFromPreservedBytes'
$p.ProgressCount=2;Check-Evidence (-not (Test-OfflineOutputEvidencePackage $p).Verified) 'StoredProgressCountNotTrusted'
$maximum=New-Object byte[] 1048576
$p=New-OfflineOutputEvidencePackage -StdoutBytes $maximum -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'maximum-fixture'
Check-Evidence ((Test-OfflineOutputEvidencePackage $p).Verified -and $p.Stdout.ByteLength -eq 1048576) 'ExactPerStreamMaximumSupported'
foreach($kind in @('missing','oversized','encoding','exit','origin')){
    $refused=$false
    try{
        switch($kind){
            missing{New-OfflineOutputEvidencePackage -StdoutBytes $null -StderrBytes $empty -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'synthetic'|Out-Null}
            oversized{New-OfflineOutputEvidencePackage -StdoutBytes (New-Object byte[] 1048577) -StderrBytes $empty -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration 'synthetic'|Out-Null}
            encoding{New-OfflineOutputEvidencePackage -StdoutBytes $empty -StderrBytes $empty -ExitCode 0 -EncodingName 'guess' -OriginDeclaration 'synthetic'|Out-Null}
            exit{New-OfflineOutputEvidencePackage -StdoutBytes $empty -StderrBytes $empty -ExitCode '0' -EncodingName 'utf-8' -OriginDeclaration 'synthetic'|Out-Null}
            origin{New-OfflineOutputEvidencePackage -StdoutBytes $empty -StderrBytes $empty -ExitCode 0 -EncodingName 'utf-8' -OriginDeclaration ''|Out-Null}
        }
    }catch{$refused=$true}
    Check-Evidence $refused ('InvalidConstructionRefused-'+$kind)
}
foreach($invalid in @($null,'not an object',@{})){
    Check-Evidence (-not (Test-OfflineOutputEvidencePackage $invalid).Verified) 'NonPackageRefused'
}
Write-Output 'LIMIT: synthetic in-memory evidence objects only;no durable archive,live capture or origin authentication.'
