# Definitions only: callers supply memory/diagnostic streams; no top-level I/O.
function Get-WPTPhases {
 return @('task.returned','pins.begin','pins.end','map.save.begin','map.save.end','map.meta.end','role.save.begin','role.save.end','exit.ready')
}
function New-WPTState([string]$Token,[string]$InputSHA256,[long]$Frequency){
 if($Token -cnotmatch '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' -or $InputSHA256 -cnotmatch '^[A-F0-9]{64}$' -or $Frequency -le 0){throw 'WPT identity/frequency'}
 return [ordered]@{Token=$Token;InputSHA256=$InputSHA256;Frequency=$Frequency;Seq=0;Bytes=0;Failed=$false;Terminal=$false;LastTicks=0L}
}
function ConvertTo-WPTRecord($State,[string]$Phase,[long]$ElapsedTicks){
 $null=New-WPTState $State.Token $State.InputSHA256 $State.Frequency
 if($State.Failed -or $State.Terminal -or $State.Seq -lt 0 -or $State.Seq -ge 10 -or $State.Bytes -lt 0 -or $ElapsedTicks -lt 0 -or $ElapsedTicks -lt $State.LastTicks){throw 'WPT state'}
 $phases=Get-WPTPhases
 if($Phase -cne 'error' -and ($State.Seq -ge 9 -or $Phase -cne $phases[$State.Seq])){throw 'WPT phase order'}
 $value=[ordered]@{Schema=1;Token=$State.Token;InputSHA256=$State.InputSHA256;Role='writer-timing';Seq=($State.Seq+1);Phase=$Phase;ElapsedTicks=$ElapsedTicks;Frequency=$State.Frequency}
 $bytes=[Text.UTF8Encoding]::new($false,$true).GetBytes((ConvertTo-Json -InputObject $value -Compress)+[char]10)
 if($bytes.Length -gt 384 -or $State.Bytes+$bytes.Length -gt 3840){throw 'WPT record budget'}
 return ,$bytes
}
function Write-WPTRecord($State,[IO.Stream]$Stream,[string]$Phase,[long]$ElapsedTicks){
 if($State.Failed -or $State.Terminal){return $false}
 try{
  $bytes=ConvertTo-WPTRecord $State $Phase $ElapsedTicks
  $State.Seq++;$State.Bytes+=$bytes.Length;$State.LastTicks=$ElapsedTicks
  if($Phase -ceq 'error'){$State.Terminal=$true}
  $Stream.Write($bytes,0,$bytes.Length);$Stream.Flush()
  return $true
 }catch{$State.Failed=$true;return $false}
}
function Read-WPTTrace([byte[]]$Bytes,[string]$Token,[string]$InputSHA256){
 $records=[Collections.Generic.List[object]]::new();$intervals=[Collections.Generic.List[object]]::new()
 $result=[ordered]@{Status='INVALID';Records=@();RawTailBase64='';Reason=$null;Intervals=@()}
 try{
  $null=New-WPTState $Token $InputSHA256 1
  if($null -eq $Bytes){throw 'WPT null bytes'}
  if($Bytes.Length -gt 3840){throw 'WPT total bound'}
  $phases=Get-WPTPhases;$encoding=[Text.UTF8Encoding]::new($false,$true)
  $start=0;$last=0L;$frequency=0L;$terminal=$false
  for($i=0;$i -lt $Bytes.Length;$i++){
   if($Bytes[$i] -ne 10){continue}
   $length=$i-$start+1
   if($length -gt 384 -or $length -lt 2 -or $records.Count -ge 10 -or $terminal){throw 'WPT line/count/order bound'}
   $line=$encoding.GetString($Bytes,$start,$length-1);$start=$i+1
   if($line.Contains([char]13)){throw 'WPT CR forbidden'}
   $doc=[Text.Json.JsonDocument]::Parse($line)
   try{
    if($doc.RootElement.ValueKind -ne [Text.Json.JsonValueKind]::Object){throw 'WPT object required'}
    $properties=@($doc.RootElement.EnumerateObject());$seen=[Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    $expected=@('Schema','Token','InputSHA256','Role','Seq','Phase','ElapsedTicks','Frequency');$record=[ordered]@{}
    if($properties.Count -ne 8){throw 'WPT field count'}
    foreach($property in $properties){
     if(-not $seen.Add($property.Name) -or $property.Name -cnotin $expected){throw 'WPT duplicate/unknown field'}
     if($property.Name -cin @('Schema','Seq','ElapsedTicks','Frequency')){
      if($property.Value.ValueKind -ne [Text.Json.JsonValueKind]::Number -or $property.Value.GetRawText() -cnotmatch '^(0|[1-9][0-9]*)$'){throw 'WPT integer required'}
      $record[$property.Name]=$property.Value.GetInt64()
     }else{
      if($property.Value.ValueKind -ne [Text.Json.JsonValueKind]::String){throw 'WPT string required'}
      $record[$property.Name]=$property.Value.GetString()
     }
    }
   }finally{$doc.Dispose()}
   if($record.Schema -ne 1 -or $record.Token -cne $Token -or $record.InputSHA256 -cne $InputSHA256 -or $record.Role -cne 'writer-timing'){throw 'WPT identity'}
   if($record.Seq -ne $records.Count+1 -or $record.ElapsedTicks -lt $last -or $record.Frequency -le 0){throw 'WPT sequence/time'}
   if($records.Count -eq 0){$frequency=$record.Frequency}elseif($record.Frequency -ne $frequency){throw 'WPT frequency changed'}
   if($record.Phase -ceq 'error'){$terminal=$true}elseif($records.Count -ge 9 -or $record.Phase -cne $phases[$records.Count]){throw 'WPT phase order'}
   $last=$record.ElapsedTicks;$records.Add($record)
  }
  if($start -lt $Bytes.Length){
   $tail=[byte[]]::new($Bytes.Length-$start);[Array]::Copy($Bytes,$start,$tail,0,$tail.Length)
   $result.RawTailBase64=[Convert]::ToBase64String($tail)
   if($tail.Length -ge 384 -or $terminal){throw 'WPT invalid trailing bytes'}
  }
  foreach($pair in @(@('pins.begin','pins.end'),@('map.save.begin','map.save.end'),@('role.save.begin','role.save.end'))){
   $begin=@($records|Where-Object {$_.Phase -ceq $pair[0]});$end=@($records|Where-Object {$_.Phase -ceq $pair[1]})
   if($begin.Count -eq 1 -and $end.Count -eq 1){$delta=$end[0].ElapsedTicks-$begin[0].ElapsedTicks;$intervals.Add([ordered]@{Begin=$pair[0];End=$pair[1];ElapsedTicks=$delta;Milliseconds=(1000.0*$delta/$frequency);IncludesInstrumentation=$true})}
  }
  $result.Status=if($terminal){'ERROR'}elseif($records.Count -eq 9 -and $start -eq $Bytes.Length){'COMPLETE'}else{'PARTIAL'}
 }catch{$result.Status='INVALID';$result.Reason=$_.Exception.Message}
 $result.Records=$records.ToArray();$result.Intervals=$intervals.ToArray()
 return $result
}
