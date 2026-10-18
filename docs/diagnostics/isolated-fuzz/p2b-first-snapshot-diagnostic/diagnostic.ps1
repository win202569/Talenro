# Single approved diagnostic; no product definitions are replaced or edited.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$global:inboxSnapshotDiagnostic=[pscustomobject][ordered]@{
    Schema=1;Mode='p2b-first-snapshot-diagnostic';HostVersion=$PSVersionTable.PSVersion.ToString()
    Core=($PSVersionTable.PSEdition -ceq 'Core');X64=[Environment]::Is64BitProcess;Admin=$null
    Created=$null;BindingMutations=$null;SnapshotCalls=0;ActualSnapshot=$null
    BreakpointLine=$null;CatchHits=0;Caught=$null;BreakpointObservationError=$null;BreakpointRemoved=$false
    Error=$null;Captured=$false;ProductChanged=$false;FullSuiteRun=$false;RuntimeApproved=$false
}
$breakpoint=$null
try{
    $id=[Security.Principal.WindowsIdentity]::GetCurrent()
    $global:inboxSnapshotDiagnostic.Admin=([Security.Principal.WindowsPrincipal]::new($id)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if($global:inboxSnapshotDiagnostic.Admin -or $PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or $PSVersionTable.PSEdition -cne 'Core' -or -not [Environment]::Is64BitProcess){throw 'Pinned ordinary PS7 required'}
    $path=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../p2b-bounded-inbox-tick-correction/inbox.psm1'))
    if((Get-FileHash -LiteralPath $path).Hash -cne '43067AD0939CB56426568DDF435695383E6C21ABFD223B5A369BC6AE6F2451C8'){throw 'Module pin'}
    $tokens=$null;$parseErrors=$null;$ast=[Management.Automation.Language.Parser]::ParseFile($path,[ref]$tokens,[ref]$parseErrors)
    if($parseErrors.Count){throw 'Module parse error'}
    $fn=$ast.Find({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq 'Get-P2BInboxSnapshot'},$true)
    $try=$fn.Body.EndBlock.Statements|Where-Object{$_ -is [Management.Automation.Language.TryStatementAst]}
    $line=$try.CatchClauses[0].Body.Statements[0].Extent.StartLineNumber
    $global:inboxSnapshotDiagnostic.BreakpointLine=$line
    $module=Import-Module $path -PassThru -DisableNameChecking
    $binding=[pscustomobject][ordered]@{
        Schema=[int]1;BindingDigest=[Convert]::FromHexString('000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F')
        SessionId=[Convert]::FromHexString('A0A1A2A3A4A5A6A7A8A9AAABACADAEAF')
        OriginTick=[long]0;Frequency=[long]1000;DeadlineTick=[long]120000
    }
    $created=New-P2BInbox $binding
    $global:inboxSnapshotDiagnostic.Created=[pscustomobject]@{Accepted=$created.Accepted;Reason=$created.Reason;HasToken=($null -ne $created.Inbox);SourceVerified=$created.SourceVerified;LiveActionsPerformed=$created.LiveActionsPerformed;RuntimeApproved=$created.RuntimeApproved}
    if(-not $created.Accepted -or $created.Reason -cne 'created'){throw 'Creation did not match prior history'}
    $token=$created.Inbox
    $binding.BindingDigest[0]=99;$binding.SessionId[0]=99;$binding.Frequency=[long]1001
    $global:inboxSnapshotDiagnostic.BindingMutations=[pscustomobject]@{BindingDigest0=[int]$binding.BindingDigest[0];SessionId0=[int]$binding.SessionId[0];Frequency=[long]$binding.Frequency}
    $breakpoint=Set-PSBreakpoint -Script $path -Line $line -Action {
        try{
            $global:inboxSnapshotDiagnostic.CatchHits++
            if($global:inboxSnapshotDiagnostic.CatchHits -eq 1){
                $entry=$global:Error[0];$chain=[object[]]::new(0);$exception=$entry.Exception
                for($n=0;$null -ne $exception -and $n -lt 8;$n++){
                    $chain+=,[pscustomobject]@{Type=$exception.GetType().FullName;Message=$exception.Message}
                    $exception=$exception.InnerException
                }
                $global:inboxSnapshotDiagnostic.Caught=[pscustomobject]@{ExceptionChain=$chain;FullyQualifiedErrorId=$entry.FullyQualifiedErrorId;Position=$entry.InvocationInfo.PositionMessage;ScriptStackTrace=$entry.ScriptStackTrace}
            }
        }catch{$global:inboxSnapshotDiagnostic.BreakpointObservationError=$_.Exception.Message}
    }
    $global:inboxSnapshotDiagnostic.SnapshotCalls++
    $global:inboxSnapshotDiagnostic.ActualSnapshot=Get-P2BInboxSnapshot $token
    $actual=$global:inboxSnapshotDiagnostic.ActualSnapshot
    $global:inboxSnapshotDiagnostic.Captured=$null -ne $actual -and $global:inboxSnapshotDiagnostic.SnapshotCalls -eq 1 -and $null -eq $global:inboxSnapshotDiagnostic.BreakpointObservationError
    if($actual.Reason -ceq 'internal-error'){$global:inboxSnapshotDiagnostic.Captured=$global:inboxSnapshotDiagnostic.Captured -and $null -ne $global:inboxSnapshotDiagnostic.Caught -and $global:inboxSnapshotDiagnostic.CatchHits -eq 1}
} catch {
    $global:inboxSnapshotDiagnostic.Error=[pscustomobject]@{Message=$_.Exception.Message;Type=$_.Exception.GetType().FullName;Position=$_.InvocationInfo.PositionMessage;Stack=$_.ScriptStackTrace}
} finally {
    if($null -ne $breakpoint){Remove-PSBreakpoint -Breakpoint $breakpoint;$global:inboxSnapshotDiagnostic.BreakpointRemoved=$true}
}
$global:inboxSnapshotDiagnostic|ConvertTo-Json -Depth 24
if(-not $global:inboxSnapshotDiagnostic.Captured -or $null -ne $global:inboxSnapshotDiagnostic.Error -or -not $global:inboxSnapshotDiagnostic.BreakpointRemoved){exit 1}
