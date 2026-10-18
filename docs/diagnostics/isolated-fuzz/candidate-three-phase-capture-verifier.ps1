# Offline text/receipt classifier only. Never dot-sources or invokes policy/tests.
# This is not a receipt authenticity, process ownership or product acceptance proof.
function Test-CandidateThreePhaseCapture {
    param([AllowEmptyString()][string]$Stdout,[AllowEmptyString()][string]$Stderr,$Receipt)
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Status='invalid-capture';CaptureValid=$false;Complete=$false;FormalProjectAcceptance=$false;Reason=$null;LineCount=0;PassCount=0;MarkerCount=0;ClosedPairs=@();OpenScopes=@()}
    try {
        $inventoryPath=Join-Path $PSScriptRoot 'integer-candidate-three-phase/expected-output.json'
        if((Get-FileHash -LiteralPath $inventoryPath -ErrorAction Stop).Hash -cne '7AC1F1B990AB0E1A63838ED7690CA857A11ED50DC0B8C696151649A4DB8A9944'){throw 'Inventory identity mismatch'}
        $inventory=Get-Content -LiteralPath $inventoryPath -Raw -ErrorAction Stop|ConvertFrom-Json
        if($null -eq $Receipt){throw 'Receipt required'}
        foreach($name in @('StopwatchFrequency','ParentAdministratorRole','ChildLaunched','StreamsSaved','TimedOut','ExitCode','ElapsedMs','CPUMs','CleanupStatus')){
            if($null -eq $Receipt.PSObject.Properties[$name]){throw ('Missing receipt field: '+$name)}
        }
        foreach($name in @('ParentAdministratorRole','ChildLaunched','StreamsSaved','TimedOut')){
            if($Receipt.$name -isnot [bool]){throw ('Boolean receipt field required: '+$name)}
        }
        if($Receipt.ParentAdministratorRole -or -not $Receipt.ChildLaunched -or -not $Receipt.StreamsSaved -or $Receipt.CleanupStatus -cne 'owned-host-exited'){throw 'Capture context/cleanup not verified'}
        $frequency=$Receipt.StopwatchFrequency
        if(($frequency -isnot [int] -and $frequency -isnot [long]) -or $frequency -le 0){throw 'Positive integer frequency required'}
        foreach($name in @('ElapsedMs','CPUMs')){
            $number=$Receipt.$name
            if(($number -isnot [int] -and $number -isnot [long] -and $number -isnot [double] -and $number -isnot [decimal]) -or [double]::IsNaN([double]$number) -or [double]::IsInfinity([double]$number) -or $number -lt 0){throw ('Finite nonnegative number required: '+$name)}
        }
        if($Receipt.ElapsedMs -le 0){throw 'Positive elapsed time required'}
        if($Receipt.ExitCode -isnot [int] -and $Receipt.ExitCode -isnot [long]){throw 'Integer exit code required'}
        if($Stderr.Length -ne 0){throw 'Nonempty stderr'}
        if($Stdout.Length -eq 0 -or -not $Stdout.EndsWith("`n",[StringComparison]::Ordinal)){throw 'Unterminated/empty stdout'}
        $lines=@($Stdout -split '\r?\n'|Where-Object {$_ -cne ''})
        $result.LineCount=$lines.Count
        if($lines.Count -lt 2 -or $lines.Count -gt $inventory.ExpectedNonemptyLines){throw 'Output length outside inventory'}
        $pairs=@();$stack=@();$previousTick=[long]-1;$markerCount=0;$passes=0
        for($i=0;$i -lt $lines.Count;$i++){
            $rule=$inventory.Lines[$i];$line=$lines[$i]
            switch -CaseSensitive ($rule.Kind){
                Context {if($line -cnotmatch $inventory.ContextRule){throw ('Context mismatch at line '+($i+1))}}
                Host {if($line -cnotmatch $inventory.HostRule){throw ('Host mismatch at line '+($i+1))}}
                Literal {if($line -cne $rule.Value){throw ('Literal mismatch at line '+($i+1))};if($line.StartsWith('PASS: ',[StringComparison]::Ordinal)){$passes++}}
                {$_ -ceq 'Boundary' -or $_ -ceq 'Subphase' -or $_ -ceq 'ProductPhase'} {
                    if($line -cnotmatch $rule.Pattern){throw ('Marker mismatch at line '+($i+1))}
                    $tick=[long]0
                    if(-not [long]::TryParse($Matches[1],[Globalization.NumberStyles]::None,[Globalization.CultureInfo]::InvariantCulture,[ref]$tick) -or $tick -lt 0 -or $tick -lt $previousTick){throw 'Invalid/nonmonotonic tick'}
                    $previousTick=$tick;$markerCount++
                    if($rule.Edge -ceq 'START'){
                        switch -CaseSensitive ($rule.Kind){
                            Boundary {
                                if($rule.Parent -cne '' -or $stack.Count -ne 0){throw 'Unexpected outer nesting'}
                            }
                            Subphase {
                                if($rule.Parent -cne 'Check4096' -or $stack.Count -ne 1 -or $stack[-1].Kind -cne 'Boundary' -or $stack[-1].Name -cne $rule.Parent){throw 'Missing expected boundary parent'}
                            }
                            ProductPhase {
                                if($rule.Parent -cne 'Evaluate4096' -or $stack.Count -ne 2 -or $stack[-1].Kind -cne 'Subphase' -or $stack[-1].Name -cne $rule.Parent){throw 'Missing expected API parent'}
                            }
                            default {throw 'Unknown timing kind'}
                        }
                        $stack+=,[pscustomobject]@{Kind=$rule.Kind;Name=$rule.Name;Parent=$rule.Parent;StartTicks=$tick}
                    }elseif($rule.Edge -ceq 'END'){
                        if($stack.Count -eq 0 -or $stack[-1].Kind -cne $rule.Kind -or $stack[-1].Name -cne $rule.Name){throw 'Unmatched stack end'}
                        $open=$stack[-1]
                        $pairs+=,[pscustomobject]@{Kind=$rule.Kind;Name=$rule.Name;Parent=$rule.Parent;StartTicks=$open.StartTicks;EndTicks=$tick;ElapsedMs=[Math]::Round(($tick-$open.StartTicks)*1000.0/$frequency,3)}
                        if($stack.Count -eq 1){$stack=@()}else{$stack=@($stack[0..($stack.Count-2)])}
                    }else{throw 'Unknown marker edge'}
                }
                default {throw 'Unknown inventory kind'}
            }
        }
        $result.CaptureValid=$true;$result.Status='valid-prefix'
        $result.PassCount=$passes;$result.MarkerCount=$markerCount;$result.ClosedPairs=$pairs;$result.OpenScopes=$stack
        $result.Complete=$lines.Count -eq $inventory.ExpectedNonemptyLines -and $passes -eq $inventory.ExpectedPassCount -and $markerCount -eq 18 -and $pairs.Count -eq 9 -and $stack.Count -eq 0 -and -not $Receipt.TimedOut -and $Receipt.ExitCode -eq 0 -and $Receipt.ElapsedMs -le 30000
        if($result.Complete){$result.Status='complete-diagnostic'}
        else{$result.Reason='Output prefix or execution gate incomplete'}
    }catch{
        $result.Reason=$_.Exception.Message
    }
    return $result
}
