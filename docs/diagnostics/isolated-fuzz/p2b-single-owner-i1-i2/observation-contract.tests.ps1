function Get-ODHistoricalContext([string]$Profile){
    $historical=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration'
    return [ordered]@{Repo=$repo;PacketRoot=$historical;ProfileRoot=[IO.Path]::GetFullPath((Join-Path $historical ('profiles/'+$Profile)));Oracle=$oracle;Sources=[ordered]@{Sinks=(Join-Path $repo 'scripts/private/offline-diagnostics/sinks.ps1');Core=(Join-Path $repo 'docs/diagnostics/isolated-fuzz/p2b-entry-runtime-migration/core.tests.ps1')}}
}
function Invoke-ODContractTests {
    . (Join-Path $PSScriptRoot observation-contract.ps1)
    $available=$true
    foreach($name in @('Assert-ODProfileObservations','Assert-ODProfileReport')){
        if($null -eq (Get-Command $name -CommandType Function -ErrorAction Ignore)){$available=$false}
    }
    if(-not $available){
        $groups=@(foreach($name in $oracle.ContractGroupNames){[ordered]@{Name=$name;Status=$(if($name -ceq 'ContractAPIAvailable'){'FAIL'}else{'NotRun'})}})
        return [pscustomobject][ordered]@{Schema=1;PASS=0;FAIL=1;NotRun=51;ExitCode=1;Groups=$groups}
    }
    LS-Result ContractAPIAvailable $true ([ordered]@{Verified=$true})
    $historical=Join-Path $repo '.superpowers/sdd/2026-10-31-offline-diagnostic-runtime-migration'
    $refs=@(
        @('runtime-green','entry-migration-core','58B2D81047526EFB2BEE3FD8155DFAF493FC3FD6ECAAE32B0E7ED19C7DB66EA0'),
        @('lifecycle','entry-migration-owned','3F7BDEE2B22F3C91EB5957259F18E1DF96E0CFD18A3F5E692D54C26460A51911'),
        @('continuation','entry-migration-live','D6729F1EBB54435B426013704B1AFD754E82F2EE7F6C725F520C02A56C532726')
    )
    $frames=[Collections.Generic.List[object]]::new()
    foreach($ref in $refs){
        $path=Join-Path $historical ($ref[0]+'.evidence.json')
        LS-Need ((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -ceq $ref[2]) 'historical baseline original hash'
        $e=Get-Content -LiteralPath $path -Raw | ConvertFrom-Json -AsHashtable
        $returns=@($e.Observations|Where-Object {$_.Name -ceq 'ProfileReturn'})
        LS-Need ($returns.Count -eq 1 -and $returns[0].Value.Profile -ceq $ref[1]) 'baseline canonical profile'
        $frames.Add($returns[0].Value.Observations)
    }
    for($caseIndex=0;$caseIndex -lt 23;$caseIndex++){
        $vector=if($caseIndex -lt 3){$caseIndex+1}else{$caseIndex-2}
        $profileIndex=if($caseIndex -lt 3){$caseIndex}elseif($vector -eq 16){2}elseif($vector -cin @(17,18)){1}else{0}
        $profile=$refs[$profileIndex][1]
        $rows=ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $frames[$profileIndex] -Depth 60)
        if($caseIndex -ge 3){
            switch($vector){
                1 {$rows=@($rows[0..1]) + @($rows[3..21])}
                2 {$rows=@($rows[0..2]) + @($rows[2]) + @($rows[3..21])}
                3 {$tmp=$rows[2];$rows[2]=$rows[3];$rows[3]=$tmp}
                4 {$rows[2].Name='Unknown'}
                5 {$rows[6].Value.Result['Extra']=0}
                6 {$null=$rows[6].Value.Result.Remove('Error')}
                7 {$rows[6].Value.Result.Error.Type='System.Exception'}
                8 {$rows[6].Value.Result.Error.ErrorId='UnrelatedError'}
                9 {$rows[6].Value.Result.Error.Message='unrelated'}
                10 {$rows[6].Value.View.Status='Active'}
                11 {$rows[6].Value.Result.Ok=$true}
                12 {$rows[6].Value.Result.Value=[ordered]@{}}
                13 {$null=$rows[1].Value.Result.Remove('Value')}
                14 {$rows[14].Value.View.Sinks.ConsoleOut.State='Healthy'}
                15 {$rows[13].Value.SecondaryErrors=@($rows[13].Value.SecondaryErrors|Select-Object -First 7)}
                16 {$rows[14].Value.A.AckConsumed=0}
                17 {$rows[4].Value.Result.Observation.OwnedExit=$false}
                18 {$rows[4].Value.Result.Observation.PID=0}
                19 {$rows[16].Value.Result.Error.ErrorId='UnrelatedError'}
                20 {$rows[16].Value.Result.Accepted='true'}
            }
        }
        $actualError=$null;$rejected=$false
        try{Assert-ODProfileObservations $rows $profile (Get-ODHistoricalContext $profile)}
        catch{
            $actualError=$_
            if($_.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $_.Exception.Message -cne 'OD: observation contract mismatch' -or $_.FullyQualifiedErrorId -cne 'OD: observation contract mismatch'){throw}
            $rejected=$true
        }
        $kind=if($caseIndex -lt 3){'Baseline'}else{'Mutation'}
        $name=if($caseIndex -lt 3){'Baseline'+$vector}else{'Reject'+$vector.ToString('00')}
        $facts=[ordered]@{Index=$vector;Kind=$kind;ActualRejected=$rejected}
        LS-Details SchemaCase ([ordered]@{Index=$vector;Kind=$kind;Profile=$profile;ActualRejected=$rejected;Error=$actualError})
        LS-Result $name ($rejected -eq ($caseIndex -ge 3)) $facts
    }
}
