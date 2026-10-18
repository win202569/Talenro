# PREPARED ONLY. Execution/compilation needs separately approved bounded PS7 host.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][ValidateSet('baseline','candidate')][string]$Variant,
    [Parameter(Mandatory)][switch]$ExecutionSeparatelyApproved
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
if(-not $ExecutionSeparatelyApproved){throw 'BLOCKED: execution not approved'}
if($PSVersionTable.PSEdition -cne 'Core'){throw 'BLOCKED: fresh PS7 required'}
if(Get-Command Assert-SupervisionPolicyToken -ErrorAction SilentlyContinue){throw 'BLOCKED: helper already loaded'}
if('TalenroTokenDirect.DescriptorSupport' -as [type]){throw 'BLOCKED: support already loaded'}
$manifest=Get-Content (Join-Path $PSScriptRoot 'source-pins.json') -Raw|ConvertFrom-Json
$selected=if($Variant -ceq 'baseline'){$manifest.Baseline}else{$manifest.Candidate}
$module=Join-Path $PSScriptRoot $selected.Path
if((Get-FileHash -LiteralPath $module -Algorithm SHA256).Hash -cne $selected.SHA256){throw 'BLOCKED: module hash'}
$support=Join-Path $PSScriptRoot 'TokenDescriptorSupport.cs'
if((Get-FileHash -LiteralPath $support -Algorithm SHA256).Hash -cne $manifest.SupportSHA256){throw 'BLOCKED: support hash'}
# Future execution only: never skip/substitute CodeProperty if compilation fails.
Add-Type -Path $support
. $module
$script:CallbackCount=0
function Direct-Type($Value){
    if($null -eq $Value){return $null}
    return [object].GetMethod('GetType').Invoke($Value.PSObject.BaseObject,$null)
}
function Fresh-Owner([string]$Text){
    if($Text.Length -eq 0){throw 'FIXTURE ERROR: decorated text must be nonempty'}
    $chars=$Text.ToCharArray()
    $fresh=[string]::new($chars)
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: wrapper constructor'}
    $wrapper=$ctor.Invoke([object[]]@($fresh))
    return [pscustomobject]@{Value=$wrapper}
}
function Check-Owner($Owner,[string]$Text,$Descriptor,$Payload){
    $field=$Owner.Value
    if((Direct-Type $field) -ne [string] -or -not [string]::Equals($field.PSObject.BaseObject,$Text,[StringComparison]::Ordinal)){
        throw 'FIXTURE ERROR: field type/value changed'
    }
    if($null -ne $Descriptor){
        $actual=$field.PSObject.Members[$Descriptor.Name]
        if($null -eq $actual -or $actual.MemberType -ne $Descriptor.MemberType -or
            -not [object]::ReferenceEquals($actual,$Descriptor)){throw 'FIXTURE ERROR: descriptor identity lost'}
        if($Descriptor.MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty -and
            -not [object]::ReferenceEquals($actual.Value,$Payload)){throw 'FIXTURE ERROR: payload identity lost'}
    }elseif(@($field.psextended.Members).Count -ne 0){throw 'FIXTURE ERROR: unexpected clean decoration'}
}
function Check-Call([string]$CaseLabel,$Owner,[AllowNull()][string[]]$Allowed,$Metadata,$Expected,$Descriptor=$null,$Payload=$null){
    $text=if((Direct-Type $Owner.Value) -eq [string]){$Owner.Value.PSObject.BaseObject}else{$null}
    if($null -ne $text){Check-Owner $Owner $text $Descriptor $Payload}
    $failure=$null;$output=@()
    try{$output=@(Assert-SupervisionPolicyToken $Owner.Value $Allowed -Metadata $Metadata)}
    catch{$failure=$_.Exception.Message}
    if($output.Count -ne 0 -or -not [string]::Equals($failure,$Expected,[StringComparison]::Ordinal)){
        throw "FAIL ${CaseLabel}: literal outcome/output mismatch"
    }
    if($null -ne $text){Check-Owner $Owner $text $Descriptor $Payload}
    if($script:CallbackCount -ne 0 -or [TalenroTokenDirect.DescriptorSupport]::Calls -ne 0){throw 'FAIL: callback executed'}
    "PASS $CaseLabel"
}
foreach($forward in @($false,$true)){
    $metadata=if($forward){New-SupervisionPolicyMetadata}else{$null}
    $mode=if($forward){'forwarded'}else{'fallback'}
    Check-Call "$mode-allowed" ([pscustomobject]@{Value='allow'}) @('allow') $metadata $null
    Check-Call "$mode-empty-allowed" ([pscustomobject]@{Value=''}) @('') $metadata $null
    Check-Call "$mode-empty-denied" ([pscustomobject]@{Value=''}) @('allow') $metadata 'Ordinal token required'
    Check-Call "$mode-unknown" ([pscustomobject]@{Value='unknown'}) @('allow') $metadata 'Ordinal token required'
    Check-Call "$mode-case" ([pscustomobject]@{Value='ALLOW'}) @('allow') $metadata 'Ordinal token required'
    Check-Call "$mode-empty-list" ([pscustomobject]@{Value='allow'}) @() $metadata 'Ordinal token required'
    Check-Call "$mode-null-list" ([pscustomobject]@{Value='allow'}) $null $metadata 'Ordinal token required'
    foreach($wrong in @($null,1,[long]1,$true,[double]1,([object[]]@('allow')),([pscustomobject]@{Text='allow'}))){
        Check-Call "$mode-wrong-type" ([pscustomobject]@{Value=$wrong}) @('allow') $metadata 'Exact string required'
    }
    foreach($kind in @('NoteProperty','ScriptProperty','CodeProperty','AliasProperty','ScriptMethod')){
        foreach($memberName in @('Metadata','Members','GetType','Length','Rank')){
            foreach($textValue in @('allow','unknown')){
                $owner=Fresh-Owner $textValue
                $clean=Fresh-Owner 'allow'
                $payload=[object[]]@('initial')
                $descriptor=switch($kind){
                    NoteProperty{[Management.Automation.PSNoteProperty]::new($memberName,$payload)}
                    ScriptProperty{[Management.Automation.PSScriptProperty]::new($memberName,{$script:CallbackCount++;throw 'callback executed'})}
                    CodeProperty{[Management.Automation.PSCodeProperty]::new($memberName,[TalenroTokenDirect.DescriptorSupport].GetMethod('Read'))}
                    AliasProperty{[Management.Automation.PSAliasProperty]::new($memberName,'Length')}
                    ScriptMethod{[Management.Automation.PSScriptMethod]::new($memberName,{$script:CallbackCount++;throw 'callback executed'})}
                }
                $owner.Value.PSObject.Members.Add($descriptor)
                # Read the stored descriptor, not its value or an executable getter.
                $stored=$owner.Value.PSObject.Members[$memberName]
                if($null -eq $stored -or $stored.MemberType -ne $descriptor.MemberType){throw 'FIXTURE ERROR: add did not survive'}
                Check-Call "$mode-clean-before" $clean @('allow') $metadata $null
                Check-Call "$mode-$kind-$memberName-$textValue" $owner @('allow') $metadata 'Decorated scalar' $stored $payload
                $payload[0]='mutated'
                Check-Owner $owner $textValue $stored $payload
                Check-Call "$mode-mutated-$kind-$memberName-$textValue" $owner @('allow') $metadata 'Decorated scalar' $stored $payload
                Check-Call "$mode-clean-after" $clean @('allow') $metadata $null
                $independent=Fresh-Owner $textValue
                Check-Owner $independent $textValue $null $null
            }
        }
    }
}
# Full API equality/refusal-index/result-ownership gates remain separate.
"PASS direct-suite-$Variant"
