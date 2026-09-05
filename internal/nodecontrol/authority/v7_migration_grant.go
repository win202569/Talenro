package authority

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/gowebpki/jcs"
	"talenro.local/platform/internal/nodecontrol/contracts"
)

const (
	authorityV7CanonicalArtifactMaxBytes = 64 * 1024
	authorityV7EvidenceBundleMaxBytes    = 1024 * 1024
)

type authorityV7StrictJSONSpec struct {
	Fields           []string
	NullablePaths    map[string]bool
	NullablePrefixes []string
	MaximumBytes     int
}

type authorityV7CanonicalBodyExpectation struct {
	Schema          string
	Fields          []string
	NullablePaths   map[string]bool
	ExpectedBodyJCS []byte
	ExpectedDigest  contracts.Digest
}

type authorityV7SignerPolicy struct {
	Schema                 string
	SignerRole             string
	SignerKeyID            string
	SignaturePolicyVersion string
	TrustRootDigest        string
	SignatureAlgorithm     string
	PublicKey              ed25519.PublicKey
}

func authorityV7FixtureSignerRole(schema string) string {
	switch schema {
	case "environment-inventory-membership-retirement.v1":
		return "release_deployment_operator"
	case "fresh-restore-import-manifest.v1":
		return "fresh_restore_export_operator"
	case "provider-protocol-downgrade-retirement-authorization.v1":
		return "provider_downgrade_retirement_admin"
	case "provider-protocol-history-zero-projection.v1":
		return "claim_v1_provider_history_auditor"
	case "provider-protocol-downgrade-retirement.v1", "claim-v1-provider-head.v1", "fresh-v7-staging-exclusion-lease.v1", "database-incarnation-proof.v1":
		return "claim_v1_provider"
	case "authority-protocol-downgrade-authorization.v1":
		return "authority_protocol_downgrade_authorizer"
	case "staging-import-capability.v1":
		return "fresh_restore_staging_import_authorizer"
	default:
		return ""
	}
}

type authorityV7EnvelopeExpectation struct {
	Body        authorityV7CanonicalBodyExpectation
	EnvelopeJCS []byte
	Policy      authorityV7SignerPolicy
}

type authorityV7EvidenceMemberExpectation struct {
	EvidenceKind string
	Envelope     *authorityV7EnvelopeExpectation
	Body         *authorityV7CanonicalBodyExpectation
}

type VerifiedAuthorityV7UpGrant struct {
	state *verifiedAuthorityV7UpGrantState
}

type verifiedAuthorityV7UpGrantState struct {
	facts    contracts.AuthorityV7UpMigrationFactsV1
	consumed atomic.Bool
}

type VerifiedAuthorityV7DownGrant struct {
	state *verifiedAuthorityV7DownGrantState
}

type authorityV7DownPersistenceView struct {
	Facts                               contracts.AuthorityV7DownMigrationFactsV1
	ProviderRetirementSetBodyJCS        []byte
	ProviderRetirementEvidenceBundleJCS []byte
}

type verifiedAuthorityV7DownGrantState struct {
	view     authorityV7DownPersistenceView
	consumed atomic.Bool
}

func validateAuthorityV7EffectiveWindowAt(issuedAt, expiresAt, now time.Time) error {
	if now.Before(issuedAt) || !now.Before(expiresAt) {
		return ErrInvalidArgument
	}
	return nil
}

func newVerifiedAuthorityV7UpGrant(facts contracts.AuthorityV7UpMigrationFactsV1) (VerifiedAuthorityV7UpGrant, error) {
	cloned := cloneAuthorityV7UpMigrationFacts(facts)
	if err := cloned.Validate(); err != nil {
		return VerifiedAuthorityV7UpGrant{}, ErrInvalidArgument
	}
	return VerifiedAuthorityV7UpGrant{state: &verifiedAuthorityV7UpGrantState{facts: cloned}}, nil
}

