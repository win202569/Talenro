package main

import (
    "archive/zip"
    "errors"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "testing/synctest"
)

const emptyH = "h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU="
const abcH = "h1:xYudg4RCFWzdnRPnaadGTWOBYxaNjPFPGOgJZ3DiUvY="
func fixture(t *testing.T, names []string, content string, expected string) Candidate {
    t.Helper()
    root:=t.TempDir(); c:=Candidate{Module:"fixture.example/mod",Version:"v1.0.0",Expected:expected}
    c.Dir=filepath.Join(root,"fixture.example","mod@v1.0.0")
    c.Zip=filepath.Join(root,"cache","download","fixture.example","mod","@v","v1.0.0.zip")
    if err:=os.MkdirAll(c.Dir,0700);err!=nil{t.Fatal(err)}
    if err:=os.MkdirAll(filepath.Dir(c.Zip),0700);err!=nil{t.Fatal(err)}
    f,err:=os.Create(c.Zip);if err!=nil{t.Fatal(err)};z:=zip.NewWriter(f)
    for _,name:=range names {w,err:=z.Create(name);if err!=nil{t.Fatal(err)};if _,err=w.Write([]byte(content));err!=nil{t.Fatal(err)}}
    if err=z.Close();err!=nil{t.Fatal(err)};if err=f.Close();err!=nil{t.Fatal(err)}
    if len(names)>0 {if err=os.WriteFile(filepath.Join(c.Dir,"a.txt"),[]byte(content),0600);err!=nil{t.Fatal(err)}}
    return c
}
func collect(c Candidate)([]Event,error){var events []Event;err:=Run([]Candidate{c},strings.Repeat("a",32),func(e Event)error{events=append(events,e);return nil});return events,err}
func good(t *testing.T)Candidate{return fixture(t,[]string{"fixture.example/mod@v1.0.0/a.txt"},"abc",abcH)}
func requireErr(t *testing.T,err error,code string){t.Helper();if err==nil||!strings.Contains(err.Error(),code){t.Fatalf("want %s rejection; got %v",code,err)}}
func TestEmptyFixture(t *testing.T){e,err:=collect(fixture(t,nil,"",emptyH));if err!=nil||len(e)!=6{t.Fatalf("empty: %v events=%d",err,len(e))};if e[3].Digest!=emptyH||e[5].Digest!=emptyH{t.Fatal("empty digest")}}
func TestOneFileEquivalence(t *testing.T){e,err:=collect(good(t));if err!=nil||len(e)!=6{t.Fatalf("abc: %v events=%d",err,len(e))};if e[3].Digest!=abcH||e[5].Digest!=abcH{t.Fatal("literal abc digest")}}
func TestChangedDirectoryContent(t *testing.T){c:=good(t);if err:=os.WriteFile(filepath.Join(c.Dir,"a.txt"),[]byte("changed"),0600);err!=nil{t.Fatal(err)};e,err:=collect(c);requireErr(t,err,"MISMATCH");if len(e)!=6||e[3].Status!="OK"||e[5].Status!="MISMATCH"{t.Fatal("phase attribution")}}
func TestExpectedDigestBinding(t *testing.T){line:="fixture.example/mod v1.0.0 "+abcH+"\n";h,err:=ReadExpected(line,abcH+"\n","fixture.example/mod","v1.0.0");if err!=nil||h!=abcH{t.Fatalf("expected parser: %v",err)};for _,s:=range []string{"",line+"fixture.example/mod v1.0.0 "+emptyH+"\n"}{_,err=ReadExpected(s,abcH,"fixture.example/mod","v1.0.0");requireErr(t,err,"EXPECTED")}}
func TestReadFailure(t *testing.T){c:=good(t);if err:=os.Remove(c.Zip);err!=nil{t.Fatal(err)};_,err:=collect(c);requireErr(t,err,"INPUT")}
func TestCandidateIdentity(t *testing.T){c:=good(t);c.Module="../escape";requireErr(t,ValidateCandidate(c),"IDENTITY")}
func TestZipTraversal(t *testing.T){c:=fixture(t,[]string{"fixture.example/mod@v1.0.0/../escape"},"abc",abcH);requireErr(t,ValidateCandidate(c),"ZIP_NAME")}
func TestZipDuplicate(t *testing.T){c:=fixture(t,[]string{"fixture.example/mod@v1.0.0/a.txt","fixture.example/mod@v1.0.0/a.txt"},"abc",abcH);requireErr(t,ValidateCandidate(c),"ZIP_DUPLICATE")}
func TestZipPrefix(t *testing.T){c:=fixture(t,[]string{"other.example/mod@v1.0.0/a.txt"},"abc",abcH);requireErr(t,ValidateCandidate(c),"ZIP_PREFIX")}
func TestFilesystemContainment(t *testing.T){c:=good(t);c.Zip=filepath.Join(t.TempDir(),"v1.0.0.zip");requireErr(t,ValidateCandidate(c),"CONTAINMENT")}
func TestReparseAttributes(t *testing.T){requireErr(t,rejectAttributes(0x400),"REPARSE");if err:=rejectAttributes(0x20);err!=nil{t.Fatalf("regular rejected: %v",err)}}
func TestInputsUnchanged(t *testing.T){c:=good(t);before,err:=os.ReadFile(c.Zip);if err!=nil{t.Fatal(err)};if _,err=collect(c);err!=nil{t.Fatal(err)};after,err:=os.ReadFile(c.Zip);if err!=nil||string(before)!=string(after){t.Fatal("ZIP mutated")};data,err:=os.ReadFile(filepath.Join(c.Dir,"a.txt"));if err!=nil||string(data)!="abc"{t.Fatal("directory mutated")}}
func TestEventOrder(t *testing.T){e,err:=collect(good(t));if err!=nil||len(e)!=6{t.Fatalf("events: %v",err)};want:=[]string{"PREPARE/START","PREPARE/END","ZIP/START","ZIP/END","DIRECTORY/START","DIRECTORY/END"};for i,v:=range e{if v.Phase+"/"+v.Kind!=want[i]||v.Sequence!=i+1||v.Schema!=1{t.Fatalf("event %d",i)}}}
func TestEmitterFailure(t *testing.T){calls:=0;err:=Run([]Candidate{good(t)},strings.Repeat("a",32),func(e Event)error{calls++;return errors.New("SINK_STOP")});requireErr(t,err,"SINK_STOP");if calls!=1{t.Fatalf("emit calls %d",calls)}}
func TestMismatchStopsLater(t *testing.T){c:=good(t);c.Expected=emptyH;var e []Event;err:=Run([]Candidate{c,good(t)},strings.Repeat("a",32),func(v Event)error{e=append(e,v);return nil});requireErr(t,err,"MISMATCH");if len(e)!=4||e[3].Status!="MISMATCH"{t.Fatalf("continued after mismatch: %d",len(e))}}
func TestMonotonicPhaseEvidence(t *testing.T){
    check:=func(t *testing.T){
        e,err:=collect(good(t));if err!=nil||len(e)!=6{t.Fatalf("run: %v events=%d",err,len(e))}
        for i,v:=range e{
            t.Logf("sequence=%d phase=%s kind=%s status=%s elapsed_ns=%d",v.Sequence,v.Phase,v.Kind,v.Status,v.ElapsedNs)
            if v.Sequence!=i+1||v.Phase!=[]string{"PREPARE","ZIP","DIRECTORY"}[i/2]{t.Fatal("phase order")}
            if i%2==0{
                if v.Kind!="START"||v.ElapsedNs!=0||v.Status!="STARTED"||v.Digest!=""{t.Fatal("start evidence")}
            }else{
                // Same-tick completion is valid; missing END or digest is not.
                if v.Kind!="END"||v.ElapsedNs<0||v.Status!="OK"||v.Expected!=abcH{t.Fatalf("invalid completion: %+v",v)}
                if v.Phase=="PREPARE"{if v.Digest!=""{t.Fatal("preparation digest")}}else if v.Digest!=abcH{t.Fatal("missing completed digest")}
                if strings.HasSuffix(t.Name(),"/same_tick")&&v.ElapsedNs!=0{t.Fatal("frozen-clock measurement was altered")}
            }
        }
    }
    t.Run("real_clock",check)
    t.Run("same_tick",func(t *testing.T){synctest.Test(t,check)})
}
