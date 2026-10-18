# PREPARED, NOT RUN: direct integer-helper contracts, not performance evidence.
# Catches coercion, lost member inspection, callback execution and error reordering.
# Standalone only; not added to the existing full-entry output inventory.
$ErrorActionPreference='Stop'
$script:integerCase='Setup'
$script:integerCallbacks=0
$script:integerTypeMethod=[object].GetMethod('GetType')
$script:integerMembersGetter=[Management.Automation.PSMemberSet].GetProperty('Members')
function Integer-Need([bool]$Value,[string]$Detail){
    if(-not $Value){throw ('ASSERTION FAILED: P1Integer'+$script:integerCase+' '+$Detail)}
}
function Integer-Case([string]$Name,[scriptblock]$Body){
    $script:integerCase=$Name;& $Body;Write-Output ('PASS: P1Integer'+$Name)
}
function Integer-Type($Value){
    if($null -eq $Value){return $null}
    return $script:integerTypeMethod.Invoke($Value.PSObject.BaseObject,$null)
}
function Integer-Owned($Value){
    # Same owned-box construction as the established ownership fixtures.
    $type=Integer-Type $Value
    if($type -eq [int] -or $type -eq [long]){
        $storage=[Array]::CreateInstance($type,1)
        $storage.SetValue($Value.PSObject.BaseObject,0);$fresh=$storage.GetValue(0)
    }elseif($type -eq [string]){$fresh=[string]::Copy($Value.PSObject.BaseObject)}
    else{throw 'FIXTURE ERROR: unsupported owned value'}
    $ctor=[Management.Automation.PSObject].GetConstructor([type[]]@([object]))
    if($null -eq $ctor -or -not $ctor.IsPublic){throw 'FIXTURE ERROR: object constructor unavailable'}
    $parameters=@($ctor.GetParameters())
    if($parameters.Count -ne 1 -or $parameters[0].ParameterType -ne [object]){throw 'FIXTURE ERROR: object constructor signature'}
    return ,($ctor.Invoke([object[]]@($fresh)))
}
function Integer-DecorationPresent($Value,[string]$Name,[string]$Kind){
    $member=$Value.PSObject.Members[$Name]
    if($null -eq $member -or $member.MemberType.ToString() -cne $Kind){throw 'FIXTURE ERROR: decoration lost'}
    # SDK getter, not an input-shadowable .Members lookup on the view.
    $view=$Value.psextended
    $entries=@($script:integerMembersGetter.GetValue($view.PSObject.BaseObject,$null))
    if($entries.Count -ne 1 -or $entries[0].Name -cne $Name -or $entries[0].MemberType.ToString() -cne $Kind){throw 'FIXTURE ERROR: extended member inventory'}
}
function Integer-Decorate($Value,[string]$Name,[ValidateSet('NoteProperty','AliasProperty','ScriptProperty','ScriptMethod')][string]$Kind){
    $typeBefore=Integer-Type $Value
    switch($Kind){
        NoteProperty {Add-Member -InputObject $Value -MemberType NoteProperty -Name $Name -Value ([object[]]@('owned')) -Force}
        AliasProperty {Add-Member -InputObject $Value -MemberType AliasProperty -Name $Name -Value 'MissingTarget' -Force}
        ScriptProperty {Add-Member -InputObject $Value -MemberType ScriptProperty -Name $Name -Value {$script:integerCallbacks++;throw 'FORBIDDEN INTEGER GETTER'} -Force}
        ScriptMethod {Add-Member -InputObject $Value -MemberType ScriptMethod -Name $Name -Value {$script:integerCallbacks++;throw 'FORBIDDEN INTEGER METHOD'} -Force}
    }
    if((Integer-Type $Value) -ne $typeBefore){throw 'FIXTURE ERROR: base type changed'}
    Integer-DecorationPresent $Value $Name $Kind
}
function Integer-Arguments($Value,[hashtable]$Options){
    $arguments=@{Value=$Value}
    foreach($key in $Options.Keys){$arguments[$key]=$Options[$key]}
    return $arguments
}
function Integer-Accepts($Value,[hashtable]$Options=@{}){
    $arguments=Integer-Arguments $Value $Options
    $output=@(Assert-SupervisionPolicyInteger @arguments)
    Integer-Need ($output.Count -eq 0) 'unexpected success output'
}
function Integer-Refuses($Value,[string]$Reason,[hashtable]$Options=@{}){
    $arguments=Integer-Arguments $Value $Options;$caught=$null
    $output=@(try{Assert-SupervisionPolicyInteger @arguments}catch{$caught=$_})
    Integer-Need ($null -ne $caught) 'missing refusal'
    Integer-Need ($caught.Exception.Message -ceq $Reason) 'wrong refusal category/order'
    Integer-Need ($output.Count -eq 0) 'unexpected refusal output'
}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
    $principal=[Security.Principal.WindowsPrincipal]::new($identity)
    $administrator=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$administrator)
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    if($administrator){throw 'Ordinary token required'}
    . (Join-Path $PSScriptRoot 'offline-supervision-policy.ps1')
    Integer-Case ApiRequired {
        Integer-Need ($null -ne (Get-Command Assert-SupervisionPolicyInteger -CommandType Function -ErrorAction SilentlyContinue)) 'integer helper missing'
        Integer-Need ($null -ne (Get-Command New-SupervisionPolicyMetadata -CommandType Function -ErrorAction SilentlyContinue)) 'metadata factory missing'
    }
    # Mutant: coerce Value to long, reject zero, change defaults, or emit output.
    Integer-Case PlainDefaults {
        foreach($value in @([int]0,[long]0,[int]1,[long]1,[int]::MaxValue,[long]::MaxValue)){Integer-Accepts $value}
        foreach($value in @([int]-1,[long]-1)){Integer-Refuses $value 'Integer bound'}
    }
    Integer-Case OwnedMetadataAndPositional {
        $context=New-SupervisionPolicyMetadata
        foreach($value in @([int]0,[long]0,[int]1,[long]::MaxValue)){Integer-Accepts $value @{Metadata=$context}}
        $output=@(Assert-SupervisionPolicyInteger ([int]-2) -2 2)
        Integer-Need ($output.Count -eq 0) 'positional defaults output'
        $output=@(Assert-SupervisionPolicyInteger ([long]2) -2 2 $context)
        Integer-Need ($output.Count -eq 0) 'positional metadata output'
    }
    # Mutant: drop either bound or turn an inclusive endpoint into exclusive.
    Integer-Case SignedBounds {
        foreach($explicit in @($false,$true)){
            $options=@{Min=-2;Max=2};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($value in @([int]-2,[long]-2,[int]0,[long]0,[int]2,[long]2)){Integer-Accepts $value $options}
            foreach($value in @([int]-3,[long]-3,[int]3,[long]3)){Integer-Refuses $value 'Integer bound' $options}
            $options.Min=[long]::MinValue;$options.Max=[long]::MaxValue
            foreach($value in @([int]::MinValue,[int]::MaxValue,[long]::MinValue,[long]::MaxValue)){Integer-Accepts $value $options}
        }
    }
    Integer-Case InvalidTypes {
        foreach($explicit in @($false,$true)){
            $options=@{};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($value in @($null,$true,$false,'0',[double]0,[single]0,[decimal]0,[uint32]0,[uint64]0,[int16]0,[byte]0,[char]'0',[DayOfWeek]::Sunday,[pscustomobject]@{})){
                Integer-Refuses $value 'Exact integer required' $options
            }
        }
    }
    # Mutant: call user-shadowable GetType/ToString instead of fixed SDK metadata.
    Integer-Case WrongTypeNoCallbacks {
        $script:integerCallbacks=0;$value=[pscustomobject]@{}
        Add-Member -InputObject $value -MemberType ScriptMethod -Name GetType -Value {$script:integerCallbacks++;return [int]} -Force
        Add-Member -InputObject $value -MemberType ScriptMethod -Name ToString -Value {$script:integerCallbacks++;return '0'} -Force
        Integer-Need ($value.PSObject.Members['GetType'].MemberType -eq [Management.Automation.PSMemberTypes]::ScriptMethod) 'shadow method fixture'
        Integer-Refuses $value 'Exact integer required'
        Integer-Refuses $value 'Exact integer required' @{Metadata=(New-SupervisionPolicyMetadata)}
        Integer-Need ($script:integerCallbacks -eq 0) 'wrong-type callback executed'
    }
    # Mutant: remove passive inspection and rely on earlier Shape checks.
    Integer-Case NoteDecorations {
        foreach($explicit in @($false,$true)){
            $options=@{};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($seed in @([int]0,[long]0)){
                foreach($name in @('Probe','Members','Length','Rank')){
                    $value=Integer-Owned $seed;Integer-Decorate $value $name NoteProperty
                    Integer-Refuses $value 'Decorated scalar' $options
                    Integer-DecorationPresent $value $name NoteProperty
                }
            }
        }
    }
    Integer-Case AliasDecorations {
        foreach($explicit in @($false,$true)){
            $options=@{};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($seed in @([int]0,[long]0)){
                foreach($name in @('Probe','Members','Length','Rank')){
                    $value=Integer-Owned $seed;Integer-Decorate $value $name AliasProperty
                    Integer-Refuses $value 'Decorated scalar' $options
                    Integer-DecorationPresent $value $name AliasProperty
                }
            }
        }
    }
    Integer-Case ExecutableDecorations {
        $script:integerCallbacks=0
        foreach($explicit in @($false,$true)){
            $options=@{};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($seed in @([int]0,[long]0)){
                foreach($kind in @('ScriptProperty','ScriptMethod')){
                    foreach($name in @('Probe','Members','GetType','Length','Rank')){
                        $value=Integer-Owned $seed;Integer-Decorate $value $name $kind
                        Integer-Refuses $value 'Decorated scalar' $options
                        Integer-DecorationPresent $value $name $kind
                        Integer-Need ($script:integerCallbacks -eq 0) 'integer descriptor executed'
                    }
                }
            }
        }
    }
    # Mutants: run range comparison before decoration, or decoration before type.
    Integer-Case DecorationBeforeBounds {
        foreach($seed in @([int]-1,[long]-1)){
            $value=Integer-Owned $seed;Integer-Decorate $value Probe NoteProperty
            Integer-Refuses $value 'Decorated scalar'
            Integer-Refuses $value 'Decorated scalar' @{Min=1;Max=2;Metadata=(New-SupervisionPolicyMetadata)}
        }
    }
    Integer-Case TypeBeforeDecoration {
        $script:integerCallbacks=0;$value=Integer-Owned '0'
        Integer-Decorate $value Probe ScriptProperty
        Integer-Refuses $value 'Exact integer required'
        Integer-Refuses $value 'Exact integer required' @{Metadata=(New-SupervisionPolicyMetadata)}
        Integer-Need ($script:integerCallbacks -eq 0) 'wrong-type getter executed'
    }
    # Mutant: cache clean acceptance across calls or retain a previous refusal.
    Integer-Case DecorationAfterClean {
        foreach($explicit in @($false,$true)){
            $options=@{};if($explicit){$options.Metadata=New-SupervisionPolicyMetadata}
            foreach($seed in @([int]0,[long]0)){
                $value=Integer-Owned $seed;Integer-Accepts $value $options
                Integer-Decorate $value Probe NoteProperty
                Integer-Refuses $value 'Decorated scalar' $options
                Integer-Refuses $value 'Decorated scalar' $options
                $value.PSObject.Members.Remove('Probe')
                $view=$value.psextended
                Integer-Need (@($script:integerMembersGetter.GetValue($view.PSObject.BaseObject,$null)).Count -eq 0) 'owned member removal fixture'
                Integer-Accepts $value $options
            }
        }
    }
    Integer-Case CleanAfterRefusal {
        $context=New-SupervisionPolicyMetadata
        $bad=Integer-Owned ([long]0);Integer-Decorate $bad Members NoteProperty
        Integer-Refuses $bad 'Decorated scalar' @{Metadata=$context}
        $plain=Integer-Owned ([long]0)
        Integer-Accepts $plain @{Metadata=$context};Integer-Accepts $plain
        Integer-Need ($null -eq $plain.PSObject.Members['Members']) 'separate owned value contaminated'
        Integer-DecorationPresent $bad Members NoteProperty
    }
    Integer-Case Complete {Integer-Need $true 'complete'}
    Write-Output 'LIMIT: direct integer-helper semantics only;no performance,full-regression or runtime acceptance.'
}catch{[Console]::Out.WriteLine('INTEGER-ERROR: '+$_.Exception.Message);exit 1}
