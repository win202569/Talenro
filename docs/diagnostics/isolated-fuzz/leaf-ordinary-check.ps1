# Fixed ordinary verification only. Importing this entry launches no workload.
. (Join-Path $PSScriptRoot 'mapping-single-check.ps1')
function Assert-LeafOrdinaryEqual($Actual,$Expected){
    if($null -eq $Expected){if($null -ne $Actual){throw 'Unexpected non-null value'};return}
    if($Expected -is [System.Collections.IDictionary]){
        if($null -eq $Actual){throw 'Map missing'}
        $map=New-Object 'System.Collections.Generic.Dictionary[string,object]' ([StringComparer]::Ordinal)
        if($Actual -is [System.Collections.IDictionary]){foreach($key in $Actual.Keys){$map.Add([string]$key,$Actual[$key])}}
        else{foreach($prop in $Actual.PSObject.Properties){$map.Add($prop.Name,$prop.Value)}}
        if($map.Count -ne $Expected.Count){throw 'Map keys differ'}
        foreach($key in $Expected.Keys){if(-not $map.ContainsKey($key)){throw 'Key missing'};Assert-LeafOrdinaryEqual $map[$key] $Expected[$key]};return
    }
    if($Expected -is [array]){
        if($Actual -isnot [array] -or $Actual.Count -ne $Expected.Count){throw 'Array count/type differs'}
        for($i=0;$i -lt $Expected.Count;$i++){Assert-LeafOrdinaryEqual $Actual[$i] $Expected[$i]};return
    }
    if($Expected -is [pscustomobject]){$map=@{};foreach($prop in $Expected.PSObject.Properties){$map[$prop.Name]=$prop.Value};Assert-LeafOrdinaryEqual $Actual $map;return}
    $valid=if($Expected -is [bool]){$Actual -is [bool] -and $Actual -eq $Expected}elseif($Expected -is [int] -or $Expected -is [long]){($Actual -is [int] -or $Actual -is [long]) -and $Actual -eq $Expected}else{$Actual -is [string] -and $Actual -ceq $Expected}
    if(-not $valid){throw 'Fixed value/type differs'}
}
function Get-LeafOrdinaryFixedPlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache,[bool]$RequireNew){
    $workspace=Get-IsolatedFullPath ([IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')))
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if([IntPtr]::Size -ne 8 -or -not(Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^leaf-ordinary-check-[0-9]{3}$' -or ($RequireNew -and (Test-Path -LiteralPath $root))){throw 'New owned ordinary identity/64-bit host required'}
    if(-not(Test-IsolatedWithin $tool $scope) -or [IO.Path]::GetFileName($tool) -cne 'toolchain' -or [IO.Path]::GetFileName((Split-Path -Parent $tool)) -cnotmatch '^uninstrumented-[0-9]{3}$'){throw 'Owned uninstrumented toolchain required'}
    foreach($inputRoot in @($tool,$modules)){if($root -ieq $inputRoot -or (Test-IsolatedWithin $root $inputRoot) -or (Test-IsolatedWithin $inputRoot $root)){throw 'Output/input overlap'}}
    $reference=Get-EnvelopeSampleFile $ReferenceReceipt
    if(-not(Test-IsolatedWithin $reference $scope) -or (Test-IsolatedWithin $reference $root)){throw 'Owned non-output reference required'}
    $inputs=@($reference,(Join-Path (Split-Path -Parent $tool) 'copy-receipt.json'),(Join-Path $tool 'bin/go.exe'),(Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod'))
    $inputRows=@(foreach($file in $inputs){$safe=Get-EnvelopeSampleFile $file;[pscustomobject]@{Path=$safe;SHA256=(Get-FileHash -LiteralPath $safe -Algorithm SHA256).Hash}})
    $sourceNames=@('leaf-ordinary-check.ps1','mapping-single-check.ps1','matrix-check.ps1','envelope-uninstrumented.ps1','envelope-sample.ps1','envelope-observe.ps1','run.ps1','isolation.ps1','provenance.ps1','process-native.cs','retained-lifetime.cs','compile-stage-trigger.cs','compile-stage-session.cs')
    $sources=@(foreach($name in $sourceNames){$file=Get-EnvelopeSampleFile (Join-Path $PSScriptRoot $name);[pscustomobject]@{Path=$name;SHA256=(Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash}})
    $binary=Join-Path $root 'package.test.exe'
    $envMap=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    [pscustomobject]@{Schema=1;Policy='leaf-ordinary-check-v1';Package='internal/nodecontrol/nodeleafbinding';VetMode='default';NoRetry=$true;Acceptance=$false;RunRoot=$root;Project=$workspace;Toolchain=$tool;ReferenceReceipt=$reference;ModuleCache=$modules;Executable=$inputRows[2].Path;Binary=$binary;WorkingDirectory=(Join-Path $workspace 'internal/nodecontrol/nodeleafbinding');Environment=$envMap;BuildLimitSeconds=600;ExecuteLimitSeconds=150;BuildArguments=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),'talenro.local/platform/internal/nodecontrol/nodeleafbinding');Arguments=@('-test.count=1','-test.timeout=120s','-test.v=true','-test.run=.');SourceBinding=$sources;InputBinding=$inputRows}
}
function Get-LeafOrdinaryCheckPlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    Get-LeafOrdinaryFixedPlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache $true
}
function Assert-LeafOrdinaryCheckPlan([object]$Plan,[bool]$RequireNew){
    $fixed=Get-LeafOrdinaryFixedPlan $Plan.RunRoot $Plan.Toolchain $Plan.ReferenceReceipt $Plan.ModuleCache $RequireNew
    Assert-LeafOrdinaryEqual $Plan $fixed
}
function Write-LeafOrdinaryJSON([string]$Path,[object]$Value){
    $safe=Get-IsolatedFullPath $Path
    if(Test-Path -LiteralPath $safe){$null=Get-EnvelopeSampleFile $safe}
    $bytes=[Text.UTF8Encoding]::new($false).GetBytes(($Value|ConvertTo-Json -Depth 30))
    $stream=[IO.File]::Open($safe,[IO.FileMode]::Create,[IO.FileAccess]::Write,[IO.FileShare]::Read)
    try{$stream.Write($bytes,0,$bytes.Length);$stream.Flush($true)}finally{$stream.Dispose()}
}
function Get-LeafOrdinaryBinding([string[]]$Files){
    [pscustomobject]@{Files=@(foreach($path in $Files){$file=Get-EnvelopeSampleFile $path;[pscustomobject]@{Path=$file;Length=(Get-Item -LiteralPath $file).Length;SHA256=(Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash}})}
}
function Assert-LeafOrdinaryBinding([object]$Binding){
    foreach($row in $Binding.Files){$file=Get-EnvelopeSampleFile $row.Path;if((Get-Item -LiteralPath $file).Length -ne $row.Length -or (Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash -cne $row.SHA256){throw 'Ordinary artifact drift'}}
}
function Get-LeafOrdinaryTestOutput([string]$Stdout,[string]$Stderr){
    $out=Get-Content -LiteralPath (Get-EnvelopeSampleFile $Stdout) -Raw -ErrorAction Stop
    $err=Get-Content -LiteralPath (Get-EnvelopeSampleFile $Stderr) -Raw -ErrorAction Stop
    $lines=@($out -split '\r?\n'|Where-Object {$_ -cne ''})
    if($lines.Count -eq 0 -or $lines[-1] -cne 'PASS' -or ($out+"`n"+$err) -match '(?m)^\s*(?:--- FAIL:|FAIL(?:\s|$)|TALENRO_FUZZ_DIAG(?:\s|$))'){throw 'Original test PASS evidence missing or contradictory'}
    $values=@{}
    foreach($kind in @('Run','Pass','Skip')){
        $pattern=if($kind -ceq 'Run'){'(?m)^\s*=== RUN\s+(Test[^\s]+)\s*$'}else{'(?m)^\s*--- '+$kind.ToUpperInvariant()+': (Test[^\s]+)\s+\('}
        $names=@([regex]::Matches($out,$pattern)|ForEach-Object {$_.Groups[1].Value})
        $values['TopLevel'+$kind]=@($names|Where-Object {-not $_.Contains('/')})
        $values['Subtest'+$kind]=@($names|Where-Object {$_.Contains('/')})
    }
    if($values.TopLevelPass.Count -lt 1){throw 'No original top-level test PASS'}
    [pscustomobject]$values
}
function Assert-LeafOrdinaryReceipt([object]$Receipt,[object]$Plan,[bool]$Build){
    Assert-ProcessCompletion $Receipt
    if($Receipt.ShutdownSchema -isnot [int] -and $Receipt.ShutdownSchema -isnot [long]){throw 'Typed ordinary schema missing'}
    if($Receipt.ShutdownSchema -ne 3 -or $Receipt.ExitCode -ne 0){throw 'Ordinary schema3/exit0 required'}
    foreach($pair in @(@('Executable',$(if($Build){$Plan.Executable}else{$Plan.Binary})),@('WorkingDirectory',$(if($Build){$Plan.Project}else{$Plan.WorkingDirectory})),@('Arguments',$(if($Build){$Plan.BuildArguments}else{$Plan.Arguments})),@('Environment',$Plan.Environment),@('StdoutPath',(Join-Path $Plan.RunRoot $(if($Build){'build/stdout.txt'}else{'execute/stdout.txt'}))),@('StderrPath',(Join-Path $Plan.RunRoot $(if($Build){'build/stderr.txt'}else{'execute/stderr.txt'}))))){Assert-LeafOrdinaryEqual $Receipt.($pair[0]) $pair[1]}
    Assert-MappingQueryEvidence $Receipt
}
function Invoke-LeafOrdinaryCheck([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    $plan=Get-LeafOrdinaryCheckPlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    $root=$plan.RunRoot
    New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
    $state=[pscustomobject]@{Schema=1;Policy=$plan.Policy;Package=$plan.Package;VetMode='default';NoRetry=$true;Acceptance=$false;PackagePassed=$false;FinalizationRequired=$true;CompilePassed=$false;DefaultVetBuildPassed=$false;ExecutionAttempted=$false;OriginalTestPassed=$false;ProductUnchanged=$false;ToolchainUnchanged=$false;ArtifactsUnchanged=$false;Failure=$null;AuditErrors=@();ToolchainReference=$null;TestOutput=$null;StartUTC=[DateTime]::UtcNow.ToString('o');EndUTC=$null;PhaseTimings=@()}
    $attempts=New-Object 'System.Collections.Generic.List[object]';$bindings=New-Object 'System.Collections.Generic.List[object]'
    $before=@();$identity=$null;$pending=$null;$phase='prepare';$total=[Diagnostics.Stopwatch]::StartNew()
    try{
        Write-LeafOrdinaryJSON (Join-Path $root 'plan.json') $plan
        Write-LeafOrdinaryJSON (Join-Path $root 'source-binding.json') $plan.SourceBinding
        Write-LeafOrdinaryJSON (Join-Path $root 'input-binding.json') $plan.InputBinding
        Write-LeafOrdinaryJSON (Join-Path $root 'progress.json') @{Attempts=@()}
        $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt;$state.ToolchainReference=$identity
        $before=@(Get-EnvelopeProductManifest);Write-LeafOrdinaryJSON (Join-Path $root 'before-product.json') $before
        foreach($name in @('go-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $root $name) -ErrorAction Stop|Out-Null}
        $state.PhaseTimings+=@([pscustomobject]@{Phase='prepare';ElapsedMs=$total.Elapsed.TotalMilliseconds})
        foreach($build in @($true,$false)){
            $phase=if($build){'build'}else{'execute'}
            Assert-LeafOrdinaryCheckPlan $plan $false
            foreach($binding in $bindings){Assert-LeafOrdinaryBinding $binding}
            $dir=Join-Path $root $phase
            $attempts.Add([pscustomobject]@{Phase=$phase;Receipt=(Join-Path $dir 'receipt.json');UTC=[DateTime]::UtcNow.ToString('o')})
            if(-not $build){$state.ExecutionAttempted=$true}
            Write-LeafOrdinaryJSON (Join-Path $root 'progress.json') @{Attempts=@($attempts.ToArray())}
            $clock=[Diagnostics.Stopwatch]::StartNew()
            try{$receipt=Invoke-MatrixRecordedProcess $(if($build){$plan.Executable}else{$plan.Binary}) $(if($build){$plan.BuildArguments}else{$plan.Arguments}) $plan.Environment $(if($build){$plan.Project}else{$plan.WorkingDirectory}) $dir $(if($build){600}else{150})}
            catch{
                $primary=$_
                try{Write-LeafOrdinaryJSON (Join-Path $dir 'launch-error.json') @{Error=$primary.Exception.Message;Acceptance=$false;NoRetry=$true}}catch{$state.AuditErrors+=@('launch-error-persistence: '+$_.Exception.Message)}
                throw $primary
            }finally{$state.PhaseTimings+=@([pscustomobject]@{Phase=$phase;ElapsedMs=$clock.Elapsed.TotalMilliseconds})}
            $files=@((Join-Path $dir 'receipt.json'),$receipt.StdoutPath,$receipt.StderrPath)
            if($build -and $receipt.ExitCode -eq 0){$files+=@($plan.Binary)}
            $bindings.Add((Get-LeafOrdinaryBinding $files))
            Assert-LeafOrdinaryReceipt $receipt $plan $build
            if($build){
                $log=Get-Content -LiteralPath (Get-EnvelopeSampleFile $receipt.StderrPath) -Raw
                Assert-MatrixCompileEvidence $log $plan.Toolchain $plan.Project $plan.Package
                Assert-CopiedCompileEvidence $log $plan.Toolchain
                $state.CompilePassed=$true;$state.DefaultVetBuildPassed=$true
            }else{
                $state.TestOutput=Get-LeafOrdinaryTestOutput $receipt.StdoutPath $receipt.StderrPath
                $state.OriginalTestPassed=$true
            }
        }
    }catch{$pending=$_;$state.Failure=$_.Exception.Message}
    finally{
        $state.EndUTC=[DateTime]::UtcNow.ToString('o')
        if($null -ne $pending){try{Write-LeafOrdinaryJSON (Join-Path $root 'stopped.json') @{Phase=$phase;Error=$state.Failure;Acceptance=$false;NoRetry=$true}}catch{$state.AuditErrors+=@('stopped-persistence: '+$_.Exception.Message)}}
        try{Write-LeafOrdinaryJSON (Join-Path $root 'result.json') $state}catch{if($null -eq $pending){throw};throw $pending}
    }
    if($null -ne $pending){throw $pending}
    return $state
}
