function Register-LSSink($Entry,[string]$Id,[string]$Path,[string]$Kind){
    if($Entry.Sinks.Count -ge 32 -or $Entry.Sinks.Contains($Id)){throw 'Sink limit/duplicate'}
    $canonical=if($Kind -ceq 'File'){[IO.Path]::GetFullPath($Path)}else{$Path}
    foreach($existing in $Entry.Sinks.Values){if($existing.Path -ieq $canonical){throw 'Path alias already registered'}}
    $Entry.Sinks[$Id]=[pscustomobject]@{Path=$canonical;Kind=$Kind;State='Healthy';LogicalIOAttempts=0;RejectedWithoutIO=0;Handle=$null}
}
function Enter-LSIO($Entry,[string]$Id,[string]$Kind){
    if($null -eq $Entry -or -not $Entry.Sinks.Contains($Id)){return $null}
    $sink=$Entry.Sinks[$Id]
    if($sink.Kind -cne $Kind){return $null}
    if($sink.State -ceq 'Failed' -or $Entry.FailureLatched -or $Entry.Status -ceq 'Closed'){
        $sink.RejectedWithoutIO++
        return $null
    }
    # One logical operation starts here; a refused access never increments it.
    $sink.LogicalIOAttempts++
    return $sink
}
function Fail-LSIO($Entry,$Sink,[Management.Automation.ErrorRecord]$Record){
    $Sink.State='Failed'
    Add-LSFailure $Entry.Handle $Record
}
function LS-IORoot([bool]$Ok,$Value,$ErrorRecord){[pscustomobject][ordered]@{Ok=$Ok;Value=$Value;Error=$ErrorRecord}}
function Write-LSLedger([object]$Session,[string]$SinkId,[object]$Value){
    $e=Find-LSSession $Session;$s=Enter-LSIO $e $SinkId File
    if($null -eq $s){return LS-IORoot $false $null $null}
    try{
        $text=ConvertTo-Json -InputObject $Value -Depth 30
        [IO.File]::WriteAllText($s.Path,$text,[Text.UTF8Encoding]::new($false))
        return LS-IORoot $true $null $null
    }catch{Fail-LSIO $e $s $_;return LS-IORoot $false $null $_}
}
function Read-LSLedger([object]$Session,[string]$SinkId){
    $e=Find-LSSession $Session;$s=Enter-LSIO $e $SinkId File
    if($null -eq $s){return LS-IORoot $false $null $null}
    try{$value=[IO.File]::ReadAllText($s.Path) | ConvertFrom-Json -AsHashtable;return LS-IORoot $true $value $null}
    catch{Fail-LSIO $e $s $_;return LS-IORoot $false $null $_}
}
function Complete-LSLedger([object]$Session,[string]$SinkId,[int]$RowIndex,[object]$Row){
    $e=Find-LSSession $Session;$s=Enter-LSIO $e $SinkId File
    if($null -eq $s){return LS-IORoot $false $null $null}
    try{
        $value=[IO.File]::ReadAllText($s.Path) | ConvertFrom-Json -AsHashtable
        if(-not $value.Contains('Rows') -or $value.Rows -isnot [array] -or $RowIndex -lt 0 -or $RowIndex -ge $value.Rows.Count){throw 'Invalid ledger row index'}
        $value.Rows[$RowIndex]=$Row
        [IO.File]::WriteAllText($s.Path,(ConvertTo-Json -InputObject $value -Depth 30),[Text.UTF8Encoding]::new($false))
        return LS-IORoot $true $null $null
    }catch{Fail-LSIO $e $s $_;return LS-IORoot $false $null $_}
}
function Write-LSOutput([object]$Session,[string]$SinkId,[string]$Text){
    $e=Find-LSSession $Session;$s=Enter-LSIO $e $SinkId Output
    if($null -eq $s){return LS-IORoot $false $null $null}
    try{
        $writer=if($SinkId -ceq 'ConsoleOut'){[Console]::Out}else{[Console]::Error}
        $writer.Write($Text);$writer.Flush()
        return LS-IORoot $true $null $null
    }catch{Fail-LSIO $e $s $_;return LS-IORoot $false $null $_}
}
function Open-LSStream($Entry,[string]$Id){
    $s=Enter-LSIO $Entry $Id File
    if($null -eq $s){return $null}
    try{$s.Handle=[IO.File]::Create($s.Path);return $s.Handle}catch{Fail-LSIO $Entry $s $_;return $null}
}
function Close-LSStream($Entry,[string]$Id){
    $s=$Entry.Sinks[$Id]
    if($null -eq $s.Handle){return}
    # Releasing this held object is not a second open/write/flush attempt.
    try{$s.Handle.Dispose()}catch{Fail-LSIO $Entry $s $_}finally{$s.Handle=$null}
}
function Read-LSBytes($Entry,[string]$Id){
    $s=Enter-LSIO $Entry $Id File
    if($null -eq $s){return LS-IORoot $false $null $null}
    try{return LS-IORoot $true ([IO.File]::ReadAllBytes($s.Path)) $null}
    catch{Fail-LSIO $Entry $s $_;return LS-IORoot $false $null $_}
}
