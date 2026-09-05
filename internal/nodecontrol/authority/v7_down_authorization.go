package authority

import (
	"context"
	"sync/atomic"
	"time"

	"talenro.local/platform/internal/nodecontrol/contracts"
)

type VerifiedAuthorityV7DownAuthorization struct {
	state *verifiedAuthorityV7DownAuthorizationState
}

type authorityV7DownAuthorizationPersistenceView struct {
	Facts                    contracts.AuthorityV7DownAuthorizationFactsV1
	AuthorizationBodyJCS     []byte
	AuthorizationEnvelopeJCS []byte
}

type verifiedAuthorityV7DownAuthorizationState struct {
	view     authorityV7DownAuthorizationPersistenceView
	consumed atomic.Bool
}

type AuthorityProtocolDowngradeAuthorizer interface {
	AuthorizeAuthorityProtocolDowngrade(context.Context, contracts.AuthorityV7DownAuthorizationRequestV1) (VerifiedAuthorityV7DownAuthorization, error)
}

func newVerifiedAuthorityV7DownAuthorization(view authorityV7DownAuthorizationPersistenceView) (VerifiedAuthorityV7DownAuthorization, error) {
	cloned := cloneAuthorityV7DownAuthorizationPersistenceView(view)
	if err := cloned.Facts.Validate(); err != nil || !equalAuthorityV7Bytes(cloned.Facts.AuthorizationEnvelopeJCS, cloned.AuthorizationEnvelopeJCS) ||
		authorityV7DomainDigest("authority-protocol-downgrade-authorization.v1", cloned.AuthorizationBodyJCS) != cloned.Facts.AuthorizationDigest {
		return VerifiedAuthorityV7DownAuthorization{}, ErrInvalidArgument
	}
	return VerifiedAuthorityV7DownAuthorization{state: &verifiedAuthorityV7DownAuthorizationState{view: cloned}}, nil
}

func ConsumeVerifiedAuthorityV7DownAuthorization(authorization VerifiedAuthorityV7DownAuthorization) (authorityV7DownAuthorizationPersistenceView, error) {
	if authorization.state == nil {
		return authorityV7DownAuthorizationPersistenceView{}, ErrInvalidArgument
	}
	if !authorization.state.consumed.CompareAndSwap(false, true) {
		return authorityV7DownAuthorizationPersistenceView{}, ErrConflict
	}
	view := cloneAuthorityV7DownAuthorizationPersistenceView(authorization.state.view)
	now := time.Now()
	if validateAuthorityV7EffectiveWindowAt(view.Facts.Authorization.IssuedAt, view.Facts.Authorization.ExpiresAt, now) != nil ||
		view.Facts.Validate() != nil || !equalAuthorityV7Bytes(view.Facts.AuthorizationEnvelopeJCS, view.AuthorizationEnvelopeJCS) ||
		authorityV7DomainDigest("authority-protocol-downgrade-authorization.v1", view.AuthorizationBodyJCS) != view.Facts.AuthorizationDigest {
		return authorityV7DownAuthorizationPersistenceView{}, ErrInvalidArgument
	}
	return view, nil
}

func cloneAuthorityV7DownAuthorizationPersistenceView(value authorityV7DownAuthorizationPersistenceView) authorityV7DownAuthorizationPersistenceView {
	result := value
	result.Facts = cloneAuthorityV7DownAuthorizationFacts(value.Facts)
	result.AuthorizationBodyJCS = cloneAuthorityV7Bytes(value.AuthorizationBodyJCS)
	result.AuthorizationEnvelopeJCS = cloneAuthorityV7Bytes(value.AuthorizationEnvelopeJCS)
	return result
}
