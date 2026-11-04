using System;
using System.IO;
using System.Threading;
using System.Threading.Tasks;

namespace Talenro.Diagnostics.BoundedDiskV1
{
    public sealed class CapacityExceededException : IOException {
        public CapacityExceededException() : base("Bounded file sink capacity exceeded") {}
    }
    // This limits admitted bytes, not filesystem allocation or I/O latency.
    public sealed class BoundedFileSink : Stream {
        private readonly Stream sink;
        private int gate, faulted, closed, closeAttempted;
        private long reserved, confirmed;
        private Exception closeError;
        private const int Limit = 65536;
        internal BoundedFileSink(Stream ownedSink) {
            sink=ownedSink ?? throw new ArgumentNullException(nameof(ownedSink));
            if(!sink.CanWrite)throw new ArgumentException("Writable sink required");
        }
        public static BoundedFileSink OpenNew(string absolutePath) {
            if(absolutePath==null)throw new ArgumentNullException(nameof(absolutePath));
            if(!Path.IsPathFullyQualified(absolutePath))throw new ArgumentException("Absolute path required");
            return new BoundedFileSink(new FileStream(absolutePath,FileMode.CreateNew,FileAccess.Write,FileShare.Read));
        }
        public long ReservedBytes => Interlocked.Read(ref reserved);
        public long ConfirmedBytes => Interlocked.Read(ref confirmed);
        public bool Faulted => Volatile.Read(ref faulted)!=0;
        public bool Closed => Volatile.Read(ref closed)!=0;
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override bool CanWrite => !Closed && !Faulted && Volatile.Read(ref closeAttempted)==0;
        private long Inspect() {
            Enter();
            try {return confirmed;} finally {Leave();}
        }
        public override long Length => Inspect();
        public override long Position {get=>Inspect();set=>throw new NotSupportedException();}
        private void Enter() {
            if(Interlocked.CompareExchange(ref gate,1,0)!=0)throw new InvalidOperationException("Operation already active");
        }
        private void Leave() {Volatile.Write(ref gate,0);}
        private void Fault() {Volatile.Write(ref faulted,1);}
        private static void Arguments(byte[] b,int offset,int count) {
            if(b==null)throw new ArgumentNullException(nameof(b));
            if(offset<0)throw new ArgumentOutOfRangeException(nameof(offset));
            if(count<0)throw new ArgumentOutOfRangeException(nameof(count));
            if(offset>b.Length || count>b.Length-offset)throw new ArgumentException("Array range");
        }
        private void EnterWrite(int count,CancellationToken token) {
            Enter();
            try {
                if(Closed)throw new ObjectDisposedException(nameof(BoundedFileSink));
                if(Faulted || closeAttempted!=0)throw new IOException("Sink latched");
                if(token.IsCancellationRequested){Fault();token.ThrowIfCancellationRequested();}
                if(count>Limit-reserved){Fault();throw new CapacityExceededException();}
                // Keep this reservation even if the underlying write fails partially.
                reserved+=count;
            }catch {Leave();throw;}
        }
        public override void Write(byte[] buffer,int offset,int count) {
            Arguments(buffer,offset,count);EnterWrite(count,CancellationToken.None);
            try {sink.Write(buffer,offset,count);Interlocked.Add(ref confirmed,count);}
            catch {Fault();throw;} finally {Leave();}
        }
        public override void Write(ReadOnlySpan<byte> buffer) {
            EnterWrite(buffer.Length,CancellationToken.None);
            try {sink.Write(buffer);Interlocked.Add(ref confirmed,buffer.Length);}
            catch {Fault();throw;} finally {Leave();}
        }
        public override void WriteByte(byte value) {
            EnterWrite(1,CancellationToken.None);
            try {sink.WriteByte(value);Interlocked.Increment(ref confirmed);}
            catch {Fault();throw;} finally {Leave();}
        }
        public override async Task WriteAsync(byte[] buffer,int offset,int count,CancellationToken token) {
            Arguments(buffer,offset,count);EnterWrite(count,token);
            try {await sink.WriteAsync(buffer,offset,count,token).ConfigureAwait(false);Interlocked.Add(ref confirmed,count);}
            catch {Fault();throw;} finally {Leave();}
        }
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> buffer,CancellationToken token=default) {
            EnterWrite(buffer.Length,token);
            try {await sink.WriteAsync(buffer,token).ConfigureAwait(false);Interlocked.Add(ref confirmed,buffer.Length);}
            catch {Fault();throw;} finally {Leave();}
        }
        private void CheckFlush() {
            if(Closed)throw new ObjectDisposedException(nameof(BoundedFileSink));
            if(Faulted || closeAttempted!=0)throw new IOException("Sink latched");
        }
        public override void Flush() {
            Enter();
            try {CheckFlush();sink.Flush();}
            catch {Fault();throw;} finally {Leave();}
        }
        public override async Task FlushAsync(CancellationToken token) {
            Enter();
            try {CheckFlush();await sink.FlushAsync(token).ConfigureAwait(false);}
            catch {Fault();throw;} finally {Leave();}
        }
        private static Exception Combine(Exception first,Exception next) {
            return first==null?next:new AggregateException(first,next);
        }
        public async Task CloseAsync() {
            if(Closed)return;
            Enter();
            try {
                if(Closed)return;
                if(closeAttempted!=0)throw new IOException("Finalization already failed",closeError);
                closeAttempted=1;
                Exception error=null;
                try {await sink.FlushAsync(CancellationToken.None).ConfigureAwait(false);}catch(Exception ex){error=ex;}
                try {await sink.DisposeAsync().ConfigureAwait(false);}catch(Exception ex){error=Combine(error,ex);}
                if(error!=null){closeError=error;Fault();throw error;}
                Volatile.Write(ref closed,1);
            }finally {Leave();}
        }
        public override ValueTask DisposeAsync() => new ValueTask(CloseAsync());
        protected override void Dispose(bool disposing) {
            if(!disposing || Closed)return;
            Enter();
            try {
                if(Closed)return;
                if(closeAttempted!=0)throw new IOException("Finalization already failed",closeError);
                closeAttempted=1;
                Exception error=null;
                try {sink.Flush();}catch(Exception ex){error=ex;}
                try {sink.Dispose();}catch(Exception ex){error=Combine(error,ex);}
                if(error!=null){closeError=error;Fault();throw error;}
                Volatile.Write(ref closed,1);
            }finally {Leave();}
        }
        public override int Read(byte[] b,int o,int n)=>throw new NotSupportedException();
        public override int Read(Span<byte> b)=>throw new NotSupportedException();
        public override int ReadByte()=>throw new NotSupportedException();
        public override Task<int> ReadAsync(byte[] b,int o,int n,CancellationToken t)=>throw new NotSupportedException();
        public override ValueTask<int> ReadAsync(Memory<byte> b,CancellationToken t=default)=>throw new NotSupportedException();
        public override long Seek(long offset,SeekOrigin origin)=>throw new NotSupportedException();
        public override void SetLength(long value)=>throw new NotSupportedException();
    }
}

