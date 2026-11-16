using System.Globalization;
namespace Talenro.Observation.CloseGuardian;
public static class GuardianActivation
{
    private const string Prefix="TALENRO_GUARDIAN_V1_";
    private static readonly string[] Names=["VERSION","RUN_ID","PIPE","KEY","GUARDIAN_PID","GUARDIAN_CREATION","POLICY","T0","FREQUENCY"];
    public static ActivationResult Parse(IReadOnlyDictionary<string, string?> environment)
    {
        ArgumentNullException.ThrowIfNull(environment);
        var relevant=environment.Where(x=>x.Key.StartsWith("TALENRO_GUARDIAN_V1",StringComparison.OrdinalIgnoreCase)).ToArray();
        if(relevant.Length==0) return new(ActivationStatus.Disabled,null,null);
        ActivationResult Reject()=>new(ActivationStatus.Rejected,null,"InvalidBootstrap");
        if(relevant.Length!=Names.Length || relevant.Any(x=>!Names.Any(n=>x.Key==Prefix+n))) return Reject();
        string? Get(string name)=>environment.TryGetValue(Prefix+name,out var value)?value:null;
        bool Number(string name,out long value)=>long.TryParse(Get(name),NumberStyles.None,CultureInfo.InvariantCulture,out value);
        var run=Get("RUN_ID"); var pipe=Get("PIPE");
        if(Get("VERSION")!="1" || !GuardianValidation.Hex(run,32) || !Guid.TryParseExact(run,"N",out var runId) || runId==Guid.Empty ||
            pipe is null || !pipe.StartsWith("talenro.guardian.",StringComparison.Ordinal) || !GuardianValidation.Hex(pipe[17..],32) ||
            !GuardianValidation.Hex(Get("KEY"),64) || !GuardianValidation.Hex(Get("POLICY"),64) ||
            !Number("GUARDIAN_PID",out var pid) || pid<=0 || pid>int.MaxValue ||
            !Number("GUARDIAN_CREATION",out var creation) || creation<=0 || creation>2650467743999999999L ||
            !Number("T0",out var t0) || !Number("FREQUENCY",out var frequency) || frequency<=0) return Reject();
        return new(ActivationStatus.Valid,new(new(runId,Get("POLICY")!,t0,frequency),pipe,(int)pid,creation,Convert.FromHexString(Get("KEY")!)),null);
    }
}