func newVerifiedAuthorityV7DownGrant(view authorityV7DownPersistenceView) (VerifiedAuthorityV7DownGrant, error) {
	cloned := cloneAuthorityV7DownPersistenceView(view)
	if err := cloned.Facts.Validate(); err != nil {
		return VerifiedAuthorityV7DownGrant{}, ErrInvalidArgument
	}
	return VerifiedAuthorityV7DownGrant{state: &verifiedAuthorityV7DownGrantState{view: cloned}}, nil
}

func ConsumeVerifiedAuthorityV7UpGrant(grant VerifiedAuthorityV7UpGrant) (contracts.AuthorityV7UpMigrationFactsV1, error) {
	if grant.state == nil {
		return contracts.AuthorityV7UpMigrationFactsV1{}, ErrInvalidArgument
	}
	if !grant.state.consumed.CompareAndSwap(false, true) {
		return contracts.AuthorityV7UpMigrationFactsV1{}, ErrConflict
	}
	facts := cloneAuthorityV7UpMigrationFacts(grant.state.facts)
	now := time.Now()
	if !now.Before(facts.ExpiresAt) || facts.Validate() != nil {
		return contracts.AuthorityV7UpMigrationFactsV1{}, ErrInvalidArgument
	}
	return facts, nil
}

func ConsumeVerifiedAuthorityV7DownGrant(grant VerifiedAuthorityV7DownGrant) (authorityV7DownPersistenceView, error) {
	if grant.state == nil {
		return authorityV7DownPersistenceView{}, ErrInvalidArgument
	}
	if !grant.state.consumed.CompareAndSwap(false, true) {
		return authorityV7DownPersistenceView{}, ErrConflict
	}
	view := cloneAuthorityV7DownPersistenceView(grant.state.view)
	now := time.Now()
	if !now.Before(view.Facts.ExpiresAt) || view.Facts.Validate() != nil {
		return authorityV7DownPersistenceView{}, ErrInvalidArgument
	}
	return view, nil
}

func cloneAuthorityV7UpMigrationFacts(value contracts.AuthorityV7UpMigrationFactsV1) contracts.AuthorityV7UpMigrationFactsV1 {
	result := value
	if value.UpgradeIntentOrNull != nil {
		intent := *value.UpgradeIntentOrNull
		if intent.CredentialPolicyUpdateID != nil {
			identifier := *intent.CredentialPolicyUpdateID
			intent.CredentialPolicyUpdateID = &identifier
		}
		result.UpgradeIntentOrNull = &intent
	}
	if value.UpgradeIntentDigestOrNull != nil {
		digest := *value.UpgradeIntentDigestOrNull
		result.UpgradeIntentDigestOrNull = &digest
	}
	return result
}

func cloneAuthorityV7DownMigrationFacts(value contracts.AuthorityV7DownMigrationFactsV1) contracts.AuthorityV7DownMigrationFactsV1 {
	result := value
	result.StableTableInventory = append([]contracts.AuthorityV7StableTableInventoryItemV1(nil), value.StableTableInventory...)
	result.ProviderRetirementSet = cloneAuthorityV7ProviderRetirementSet(value.ProviderRetirementSet)
	return result
}

func cloneAuthorityV7DownPersistenceView(value authorityV7DownPersistenceView) authorityV7DownPersistenceView {
	result := value
	result.Facts = cloneAuthorityV7DownMigrationFacts(value.Facts)
	result.ProviderRetirementSetBodyJCS = cloneAuthorityV7Bytes(value.ProviderRetirementSetBodyJCS)
	result.ProviderRetirementEvidenceBundleJCS = cloneAuthorityV7Bytes(value.ProviderRetirementEvidenceBundleJCS)
	return result
}

func cloneAuthorityV7ProviderRetirementSet(value contracts.AuthorityV7ProviderRetirementSetFactsV1) contracts.AuthorityV7ProviderRetirementSetFactsV1 {
	result := value
	result.Retirements = append([]contracts.AuthorityV7ProviderRetirementFactsV1(nil), value.Retirements...)
	result.RetiredMembers = append([]contracts.AuthorityV7RetiredEnvironmentMemberFactsV1(nil), value.RetiredMembers...)
	return result
}

