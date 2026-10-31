param(
    [Parameter(Mandatory)][string]$PacketRoot,
    [Parameter(Mandatory)][ValidatePattern('^[A-F0-9]{64}$')][string]$SeedSHA256
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$rcRepo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$rcExpectedPacket=Join-Path $rcRepo '.superpowers/sdd/2026-10-31-offline-record-chain-diagnosis'
if([IO.Path]::GetFullPath($PacketRoot) -cne $rcExpectedPacket){throw 'Fixed packet required'}
$rcSeed=Join-Path $rcRepo '.superpowers/sdd/2026-10-31-offline-diagnostic-review-repair/seed.frozen.json'
if($SeedSHA256 -cne '9091825652E469001E51D8A510B12E0F84013C38509D740C649942DC2EAAFF6B' -or (Get-FileHash -LiteralPath $rcSeed).Hash -cne $SeedSHA256){throw 'Seed identity mismatch'}
if($PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Pinned engine required'}
$rcIdentity=[Security.Principal.WindowsIdentity]::GetCurrent()
if($rcIdentity.Name -cne 'DESKTOP-H4SVCLU\Lenovo' -or ([Security.Principal.WindowsPrincipal]::new($rcIdentity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Fixed non-admin context required'}
$rcWriter=Join-Path $rcRepo 'scripts/offline-diagnostics-v2/private/write-boundary.ps1'
$rcCapture=Join-Path $rcRepo 'docs/diagnostics/isolated-fuzz/p2b-review-repair/capture.ps1'
$rcTokens=$null;$rcParseErrors=$null
$rcAst=[Management.Automation.Language.Parser]::ParseFile($rcCapture,[ref]$rcTokens,[ref]$rcParseErrors)
if($rcParseErrors.Count -ne 0){throw 'Capture AST invalid'}
$rcCollectors=@($rcAst.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
if($rcCollectors.Count -ne 1){throw 'Unique collector required'}
. $rcWriter
. ([ScriptBlock]::Create($rcCollectors[0].Extent.Text))
$script:HostPrimary=$null
$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
$script:HostOverflow=$false
$rcNames=@('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Capture-Failure')
$rcOriginals=@{}
foreach($rcName in $rcNames){$rcOriginals[$rcName]=(Get-Command $rcName -CommandType Function -ErrorAction Stop).ScriptBlock}

# Observer state is passed as a local reference, not held in tested script vars.
$rcObserver=[pscustomobject]@{Stream=$null;Bytes=0;Sequence=0;PublishFailed=$false}
$rcExit=2
try{
    $rcObserver.Stream=[IO.File]::Open((Join-Path $PacketRoot trace.jsonl),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    function Write-RCDiagnosticStage($Observer,$Value){
        $rcBytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 15 -Compress)+"`n")
        if($Observer.Bytes+$rcBytes.Length -gt 1048576){$Observer.PublishFailed=$true;throw 'Trace byte limit exceeded'}
        try{$Observer.Stream.Write($rcBytes,0,$rcBytes.Length);$Observer.Stream.Flush();$Observer.Bytes+=$rcBytes.Length;$Observer.Sequence++}
        catch{$Observer.PublishFailed=$true;throw}
    }
    function Get-RCDiagnosticCommands($Names,$Originals){
        foreach($rcName in $Names){
            $rcCommand=Get-Command $rcName -CommandType Function -ErrorAction Stop
            [ordered]@{Name=$rcName;SameOriginal=[object]::ReferenceEquals($Originals[$rcName],$rcCommand.ScriptBlock);ModuleName=$rcCommand.ModuleName;ModuleGuid=$(if($null -eq $rcCommand.Module){$null}else{$rcCommand.Module.Guid.ToString()})}
        }
    }
    function Get-RCDiagnosticState {
        $rcPrimary=Get-Variable -Name HostPrimary -Scope Script -ErrorAction Ignore
        $rcSecondary=Get-Variable -Name HostSecondary -Scope Script -ErrorAction Ignore
        $rcSentinel=Get-Variable -Name RCDiagnosisSentinel_20261031 -Scope Script -ErrorAction Ignore
        [ordered]@{PrimaryPresent=($null -ne $rcPrimary);PrimaryNull=$(if($null -eq $rcPrimary){$null}else{$null -eq $rcPrimary.Value});SecondaryPresent=($null -ne $rcSecondary);SecondaryCount=$(if($null -eq $rcSecondary){$null}else{$rcSecondary.Value.Count});SentinelPresent=($null -ne $rcSentinel);SentinelValue=$(if($null -eq $rcSentinel){$null}else{$rcSentinel.Value})}
    }
    function Invoke-RCDiagnosticTopology($Observer,$Writer,$CollectorText,$Names,$Originals){
        $rcOuterState=Get-RCDiagnosticState
        Write-RCDiagnosticStage $Observer ([ordered]@{Sequence=$Observer.Sequence;Stage='outer-before';Commands=@(Get-RCDiagnosticCommands $Names $Originals);OuterState=$rcOuterState;ModuleState=$null})
        # Verbatim construction from frozen record-boundary.tests.ps1:16-21.
        $module=New-Module -ArgumentList $Writer,$CollectorText -ScriptBlock {
            param($WriterPath,$CollectorText)
            . $WriterPath
            . ([ScriptBlock]::Create($CollectorText))
            $script:OriginalFactory=(Get-Command New-ODRecordStream).ScriptBlock
        }
        # A module-bound read-only state reader; it calls no tested function.
        $rcModuleRead={
            $rcPrimary=Get-Variable -Name HostPrimary -Scope Script -ErrorAction Ignore
            $rcSecondary=Get-Variable -Name HostSecondary -Scope Script -ErrorAction Ignore
            $rcSentinel=Get-Variable -Name RCDiagnosisSentinel_20261031 -Scope Script -ErrorAction Ignore
            [ordered]@{PrimaryPresent=($null -ne $rcPrimary);PrimaryNull=$(if($null -eq $rcPrimary){$null}else{$null -eq $rcPrimary.Value});SecondaryPresent=($null -ne $rcSecondary);SecondaryCount=$(if($null -eq $rcSecondary){$null}else{$rcSecondary.Value.Count});SentinelPresent=($null -ne $rcSentinel);SentinelValue=$(if($null -eq $rcSentinel){$null}else{$rcSentinel.Value})}
        }
        $rcModuleState=& $module $rcModuleRead
        Write-RCDiagnosticStage $Observer ([ordered]@{Sequence=$Observer.Sequence;Stage='module-created';Commands=@(Get-RCDiagnosticCommands $Names $Originals);OuterState=(Get-RCDiagnosticState);ModuleState=$rcModuleState})
        $rcModuleObservation=& $module {
            param($Names,$Originals)
            $script:RCDiagnosisSentinel_20261031='scope-only-20261031'
            $rcCommands=@(foreach($rcName in $Names){
                $rcCommand=Get-Command $rcName -CommandType Function -ErrorAction Stop
                [ordered]@{Name=$rcName;SameOriginal=[object]::ReferenceEquals($Originals[$rcName],$rcCommand.ScriptBlock);ModuleName=$rcCommand.ModuleName;ModuleGuid=$(if($null -eq $rcCommand.Module){$null}else{$rcCommand.Module.Guid.ToString()})}
            })
            $rcPrimary=Get-Variable -Name HostPrimary -Scope Script -ErrorAction Ignore
            $rcSecondary=Get-Variable -Name HostSecondary -Scope Script -ErrorAction Ignore
            $rcSentinel=Get-Variable -Name RCDiagnosisSentinel_20261031 -Scope Script -ErrorAction Ignore
            [ordered]@{Commands=$rcCommands;State=[ordered]@{PrimaryPresent=($null -ne $rcPrimary);PrimaryNull=$(if($null -eq $rcPrimary){$null}else{$null -eq $rcPrimary.Value});SecondaryPresent=($null -ne $rcSecondary);SecondaryCount=$(if($null -eq $rcSecondary){$null}else{$rcSecondary.Value.Count});SentinelPresent=($null -ne $rcSentinel);SentinelValue=$(if($null -eq $rcSentinel){$null}else{$rcSentinel.Value})}}
        } $Names $Originals
        Write-RCDiagnosticStage $Observer ([ordered]@{Sequence=$Observer.Sequence;Stage='module-scriptblock';Commands=$rcModuleObservation.Commands;OuterState=(Get-RCDiagnosticState);ModuleState=$rcModuleObservation.State})
        $rcModuleState=& $module $rcModuleRead
        Write-RCDiagnosticStage $Observer ([ordered]@{Sequence=$Observer.Sequence;Stage='outer-after';Commands=@(Get-RCDiagnosticCommands $Names $Originals);OuterState=(Get-RCDiagnosticState);ModuleState=$rcModuleState})
        Write-RCDiagnosticStage $Observer ([ordered]@{Sequence=$Observer.Sequence;Stage='complete';Commands=@(Get-RCDiagnosticCommands $Names $Originals);OuterState=(Get-RCDiagnosticState);ModuleState=$rcModuleState})
    }
    Invoke-RCDiagnosticTopology $rcObserver $rcWriter $rcCollectors[0].Extent.Text $rcNames $rcOriginals
    $rcExit=0
}catch{
    $rcActualError=$_
    if($null -ne $rcObserver.Stream -and -not $rcObserver.PublishFailed){
        try{Write-RCDiagnosticStage $rcObserver ([ordered]@{Sequence=$rcObserver.Sequence;Stage='unexpected-error';Error=[ordered]@{Type=$rcActualError.Exception.GetType().FullName;Message=$rcActualError.Exception.Message;ErrorId=$rcActualError.FullyQualifiedErrorId;Position=$rcActualError.InvocationInfo.PositionMessage;Stack=$rcActualError.ScriptStackTrace}})}catch{}
    }
}finally{
    if($null -ne $rcObserver.Stream){try{$rcObserver.Stream.Dispose()}catch{$rcExit=2}}
}
exit $rcExit
