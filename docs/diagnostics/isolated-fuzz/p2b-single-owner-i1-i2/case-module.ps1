function New-ODCaseModule([string]$WriterPath,[string]$CollectorText){
    return New-Module -ArgumentList $WriterPath,$CollectorText -ScriptBlock {
        param($WriterPath,$CollectorText)
        Set-StrictMode -Version Latest
        $ErrorActionPreference='Stop'
        . $WriterPath
        . (Join-Path ([IO.Path]::GetDirectoryName($WriterPath)) record-owner.ps1)
        . ([ScriptBlock]::Create($CollectorText))
        $script:OriginalFactory=(Get-Command New-ODRecordStream).ScriptBlock
        $script:HostPrimary=$null
        $script:HostSecondary=[Collections.Generic.List[Management.Automation.ErrorRecord]]::new()
        $script:HostOverflow=$false
        Export-ModuleMember
    }
}