func cloneAuthorityV7DownAuthorizationRequest(value contracts.AuthorityV7DownAuthorizationRequestV1) contracts.AuthorityV7DownAuthorizationRequestV1 {
	result := value
	result.PristineInventory.StableTableInventory = append([]contracts.AuthorityV7StableTableInventoryItemV1(nil), value.PristineInventory.StableTableInventory...)
	result.ProviderRetirementSet = cloneAuthorityV7ProviderRetirementSet(value.ProviderRetirementSet)
	return result
}

func cloneAuthorityV7DownAuthorizationFacts(value contracts.AuthorityV7DownAuthorizationFactsV1) contracts.AuthorityV7DownAuthorizationFactsV1 {
	result := value
	result.Request = cloneAuthorityV7DownAuthorizationRequest(value.Request)
	result.AuthorizationEnvelopeJCS = append([]byte(nil), value.AuthorizationEnvelopeJCS...)
	return result
}

func cloneFreshRestoreImportProjectionInput(value contracts.FreshRestoreImportProjectionInputV1) contracts.FreshRestoreImportProjectionInputV1 {
	result := value
	result.ManifestTopology.Objects = make([]contracts.FreshRestoreImportManifestObjectV1, len(value.ManifestTopology.Objects))
	for index, object := range value.ManifestTopology.Objects {
		result.ManifestTopology.Objects[index] = object
		result.ManifestTopology.Objects[index].Payload = append([]byte(nil), object.Payload...)
	}
	if value.LatestDatabaseAuthorityRebindResultDigestOrNull != nil {
		digest := *value.LatestDatabaseAuthorityRebindResultDigestOrNull
		result.LatestDatabaseAuthorityRebindResultDigestOrNull = &digest
	}
	return result
}

func cloneFreshImportTopologyProjection(value contracts.FreshImportTopologyProjectionV1) contracts.FreshImportTopologyProjectionV1 {
	result := value
	result.Objects = nil
	if value.Objects != nil {
		result.Objects = make([]contracts.FreshImportTopologyProjectionObjectV1, len(value.Objects))
		for index, object := range value.Objects {
			result.Objects[index] = object
			result.Objects[index].NormalizedPayload = cloneAuthorityV7Bytes(object.NormalizedPayload)
		}
	}
	return result
}

func cloneAuthorityV7Bytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	result := make([]byte, len(value))
	copy(result, value)
	return result
}

func equalAuthorityV7Bytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func authorityV7DomainDigest(schema string, bodyJCS []byte) contracts.Digest {
	hash := sha256.New()
	_, _ = hash.Write([]byte("talenro.c12." + schema))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write(bodyJCS)
	var result contracts.Digest
	copy(result[:], hash.Sum(nil))
	return result
}

