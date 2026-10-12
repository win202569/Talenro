package serving

import (
	"context"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// Private structural staging only: this is not an authorized production reader.
func readDesiredProjection(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID) (desiredProjection, error) {
	return readDesiredProjectionWithEpoch(ctx, source, nodeID, authority.BoundReadEpoch)
}

// The resolver seam is private and test-only in purpose. The real entry fixes
// it to BoundReadEpoch; neither it nor an arbitrary source is public configuration.
func readDesiredProjectionWithEpoch(ctx context.Context, source BoundAuthorityReadSource, nodeID uuid.UUID, resolve func(store.DBTX) (uint64, error)) (desiredProjection, error) {
	if ctx != nil && ctx.Err() != nil {
		return desiredProjection{}, authority.ErrCanceled
	}
	if ctx == nil {
		return desiredProjection{}, authority.ErrAuthorityUnavailable
	}
	if !factsUUID(nodeID) {
		return desiredProjection{}, authority.ErrInvalidArgument
	}
	return guardedRead(ctx, source, func(boundCtx context.Context, db store.DBTX) (desiredProjection, error) {
		if resolve == nil {
			return desiredProjection{}, authority.ErrAuthorityUnavailable
		}
		// Resolve against the exact callback capability before the adapter wraps it.
		epoch, err := resolve(db)
		if readCanceled(err) {
			return desiredProjection{}, authority.ErrCanceled
		}
		if err != nil || !factsCoordinate(epoch) {
			return desiredProjection{}, authority.ErrAuthorityUnavailable
		}
		return queryDesiredProjection(boundCtx, db, nodeID, epoch)
	}, discardDesiredProjection)
}
