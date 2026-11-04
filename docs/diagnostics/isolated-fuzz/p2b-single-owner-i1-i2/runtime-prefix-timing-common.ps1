# Definitions only; no operational entry point is evaluated.
function Write-PTOnce([string]$Path,$Value){
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $Value -Depth 100))
    $stream=[IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}finally{$stream.Dispose()}
}
function Get-PTHash([string]$Path){return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash}
function Assert-PTInputs($Binding){
    $observed=[ordered]@{}
    foreach($path in $Binding.AbsentPaths){if(Test-Path -LiteralPath $path){throw ('PT: absent pin occupied '+$path)}}
    foreach($map in @($Binding.SourceSHA256,$Binding.PreservedSHA256,$Binding.HistoricalEvidenceSHA256,$Binding.HelperSHA256)){
        foreach($path in $map.Keys){
            $absolute=if([IO.Path]::IsPathFullyQualified($path)){$path}else{Join-Path $script:PTRepo $path}
            $hash=Get-PTHash $absolute
            if($hash -cne $map[$path]){throw ('PT: immutable drift '+$path)}
            $observed[$path]=$hash
        }
    }
    return $observed
}
function Get-PTDefinitions([string]$SourcePath,[string[]]$Names){
    $allowed=[ordered]@{
        'source-model.ps1'=@('Get-ODEpochManifestName','Get-ODPinnedSHA256','Get-ODSourceOwner','Get-ODFailureExtent','Read-ODSourceModel','New-ODFixedContext','New-ODHistoricalContext')
        'observation-contract.ps1'=@('OC-Need','OC-Keys','OC-Integer','OC-Expand','OC-SourceError','OC-Match','Assert-ODProfileObservations')
    }
    $leaf=[IO.Path]::GetFileName($SourcePath)
    if(-not $allowed.Contains($leaf) -or $Names.Count -ne $allowed[$leaf].Count -or @(Compare-Object -CaseSensitive $Names $allowed[$leaf]).Count){throw 'PT: definition allowlist'}
    $expected=Join-Path $script:PTRepo ('docs/diagnostics/isolated-fuzz/p2b-single-owner-i1-i2/'+$leaf)
    if([IO.Path]::GetFullPath($SourcePath) -cne [IO.Path]::GetFullPath($expected)){throw 'PT: definition path'}
    $relative=[IO.Path]::GetRelativePath($script:PTRepo,$expected).Replace('\','/')
    if((Get-PTHash $expected) -cne $script:PTBinding.SourceSHA256[$relative]){throw 'PT: source identity'}
    $tokens=$null;$errors=$null
    $tree=[Management.Automation.Language.Parser]::ParseFile($expected,[ref]$tokens,[ref]$errors)
    if($errors.Count){throw 'PT: source parse'}
    foreach($name in $Names){
        $nodes=@($tree.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq $name},$true))
        if($nodes.Count -ne 1){throw ('PT: definition count '+$name)}
        $nodes[0].Extent.Text
    }
}
function Write-PTEvent([string]$Stage,[string]$Kind,[string]$Profile,[string]$Source){
    if($script:PTJournalFailed){throw 'PT: journal latched'}
    try{
        if($Kind -cnotin @('begin','end','error')){throw 'PT: event kind'}
        if($Profile -cnotin @('','entry-migration-core','entry-migration-owned','entry-migration-live')){throw 'PT: event profile'}
        if($Source){
            if([IO.Path]::IsPathFullyQualified($Source)){$Source=[IO.Path]::GetRelativePath($script:PTRepo,$Source).Replace('\','/')}
            if(-not $script:PTBinding.SourceSHA256.Contains($Source) -and -not $script:PTBinding.PreservedSHA256.Contains($Source) -and -not $script:PTBinding.HistoricalEvidenceSHA256.Contains($Source)){throw 'PT: event source'}
        }
        if($Kind -ceq 'begin'){
            $script:PTEventId++
            $parent=if($script:PTStack.Count){$script:PTStack.Peek().EventId}else{$null}
            $frame=[ordered]@{EventId=$script:PTEventId;ParentId=$parent;Stage=$Stage;Profile=$Profile;Source=$Source}
        }else{
            if(-not $script:PTStack.Count){throw 'PT: event stack empty'}
            $frame=$script:PTStack.Peek()
            if($frame.Stage -cne $Stage -or $frame.Profile -cne $Profile -or $frame.Source -cne $Source){throw 'PT: event stack mismatch'}
        }
        $record=[ordered]@{Schema=1;Seq=([long]$script:PTSeq+1);EventId=[long]$frame.EventId;ParentId=$frame.ParentId;Stage=$Stage;Kind=$Kind;Profile=$Profile;Source=$Source;ElapsedTicks=[long]$script:PTClock.ElapsedTicks;Frequency=[long][Diagnostics.Stopwatch]::Frequency}
        $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject $record -Depth 8 -Compress)+"\n".Replace('\n',[string][char]10))
        if($script:PTSeq -ge 4096 -or $script:PTJournalBytes+$bytes.Length -gt 1048576){throw 'PT: journal bound'}
        $script:PTJournal.Write($bytes,0,$bytes.Length);$script:PTJournal.Flush()
        $script:PTSeq++;$script:PTJournalBytes+=$bytes.Length;$script:PTLastEvent=$record
        if($Kind -ceq 'begin'){$script:PTStack.Push($frame)}else{$null=$script:PTStack.Pop()}
    }catch{$script:PTJournalFailed=$true;throw}
}
function Get-PTError($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
