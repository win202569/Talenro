package deviceauth

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	deviceauthv1 "talenro.local/platform/gen/go/talenro/deviceauth/v1"
	eventsv1 "talenro.local/platform/gen/go/talenro/events/v1"
	"talenro.local/platform/internal/apierrors"
	"talenro.local/platform/internal/contracts/events"
	"talenro.local/platform/internal/idempotency"
	"talenro.local/platform/internal/identity"
	"talenro.local/platform/internal/store"
)

const (
	privateRevocationBindingDomain = "deviceauth/revoke-request/v1"
	revokeDeviceOperation          = "revoke_device"
	deviceRevocationStatus         = 204
)

type lockedDeviceRevocation struct {
	authorization store.DeviceauthDeviceAuthorization
	device        store.DeviceauthDevice
	families      []store.DeviceauthDeviceTokenFamily
	refresh       []store.DeviceauthDeviceRefreshToken
}

func (authority *lockedDeviceRevocation) clear() {
	if authority == nil {
		return
	}
	clearDeviceRowSecrets(&authority.device)
	for index := range authority.families {
		clear(authority.families[index].AccessTokenHash)
	}
	clearDeviceRefreshRows(authority.refresh)
	*authority = lockedDeviceRevocation{}
}

// RevokeDevice atomically removes exactly one device authority graph and its bound identity sessions.
func (service *Service) RevokeDevice(ctx context.Context, command RevokeDeviceCommand) (resultErr error) {
	defer func() {
		if recover() != nil {
			resultErr = deviceDependencyUnavailable()
		}
	}()
	principalID, principalErr := canonicalDeviceauthUUID(string(command.AccountPrincipal))
	if !validDeviceauthService(service) || nilDeviceauthValue(ctx) || principalErr != nil || !validRevokeDeviceCommand(command) {
		return deviceMalformedRequest()
	}
	canonical, err := privateRevocationBinding(service, command)
	if err != nil {
		return deviceDependencyUnavailable()
	}
	defer clear(canonical)
	scope, err := idempotency.AuthenticatedScope(principalID, "device_authorization", revokeDeviceOperation)
	if err != nil {
		return deviceDependencyUnavailable()
	}
	operationContext, cancel := context.WithTimeout(ctx, service.security.RequestDeadline)
	defer cancel()
	now, ok := service.now()
	if !ok {
		return deviceDependencyUnavailable()
	}
	err = service.repository.WithinTransaction(operationContext, func(transactionContext context.Context, base Transaction) error {
		transaction, typed := base.(deviceTokenTransaction)
		if !typed || nilDeviceauthValue(transaction) {
			return deviceDependencyUnavailable()
		}
		record, outcome, beginErr := transaction.BeginIdempotency(
			transactionContext, scope, command.IdempotencyKey, canonical, now, now.Add(deviceRotationIdempotencyRetention),
		)
		if beginErr != nil {
			return deviceDependencyUnavailable()
		}
		switch outcome {
		case idempotency.Replay:
			if record.ResponseStatus() != deviceRevocationStatus {
				return deviceDependencyUnavailable()
			}
			body, owned := record.TakeResponseBody()
			defer clear(body)
			if !owned || len(body) != 0 {
				return deviceDependencyUnavailable()
			}
			return nil
		case idempotency.Conflict:
			return apierrors.New(apierrors.IdempotencyConflict, apierrors.ContactSupport)
		case idempotency.InProgress:
			return apierrors.NewRetryAfter(apierrors.StateConflict, apierrors.Retry, time.Second)
		case idempotency.Started:
		default:
			return deviceDependencyUnavailable()
		}
		locked, valid, lockErr := service.lockDeviceRevocationAuthority(transactionContext, transaction, principalID, command, now)
		if lockErr != nil {
			return lockErr
		}
		defer locked.clear()
		if !valid {
			return deviceAuthenticationFailed()
		}
		for index := range locked.families {
			family := locked.families[index]
			if _, revokeErr := transaction.RevokeDeviceRefreshTokens(transactionContext, store.RevokeDeviceRefreshTokensParams{
				FamilyID: family.ID, RevokedAt: sql.NullTime{Time: now, Valid: true},
			}); revokeErr != nil {
				return deviceDependencyUnavailable()
			}
			rows, revokeErr := transaction.RevokeDeviceTokenFamily(transactionContext, store.RevokeDeviceTokenFamilyParams{ID: family.ID, UpdatedAt: now})
			if revokeErr != nil || rows != 1 {
				return deviceDependencyUnavailable()
			}
		}
		rows, revokeErr := transaction.RevokeDeviceAuthorization(transactionContext, store.RevokeDeviceAuthorizationParams{ID: locked.authorization.ID, UpdatedAt: now})
		if revokeErr != nil || rows != 1 {
			return deviceDependencyUnavailable()
		}
		rows, revokeErr = transaction.RevokeDeviceRecord(transactionContext, store.RevokeDeviceRecordParams{ID: locked.device.ID, UpdatedAt: now})
		if revokeErr != nil || rows != 1 {
			return deviceDependencyUnavailable()
		}
		if err := service.identity.RevokeAuthorizationSessions(transactionContext, transaction.DBTX(), locked.authorization.ID, now); err != nil {
			return deviceDependencyUnavailable()
		}
		eventID, eventErr := service.randomUUID()
		if eventErr != nil {
			return deviceDependencyUnavailable()
		}
		version := locked.authorization.StateVersion + 1
		if version <= 0 {
			return deviceDependencyUnavailable()
		}
		payload, eventErr := events.MarshalPayload(events.DeviceAuthorizationChangedType, &deviceauthv1.DeviceAuthorizationChanged{
			PrincipalId: principalID.String(), DeviceId: locked.device.ID.String(), AuthorizationId: locked.authorization.ID.String(),
			State: "revoked", Version: uint64(version),
		})
		if eventErr != nil {
			return deviceDependencyUnavailable()
		}
		event := &eventsv1.EventEnvelope{EventId: eventID.String(), EventType: events.DeviceAuthorizationChangedType, OccurredAt: timestamppb.New(now),
			Producer: "deviceauth", AggregateType: "device_authorization", AggregateId: locked.authorization.ID.String(), AggregateVersion: uint64(version),
			IdempotencyKey: eventID.String(), Payload: payload}
		if transaction.AppendEvent(transactionContext, event) != nil ||
			transaction.CompleteIdempotency(transactionContext, record, deviceRevocationStatus, nil) != nil {
			return deviceDependencyUnavailable()
		}
		return nil
	})
	return err
}

