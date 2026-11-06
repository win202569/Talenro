param([Parameter(Mandatory)][ValidateSet('Red','Green','Combined')][string]$Phase)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$repo=Split-Path (Split-Path $PSScriptRoot)
$root=Join-Path $repo ('.superpowers/sdd/2026-11-06-c12-shadow-audit-integration-v1/Hooks'+$Phase)
if(Test-Path -LiteralPath $root){throw 'occupied hook phase'}
$null=[IO.Directory]::CreateDirectory($root)
$utf8=[Text.UTF8Encoding]::new($false)
function Save-Hook($path,$value){
 $bytes=$utf8.GetBytes((ConvertTo-Json -InputObject $value -Depth 30))
 $stream=[IO.File]::Open($path,[IO.FileMode]::CreateNew)
 try{$stream.Write($bytes,0,$bytes.Length)}finally{$stream.Dispose()}
}
function Check-Hook($ok,$why){if(-not $ok){throw "HOOK_TEST: $why"}}
function Block-Hook($ast){[scriptblock]::Create($ast.Extent.Text.Trim().TrimStart('{').TrimEnd('}'))}
Save-Hook (Join-Path $root 'claim.json') @{Phase=$Phase;Consumed=$true;Synthetic=$true;Overall='NOT_ACCEPTED'}
$results=[Collections.Generic.List[object]]::new();$current='extract'
try{
 $tokens=$null;$errors=$null
 $ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $repo 'scripts/run-c12-integration.ps1'),[ref]$tokens,[ref]$errors)
 Check-Hook ($errors.Count -eq 0) 'runner parse errors'
 if('ShadowAudit' -cnotin @($ast.ParamBlock.Parameters.Name.VariablePath.UserPath)){throw 'SHADOW_RED: missing switch/hooks'}
 function Function-AST($name){$ast.Find({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $name},$true)}
 $groupAST=Function-AST 'Invoke-C12Group'
 $groupTry=@($groupAST.Body.EndBlock.Statements|Where-Object {$_ -is [Management.Automation.Language.TryStatementAst]})[0]
 $groupCleanup=Block-Hook $groupTry.Finally
 $groupTail=[scriptblock]::Create((@($groupAST.Body.EndBlock.Statements|Where-Object {$_.Extent.StartOffset -gt $groupTry.Extent.EndOffset}|ForEach-Object {$_.Extent.Text}) -join "`n"))
 $startAST=$groupAST.Find({param($n) $n -is [Management.Automation.Language.IfStatementAst] -and $n.Extent.Text.Contains('Start-C12ShadowGroup')},$true)
 Check-Hook ($null -ne $startAST) 'start hook absent'
 $startHook=[scriptblock]::Create($startAST.Extent.Text)
 $returnedAST=$groupAST.Find({param($n) $n -is [Management.Automation.Language.IfStatementAst] -and $n.Extent.Text -match '^if \(\$ShadowAudit\).*shadowNativeReturned = \$true'},$true)
 Check-Hook ($null -ne $returnedAST) 'native return hook absent'
 $returnedHook=[scriptblock]::Create($returnedAST.Extent.Text)
 $top=@($ast.EndBlock.Statements)
 $initAST=@($top|Where-Object {$_ -is [Management.Automation.Language.IfStatementAst] -and $_.Extent.Text.Contains('New-C12ShadowSession')})[0]
 $finalAST=@($top|Where-Object {$_ -is [Management.Automation.Language.IfStatementAst] -and $_.Extent.Text.Contains('Complete-C12ShadowSession')})[0]
 $initHook=[scriptblock]::Create($initAST.Extent.Text);$finalHook=[scriptblock]::Create($finalAST.Extent.Text)
 $mainAST=@($top|Where-Object {$_ -is [Management.Automation.Language.TryStatementAst] -and $_.Extent.Text.Contains('Assert-C12NoInheritedDependencies')})[0]
 $mainHook=[scriptblock]::Create($mainAST.Extent.Text)
 $suiteAST=Function-AST 'Invoke-C12SuiteMode';$suiteDefinition=[scriptblock]::Create($suiteAST.Extent.Text)
 $envNames=@('GOOS','GOARCH','CGO_ENABLED','TALENRO_DATABASE_URL','TALENRO_REDIS_ADDRESS','TALENRO_NATS_URL')
 foreach($current in @('off-valid','off-missing-module','matched','drift','audit-io','formal-failure','group-cleanup','unknown-ownership','suite-roots','suite-cleanup','restore-failure','final-deadline')){
  & {
   $caseRoot=Join-Path $root $current;$invocation=Join-Path $caseRoot 'invocation';$candidateRoot=Join-Path $caseRoot 'candidate'
   $null=[IO.Directory]::CreateDirectory((Join-Path $invocation 'scripts/private'))
   $null=[IO.Directory]::CreateDirectory((Join-Path $candidateRoot 'testdata/c12'))
   if($current -ne 'off-missing-module'){[IO.File]::Copy((Join-Path $repo 'scripts/private/c12-shadow-audit.ps1'),(Join-Path $invocation 'scripts/private/c12-shadow-audit.ps1'))}
   foreach($source in @($invocation,$candidateRoot)){
    $null=[IO.Directory]::CreateDirectory((Join-Path $source 'testdata/c12'))
    [IO.File]::WriteAllText((Join-Path $source 'a.txt'),'abc',$utf8)
    Save-Hook (Join-Path $source 'testdata/c12/shadow-source-manifest.v1.json') @{Schema=1;Mode='SHADOW';Files=@(@{Path='a.txt';SHA256='BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD'})}
   }
   if($current -eq 'drift'){[IO.File]::WriteAllText((Join-Path $invocation 'a.txt'),'changed',$utf8)}
   if($current -like 'suite-*'){[IO.File]::WriteAllText((Join-Path $invocation 'a.txt'),'invocation differs',$utf8)}
   Save-Hook (Join-Path $candidateRoot 'testdata/c12/integration-contracts-schema-authority.v1.json') @{groups=@(@{id='g';profile='base';package='./internal/testinfra';timeout='1m';tests=@('TestSynthetic')})}
   $snapshot=@{};foreach($file in Get-ChildItem -LiteralPath $candidateRoot -Recurse -File){$snapshot[$file.FullName]=(Get-FileHash -LiteralPath $file.FullName).Hash}
   $savedEnv=@{};foreach($name in $envNames){$savedEnv[$name]=[Environment]::GetEnvironmentVariable($name,'Process')}
   $priorGoEnvironment=@{GOOS=$savedEnv.GOOS;GOARCH=$savedEnv.GOARCH;CGO_ENABLED=$savedEnv.CGO_ENABLED}
   if($current -eq 'restore-failure'){$priorGoEnvironment=@{''='invalid-name'}}
   $ShadowAudit=($current -notlike 'off-*');$script:c12RepositoryRoot=$invocation
   $script:c12SuiteDeadline=[datetime]::UtcNow.AddMinutes(5);$script:c12NativeDeadline=$script:c12SuiteDeadline
   $script:c12GroupCleanupBudget=[timespan]::FromSeconds(1);$script:c12ProfileAllowances=@{base=[timespan]::FromMinutes(1);'authority-v7-pitr'=[timespan]::FromMinutes(1)}
   $Suite='batch01';$Timeout='120m';$selectedMode=if($current -like 'suite-*'){'Suite'}else{'Focused'}
   $scriptExitCode=0;$outerThrown=$false
   $script:hookUnexpectedAuditCalls=0
   function New-C12ShadowSession {$script:hookUnexpectedAuditCalls++;throw 'OFF audit factory reached'}
   function Start-C12ShadowGroup {$script:hookUnexpectedAuditCalls++;throw 'OFF audit group reached'}
   function Complete-C12ShadowGroup {$script:hookUnexpectedAuditCalls++;throw 'OFF audit completion reached'}
   function Complete-C12ShadowSession {$script:hookUnexpectedAuditCalls++;throw 'OFF audit finalization reached'}
   function Assert-C12NoInheritedDependencies {}
   function Clear-C12InheritedCapabilities {}
   function Close-C12AllPreparedArtifacts {}
   function Remove-C12Container {throw 'synthetic cleanup failure'}
   function ConvertFrom-C12Duration {param($Value) if($Value -eq '120m'){[timespan]::FromMinutes(120)}else{[timespan]::FromMinutes(1)}}
   function New-C12CandidateSnapshot {return @{Tree=('a'*40);OwnerRoot=$candidateRoot;Ownership=@{}}}
   function New-C12CandidateMaterialization {return @{Root=$candidateRoot}}
   function Invoke-C12TrustedValidator {}
   function Assert-C12ExecutionPlan {}
   function Assert-C12CandidateSnapshotClean {
    Check-Hook (@(Get-ChildItem -LiteralPath $candidateRoot -Recurse -File).Count -eq $snapshot.Count) 'snapshot gained files'
    foreach($path in $snapshot.Keys){Check-Hook ((Get-FileHash -LiteralPath $path).Hash -ceq $snapshot[$path]) 'snapshot content changed'}
   }
   function Remove-C12CandidateMaterialization {}
   function Remove-C12CandidateSnapshot {if($current -eq 'suite-cleanup'){throw 'synthetic outer cleanup failure'}}
   function Invoke-C12Group {
    param($GroupID='g',$GroupProfile='base',$Package='./internal/testinfra',$RunPattern,$GroupTimeout,$ExpectedTests,$CandidateTree='',$AbsoluteDeadline)
    $runSuffix='0123456789abcdef0123456789abcdef';$groupDeadline=$script:c12SuiteDeadline;$priorNativeDeadline=$script:c12NativeDeadline
    . $startHook
    $primaryFailure=if($current -in @('formal-failure','unknown-ownership')){'synthetic formal failure'}else{$null}
    $cleanupFailures=@();$preparedInitializer=$null;$preparedArtifactRoot=$null;$pitrRunRoot=$null;$resources=@();$pitrNonce='';$pitrHMACKey=''
    if($current -eq 'group-cleanup'){$resources=@(@{Kind='postgres'})}
    if($current -ne 'unknown-ownership'){. $returnedHook}
    . $groupCleanup
    . $groupTail
   }
   function Invoke-C12FocusedMode {Invoke-C12Group}
   . $suiteDefinition
   try{
    . $initHook
    if($ShadowAudit){
     Check-Hook ($null -ne $script:c12ShadowSession -and -not $script:c12ShadowSession.Synthetic) 'real default changed'
     # The actual init is verified above. Subsequent fixture maps are explicitly synthetic.
     $script:c12ShadowSession.Synthetic=$true
     if($current -eq 'audit-io'){
      Save-Hook (Join-Path $script:c12ShadowSession.EvidenceRoot 'run.json') @{Sentinel='preserve'}
      function Write-Warning {param($Message) throw 'synthetic warning failure'}
     }
    }
    try{. $mainHook}catch{$outerThrown=$true}
    if(-not $outerThrown){
     if($current -eq 'final-deadline'){$script:c12ShadowDeadline=[datetime]::UtcNow.AddSeconds(-1)}
     . $finalHook
    }
    $expectedExit=if($current -in @('formal-failure','group-cleanup','unknown-ownership','suite-cleanup')){1}else{0}
    Check-Hook ($scriptExitCode -eq $expectedExit) "formal exit $current"
    Check-Hook ($outerThrown -eq ($current -eq 'restore-failure')) 'restoration throw changed'
    $state='DISABLED'
    if(-not $ShadowAudit){Check-Hook ($script:hookUnexpectedAuditCalls -eq 0 -and -not(Test-Path -LiteralPath (Join-Path $invocation '.superpowers'))) 'off call or side effect'}
    else{
     $runPath=Join-Path $script:c12ShadowSession.EvidenceRoot 'run.json'
     if($current -in @('restore-failure','final-deadline')){Check-Hook (-not(Test-Path -LiteralPath $runPath)) 'invalid final summary';$state='INCOMPLETE'}
     elseif($current -eq 'audit-io'){$data=Get-Content -LiteralPath $runPath -Raw|ConvertFrom-Json;Check-Hook ($data.Sentinel -ceq 'preserve') 'occupied summary overwritten';$state='INCOMPLETE'}
     else{
      $data=Get-Content -LiteralPath $runPath -Raw|ConvertFrom-Json
      $want=if($current -eq 'drift'){'DRIFT'}elseif($current -in @('group-cleanup','unknown-ownership','suite-cleanup')){'INCOMPLETE'}else{'MATCHED'}
      Check-Hook ($data.State -ceq $want -and $data.FormalExitCode -eq $expectedExit) "shadow state $current"
      $state=$data.State
     }
     if($current -like 'suite-*'){
      Check-Hook ($script:c12ShadowSession.Groups[0].SourceRoot -ceq $candidateRoot -and $script:c12ShadowSession.EvidenceRoot.StartsWith($invocation)) 'Suite root binding'
      Assert-C12CandidateSnapshotClean
     }
     if($current -in @('group-cleanup','unknown-ownership')){Check-Hook ($null -eq $script:c12ShadowSession.Groups[0].After) 'unsafe after scan'}
    }
    $results.Add(@{Case=$current;Matched=$true;FormalExit=$scriptExitCode;RestorationThrew=$outerThrown;ShadowState=$state;Synthetic=$true})
    Write-Output "PASS: $current"
   }finally{foreach($name in $savedEnv.Keys){[Environment]::SetEnvironmentVariable($name,$savedEnv[$name],'Process')}}
  }
 }
 Save-Hook (Join-Path $root 'result.json') @{Status='SHADOW_HOOKS_VERIFIED';Results=$results.ToArray();CoverageBlocked=@('reparse');Overall='NOT_ACCEPTED';Synthetic=$true}
 'SHADOW_HOOKS_VERIFIED'
}catch{
 Save-Hook (Join-Path $root 'result.json') @{Status='STOPPED';Case=$current;Reason=$_.Exception.Message;Stack=$_.ScriptStackTrace;Results=$results.ToArray();NoRetry=$true;Overall='NOT_ACCEPTED'}
 throw
}
