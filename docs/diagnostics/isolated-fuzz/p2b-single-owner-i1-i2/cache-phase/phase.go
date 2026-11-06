package main

import (
    "archive/zip"
    "encoding/base64"
    "errors"
    "io"
    "io/fs"
    "os"
    "path"
    "path/filepath"
    "regexp"
    "strings"
    "syscall"
    "time"

    "talenro.local/c12-cache-phase/internal/dirhash"
)

type Candidate struct { Module, Version, Zip, Dir, Expected, SumFile string }
type Event struct {
    Schema, Sequence int
    Session, Module, Version, Phase, Kind, Status, Digest, Expected string
    ElapsedNs int64
}
func rejectAttributes(attributes uint32) error {
    if attributes&0x400 != 0 { return errors.New("REPARSE") }
    return nil
}

// Check ancestors before querying descendants. This is not an atomic snapshot:
// concurrent cache mutation remains outside this diagnostic's guarantees.
func safePath(p string) error {
    if !filepath.IsAbs(p) { return errors.New("CONTAINMENT") }
    parent := filepath.Dir(p)
    if parent != p { if err := safePath(parent); err != nil { return err } }
    ptr, err := syscall.UTF16PtrFromString(p)
    if err != nil { return errors.New("INPUT") }
    attributes, err := syscall.GetFileAttributes(ptr)
    if err != nil { return errors.New("INPUT") }
    return rejectAttributes(attributes)
}

func ValidateCandidate(c Candidate) error {
    if !regexp.MustCompile(`^[a-z0-9.-]+(/[a-z0-9._-]+)+$`).MatchString(c.Module) ||
       !regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(c.Version) {
        return errors.New("IDENTITY")
    }
    for _, part := range strings.Split(c.Module,"/") { if part=="." || part==".." { return errors.New("IDENTITY") } }
    if !filepath.IsAbs(c.Dir) || !filepath.IsAbs(c.Zip) { return errors.New("CONTAINMENT") }
    root := filepath.Clean(c.Dir)
    for range strings.Split(c.Module,"/") { root=filepath.Dir(root) }
    wantDir:=filepath.Join(root,filepath.FromSlash(c.Module+"@"+c.Version))
    wantZip:=filepath.Join(root,"cache","download",filepath.FromSlash(c.Module),"@v",c.Version+".zip")
    if !strings.EqualFold(filepath.Clean(c.Dir),wantDir) || !strings.EqualFold(filepath.Clean(c.Zip),wantZip) { return errors.New("CONTAINMENT") }
    if err:=safePath(c.Dir);err!=nil{return err}
    if err:=safePath(c.Zip);err!=nil{return err}
    info,err:=os.Lstat(c.Zip);if err!=nil||!info.Mode().IsRegular(){return errors.New("INPUT")}
    info,err=os.Lstat(c.Dir);if err!=nil||!info.IsDir(){return errors.New("INPUT")}
    if err=filepath.WalkDir(c.Dir,func(p string,d fs.DirEntry,walkErr error)error{
        if walkErr!=nil{return errors.New("INPUT")}
        if err:=safePath(p);err!=nil{return err}
        info,err:=d.Info();if err!=nil{return errors.New("INPUT")}
        if !info.IsDir()&&!info.Mode().IsRegular(){return errors.New("INPUT")}
        return nil
    });err!=nil{return err}
    z,err:=zip.OpenReader(c.Zip);if err!=nil{return errors.New("INPUT")};defer z.Close()
    seen:=map[string]bool{};prefix:=c.Module+"@"+c.Version+"/"
    for _,f:=range z.File{
        n:=f.Name
        if strings.ContainsAny(n,"\\:\r\n\x00")||strings.HasPrefix(n,"/")||path.Clean(n)!=strings.TrimSuffix(n,"/") {return errors.New("ZIP_NAME")}
        for _,part:=range strings.Split(n,"/"){if part==".."||part=="."{return errors.New("ZIP_NAME")}}
        if !strings.HasPrefix(n,prefix){return errors.New("ZIP_PREFIX")}
        if seen[n]{return errors.New("ZIP_DUPLICATE")};seen[n]=true
        if !f.Mode().IsRegular()&&!f.Mode().IsDir(){return errors.New("ZIP_NAME")}
    }
    return nil
}

func validDigest(h string) bool {
    if !strings.HasPrefix(h,"h1:"){return false}
    b,err:=base64.StdEncoding.DecodeString(h[3:])
    return err==nil&&len(b)==32&&base64.StdEncoding.EncodeToString(b)==h[3:]
}
func ReadExpected(sumText,ziphashText,module,version string)(string,error){
    expected:=strings.TrimSpace(ziphashText)
    if !validDigest(expected){return "",errors.New("EXPECTED")}
    found:=false
    for _,line:=range strings.Split(sumText,"\n"){
        fields:=strings.Fields(line)
        if len(fields)>=2&&fields[0]==module&&fields[1]==version{
            if len(fields)!=3||fields[2]!=expected{return "",errors.New("EXPECTED")};found=true
        }
    }
    if !found{return "",errors.New("EXPECTED")};return expected,nil
}
func boundedText(p string,limit int64)(string,error){
    if err:=safePath(p);err!=nil{return "",err}
    info,err:=os.Lstat(p);if err!=nil||!info.Mode().IsRegular()||info.Size()>limit{return "",errors.New("EXPECTED")}
    f,err:=os.Open(p);if err!=nil{return "",errors.New("INPUT")};defer f.Close()
    b,err:=io.ReadAll(io.LimitReader(f,limit+1));if err!=nil||int64(len(b))>limit{return "",errors.New("EXPECTED")}
    return string(b),nil
}
func Run(candidates []Candidate,session string,emit func(Event)error)error{
    if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(session)||emit==nil{return errors.New("SESSION")}
    sequence:=0
    for _,c:=range candidates{
        for _,phase:=range []string{"PREPARE","ZIP","DIRECTORY"}{
            sequence++
            event:=Event{Schema:1,Sequence:sequence,Session:session,Module:c.Module,Version:c.Version,Phase:phase,Kind:"START",Status:"STARTED",Expected:c.Expected}
            if err:=emit(event);err!=nil{return err}
            start:=time.Now();digest:="";var err error
            switch phase{
            case "PREPARE":
                err=ValidateCandidate(c)
                if err==nil&&c.Expected==""{
                    var sum,zh string
                    sum,err=boundedText(c.SumFile,4<<20)
                    if err==nil{zh,err=boundedText(strings.TrimSuffix(c.Zip,".zip")+".ziphash",1024)}
                    if err==nil{c.Expected,err=ReadExpected(sum,zh,c.Module,c.Version)}
                }
                if err==nil&&!validDigest(c.Expected){err=errors.New("EXPECTED")}
            case "ZIP": digest,err=dirhash.HashZip(c.Zip,dirhash.DefaultHash)
            case "DIRECTORY": digest,err=dirhash.HashDir(c.Dir,c.Module+"@"+c.Version,dirhash.DefaultHash)
            }
            sequence++;event.Sequence=sequence;event.Kind="END";event.ElapsedNs=time.Since(start).Nanoseconds();event.Expected=c.Expected;event.Digest=digest;event.Status="OK"
            if err!=nil{event.Status="ERROR"}else if phase!="PREPARE"&&digest!=c.Expected{event.Status="MISMATCH";err=errors.New("MISMATCH")}
            if emitErr:=emit(event);emitErr!=nil{return emitErr}
            if err!=nil{return err}
        }
    }
    return nil
}
