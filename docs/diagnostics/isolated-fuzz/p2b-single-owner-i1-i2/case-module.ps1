param([string]$WriterPath,[string]$CollectorText)
$caseModule=New-Module -ArgumentList $WriterPath,$CollectorText -ScriptBlock {
    param($WriterPath,$CollectorText)
    . $WriterPath
    . ([ScriptBlock]::Create($CollectorText))
    $script:OriginalFactory=(Get-Command New-ODRecordStream).ScriptBlock
    $script:HostPrimary=$null
    $script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
    $script:HostOverflow=$false
    Export-ModuleMember
}
