# Definition-only comparison of the complete key set and hashes.
function Assert-C12ProfileManifest([hashtable]$Expected,[hashtable]$Actual) {
 if($Expected.Count -ne $Actual.Count){throw 'SOURCE_SET_CHANGED'}
 foreach($k in $Expected.Keys){if(-not $Actual.ContainsKey($k) -or $Actual[$k] -cne $Expected[$k]){throw 'SOURCE_CHANGED'}}
}
