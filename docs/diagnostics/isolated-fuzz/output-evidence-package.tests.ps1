$ErrorActionPreference='Stop'
function Check-Evidence([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'output-evidence-package.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Evidence ($null -ne (Get-Command New-OfflineOutputEvidencePackage -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineEvidencePackageRequired'
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