// validateAuthorityV7StrictJSON is the sole byte-level JSON/JCS admission
// gate used by the v7 fixture verifiers. Schema and cross-artifact checks are
// deliberately layered after this structural gate.
func validateAuthorityV7StrictJSON(raw []byte, spec authorityV7StrictJSONSpec) (map[string]json.RawMessage, error) {
	maximum := spec.MaximumBytes
	if maximum == 0 {
		maximum = authorityV7CanonicalArtifactMaxBytes
	}
	if len(raw) == 0 || len(raw) > maximum || !utf8.Valid(raw) || bytes.HasPrefix(raw, []byte{0xef, 0xbb, 0xbf}) {
		return nil, ErrInvalidArgument
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nulls := make([]string, 0)
	if err := inspectAuthorityV7JSONValue(decoder, "", &nulls); err != nil {
		return nil, ErrInvalidArgument
	}
	if token, err := decoder.Token(); err != io.EOF || token != nil {
		return nil, ErrInvalidArgument
	}

	valueDecoder := json.NewDecoder(bytes.NewReader(raw))
	valueDecoder.UseNumber()
	var value json.RawMessage
	if err := valueDecoder.Decode(&value); err != nil {
		return nil, ErrInvalidArgument
	}
	var trailing json.RawMessage
	if err := valueDecoder.Decode(&trailing); err != io.EOF {
		return nil, ErrInvalidArgument
	}

	objectDecoder := json.NewDecoder(bytes.NewReader(value))
	objectDecoder.UseNumber()
	var object map[string]json.RawMessage
	if err := objectDecoder.Decode(&object); err != nil || object == nil {
		return nil, ErrInvalidArgument
	}
	if len(object) != len(spec.Fields) {
		return nil, ErrInvalidArgument
	}
	for _, field := range spec.Fields {
		fieldValue, present := object[field]
		if !present || len(fieldValue) == 0 {
			return nil, ErrInvalidArgument
		}
	}
	for field := range object {
		if !authorityV7ContainsString(spec.Fields, field) {
			return nil, ErrInvalidArgument
		}
	}
	for _, path := range nulls {
		if spec.NullablePaths[path] {
			continue
		}
		allowed := false
		for _, prefix := range spec.NullablePrefixes {
			if strings.HasPrefix(path, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, ErrInvalidArgument
		}
	}

	canonical, err := jcs.Transform(raw)
	if err != nil || !bytes.Equal(canonical, raw) {
		return nil, ErrInvalidArgument
	}
	return object, nil
}

func inspectAuthorityV7JSONValue(decoder *json.Decoder, path string, nulls *[]string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		*nulls = append(*nulls, path)
		return nil
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, keyErr := decoder.Token()
			key, stringKey := keyToken.(string)
			if keyErr != nil || !stringKey || seen[key] {
				return ErrInvalidArgument
			}
			seen[key] = true
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			if err := inspectAuthorityV7JSONValue(decoder, childPath, nulls); err != nil {
				return err
			}
		}
		end, endErr := decoder.Token()
		if endErr != nil || end != json.Delim('}') {
			return ErrInvalidArgument
		}
	case '[':
		arrayPath := path + "[]"
		for decoder.More() {
			if err := inspectAuthorityV7JSONValue(decoder, arrayPath, nulls); err != nil {
				return err
			}
		}
		end, endErr := decoder.Token()
		if endErr != nil || end != json.Delim(']') {
			return ErrInvalidArgument
		}
	default:
		return ErrInvalidArgument
	}
	return nil
}

func authorityV7ContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func verifyAuthorityV7CanonicalBody(bodyJCS []byte, expectation authorityV7CanonicalBodyExpectation) (map[string]json.RawMessage, error) {
	fields, err := validateAuthorityV7StrictJSON(bodyJCS, authorityV7StrictJSONSpec{
		Fields:        expectation.Fields,
		NullablePaths: expectation.NullablePaths,
		MaximumBytes:  authorityV7CanonicalArtifactMaxBytes,
	})
	if err != nil || expectation.Schema == "" {
		return nil, ErrInvalidArgument
	}
	if expectation.ExpectedBodyJCS != nil && !bytes.Equal(bodyJCS, expectation.ExpectedBodyJCS) {
		return nil, ErrInvalidArgument
	}
	if authorityV7DomainDigest(expectation.Schema, bodyJCS) != expectation.ExpectedDigest {
		return nil, ErrInvalidArgument
	}
	return fields, nil
}

