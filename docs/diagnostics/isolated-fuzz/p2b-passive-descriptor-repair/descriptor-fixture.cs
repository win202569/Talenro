// Test-only managed counter fixture. No external operations or callbacks.
using System.Management.Automation;
namespace P2BDescriptorFixture {
    public sealed class CounterNote : PSNoteProperty {
        private object stored;
        public static int MemberReads;
        public static int ValueReads;
        public static int Copies;
        public static int Writes;
        public static int Total { get { return MemberReads + ValueReads + Copies + Writes; } }
        public static void Reset() { MemberReads = 0; ValueReads = 0; Copies = 0; Writes = 0; }
        public CounterNote(string name, object value) : base(name, value) { stored = value; }
        public override PSMemberTypes MemberType { get { MemberReads++; return PSMemberTypes.NoteProperty; } }
        public override object Value { get { ValueReads++; return stored; } set { Writes++; stored = value; } }
        public override PSMemberInfo Copy() { Copies++; return new CounterNote(Name, stored); }
    }
}
