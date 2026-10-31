using System;
using System.IO;
namespace ODRepairFixture {
    public sealed class FaultStream : Stream {
        private readonly FileStream healthy;
        private readonly FileStream closed;
        private readonly string mode;
        private readonly string missing;
        public int WriteCalls { get; private set; }
        public int FlushCalls { get; private set; }
        public int DisposeCalls { get; private set; }
        public FaultStream(FileStream healthyStream, string selectedMode, string absentPath, string closedPath) {
            healthy=healthyStream; mode=selectedMode; missing=absentPath;
            closed=File.Open(closedPath,FileMode.CreateNew,FileAccess.ReadWrite,FileShare.None);
            closed.Dispose();
        }
        public override bool CanRead { get { return false; } }
        public override bool CanSeek { get { return false; } }
        public override bool CanWrite { get { return true; } }
        public override long Length { get { return healthy.Length; } }
        public override long Position { get { return healthy.Position; } set { healthy.Position=value; } }
        public override void Write(byte[] buffer,int offset,int count) {
            WriteCalls++;
            if(mode=="WriteDispose") closed.Write(buffer,offset,count);
            else healthy.Write(buffer,offset,count);
        }
        public override void Flush() {
            FlushCalls++;
            if(mode=="FlushDispose") closed.Flush(); else healthy.Flush();
        }
        protected override void Dispose(bool disposing) {
            DisposeCalls++;
            if(disposing) {
                healthy.Dispose();
                if(mode!="Success") File.ReadAllText(missing);
            }
        }
        public override int Read(byte[] b,int o,int c) { return healthy.Read(b,o,c); }
        public override long Seek(long o,SeekOrigin s) { return healthy.Seek(o,s); }
        public override void SetLength(long n) { healthy.SetLength(n); }
    }
}
