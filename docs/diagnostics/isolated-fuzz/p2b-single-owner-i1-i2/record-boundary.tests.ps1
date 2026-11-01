# Actual production writer functions run in isolated private test scopes.
function Invoke-ODWriterTests {
    # Availability is the entire RED computation. Do not compile or load a
    # fixture/private module before this conjunctive guard succeeds.
    $runtime=Join-Path $PSScriptRoot '../../../../scripts/offline-diagnostics-v4/private'
    if($null -eq (Get-Command New-ODRecordStream -CommandType Function -ErrorAction Ignore)){
        . (Join-Path $runtime write-boundary.ps1)
        . (Join-Path $runtime record-owner.ps1)
    }
    $available=$true
    foreach($name in @('New-ODRecordStream','Write-ODCaptureRecord','Write-ODProfileIntentRecord','Register-ODRecordOutcome','Import-ODProfileFailure')){
        if($null -eq (Get-Command $name -CommandType Function -ErrorAction Ignore)){$available=$false}
    }
    if(-not $available){
        $groups=@(foreach($name in $oracle.WriterGroupNames){[ordered]@{Name=$name;Status=$(if($name -ceq 'OwnerAPIAvailable'){'FAIL'}else{'NotRun'})}})
        return [pscustomobject][ordered]@{Schema=1;PASS=0;FAIL=1;NotRun=25;ExitCode=1;Groups=$groups}
    }
    LS-Result OwnerAPIAvailable $true ([ordered]@{Available=$true})
    . (Join-Path $PSScriptRoot case-module.ps1)
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $PSScriptRoot fault-stream.cs))) -Language CSharp -ErrorAction Stop
    $source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot capture.ps1))
    $tokens=$null;$parseErrors=$null
    $ast=[Management.Automation.Language.Parser]::ParseInput($source,[ref]$tokens,[ref]$parseErrors)
    LS-Need ($parseErrors.Count -eq 0) 'capture source AST'
    $collector=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
    LS-Need ($collector.Count -eq 1) 'fixed collector definition'
    $writer=Join-Path $repo 'scripts/offline-diagnostics-v4/private/write-boundary.ps1'
    $outer=(Get-Command New-ODRecordStream).ScriptBlock
    foreach($definition in $oracle.WriterCases){
        $root=Join-Path $packet ('unit/'+$Grant+'/'+$definition.Boundary+'/'+$definition.Case)
        LS-Need (-not (Test-Path -LiteralPath $root)) 'unit case no retry'
        $null=[IO.Directory]::CreateDirectory($root)
        $module=New-ODCaseModule $writer $collector[0].Extent.Text
        $script:CaseModules.Add([pscustomobject]@{Boundary=$definition.Boundary;Case=$definition.Case;Module=$module})
        Observe-ODRecorderGuard $script:RecorderGuard Created $definition.Name
        try{
        if(($module.ExportedFunctions.Count+$module.ExportedAliases.Count+$module.ExportedVariables.Count+$module.ExportedCmdlets.Count) -ne 0){throw 'ASSERT: no case exports'}
        & $module {
            param($Boundary,$Case,$Root)
            $script:HostPrimary=$null;$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new();$script:HostOverflow=$false
            $script:OpenCount=0;$script:Wrapper=$null;$script:Mode=$Case;$script:CaseRoot=$Root
            $path=Join-Path $Root record.bin
            [byte[]]$payload=@(65,66,0,255)
            if($Case -ceq 'OpenFailureOnce'){[IO.File]::WriteAllBytes($path,[byte[]]@(80,82,73,79,82))}
            function script:New-ODRecordStream([string]$Path){
                $script:OpenCount++
                $stream=& $script:OriginalFactory $Path
                if($script:Mode -ceq 'OpenFailureOnce'){return $stream}
                $mode=switch -CaseSensitive ($script:Mode){'WriteDisposeFirstCause'{'WriteDispose'}'FlushDisposeFirstCause'{'FlushDispose'}'DisposeOnly'{'DisposeOnly'}'SuccessOnce'{'Success'}}
                $script:Wrapper=[ODRepairFixture.FaultStream]::new($stream,$mode,(Join-Path $script:CaseRoot absent.txt),(Join-Path $script:CaseRoot closed.bin))
                return $script:Wrapper
            }
        } $definition.Boundary $definition.Case $root
        Observe-ODRecorderGuard $script:RecorderGuard Replaced $definition.Name
        $result=& $module {
            param($Boundary,$Case,$Root)
            $path=Join-Path $Root record.bin
            [byte[]]$payload=@(65,66,0,255)
            $outcome=$null
            # Factory replacement stays inside this explicitly private module.
            if($Boundary -ceq 'Capture'){$outcome=Write-ODCaptureRecord $path $payload}
            else{$outcome=Write-ODProfileIntentRecord $path $payload}
            $consumption=[pscustomobject]@{Consumed=$false}
            Register-ODRecordOutcome $outcome $consumption
            $all=@();if($null -ne $script:HostPrimary){$all+=,$script:HostPrimary};$all+=@($script:HostSecondary.ToArray())
            $wanted=if($Case -cin @('WriteDisposeFirstCause','FlushDisposeFirstCause')){'ObjectDisposedException'}elseif($Case -ceq 'DisposeOnly'){'FileNotFoundException'}elseif($Case -ceq 'OpenFailureOnce'){'IOException'}else{''}
            $first=@($all | Where-Object {$_.FullyQualifiedErrorId -ceq $wanted})
            $dispose=@($all | Where-Object {$_.FullyQualifiedErrorId -ceq 'FileNotFoundException'})
            if($Case -cne 'SuccessOnce' -and $first.Count -ne 1){throw 'ASSERT: first real error retained once'}
            $isFirst=if($Case -ceq 'SuccessOnce'){$null -eq $outcome.PrimaryError}else{[object]::ReferenceEquals($first[0],$outcome.PrimaryError)}
            $primaryFirst=if($Case -ceq 'SuccessOnce'){$null -eq $script:HostPrimary}else{[object]::ReferenceEquals($first[0],$script:HostPrimary)}
            $compound=$Case -cin @('WriteDisposeFirstCause','FlushDisposeFirstCause')
            $secondaryDispose=if($compound){$dispose.Count -eq 1 -and $script:HostSecondary.Count -eq 1 -and [object]::ReferenceEquals($script:HostSecondary[0],$dispose[0])}else{$script:HostSecondary.Count -eq 0}
            $bytesExact=$Case -ceq 'SuccessOnce' -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($path),$payload)
            $unchanged=$Case -ceq 'OpenFailureOnce' -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($path),[byte[]]@(80,82,73,79,82))
            [pscustomobject][ordered]@{
                Boundary=$Boundary;Case=$Case;Outcome=$outcome;Consumed=$consumption.Consumed;Primary=$script:HostPrimary;Secondary=$script:HostSecondary.ToArray()
                OpenCalls=$script:OpenCount;WriteCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.WriteCalls})
                FlushCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.FlushCalls});DisposeCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.DisposeCalls})
                OutcomePrimaryIsFirst=$isFirst;CollectorPrimaryIsFirst=$primaryFirst;SecondaryIsDispose=$secondaryDispose
                BytesExact=$bytesExact;ExistingUnchanged=$unchanged;Overflow=$script:HostOverflow
            }
        } $definition.Boundary $definition.Case $root
        }finally{Observe-ODRecorderGuard $script:RecorderGuard Finally $definition.Name}
        $same=[object]::ReferenceEquals($outer,(Get-Command New-ODRecordStream).ScriptBlock)
        LS-Need $same 'unit factory cannot change capture factory'
        $facts=[ordered]@{
            Boundary=$result.Boundary;Case=$result.Case;Succeeded=$result.Outcome.Succeeded;OutcomePrimaryIsFirst=$result.OutcomePrimaryIsFirst;CollectorPrimaryIsFirst=$result.CollectorPrimaryIsFirst;SecondaryIsDispose=$result.SecondaryIsDispose
            OpenCalls=$result.OpenCalls;WriteCalls=$result.WriteCalls;FlushCalls=$result.FlushCalls;DisposeCalls=$result.DisposeCalls
            SecondaryCount=$result.Secondary.Count;BytesExact=$result.BytesExact;ExistingUnchanged=$result.ExistingUnchanged;FactoryIsolated=$same
        }
        LS-Need ($result.Consumed -and $result.Outcome -is [Management.Automation.PSCustomObject] -and $result.Outcome.Succeeded -is [bool] -and $result.Outcome.SecondaryErrors -is [Management.Automation.ErrorRecord[]]) 'live typed outcome consumed'
        LS-Need ((@($result.Outcome.PSObject.Properties.Name) -join '|') -ceq 'Succeeded|PrimaryError|SecondaryErrors') 'ordered outcome shape'
        LS-Details RecordCase ([ordered]@{Boundary=$result.Boundary;Case=$result.Case;Outcome=$result.Outcome;Consumed=$result.Consumed;Primary=$result.Primary;Secondary=$result.Secondary;OpenCalls=$result.OpenCalls;WriteCalls=$result.WriteCalls;FlushCalls=$result.FlushCalls;DisposeCalls=$result.DisposeCalls;BytesExact=$result.BytesExact;ExistingUnchanged=$result.ExistingUnchanged;FactoryIsolated=$same;Overflow=$result.Overflow})
        $expected=@($oracle.Grants['writer-green'].Groups | Where-Object Name -CEQ $definition.Name)[0].Facts
        LS-Result $definition.Name ((ConvertTo-Json -InputObject $facts -Depth 10 -Compress) -ceq (ConvertTo-Json -InputObject $expected -Depth 10 -Compress)) $facts
        $module=$null
    }
    Invoke-ODOwnerControls $writer $collector[0].Extent.Text $source
}

