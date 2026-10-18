# Test-only public SDK setter trampoline. No private fields or product operations.
function S-SetStored($Field,$Value){
    $valueArguments=[object[]]@($Value)
    $invokeArguments=[object[]]@($Field,$valueArguments)
    $null=$script:invokeMethod.Invoke($script:storageSetter,$invokeArguments)
}