func verifyAuthorityV7SignedEnvelope(bodyJCS, envelopeJCS []byte, expectation authorityV7EnvelopeExpectation) error {
	if _, err := verifyAuthorityV7CanonicalBody(bodyJCS, expectation.Body); err != nil {
		return ErrInvalidArgument
	}
	if expectation.EnvelopeJCS != nil && !bytes.Equal(envelopeJCS, expectation.EnvelopeJCS) {
		return ErrInvalidArgument
	}
	fields, err := validateAuthorityV7StrictJSON(envelopeJCS, authorityV7StrictJSONSpec{
		Fields: []string{
			"schema", "body", "body_digest", "signer_role", "signer_key_id",
			"signature_policy_version", "trust_root_digest", "signature_algorithm", "signature",
		},
		NullablePrefixes: []string{"body."},
		MaximumBytes:     authorityV7CanonicalArtifactMaxBytes,
	})
	if err != nil {
		return ErrInvalidArgument
	}
	if !bytes.Equal(fields["body"], bodyJCS) {
		return ErrInvalidArgument
	}
	policy := expectation.Policy
	schema, err := authorityV7JSONString(fields["schema"])
	if err != nil || schema != expectation.Body.Schema || policy.Schema != schema {
		return ErrInvalidArgument
	}
	bodyDigest, err := authorityV7JSONDigest(fields["body_digest"])
	if err != nil || bodyDigest != expectation.Body.ExpectedDigest {
		return ErrInvalidArgument
	}
	signerRole, err := authorityV7JSONString(fields["signer_role"])
	if err != nil || signerRole != policy.SignerRole || !authorityV7PrintableASCII(signerRole, 1, 128) {
		return ErrInvalidArgument
	}
	signerKeyID, err := authorityV7JSONString(fields["signer_key_id"])
	if err != nil || signerKeyID != policy.SignerKeyID || !authorityV7PrintableASCII(signerKeyID, 1, 128) {
		return ErrInvalidArgument
	}
	policyVersion, err := authorityV7JSONString(fields["signature_policy_version"])
	if err != nil || policyVersion != policy.SignaturePolicyVersion || !authorityV7CanonicalPositiveDecimal(policyVersion) {
		return ErrInvalidArgument
	}
	trustRootDigest, err := authorityV7JSONString(fields["trust_root_digest"])
	if err != nil || trustRootDigest != policy.TrustRootDigest || !authorityV7CanonicalDigestText(trustRootDigest) {
		return ErrInvalidArgument
	}
	algorithm, err := authorityV7JSONString(fields["signature_algorithm"])
	if err != nil || algorithm != policy.SignatureAlgorithm || algorithm != "ed25519" || len(policy.PublicKey) != ed25519.PublicKeySize {
		return ErrInvalidArgument
	}
	signatureText, err := authorityV7JSONString(fields["signature"])
	if err != nil {
		return ErrInvalidArgument
	}
	signature, err := base64.RawURLEncoding.DecodeString(signatureText)
	if err != nil || len(signature) != ed25519.SignatureSize || base64.RawURLEncoding.EncodeToString(signature) != signatureText {
		return ErrInvalidArgument
	}

	metadata, err := json.Marshal(map[string]string{
		"schema":                   schema,
		"body_digest":              hex.EncodeToString(bodyDigest[:]),
		"signer_role":              signerRole,
		"signer_key_id":            signerKeyID,
		"signature_policy_version": policyVersion,
		"trust_root_digest":        trustRootDigest,
		"signature_algorithm":      algorithm,
	})
	if err != nil {
		return ErrInvalidArgument
	}
	metadata, err = jcs.Transform(metadata)
	if err != nil {
		return ErrInvalidArgument
	}
	signatureInput := make([]byte, 0, len("talenro.c12.signature-envelope.v1")+1+len(metadata))
	signatureInput = append(signatureInput, "talenro.c12.signature-envelope.v1"...)
	signatureInput = append(signatureInput, 0)
	signatureInput = append(signatureInput, metadata...)
	if !ed25519.Verify(policy.PublicKey, signatureInput, signature) {
		return ErrInvalidArgument
	}
	return nil
}

