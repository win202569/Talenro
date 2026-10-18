# Test-fixture helper only; no product imports or load-time actions.
# Preserve explicit malformed/decorated arguments; own plain numeric fixture storage.
function New-TBRequest([object]$OriginTick=[long]0,[object]$Frequency=[long]1000,[object]$ObservedTick=[long]0,[object]$StopTick=$null,[object]$Schema=[int]1){
    $typeMethod=[object].GetMethod('GetType',[type[]]@())
    $membersProperty=[Management.Automation.PSMemberSet].GetProperty('Members')
    $provided=@{Schema=$Schema;OriginTick=$OriginTick;Frequency=$Frequency;ObservedTick=$ObservedTick;StopTick=$StopTick}
    $request=[pscustomobject][ordered]@{}
    foreach($name in @('Schema','OriginTick','Frequency','ObservedTick','StopTick')){
        $value=$provided[$name];$scalar=$value
        if($null -ne $value){
            $base=$value.PSObject.BaseObject
            $type=$typeMethod.Invoke($base,$null)
            if($type -eq [int] -or $type -eq [long]){
                $copy=$true
                if($PSBoundParameters.ContainsKey($name)){
                    # Caller decorations are intentional invalid fixtures, not defaults
                    # to sanitize. Read descriptors with fixed SDK metadata only.
                    $view=$value.psextended
                    foreach($member in $membersProperty.GetValue($view.PSObject.BaseObject,$null)){$copy=$false;break}
                }
                if($copy){
                    $storage=[Array]::CreateInstance($type,1)
                    $storage.SetValue($base,0)
                    $scalar=$storage.GetValue(0)
                }
            }
        }
        $request.PSObject.Properties.Add([Management.Automation.PSNoteProperty]::new($name,$scalar))
    }
    return $request
}
