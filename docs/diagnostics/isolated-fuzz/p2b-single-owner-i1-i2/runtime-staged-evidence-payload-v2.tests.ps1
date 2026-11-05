Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$helper=Join-Path $PSScriptRoot runtime-staged-evidence-fixtures-v2.ps1
if(-not(Test-Path $helper)){throw 'PAYLOAD_RED: versioned decoder missing'}
. $helper
function Check($ok,$name){if(-not $ok){throw "PAYLOAD_FAIL: $name"};Write-Output "PASS: $name"}
function Reject($entry,$pattern,$name){
 $reason=$null
 try{$null=ConvertFrom-SEPayload $entry}catch{$reason=$_.Exception.Message}
 Check ($null -ne $reason -and $reason -match $pattern) $name
}
# These catch selecting only the newer field, coercing invalid types, ambiguous
# fallback, losing empty arrays, and failing to bind restored bytes to metadata.
$sha='BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD'
foreach($field in @('Base64','RawBase64')){
 $e=@{Length=3;SHA256=$sha};$e[$field]='YWJj'
 $b=ConvertFrom-SEPayload $e
 Check ($b -is [byte[]] -and [Convert]::ToHexString($b) -ceq '616263') "$field literal bytes"
 $empty=@{Length=0;SHA256='E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855'};$empty[$field]=''
 $b=ConvertFrom-SEPayload $empty
 Check ($b -is [byte[]] -and $b.Length -eq 0) "$field empty bytes"
}
Reject @{Length=3;SHA256=$sha} 'exactly one' 'missing field'
Reject @{Length=3;SHA256=$sha;Base64='YWJj';RawBase64='YWJj'} 'exactly one' 'ambiguous fields'
Reject @{Length=3;SHA256=$sha;Base64=$null} 'string' 'null field'
Reject @{Length=3;SHA256=$sha;RawBase64=123} 'string' 'non-string field'
Reject @{Length=3;SHA256=$sha;Base64='!'} 'base64' 'malformed encoding'
Reject @{Length=4;SHA256=$sha;Base64='YWJj'} 'bytes' 'length mismatch'
Reject @{Length=3;SHA256=('0'*64);RawBase64='YWJj'} 'bytes' 'hash mismatch'
# Real archive entries exercise both formats without writing fixtures or running a verifier.
$prior=Get-Content (Join-Path $PSScriptRoot portable-runtime-copy-transport-final.json) -Raw|ConvertFrom-Json -AsHashtable
$e=@($prior.Files|Where-Object Path -CEQ '.superpowers/sdd/2026-11-04-runtime-copy-transport-v1/stages.jsonl')[0]
$b=ConvertFrom-SEPayload $e
Check ($b.Length -eq 8920) 'real legacy stages'
$recent=Get-Content (Join-Path $PSScriptRoot portable-runtime-staged-integrity-review-v2.json) -Raw|ConvertFrom-Json -AsHashtable
foreach($e in $recent.Files){$null=ConvertFrom-SEPayload $e}
Check ($recent.Files.Count -eq 11) 'real current archive'
'PAYLOAD_V2_VERIFIED'