func verifyAuthorityV7EvidenceBundle(bundleJCS []byte, messageSchema string, messageDigest contracts.Digest, expected []authorityV7EvidenceMemberExpectation) error {
	fields, err := validateAuthorityV7StrictJSON(bundleJCS, authorityV7StrictJSONSpec{
		Fields: []string{"message_schema", "message_body_digest", "evidence_count", "evidence"},
		NullablePaths: map[string]bool{
			"evidence[].canonical_body_or_null":     true,
			"evidence[].canonical_envelope_or_null": true,
		},
		NullablePrefixes: []string{"evidence[].canonical_body_or_null.", "evidence[].canonical_envelope_or_null.body."},
		MaximumBytes:     authorityV7EvidenceBundleMaxBytes,
	})
	if err != nil {
		return ErrInvalidArgument
	}
	gotSchema, err := authorityV7JSONString(fields["message_schema"])
	if err != nil || gotSchema != messageSchema {
		return ErrInvalidArgument
	}
	gotDigest, err := authorityV7JSONDigest(fields["message_body_digest"])
	if err != nil || gotDigest != messageDigest {
		return ErrInvalidArgument
	}
	countText, err := authorityV7JSONString(fields["evidence_count"])
	if err != nil || !authorityV7CanonicalNonnegativeDecimal(countText) {
		return ErrInvalidArgument
	}
	count, err := strconv.ParseUint(countText, 10, 64)
	if err != nil || count != uint64(len(expected)) {
		return ErrInvalidArgument
	}
	var items []json.RawMessage
	if err := json.Unmarshal(fields["evidence"], &items); err != nil || items == nil || len(items) != len(expected) {
		return ErrInvalidArgument
	}

	var previousDigest contracts.Digest
	previousSchema := ""
	for index, itemJCS := range items {
		itemFields, itemErr := validateAuthorityV7StrictJSON(itemJCS, authorityV7StrictJSONSpec{
			Fields: []string{"evidence_kind", "schema", "body_digest", "canonical_body_or_null", "canonical_envelope_or_null"},
			NullablePaths: map[string]bool{
				"canonical_body_or_null":     true,
				"canonical_envelope_or_null": true,
			},
			NullablePrefixes: []string{"canonical_body_or_null.", "canonical_envelope_or_null.body."},
			MaximumBytes:     authorityV7CanonicalArtifactMaxBytes,
		})
		if itemErr != nil {
			return ErrInvalidArgument
		}
		kind, err := authorityV7JSONString(itemFields["evidence_kind"])
		if err != nil || kind != expected[index].EvidenceKind {
			return ErrInvalidArgument
		}
		schema, err := authorityV7JSONString(itemFields["schema"])
		if err != nil {
			return ErrInvalidArgument
		}
		digest, err := authorityV7JSONDigest(itemFields["body_digest"])
		if err != nil || index > 0 && authorityV7CompareEvidenceKeys(previousDigest, previousSchema, digest, schema) >= 0 {
			return ErrInvalidArgument
		}
		previousDigest, previousSchema = digest, schema
		bodyIsNull := bytes.Equal(itemFields["canonical_body_or_null"], []byte("null"))
		envelopeIsNull := bytes.Equal(itemFields["canonical_envelope_or_null"], []byte("null"))
		if bodyIsNull == envelopeIsNull {
			return ErrInvalidArgument
		}
		switch kind {
		case "database_immutable_body":
			bodyExpectation := expected[index].Body
			if bodyExpectation == nil || expected[index].Envelope != nil || bodyIsNull || !envelopeIsNull || schema != bodyExpectation.Schema || digest != bodyExpectation.ExpectedDigest {
				return ErrInvalidArgument
			}
			if _, err := verifyAuthorityV7CanonicalBody(itemFields["canonical_body_or_null"], *bodyExpectation); err != nil {
				return ErrInvalidArgument
			}
		case "external_signed_envelope":
			envelopeExpectation := expected[index].Envelope
			if envelopeExpectation == nil || expected[index].Body != nil || !bodyIsNull || envelopeIsNull || schema != envelopeExpectation.Body.Schema || digest != envelopeExpectation.Body.ExpectedDigest {
				return ErrInvalidArgument
			}
			var envelopeObject map[string]json.RawMessage
			if err := json.Unmarshal(itemFields["canonical_envelope_or_null"], &envelopeObject); err != nil {
				return ErrInvalidArgument
			}
			bodyJCS := envelopeObject["body"]
			if len(bodyJCS) == 0 || verifyAuthorityV7SignedEnvelope(bodyJCS, itemFields["canonical_envelope_or_null"], *envelopeExpectation) != nil {
				return ErrInvalidArgument
			}
		default:
			return ErrInvalidArgument
		}
	}
	return nil
}

