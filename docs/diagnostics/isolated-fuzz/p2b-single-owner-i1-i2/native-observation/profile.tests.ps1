param([ValidateSet('Red','Green')][string]$Phase='Green',[string]$SchemaPath='C:/Program Files (x86)/Windows Kits/10/Windows Performance Toolkit/WPRControlProfiles.xsd')
$ErrorActionPreference='Stop'
# Mutations catch weakened allowlists, unsafe XML parsing and capacity arithmetic.
$candidate=@'
<WindowsPerformanceRecorder Version="1.0"><Profiles>
<SystemCollector Id="C12SystemCollector" Name="C12 Native Observation"><BufferSize Value="1024"/><Buffers Value="128"/></SystemCollector>
<SystemProvider Id="C12SystemProvider"><Keywords><Keyword Value="ProcessThread"/><Keyword Value="Loader"/><Keyword Value="SampledProfile"/><Keyword Value="FileIO"/><Keyword Value="FileIOInit"/></Keywords></SystemProvider>
<Profile Id="C12NativeObservation.Light.Memory" Name="C12NativeObservation" Description="Bounded native observation candidate" LoggingMode="Memory" DetailLevel="Light"><Collectors><SystemCollectorId Value="C12SystemCollector"><SystemProviderId Value="C12SystemProvider"/></SystemCollectorId></Collectors></Profile>
</Profiles></WindowsPerformanceRecorder>
'@
if(-not(Test-Path $SchemaPath)){throw 'TEST_SCHEMA_MISSING'}
if(Test-Path (Join-Path $PSScriptRoot 'profile.ps1')){. (Join-Path $PSScriptRoot 'profile.ps1')}
if(-not(Get-Command Test-C12ObservationProfile -ErrorAction SilentlyContinue)){throw 'ASSERT_PROFILE_POLICY_ABSENT'}
$cases=@(
 @{Name='candidate';Xml=$candidate;Valid=$true},
 @{Name='over-cap';Xml=$candidate.Replace('Value="128"','Value="129"');Valid=$false},
 @{Name='file-mode';Xml=$candidate.Replace('LoggingMode="Memory"','LoggingMode="File"');Valid=$false},
 @{Name='extra-keyword';Xml=$candidate.Replace('</Keywords>','<Keyword Value="NetworkTrace"/></Keywords>');Valid=$false},
 @{Name='extra-provider';Xml=$candidate.Replace('</Profiles>','<EventProvider Id="extra" Name="extra"/></Profiles>');Valid=$false},
 @{Name='stacks';Xml=$candidate.Replace('</SystemProvider>','<Stacks><Stack Value="FileRead"/></Stacks></SystemProvider>');Valid=$false},
 @{Name='inheritance';Xml=$candidate.Replace('<Profile Id=','<Profile Base="other" Id=');Valid=$false},
 @{Name='percentage';Xml=$candidate.Replace('<Buffers Value=','<Buffers PercentageOfTotalMemory="true" Value=');Valid=$false},
 @{Name='per-cpu';Xml=$candidate.Replace('<Buffers Value="128"/>','<BuffersPerCPU Value="128"/>');Valid=$false},
 @{Name='duplicate-id';Xml=$candidate.Replace('</Profiles>','<SystemProvider Id="C12SystemProvider"/></Profiles>');Valid=$false},
 @{Name='unresolved-id';Xml=$candidate.Replace('SystemProviderId Value="C12SystemProvider"','SystemProviderId Value="missing"');Valid=$false},
 @{Name='malformed';Xml='<bad';Valid=$false},
 @{Name='dtd';Xml='<!DOCTYPE WindowsPerformanceRecorder [<!ENTITY a "x">]>'+$candidate;Valid=$false},
 @{Name='external-entity';Xml='<!DOCTYPE WindowsPerformanceRecorder [<!ENTITY a SYSTEM "file:///C:/nonexistent-c12">]>'+$candidate;Valid=$false},
 @{Name='zero-size';Xml=$candidate.Replace('Value="1024"','Value="0"');Valid=$false},
 @{Name='negative-count';Xml=$candidate.Replace('Value="128"','Value="-1"');Valid=$false},
 @{Name='overflow';Xml=$candidate.Replace('Value="1024"','Value="999999999999999999999999"');Valid=$false},
 @{Name='extra-profile';Xml=$candidate.Replace('</Profiles>','<Profile Id="extra" Name="extra" LoggingMode="Memory" DetailLevel="Light" Description="extra"/></Profiles>');Valid=$false},
 @{Name='unknown-attribute';Xml=$candidate.Replace('<Buffers Value=','<Buffers Unexpected="true" Value=');Valid=$false},
 @{Name='missing-keyword';Xml=$candidate.Replace('<Keyword Value="FileIOInit"/>','');Valid=$false},
 @{Name='missing-schema';Xml=$candidate;Schema='C:/nonexistent-c12-schema.xsd';Valid=$false}
)
foreach($c in $cases){$schema=if($c.ContainsKey('Schema')){$c.Schema}else{$SchemaPath};$r=Test-C12ObservationProfile $c.Xml $schema;if($r.Valid -isnot [bool] -or $r.Valid -ne $c.Valid){throw "ASSERT_$($c.Name): $($r|ConvertTo-Json -Compress)"};if($c.Valid){if($r.DeclaredBufferBytes -ne 134217728 -or @($r.Reasons).Count -ne 0){throw 'ASSERT_CAPACITY'}}elseif(@($r.Reasons).Count -eq 0){throw 'ASSERT_REASON'};"PASS $($c.Name)"}
if($Phase -eq 'Green'){$actual=Get-Content (Join-Path $PSScriptRoot 'native-observation.wprp') -Raw;$r=Test-C12ObservationProfile $actual $SchemaPath;if(-not $r.Valid){throw 'ASSERT_SHIPPED_PROFILE'};'PASS shipped-profile'}
"PASS profile cases=$($cases.Count) phase=$Phase"
