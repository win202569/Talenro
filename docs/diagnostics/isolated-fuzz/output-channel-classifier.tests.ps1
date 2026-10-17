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
