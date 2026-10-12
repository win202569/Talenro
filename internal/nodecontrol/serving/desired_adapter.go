package serving

import (
	"context"
	"crypto/sha256"
	"github.com/google/uuid"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// Private staging only; the reader must supply the same callback DBTX and epoch.
func queryDesiredProjection(ctx context.Context, db store.DBTX, nodeID uuid.UUID, epoch uint64) (desiredProjection, error) {
	if ctx != nil && ctx.Err() != nil {
		return desiredProjection{}, authority.ErrCanceled
	}
	if ctx == nil || nilCapability(db) {
		return desiredProjection{}, authority.ErrAuthorityUnavailable
	}
	if !factsUUID(nodeID) || !factsCoordinate(epoch) {
		return desiredProjection{}, authority.ErrInvalidArgument
	}
	life := desiredQueryLifetime{db: desiredScanDB{DBTX: db}}
	defer life.cleanup()
	transcript := append([]byte("TALENRO-NODE-AUTHORITY-SCOPE-V1\x00"), nodeID[:]...)
	life.nodeScope = sha256.Sum256(transcript)
	clear(transcript)
	life.trustScope = sha256.Sum256([]byte("TALENRO-GLOBAL-NODE-TRUST-SCOPE-V1\x00global_node_trust"))
	rows, err := store.New(&life.db).ListExactDesiredProjection(ctx, store.ListExactDesiredProjectionParams{
		NodeID: nodeID, AuthorityEpoch: int64(epoch), NodeScopeDigest: life.nodeScope[:], TrustScopeDigest: life.trustScope[:],
	})
	life.rows = rows
	if ctx.Err() != nil || readCanceled(err) {
		return desiredProjection{}, authority.ErrCanceled
	}
	if err != nil {
		return desiredProjection{}, authority.ErrAuthorityUnavailable
	}
	life.staged, err = desiredProjectionFromRows(nodeID, epoch, rows)
	if ctx.Err() != nil {
		return desiredProjection{}, authority.ErrCanceled
	}
	if err != nil {
		return desiredProjection{}, err
	}
	return life.finish(ctx)
}

// Owns only decoded storage and the not-yet-delivered projection. It cannot
// erase driver buffers, unreachable runtime copies, or caller-owned fixtures.
type desiredQueryLifetime struct {
	db                    desiredScanDB
	rows                  []store.ListExactDesiredProjectionRow
	nodeScope, trustScope [32]byte
	staged                desiredProjection
}

func (l *desiredQueryLifetime) cleanup() {
	defer func() {
		clear(l.rows)
		l.rows = nil
		clear(l.nodeScope[:])
		clear(l.trustScope[:])
		discardDesiredProjection(&l.staged)
	}()
	l.db.cleanup()
}
func (l *desiredQueryLifetime) finish(ctx context.Context) (desiredProjection, error) {
	// The outer cleanup still owns staged if Close panics or cancellation wins.
	l.db.cleanup()
	if ctx.Err() != nil {
		return desiredProjection{}, authority.ErrCanceled
	}
	out := l.staged
	l.staged = desiredProjection{}
	return out, nil
}
