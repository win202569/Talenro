# Actual production writer functions run in isolated private test scopes.
function Invoke-ODWriterTests {
    Add-Type -TypeDefinition ([IO.File]::ReadAllText((Join-Path $PSScriptRoot fault-stream.cs))) -Language CSharp -ErrorAction Stop
    $source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot capture.ps1))
    $tokens=$null;$parseErrors=$null
    $ast=[Management.Automation.Language.Parser]::ParseInput($source,[ref]$tokens,[ref]$parseErrors)
    LS-Need ($parseErrors.Count -eq 0) 'capture source AST'
    $collector=@($ast.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Capture-Failure'},$false))
    LS-Need ($collector.Count -eq 1) 'fixed collector definition'
    $writer=Join-Path $repo 'scripts/offline-diagnostics-v3/private/write-boundary.ps1'
    $outer=(Get-Command New-ODRecordStream).ScriptBlock
    foreach($definition in $oracle.WriterCases){
        $root=Join-Path $packet ('unit/'+$Grant+'/'+$definition.Boundary+'/'+$definition.Case)
        LS-Need (-not (Test-Path -LiteralPath $root)) 'unit case no retry'
        $null=[IO.Directory]::CreateDirectory($root)
        . (Join-Path $PSScriptRoot case-module.ps1) $writer $collector[0].Extent.Text
        $module=$caseModule
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
            $thrown=$null
            # Factory replacement stays inside this explicitly private module.
            try{
                if($Boundary -ceq 'Capture'){Write-ODCaptureRecord $path $payload}
                else{Write-ODProfileIntentRecord $path $payload}
            }catch{$thrown=$_;Capture-Failure $_}
            $all=@();if($null -ne $script:HostPrimary){$all+=,$script:HostPrimary};$all+=@($script:HostSecondary.ToArray())
            $wanted=if($Case -cin @('WriteDisposeFirstCause','FlushDisposeFirstCause')){'ObjectDisposedException'}elseif($Case -ceq 'DisposeOnly'){'FileNotFoundException'}elseif($Case -ceq 'OpenFailureOnce'){'IOException'}else{''}
            $first=@($all | Where-Object {$_.FullyQualifiedErrorId -ceq $wanted})
            $dispose=@($all | Where-Object {$_.FullyQualifiedErrorId -ceq 'FileNotFoundException'})
            if($Case -cne 'SuccessOnce' -and $first.Count -ne 1){throw 'ASSERT: first real error retained once'}
            $isFirst=if($Case -ceq 'SuccessOnce'){$null -eq $thrown}else{[object]::ReferenceEquals($first[0],$thrown)}
            $primaryFirst=if($Case -ceq 'SuccessOnce'){$null -eq $script:HostPrimary}else{[object]::ReferenceEquals($first[0],$script:HostPrimary)}
            $compound=$Case -cin @('WriteDisposeFirstCause','FlushDisposeFirstCause')
            $secondaryDispose=if($compound){$dispose.Count -eq 1 -and $script:HostSecondary.Count -eq 1 -and [object]::ReferenceEquals($script:HostSecondary[0],$dispose[0])}else{$script:HostSecondary.Count -eq 0}
            $bytesExact=$Case -ceq 'SuccessOnce' -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($path),$payload)
            $unchanged=$Case -ceq 'OpenFailureOnce' -and [Linq.Enumerable]::SequenceEqual([byte[]][IO.File]::ReadAllBytes($path),[byte[]]@(80,82,73,79,82))
            [pscustomobject][ordered]@{
                Boundary=$Boundary;Case=$Case;Thrown=$thrown;Primary=$script:HostPrimary;Secondary=$script:HostSecondary.ToArray()
                OpenCalls=$script:OpenCount;WriteCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.WriteCalls})
                FlushCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.FlushCalls});DisposeCalls=$(if($null -eq $script:Wrapper){0}else{$script:Wrapper.DisposeCalls})
                ThrownIsFirst=$isFirst;PrimaryIsFirst=$primaryFirst;SecondaryIsDispose=$secondaryDispose
                BytesExact=$bytesExact;ExistingUnchanged=$unchanged;Overflow=$script:HostOverflow
            }
        } $definition.Boundary $definition.Case $root
        }finally{Observe-ODRecorderGuard $script:RecorderGuard Finally $definition.Name}
        $same=[object]::ReferenceEquals($outer,(Get-Command New-ODRecordStream).ScriptBlock)
        LS-Need $same 'unit factory cannot change capture factory'
        $facts=[ordered]@{
            Boundary=$result.Boundary;Case=$result.Case;ThrownIsFirst=$result.ThrownIsFirst;PrimaryIsFirst=$result.PrimaryIsFirst;SecondaryIsDispose=$result.SecondaryIsDispose
            OpenCalls=$result.OpenCalls;WriteCalls=$result.WriteCalls;FlushCalls=$result.FlushCalls;DisposeCalls=$result.DisposeCalls
            SecondaryCount=$result.Secondary.Count;BytesExact=$result.BytesExact;ExistingUnchanged=$result.ExistingUnchanged;FactoryIsolated=$same
        }
        LS-Details RecordCase ([ordered]@{Boundary=$result.Boundary;Case=$result.Case;Thrown=$result.Thrown;Primary=$result.Primary;Secondary=$result.Secondary;OpenCalls=$result.OpenCalls;WriteCalls=$result.WriteCalls;FlushCalls=$result.FlushCalls;DisposeCalls=$result.DisposeCalls;BytesExact=$result.BytesExact;ExistingUnchanged=$result.ExistingUnchanged;FactoryIsolated=$same;Overflow=$result.Overflow})
        $expected=@($oracle.Grants['writer-green'].Groups | Where-Object Name -CEQ $definition.Name)[0].Facts
        LS-Result $definition.Name ((ConvertTo-Json -InputObject $facts -Depth 10 -Compress) -ceq (ConvertTo-Json -InputObject $expected -Depth 10 -Compress)) $facts
        $module=$null
    }
}
