# Seed: mechanically preserve the two reviewed failure orderings for RED.
function New-ODRecordStream([string]$Path){
    return [IO.File]::Open($Path,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
}
function Write-ODCaptureRecord([string]$Path,[byte[]]$Bytes){
    $stream=$null;$primary=$null
    try{
        $stream=New-ODRecordStream $Path
        $stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()
    }catch{$primary=$_}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{if($null -eq $primary){$primary=$_}else{Capture-Failure $_}}}}
    if($null -ne $primary){throw $primary}
}
function Write-ODProfileIntentRecord([string]$Path,[byte[]]$Bytes){
    $stream=$null
    try{
        $stream=New-ODRecordStream $Path
        $stream.Write($Bytes,0,$Bytes.Length);$stream.Flush()
    }catch{Capture-Failure $_;throw}
    finally{if($null -ne $stream){try{$stream.Dispose()}catch{Capture-Failure $_;throw}}}
}
