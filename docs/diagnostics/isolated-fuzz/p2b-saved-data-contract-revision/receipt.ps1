# Fixed new sink only. Structural assurance; compound I/O fault not induced.
$script:ReceiptAttempted=$false
$script:ReceiptFailed=$false
$script:ReceiptSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
function Write-ODSavedDataReceipt {
    param([byte[]]$Bytes)
    if($script:ReceiptAttempted){throw 'ODR: receipt attempt already consumed'}
    $script:ReceiptAttempted=$true
    $receiptPath='D:\Projects\Talenro\.worktrees\c12-b01-task9-coordinator\.superpowers\sdd\2026-11-01-offline-checker-unified-revision\result.json'
    $heldStream=$null;$firstCause=$null
    try{
        $heldStream=[IO.File]::Open($receiptPath,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::Read)
        $heldStream.Write($Bytes,0,$Bytes.Length)
        $heldStream.Flush()
    }catch{$firstCause=$_;$script:ReceiptFailed=$true}
    finally{
        if($null -ne $heldStream){
            try{$heldStream.Dispose()}
            catch{
                $script:ReceiptFailed=$true
                if($null -eq $firstCause){$firstCause=$_}
                elseif($script:ReceiptSecondary.Count -lt 8){$script:ReceiptSecondary.Add($_)}
            }
        }
    }
    if($null -ne $firstCause){throw $firstCause}
}
