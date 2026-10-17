$ErrorActionPreference='Stop'
function Check-Output([bool]$Value,[string]$Name){if(-not $Value){throw ('ASSERTION FAILED: '+$Name)};Write-Output ('PASS: '+$Name)}
$helper=Join-Path $PSScriptRoot 'output-channel-classifier.ps1'
if(Test-Path -LiteralPath $helper){. $helper}
Check-Output ($null -ne (Get-Command Get-OfflineOutputClassification -CommandType Function -ErrorAction SilentlyContinue)) 'OfflineClassifierRequired'
$utf8=New-Object Text.UTF8Encoding($false,$true)
$utf16=New-Object Text.UnicodeEncoding($false,$false,$true)
$prefix='#< CLIXML'+"`r`n"+'<Objs Version="1.1.0.1" xmlns="http://schemas.microsoft.com/powershell/2004/04">'
$record='<Obj S="progress" RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN><MS><I64 N="SourceId">1</I64><PR N="Record"><AV>preparing module</AV><AI>0</AI><Nil /><PI>-1</PI><PC>-1</PC><T>Completed</T><SR>-1</SR><SD> </SD></PR></MS></Obj>'
$record2=$record.Replace('RefId="0"><TN RefId="0"><T>System.Management.Automation.PSCustomObject</T><T>System.Object</T></TN>','RefId="1"><TNRef RefId="0" />')
$text=$prefix+$record+'</Objs>'
# Break caught:legal XML may contain ignorable characters in protocol values.
# Literal expected labels/counts;independent SHA calculation also checks byte retention.
$script:classifierExactFailures=0
function Check-ClassifierExact([bool]$Value,[string]$Name){
    if($Value){Write-Output ('PASS: '+$Name)}else{$script:classifierExactFailures++;Write-Output ('FAIL: '+$Name)}
}
function Check-ClassifierExactResult($Actual,[byte[]]$FixtureBytes,[string]$Expected,[int]$Count,[string]$Name){
    $sha=[Security.Cryptography.SHA256]::Create()
    try{$digest=[BitConverter]::ToString($sha.ComputeHash($FixtureBytes)).Replace('-','')}finally{$sha.Dispose()}
    Check-ClassifierExact ([string]::Equals($Actual.Classification,$Expected,[StringComparison]::Ordinal) -and $Actual.ProgressCount -eq $Count -and -not $Actual.Acceptance -and -not $Actual.RuntimeApproved -and $Actual.Stderr.ByteLength -eq $FixtureBytes.Length -and [string]::Equals($Actual.Stderr.SHA256,$digest,[StringComparison]::Ordinal)) $Name
}
$exactXmlCases=[ordered]@{
    channelIgnorable=$text.Replace('S="progress"','S="pro&#173;gress"')
    sourceNameIgnorable=$text.Replace('N="SourceId"','N="Sou&#173;rceId"')
    recordNameIgnorable=$text.Replace('N="Record"','N="Rec&#173;ord"')
    completedIgnorable=$text.Replace('<T>Completed</T>','<T>Com&#173;pleted</T>')
    processingIgnorable=$text.Replace('<T>Completed</T>','<T>Proce&#173;ssing</T>')
    customTypeIgnorable=$text.Replace('System.Management.Automation.PSCustomObject','System.Management.Automation.PSCustomOb&#173;ject')
    baseTypeIgnorable=$text.Replace('System.Object','System.Ob&#173;ject')
    versionIgnorable=$text.Replace('Version="1.1.0.1"','Version="1.1.0.&#173;1"')
    namespaceIgnorable=$text.Replace('http://schemas.microsoft.com/powershell/2004/04','http://schemas.microsoft.com/powershell/2004/0&#173;4')
    channelCase=$text.Replace('S="progress"','S="Progress"')
    progressCase=$text.Replace('<T>Completed</T>','<T>completed</T>')
    nonProgressIgnorable=$text.Replace('S="progress"','S="e&#173;rror"')
}
foreach($exactEncoding in @('utf-8','utf-16le')){
    $exactCodec=if($exactEncoding -eq 'utf-8'){$utf8}else{$utf16}
    foreach($exactName in $exactXmlCases.Keys){
        $fixtureBytes=$exactCodec.GetBytes($exactXmlCases[$exactName])
        $exactResult=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $fixtureBytes -ExitCode 0 -EncodingName $exactEncoding
        Check-ClassifierExactResult $exactResult $fixtureBytes 'refused' 0 ('ExactClassifierRefusal-'+$exactName+'-'+$exactEncoding)
    }
    foreach($exactState in @('Completed','Processing')){
        $display=([string][char]0x00E9)+'e'+[char]0x0301+[char]0x00AD+[char]0x6A21+[char]0x5757
        $displayXml=$text.Replace('preparing module',$display).Replace('<T>Completed</T>',('<T>'+$exactState+'</T>'))
        $fixtureBytes=$exactCodec.GetBytes($displayXml);$originalFixture=[Convert]::ToBase64String($fixtureBytes)
        $exactResult=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $fixtureBytes -ExitCode 0 -EncodingName $exactEncoding
        Check-ClassifierExactResult $exactResult $fixtureBytes 'progress-only' 1 ('ExactUnicodeDisplayAndStatePreserved-'+$exactState+'-'+$exactEncoding)
        if(-not [string]::Equals([Convert]::ToBase64String($fixtureBytes),$originalFixture,[StringComparison]::Ordinal)){throw 'Classifier changed fixture bytes'}
    }
}
$nul=[string][char]0;$softHyphen=[string][char]0x00AD
foreach($encodingCase in @(
    @{Token=('utf'+$nul+'-8');Codec=$utf8;Name='utf8-null'},
    @{Token=('utf-16'+$nul+'le');Codec=$utf16;Name='utf16-null'},
    @{Token=('utf'+$softHyphen+'-8');Codec=$utf8;Name='utf8-ignorable'},
    @{Token=('utf-16'+$softHyphen+'le');Codec=$utf16;Name='utf16-ignorable'},
    @{Token='UTF-8';Codec=$utf8;Name='utf8-case'},
    @{Token='UTF-16LE';Codec=$utf16;Name='utf16-case'})){
    $fixtureBytes=$encodingCase.Codec.GetBytes($text)
    $exactResult=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $fixtureBytes -ExitCode 0 -EncodingName $encodingCase.Token
    Check-ClassifierExactResult $exactResult $fixtureBytes 'refused' 0 ('ExactClassifierEncodingRefusal-'+$encodingCase.Name)
}
Check-Output ($script:classifierExactFailures -eq 0) 'ExactClassifierMarkersRequired'
$bytes=$utf8.GetBytes($text);$copy=[Convert]::ToBase64String($bytes)
$empty=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@(97,98,99)) -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8'
Check-Output ($empty.Classification -ceq 'empty' -and -not $empty.Acceptance -and -not $empty.RuntimeApproved) 'EmptyIsClassificationNotAcceptance'
Check-Output ($empty.Stdout.SHA256 -ceq 'BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD' -and $empty.Stdout.ByteLength -eq 3 -and $empty.Stderr.SHA256 -ceq 'E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855' -and $empty.Stderr.ByteLength -eq 0) 'OriginalByteHashesAndLengths'
$progress=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $bytes -ExitCode 0 -EncodingName 'utf-8'
Check-Output ($progress.Classification -ceq 'progress-only' -and $progress.ProgressCount -eq 1 -and -not $progress.Acceptance -and -not $progress.RuntimeApproved) 'ExplicitProgressOnlyRecognized'
Check-Output ([Convert]::ToBase64String($bytes) -ceq $copy) 'CallerBytesUnchanged'
$multiple=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes ($utf8.GetBytes($prefix+$record+$record2+'</Objs>')) -ExitCode 0 -EncodingName 'utf-8'
Check-Output ($multiple.Classification -ceq 'progress-only' -and $multiple.ProgressCount -eq 2) 'DefinedTypeReferenceRecognized'
$unicodeText=$text.Replace('preparing module',([string][char]0x6A21+[char]0x5757))
foreach($encoding in @('utf-8','utf-16le')){
    $encoder=if($encoding -ceq 'utf-8'){$utf8}else{$utf16}
    $encoded=$encoder.GetBytes($unicodeText)
    $result=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $encoded -ExitCode 0 -EncodingName $encoding
    Check-Output ($result.Classification -ceq 'progress-only' -and $result.EncodingName -ceq $encoding -and $result.Stderr.ByteLength -eq $encoded.Length) ('ExplicitStrictEncoding-'+$encoding)
}
$cases=[ordered]@{
    terminalLfActivity=$text.Replace('<AI>0</AI>','<AI>0&#10;</AI>')
    terminalLfParent=$text.Replace('<PI>-1</PI>','<PI>-1&#10;</PI>')
    terminalLfPercent=$text.Replace('<PC>-1</PC>','<PC>-1&#10;</PC>')
    terminalLfSeconds=$text.Replace('<SR>-1</SR>','<SR>-1&#10;</SR>')
    terminalLfSource=$text.Replace('<I64 N="SourceId">1</I64>','<I64 N="SourceId">1&#10;</I64>')
    terminalLfObjectId=$text.Replace('<Obj S="progress" RefId="0">','<Obj S="progress" RefId="0&#10;">')
    terminalLfTypeId=$text.Replace('<TN RefId="0">','<TN RefId="0&#10;">')
    terminalLfTypeReference=$prefix+$record.Replace('<TN RefId="0">','<TN RefId="0&#10;">')+$record2.Replace('<TNRef RefId="0"','<TNRef RefId="0&#10;"')+'</Objs>'
    error=$text.Replace('S="progress"','S="error"')
    warning=$text.Replace('S="progress"','S="warning"')
    unknown=$text.Replace('S="progress"','S="other"')
    mixed=$prefix+$record+$record.Replace('S="progress"','S="error"').Replace('RefId="0"','RefId="2"')+'</Objs>'
    truncated=$text.Substring(0,$text.Length-4)
    trailing=$text+'garbage'
    secondRoot=$text+'<Objs />'
    wrongNamespace=$text.Replace('http://schemas.microsoft.com/powershell/2004/04','urn:other')
    wrongVersion=$text.Replace('Version="1.1.0.1"','Version="2"')
    unexpectedAttribute=$text.Replace('<MS>','<MS Extra="x">')
    hiddenElement=$text.Replace('<Nil />','<Nil /><S>hidden</S>')
    hiddenText=$text.Replace('<MS>','<MS>hidden')
    comment=$text.Replace('<MS>','<MS><!-- hidden -->')
    instruction=$text.Replace('<MS>','<MS><?hidden data?>')
    wrongType=$text.Replace('System.Management.Automation.PSCustomObject','System.Management.Automation.ErrorRecord')
    unknownReference=$prefix+$record2.Replace('TNRef RefId="0"','TNRef RefId="7"')+'</Objs>'
    duplicateObject=$prefix+$record+$record+'</Objs>'
    invalidInteger=$text.Replace('<AI>0</AI>','<AI>no</AI>')
    invalidPercent=$text.Replace('<PC>-1</PC>','<PC>101</PC>')
    nestedText=$text.Replace('preparing module','<S>hidden</S>')
    dtd='#< CLIXML'+"`n"+'<!DOCTYPE Objs [<!ENTITY ex SYSTEM "file:///never-read">]>'+($text.Substring($text.IndexOf('<Objs')))
    plain='a real error message'
    whitespace=' '
    noHeader=$text.Substring($text.IndexOf('<Objs'))
    emptyEnvelope=$prefix+'</Objs>'
}
foreach($name in $cases.Keys){
    $result=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes ($utf8.GetBytes($cases[$name])) -ExitCode 0 -EncodingName 'utf-8'
    Check-Output ($result.Classification -ceq 'refused' -and -not $result.Acceptance -and -not $result.RuntimeApproved -and $result.Stderr.ByteLength -gt 0 -and $result.Stderr.SHA256 -match '^[A-F0-9]{64}$') ('UnsafeOutputRefused-'+$name)
}
foreach($code in @(1,-1,'0',$true)){
    $result=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $bytes -ExitCode $code -EncodingName 'utf-8'
    Check-Output ($result.Classification -ceq 'refused') ('NonzeroOrNonIntegerExitRefused-'+[string]$code)
}
foreach($case in @(@([byte[]]@(255),'utf-8'),@([byte[]]@(65),'utf-16le'))){
    $result=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $case[0] -ExitCode 0 -EncodingName $case[1]
    Check-Output ($result.Classification -ceq 'refused') ('MalformedEncodingRefused-'+$case[1])
}
$badOut=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@(255)) -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8'
Check-Output ($badOut.Classification -ceq 'refused') 'MalformedStdoutAlsoRefused'
$badEncoding=Get-OfflineOutputClassification -StdoutBytes ([byte[]]@()) -StderrBytes $bytes -ExitCode 0 -EncodingName 'guess'
Check-Output ($badEncoding.Classification -ceq 'refused') 'UnknownEncodingRefused'
foreach($stream in @('stdout','stderr')){
    $large=New-Object byte[] 1048577;$small=[byte[]]@()
    $out=$small;$err=$small
    if($stream -ceq 'stdout'){$out=$large}else{$err=$large}
    $result=Get-OfflineOutputClassification -StdoutBytes $out -StderrBytes $err -ExitCode 0 -EncodingName 'utf-8'
    Check-Output ($result.Classification -ceq 'refused') ('OversizedStreamRefused-'+$stream)
}
$nullResult=Get-OfflineOutputClassification -StdoutBytes $null -StderrBytes ([byte[]]@()) -ExitCode 0 -EncodingName 'utf-8'
Check-Output ($nullResult.Classification -ceq 'refused') 'MissingBytesNotEmptyStream'
Write-Output 'LIMIT: synthetic byte fixtures only;no process launcher,source loader,OS adapter or real diagnostic invoked.'
