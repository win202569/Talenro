# Standalone expression diagnostic only. No product/test import or API call.
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$principal=[Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
$admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if($admin -or $PSVersionTable.PSVersion.ToString() -cne '7.6.5' -or -not [Environment]::Is64BitProcess){throw 'Ordinary pinned PS7 host required'}
[Console]::Out.WriteLine('HOST: '+$PSVersionTable.PSVersion.ToString()+' admin='+$admin)
$len=16
$original='$len-1,$len+1,113'
$tokens=$null;$errors=$null
$ast=[Management.Automation.Language.Parser]::ParseInput($original,[ref]$tokens,[ref]$errors)
foreach($node in $ast.FindAll({param($n) $n -is [Management.Automation.Language.BinaryExpressionAst]},$true)){
    [Console]::Out.WriteLine('AST: '+$node.Operator+' left=['+$node.Left.Extent.Text+'] right=['+$node.Right.Extent.Text+']')
}
$entered=0
try{
    foreach($n in @($len-1,$len+1,113)){$entered++}
    [Console]::Out.WriteLine('ORIGINAL: no exception')
}catch{
    [Console]::Out.WriteLine('ORIGINAL: '+$_.FullyQualifiedErrorId)
    [Console]::Out.WriteLine('MESSAGE: '+$_.Exception.Message)
    [Console]::Out.WriteLine('LOCATION: '+$_.InvocationInfo.ScriptLineNumber)
}
[Console]::Out.WriteLine('BODY_ENTRIES: '+$entered)
$control=@(($len-1),($len+1),113)
[Console]::Out.WriteLine('PARENTHESIZED: '+($control -join ','))
[Console]::Out.WriteLine('LIMIT: expression only; no product import or acceptance')
