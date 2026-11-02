# Finite pin regression. Definitions only; no runtime/profile/module bodies.
function PB-Need([bool]$Good,[string]$Label){if(-not $Good){throw ('Pin binding regression: '+$Label)}}
function PB-Error($Record){
    if($null -eq $Record){return $null}
    $position=$null;if($null -ne $Record.InvocationInfo){$position=$Record.InvocationInfo.PositionMessage}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$position;Stack=$Record.ScriptStackTrace}
}
function PB-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{$out=@(& $Body);PB-Need ($out.Count -eq 0) ('unexpected output '+$Name);$State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})}
    catch{if($null -eq $State.RootError){$State.RootError=$_};$State.Stopped=$true;$State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})}
}
function PB-Reject([scriptblock]$Body,[string]$Message){
    $caught=$null;$out=@();try{$out=@(& $Body)}catch{$caught=$_}
    PB-Need ($null -ne $caught -and $out.Count -eq 0) 'must reject'
    if($caught.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $caught.Exception.Message -cne $Message -or $caught.FullyQualifiedErrorId -cne $Message){throw $caught}
}
function Invoke-ODPinBindingRegression {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][ValidateSet('Red','Green')][string]$Phase)
    Set-StrictMode -Version Latest;$ErrorActionPreference='Stop'
    $state=[pscustomobject]@{Groups=[Collections.Generic.List[object]]::new();RootError=$null;RecordingError=$null;Stopped=$false}
    $names=@('ValidPinnedHash','MissingPinRejected','AbsentPinRejected','MalformedHashRejected','ConflictingPinsRejected','ActualHashMismatchRejected','IdenticalPinsAccepted')
    $testHash=$null;$inputHash=$null;$before=[ordered]@{};$after=[ordered]@{};$code=2
    try{
        $leaf=Split-Path (Get-Command Invoke-ODPinBindingRegression -CommandType Function).ScriptBlock.File
        $repo=[IO.Path]::GetFullPath((Join-Path $leaf '../../../..'));$packet=Join-Path $repo '.superpowers/sdd/2026-11-01-offline-i1-i2-single-owner'
        $testPath=Join-Path $leaf pin-binding.tests.ps1;$inputPath=Join-Path $packet pin-binding.input.json
        $testHash=(Get-FileHash $testPath).Hash;$inputHash=(Get-FileHash $inputPath).Hash
        $binding=Get-Content -Raw $inputPath|ConvertFrom-Json -AsHashtable
        PB-Need ($binding.Schema -eq 1 -and $binding.TestSHA256 -ceq $testHash) 'frozen input/test'
        foreach($key in $binding.HistoricalFiles.Keys){PB-Need ((Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash -ceq $binding.HistoricalFiles[$key]) ('history '+$key)}
        $expected=$binding.BaselineSources;$intermediateHash=$null
        if($Phase -ceq 'Green'){
            $intermediatePath=Join-Path $packet pin-binding-intermediate.input.json;$intermediateHash=(Get-FileHash $intermediatePath).Hash
            $intermediate=Get-Content -Raw $intermediatePath|ConvertFrom-Json -AsHashtable
            PB-Need ($intermediate.Schema -eq 1 -and $intermediate.InputSHA256 -ceq $inputHash) 'intermediate binding';$expected=$intermediate.SourceSHA256
        }
        foreach($key in $expected.Keys){$before[$key]=(Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash;PB-Need ($before[$key] -ceq $expected[$key]) ('source '+$key)}
        $t=$null;$e=$null;$tree=[Management.Automation.Language.Parser]::ParseFile((Join-Path $leaf source-model.ps1),[ref]$t,[ref]$e)
        PB-Need ($e.Count -eq 0) 'source parse'
        $definitions=@($tree.FindAll({param($n)$n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq 'Get-ODPinnedSHA256'},$true))
        PB-Need ($definitions.Count -eq 1) 'unique actual definition';. ([scriptblock]::Create($definitions[0].Extent.Text))
        $manifestPath=Join-Path $packet writer-projection-repair.frozen.json
        PB-Need ((Get-FileHash $manifestPath).Hash -ceq $binding.ParentSHA256) 'actual parent'
        $manifest=Get-Content -Raw $manifestPath|ConvertFrom-Json -AsHashtable
        $path=Join-Path $leaf record-boundary.tests.ps1;$actual=(Get-FileHash $path).Hash
        $selected=@($manifest.Pins|Where-Object {$_.VerificationPath -ieq $path});PB-Need ($selected.Count -eq 1 -and $selected[0].SHA256 -ceq $actual) 'actual pinned source'
        $pin=$selected[0];$wrong='0'*64;PB-Need ($actual -cne $wrong) 'independent wrong hash'
        PB-Case $names[0] {$value=Get-ODPinnedSHA256 $path $manifest;PB-Need ($value -is [string] -and $value -cmatch '^[A-F0-9]{64}$' -and $value -ceq $actual) 'valid returns independent hash'} $state
        if($Phase -ceq 'Red'){
            PB-Need ($null -ne $state.RootError -and $state.RootError.Exception.GetType().FullName -ceq 'System.Management.Automation.PropertyNotFoundException' -and $state.RootError.FullyQualifiedErrorId -ceq 'PropertyNotFoundStrict' -and $state.RootError.Exception.Message -ceq "The property 'SHA256' cannot be found on this object. Verify that the property exists.") 'exact observed RED'
            for($i=1;$i -lt $names.Count;$i++){PB-Case $names[$i] {} $state};$code=1
        }else{
            PB-Case $names[1] {PB-Reject {Get-ODPinnedSHA256 $path ([ordered]@{Pins=@()})} 'OD: source model input not pinned'} $state
            PB-Case $names[2] {$v=[ordered]@{VerificationPath=$path;SHA256=$actual;ExpectedState='Absent'};PB-Reject {Get-ODPinnedSHA256 $path ([ordered]@{Pins=@($v)})} 'OD: source model pin conflict'} $state
            PB-Case $names[3] {$v=[ordered]@{VerificationPath=$path;SHA256='not-a-hash';ExpectedState='Present'};PB-Reject {Get-ODPinnedSHA256 $path ([ordered]@{Pins=@($v)})} 'OD: source model pin conflict'} $state
            PB-Case $names[4] {$v=[ordered]@{VerificationPath=$path;SHA256=$wrong;ExpectedState='Present'};PB-Reject {Get-ODPinnedSHA256 $path ([ordered]@{Pins=@($pin,$v)})} 'OD: source model pin conflict'} $state
            PB-Case $names[5] {$v=[ordered]@{VerificationPath=$path;SHA256=$wrong;ExpectedState='Present'};PB-Reject {Get-ODPinnedSHA256 $path ([ordered]@{Pins=@($v)})} 'OD: source model input changed'} $state
            PB-Case $names[6] {$v=[ordered]@{VerificationPath=$path;SHA256=$actual;ExpectedState='Present'};PB-Need ((Get-ODPinnedSHA256 $path ([ordered]@{Pins=@($pin,$v)})) -ceq $actual) 'identical duplicates accepted'} $state
            $code=if($state.Stopped){2}else{0}
        }
        foreach($key in $expected.Keys){$after[$key]=(Get-FileHash -LiteralPath (Join-Path $repo $key)).Hash;PB-Need ($after[$key] -ceq $before[$key]) ('post source '+$key)}
        PB-Need ((Get-FileHash $testPath).Hash -ceq $testHash -and (Get-FileHash $inputPath).Hash -ceq $inputHash -and (Get-FileHash $path).Hash -ceq $actual -and (Get-FileHash $manifestPath).Hash -ceq $binding.ParentSHA256) 'immutable inputs'
        if($Phase -ceq 'Green'){PB-Need ((Get-FileHash $intermediatePath).Hash -ceq $intermediateHash) 'intermediate unchanged'}
    }catch{if($null -eq $state.RootError){$state.RootError=$_}else{$state.RecordingError=$_};$code=2}
    foreach($name in $names){if($name -cnotin @($state.Groups|ForEach-Object Name)){$state.Groups.Add([ordered]@{Name=$name;Status='NotRun'})}}
    return [ordered]@{Schema=1;Phase=$Phase;TestSHA256=$testHash;InputSHA256=$inputHash;SourceSHA256=$before;PostSourceSHA256=$after;Groups=$state.Groups.ToArray();PASS=@($state.Groups|Where-Object Status -CEQ PASS).Count;FAIL=@($state.Groups|Where-Object Status -CEQ FAIL).Count;NotRun=@($state.Groups|Where-Object Status -CEQ NotRun).Count;ExitCode=$code;RootError=(PB-Error $state.RootError);RecordingError=(PB-Error $state.RecordingError)}
}
