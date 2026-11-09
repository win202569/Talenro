#requires -Version 7.0
[CmdletBinding()]
param([Parameter(Mandatory)][string]$OrdinaryAssembly,
      [Parameter(Mandatory)][string]$LayoutAssembly,
      [Parameter(Mandatory)][string]$CloseStageAssembly)
$ErrorActionPreference='Stop'
$paths=@($OrdinaryAssembly,$LayoutAssembly,$CloseStageAssembly | ForEach-Object { (Resolve-Path -LiteralPath $_).Path })
if (@($paths | Select-Object -Unique).Count -ne 3) { throw 'Three distinct assemblies required.' }
$required=@('CloseStageRecord','CloseStageBuffer','CloseStageScope','CloseStageSession','CloseStageReceiptReader','CloseStageActivation')
for ($i=0; $i -lt 3; $i++) {
    $stream=[IO.File]::OpenRead($paths[$i]); $pe=$null
    try {
        $pe=[Reflection.PortableExecutable.PEReader]::new($stream)
        $metadata=[Reflection.Metadata.PEReaderExtensions]::GetMetadataReader($pe)
        $types=@($metadata.TypeDefinitions | ForEach-Object {
            $type=$metadata.GetTypeDefinition($_)
            if ($metadata.GetString($type.Namespace) -eq 'Talenro.Demo.App.Layout.CloseStages') { $metadata.GetString($type.Name) }
        })
        $references=@($metadata.TypeReferences | ForEach-Object {
            $type=$metadata.GetTypeReference($_)
            $metadata.GetString($type.Namespace)
        })
        if ($i -lt 2) {
            if ($types.Count -ne 0 -or $references -contains 'Talenro.Demo.App.Layout.CloseStages') { throw 'Recorder leaked into non-recording assembly.' }
        } else {
            foreach ($name in $required) { if ($name -cnotin $types) { throw "Recorder type missing: $name" } }
        }
    } finally { if ($pe) { $pe.Dispose() }; $stream.Dispose() }
}
'PASS: static PE metadata isolation; no assembly executed, no runtime verdict.'
