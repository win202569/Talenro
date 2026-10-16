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
