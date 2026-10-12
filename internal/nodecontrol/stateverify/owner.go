package stateverify

import "talenro.local/platform/internal/nodecontrol/stateformat"

type verificationOwner struct {
	artifacts   [36]*stateformat.ParsedArtifact
	roots       [33]stateformat.RootMaterial
	metadata    stateformat.MetadataMaterial
	shares      [33]stateformat.SignatureMaterial
	transcripts [34][]byte
	evidence    *SignatureEvidence
}

func (o *verificationOwner) Clear() {
	if o == nil {
		return
	}
	for _, a := range o.artifacts {
		a.Clear()
	}
	for i := range o.roots {
		o.roots[i].Clear()
	}
	o.metadata.Clear()
	for i := range o.shares {
		o.shares[i].Clear()
	}
	for _, b := range o.transcripts {
		clear(b)
	}
	o.evidence.Clear()
	*o = verificationOwner{}
}
