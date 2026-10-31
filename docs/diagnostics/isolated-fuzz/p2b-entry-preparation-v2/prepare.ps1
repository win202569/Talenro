[CmdletBinding(PositionalBinding=$false)]
param()
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'

function Convert-ODInheritedPin($Pin){
    if($null -eq $Pin){throw 'OD: required pin identity invalid'}
    $path=$Pin.PSObject.Properties['VerificationPath']
    $hash=$Pin.PSObject.Properties['SHA256']
    if($null -eq $path -or $null -eq $hash -or
       $path.Value -isnot [string] -or [string]::IsNullOrWhiteSpace($path.Value) -or
       $hash.Value -isnot [string] -or $hash.Value -cnotmatch '^[A-F0-9]{64}$'){
        throw 'OD: required pin identity invalid'
    }
    [pscustomobject][ordered]@{
        VerificationPath=[IO.Path]::GetFullPath($path.Value)
        SHA256=$hash.Value
        Role=$Pin.Role
        ExpectedState='Present'
    }
}

throw 'OD: preparation body not implemented'
