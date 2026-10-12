package statecontract

import (
	"fmt"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

func faultID(i int) string { return fmt.Sprintf("12345678-1234-4234-8234-%012d", i) }
func recoveryFixture() RecoverySection {
	return RecoverySection{Header: SectionHeader{DomainRecovery, Present}, Mark: StreamMark{1, 1, ref(RecoverySnapshot, 60)}, IdentityEpoch: 1, RecoveryID: "12345678-1234-4234-8234-123456789abc", SessionVersion: 1, Session: ref(RecoverySession, 62), RootPayload: d(20), MetadataPayload: d(22), SigningKeyID: d(2), Window: TimeWindow{0, 100}}
}

func d(n byte) contracts.Digest { return contracts.Digest{n} }
func ref(k ArtifactKind, n byte) ArtifactRef {
	r := ArtifactRef{Kind: k, Payload: d(n)}
	switch k {
	case RootRotation, Metadata, TrustBundle, ResourceEnvelope, HostMemoryPolicy, DesiredSnapshot, RecoverySnapshot:
		r.Envelope = d(n + 1)
	}
	return r
}
func fixture() MainStateContentModel {
	h := func(d Domain) SectionHeader { return SectionHeader{d, Present} }
	mark := func(k ArtifactKind, n byte) StreamMark { return StreamMark{1, 1, ref(k, n)} }
	scope := MainStateScope{NodeID: "12345678-1234-4234-8234-123456789abc", AuthorityEpoch: 1, DeploymentID: d(1), TrustAnchorID: d(2), DeploymentKeySetID: d(3), TimeProviderID: d(4), Main: GuardBinding{d(5), d(6)}, Supervisor: GuardBinding{d(7), d(8)}, Latch: GuardBinding{d(9), d(10)}}
	desired := DesiredBinding{1, 1, d(40)}
	bundle := func(p BundlePurpose, n byte) BundleRef {
		return BundleRef{Purpose: p, TrustDomain: "example.test", AuthorityEpoch: 1, Mark: mark(TrustBundle, n), DeploymentKeyID: d(11), DeauthorizedSetDigest: d(12)}
	}
	return MainStateContentModel{
		ModelVersion: 1, Scope: scope,
		Checkpoint:  CheckpointSection{h(DomainCheckpoint), scope.NodeID, 1, 100},
		Root:        RootSection{h(DomainRoot), mark(RootPayload, 20), 1, []contracts.Digest{d(1)}},
		Metadata:    MetadataSection{h(DomainMetadata), mark(Metadata, 22), 1, d(20), []contracts.Digest{d(2)}, nil, d(23), ref(MetadataHistory, 24), TimeWindow{0, 100}},
		Identity:    IdentitySection{h(DomainIdentity), 1, scope.NodeID, 200, d(1), d(2), d(3), d(4), ref(CertificateDER, 25), ref(IdentityHistory, 26), TimeWindow{0, 100}},
		ServerTrust: ServerTrustSection{h(DomainServerTrust), bundle(BootstrapServer, 27), bundle(AgentServer, 29)},
		Releases:    ReleaseSection{h(DomainReleases), d(30), VersionMark{1, ref(ApprovedManifest, 31)}, VersionMark{1, ref(InstalledMap, 32)}},
		Resources:   ResourceSection{h(DomainResources), mark(ResourceEnvelope, 33), scope.NodeID, 1, d(11)},
		HostPolicy:  HostPolicySection{h(DomainHostPolicy), mark(HostMemoryPolicy, 35), scope.NodeID, 1, d(11)},
		Desired:     DesiredSection{h(DomainDesired), mark(DesiredSnapshot, 40), 1, 1, d(33), d(20), d(22), d(2), TimeWindow{0, 100}, []string{"slot"}},
		Recovery:    RecoverySection{Header: SectionHeader{DomainRecovery, NeverEstablished}},
		Supervisor:  SupervisorSection{h(DomainSupervisor), d(45), scope.Supervisor, ref(SupervisorState, 46), AppliedRecord{Present, desired}, AppliedRecord{Present, desired}, TransitionRef{Coverage: NeverEstablished}},
		LKG:         LKGSection{h(DomainLKG), desired, ref(LKGHistory, 47), []LKGEntry{{"slot", desired, ref(ProcessSpec, 48), ref(ConfigMaterial, 49), ref(ReleaseMaterial, 50), d(20), d(22), d(2), 100}}},
	}
}
