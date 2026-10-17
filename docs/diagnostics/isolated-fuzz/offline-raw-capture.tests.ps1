param([ValidateSet('all','ownership-mutation','stop-mutation')][string]$Group='all')
$ErrorActionPreference='Stop'
# Tests name the breaks: wrong byte count/EOF, bad stop mapping, coercion, and shared mutable inputs.
function Assert-Raw([bool]$Value,[string]$Label){if(-not $Value){throw ('ASSERTION FAILED: '+$Label)}}
function Need([bool]$Value){Assert-Raw $Value $script:rawCase}
function Raw-Case([string]$Name,[scriptblock]$Body){
    if($Group -eq 'ownership-mutation' -and $Name -cnotin @('ApiRequired','CopyOwnership','Complete')){return}
    if($Group -eq 'stop-mutation' -and $Name -cnotin @('ApiRequired','DeadlineEdges','Complete')){return}
    $script:rawCase='RawCapture'+$Name;& $Body;Write-Output ('PASS: '+$script:rawCase)
}
function New-RawReadAction($Channel='stdout',$Count=4,$ElapsedMs=0){[pscustomobject]@{Kind='read';ElapsedMs=$ElapsedMs;Channel=$Channel;Count=$Count}}
function Raw-Exit($Code=0,$Time=0){[pscustomobject]@{Kind='exit';ElapsedMs=$Time;ExitCode=$Code}}
function Raw-Tick($Time=0){[pscustomobject]@{Kind='tick';ElapsedMs=$Time}}
function New-RawTestScenario([byte[]]$Out=[byte[]]@(),[byte[]]$Err=[byte[]]@(),[object[]]$Actions=[object[]]@()){
    [pscustomobject]@{Schema=1;SourceMode='synthetic';Header=[pscustomobject]@{SessionLabel='s';ExecutableDeclaration='x';ArgumentBytes=[byte[]]@();EncodingName='utf-8';OriginDeclaration='o';StartedUTC='2026-10-17T00:00:00.0000000Z';EndedUTC='2026-10-17T00:00:00.0000000Z'};StdoutBytes=$Out;StderrBytes=$Err;Actions=$Actions}
}
function Raw-Shape($r){
    $names=@('Schema','SourceMode','Created','Reason','Record','Actions','FailureActionIndex','OriginVerified','Acceptance','RuntimeApproved')
    Need (@($r.PSObject.Properties).Count -eq 10)
    foreach($p in $r.PSObject.Properties){Need ($p.Name -cin $names)}
    Need ($r.Schema -eq 1 -and $r.SourceMode -ceq 'synthetic' -and $r.Created -is [bool])
    foreach($name in @('OriginVerified','Acceptance','RuntimeApproved')){Need ($r.$name -is [bool] -and -not $r.$name)}
}
function Assert-RawRecord($Result){
    Raw-Shape $Result;Need ($Result.Created -and $Result.Reason -ceq 'record-created' -and $null -ne $Result.Record)
    Need (@($Result.Record.PSObject.Properties).Count -eq 20)
    Need ((Test-SyntheticCaptureRecord $Result.Record).Verified)
    Need (-not $Result.Record.LiveActionsPerformed -and -not $Result.Record.OriginVerified -and -not $Result.Record.Acceptance -and -not $Result.Record.RuntimeApproved -and $Result.Record.CleanupStatus -ceq 'unknown')
}
function Raw-Good($Scenario){$r=New-OfflineRawCaptureRecord $Scenario;Assert-RawRecord $r;return $r}
function Raw-Bad($Scenario,[string]$Reason){$r=New-OfflineRawCaptureRecord $Scenario;Raw-Shape $r;Need (-not $r.Created -and $null -eq $r.Record -and $null -eq $r.Actions -and $r.Reason -ceq $Reason)}
function Raw-Hash([byte[]]$Bytes){$s=[Security.Cryptography.SHA256]::Create();try{return [BitConverter]::ToString($s.ComputeHash($Bytes)).Replace('-','')}finally{$s.Dispose()}}
function Raw-Hex([string]$Text){$b=New-Object byte[] ($Text.Length/2);for($i=0;$i -lt $b.Length;$i++){$b[$i]=[Convert]::ToByte($Text.Substring(2*$i,2),16)};return ,$b}
function Raw-Shadow($Value,[string]$Name,$Data){$Value=Add-Member -InputObject $Value -MemberType NoteProperty -Name $Name -Value $Data -Force -PassThru;Need ($Value.PSObject.Properties[$Name].MemberType -eq [Management.Automation.PSMemberTypes]::NoteProperty);return ,$Value}
try{
    $identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=New-Object Security.Principal.WindowsPrincipal($identity)
    Write-Output ('CONTEXT: '+$identity.Name+' admin='+$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
    $helper=Join-Path $PSScriptRoot 'offline-raw-capture.ps1';if(Test-Path -LiteralPath $helper){. $helper}
    Raw-Case ApiRequired {Need ($null -ne (Get-Command New-OfflineRawCaptureRecord -CommandType Function -ErrorAction SilentlyContinue))}
    Write-Output ('HOST: '+$PSVersionTable.PSVersion.ToString()+' '+$PSVersionTable.PSEdition+' x64='+[Environment]::Is64BitProcess)
    Raw-Case NormalBytes {
        $p=New-RawTestScenario ([byte[]]@(0,1,2)) ([byte[]]@(255)) @((New-RawReadAction stdout),(New-RawReadAction stderr),(New-RawReadAction stdout),(New-RawReadAction stderr),(Raw-Exit))
        $r=Raw-Good $p;Need ($r.Record.Lifecycle.Status -ceq 'complete' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'AAEC' -and $r.Record.Lifecycle.Stderr.Base64 -ceq '/w==')
        Need ($r.Record.Lifecycle.Stdout.SHA256 -ceq (Raw-Hash ([byte[]]@(0,1,2))) -and $r.Record.Lifecycle.Stderr.SHA256 -ceq (Raw-Hash ([byte[]]@(255))))
    }
    Raw-Case Utf8Split {
        $r=Raw-Good (New-RawTestScenario ([byte[]]@(228,184,173)) ([byte[]]@()) @((New-RawReadAction stdout 1),(New-RawReadAction stdout 2)))
        Need ($r.Record.Lifecycle.Stdout.Base64 -ceq '5Lit' -and $r.Record.Events[0].Data.Base64 -ceq '5A==' -and $r.Record.Events[1].Data.Base64 -ceq 'uK0=')
    }
    Raw-Case Utf16Split {
        $p=New-RawTestScenario ([byte[]]@(45,78)) ([byte[]]@()) @((New-RawReadAction stdout 1),(New-RawReadAction stdout 1));$p.Header.EncodingName='utf-16le'
        $r=Raw-Good $p;Need ($r.Record.Lifecycle.Stdout.Base64 -ceq 'LU4=' -and $r.Record.Events[0].Data.Base64 -ceq 'LQ==' -and $r.Record.Events[1].Data.Base64 -ceq 'Tg==')
    }
    Raw-Case ShortRead {
        $r=Raw-Good (New-RawTestScenario ([byte[]]@(1,2)) ([byte[]]@()) @((New-RawReadAction stdout 4)))
        Need ($r.Record.Lifecycle.Stdout.Base64 -ceq 'AQI=' -and -not $r.Record.Lifecycle.StdoutEOF -and $r.Record.Events[0].Data.ByteLength -eq 2)
    }
    Raw-Case EmptyEOF {
        $r=Raw-Good (New-RawTestScenario -Actions @((New-RawReadAction stdout),(New-RawReadAction stderr),(Raw-Exit)))
        Need ($r.Record.Lifecycle.Status -ceq 'complete' -and $r.Record.Events[0].Kind -ceq 'eof' -and $r.Record.Events[1].Kind -ceq 'eof' -and $r.Record.Lifecycle.Stdout.ByteLength -eq 0)
    }
    Raw-Case EarlyExit {
        $r=Raw-Good (New-RawTestScenario ([byte[]]@(1)) ([byte[]]@()) @((Raw-Exit),(New-RawReadAction stdout),(New-RawReadAction stdout),(New-RawReadAction stderr)))
        Need ($r.Record.Lifecycle.Status -ceq 'complete' -and $r.Record.Events[0].Kind -ceq 'exit' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'AQ==')
    }
    Raw-Case UnknownExit {$r=Raw-Good (New-RawTestScenario -Actions @((New-RawReadAction stdout),(New-RawReadAction stderr)));Need ($r.Record.Lifecycle.Status -ceq 'incomplete' -and $null -eq $r.Record.Lifecycle.ExitCode -and -not $r.Record.Lifecycle.ExitObserved)}
    Raw-Case UnreadTail {$r=Raw-Good (New-RawTestScenario ([byte[]]@(1,2)) ([byte[]]@()) @((New-RawReadAction stdout 1)));Need ($r.Record.Lifecycle.Status -ceq 'incomplete' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'AQ==' -and $r.Record.Events.Count -eq 1)}
    Raw-Case DeadlineEdges {
        foreach($t in @(119999,120000,120001)){
            $r=Raw-Good (New-RawTestScenario ([byte[]]@(9)) ([byte[]]@()) @((New-RawReadAction stdout 1 $t)))
            if($t -eq 119999){Need ($r.Record.Events[0].Kind -ceq 'chunk' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'CQ==' -and $r.Record.Lifecycle.Status -ceq 'incomplete')}
            else{Need ($r.Record.Events[0].Kind -ceq 'tick' -and $null -eq $r.Record.Events[0].PSObject.Properties['Data'] -and $r.Record.Lifecycle.Status -ceq 'timed-out' -and $r.Record.Lifecycle.Stdout.ByteLength -eq 0 -and $r.FailureActionIndex -eq 0)}
        }
    }
    Raw-Case EqualTime {$r=Raw-Good (New-RawTestScenario -Actions @((Raw-Exit),(New-RawReadAction stderr),(New-RawReadAction stdout)));Need ($r.Record.Events[0].Kind -ceq 'exit' -and $r.Record.Events[1].Channel -ceq 'stderr' -and $r.Record.Events[2].Channel -ceq 'stdout' -and $r.Record.Lifecycle.Status -ceq 'complete')}
    Raw-Case BackwardTime {
        $r=Raw-Good (New-RawTestScenario ([byte[]]@(1,2)) ([byte[]]@()) @((New-RawReadAction stdout 1 2),(New-RawReadAction stdout 1 1)))
        Need ($r.Record.Lifecycle.Reason -ceq 'time-went-backward' -and $r.Record.Events[1].Kind -ceq 'tick' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'AQ==' -and $r.FailureActionIndex -eq 1)
    }
    Raw-Case AfterEOF {
        $r=Raw-Good (New-RawTestScenario -Actions @((New-RawReadAction stdout),(New-RawReadAction stdout),(Raw-Exit)))
        Need ($r.Record.Lifecycle.Reason -ceq 'chunk-after-eof' -and $r.Record.Events[1].Kind -ceq 'chunk' -and $r.Record.Events[1].Data.ByteLength -eq 0 -and $r.FailureActionIndex -eq 1 -and $r.Actions.Count -eq 3 -and $r.Record.Events.Count -eq 2)
    }
    Raw-Case DuplicateExit {$r=Raw-Good (New-RawTestScenario -Actions @((Raw-Exit),(Raw-Exit)));Need ($r.Record.Lifecycle.Reason -ceq 'duplicate-exit' -and $r.Record.Lifecycle.AcceptedEventCount -eq 1 -and $r.FailureActionIndex -eq 1)}
    Raw-Case Trailing {
        $r=Raw-Good (New-RawTestScenario -Actions @((New-RawReadAction stdout),(New-RawReadAction stderr),(Raw-Exit),(New-RawReadAction stdout)))
        Need ($r.Record.Lifecycle.Reason -ceq 'trailing-after-complete' -and $r.Record.Events[3].Kind -ceq 'tick' -and $r.FailureActionIndex -eq 3 -and $null -eq $r.Record.Lifecycle.Package)
    }
    Raw-Case InvalidScenario {
        Raw-Bad $null invalid-scenario
        foreach($v in @('live','Synthetic')){$p=New-RawTestScenario;$p.SourceMode=$v;Raw-Bad $p invalid-scenario}
        $p=New-RawTestScenario;$p.Schema='1';Raw-Bad $p invalid-scenario
        $p=New-RawTestScenario;$p|Add-Member NoteProperty Extra 1;Raw-Bad $p invalid-scenario
    }
    Raw-Case InvalidHeader {
        foreach($v in @('utf-7','UTF-8')){$p=New-RawTestScenario;$p.Header.EncodingName=$v;Raw-Bad $p invalid-header}
        $p=New-RawTestScenario;$p.Header.ArgumentBytes=[byte[]]@(1);Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.ArgumentBytes=[byte[]]@(255,254);Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.ArgumentBytes=[byte[]]@(0,216);Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.EndedUTC='2026-10-16T00:00:00.0000000Z';Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.StartedUTC='2026-10-17';Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.ExecutableDeclaration='x'*1025;Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.ArgumentBytes=New-Object byte[] 65538;Raw-Bad $p invalid-header
        $p=New-RawTestScenario;$p.Header.PSObject.Properties.Remove('SessionLabel');Raw-Bad $p invalid-header
    }
    Raw-Case InvalidActions {
        $p=New-RawTestScenario -Actions @((Raw-Tick 120000),[pscustomobject]@{Kind='bad';ElapsedMs=0});Raw-Bad $p invalid-action
        $p=New-RawTestScenario -Actions @((New-RawReadAction STDOUT));Raw-Bad $p invalid-action
        $p=New-RawTestScenario -Actions @((Raw-Tick));$p.Actions[0].ElapsedMs=-1;Raw-Bad $p invalid-action
        $p=New-RawTestScenario -Actions @((Raw-Exit));$p.Actions[0].ExitCode=[long]2147483648;Raw-Bad $p invalid-action
        $p=New-RawTestScenario -Actions @((Raw-Tick));$p.Actions[0].Kind='tıck';Raw-Bad $p invalid-action
        $p=New-RawTestScenario -Actions @((Raw-Tick));$p.Actions[0]|Add-Member NoteProperty Extra 1;Raw-Bad $p invalid-action
    }
    Raw-Case PassiveProperties {
        $script:rawGetter=0;$p=New-RawTestScenario;$p.Header.PSObject.Properties.Remove('SessionLabel');$p.Header|Add-Member ScriptProperty SessionLabel {$script:rawGetter++;'s'}
        Raw-Bad $p invalid-header;Need ($script:rawGetter -eq 0)
        $p=New-RawTestScenario;$p|Add-Member ScriptMethod GetType {$script:rawGetter++;throw 'must-not-dispatch'} -Force
        $null=Raw-Good $p;Need ($script:rawGetter -eq 0)
    }
    Raw-Case ArrayAndByteBounds {
        $p=New-RawTestScenario;$p.StdoutBytes=[object[]]@(1);Raw-Bad $p invalid-scenario
        $p=New-RawTestScenario;$p.Actions=New-Object 'object[,]' 1,1;Raw-Bad $p invalid-action
        $p=New-RawTestScenario;$p.StdoutBytes=Raw-Shadow ([byte[]]@(1)) Length 0;Raw-Bad $p invalid-scenario
        $p=New-RawTestScenario;$p.Actions=Raw-Shadow ([object[]]@()) Rank 1;Raw-Bad $p invalid-action
        $p=New-RawTestScenario (New-Object byte[] 1048576) (New-Object byte[] 1048576);$null=Raw-Good $p
        $p=New-RawTestScenario;$p.StdoutBytes=New-Object byte[] 1048577;Raw-Bad $p invalid-scenario
        $p=New-RawTestScenario;$p.StderrBytes=New-Object byte[] 1048577;Raw-Bad $p invalid-scenario
    }
    Raw-Case ActionCountBounds {
        $many=New-Object object[] 4096;for($i=0;$i -lt $many.Length;$i++){$many[$i]=Raw-Tick}
        $r=Raw-Good (New-RawTestScenario -Actions $many);Need ($r.Actions.Count -eq 4096 -and $r.Record.Lifecycle.AcceptedEventCount -eq 4096)
        $p=New-RawTestScenario;$p.Actions=[object[]]($many+@((Raw-Tick)));Raw-Bad $p invalid-action
    }
    Raw-Case CountBounds {
        foreach($v in @(1,65536)){$r=Raw-Good (New-RawTestScenario ([byte[]]@(1)) ([byte[]]@()) @((New-RawReadAction stdout $v)));Need ($r.Record.Lifecycle.Stdout.Base64 -ceq 'AQ==')}
        foreach($v in @(0,65537,$true,'1',1.0,[long]::MaxValue)){$p=New-RawTestScenario -Actions @((New-RawReadAction stdout $v));Raw-Bad $p invalid-action}
    }
    Raw-Case CopyOwnership {
        $p=New-RawTestScenario ([byte[]]@(1,2)) ([byte[]]@()) @((New-RawReadAction stdout 1));$p.Header.ArgumentBytes=[byte[]]@(65,0)
        $r=Raw-Good $p;$before=$r.Record.BindingSHA256
        $p.Actions[0].Count=2;$p.Header.ExecutableDeclaration='changed';$p.Header.ArgumentBytes[0]=66;$p.StdoutBytes[0]=9
        Need ($r.Actions[0].Count -eq 1 -and $r.Record.ExecutableDeclaration -ceq 'x' -and $r.Record.Arguments.Base64 -ceq 'QQA=' -and $r.Record.Lifecycle.Stdout.Base64 -ceq 'AQ==')
        $r.Actions[0].Count=3;Need ($r.Record.BindingSHA256 -ceq $before -and (Test-SyntheticCaptureRecord $r.Record).Verified)
        Need (-not [object]::ReferenceEquals($r.Actions,$p.Actions) -and -not [object]::ReferenceEquals($r.Actions[0],$p.Actions[0]))
    }
    Raw-Case ResultShapeFlags {$r=Raw-Good (New-RawTestScenario);Need ($null -eq $r.FailureActionIndex -and $r.Record.StopIntent -ceq 'none' -and $r.Record.CleanupStatus -ceq 'unknown');Raw-Bad $null invalid-scenario}
    Raw-Case DeclaredBinding {
        # Hand-encoded TLRCAP01 header for s/x/o, empty arguments and zero events; not a production encoder.
        $literal=Raw-Hex '544C524341503031010000000100000001000000730100000078010000006F1C000000323032362D31302D31375430303A30303A30302E303030303030305A1C000000323032362D31302D31375430303A30303A30302E303030303030305A0000000000000000'
        $r=Raw-Good (New-RawTestScenario);Need ($r.Record.BindingSHA256 -ceq (Raw-Hash $literal));$base=$r.Record.BindingSHA256
        foreach($field in @('ExecutableDeclaration','SessionLabel','EndedUTC','ArgumentBytes')){
            $p=New-RawTestScenario
            switch($field){ExecutableDeclaration{$p.Header.ExecutableDeclaration='y'} SessionLabel{$p.Header.SessionLabel='t'} EndedUTC{$p.Header.EndedUTC='2026-10-17T00:00:01.0000000Z'} ArgumentBytes{$p.Header.ArgumentBytes=[byte[]]@(65,0)}}
            $r=Raw-Good $p;Need ($r.Record.BindingSHA256 -cne $base)
        }
    }
    Raw-Case RefusedClassification {$r=Raw-Good (New-RawTestScenario -Actions @((New-RawReadAction stdout),(New-RawReadAction stderr),(Raw-Exit (-1))));Need ($r.Record.Lifecycle.Status -ceq 'complete' -and $r.Record.Lifecycle.Package.Classification -ceq 'refused' -and -not $r.Acceptance)}
    Raw-Case DerivedConsistency {
        foreach($kind in @('complete','incomplete','timed-out','refused')){
            $a=switch($kind){complete{@((New-RawReadAction stdout),(New-RawReadAction stderr),(Raw-Exit))} incomplete{@((Raw-Tick))} timed-out{@((Raw-Tick 120000))} refused{@((Raw-Exit),(Raw-Exit))}}
            $r=Raw-Good (New-RawTestScenario -Actions ([object[]]$a));Need ($r.Record.Lifecycle.Status -ceq $kind)
            if($kind -eq 'timed-out'){Need ($r.Record.StopIntent -ceq 'would-request-owned-stop' -and $r.FailureActionIndex -eq 0)}
            if($kind -eq 'refused'){Need ($r.Record.StopIntent -ceq 'would-refuse-and-stop-if-owned' -and $r.FailureActionIndex -eq 1)}
        }
    }
    Raw-Case Complete {Need $true}
    Write-Output 'LIMIT: synthetic memory byte transfer only;no real pipes,watchdog,origin or cleanup proof.'
}catch{
    if($_.Exception.Message -cmatch '^ASSERTION FAILED: RawCapture[A-Za-z0-9]+$'){Write-Output $_.Exception.Message;exit 1}
    Write-Error $_;exit 1
}
