package statecontract

import "talenro.local/platform/internal/nodecontrol/contracts"

type Coverage uint8

const (
	Unknown Coverage = iota
	Present
	NeverEstablished
)

type Domain uint8

const (
	DomainUnknown Domain = iota
	DomainCheckpoint
	DomainRoot
	DomainMetadata
	DomainIdentity
	DomainServerTrust
	DomainReleases
	DomainResources
	DomainHostPolicy
	DomainDesired
	DomainRecovery
	DomainSupervisor
	DomainLKG
)

type SectionHeader struct {
	Domain   Domain
	Coverage Coverage
}
type GuardBinding struct{ CounterID, SealKeyID contracts.Digest }
type OptionalTime struct {
	Present     bool
	UnixSeconds int64
}
type TimeWindow struct{ Issued, Until int64 }
type MainStateScope struct {
	NodeID                                                          string
	AuthorityEpoch                                                  uint64
	DeploymentID, TrustAnchorID, DeploymentKeySetID, TimeProviderID contracts.Digest
	Main, Supervisor, Latch                                         GuardBinding
	TimeFloor                                                       OptionalTime
}
type ArtifactKind uint8

const (
	ArtifactUnknown ArtifactKind = iota
	RootPayload
	RootRotation
	Metadata
	MetadataHistory
	CertificateDER
	IdentityHistory
	TrustBundle
	ApprovedManifest
	InstalledMap
	ResourceEnvelope
	HostMemoryPolicy
	DesiredSnapshot
	RecoverySnapshot
	RecoverySession
	SupervisorState
	TransitionCapsule
	LKGHistory
	ProcessSpec
	ConfigMaterial
	ReleaseMaterial
)

type ArtifactRef struct {
	Kind              ArtifactKind
	Payload, Envelope contracts.Digest
}
type StreamMark struct {
	Version, Sequence uint64
	Material          ArtifactRef
}
type VersionMark struct {
	Version  uint64
	Material ArtifactRef
}
type DesiredBinding struct {
	Generation, Sequence uint64
	Payload              contracts.Digest
}
type CheckpointSection struct {
	Header                   SectionHeader
	NodeID                   string
	AuthorityEpoch, Sequence uint64
}
type RootSection struct {
	Header    SectionHeader
	Mark      StreamMark
	Threshold uint8
	KeyIDs    []contracts.Digest
}
type MetadataSection struct {
	Header             SectionHeader
	Mark               StreamMark
	RootVersion        uint64
	RootPayload        contracts.Digest
	KeyIDs, RevokedIDs []contracts.Digest
	RevokedSetDigest   contracts.Digest
	History            ArtifactRef
	Window             TimeWindow
}
type IdentitySection struct {
	Header                                 SectionHeader
	IdentityEpoch                          uint64
	LineageID                              string
	LineageUntil                           int64
	IssuerID, Serial, PublicKey, KeyHandle contracts.Digest
	Leaf, History                          ArtifactRef
	Window                                 TimeWindow
}
type ServerTrustSection struct {
	Header           SectionHeader
	Bootstrap, Agent BundleRef
}
type ReleaseSection struct {
	Header        SectionHeader
	BuildID       contracts.Digest
	Manifest, Map VersionMark
}
type ResourceSection struct {
	Header          SectionHeader
	Mark            StreamMark
	NodeID          string
	AuthorityEpoch  uint64
	DeploymentKeyID contracts.Digest
}
type HostPolicySection struct {
	Header          SectionHeader
	Mark            StreamMark
	NodeID          string
	AuthorityEpoch  uint64
	DeploymentKeyID contracts.Digest
}
type DesiredSection struct {
	Header                                                      SectionHeader
	Mark                                                        StreamMark
	InventoryVersion, ResourceVersion                           uint64
	ResourcePayload, RootPayload, MetadataPayload, SigningKeyID contracts.Digest
	Window                                                      TimeWindow
	SlotIDs                                                     []string
}
type RecoverySection struct {
	Header                                     SectionHeader
	Mark                                       StreamMark
	IdentityEpoch                              uint64
	RecoveryID                                 string
	SessionVersion                             uint64
	Session                                    ArtifactRef
	RootPayload, MetadataPayload, SigningKeyID contracts.Digest
	Window                                     TimeWindow
	Local                                      []LocalFaultBinding
	Supervisor                                 []SupervisorFaultBinding
}
type SupervisorSection struct {
	Header            SectionHeader
	Identity          contracts.Digest
	Guard             GuardBinding
	State             ArtifactRef
	Verified, Applied AppliedRecord
	Transition        TransitionRef
}
type LKGSection struct {
	Header          SectionHeader
	HighestAccepted DesiredBinding
	History         ArtifactRef
	Entries         []LKGEntry
}
type BundlePurpose uint8

const (
	PurposeUnknown BundlePurpose = iota
	BootstrapServer
	AgentServer
)

type BundleRef struct {
	Purpose               BundlePurpose
	TrustDomain           string
	AuthorityEpoch        uint64
	Mark                  StreamMark
	DeploymentKeyID       contracts.Digest
	Deauthorized          []contracts.Digest
	DeauthorizedSetDigest contracts.Digest
}
type AppliedRecord struct {
	Coverage Coverage
	Desired  DesiredBinding
}
type TransitionRef struct {
	Coverage                        Coverage
	Material                        ArtifactRef
	PriorApplied, CandidateVerified DesiredBinding
}
type LocalFaultBinding struct{ FaultID, IncidentID string }
type SupervisorFaultBinding struct {
	FaultID, IncidentID string
	Evidence            contracts.Digest
}
type LKGEntry struct {
	SlotID                                     string
	Desired                                    DesiredBinding
	Process, Config, Release                   ArtifactRef
	RootPayload, MetadataPayload, SigningKeyID contracts.Digest
	Deadline                                   int64
}
type RotationRef struct {
	PreviousVersion, NewVersion, Sequence uint64
	OldRootPayload, NewRootPayload        contracts.Digest
	Material                              ArtifactRef
}
type MainStateContentModel struct {
	ModelVersion uint32
	Scope        MainStateScope
	Checkpoint   CheckpointSection
	Root         RootSection
	Metadata     MetadataSection
	Identity     IdentitySection
	ServerTrust  ServerTrustSection
	Releases     ReleaseSection
	Resources    ResourceSection
	HostPolicy   HostPolicySection
	Desired      DesiredSection
	Recovery     RecoverySection
	Supervisor   SupervisorSection
	LKG          LKGSection
}
type ModelInput struct {
	Previous, Proposed MainStateContentModel
	RootTransition     []RotationRef
}
