// Package serving contains the internal certificate/desired read boundary.
// It does not authorize endpoints or expose a production reader factory yet.
package serving

import (
	"context"
	"errors"
	"reflect"

	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// BoundAuthorityReadSource is the indivisible capability specified by Task 10.
// Its production implementation must be created only by the Coordinator from
// its own PostgresRepository, never from a separately supplied pool/readiness
// pair. This interface alone is NOT proof that an implementation is trusted.
//
// The source synchronously calls query exactly once between readiness checks on
// one physical connection, without holding a transaction across provider calls.
// It must perform the post-check even on a domain error, and return nil only
// after both ready snapshots match exactly. Connection loss/replacement and
// identity/head mismatches fail closed. The callback/DBTX must not be retained.
// The source reports post-check cancellation/deadline via its return value;
// normal cleanup may cancel its scoped callback context after a successful read.
// Task 15's v7 incarnation, activation and lease gates remain mandatory before
// production serving; ordinary readiness is not a substitute.
type BoundAuthorityReadSource interface {
	WithConsistentReadyRead(context.Context, func(context.Context, store.DBTX) error) error
}

// guardedRead is the reader-side staging boundary. Future typed readers must
// query/validate into exclusively owned, defensive facts and provide a discard
// function that clears all temporary material. No constructor accepting an
// arbitrary source is exported. Physical readiness remains the source's job.
func guardedRead[T any](ctx context.Context, source BoundAuthorityReadSource, query func(context.Context, store.DBTX) (T, error), discard func(*T)) (T, error) {
	var zero T
	if ctx != nil && ctx.Err() != nil {
		return zero, authority.ErrCanceled
	}
	if ctx == nil || nilCapability(source) || query == nil || discard == nil {
		return zero, authority.ErrAuthorityUnavailable
	}
	var staged T
	var queryErr error
	called, queried, invalid, published := false, false, false, false
	defer func() {
		if queried && !published {
			discard(&staged)
		}
	}()
	sourceErr := source.WithConsistentReadyRead(ctx, func(boundCtx context.Context, db store.DBTX) error {
		if called {
			invalid = true
			return authority.ErrAuthorityUnavailable
		}
		called = true
		if ctx.Err() != nil || (boundCtx != nil && boundCtx.Err() != nil) {
			queryErr = authority.ErrCanceled
			return queryErr
		}
		if boundCtx == nil || nilCapability(db) {
			invalid = true
			return authority.ErrAuthorityUnavailable
		}
		queried = true
		staged, queryErr = query(boundCtx, db)
		// Hold domain errors until the source has completed its post-check.
		// Cancellation is the one query outcome allowed to stop that check.
		if readCanceled(queryErr) || ctx.Err() != nil || boundCtx.Err() != nil {
			queryErr = authority.ErrCanceled
			return queryErr
		}
		return nil
	})
	if ctx.Err() != nil || readCanceled(queryErr) || readCanceled(sourceErr) {
		return zero, authority.ErrCanceled
	}
	if sourceErr != nil || !called || invalid {
		return zero, authority.ErrAuthorityUnavailable
	}
	if queryErr != nil {
		// Return only finite, value-free classifications, never driver details.
		switch {
		case errors.Is(queryErr, authority.ErrAuthorityUnavailable):
			return zero, authority.ErrAuthorityUnavailable
		case errors.Is(queryErr, authority.ErrNotFound):
			return zero, authority.ErrNotFound
		case errors.Is(queryErr, authority.ErrConflict):
			return zero, authority.ErrConflict
		default:
			return zero, authority.ErrAuthorityUnavailable
		}
	}
	published = true
	return staged, nil
}

func readCanceled(err error) bool {
	return errors.Is(err, authority.ErrCanceled) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func nilCapability(value any) bool {
	if value == nil {
		return true
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflect.ValueOf(value).IsNil()
	default:
		return false
	}
}
