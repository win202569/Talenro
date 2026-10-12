package stateformat

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"
)

var formatKinds = map[string]Kind{"desired": Desired, "root-initial": InitialRoot, "root-rotation": RootRotation, "metadata": Metadata}

// Golden payload/envelope hashes and transcript bytes were independently frozen
// with .NET, not calculated with the implementation under test.
func TestArtifactVectors(t *testing.T) {
	var vectors []struct {
		Name, Payload, Envelope string
		PayloadHash             string  `json:"payload_sha256"`
		EnvelopeHash            string  `json:"envelope_sha256"`
		Transcript              *string `json:"transcript_hex"`
	}
	if e := json.Unmarshal(fixture(t, "vectors"), &vectors); e != nil {
		t.Fatal(e)
	}
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			p := fixture(t, v.Name+".payload")
			e := fixture(t, v.Name+".envelope")
			a, err := ParseEnvelope(formatKinds[v.Name], e)
			if err != nil || a == nil {
				t.Fatal(err)
			}
			defer a.Clear()
			pd, ed := a.PayloadDigest(), a.EnvelopeDigest()
			if hex.EncodeToString(pd[:]) != v.PayloadHash || hex.EncodeToString(ed[:]) != v.EnvelopeHash || !bytes.Equal(a.Payload(), p) || !bytes.Equal(a.Envelope(), e) {
				t.Fatal("vector mismatch")
			}
			msg, err := a.Transcript()
			if v.Transcript == nil {
				if err != ErrInvalidFormat || msg != nil {
					t.Fatal("initial root transcript")
				}
			} else if err != nil || hex.EncodeToString(msg) != *v.Transcript {
				t.Fatal("transcript mismatch", err)
			}
		})
	}
}
