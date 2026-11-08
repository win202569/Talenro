using System;
using System.Globalization;
using System.Runtime.InteropServices;
using Microsoft.Win32.SafeHandles;

namespace Talenro.Diagnostics.NativeObservationV1
{
    [StructLayout(LayoutKind.Sequential)]
    public struct NativeFileTime { public uint Low; public uint High; }

    public sealed class TimeParts
    {
        public uint High { get; set; }
        public uint Low { get; set; }
        internal static TimeParts From(NativeFileTime value) => new TimeParts { High = value.High, Low = value.Low };
    }
    public sealed class ProcessTimes
    {
        public TimeParts Creation { get; set; }
        public TimeParts Exit { get; set; }
        public TimeParts Kernel { get; set; }
        public TimeParts User { get; set; }
    }
    public sealed class ProbeCall
    {
        public string StartedUtc { get; set; }
        public string EndedUtc { get; set; }
        public bool NotRun { get; set; } = true;
        public string ExceptionType { get; set; }
        public bool? ApiSucceeded { get; set; }
        public uint? RawResult { get; set; }
        public uint? ExitCode { get; set; }
        public ProcessTimes Times { get; set; }
        public int? LastError { get; set; }
    }
    public sealed class ProbeSnapshot
    {
        public string StartedUtc { get; set; }
        public string EndedUtc { get; set; }
        public string SnapshotError { get; set; }
        public ProbeCall Wait { get; set; } = new ProbeCall();
        public ProbeCall ExitCode { get; set; } = new ProbeCall();
        public ProbeCall Times { get; set; } = new ProbeCall();
    }

    // No static initializer: compilation/reflection is not native verification.
    public static class NativeProbe
    {
        [DllImport("kernel32.dll", ExactSpelling = true, SetLastError = true)]
        private static extern uint WaitForSingleObject(IntPtr handle, uint milliseconds);
        [DllImport("kernel32.dll", ExactSpelling = true, SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool GetExitCodeProcess(IntPtr handle, out uint exitCode);
        [DllImport("kernel32.dll", ExactSpelling = true, SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool GetProcessTimes(IntPtr handle, out NativeFileTime creation,
            out NativeFileTime exit, out NativeFileTime kernel, out NativeFileTime user);

        private static string Now() => DateTime.UtcNow.ToString("o", CultureInfo.InvariantCulture);

        public static ProbeSnapshot Capture(SafeProcessHandle handle)
        {
            var result = new ProbeSnapshot { StartedUtc = Now() };
            bool held = false;
            try
            {
                if (handle == null || handle.IsClosed || handle.IsInvalid)
                    throw new ArgumentException("A retained valid process handle is required.");
                handle.DangerousAddRef(ref held);
                if (handle.IsClosed || handle.IsInvalid)
                    throw new ArgumentException("Handle unavailable.");
                IntPtr pointer = handle.DangerousGetHandle();
                result.Wait = ReadWait(pointer);
                result.ExitCode = ReadExitCode(pointer);
                result.Times = ReadTimes(pointer);
            }
            catch (Exception error) { result.SnapshotError = error.GetType().FullName; }
            finally
            {
                if (held) handle.DangerousRelease();
                result.EndedUtc = Now();
            }
            return result;
        }

        private static ProbeCall ReadWait(IntPtr pointer)
        {
            var call = new ProbeCall { StartedUtc = Now(), NotRun = false };
            try
            {
                uint raw = WaitForSingleObject(pointer, 0);
                int? error = raw == uint.MaxValue ? Marshal.GetLastPInvokeError() : (int?)null;
                call.RawResult = raw;
                call.LastError = error;
            }
            catch (Exception error) { call.ExceptionType = error.GetType().FullName; }
            finally { call.EndedUtc = Now(); }
            return call;
        }

        private static ProbeCall ReadExitCode(IntPtr pointer)
        {
            var call = new ProbeCall { StartedUtc = Now(), NotRun = false };
            try
            {
                bool success = GetExitCodeProcess(pointer, out uint code);
                int? error = !success ? Marshal.GetLastPInvokeError() : (int?)null;
                call.ApiSucceeded = success;
                call.ExitCode = success ? code : (uint?)null;
                call.LastError = error;
            }
            catch (Exception error) { call.ExceptionType = error.GetType().FullName; }
            finally { call.EndedUtc = Now(); }
            return call;
        }

        private static ProbeCall ReadTimes(IntPtr pointer)
        {
            var call = new ProbeCall { StartedUtc = Now(), NotRun = false };
            try
            {
                bool success = GetProcessTimes(pointer, out NativeFileTime creation,
                    out NativeFileTime exit, out NativeFileTime kernel, out NativeFileTime user);
                int? error = !success ? Marshal.GetLastPInvokeError() : (int?)null;
                call.ApiSucceeded = success;
                call.LastError = error;
                if (success) call.Times = new ProcessTimes {
                    Creation = TimeParts.From(creation), Exit = TimeParts.From(exit),
                    Kernel = TimeParts.From(kernel), User = TimeParts.From(user)
                };
            }
            catch (Exception error) { call.ExceptionType = error.GetType().FullName; }
            finally { call.EndedUtc = Now(); }
            return call;
        }
    }
}
