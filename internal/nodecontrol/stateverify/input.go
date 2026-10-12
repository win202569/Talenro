package stateverify

// Batch borrows immutable-for-the-call envelopes. Anchor and expected coordinates
// must be supplied independently; this package does not authenticate their source.
type Batch struct {
	ExpectedNodeID                                        string
	ExpectedEpoch                                         uint64
	AnchorEnvelope                                        []byte
	Rotations                                             [][]byte
	ActiveRootEnvelope, MetadataEnvelope, DesiredEnvelope []byte
}
