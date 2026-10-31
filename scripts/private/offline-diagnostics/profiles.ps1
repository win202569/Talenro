# Fixed, one-shot dispatch. No caller path, callback, manifest or resume input.
function Invoke-ODProfile {
    [CmdletBinding(PositionalBinding=$false)]
    param([Parameter(Mandatory)][string]$Profile)
    if($script:Invoked){throw 'OD: profile already invoked'}
    $script:Invoked=$true
    if($Profile -cnotin @('entry-migration-core','entry-migration-owned','entry-migration-live')){throw 'OD: unsupported profile'}
    $script:Groups=[Collections.Generic.List[object]]::new()
    $script:Details=[Collections.Generic.List[object]]::new()
    $script:ActiveSessions=[Collections.Generic.List[object]]::new()
    $script:HostPrimary=$null
    $script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
    $script:HostOverflow=$false;$script:CurrentGroup=$null
    $script:Grant=$Profile
    $Grant=$Profile
    $repo=$script:Repo;$packet=$script:Packet
    $manifestPath=Join-Path $packet 'runtime-green.frozen.json'
    $FrozenSHA256=(Get-FileHash -LiteralPath $manifestPath -Algorithm SHA256).Hash
    $script:FrozenSHA256=$FrozenSHA256
    $manifest=Get-Content -LiteralPath $manifestPath -Raw | ConvertFrom-Json
    if($manifest.Schema -ne 2 -or $manifest.Phase -cne 'runtime-green' -or $manifest.Pins.Count -ne 2129){throw 'OD: unbound runtime manifest'}
    $seen=@{}
    foreach($pin in $manifest.Pins){
        if($pin.ExpectedState -cne 'Present' -or (Get-FileHash -LiteralPath $pin.VerificationPath -Algorithm SHA256).Hash -cne $pin.SHA256){throw 'OD: frozen input changed'}
        $key=[IO.Path]::GetFullPath($pin.VerificationPath)
        if($seen.ContainsKey($key) -and $seen[$key] -cne $pin.SHA256){throw 'OD: conflicting input identity'}
        $seen[$key]=$pin.SHA256
    }
    $script:Oracle=Get-Content -LiteralPath (Join-Path $script:TestLeaf expected-output.json) -Raw | ConvertFrom-Json -AsHashtable
    $entryPath=[IO.Path]::GetFullPath((Join-Path $repo 'scripts/diagnose-offline.ps1'))
    $entryHash=(Get-FileHash -LiteralPath $entryPath -Algorithm SHA256).Hash
    $kernelHash=(Get-FileHash -LiteralPath $script:KernelPath -Algorithm SHA256).Hash
    $root=Join-Path $packet ('profiles/'+$Profile)
    if(Test-Path -LiteralPath $root){throw 'OD: profile already attempted'}
    $null=[IO.Directory]::CreateDirectory($root)
    $stream=$null
    try{
        $stream=[IO.File]::Open((Join-Path $root profile.intent.json),[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
        $bytes=[Text.UTF8Encoding]::new($false).GetBytes((ConvertTo-Json -InputObject ([ordered]@{Profile=$Profile;AttemptConsumed=$true;FrozenSHA256=$FrozenSHA256})))
        $stream.Write($bytes,0,$bytes.Length);$stream.Flush()
    }catch{Capture-Failure $_;throw}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{Capture-Failure $_;throw}}}
    $closed=$true
    try{
        . (Join-Path $script:TestLeaf test-helpers.ps1)
        switch -CaseSensitive ($Profile){
            'entry-migration-core' {
                . (Join-Path $script:TestLeaf prefix.tests.ps1)
                . (Join-Path $script:TestLeaf core.tests.ps1)
                Invoke-LSPrefixTests
                Invoke-LSCoreTests
            }
            'entry-migration-owned' {
                . (Join-Path $script:TestLeaf owned.tests.ps1)
                Invoke-LSOwnedTests
            }
            'entry-migration-live' {
                . (Join-Path $script:TestLeaf live.tests.ps1)
                Invoke-LSLiveTests
            }
        }
    }catch{Capture-Failure $_}
    finally{
        foreach($handle in $script:ActiveSessions){try{$null=Close-LSSession $handle Stop}catch{$closed=$false;Capture-Failure $_}}
        try{[Console]::Out.Flush()}catch{Capture-Failure $_}
        $script:CurrentGroup=$null
    }
    $good=$closed -and $null -eq $script:HostPrimary -and $script:HostSecondary.Count -eq 0 -and -not $script:HostOverflow -and @($script:Groups | Where-Object {-not $_.Passed}).Count -eq 0
    [pscustomobject][ordered]@{
        Schema=1;Profile=$Profile;EntryPath=$entryPath;EntrySHA256=$entryHash
        KernelPath=$script:KernelPath;KernelSHA256=$kernelHash;FrozenSHA256=$FrozenSHA256
        Status=$(if($good){'COMPLETE'}else{'INCOMPLETE'});Groups=$script:Groups.ToArray();Observations=$script:Details.ToArray()
        PrimaryError=$script:HostPrimary;SecondaryErrors=$script:HostSecondary.ToArray();Overflow=$script:HostOverflow;Closed=$closed
        SourceVerified=$false;LiveActionsPerformed=$false;RuntimeApproved=$false;FormalAcceptance=$false;RealGateImplemented=$false
    }
}
