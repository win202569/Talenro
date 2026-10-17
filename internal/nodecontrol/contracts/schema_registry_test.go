package contracts

import (
	"encoding/json"
	"os"
	"testing"
)

type metadataFixtureEntry struct {
	Schema           string    `json:"schema"`
	ObjectName       string    `json:"object_name"`
	SignerRoleOrNull *string   `json:"signer_role_or_null"`
	EvidenceKinds    []*string `json:"evidence_kinds"`
}
type metadataFramingVector struct {
	Schema        string `json:"schema"`
	CanonicalBody string `json:"canonical_body"`
	PreimageHex   string `json:"preimage_hex"`
	SHA256Hex     string `json:"sha256_hex"`
}
type metadataRegistryFixture struct {
	Schema   string                 `json:"schema"`
	Coverage string                 `json:"coverage"`
	Entries  []metadataFixtureEntry `json:"entries"`
}
type metadataGoldenFixture struct {
	Schema          string                  `json:"schema"`
	Coverage        string                  `json:"coverage"`
	ExpectedEntries []metadataFixtureEntry  `json:"expected_entries"`
	FramingVectors  []metadataFramingVector `json:"framing_vectors"`
}

// Catches missing or incomplete independent literals,including same-count identity drift.
func TestAuthorityV7MetadataFixtureInventory(t *testing.T) {
	rb, err := os.ReadFile(nameFixtureBase + "schema-registry.v1.json")
	if err != nil {
		t.Fatalf("required metadata fixture schema-registry.v1.json: %v", err)
	}
	gb, err := os.ReadFile(nameFixtureBase + "schema-registry-golden.v1.json")
	if err != nil {
		t.Fatalf("required metadata fixture schema-registry-golden.v1.json: %v", err)
	}
	var registry metadataRegistryFixture
	var golden metadataGoldenFixture
	if err := json.Unmarshal(rb, &registry); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(gb, &golden); err != nil {
		t.Fatal(err)
	}
	if registry.Schema != "talenro.c12.authority-v7.schema-registry.v1" || registry.Coverage != "metadata-registry" || golden.Schema != "talenro.c12.authority-v7.schema-registry-golden.v1" || golden.Coverage != "registry-and-domain-framing" {
		t.Fatal("metadata fixture markers")
	}
	entries, vectors := golden.ExpectedEntries, golden.FramingVectors
	if len(entries) != 146 || len(registry.Entries) != 146 {
		t.Fatal("expected exactly146 metadata rows")
	}
	if len(vectors) != 3 {
		t.Fatalf("vectors=%d", len(vectors))
	}
	names, _, err := readNameGoldenFixture()
	if err != nil {
		t.Fatal(err)
	}
	roles := map[string]bool{}
	counts := map[string]int{}
	objects := map[string]bool{}
	for i, entry := range entries {
		if entry.Schema != names[i].Schema || entry.ObjectName != names[i].ObjectName || objects[entry.ObjectName] || (i > 0 && entries[i-1].Schema >= entry.Schema) {
			t.Fatalf("metadata identity row%d", i)
		}
		objects[entry.ObjectName] = true
		if entry.EvidenceKinds == nil {
			t.Fatal("missing/null evidence_kinds")
		}
		role := "null"
		if entry.SignerRoleOrNull != nil {
			role = *entry.SignerRoleOrNull
			if role == "" {
				t.Fatal("empty signer role")
			}
			roles[role] = true
		}
		kind := "[]"
		if len(entry.EvidenceKinds) > 1 {
			t.Fatal("too many evidence kinds")
		}
		if len(entry.EvidenceKinds) == 1 {
			if entry.EvidenceKinds[0] == nil {
				t.Fatal("null evidence kind")
			}
			kind = *entry.EvidenceKinds[0]
		}
		counts[kind+"/"+map[bool]string{true: "signed", false: "null"}[role != "null"]]++
		peer := registry.Entries[i]
		if entry.Schema != peer.Schema || entry.ObjectName != peer.ObjectName || (entry.SignerRoleOrNull == nil) != (peer.SignerRoleOrNull == nil) || len(entry.EvidenceKinds) != len(peer.EvidenceKinds) || peer.EvidenceKinds == nil {
			t.Fatalf("registry/golden mismatch row%d", i)
		}
		if entry.SignerRoleOrNull != nil && *entry.SignerRoleOrNull != *peer.SignerRoleOrNull {
			t.Fatal("role mismatch")
		}
		if len(entry.EvidenceKinds) == 1 && (peer.EvidenceKinds[0] == nil || *entry.EvidenceKinds[0] != *peer.EvidenceKinds[0]) {
			t.Fatal("kind mismatch")
		}
	}
	if len(roles) != 37 || counts["external_signed_envelope/signed"] != 71 || counts["database_immutable_body/null"] != 27 || counts["[]/signed"] != 32 || counts["[]/null"] != 16 || len(counts) != 4 {
		t.Fatalf("metadata categories: roles=%d counts=%v", len(roles), counts)
	}
}
