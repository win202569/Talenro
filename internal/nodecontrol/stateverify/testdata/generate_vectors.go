//go:build ignore

// Offline, test-only vector generator. Writes a JSON transport to stdout, never
// imports product code. All values are ASCII; encoding/json sorts map keys.
package main
import("crypto/ed25519";"crypto/sha256";"encoding/base64";"encoding/hex";"encoding/json";"fmt")
func enc(v any)[]byte{b,e:=json.Marshal(v);if e!=nil{panic(e)};return b}
func b64(b []byte)string{return base64.RawURLEncoding.EncodeToString(b)}
func hash(b []byte)string{h:=sha256.Sum256(b);return hex.EncodeToString(h[:])}
func key(n byte)ed25519.PrivateKey{seed:=make([]byte,32);seed[0]=n;return ed25519.NewKeyFromSeed(seed)}
func main(){
 rootKey,desiredKey:=key(1),key(2);rp,dp:=rootKey.Public().(ed25519.PublicKey),desiredKey.Public().(ed25519.PublicKey)
 root:=func(version,seq string)map[string]any{return map[string]any{"schema_version":"node-state-root-set.v1","control_plane_authority_epoch":"19","authority_sequence":seq,"root_version":version,"threshold":"1","keys":[]any{map[string]any{"algorithm":"ed25519","key_id":hash(rp),"public_key":b64(rp)}}}}
 payloads:=map[string][]byte{
 "anchor":enc(root("1","3")),
 "rotation":enc(map[string]any{"schema_version":"node-state-root-rotation.v1","control_plane_authority_epoch":"19","authority_sequence":"4","previous_root_version":"1","new_root_set":root("2","4"),"issued_at":"2020-01-01T00:00:00Z","valid_until":"2020-01-08T00:00:00Z"}),
 "metadata":enc(map[string]any{"schema_version":"node-state-trust-metadata.v1","control_plane_authority_epoch":"19","authority_sequence":"5","root_version":"2","metadata_version":"6","role":"node_state","issued_at":"2020-01-01T00:00:00Z","valid_until":"2020-01-08T00:00:00Z","next_refresh_at":"2020-01-06T00:00:00Z","revoked_key_ids":[]any{},"keys":[]any{map[string]any{"algorithm":"ed25519","key_id":hash(dp),"public_key":b64(dp),"status":"active","not_before":"2019-12-31T00:00:00Z","not_after":"2020-01-10T00:00:00Z","revoked_at":nil,"revocation_reason":nil,"incident_id":nil}}}),
 "desired":enc(map[string]any{"schema_version":"node-desired-state.v1","control_plane_authority_epoch":"19","authority_sequence":"9","node_id":"12345678-1234-4234-8234-123456789abc","generation":"7","inventory_version":"8","node_resource_envelope_version":"9","node_resource_envelope_digest":"abababababababababababababababababababababababababababababababab","issued_at":"2020-01-01T00:00:00Z","valid_until":"2020-01-02T00:00:00Z","minimum_agent_version":"1.2.3","minimum_supervisor_version":"0.0.0","processes":[]any{},"reason":"initial"}),
 }
 files:=map[string]string{};vectors:=[]map[string]string{}
 for _,name:=range []string{"anchor","rotation","metadata","desired"}{
  p:=payloads[name];e:=map[string]any{"payload_b64url":b64(p),"payload_sha256":hash(p)}
  v:=map[string]string{"Name":name,"PayloadSHA256":hash(p)}
  domain:="";signer:=rootKey;pub:=rp
  switch name{
  case "anchor":e["schema_version"]="talenro-node-state-root-initial/v1";e["root_version"]="1";e["authority_sequence"]="3"
  case "rotation":e["schema_version"]="talenro-node-state-root-rotation/v1";e["root_version"]="2";e["previous_root_version"]="1";e["authority_sequence"]="4";domain="TALENRO-NODE-STATE-ROOT-ROTATION-V1"
  case "metadata":e["schema_version"]="talenro-node-state-metadata/v1";e["root_version"]="2";e["authority_sequence"]="5";domain="TALENRO-NODE-STATE-METADATA-V1"
  case "desired":e["schema_version"]="talenro-node-desired-state/v1";domain="TALENRO-NODE-DESIRED-STATE-V1";signer=desiredKey;pub=dp
  }
  if name!="desired"{e["control_plane_authority_epoch"]="19"}
  if domain!=""{
   msg:=append([]byte(domain+"\x00"),p...);sig:=ed25519.Sign(signer,msg);share:=map[string]any{"key_id":hash(pub),"signature":b64(sig)}
   switch name{case "rotation":e["current_signatures"]=[]any{share};e["new_signatures"]=[]any{share};case "metadata":e["signatures"]=[]any{share};case "desired":e["key_id"]=hash(pub);e["signature"]=b64(sig)}
   v["TranscriptHex"]=hex.EncodeToString(msg);v["PublicKey"]=b64(pub);v["Signature"]=b64(sig)
   files[name+".transcript.hex"]=b64([]byte(v["TranscriptHex"]))
  }
  raw:=enc(e);v["EnvelopeSHA256"]=hash(raw);vectors=append(vectors,v)
  files[name+".payload.json"]=b64(p);files[name+".envelope.json"]=b64(raw)
 }
 files["vectors.json"]=b64(enc(vectors))
 files["README.md"]=b64([]byte("Test-only fixed vectors. Go standard Ed25519, seeds: 32 bytes all zero except byte 0 = 1 (root), 2 (desired). Never production keys. ASCII literal payloads with sorted JSON keys; ordinary signatures over complete domain + NUL + payload. Transcript hex is decoded before comparison. Generated independently of stateformat/stateverify.\n"))
 fmt.Print(string(enc(files)))
}
