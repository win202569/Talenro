function Read-P2BProcessAssociation([object]$Process){
    $result=[pscustomobject][ordered]@{State='Unknown';ProcessId=$null;Error=$null}
    if($Process -isnot [Diagnostics.Process]){return $result}
    try{
        $property=[Diagnostics.Process].GetProperty('Id',([Reflection.BindingFlags]::Public -bor [Reflection.BindingFlags]::Instance))
        [object]$value=$property.GetValue($Process)
        if($value -is [int] -and $value -gt 0){$result.State='Associated';$result.ProcessId=$value}
    }catch{
        $result.Error=$_
        [Exception]$node=$_.Exception
        $seen=[Collections.Generic.List[Exception]]::new()
        while($null -ne $node){
            if($seen.Count -ge 8){return $result}
            foreach($prior in $seen){if([object]::ReferenceEquals($prior,$node)){return $result}}
            $seen.Add($node)
            if($null -eq $node.InnerException){
                if($node.GetType() -eq [InvalidOperationException]){$result.State='Unassociated'}
                return $result
            }
            if($node.GetType() -ne [Management.Automation.MethodInvocationException] -and $node.GetType() -ne [Reflection.TargetInvocationException]){return $result}
            $node=$node.InnerException
        }
    }
    return $result
}
