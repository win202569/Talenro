# Separate supplement; the original frozen 42-group source stays unchanged.
# Catches culture-sensitive BOM rejection of valid data and genuine BOM leaks.
function Invoke-ODBOMRegression([byte[]]$OriginalBytes,[Collections.IDictionary]$Context,$ExpectedMeasurements){
    $priorCulture=[Globalization.CultureInfo]::CurrentCulture
    $groups=[Collections.Generic.List[object]]::new()
    try{
        [Globalization.CultureInfo]::CurrentCulture=[Globalization.CultureInfo]::GetCultureInfo('zh-CN')
        $groupStatus='PASS';$groupDetail=''
        try{
            if($OriginalBytes.Length -lt 3 -or $OriginalBytes[0] -ne 123){throw 'REGRESSION: fixed BOM-free fixture must start with literal ASCII {'}
            $report=Assert-ODSavedReferenceData -Bytes $OriginalBytes -Context $Context
            Assert-ODTestReport $report $ExpectedMeasurements
        }catch{$groupStatus='FAIL';$groupDetail=$_.Exception.GetType().FullName+': '+$_.Exception.Message+' | '+$_.ScriptStackTrace}
        $groups.Add([ordered]@{Name='BOMFreeASCIIUnderZhCN';Status=$groupStatus;Detail=$groupDetail})
        $groupStatus='PASS';$groupDetail=''
        try{
            [byte[]]$withBOM=[byte[]](@(239,187,191)+@($OriginalBytes))
            $didReject=$false
            try{$unused=Assert-ODSavedReferenceData -Bytes $withBOM -Context $Context}
            catch{
                if($_.Exception -isnot [Management.Automation.RuntimeException] -or $_.Exception.Message -cne 'ODR: saved data contract mismatch' -or $_.FullyQualifiedErrorId -cne 'ODR: saved data contract mismatch'){throw}
                $didReject=$true
            }
            if(-not $didReject){throw 'REGRESSION: actual UTF8 BOM accepted'}
        }catch{$groupStatus='FAIL';$groupDetail=$_.Exception.GetType().FullName+': '+$_.Exception.Message+' | '+$_.ScriptStackTrace}
        $groups.Add([ordered]@{Name='GenuineUTF8BOMRejectedUnderZhCN';Status=$groupStatus;Detail=$groupDetail})
    }finally{[Globalization.CultureInfo]::CurrentCulture=$priorCulture}
    $failures=@($groups|Where-Object{$_.Status -ceq 'FAIL'}).Count
    [ordered]@{Schema=1;PASS=2-$failures;FAIL=$failures;CultureRestored=([object]::Equals([Globalization.CultureInfo]::CurrentCulture,$priorCulture));Groups=$groups.ToArray()}
}