# A deliberate control-only scope uses the actual collector/owners and capture
# function ASTs. Synthetic ErrorRecord construction appears in exactly one
# labelled contract control; all other causes come from real resource calls.
function Invoke-ODOwnerControls([string]$WriterPath,[string]$CollectorText,[string]$CaptureText){
    . (Join-Path $PSScriptRoot case-module.ps1)
    $controls=New-ODCaseModule $WriterPath $CollectorText
    $script:CaseModules.Add([pscustomobject]@{Boundary='Owner';Case='Controls';Module=$controls})
    $failurePublisher=New-ODCaseModule $WriterPath $CollectorText
    $script:CaseModules.Add([pscustomobject]@{Boundary='Capture';Case='FinalWriteControl';Module=$failurePublisher})
    & $failurePublisher {
        $script:ControlOpens=0
        function script:New-ODRecordStream([string]$Path){$script:ControlOpens++;return & $script:OriginalFactory $Path}
    }
    $controlRoot=Join-Path $packet 'unit/writer-green/owner-controls'
    LS-Need (-not (Test-Path -LiteralPath $controlRoot)) 'owner controls no retry'
    $null=[IO.Directory]::CreateDirectory($controlRoot)
    $captureTokens=$null;$captureIssues=$null
    $captureAst=[Management.Automation.Language.Parser]::ParseInput($CaptureText,[ref]$captureTokens,[ref]$captureIssues)
    LS-Need ($captureIssues.Count -eq 0) 'owner capture AST'
    $functions=@(foreach($name in @('Capture-Value','Capture-Once','Capture-ErrorSummary')){
        $found=@($captureAst.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $name},$false))
        LS-Need ($found.Count -eq 1) ('actual owner function '+$name)
        $found[0].Extent.Text
    })
    $profileTokens=$null;$profileIssues=$null
    $profileAst=[Management.Automation.Language.Parser]::ParseFile((Join-Path ([IO.Path]::GetDirectoryName($WriterPath)) profiles.ps1),[ref]$profileTokens,[ref]$profileIssues)
    $profileDefinition=@($profileAst.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Invoke-ODProfile'},$false))
    LS-Need ($profileIssues.Count -eq 0 -and $profileDefinition.Count -eq 1) 'actual intent owner AST'
    & $controls {
        param($Definitions,$ProfileText,$Root,$Repo,$Packet,$Leaf,$Publisher,$FailurePublisher)
        foreach($text in $Definitions){. ([ScriptBlock]::Create($text))}
        . ([ScriptBlock]::Create($ProfileText))
        $script:ControlRoot=$Root;$script:Repo=$Repo;$script:Packet=$Packet;$script:TestLeaf=$Leaf
        $script:KernelPath=Join-Path $Repo 'scripts/offline-diagnostics-v4/private/session.psm1'
        $script:ProfilePhase='writer-green';$script:ProfileManifestPath=Join-Path $Packet 'writer-green.frozen.json'
        $script:ProfileRootBase=Join-Path $Root intent-profile
        $script:Publisher=$Publisher
        $script:FailurePublisher=$FailurePublisher
        function script:Control-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('ASSERT: '+$Label)}}
        function script:Control-Fault {
            try{$null=[IO.File]::ReadAllText((Join-Path $script:ControlRoot missing.txt))}
            catch{return $_}
            throw 'ASSERT: real missing-file fault absent'
        }
        function script:Control-Outcome($Primary,[Management.Automation.ErrorRecord[]]$Secondary=@()){
            return [pscustomobject][ordered]@{Succeeded=($null -eq $Primary);PrimaryError=$Primary;SecondaryErrors=$Secondary}
        }
        function script:Control-ExactRejection([ScriptBlock]$Body){
            $caught=$null
            try{& $Body}catch{$caught=$_}
            Control-Need ($null -ne $caught) 'contract must reject'
            if($caught.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $caught.Exception.Message -cne 'OD: record outcome contract mismatch' -or $caught.FullyQualifiedErrorId -cne 'OD: record outcome contract mismatch'){throw $caught}
        }
    } $functions $profileDefinition[0].Extent.Text $controlRoot $repo $packet $PSScriptRoot $script:Publisher $failurePublisher
    $controlResults=[Collections.Generic.List[object]]::new()
    foreach($name in $oracle.OwnerControlNames){
        $result=& $controls {
            param($Name)
            $script:HostPrimary=$null;$script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new();$script:HostOverflow=$false
            $first=Control-Fault
            $assertions=0
            switch -CaseSensitive ($Name){
                'SameRecordOnce' {
                    Capture-Failure $first;Capture-Failure $first
                    Control-Need ([object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 0) $Name
                    $assertions=1
                }
                'DistinctSameTextKept' {
                    $second=Control-Fault
                    Control-Need (-not [object]::ReferenceEquals($first,$second) -and $first.Exception.Message -ceq $second.Exception.Message) 'two independent same-text causes'
                    Capture-Failure $first;Capture-Failure $second
                    Control-Need ([object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 1 -and [object]::ReferenceEquals($second,$script:HostSecondary[0])) $Name
                    $assertions=2
                }
                'SharedExceptionDifferentRecordsKept' {
                    # Synthetic wrapper control, NOT an observed I/O event.
                    $a=[Management.Automation.ErrorRecord]::new($first.Exception,'ContractOnly',[Management.Automation.ErrorCategory]::NotSpecified,$null)
                    $b=[Management.Automation.ErrorRecord]::new($first.Exception,'ContractOnly',[Management.Automation.ErrorCategory]::NotSpecified,$null)
                    Capture-Failure $a;Capture-Failure $b
                    Control-Need ([object]::ReferenceEquals($a.Exception,$b.Exception) -and -not [object]::ReferenceEquals($a,$b) -and [object]::ReferenceEquals($a,$script:HostPrimary) -and $script:HostSecondary.Count -eq 1 -and [object]::ReferenceEquals($b,$script:HostSecondary[0])) $Name
                    $assertions=1
                }
                'ExistingPrimaryPreserved' {
                    $operation=Control-Fault;$cleanup=Control-Fault;Capture-Failure $first
                    $outcome=Control-Outcome $operation ([Management.Automation.ErrorRecord[]]@($cleanup))
                    $once=[pscustomobject]@{Consumed=$false};Register-ODRecordOutcome $outcome $once
                    Control-Need ([object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 2 -and [object]::ReferenceEquals($operation,$script:HostSecondary[0]) -and [object]::ReferenceEquals($cleanup,$script:HostSecondary[1])) $Name
                    $assertions=1
                }
                'OverflowBound8' {
                    Capture-Failure $first
                    $causes=@(1..9|ForEach-Object {Control-Fault})
                    foreach($cause in $causes){Capture-Failure $cause}
                    Control-Need ([object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 8 -and $script:HostOverflow) $Name
                    for($i=0;$i -lt 8;$i++){Control-Need ([object]::ReferenceEquals($causes[$i],$script:HostSecondary[$i])) 'secondary original order'}
                    $assertions=9
                }
                'OutcomeConsumedOnce' {
                    $outcome=Control-Outcome $first;$once=[pscustomobject]@{Consumed=$false}
                    Register-ODRecordOutcome $outcome $once
                    Control-ExactRejection {Register-ODRecordOutcome $outcome $once}
                    Control-Need ($once.Consumed -and [object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 0) $Name
                    $assertions=2
                }
                'MalformedOutcomeStops' {
                    Capture-Failure $first
                    $cleanup=Control-Fault
                    $bad=@(
                        [pscustomobject][ordered]@{Succeeded='true';PrimaryError=$null;SecondaryErrors=[Management.Automation.ErrorRecord[]]@()},
                        [pscustomobject][ordered]@{PrimaryError=$null;Succeeded=$true;SecondaryErrors=[Management.Automation.ErrorRecord[]]@()},
                        [pscustomobject][ordered]@{Succeeded=$true;PrimaryError=$first;SecondaryErrors=[Management.Automation.ErrorRecord[]]@()},
                        [pscustomobject][ordered]@{Succeeded=$false;PrimaryError=$null;SecondaryErrors=[Management.Automation.ErrorRecord[]]@()},
                        [pscustomobject][ordered]@{Succeeded=$false;PrimaryError=$first;SecondaryErrors=$cleanup},
                        [pscustomobject][ordered]@{Succeeded=$false;PrimaryError=$first;SecondaryErrors=[Management.Automation.ErrorRecord[]]@($cleanup,$cleanup)},
                        [pscustomobject][ordered]@{Succeeded=$true;PrimaryError=$null;SecondaryErrors=[Management.Automation.ErrorRecord[]]@();Extra=0}
                    )
                    foreach($value in $bad){
                        $once=[pscustomobject]@{Consumed=$false}
                        Control-ExactRejection {Register-ODRecordOutcome $value $once}
                        Control-Need ($once.Consumed -and [object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 0) 'malformed consumed without collector mutation'
                    }
                    $assertions=$bad.Count*2
                }
                'IntentFailureNoWork' {
                    $script:Invoked=$false;$script:HostStarts=0;$script:Sessions=[Collections.Generic.List[object]]::new()
                    $existing=Join-Path $script:ControlRoot intent-existing.bin
                    [IO.File]::WriteAllBytes($existing,[byte[]]@(80,82,73,79,82))
                    $script:FaultPath=$existing;$script:ControlOpens=0
                    function script:New-ODRecordStream([string]$Path){$script:ControlOpens++;return & $script:OriginalFactory $script:FaultPath}
                    try{$report=Invoke-ODProfile -Profile entry-migration-core}
                    finally{Set-Item Function:script:New-ODRecordStream $script:OriginalFactory}
                    Control-Need ($report.Schema -eq 2 -and $report.Status -ceq 'INCOMPLETE' -and $report.Groups -is [array] -and $report.Groups.Count -eq 0 -and $report.Observations -is [array] -and $report.Observations.Count -eq 0 -and $report.Closed -and $script:Invoked -and $script:HostStarts -eq 0 -and $script:Sessions.Count -eq 0 -and $script:ActiveSessions.Count -eq 0 -and $script:ControlOpens -eq 1 -and [object]::ReferenceEquals($report.PrimaryError,$script:HostPrimary) -and $report.PrimaryError -is [Management.Automation.ErrorRecord]) $Name
                    $script:IntentFailureReport=$report
                    $assertions=1
                }
                'PublisherSingleOwner' {
                    $view=& $script:Publisher {
                        $absent=$true
                        foreach($key in @('HostPrimary','HostSecondary','HostOverflow')){if($null -ne (Get-Variable $key -Scope Script -ErrorAction Ignore)){$absent=$false}}
                        return $absent
                    }
                    $outcome=& $script:Publisher {param($Path)Write-ODCaptureRecord $Path ([byte[]]@(65,66,0,255))} (Join-Path $script:ControlRoot publisher.bin)
                    $once=[pscustomobject]@{Consumed=$false};Register-ODRecordOutcome $outcome $once
                    Control-Need ($view -and $outcome.Succeeded -and $once.Consumed -and $null -eq $script:HostPrimary -and $script:HostSecondary.Count -eq 0) $Name
                    $assertions=1
                }
                'SummaryFaultKeepsPrimary' {
                    Capture-Failure $first;$script:ErrorOutputAttempted=$false
                    $original=[Console]::Error
                    $stream=[IO.MemoryStream]::new();$closed=[IO.StreamWriter]::new($stream);$closed.Dispose()
                    try{[Console]::SetError($closed);Capture-ErrorSummary;Capture-ErrorSummary}
                    finally{[Console]::SetError($original)}
                    Control-Need ($script:ErrorOutputAttempted -and [object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 1 -and $script:HostSecondary[0].Exception -is [Management.Automation.MethodInvocationException]) $Name
                    $assertions=1
                }
                'ProjectionFaultKeepsPrimary' {
                    Capture-Failure $first;$caught=$null
                    try{$null=Capture-Value ([object]::new())}catch{$caught=$_;Capture-Failure $_}
                    Control-Need ($null -ne $caught -and $caught.Exception.Message -ceq 'Unsupported evidence value' -and [object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 1 -and [object]::ReferenceEquals($caught,$script:HostSecondary[0])) $Name
                    $assertions=1
                }
                'FinalWriteFailureNoRetry' {
                    # A private writer scope; never replace the real publisher's
                    # factory. Capture-Once itself is the actual AST above.
                    Capture-Failure $first
                    $path=Join-Path $script:ControlRoot final-existing.bin
                    [byte[]]$originalBytes=@(80,82,73,79,82)
                    [IO.File]::WriteAllBytes($path,$originalBytes)
                    $script:EvidenceAttempted=$false;$script:Aborted=$false
                    $originalPublisher=$script:Publisher
                    $script:Publisher=$script:FailurePublisher
                    $repeat=$null
                    try{
                        $published=Capture-Once $path ([ordered]@{Probe='failure'})
                        try{$null=Capture-Once $path ([ordered]@{Probe='retry'})}catch{$repeat=$_}
                    }finally{$script:Publisher=$originalPublisher}
                    $opens=& $script:FailurePublisher {$script:ControlOpens}
                    Control-Need (-not $published -and $script:Aborted -and $script:EvidenceAttempted -and $opens -eq 1 -and $null -ne $repeat -and $repeat.Exception.Message -ceq 'Evidence attempt consumed' -and [object]::ReferenceEquals($first,$script:HostPrimary) -and $script:HostSecondary.Count -eq 1 -and $script:HostSecondary[0].FullyQualifiedErrorId -ceq 'IOException' -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($path),$originalBytes)) $Name
                    $assertions=1
                }
                default {throw 'ASSERT: unknown owner control'}
            }
            [ordered]@{Name=$Name;Assertions=$assertions;ContractOnly=($Name -ceq 'SharedExceptionDifferentRecordsKept');Primary=$script:HostPrimary;Secondary=$script:HostSecondary.ToArray();Overflow=$script:HostOverflow;IntentReport=$(if($Name -ceq 'IntentFailureNoWork'){$script:IntentFailureReport}else{$null})}
        } $name
        $controlResults.Add($result)
        LS-Need ($result.Assertions -gt 0) ('owner control assertions '+$name)
        LS-Result $name $true ([ordered]@{Verified=$true})
    }
    LS-Details OwnerControls ($controlResults.ToArray())
    & $controls {$script:HostPrimary=$null;$script:HostSecondary.Clear();$script:HostOverflow=$false}
}
