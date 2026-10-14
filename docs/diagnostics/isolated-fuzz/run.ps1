Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'isolation.ps1')
if(-not ('IsolatedFuzzProcess' -as [type])) { Add-Type -Path (Join-Path $PSScriptRoot 'process-native.cs') }
function Invoke-RecordedProcess([string]$Executable,[string[]]$Arguments,[hashtable]$Environment,[string]$ReceiptDir,[int]$LimitSeconds) {
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $allowed=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $receipt=Get-IsolatedFullPath $ReceiptDir
    $exe=Get-IsolatedFullPath $Executable
    if(-not (Test-IsolatedWithin $receipt $allowed) -or (Test-Path -LiteralPath $receipt)) {throw 'Receipt scope invalid or already exists'}
    if(-not (Test-Path -LiteralPath $exe -PathType Leaf) -or $LimitSeconds -lt 1 -or $LimitSeconds -gt 600) {throw 'Executable or bound invalid'}
    if($Environment.ContainsKey('GOROOT')) {
        $goroot=Get-IsolatedFullPath $Environment.GOROOT
        if(-not (Test-IsolatedWithin $goroot $allowed) -or -not (Test-Path -LiteralPath (Join-Path $goroot 'bin/go.exe') -PathType Leaf)) {throw 'Invalid copy GOROOT'}
    }
    New-Item -ItemType Directory -Path $receipt -ErrorAction Stop | Out-Null
    $stdout=Join-Path $receipt 'stdout.txt';$stderr=Join-Path $receipt 'stderr.txt'
    try {
        $native=[IsolatedFuzzProcess]::Run($exe,$Arguments,$Environment,(Split-Path -Parent $receipt),$stdout,$stderr,$LimitSeconds)
        $result=[pscustomobject]@{ExitCode=$native.ExitCode;TimedOut=$native.TimedOut;DescendantsRemaining=$native.DescendantsRemaining;CleanupVerified=$native.CleanupVerified;ActiveAfterCleanup=$native.ActiveAfterCleanup;TotalProcesses=$native.TotalProcesses;RootId=$native.RootId;RootExited=$native.RootExited;ActiveAtRootWait=$native.ActiveAtRootWait;MembersAtRootWait=$native.MembersAtRootWait;MembersAfterCleanup=$native.MembersAfterCleanup;TerminationRequested=$native.TerminationRequested;StartUTC=$native.StartUTC;EndUTC=$native.EndUTC;StdoutPath=$stdout;StderrPath=$stderr;Executable=$exe;Arguments=$Arguments;Environment=$Environment}
        $result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $receipt 'receipt.json')
        return $result
    } catch {
        [pscustomobject]@{StartUTC=[DateTime]::UtcNow.ToString('o');LaunchError=$_.Exception.GetType().FullName;CleanupVerified=$false} | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $receipt 'launch-error.json')
        throw
    }
}
function Assert-CopiedCompileEvidence([string]$Log,[string]$Toolchain) {
    $compiler=Join-Path $Toolchain 'pkg/tool/windows_amd64/compile.exe'
    foreach($package in @('context','internal/fuzz')) {
        $source=Join-Path $Toolchain ('src/'+$package+'/'+$(if($package -eq 'context'){'context.go'}else{'fuzz.go'}))
        $found=$false
        foreach($line in ($Log -split '\r?\n')) {
            $tokens=@([regex]::Matches($line,'"(?:\\.|[^"\\])*"|[^\s"]+') | ForEach-Object {
                $value=$_.Value
                if($value.StartsWith('"')) {$value=$value.Substring(1,$value.Length-2).Replace('\\','\').Replace('\"','"')}
                $value
            })
            if($tokens.Count -lt 4 -or $tokens[0] -cne $compiler) {continue}
            $hasPackage=$false
            for($i=1;$i -lt $tokens.Count-1;$i++) {if($tokens[$i] -ceq '-p' -and $tokens[$i+1] -ceq $package){$hasPackage=$true}}
            if($hasPackage -and $tokens -ccontains $source) {$found=$true;break}
        }
        if(-not $found) {throw 'Actual copied standard-library compiler/source tokens missing'}
    }
}
function Assert-ArmResult([string]$Arm,[object]$Receipt,[object[]]$Trace) {
    if($Receipt.TimedOut -or $Receipt.DescendantsRemaining -ne 0 -or -not $Receipt.CleanupVerified -or $Receipt.ActiveAfterCleanup -ne 0) {throw 'Unfinished or timed-out process group'}
    if($Arm -eq 'baseline') {if($Receipt.ExitCode -ne 0){throw 'Baseline failed'};return}
    if($Trace.Count -lt 2) {throw 'Trace incomplete'}
    $expectedMode=if($Arm -eq 'crash'){'observe'}else{$Arm}
    $seq=0
    foreach($row in $Trace) {
        if($row.mode -ne $expectedMode -or $row.schema -ne 1 -or $row.seq -ne (++$seq) -or $row.watchdog) {throw 'Trace mode/schema/order/watchdog failure'}
        foreach($key in @('incoming','childErr','parentErr','oldError','newError')) {if($row.$key -like 'unsupported:*'){throw 'Unsupported identity in actual trace'}}
        if(-not $row.parentRead -and $row.parentErr -ne 'unread') {throw 'Unread operand misrepresented'}
        if($row.event -eq 'stop' -and -not $row.stoppingRead) {throw 'Stop-state read provenance missing'}
        if($row.event -eq 'return' -and ($row.childErr -ne 'unread' -or $row.oldError -ne 'unread' -or $row.stoppingRead)) {throw 'Return invents unobserved state'}
    }
    $final=$Trace[-1]
    if($final.event -ne 'return' -or $final.activeWorkers -ne 0) {throw 'Final worker/return trace missing'}
    $stops=@($Trace | Where-Object {$_.event -eq 'stop'})
    if(-not ($stops | Where-Object {$_.source -eq 'worker' -and $_.activeWorkers -eq 1 -and $_.stopping})) {throw 'Worker shutdown evidence missing'}
    switch($Arm) {
        'parent-first' {if($Receipt.ExitCode -eq 0 -or $final.newError -ne 'deadline' -or -not ($stops | Where-Object {$_.source -eq 'parent-done' -and $_.incoming -eq 'deadline' -and $_.childErr -eq 'nil' -and $_.gateEntered -and $_.newError -eq 'deadline'})){throw 'Forced deadline mechanism not established'}}
        'synchronized' {if($Receipt.ExitCode -ne 0 -or $final.newError -ne 'nil' -or -not ($stops | Where-Object {$_.source -eq 'parent-done' -and $_.childErr -eq 'deadline' -and $_.propagated -and $_.newError -eq 'nil'})){throw 'Synchronized deadline control failed'}}
        'crash' {if($Receipt.ExitCode -eq 0 -or $final.newError -notlike 'error:*' -or -not ($stops | Where-Object {$_.source -eq 'crasher' -and $_.newError -like 'error:*'})){throw 'Real worker crasher not preserved'}}
        'cancel' {if($Receipt.ExitCode -eq 0 -or $final.newError -ne 'canceled' -or -not ($stops | Where-Object {$_.incoming -eq 'canceled' -and $_.childErr -eq 'nil' -and $_.gateEntered})){throw 'Selected early cancellation not established'}}
        'observe' {if($Receipt.ExitCode -ne 0){throw 'Observation failed; preserve evidence, do not retry'}}
        default {throw 'Unknown diagnostic arm'}
    }
}
function Get-DiagnosticAttemptPaths([string]$RunRoot,[string]$Arm,[string]$BaselineAttempt='') {
    $suffix=''
    if($BaselineAttempt) {
        if($Arm -ne 'baseline' -or $BaselineAttempt -cnotmatch '^retry-[0-9]{3}$') {throw 'Invalid baseline attempt identity'}
        $prior=Join-Path $RunRoot 'arms/baseline'
        if(-not (Test-Path -LiteralPath $prior -PathType Container) -or (Test-Path -LiteralPath (Join-Path $prior 'execute'))) {throw 'Baseline retry only supports a failed preparation, not a completed fuzz run'}
        foreach($attempt in Get-ChildItem -LiteralPath (Split-Path -Parent $prior) -Directory -Force) {
            if($attempt.Name -cmatch '^baseline-retry-[0-9]{3}$') {
                $validated=Get-IsolatedFullPath $attempt.FullName
                if(Test-Path -LiteralPath (Join-Path $validated 'execute')) {throw 'A baseline retry already launched fuzz; no further retry allowed'}
            }
        }
        $suffix='-'+$BaselineAttempt
    }
    $armDir=Join-Path $RunRoot ('arms/'+$Arm+$suffix)
    $buildDir=Join-Path $RunRoot ('binaries/'+$(if($Arm -eq 'baseline'){'plain'+$suffix}else{'diagnostic'}))
    if(Test-Path -LiteralPath $armDir) {throw 'Arm already exists'}
    if($BaselineAttempt -and (Test-Path -LiteralPath $buildDir)) {throw 'Retry build identity already exists'}
    [pscustomobject]@{ArmDir=$armDir;BuildDir=$buildDir}
}
function Invoke-DiagnosticArm([string]$RunRoot,[string]$Arm,[string]$BaselineAttempt='') {
    $paths=Get-DiagnosticAttemptPaths $RunRoot $Arm $BaselineAttempt
    if($Arm -notin @('baseline','observe','parent-first','synchronized','crash','cancel')) {throw 'Unknown arm'}
    $workspace=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd('\','/')
    $allowed=Join-Path $workspace '.superpowers/diagnostics/2026-10-14-isolated-fuzz-coordinator'
    $root=Get-IsolatedFullPath $RunRoot
    if(-not (Test-IsolatedWithin $root $allowed)) {throw 'Run root outside owned scope'}
    $tool=Get-IsolatedFullPath (Join-Path $root 'toolchain')
    $copy=Get-Content -Raw -LiteralPath (Join-Path $root 'copy-receipt.json') | ConvertFrom-Json
    if($copy.Toolchain -ne $tool -or $copy.BeforeManifest.Count -ne 11536) {throw 'Verified copy provenance missing'}
    $go=Join-Path $tool 'bin/go.exe'
    $goPin=@($copy.BeforeManifest | Where-Object {$_.Path -eq 'bin/go.exe'})
    if($goPin.Count -ne 1 -or (Get-FileHash -LiteralPath $go).Hash -ne $goPin[0].SHA256) {throw 'Copied go binary drift'}
    $sourcePins=@{'src/context/context.go'='765F99F7CC0710B0117FFA183E432524F429C21496FCBAD9B63FB9886FDBAAC7';'src/internal/fuzz/fuzz.go'='D9DE12569D5C18FB629C8E79A4D629EB2985B719E4FBFE30C3F3CE23FC44FEB5'}
    if($Arm -eq 'baseline') {
        foreach($path in $sourcePins.Keys) {if((Get-FileHash -LiteralPath (Join-Path $tool $path)).Hash -ne $sourcePins[$path]) {throw 'Baseline source is instrumented or drifted'}}
        if(Test-Path -LiteralPath (Join-Path $root 'patch-receipt.json')) {throw 'Baseline cannot run after instrumentation'}
    } else {
        $patchRecord=Get-Content -Raw -LiteralPath (Join-Path $root 'patch-receipt.json') | ConvertFrom-Json
        if($patchRecord.Toolchain -ne $tool -or $patchRecord.PatchHash -ne (Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'coordinator.patch')).Hash) {throw 'Patch provenance drift'}
        foreach($row in $patchRecord.InstalledHashes.PSObject.Properties) {if((Get-FileHash -LiteralPath (Join-Path $tool $row.Name)).Hash -ne $row.Value) {throw 'Instrumented source drift'}}
    }
    $armDir=Get-IsolatedFullPath $paths.ArmDir
    New-Item -ItemType Directory -Path $armDir -ErrorAction Stop | Out-Null
    foreach($dir in @('runtime-cache','fuzz-cache','tmp')) {New-Item -ItemType Directory -Path (Join-Path $armDir $dir) | Out-Null}
    $mode=if($Arm -eq 'baseline'){''}elseif($Arm -eq 'crash'){'observe'}else{$Arm}
    $environment=@{GOROOT=$tool;GOCACHE=(Join-Path $armDir 'runtime-cache');GOTOOLCHAIN='local';GOENV='off';GOPROXY='off';GOSUMDB='off';GOAUTH='off';GOVCS='all:off';GO111MODULE='off';GOFLAGS='';GOOS='windows';GOARCH='amd64';TEMP=(Join-Path $armDir 'tmp');TMP=(Join-Path $armDir 'tmp');GOTMPDIR=(Join-Path $armDir 'tmp');TALENRO_FUZZ_DIAG_MODE=$mode;TALENRO_FUZZ_FIXTURE=$(if($Arm -eq 'crash'){'crash'}else{'normal'})}
    $version=Invoke-RecordedProcess $go @('version') $environment (Join-Path $armDir 'version') 60
    if($version.ExitCode -ne 0 -or (Get-Content -Raw -LiteralPath $version.StdoutPath).Trim() -ne 'go version go1.26.5 windows/amd64') {throw 'Pinned version verification failed'}
    $origin=Invoke-RecordedProcess $go @('env','GOROOT','GOTOOLDIR','GOEXPERIMENT') $environment (Join-Path $armDir 'origin') 60
    $originLines=@(Get-Content -LiteralPath $origin.StdoutPath)
    if($origin.ExitCode -ne 0 -or $originLines.Count -lt 2 -or $originLines[0] -ne $tool -or -not $originLines[1].StartsWith($tool+'\',[StringComparison]::OrdinalIgnoreCase)) {throw 'Copied compiler origin verification failed'}
    $variant=if($Arm -eq 'baseline'){'plain'}else{'diagnostic'}
    $buildDir=Get-IsolatedFullPath $paths.BuildDir
    $binary=Join-Path $buildDir 'fixture.exe'
    $fixtureHash=(Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'fixture_test.go.txt')).Hash
    if(-not (Test-Path -LiteralPath $buildDir)) {
        New-Item -ItemType Directory -Path $buildDir -ErrorAction Stop | Out-Null
        foreach($dir in @('cache','tmp')) {New-Item -ItemType Directory -Path (Join-Path $buildDir $dir) | Out-Null}
        $fixture=Join-Path $buildDir 'fixture_test.go'
        Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'fixture_test.go.txt') -Destination $fixture
        $buildEnvironment=$environment.Clone();$buildEnvironment.GOCACHE=Join-Path $buildDir 'cache';$buildEnvironment.GOTMPDIR=Join-Path $buildDir 'tmp';$buildEnvironment.TEMP=$buildEnvironment.GOTMPDIR;$buildEnvironment.TMP=$buildEnvironment.GOTMPDIR
        $build=Invoke-RecordedProcess $go @('test','-c','-x','-p=1','-count=1','-run=^$','-fuzz=^FuzzDiagnostic$','-fuzztime=5s','-parallel=1','-timeout=30s',('-o='+$binary),$fixture) $buildEnvironment (Join-Path $buildDir 'build') 600
        if($build.ExitCode -ne 0 -or $build.TimedOut -or -not $build.CleanupVerified -or $build.DescendantsRemaining -ne 0) {throw 'Fixture build failed; do not retry arm'}
        $buildLog=Get-Content -Raw -LiteralPath $build.StderrPath
        Assert-CopiedCompileEvidence $buildLog $tool
        [pscustomobject]@{FixtureHash=$fixtureHash;BinaryHash=(Get-FileHash -LiteralPath $binary).Hash;Variant=$variant;GOROOT=$tool;CompilerEvidence=$build.StderrPath} | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $buildDir 'binary-receipt.json')
    }
    $binaryRecord=Get-Content -Raw -LiteralPath (Join-Path $buildDir 'binary-receipt.json') | ConvertFrom-Json
    if($binaryRecord.FixtureHash -ne $fixtureHash -or $binaryRecord.BinaryHash -ne (Get-FileHash -LiteralPath $binary).Hash -or $binaryRecord.GOROOT -ne $tool -or $binaryRecord.Variant -ne $variant) {throw 'Binary provenance invalid'}
    $arguments=@('-test.run=^$','-test.fuzz=^FuzzDiagnostic$','-test.fuzztime=5s','-test.parallel=1','-test.timeout=30s','-test.count=1',('-test.fuzzcachedir='+(Join-Path $armDir 'fuzz-cache')))
    $receipt=Invoke-RecordedProcess $binary $arguments $environment (Join-Path $armDir 'execute') 60
    $trace=@(Get-Content -LiteralPath $receipt.StderrPath | Where-Object {$_.StartsWith('TALENRO_FUZZ_DIAG ')} | ForEach-Object {$_.Substring(18) | ConvertFrom-Json})
    $failure=$null
    try {
        if($receipt.TotalProcesses -lt 2) {throw 'No actual worker process recorded'}
        if($Arm -eq 'crash' -and -not (Get-Content -Raw -LiteralPath $receipt.StdoutPath).Contains('isolated-fixture-crash') -and -not (Get-Content -Raw -LiteralPath $receipt.StderrPath).Contains('isolated-fixture-crash')) {throw 'Injected crash marker absent'}
        Assert-ArmResult $Arm $receipt $trace
    } catch {$failure=$_.Exception.Message}
    [pscustomobject]@{Arm=$Arm;ValidationPassed=($null -eq $failure);ValidationFailure=$failure;Receipt=$receipt;Trace=$trace;FixtureHash=$fixtureHash;BinaryHash=$binaryRecord.BinaryHash} | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $armDir 'arm-result.json')
    if($receipt.TimedOut -or $receipt.DescendantsRemaining -ne 0 -or -not $receipt.CleanupVerified -or $receipt.ActiveAfterCleanup -ne 0 -or @($trace | Where-Object {$_.watchdog}).Count -gt 0) {throw 'Safety/cleanup/watchdog gate failed; stop later arms'}
    [pscustomobject]@{Arm=$Arm;ValidationPassed=($null -eq $failure);ValidationFailure=$failure;ExitCode=$receipt.ExitCode;TraceCount=$trace.Count;TotalProcesses=$receipt.TotalProcesses;ResultPath=(Join-Path $armDir 'arm-result.json')}
}
