# UNFINISHED DEFINITIONS ONLY. Do not freeze or invoke RED25 yet.
# No source imports, filesystem writes or test execution at module load.
function RT-Need([bool]$Condition,[string]$Label){
    if(-not $Condition){throw ('Runtime binding regression: '+$Label)}
}
function RT-Equal($Left,$Right){
    return ((ConvertTo-Json -InputObject $Left -Depth 100 -Compress) -ceq (ConvertTo-Json -InputObject $Right -Depth 100 -Compress))
}
function RT-Copy($Value){
    return ConvertFrom-Json -AsHashtable -InputObject (ConvertTo-Json -InputObject $Value -Depth 100)
}
function RT-Error($Record){
    if($null -eq $Record){return $null}
    return [ordered]@{Type=$Record.Exception.GetType().FullName;Message=$Record.Exception.Message;ErrorId=$Record.FullyQualifiedErrorId;Position=$Record.InvocationInfo.PositionMessage;Stack=$Record.ScriptStackTrace}
}
function RT-Tree([string]$Text){
    $tokens=$null;$errors=$null
    $tree=[Management.Automation.Language.Parser]::ParseInput($Text,[ref]$tokens,[ref]$errors)
    RT-Need ($errors.Count -eq 0) 'source parse'
    return $tree
}
function RT-Definition($Tree,[string]$Name){
    $nodes=@($Tree.FindAll({param($n)
        $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -ceq $Name
    },$true))
    RT-Need ($nodes.Count -eq 1) ('unique definition '+$Name)
    return $nodes[0]
}
function RT-Calls($Tree,[string]$Name){
    # Deliberately emit individual nodes, never a comma-wrapped array.
    foreach($node in $Tree.FindAll({param($n)
        $n -is [Management.Automation.Language.CommandAst] -and $n.GetCommandName() -ceq $Name
    },$true)){$node}
}
function RT-SingleCall($Tree,[string]$Name){
    $calls=@(RT-Calls $Tree $Name)
    RT-Need ($calls.Count -eq 1) ('unique call '+$Name)
    RT-Need ($calls[0] -is [Management.Automation.Language.CommandAst]) ('scalar command '+$Name)
    RT-Need ($calls[0].Extent.StartOffset -is [int]) ('integer call offset '+$Name)
    return $calls[0]
}
function RT-ReadOnlyDefinition($Tree,[string]$Name){
    $allowed=@('Get-ODAttemptSlot','Get-ODEpochManifestName','Driver-Hash',
        'Driver-CheckHash','Driver-Need','Driver-EqualJSON','Driver-Keys',
        'Driver-VerifyHistory','Driver-ValidateRuntimeInputs','Driver-RequireFinal',
        'Get-ODWriterManifestPath','Driver-FalseFlags','Assert-ODRuntimeLineage')
    RT-Need ($Name -cin $allowed) ('extraction allowlist '+$Name)
    $definition=RT-Definition $Tree $Name
    # Return only a scriptblock; the caller dot-sources it in its local scope.
    # Gate, receipt orchestration, acquisition and Prepare are never extractable.
    return [scriptblock]::Create(([string][char]10 * ($definition.Extent.StartLineNumber-1))+$definition.Extent.Text)
}
function RT-Reject([scriptblock]$Body,[string]$Message){
    $caught=$null;$emitted=@()
    try{$emitted=@(& $Body)}catch{$caught=$_}
    RT-Need ($null -ne $caught -and $emitted.Count -eq 0) 'silent validation rejection'
    RT-Need ($caught.Exception.GetType().FullName -ceq 'System.Management.Automation.RuntimeException' -and $caught.Exception.Message -ceq $Message -and $caught.FullyQualifiedErrorId -ceq $Message) ('precise rejection '+$Message)
}
function RT-Case([string]$Name,[scriptblock]$Body,$State){
    if($State.Stopped){$State.Groups.Add([ordered]@{Name=$Name;Status='NotRun'});return}
    try{
        $emitted=@(& $Body)
        RT-Need ($emitted.Count -eq 0) ('silent group '+$Name)
        $State.Groups.Add([ordered]@{Name=$Name;Status='PASS'})
    }catch{
        $State.RootError=$_;$State.Stopped=$true
        $State.Groups.Add([ordered]@{Name=$Name;Status='FAIL'})
    }
}
function Get-RTGroupNames {
    @('NewEpochBinding','Green24Identity','HistoricalEpoch','LiveCandidate',
      'ApprovalIdentity','PriorFailurePreserved','ExternalHistory','RelativeHistory',
      'ParentLineage','WriterIdentity','InvalidReceiptRejected','InvalidCandidateRejected',
      'EpochRejected','DisjointSlots','ScalarCallSelection','AttemptOwnership',
      'ActualGateCall','SourceWindow','OriginalGuardsUnchanged','OriginalSuiteUnchanged')
}
# The full Invoke-ODRuntimeBindingRegression entry point is intentionally absent
# until all20group bodies, fixture checks and pre/post recording are authored.
# This draft must never be used to consume an auxiliary attempt.
