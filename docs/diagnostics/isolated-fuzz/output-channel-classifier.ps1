# Pure offline byte classification. No file reads, process launch or deserialization.
function Test-OfflineOutputExact($Left,$Right){
    if($Left -isnot [string] -or $Right -isnot [string]){return $false}
    return [string]::Equals($Left,$Right,[StringComparison]::Ordinal)
}
function Test-OfflineOutputExactMember($Value,[string[]]$Names){
    if($Value -isnot [string]){return $false}
    foreach($name in $Names){if([string]::Equals($Value,$name,[StringComparison]::Ordinal)){return $true}}
    return $false
}
function Get-OfflineOutputByteInfo([AllowNull()][AllowEmptyCollection()][byte[]]$Bytes){
    if($null -eq $Bytes){return [pscustomobject]@{ByteLength=$null;SHA256=$null}}
    $sha=[Security.Cryptography.SHA256]::Create()
    try{$hash=([BitConverter]::ToString($sha.ComputeHash($Bytes))).Replace('-','')}finally{$sha.Dispose()}
    [pscustomobject]@{ByteLength=$Bytes.Length;SHA256=$hash}
}
function Assert-OfflineOutputElement([Xml.XmlNode]$Node,[string]$Name,[string[]]$Attributes=@()){
    if($null -eq $Node -or $Node.NodeType -ne [Xml.XmlNodeType]::Element -or -not (Test-OfflineOutputExact $Node.LocalName $Name) -or -not (Test-OfflineOutputExact $Node.Prefix '') -or -not (Test-OfflineOutputExact $Node.NamespaceURI 'http://schemas.microsoft.com/powershell/2004/04')){throw 'Unexpected XML element'}
    if($Node.Attributes.Count -ne $Attributes.Count){throw 'Unexpected XML attributes'}
    foreach($attribute in $Node.Attributes){if(-not (Test-OfflineOutputExactMember $attribute.Name $Attributes)){throw 'Unexpected XML attribute'}}
}
function Assert-OfflineOutputText([Xml.XmlNode]$Node){
    foreach($child in $Node.ChildNodes){
        if($child.NodeType -notin @([Xml.XmlNodeType]::Text,[Xml.XmlNodeType]::Whitespace,[Xml.XmlNodeType]::SignificantWhitespace)){throw 'Unexpected scalar XML content'}
    }
}
function Assert-OfflineOutputReference([string]$Value){
    $number=[uint32]0
    if($Value -cnotmatch '^(0|[1-9][0-9]{0,9})\z' -or -not [uint32]::TryParse($Value,[ref]$number)){throw 'Invalid XML reference'}
}
function Assert-OfflineOutputProgress([Xml.XmlNode]$Node){
    Assert-OfflineOutputElement $Node 'MS'
    $members=@($Node.ChildNodes)
    if($members.Count -ne 2){throw 'Unexpected progress members'}
    Assert-OfflineOutputElement $members[0] 'I64' @('N')
    Assert-OfflineOutputText $members[0]
    $source=[long]0
    if(-not (Test-OfflineOutputExact ($members[0].GetAttribute('N')) 'SourceId') -or $members[0].InnerText -cnotmatch '^(0|[1-9][0-9]{0,18})\z' -or -not [long]::TryParse($members[0].InnerText,[ref]$source)){throw 'Invalid progress source'}
    Assert-OfflineOutputElement $members[1] 'PR' @('N')
    if(-not (Test-OfflineOutputExact ($members[1].GetAttribute('N')) 'Record')){throw 'Unexpected progress record'}
    $fields=@($members[1].ChildNodes);$names=@('AV','AI','Nil','PI','PC','T','SR','SD')
    if($fields.Count -ne $names.Count){throw 'Unexpected progress fields'}
    for($index=0;$index -lt $names.Count;$index++){
        Assert-OfflineOutputElement $fields[$index] $names[$index]
        Assert-OfflineOutputText $fields[$index]
        if($names[$index] -ceq 'Nil' -and $fields[$index].ChildNodes.Count -ne 0){throw 'Nonempty Nil field'}
    }
    if(-not (Test-OfflineOutputExactMember $fields[5].InnerText @('Completed','Processing'))){throw 'Unknown progress type'}
    foreach($index in @(1,3,4,6)){
        $value=[int]0;$text=$fields[$index].InnerText
        if($text -cnotmatch '^-?(0|[1-9][0-9]*)\z' -or -not [int]::TryParse($text,[ref]$value)){throw 'Invalid progress integer'}
        if(($index -eq 1 -and $value -lt 0) -or ($index -ne 1 -and $value -lt -1) -or ($index -eq 4 -and $value -gt 100)){throw 'Out of range progress integer'}
    }
}
function Get-OfflineOutputProgressCount([string]$Text){
    if($Text.Length -gt 0 -and $Text[0] -eq [char]0xFEFF){$Text=$Text.Substring(1)}
    if(-not $Text.StartsWith("#< CLIXML`r`n",[StringComparison]::Ordinal) -and -not $Text.StartsWith("#< CLIXML`n",[StringComparison]::Ordinal)){throw 'CLIXML header required'}
    $xml=$Text.Substring($Text.IndexOf("`n")+1)
    $settings=New-Object Xml.XmlReaderSettings
    $settings.DtdProcessing=[Xml.DtdProcessing]::Prohibit;$settings.XmlResolver=$null
    $settings.ConformanceLevel=[Xml.ConformanceLevel]::Document
    $settings.IgnoreWhitespace=$true;$settings.IgnoreComments=$false;$settings.IgnoreProcessingInstructions=$false
    $settings.MaxCharactersInDocument=1048576
    $inputReader=New-Object IO.StringReader($xml);$reader=$null
    try{
        $reader=[Xml.XmlReader]::Create($inputReader,$settings)
        $document=New-Object Xml.XmlDocument;$document.XmlResolver=$null
        $document.Load($reader)
        if($document.ChildNodes.Count -ne 1){throw 'Unexpected document content'}
        $root=$document.DocumentElement
        Assert-OfflineOutputElement $root 'Objs' @('Version','xmlns')
        if(-not (Test-OfflineOutputExact ($root.GetAttribute('Version')) '1.1.0.1') -or -not (Test-OfflineOutputExact ($root.GetAttribute('xmlns')) 'http://schemas.microsoft.com/powershell/2004/04')){throw 'Unexpected CLIXML envelope'}
        $objects=@($root.ChildNodes)
        if($objects.Count -eq 0){throw 'Empty CLIXML envelope'}
        $objectIds=@{};$typeIds=@{}
        foreach($node in $objects){
            Assert-OfflineOutputElement $node 'Obj' @('S','RefId')
            if(-not (Test-OfflineOutputExact ($node.GetAttribute('S')) 'progress')){throw 'Non-progress channel refused'}
            $id=$node.GetAttribute('RefId');Assert-OfflineOutputReference $id
            if($objectIds.ContainsKey($id)){throw 'Duplicate object reference'};$objectIds[$id]=$true
            $children=@($node.ChildNodes)
            if($children.Count -ne 2){throw 'Unexpected progress object'}
            $type=$children[0];$typeId=$type.GetAttribute('RefId');Assert-OfflineOutputReference $typeId
            if(Test-OfflineOutputExact $type.LocalName 'TN'){
                Assert-OfflineOutputElement $type 'TN' @('RefId')
                if($typeIds.ContainsKey($typeId)){throw 'Duplicate type reference'}
                $types=@($type.ChildNodes)
                if($types.Count -ne 2){throw 'Unexpected progress type list'}
                for($index=0;$index -lt 2;$index++){Assert-OfflineOutputElement $types[$index] 'T';Assert-OfflineOutputText $types[$index]}
                if(-not (Test-OfflineOutputExact $types[0].InnerText 'System.Management.Automation.PSCustomObject') -or -not (Test-OfflineOutputExact $types[1].InnerText 'System.Object')){throw 'Unknown progress object type'}
                $typeIds[$typeId]=$true
            }else{
                Assert-OfflineOutputElement $type 'TNRef' @('RefId')
                if($type.ChildNodes.Count -ne 0 -or -not $typeIds.ContainsKey($typeId)){throw 'Unresolved type reference'}
            }
            Assert-OfflineOutputProgress $children[1]
        }
        return $objects.Count
    }finally{if($null -ne $reader){$reader.Dispose()};$inputReader.Dispose()}
}
function Get-OfflineOutputClassification(
    [AllowNull()][AllowEmptyCollection()][byte[]]$StdoutBytes,
    [AllowNull()][AllowEmptyCollection()][byte[]]$StderrBytes,
    [object]$ExitCode,[string]$EncodingName
){
    Set-StrictMode -Version Latest
    $result=[pscustomobject]@{Schema=1;Classification='refused';Reason='unclassified';ProgressCount=0;EncodingName=$EncodingName;ExitCode=$ExitCode;Stdout=(Get-OfflineOutputByteInfo $StdoutBytes);Stderr=(Get-OfflineOutputByteInfo $StderrBytes);Acceptance=$false;RuntimeApproved=$false}
    try{
        if($null -eq $StdoutBytes -or $null -eq $StderrBytes){throw 'Missing stream bytes'}
        if($StdoutBytes.Length -gt 1048576 -or $StderrBytes.Length -gt 1048576){throw 'Stream exceeds 1MiB bound'}
        if(($ExitCode -isnot [int] -and $ExitCode -isnot [long]) -or $ExitCode -lt [int]::MinValue -or $ExitCode -gt [int]::MaxValue){throw 'Integer child exit required'}
        if($ExitCode -ne 0){throw 'Nonzero child exit'}
        if(Test-OfflineOutputExact $EncodingName 'utf-8'){$encoding=New-Object Text.UTF8Encoding($false,$true)}
        elseif(Test-OfflineOutputExact $EncodingName 'utf-16le'){$encoding=New-Object Text.UnicodeEncoding($false,$false,$true)}
        else{throw 'Explicit supported encoding required'}
        [void]$encoding.GetString($StdoutBytes)
        $text=$encoding.GetString($StderrBytes)
        if($StderrBytes.Length -eq 0){$result.Classification='empty'}
        else{$result.ProgressCount=Get-OfflineOutputProgressCount $text;$result.Classification='progress-only'}
        $result.Reason='classified-only'
    }catch{$result.Classification='refused';$result.ProgressCount=0;$result.Reason=$_.Exception.Message}
    return $result
}
