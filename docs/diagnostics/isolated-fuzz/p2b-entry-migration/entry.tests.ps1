# Actual public calls: moving execution into Inspect or accepting an old argument
# must fail through observable imports/tree changes, not a source-text assertion.
function OD-Equal($A,$B){(ConvertTo-Json -InputObject $A -Depth 50 -Compress) -ceq (ConvertTo-Json -InputObject $B -Depth 50 -Compress)}
function OD-Expand($Value,$Manifest){
    if($null -eq $Value){return $null}
    if($Value -is [Collections.IDictionary]){
        $result=[ordered]@{};foreach($key in $Value.Keys){$result[$key]=OD-Expand $Value[$key] $Manifest};return $result
    }
    if($Value -is [array]){$rows=[Collections.Generic.List[object]]::new();foreach($item in $Value){$rows.Add((OD-Expand $item $Manifest))};return ,($rows.ToArray())}
    if($Value -is [string]){
        if($Value.StartsWith('@repo:')){return [IO.Path]::GetFullPath((Join-Path $repo $Value.Substring(6)))}
        if($Value -ceq '@frozen'){return $FrozenSHA256}
        if($Value.StartsWith('@pin:')){
            $path=[IO.Path]::GetFullPath((Join-Path $repo $Value.Substring(5)))
            $pins=@($Manifest.Pins | Where-Object {$_.VerificationPath -ieq $path})
            LS-Need ($pins.Count -eq 1 -and $pins[0].ExpectedState -ceq 'Present') 'unique expected source pin'
            return $pins[0].SHA256
        }
    }
    return $Value
}
function Get-ODProfileTree {
    $path=Join-Path $packet profiles
    $rows=[Collections.Generic.List[object]]::new()
    if(Test-Path -LiteralPath $path){
        foreach($item in (Get-ChildItem -LiteralPath $path -Recurse -Force | Sort-Object FullName)){
            $relative=[IO.Path]::GetRelativePath($path,$item.FullName).Replace('\','/')
            if($item.PSIsContainer){$rows.Add([ordered]@{Path=$relative;Kind='Directory';Length=$null;SHA256=$null})}
            else{$rows.Add([ordered]@{Path=$relative;Kind='File';Length=$item.Length;SHA256=(Get-FileHash -LiteralPath $item.FullName -Algorithm SHA256).Hash})}
        }
    }
    return ,($rows.ToArray())
}
function Get-ODProfileRoots {
    $path=Join-Path $packet profiles
    if(-not (Test-Path -LiteralPath $path)){return ,([string[]]@())}
    return ,([string[]]@(Get-ChildItem -LiteralPath $path -Directory | Sort-Object Name | ForEach-Object {$_.Name}))
}
function OD-RuntimeLoaded([string]$Kernel){return (@(Get-Module -All | Where-Object {$_.Path -ieq $Kernel}).Count -gt 0)}
function OD-CheckInspect($Report,[string]$CurrentGrant){
    LS-Need ($null -ne $Report -and (OD-Equal @($Report.PSObject.Properties.Name) $oracle.InspectRootNames)) 'inspect complete ordered root'
    $available=$CurrentGrant -cin @('lifecycle','continuation')
    $phase=if($available){'runtime-green'}else{'entry-green'}
    $manifestPath=Join-Path $packet ($phase+'.frozen.json')
    LS-Need ((Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash -ceq $FrozenSHA256) 'inspect expected manifest binding'
    $manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json -AsHashtable
    $expected=OD-Expand $oracle.InspectTemplates[$(if($available){'Present'}else{'Missing'})] $manifest
    LS-Need (OD-Equal $Report $expected) 'inspect full literal data contract'
}
function Invoke-ODEntryContract([string]$CurrentGrant){
    $entry=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/diagnose-offline.ps1'))
    $kernel=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/private/offline-diagnostics/session.psm1'))
    $exists=[IO.File]::Exists($entry)
    LS-Result EntryPresent $exists ([ordered]@{Exists=$exists})
    if(-not $exists){return}
    $before=Get-ODProfileTree;$rootsBefore=Get-ODProfileRoots
    LS-Need (OD-Equal $rootsBefore $oracle.Grants[$CurrentGrant].ExistingProfiles) 'declared pre-existing profile roots'
    LS-Need (-not (OD-RuntimeLoaded $kernel)) 'fresh capture runtime unloaded'
    $r=& $entry
    $afterDefault=Get-ODProfileTree
    LS-Details InspectDefault ([ordered]@{Report=$r;Before=$before;After=$afterDefault;Roots=$rootsBefore;RuntimeLoaded=(OD-RuntimeLoaded $kernel)})
    OD-CheckInspect $r $CurrentGrant
    $same=OD-Equal $before $afterDefault;$loaded=OD-RuntimeLoaded $kernel
    LS-Need ($same -and -not $loaded) 'default inspect no preparation'
    LS-Result InspectDefaultReadOnly ($same -and -not $loaded -and $r.Authority -ceq 'None') ([ordered]@{Authority=$r.Authority;RuntimeLoaded=$loaded;ProfileTreeUnchanged=$same})
    $explicit=& $entry -Action Inspect
    $afterExplicit=Get-ODProfileTree
    LS-Details InspectExplicit ([ordered]@{Report=$explicit;Before=$before;After=$afterExplicit;Roots=(Get-ODProfileRoots);RuntimeLoaded=(OD-RuntimeLoaded $kernel)})
    OD-CheckInspect $explicit $CurrentGrant
    $same=OD-Equal $before $afterExplicit;$loaded=OD-RuntimeLoaded $kernel
    LS-Need ($same -and -not $loaded) 'explicit inspect no preparation'
    LS-Result InspectExplicitReadOnly ($same -and -not $loaded) ([ordered]@{Authority=$explicit.Authority;RuntimeLoaded=$loaded;ProfileTreeUnchanged=$same})
    $rejected=0
    foreach($request in $oracle.Rejections){
        $actualError=$null
        try{
            if($request.Positional){$unexpected=& $entry unexpected}
            else{$requestParameters=[hashtable]$request.Parameters;$unexpected=& $entry @requestParameters}
        }catch{$actualError=$_}
        LS-Details RejectedRequest ([ordered]@{Name=$request.Name;Error=$actualError;RuntimeLoaded=(OD-RuntimeLoaded $kernel)})
        LS-Need ($actualError -is [Management.Automation.ErrorRecord]) ('public request must reject '+$request.Name)
        LS-Need ($actualError.Exception.GetType().FullName -ceq $request.Type -and $actualError.FullyQualifiedErrorId -cin $request.Ids -and $actualError.Exception.Message -ceq $request.Message) ('fixed actual rejection '+$request.Name)
        $rejected++
    }
    $afterRejected=Get-ODProfileTree;$same=OD-Equal $before $afterRejected;$loaded=OD-RuntimeLoaded $kernel
    LS-Details EntryTree ([ordered]@{Before=$before;After=$afterRejected;Roots=(Get-ODProfileRoots)})
    LS-Need ($same -and -not $loaded -and $rejected -eq 14) 'request guards no preparation'
    LS-Result UnsupportedRequestNoRuntime ($rejected -eq 14 -and $same -and -not $loaded) ([ordered]@{Rejected=$rejected;RuntimeLoaded=$loaded;ProfileTreeUnchanged=$same})
    $missing=@($explicit.Dependencies | Where-Object {$_.Role.StartsWith('Impl:') -and -not $_.Exists}).Count
    $freeze=@($explicit.Dependencies | Where-Object {$_.Role -ceq 'FreezeManifest'})
    LS-Need ($freeze.Count -eq 1) 'single freeze observation'
    LS-Result InspectDependencyStateReadOnly ($missing -eq $(if($CurrentGrant -ceq 'entry-green'){6}else{0})) ([ordered]@{MissingImplementation=$missing;FreezeManifestPresent=$freeze[0].Exists})
    LS-Result InspectHistoryDataOnly ($explicit.Authority -ceq 'None' -and $explicit.History.State -ceq 'Consumed' -and $explicit.History.Role -ceq 'evidence-only') ([ordered]@{State=$explicit.History.State;Role=$explicit.History.Role;Authority=$explicit.Authority})
}
function Invoke-ODRuntimeContract([string]$CurrentGrant){
    $entry=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/diagnose-offline.ps1'))
    $kernel=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/private/offline-diagnostics/session.psm1'))
    $profile=$oracle.Grants[$CurrentGrant].Profile
    $report=$null;$entryError=$null
    try{$report=& $entry -Action RunApproved -Profile $profile}catch{$entryError=$_}
    if($null -ne $entryError){
        LS-Details EntryInvocation ([ordered]@{EntryPath=$entry;Profile=$profile;Error=$entryError;RuntimeLoaded=(OD-RuntimeLoaded $kernel)})
        if($CurrentGrant -cne 'runtime-red' -or $entryError.Exception.GetType().FullName -cne 'System.Management.Automation.RuntimeException' -or $entryError.Exception.Message -cne 'OD: backend unavailable' -or $entryError.FullyQualifiedErrorId -cne 'OD: backend unavailable'){Capture-Failure $entryError;throw $entryError}
        LS-Result RuntimeProfileBound $false ([ordered]@{Profile=$profile;PublicEntry=$true;PrivateModule=$false;Closed=$false;SourceBound=$false})
        return
    }
    LS-Details ProfileReturn $report
    LS-Need ($null -ne $report -and (OD-Equal @($report.PSObject.Properties.Name) $oracle.ProfileRootNames)) 'full profile ordered root'
    foreach($group in $report.Groups){$script:Groups.Add($group)}
    if($null -ne $report.PrimaryError){Capture-Failure $report.PrimaryError}
    LS-Need ($report.Status -ceq 'COMPLETE' -and $report.Closed -and $null -eq $report.PrimaryError -and $report.SecondaryErrors.Count -eq 0 -and -not $report.Overflow) 'complete profile without controller fault'
    $modules=@(Get-Module -All | Where-Object {$_.Path -ieq $kernel})
    $source=$modules.Count -eq 1 -and $report.EntryPath -ceq $entry -and $report.KernelPath -ceq $kernel -and $report.EntrySHA256 -ceq (Get-FileHash -LiteralPath $entry -Algorithm SHA256).Hash -and $report.KernelSHA256 -ceq (Get-FileHash -LiteralPath $kernel -Algorithm SHA256).Hash -and $report.FrozenSHA256 -ceq $FrozenSHA256
    LS-Result RuntimeProfileBound $source ([ordered]@{Profile=$report.Profile;PublicEntry=($report.EntryPath -ceq $entry);PrivateModule=($modules.Count -eq 1);Closed=$report.Closed;SourceBound=$source})
    LS-Need $source 'public entry bound to actual new module'
    $moduleBefore=$modules[0];$before=Get-ODProfileTree;$againError=$null
    try{$unexpected=& $entry -Action RunApproved -Profile $profile}catch{$againError=$_}
    $after=Get-ODProfileTree;$moduleAfter=@(Get-Module -All | Where-Object {$_.Path -ieq $kernel})
    $sameModule=$moduleAfter.Count -eq 1 -and [object]::ReferenceEquals($moduleBefore,$moduleAfter[0])
    $sameTree=OD-Equal $before $after
    LS-Details ProfileReentry ([ordered]@{Error=$againError;SameModule=$sameModule;Before=$before;After=$after})
    $rejected=$null -ne $againError -and $againError.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $againError.FullyQualifiedErrorId -ceq 'OD: profile already invoked' -and $againError.Exception.Message -ceq 'OD: profile already invoked'
    LS-Need ($rejected -and $sameModule -and $sameTree) 'same host profile invocation cannot reset runtime'
    LS-Result ProfileReentryRejected ($rejected -and $sameModule -and $sameTree) ([ordered]@{Rejected=$rejected;SameModule=$sameModule;ProfileTreeUnchanged=$sameTree})
}
