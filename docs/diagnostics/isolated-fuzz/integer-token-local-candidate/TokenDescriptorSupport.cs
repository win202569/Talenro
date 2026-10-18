// Test-only support. NOT COMPILED during source preparation.
using System.Management.Automation;
namespace TalenroTokenDirect {
    public static class DescriptorSupport {
        public static int Calls;
        public static object Read(PSObject owner) {
            Calls++;
            return "callback-must-not-run";
        }
    }
}
