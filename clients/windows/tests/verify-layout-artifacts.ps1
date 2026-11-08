#requires -Version 7.0
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$OrdinaryAssembly,
    [Parameter(Mandatory)][string]$DiagnosticAssembly
)
$ErrorActionPreference = 'Stop'

function Read-LayoutMetadata([string]$AssemblyPath) {
    $resolved = (Resolve-Path -LiteralPath $AssemblyPath).Path
    $stream = [System.IO.File]::OpenRead($resolved)
    $pe = $null
    try {
        $pe = [System.Reflection.PortableExecutable.PEReader]::new($stream)
        if (-not $pe.HasMetadata) { throw 'Expected a managed App assembly, not its native executable.' }
        $metadata = [System.Reflection.Metadata.PEReaderExtensions]::GetMetadataReader($pe)
        $types = @($metadata.TypeDefinitions | ForEach-Object {
            $definition = $metadata.GetTypeDefinition($_)
            if ($metadata.GetString($definition.Namespace) -eq 'Talenro.Demo.App.Layout') {
                $metadata.GetString($definition.Name)
            }
        })
        $resources = @($metadata.ManifestResources | ForEach-Object {
            $metadata.GetString($metadata.GetManifestResource($_).Name)
        })
        [pscustomobject]@{ Path=$resolved; Types=$types; Resources=$resources }
    }
    finally {
        if ($null -ne $pe) { $pe.Dispose() }
        $stream.Dispose()
    }
}

$ordinary = Read-LayoutMetadata $OrdinaryAssembly
$diagnostic = Read-LayoutMetadata $DiagnosticAssembly
if ($ordinary.Path -eq $diagnostic.Path) { throw 'Ordinary and diagnostic outputs must differ.' }
$requiredTypes = @('LayoutOptions','LayoutParseResult','LayoutFixtureCatalog','LayoutPresentation',
    'LayoutResourceCatalog','LayoutStaticBindings','LayoutMeasurement','LayoutTitle','LayoutWindowObserver')
$requiredResources = @('LayoutFixtures.json','LayoutStrings.zh-Hans.xml','LayoutStrings.en-US.xml')
if (@($ordinary.Types | Where-Object { $_ -ne 'ReservedLayoutArguments' }).Count -ne 0) {
    throw 'Diagnostic types leaked into ordinary App.'
}
if (@($ordinary.Resources | Where-Object { $_ -like 'Layout*' }).Count -ne 0) {
    throw 'Diagnostic resources leaked into ordinary App.'
}
foreach ($name in $requiredTypes) {
    if ($name -cnotin $diagnostic.Types) { throw "Missing diagnostic type: $name" }
}
foreach ($name in $requiredResources) {
    if ($name -cnotin $diagnostic.Resources) { throw "Missing diagnostic resource: $name" }
}
if ('ReservedLayoutArguments' -cnotin $ordinary.Types) { throw 'Ordinary argument guard is missing.' }
[pscustomobject]@{
    Result='PASS - static metadata isolation only; no assembly executed'
    Ordinary=$ordinary.Path
    Diagnostic=$diagnostic.Path
    DiagnosticTypeCount=$requiredTypes.Count
    DiagnosticResourceCount=$requiredResources.Count
}