func authorityV7JSONString(raw json.RawMessage) (string, error) {
	var value string
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", ErrInvalidArgument
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return "", ErrInvalidArgument
	}
	return value, nil
}

func authorityV7CanonicalTime(raw json.RawMessage) (time.Time, error) {
	text, err := authorityV7JSONString(raw)
	if err != nil || !strings.HasSuffix(text, "Z") {
		return time.Time{}, ErrInvalidArgument
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != text {
		return time.Time{}, ErrInvalidArgument
	}
	return parsed, nil
}

func validateAuthorityV7CanonicalWindowAt(bodyJCS []byte, startField string, maximum time.Duration, now time.Time) (time.Time, error) {
	if startField == "" || maximum <= 0 {
		return time.Time{}, ErrInvalidArgument
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(bodyJCS, &fields) != nil || fields == nil {
		return time.Time{}, ErrInvalidArgument
	}
	start, err := authorityV7CanonicalTime(fields[startField])
	if err != nil {
		return time.Time{}, ErrInvalidArgument
	}
	expiresAt, err := authorityV7CanonicalTime(fields["expires_at"])
	if err != nil || !expiresAt.After(start) || expiresAt.Sub(start) > maximum || now.Before(start) || !now.Before(expiresAt) {
		return time.Time{}, ErrInvalidArgument
	}
	return expiresAt, nil
}

func authorityV7JSONDigest(raw json.RawMessage) (contracts.Digest, error) {
	text, err := authorityV7JSONString(raw)
	if err != nil || !authorityV7CanonicalDigestText(text) {
		return contracts.Digest{}, ErrInvalidArgument
	}
	decoded, err := hex.DecodeString(text)
	if err != nil || len(decoded) != sha256.Size {
		return contracts.Digest{}, ErrInvalidArgument
	}
	var result contracts.Digest
	copy(result[:], decoded)
	return result, nil
}

func authorityV7CanonicalDigestText(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}

func authorityV7CanonicalPositiveDecimal(value string) bool {
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && parsed > 0 && strconv.FormatUint(parsed, 10) == value
}

func authorityV7CanonicalNonnegativeDecimal(value string) bool {
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && strconv.FormatUint(parsed, 10) == value
}

func authorityV7PrintableASCII(value string, minimum, maximum int) bool {
	if len(value) < minimum || len(value) > maximum {
		return false
	}
	for index := range value {
		if value[index] < 0x20 || value[index] > 0x7e {
			return false
		}
	}
	return true
}

func authorityV7CompareEvidenceKeys(leftDigest contracts.Digest, leftSchema string, rightDigest contracts.Digest, rightSchema string) int {
	if compared := bytes.Compare(leftDigest[:], rightDigest[:]); compared != 0 {
		return compared
	}
	return strings.Compare(leftSchema, rightSchema)
}

func cloneFreshRestoreImportApplication(value contracts.FreshRestoreImportApplicationV1) contracts.FreshRestoreImportApplicationV1 {
	result := value
	if value.StagingImportCapabilityRecoveryIntentDigestOrNull != nil {
		digest := *value.StagingImportCapabilityRecoveryIntentDigestOrNull
		result.StagingImportCapabilityRecoveryIntentDigestOrNull = &digest
	}
	if value.StagingImportCapabilityRecoveryApplicationDigestOrNull != nil {
		digest := *value.StagingImportCapabilityRecoveryApplicationDigestOrNull
		result.StagingImportCapabilityRecoveryApplicationDigestOrNull = &digest
	}
	return result
}
