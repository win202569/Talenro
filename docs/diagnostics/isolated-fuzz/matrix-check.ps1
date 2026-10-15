. (Join-Path $PSScriptRoot 'envelope-uninstrumented.ps1')
function Get-MatrixPackages {
    @('internal/nodecontrol/nodeissuermapping','internal/nodecontrol/nodeissuerbinding','internal/nodecontrol/nodeleafbinding','internal/nodecontrol/nodecerttime','internal/nodecontrol/nodeissuer','internal/nodecontrol/nodeca','internal/nodecontrol/nodeleaf','internal/nodecontrol/nodeleafkeys','internal/nodecontrol/contracts','internal/nodecontrol/serving','internal/nodecontrol/statebinding','internal/nodecontrol/statecomposition','internal/nodecontrol/statecontract','internal/nodecontrol/stateformat','internal/nodecontrol/statematerials','internal/nodecontrol/statepolicy','internal/nodecontrol/statepolicybinding','internal/nodecontrol/staterecovery','internal/nodecontrol/statetime','internal/nodecontrol/statetimebinding','internal/nodecontrol/stateverify','internal/readiness','internal/nodecontrol/authority','internal/store')
}
function Get-LimitedMatrixPlan([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot;$tool=Get-IsolatedFullPath $Toolchain;$modules=Get-IsolatedFullPath $ModuleCache
    if(-not (Test-IsolatedWithin $root $scope) -or [IO.Path]::GetFileName($root) -cnotmatch '^matrix-check-[0-9]{3}$' -or (Test-Path -LiteralPath $root)){throw 'New owned matrix identity required; no replay'}
    $go=Get-EnvelopeSampleFile (Join-Path $tool 'bin/go.exe')
    if(-not (Test-Path -LiteralPath (Join-Path $modules 'github.com/google/uuid@v1.6.0/go.mod') -PathType Leaf)){throw 'Offline dependencies missing; no download'}
    if((Get-FileHash -LiteralPath (Join-Path $workspace 'internal/nodecontrol/nodeissuer/fuzz_test.go')).Hash -cne '318CF8DE04BA5D95E71E625ABF6AAED8DFD54B854A9A925B7A3004D4132D989E'){throw 'Original callbacks changed'}
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $root 'go-cache');GOMODCACHE=$modules;GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='on';GOWORK='off';GOFLAGS='-mod=readonly';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $root 'tmp');TMP=(Join-Path $root 'tmp');GOTMPDIR=(Join-Path $root 'tmp');TALENRO_FUZZ_DIAG_MODE='';TALENRO_FUZZ_FIXTURE='';GODEBUG=''}
    $checks=New-Object 'System.Collections.Generic.List[object]';$index=0
    foreach($package in @(Get-MatrixPackages)){
        $index++;$name=($package -split '/')[-1];$directory=Join-Path $root ('{0:D2}-{1}' -f $index,$name);$binary=Join-Path $directory 'package.test.exe'
        $build=@('test','-c','-x','-p=1','-count=1','-timeout=120s',('-o='+$binary),('talenro.local/platform/'+$package))
        if($index -eq 5){$build=@('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzIssuerEnvelope$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$binary),('talenro.local/platform/'+$package))}
        $run=if($index -eq 23){'^(TestBoundRead|TestCoordinator)'}else{'.'}
        $checks.Add([pscustomobject]@{Index=$index;Kind=$(if($index -eq 24){'compile-only'}else{'ordinary'});Target=$package;Package=$package;Directory=$directory;Binary=$binary;BuildArguments=$build;Arguments=@('-test.count=1','-test.timeout=120s','-test.v=true',('-test.run='+$run));ExecuteLimitSeconds=150;WorkingDirectory=(Join-Path $workspace $package)})
    }
    foreach($target in @('FuzzIssuerEnvelope','FuzzIssuerPair')){
        $index++;$directory=Join-Path $root ('{0:D2}-{1}' -f $index,$target)
        $checks.Add([pscustomobject]@{Index=$index;Kind='fuzz';Target=$target;Package='internal/nodecontrol/nodeissuer';Directory=$directory;Binary=$checks[4].Binary;BuildArguments=@();Arguments=@('-test.run=^$',('-test.fuzz=^'+$target+'$'),'-test.fuzztime=5s','-test.parallel=1','-test.timeout=30s','-test.count=1',('-test.fuzzcachedir='+(Join-Path $directory 'fuzz-cache')));ExecuteLimitSeconds=60;WorkingDirectory=$checks[4].WorkingDirectory})
    }
    [pscustomobject]@{RunRoot=$root;Project=$workspace;Toolchain=$tool;ReferenceReceipt=$ReferenceReceipt;Executable=$go;Environment=$environment;BuildLimitSeconds=600;Checks=@($checks.ToArray())}
}
function Invoke-MatrixNative([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$WorkingDirectory,[string]$Stdout,[string]$Stderr,[int]$LimitSeconds){
    [IsolatedFuzzProcess]::Run($Executable,$Arguments,$Environment,$WorkingDirectory,$Stdout,$Stderr,$LimitSeconds,$false,$false)
}
function Invoke-MatrixRecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$WorkingDirectory,[string]$ReceiptDir,[int]$LimitSeconds){
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $scope=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $exe=Get-EnvelopeSampleFile $Executable;$receipt=Get-IsolatedFullPath $ReceiptDir;$cwd=Get-IsolatedFullPath $WorkingDirectory
    $allowed=@($workspace)+@(Get-MatrixPackages|ForEach-Object {Join-Path $workspace $_})
    if(-not (Test-IsolatedWithin $exe $scope) -or -not (Test-IsolatedWithin $receipt $scope) -or (Test-Path -LiteralPath $receipt) -or $allowed -inotcontains $cwd -or -not (Test-Path -LiteralPath $cwd -PathType Container) -or $LimitSeconds -lt 1 -or $LimitSeconds -gt 600){throw 'Matrix process path/cwd/budget invalid'}
    New-Item -ItemType Directory -Path $receipt -ErrorAction Stop|Out-Null
    $stdout=Join-Path $receipt 'stdout.txt';$stderr=Join-Path $receipt 'stderr.txt'
    try{
        $native=Invoke-MatrixNative $exe $Arguments $Environment $cwd $stdout $stderr $LimitSeconds
        $fields=@{};foreach($property in $native.PSObject.Properties){$fields[$property.Name]=$property.Value}
        foreach($pair in @(@('Executable',$exe),@('Arguments',$Arguments),@('Environment',$Environment),@('WorkingDirectory',$cwd),@('StdoutPath',$stdout),@('StderrPath',$stderr))){$fields[$pair[0]]=$pair[1]}
        $result=[pscustomobject]$fields
        $result|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $receipt 'receipt.json') -ErrorAction Stop
        return $result
    }catch{
        [pscustomobject]@{UTC=[DateTime]::UtcNow.ToString('o');LaunchOrReceiptError=$_.Exception.Message;CleanupVerified=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $receipt 'launch-error.json') -ErrorAction Stop
        throw
    }
}
function Assert-MatrixCompileEvidence([string]$Log,[string]$Toolchain,[string]$Project,[string]$Package){
    $compiler=Join-Path $Toolchain 'pkg/tool/windows_amd64/compile.exe';$directory=Join-Path $Project $Package
    foreach($line in ($Log -split '\r?\n')){
        $tokens=@([regex]::Matches($line,'"(?:\\.|[^"\\])*"|[^\s"]+')|ForEach-Object {$v=$_.Value;if($v.StartsWith('"')){$v=$v.Substring(1,$v.Length-2).Replace('\\','\').Replace('\"','"')};$v})
        if($tokens.Count -lt 4 -or $tokens[0] -cne $compiler){continue}
        $packageMatches=$false
        for($i=1;$i -lt $tokens.Count-1;$i++){if($tokens[$i] -ceq '-p' -and $tokens[$i+1] -cin @(('talenro.local/platform/'+$Package),('talenro.local/platform/'+$Package+'_test'))){$packageMatches=$true}}
        if(-not $packageMatches){continue}
        foreach($token in $tokens){if($token.EndsWith('_test.go',[StringComparison]::Ordinal) -and [IO.Path]::IsPathRooted($token)){$file=Get-IsolatedFullPath $token;if((Test-IsolatedWithin $file $directory) -and (Test-Path -LiteralPath $file -PathType Leaf)){return}}}
    }
    throw 'Compiled original package test source evidence missing'
}
function Assert-MatrixBinding([object]$Binding){
    foreach($row in $Binding.Files){if((Get-FileHash -LiteralPath (Get-EnvelopeSampleFile $row.Path)).Hash -cne $row.SHA256){throw 'Matrix binary/compile artifact drift'}}
}
function Save-MatrixProgress([object]$State,[string]$Root){$State|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $Root 'progress.json') -ErrorAction Stop}
function Invoke-LimitedMatrix([string]$RunRoot,[string]$Toolchain,[string]$ReferenceReceipt,[string]$ModuleCache){
    $plan=Get-LimitedMatrixPlan $RunRoot $Toolchain $ReferenceReceipt $ModuleCache
    $identity=Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt
    $before=@(Get-EnvelopeProductManifest)
    $state=[pscustomobject]@{Scope='One new limited26-check matrix; not historical replay';MatrixPassed=$false;Acceptance=$false;Passed=0;Failed=0;Unrun=26;Current=0;Phase='initialize';Failure=$null;FinalizationFailed=$false;ToolchainReference=$identity;Checks=(New-Object 'System.Collections.Generic.List[object]');ProcessAttempts=(New-Object 'System.Collections.Generic.List[object]');Bindings=(New-Object 'System.Collections.Generic.List[object]');RuntimeEvidence=(New-Object 'System.Collections.Generic.List[object]')}
    New-Item -ItemType Directory -Path $plan.RunRoot -ErrorAction Stop|Out-Null
    try{
        $plan|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'plan.json') -ErrorAction Stop
        $before|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'before-product.json') -ErrorAction Stop
        foreach($name in @('go-cache','tmp')){New-Item -ItemType Directory -Path (Join-Path $plan.RunRoot $name) -ErrorAction Stop|Out-Null}
        foreach($check in $plan.Checks){
            $state.Current=$check.Index
            New-Item -ItemType Directory -Path $check.Directory -ErrorAction Stop|Out-Null
            if($check.Kind -eq 'fuzz'){New-Item -ItemType Directory -Path (Join-Path $check.Directory 'fuzz-cache') -ErrorAction Stop|Out-Null}
            if($check.Kind -ne 'fuzz'){
                $state.Phase='compile';$receiptDir=Join-Path $check.Directory 'build'
                $state.ProcessAttempts.Add([pscustomobject]@{Index=$check.Index;Phase=$state.Phase;Receipt=(Join-Path $receiptDir 'receipt.json')});Save-MatrixProgress $state $plan.RunRoot
                $build=Invoke-MatrixRecordedProcess $plan.Executable $check.BuildArguments $plan.Environment $plan.Project $receiptDir $plan.BuildLimitSeconds
                Assert-ProcessCompletion $build
                if($build.ShutdownSchema -ne 3 -or $build.ExitCode -ne 0){throw 'Matrix compile failed'}
                $log=Get-Content -LiteralPath $build.StderrPath -Raw -ErrorAction Stop
                Assert-MatrixCompileEvidence $log $plan.Toolchain $plan.Project $check.Package
                if($check.Index -eq 1){Assert-CopiedCompileEvidence $log $plan.Toolchain}
                if($check.Index -eq 5){Assert-EnvelopeCompileEvidence $log $plan.Toolchain $plan.Project}
                $files=@(foreach($file in @($check.Binary,(Join-Path $receiptDir 'receipt.json'),$build.StdoutPath,$build.StderrPath)){$path=Get-EnvelopeSampleFile $file;[pscustomobject]@{Path=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash}})
                $binding=[pscustomobject]@{Index=$check.Index;Binary=$check.Binary;Files=$files};$state.Bindings.Add($binding)
            }else{$binding=@($state.Bindings|Where-Object {$_.Index -eq 5})[0]}
            $executions=$null
            if($check.Kind -ne 'compile-only'){
                Assert-MatrixBinding $binding
                $state.Phase='execute';$receiptDir=Join-Path $check.Directory 'execute'
                $state.ProcessAttempts.Add([pscustomobject]@{Index=$check.Index;Phase=$state.Phase;Receipt=(Join-Path $receiptDir 'receipt.json')});Save-MatrixProgress $state $plan.RunRoot
                $run=Invoke-MatrixRecordedProcess $check.Binary $check.Arguments $plan.Environment $check.WorkingDirectory $receiptDir $check.ExecuteLimitSeconds
                Assert-ProcessCompletion $run
                if($run.ShutdownSchema -ne 3 -or $run.ExitCode -ne 0){throw 'Matrix test execution failed'}
                $stdout=@(Get-Content -LiteralPath $run.StdoutPath -ErrorAction Stop|Where-Object {$_ -ne ''})
                $stderr=Get-Content -LiteralPath $run.StderrPath -Raw -ErrorAction Stop
                if($stdout.Count -eq 0 -or $stdout[-1] -cne 'PASS' -or $stderr -match '(?m)^TALENRO_FUZZ_DIAG '){throw 'Matrix PASS/uninstrumented output invalid'}
                if($check.Kind -eq 'fuzz'){
                    $executionMatches=[regex]::Matches($stderr,'(?m)^fuzz:.*execs: ([0-9]+) ');$executions=[UInt64]0
                    if($run.TotalProcesses -lt 2 -or $stderr -notmatch '(?m)^fuzz:.*now fuzzing with 1 workers\r?$' -or $executionMatches.Count -eq 0 -or -not [UInt64]::TryParse($executionMatches[$executionMatches.Count-1].Groups[1].Value,[ref]$executions) -or $executions -eq 0){throw 'Matrix fuzz worker/execution evidence missing'}
                }else{
                    $pattern=if($check.Index -eq 23){'^--- PASS: (TestBoundRead|TestCoordinator)'}else{'^--- PASS: (Test|Example|Fuzz)'}
                    if(@($stdout|Where-Object {$_ -cmatch $pattern}).Count -eq 0){throw 'Matrix matched test execution evidence missing'}
                }
                Assert-MatrixBinding $binding
                $runtimeFiles=@(foreach($file in @((Join-Path $receiptDir 'receipt.json'),$run.StdoutPath,$run.StderrPath)){$path=Get-EnvelopeSampleFile $file;[pscustomobject]@{Path=$path;SHA256=(Get-FileHash -LiteralPath $path).Hash}})
                $state.RuntimeEvidence.Add([pscustomobject]@{Index=$check.Index;Files=$runtimeFiles})
            }
            $state.Checks.Add([pscustomobject]@{Index=$check.Index;Kind=$check.Kind;Target=$check.Target;Passed=$true;RecordedExecutions=$executions});$state.Passed++;$state.Unrun=26-$state.Passed;Save-MatrixProgress $state $plan.RunRoot
        }
    }catch{
        $state.Failure=$_.Exception.Message
        if($state.Current -gt $state.Passed){$state.Failed=1;$state.Unrun=26-$state.Passed-1}
        [pscustomobject]@{Index=$state.Current;Stage=$state.Phase;Error=$state.Failure;NoRetry=$true;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
        throw
    }finally{
        try{
            $after=@(Get-EnvelopeProductManifest);$after|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'after-product.json') -ErrorAction Stop
            Assert-EnvelopeProductUnchanged $before $after
            if((Assert-UninstrumentedEnvelopeToolchain $plan.Toolchain $plan.ReferenceReceipt) -cne $identity){throw 'Matrix copied toolchain drift'}
            foreach($binding in $state.Bindings){Assert-MatrixBinding $binding}
            foreach($evidence in $state.RuntimeEvidence){Assert-MatrixBinding $evidence}
            $state.MatrixPassed=($state.Passed -eq 26 -and $null -eq $state.Failure)
            $state|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'matrix-result.json') -ErrorAction Stop
            if($state.MatrixPassed){Write-Output $state}
        }catch{
            $state.MatrixPassed=$false;$state.FinalizationFailed=$true;$state.Failure=$_.Exception.Message
            [pscustomobject]@{Index=$state.Current;Stage='finalize';PriorStage=$state.Phase;Error=$state.Failure;NoRetry=$true;Acceptance=$false}|ConvertTo-Json|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'stopped.json') -ErrorAction Stop
            try{$state|ConvertTo-Json -Depth 9|Set-Content -LiteralPath (Join-Path $plan.RunRoot 'matrix-result.json') -ErrorAction Stop}catch{}
            throw
        }
    }
}
