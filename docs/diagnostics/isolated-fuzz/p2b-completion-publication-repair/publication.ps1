# Fixed private publication boundaries; no caller-supplied callbacks or sinks.
function Save-P2BIntegrationPublication([object]$State,[string]$Path,[int]$RowIndex,[object]$Row,[object]$Rows){
    if($null -ne $State.PrimaryError -or $State.SecondaryErrors.Count -ne 0 -or $State.SecondaryOverflow){return $false}
    # Read the existing ledger again; never rebuild it from stale in-memory rows.
    return (Complete-P2BLedger $State $Path $RowIndex $Row 'integration')
}
function Publish-P2BHostOutcome([object]$State,[string]$Path,[object]$Rows,[object]$Row,[string]$Raw){
    if($null -ne $State.PrimaryError -or $State.SecondaryErrors.Count -ne 0 -or $State.SecondaryOverflow){return $false}
    try{
        $rendered=$Row|ConvertTo-Json -Depth 12
        [Console]::Out.WriteLine($rendered)
        [Console]::Out.Write($Raw)
        [Console]::Out.Flush()
        # No normal output follows the authoritative completion write.
        return (Save-P2BLedger $State $Path $Rows)
    }catch{Add-P2BFailure $State $_;return $false}
}
