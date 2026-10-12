package statepolicy

import (
	"talenro.local/platform/internal/nodecontrol/stateformat"
	"talenro.local/platform/internal/nodecontrol/stateverify"
)

type artifactFact struct {
	bindings stateformat.Bindings
	id       stateverify.ArtifactIdentity
	root     stateformat.RootMaterial
	metadata stateformat.MetadataMaterial
}

func (f *artifactFact) clear() { f.root.Clear(); f.metadata.Clear(); *f = artifactFact{} }

type verifiedFacts struct {
	anchor, metadata, desired artifactFact
	rotations                 []artifactFact
}

func (f *verifiedFacts) Clear() {
	if f == nil {
		return
	}
	f.anchor.clear()
	f.metadata.clear()
	f.desired.clear()
	for i := range f.rotations {
		f.rotations[i].clear()
	}
	clear(f.rotations)
	*f = verifiedFacts{}
}
