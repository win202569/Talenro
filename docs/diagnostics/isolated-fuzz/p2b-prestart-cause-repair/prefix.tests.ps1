# These literal prefixes are independent of the production validator.
function Prefix-Eval($Bytes,[string]$Channel,[bool]$AllowPrefix){
    return (& $script:LSModule {param($bytes,$channel,$prefix) Test-LSStreamContract $bytes $channel $prefix} $Bytes $Channel $AllowPrefix)
}
function Invoke-LSPrefixTests {
    $literal=[ordered]@{
        Out=@('','P','P2','P2B','P2B-','P2B-C','P2B-CO','P2B-CON','P2B-CONT','P2B-CONTA','P2B-CONTAI','P2B-CONTAIN','P2B-CONTAINM','P2B-CONTAINME','P2B-CONTAINMEN','P2B-CONTAINMENT','P2B-CONTAINMENT-','P2B-CONTAINMENT-O','P2B-CONTAINMENT-OU','P2B-CONTAINMENT-OUT')
        Err=@('','P','P2','P2B','P2B-','P2B-C','P2B-CO','P2B-CON','P2B-CONT','P2B-CONTA','P2B-CONTAI','P2B-CONTAIN','P2B-CONTAINM','P2B-CONTAINME','P2B-CONTAINMEN','P2B-CONTAINMENT','P2B-CONTAINMENT-','P2B-CONTAINMENT-E','P2B-CONTAINMENT-ER','P2B-CONTAINMENT-ERR')
    }
    $accepted=[ordered]@{EmptyPrefix=0;ProperPrefix=0;CompletePrefix=0;InvalidPrefix=0;NormalExact=0;NormalShortRejected=0;NullAndTypeRejected=0}
    foreach($channel in @('Out','Err')){
        foreach($index in 0..19){
            [byte[]]$bytes=[Text.Encoding]::ASCII.GetBytes($literal[$channel][$index])
            $group=if($index -eq 0){'EmptyPrefix'}elseif($index -eq 19){'CompletePrefix'}else{'ProperPrefix'}
            if(Prefix-Eval $bytes $channel $true){$accepted[$group]++}
            if(Prefix-Eval $bytes $channel $false){if($index -eq 19){$accepted.NormalExact++}else{$accepted.NormalShortRejected++}}
        }
        $full=if($channel -ceq 'Out'){'P2B-CONTAINMENT-OUT'}else{'P2B-CONTAINMENT-ERR'}
        foreach($bad in @(('X'+$full.Substring(1)),($full+'X'),'P2B-X','CONTAINMENT','garbage')){
            if(Prefix-Eval ([Text.Encoding]::ASCII.GetBytes($bad)) $channel $true){$accepted.InvalidPrefix++}
        }
        foreach($mode in @($false,$true)){
            if(Prefix-Eval $null $channel $mode){$accepted.NullAndTypeRejected++}
            if(Prefix-Eval 'P2B-CONTAINMENT-OUT' $channel $mode){$accepted.NullAndTypeRejected++}
        }
    }
    # Literal expected accepted counts. Do not derive them from a validator.
    $checks=[ordered]@{EmptyPrefix=2;ProperPrefix=36;CompletePrefix=2;InvalidPrefix=10;NormalExact=2;NormalShortRejected=38;NullAndTypeRejected=8}
    $want=[ordered]@{EmptyPrefix=2;ProperPrefix=36;CompletePrefix=2;InvalidPrefix=0;NormalExact=2;NormalShortRejected=0;NullAndTypeRejected=0}
    foreach($name in $checks.Keys){
        LS-Result $name ($accepted[$name] -eq $want[$name]) ([ordered]@{Checks=$checks[$name];Accepted=$accepted[$name];Rejected=($checks[$name]-$accepted[$name])})
    }
}
