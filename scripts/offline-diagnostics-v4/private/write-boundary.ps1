# Private live outcomes; collection belongs only to the calling boundary.
function New-ODRecordStream([string]$Path){
    [IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
}

function Write-ODCaptureRecord([string]$Path,[byte[]]$Bytes){
    if([string]::IsNullOrEmpty($Path) -or $null -eq $Bytes){throw 'OD: invalid record arguments'}
    $stream=$null;$primary=$null
    [Management.Automation.ErrorRecord[]]$secondary=@()
    try{
        try{$stream=New-ODRecordStream $Path}catch{$primary=$_}
        if($null -ne $stream){
            try{$stream.Write($Bytes,0,$Bytes.Length)}catch{$primary=$_}
            if($null -eq $primary){try{$stream.Flush()}catch{$primary=$_}}
        }
    }finally{
        if($null -ne $stream){
            try{$stream.Dispose()}catch{
                if($null -eq $primary){$primary=$_}else{$secondary=[Management.Automation.ErrorRecord[]]@($_)}
            }
        }
    }
    [pscustomobject][ordered]@{Succeeded=($null -eq $primary);PrimaryError=$primary;SecondaryErrors=$secondary}
}

function Write-ODProfileIntentRecord([string]$Path,[byte[]]$Bytes){
    if([string]::IsNullOrEmpty($Path) -or $null -eq $Bytes){throw 'OD: invalid record arguments'}
    $stream=$null;$primary=$null
    [Management.Automation.ErrorRecord[]]$secondary=@()
    try{
        try{$stream=New-ODRecordStream $Path}catch{$primary=$_}
        if($null -ne $stream){
            try{$stream.Write($Bytes,0,$Bytes.Length)}catch{$primary=$_}
            if($null -eq $primary){try{$stream.Flush()}catch{$primary=$_}}
        }
    }finally{
        if($null -ne $stream){
            try{$stream.Dispose()}catch{
                if($null -eq $primary){$primary=$_}else{$secondary=[Management.Automation.ErrorRecord[]]@($_)}
            }
        }
    }
    [pscustomobject][ordered]@{Succeeded=($null -eq $primary);PrimaryError=$primary;SecondaryErrors=$secondary}
}
