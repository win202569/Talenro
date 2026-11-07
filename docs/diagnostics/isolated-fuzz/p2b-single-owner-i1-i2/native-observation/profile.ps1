# Definition only. This validates a fixed candidate; it never controls recording.
function Test-C12ObservationProfile([string]$Xml,[string]$SchemaPath) {
    $answer=@{Valid=$false;DeclaredBufferBytes=[long]0;Reasons=@()}
    if(-not [IO.File]::Exists($SchemaPath)){$answer.Reasons=@('SCHEMA_MISSING');return $answer}
    if([string]::IsNullOrWhiteSpace($Xml) -or $Xml.Length -gt 65536){$answer.Reasons=@('XML_LIMIT');return $answer}
    $reader=$null;$schemaReader=$null
    try {
        $settings=[Xml.XmlReaderSettings]::new();$settings.DtdProcessing=[Xml.DtdProcessing]::Prohibit;$settings.XmlResolver=$null
        $settings.MaxCharactersInDocument=65536
        $reader=[Xml.XmlReader]::Create([IO.StringReader]::new($Xml),$settings)
        $doc=[Xml.XmlDocument]::new();$doc.XmlResolver=$null;$doc.Load($reader)
        $schemaSettings=[Xml.XmlReaderSettings]::new();$schemaSettings.DtdProcessing=[Xml.DtdProcessing]::Prohibit;$schemaSettings.XmlResolver=$null
        $schemaSettings.MaxCharactersInDocument=1048576
        $schemaReader=[Xml.XmlReader]::Create($SchemaPath,$schemaSettings)
        $schemas=[Xml.Schema.XmlSchemaSet]::new();$schemas.XmlResolver=$null
        $null=$schemas.Add($null,$schemaReader);$schemas.Compile();$doc.Schemas=$schemas
        $doc.Validate($null)
    } catch {$answer.Reasons=@('XML_OR_SCHEMA_INVALID');return $answer}
    finally {if($null -ne $reader){$reader.Dispose()};if($null -ne $schemaReader){$schemaReader.Dispose()}}
    # Exact element/attribute allowlist prevents schema-valid inherited or extra providers.
    $p='/WindowsPerformanceRecorder/Profiles'
    $rules=@(
        @{Path='/WindowsPerformanceRecorder';Attrs=@{Version='1.0'}},
        @{Path=$p;Attrs=@{}},
        @{Path="$p/SystemCollector";Attrs=@{Id='C12SystemCollector';Name='C12 Native Observation'}},
        @{Path="$p/SystemCollector/BufferSize";Attrs=@{Value='1024'}},
        @{Path="$p/SystemCollector/Buffers";Attrs=@{Value='128'}},
        @{Path="$p/SystemProvider";Attrs=@{Id='C12SystemProvider'}},
        @{Path="$p/SystemProvider/Keywords";Attrs=@{}},
        @{Path="$p/Profile";Attrs=@{Id='C12NativeObservation.Light.Memory';Name='C12NativeObservation';Description='Bounded native observation candidate';LoggingMode='Memory';DetailLevel='Light'}},
        @{Path="$p/Profile/Collectors";Attrs=@{}},
        @{Path="$p/Profile/Collectors/SystemCollectorId";Attrs=@{Value='C12SystemCollector'}},
        @{Path="$p/Profile/Collectors/SystemCollectorId/SystemProviderId";Attrs=@{Value='C12SystemProvider'}}
    )
    foreach($key in @('ProcessThread','Loader','SampledProfile','FileIO','FileIOInit')){$rules+=@{Path="$p/SystemProvider/Keywords/Keyword[@Value='$key']";Attrs=@{Value=$key}}}
    if($doc.SelectNodes('//*').Count -ne $rules.Count){$answer.Reasons=@('PROFILE_SHAPE');return $answer}
    foreach($rule in $rules){
        $nodes=$doc.SelectNodes($rule.Path)
        if($nodes.Count -ne 1){$answer.Reasons=@('PROFILE_SHAPE');return $answer}
        $node=$nodes[0]
        if($node.Attributes.Count -ne $rule.Attrs.Count){$answer.Reasons=@('PROFILE_ATTRIBUTES');return $answer}
        foreach($key in $rule.Attrs.Keys){if($node.GetAttribute($key) -cne $rule.Attrs[$key]){$answer.Reasons=@('PROFILE_VALUE');return $answer}}
    }
    $answer.Valid=$true;$answer.DeclaredBufferBytes=[long]1024*128*1024
    return $answer
}
