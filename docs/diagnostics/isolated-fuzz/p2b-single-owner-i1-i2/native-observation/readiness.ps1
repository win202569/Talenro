# Definition-only assessment of supplied evidence. Does not authenticate claims or grant permission.
function Get-C12ObservationReadiness([hashtable]$Evidence) {
    $required=@('Profile','Ownership','StartStopBounds','StorageEnforcement','OfflineAnalysis','PrivacyAcknowledgement','PrivilegeSeparation')
    $blockers=[Collections.Generic.List[string]]::new()
    if($null -eq $Evidence){$blockers.Add('EVIDENCE_MISSING')}else{
        foreach($key in $Evidence.Keys){if($key -cnotin $required){$blockers.Add('UNKNOWN_KEY')}}
        foreach($key in $required){
            $record=$Evidence[$key]
            if($record -isnot [hashtable]){$blockers.Add("${key}:MISSING_RECORD");continue}
            $fields=@('Status','Kind','Source','Issues')
            if($record.Count -ne 4 -or @($record.Keys|Where-Object {$_ -cnotin $fields}).Count -gt 0){$blockers.Add("${key}:RECORD_SCHEMA");continue}
            if($record.Status -isnot [string] -or $record.Status -cne 'Verified'){$blockers.Add("${key}:NOT_VERIFIED")}
            if($record.Source -isnot [string] -or [string]::IsNullOrWhiteSpace($record.Source)){$blockers.Add("${key}:SOURCE_REQUIRED")}
            # Only profile semantics can be established statically in this contract.
            $kinds=if($key -ceq 'Profile'){@('Static','Observed')}else{@('Observed')}
            if($record.Kind -isnot [string] -or $record.Kind -cnotin $kinds){$blockers.Add("${key}:OBSERVATION_REQUIRED")}
            if($record.Issues -isnot [array]){$blockers.Add("${key}:ISSUES_REQUIRED")}elseif($record.Issues.Count -gt 0){$blockers.Add("${key}:UNRESOLVED_ISSUES")}
        }
    }
    return @{Status=$(if($blockers.Count){'BLOCKED'}else{'READY_FOR_RUNTIME_PLAN'});Blockers=$blockers.ToArray();RuntimeAuthorized=$false}
}
