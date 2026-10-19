function New-P2BFailureState {
    [pscustomobject][ordered]@{PrimaryError=$null;SecondaryErrors=[Management.Automation.ErrorRecord[]]@();SecondaryOverflow=$false;Outcome='INCOMPLETE';FinalLedgerWritten=$false}
}
function Add-P2BFailure([object]$State,[Management.Automation.ErrorRecord]$Record){
    if($null -eq $Record){return}
    $State.Outcome='INCOMPLETE';$State.FinalLedgerWritten=$false
    if($null -eq $State.PrimaryError){$State.PrimaryError=$Record;return}
    if([object]::ReferenceEquals($State.PrimaryError,$Record)){return}
    foreach($saved in $State.SecondaryErrors){if([object]::ReferenceEquals($saved,$Record)){return}}
    if($State.SecondaryErrors.Count -lt 8){$State.SecondaryErrors=[Management.Automation.ErrorRecord[]](@($State.SecondaryErrors)+@($Record))}else{$State.SecondaryOverflow=$true}
}
function Save-P2BLedger([object]$State,[string]$Path,[object]$Rows){
    try{
        $base=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../.superpowers/sdd/2026-10-19-p2b-completion-publication-repair'))+[IO.Path]::DirectorySeparatorChar
        $full=[IO.Path]::GetFullPath($Path)
        if(-not $full.StartsWith($base,[StringComparison]::OrdinalIgnoreCase)){throw 'Fixed packet paths only'}
        [IO.File]::WriteAllText($full,($Rows|ConvertTo-Json -Depth 30),[Text.UTF8Encoding]::new($false))
        return $true
    }catch{Add-P2BFailure $State $_;return $false}
}
function Complete-P2BLedger([object]$State,[string]$Path,[int]$RowIndex,[object]$Row,
    [ValidateSet('child','integration','parent')][string]$Scope){
    try{
        $base=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../../.superpowers/sdd/2026-10-19-p2b-completion-publication-repair'))+[IO.Path]::DirectorySeparatorChar
        $full=[IO.Path]::GetFullPath($Path)
        if(-not $full.StartsWith($base,[StringComparison]::OrdinalIgnoreCase)){throw 'Fixed packet paths only'}
        $rows=@([IO.File]::ReadAllText($full)|ConvertFrom-Json)
        if($RowIndex -lt 0 -or $RowIndex -ge $rows.Count){throw 'Existing ledger row required'}
        $rows[$RowIndex]=$Row
        $ok=Save-P2BLedger $State $full $rows
        $State.FinalLedgerWritten=$ok
        return $ok
    }catch{Add-P2BFailure $State $_;return $false}
}
function Write-P2BFailureFallback([object]$State){
    try{
        $records=@();if($null -ne $State.PrimaryError){$records+=,$State.PrimaryError};$records+=@($State.SecondaryErrors)
        $items=[Collections.Generic.List[string]]::new()
        foreach($record in $records){
            $text=$record.Exception.GetType().FullName+': '+$record.Exception.Message
            $items.Add($text.Substring(0,[Math]::Min(1024,$text.Length)))
        }
        [Console]::Error.WriteLine('INCOMPLETE: '+($items.ToArray() -join ' | ')+' overflow='+$State.SecondaryOverflow)
    }catch{Add-P2BFailure $State $_}
}
