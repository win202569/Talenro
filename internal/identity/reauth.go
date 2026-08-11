package identity

import (
	"context"
	"encoding/binary"
	"math"
	"time"

	"github.com/google/uuid"

	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/sensitive"
	"talenro.local/platform/internal/store"
)

const strongAuthRequestDigestDomain = "identity_strong_auth_request_v1"

func (service *Service) loadPasswordReauthenticationCredential(
	ctx context.Context,
	transaction Transaction,
	principalID uuid.UUID,
) (store.IdentityPasswordCredential, error) {
	credential, found, err := transaction.GetPasswordCredential(ctx, principalID)
	if err != nil {
		return store.IdentityPasswordCredential{}, dependencyUnavailable()
	}
	if !found || credential.PrincipalID != principalID {
		if !service.performDummyPasswordWork() {
			return store.IdentityPasswordCredential{}, dependencyUnavailable()
		}
		return store.IdentityPasswordCredential{}, authenticationFailed()
	}
	return credential, nil
}

func (service *Service) verifyLockedPasswordReauthentication(
	ctx context.Context,
	transaction Transaction,
	principalID uuid.UUID,
	reauthentication Reauthentication,
	credentialRow store.IdentityPasswordCredential,
	now time.Time,
) (store.IdentityAccount, error) {
	sessionID, sessionOK := parseCanonicalIdentityUUID(string(reauthentication.SessionID))
	if !sessionOK || reauthentication.Method != ReauthPassword || !validPasswordSecret(reauthentication.Proof) {
		return store.IdentityAccount{}, malformedRequest()
	}
	session, sessionFound, err := transaction.GetAccountSessionForUpdate(ctx, store.GetAccountSessionForUpdateParams{
		ID: sessionID, PrincipalID: principalID,
	})
	if err != nil {
		return store.IdentityAccount{}, dependencyUnavailable()
	}
	account, accountFound, err := transaction.GetAccountForUpdate(ctx, principalID)
	if err != nil {
		return store.IdentityAccount{}, dependencyUnavailable()
	}
	if !sessionFound || !accountFound || session.ID != sessionID || session.PrincipalID != principalID || account.ID != principalID ||
		account.State != "active" || session.State != "active" || !session.AccessExpiresAt.After(now) || !session.AbsoluteExpiresAt.After(now) {
		if !service.performDummyPasswordWork() {
			return store.IdentityAccount{}, dependencyUnavailable()
		}
		return store.IdentityAccount{}, authenticationFailed()
	}
	credential, err := passwordCredentialFromStore(credentialRow)
	if err != nil {
		if !service.performDummyPasswordWork() {
			return store.IdentityAccount{}, dependencyUnavailable()
		}
		return store.IdentityAccount{}, dependencyUnavailable()
	}
	proof := reauthentication.Proof.Copy()
	defer clear(proof)
	matched, _ := VerifyPasswordWithDeriver(proof, credential, CurrentPasswordPolicy(), service.derive)
	if !matched {
		return store.IdentityAccount{}, authenticationFailed()
	}
	return account, nil
}

func strongAuthCanonicalRequest(
	protector sensitive.Protector,
	operation string,
	parts ...[]byte,
) ([]byte, error) {
	if nilIdentityValue(protector) || !validStrongAuthOperation(operation) || len(parts) == 0 || len(parts) > math.MaxUint32 || len(operation) > math.MaxUint32 {
		return nil, errStrongAuthDependency
	}
	materialLength := 8 + len(operation)
	for _, part := range parts {
		if len(part) > math.MaxUint32 || materialLength > math.MaxInt-4-len(part) {
			return nil, errStrongAuthDependency
		}
		materialLength += 4 + len(part)
	}
	material := make([]byte, 0, materialLength)
	var length [4]byte
	// #nosec G115 -- the operation and part counts are explicitly bounded above.
	binary.BigEndian.PutUint32(length[:], uint32(len(operation)))
	material = append(material, length[:]...)
	material = append(material, operation...)
	// #nosec G115 -- the part count is explicitly bounded above.
	binary.BigEndian.PutUint32(length[:], uint32(len(parts)))
	material = append(material, length[:]...)
	for _, part := range parts {
		// #nosec G115 -- each part length is explicitly bounded above.
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		material = append(material, length[:]...)
		material = append(material, part...)
	}
	clear(length[:])
	defer clear(material)
	digest := protector.LookupDigest(strongAuthRequestDigestDomain, material)
	if digest == [32]byte{} {
		return nil, errStrongAuthDependency
	}
	result := append([]byte(nil), digest[:]...)
	clear(digest[:])
	return result, nil
}

func validStrongAuthOperation(operation string) bool {
	switch operation {
	case "begin_totp_enrollment", "verify_totp_enrollment", "revoke_totp",
		"rotate_recovery_codes", "consume_recovery_code",
		"begin_passkey_registration", "finish_passkey_registration", "finish_passkey_authentication", "revoke_passkey":
		return true
	default:
		return false
	}
}

func strongAuthStateConflict() error {
	return apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
}