func (service *Service) lockDeviceRevocationAuthority(
	ctx context.Context,
	transaction deviceTokenTransaction,
	principalID uuid.UUID,
	command RevokeDeviceCommand,
	now time.Time,
) (lockedDeviceRevocation, bool, error) {
	discovered, found, err := transaction.DiscoverDeviceAuthorization(ctx, command.DeviceID)
	if err != nil {
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	if !found || discovered.ID == uuid.Nil {
		return lockedDeviceRevocation{}, false, nil
	}
	discoveredFamilies, err := transaction.ListDeviceAuthorizationFamilies(ctx, discovered.ID)
	if err != nil {
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	if !stableDeviceFamilySet(discoveredFamilies, discovered.ID) {
		clearDeviceFamilies(discoveredFamilies)
		return lockedDeviceRevocation{}, false, nil
	}
	lockedRefresh := make([]store.DeviceauthDeviceRefreshToken, 0)
	for index := range discoveredFamilies {
		rows, lockErr := transaction.LockDeviceFamilyRefreshTokens(ctx, discoveredFamilies[index].ID)
		if lockErr != nil {
			clearDeviceFamilies(discoveredFamilies)
			clearDeviceRefreshRows(lockedRefresh)
			return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
		}
		lockedRefresh = append(lockedRefresh, rows...)
	}
	lockedFamilies := make([]store.DeviceauthDeviceTokenFamily, 0, len(discoveredFamilies))
	for index := range discoveredFamilies {
		family, familyFound, familyErr := transaction.GetDeviceTokenFamilyForUpdate(ctx, discoveredFamilies[index].ID)
		if familyErr != nil {
			clearDeviceFamilies(discoveredFamilies)
			clearDeviceFamilies(lockedFamilies)
			clearDeviceRefreshRows(lockedRefresh)
			return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
		}
		if !familyFound {
			clearDeviceFamilies(discoveredFamilies)
			clearDeviceFamilies(lockedFamilies)
			clearDeviceRefreshRows(lockedRefresh)
			return lockedDeviceRevocation{}, false, nil
		}
		lockedFamilies = append(lockedFamilies, family)
	}
	authorization, authorizationFound, err := transaction.GetDeviceAuthorizationForUpdate(ctx, discovered.ID)
	if err != nil {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	if !authorizationFound {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		return lockedDeviceRevocation{}, false, nil
	}
	device, deviceFound, err := transaction.GetDeviceForUpdate(ctx, command.DeviceID)
	if err != nil {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	if !deviceFound {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		return lockedDeviceRevocation{}, false, nil
	}
	identityAuthority, identityFound, err := service.identity.ValidateDeviceRevocation(ctx, transaction.DBTX(), identity.DeviceRevocationRequest{
		PrincipalID: command.AccountPrincipal, Reauthentication: command.Reauthentication,
	}, now)
	if err != nil {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		clearDeviceRowSecrets(&device)
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	freshFamilies, err := transaction.ListDeviceAuthorizationFamilies(ctx, discovered.ID)
	if err != nil {
		clearDeviceFamilies(discoveredFamilies)
		clearDeviceFamilies(lockedFamilies)
		clearDeviceRefreshRows(lockedRefresh)
		clearDeviceRowSecrets(&device)
		return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
	}
	freshRefresh := make([]store.DeviceauthDeviceRefreshToken, 0, len(lockedRefresh))
	for index := range lockedFamilies {
		rows, listErr := transaction.ListDeviceFamilyRefreshTokens(ctx, lockedFamilies[index].ID)
		if listErr != nil {
			clearDeviceFamilies(discoveredFamilies)
			clearDeviceFamilies(lockedFamilies)
			clearDeviceFamilies(freshFamilies)
			clearDeviceRefreshRows(lockedRefresh)
			clearDeviceRefreshRows(freshRefresh)
			clearDeviceRowSecrets(&device)
			return lockedDeviceRevocation{}, false, deviceDependencyUnavailable()
		}
		freshRefresh = append(freshRefresh, rows...)
	}
	sortDeviceRefreshRows(lockedRefresh)
	sortDeviceRefreshRows(freshRefresh)
	valid := identityFound && identityAuthority.PrincipalID() == command.AccountPrincipal &&
		identityAuthority.SessionID() == command.Reauthentication.SessionID && sameDeviceFamilySet(discoveredFamilies, lockedFamilies) &&
		sameDeviceFamilySet(lockedFamilies, freshFamilies) && sameDeviceRefreshSet(lockedRefresh, freshRefresh) &&
		validDeviceRevocationGraph(principalID, command.DeviceID, authorization, device, lockedFamilies, lockedRefresh)
	clearDeviceFamilies(discoveredFamilies)
	clearDeviceFamilies(freshFamilies)
	clearDeviceRefreshRows(freshRefresh)
	result := lockedDeviceRevocation{authorization: authorization, device: device, families: lockedFamilies, refresh: lockedRefresh}
	if !valid {
		result.clear()
		return lockedDeviceRevocation{}, false, nil
	}
	return result, true, nil
}

func stableDeviceFamilySet(families []store.DeviceauthDeviceTokenFamily, authorizationID uuid.UUID) bool {
	if len(families) == 0 || authorizationID == uuid.Nil {
		return false
	}
	for index := range families {
		if families[index].ID == uuid.Nil || families[index].AuthorizationID != authorizationID || families[index].State == "revoked" ||
			(index > 0 && bytes.Compare(families[index-1].ID[:], families[index].ID[:]) >= 0) {
			return false
		}
	}
	return true
}

func sameDeviceFamilySet(left, right []store.DeviceauthDeviceTokenFamily) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ID != right[index].ID || left[index].AuthorizationID != right[index].AuthorizationID || left[index].State != right[index].State ||
			left[index].StateVersion != right[index].StateVersion {
			return false
		}
	}
	return true
}

func sortDeviceRefreshRows(rows []store.DeviceauthDeviceRefreshToken) {
	for outer := 1; outer < len(rows); outer++ {
		for inner := outer; inner > 0 && bytes.Compare(rows[inner-1].TokenHash, rows[inner].TokenHash) > 0; inner-- {
			rows[inner-1], rows[inner] = rows[inner], rows[inner-1]
		}
	}
}

func clearDeviceFamilies(rows []store.DeviceauthDeviceTokenFamily) {
	for index := range rows {
		clear(rows[index].AccessTokenHash)
	}
}

func validRevokeDeviceCommand(command RevokeDeviceCommand) bool {
	if command.DeviceID == uuid.Nil || !validDeviceauthIdempotencyKey(command.IdempotencyKey) {
		return false
	}
	if _, err := canonicalDeviceauthUUID(string(command.AccountPrincipal)); err != nil {
		return false
	}
	if _, err := canonicalDeviceauthUUID(string(command.Reauthentication.SessionID)); err != nil {
		return false
	}
	proof := command.Reauthentication.Proof.Copy()
	defer clear(proof)
	switch command.Reauthentication.Method {
	case identity.ReauthPassword, identity.ReauthPasskey, identity.ReauthTOTP, identity.ReauthRecoveryCode:
	default:
		return false
	}
	return len(proof) >= 1 && len(proof) <= 4096
}

func privateRevocationBinding(service *Service, command RevokeDeviceCommand) ([]byte, error) {
	if service == nil || nilDeviceauthValue(service.protector) || !validRevokeDeviceCommand(command) {
		return nil, ErrInvalidApplication
	}
	principal := []byte(command.AccountPrincipal)
	session := []byte(command.Reauthentication.SessionID)
	method := []byte(command.Reauthentication.Method)
	proof := command.Reauthentication.Proof.Copy()
	defer clear(principal)
	defer clear(session)
	defer clear(method)
	defer clear(proof)
	material := make([]byte, 0, 160+len(principal)+len(session)+len(method)+len(proof))
	material = appendRegistrationFrame(material, principal)
	material = appendRegistrationFrame(material, command.DeviceID[:])
	material = appendRegistrationFrame(material, session)
	material = appendRegistrationFrame(material, method)
	material = appendRegistrationFrame(material, proof)
	defer clear(material)
	digest := service.protector.LookupDigest(privateRevocationBindingDomain, material)
	if digest == [32]byte{} {
		return nil, ErrInvalidApplication
	}
	result := append([]byte(nil), digest[:]...)
	clear(digest[:])
	return result, nil
}

func (lockedDeviceRevocation) Format(state fmt.State, _ rune) {
	redactDeviceauthValue(state, "lockedDeviceRevocation")
}

func (lockedDeviceRevocation) LogValue() slog.Value {
	return slog.StringValue("deviceauth.lockedDeviceRevocation([REDACTED])")
}

func (lockedDeviceRevocation) MarshalJSON() ([]byte, error) {
	return nil, errors.New("deviceauth: locked revocation serialization forbidden")
}

func validDeviceRevocationGraph(
	principalID, deviceID uuid.UUID,
	authorization store.DeviceauthDeviceAuthorization,
	device store.DeviceauthDevice,
	families []store.DeviceauthDeviceTokenFamily,
	refresh []store.DeviceauthDeviceRefreshToken,
) bool {
	if principalID == uuid.Nil || deviceID == uuid.Nil || authorization.ID == uuid.Nil || authorization.PrincipalID != principalID ||
		authorization.DeviceID != deviceID || authorization.State == "revoked" || device.ID != deviceID || device.PrincipalID != principalID ||
		device.State == "revoked" || len(families) == 0 {
		return false
	}
	familyIDs := make(map[uuid.UUID]struct{}, len(families))
	var previous uuid.UUID
	for index := range families {
		family := families[index]
		if family.ID == uuid.Nil || family.AuthorizationID != authorization.ID || (index > 0 && bytes.Compare(previous[:], family.ID[:]) >= 0) {
			return false
		}
		familyIDs[family.ID] = struct{}{}
		previous = family.ID
	}
	var previousHash []byte
	for index := range refresh {
		if _, ok := familyIDs[refresh[index].FamilyID]; !ok || len(refresh[index].TokenHash) != 32 {
			return false
		}
		if previousHash != nil && bytes.Compare(previousHash, refresh[index].TokenHash) >= 0 {
			return false
		}
		previousHash = refresh[index].TokenHash
	}
	return true
}
