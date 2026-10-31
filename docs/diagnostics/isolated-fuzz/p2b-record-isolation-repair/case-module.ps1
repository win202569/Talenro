param([string]$WriterPath,[string]$CollectorText)
# Seed retains the diagnosed default-export defect; the only GREEN edit is below.
$riCaseModule=New-Module -ArgumentList $WriterPath,$CollectorText -ScriptBlock {
    param($WriterPath,$CollectorText)
    . $WriterPath
    . ([ScriptBlock]::Create($CollectorText))
    $script:OriginalFactory=(Get-Command New-ODRecordStream).ScriptBlock
    Export-ModuleMember
}
