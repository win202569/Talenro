package serving

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"
	"talenro.local/platform/internal/nodecontrol/authority"
	"talenro.local/platform/internal/store"
)

// queryCertificateFacts is a private staging adapter, NOT an authorized reader.
// Its DBTX and expected epoch must eventually come from one bound ready read.
// It owns decoded Scan destinations, not pgx's internal network buffers, and
// makes no secure-erasure claim about immutable strings or runtime copies.
func queryCertificateFacts(ctx context.Context, db store.DBTX, lookup CertificateLookup, epoch uint64) (CertificateFacts, error) {
	if ctx != nil && ctx.Err() != nil {
		return CertificateFacts{}, authority.ErrCanceled
	}
	if ctx == nil || nilCapability(db) {
		return CertificateFacts{}, authority.ErrAuthorityUnavailable
	}
	p, err := certificateQueryArguments(lookup, epoch)
	if err != nil {
		return CertificateFacts{}, err
	}
	defer func() {
		clear(p.Identity.SerialBytes)
		clear(p.Identity.LeafDERDigest[:])
		clear(p.Identity.PublicKeyDigest[:])
		clear(p.NodeScopeDigest[:])
		clear(p.TrustScopeDigest[:])
	}()
	owned := &certificateScanDB{DBTX: db}
	defer owned.cleanup()
	rows, err := store.New(owned).ListExactCertificateProjection(ctx, store.ListExactCertificateProjectionParams{
		NodeScopeDigest: p.NodeScopeDigest[:], RequireAgentTrust: p.RequireAgentTrust, TrustDomain: p.TrustDomain,
		TrustScopeDigest: p.TrustScopeDigest[:], NodeID: p.Identity.NodeID, IssuerID: p.Identity.IssuerID,
		SerialBytes: p.Identity.SerialBytes, LeafDerDigest: p.Identity.LeafDERDigest[:], PublicKeyDigest: p.Identity.PublicKeyDigest[:], AuthorityEpoch: p.AuthorityEpoch,
	})
	defer clear(rows)
	if ctx.Err() != nil || readCanceled(err) {
		return CertificateFacts{}, authority.ErrCanceled
	}
	// A transport error is never trusted as a domain classification. Only the
	// converter below may produce NotFound/Conflict after a successful query.
	if err != nil {
		return CertificateFacts{}, authority.ErrAuthorityUnavailable
	}
	facts, err := certificateFactsFromRows(lookup, epoch, rows)
	if ctx.Err() != nil {
		discardCertificateFacts(&facts)
		return CertificateFacts{}, authority.ErrCanceled
	}
	return facts, err
}

// The generated :many method drops partial rows on Scan/iteration errors.
// Retain its destination pointers so even those partial decodes can be cleared.
type certificateScanDB struct {
	store.DBTX
	rows *certificateScanRows
}

func (d *certificateScanDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	r, err := d.DBTX.Query(ctx, sql, args...)
	if !nilCapability(r) {
		d.rows = &certificateScanRows{Rows: r}
	}
	if err != nil {
		return nil, err
	}
	if d.rows == nil {
		return nil, authority.ErrAuthorityUnavailable
	}
	return d.rows, nil
}
func (d *certificateScanDB) cleanup() {
	if d.rows == nil {
		return
	}
	defer func() {
		for _, target := range d.rows.targets {
			if b, ok := target.(*[]byte); ok {
				clear(*b)
			}
			reflect.ValueOf(target).Elem().SetZero()
		}
		clear(d.rows.targets)
		d.rows.targets = nil
	}()
	d.rows.Close()
}

type certificateScanRows struct {
	pgx.Rows
	targets []any
	closed  bool
}

func (r *certificateScanRows) Scan(dest ...any) error {
	r.targets = append(r.targets, dest...)
	return r.Rows.Scan(dest...)
}
func (r *certificateScanRows) Close() {
	if !r.closed {
		r.closed = true
		r.Rows.Close()
	}
}

// Used on cancellation now; the eventual guarded reader must also invoke this
// when its authority post-check rejects staged facts.
func discardCertificateFacts(f *CertificateFacts) {
	clear(f.snapshot.Identity.SerialBytes)
	clear(f.snapshot.LeafDER)
	if f.snapshot.ServerTrust != nil {
		*f.snapshot.ServerTrust = ServerTrustFacts{}
	}
	*f = CertificateFacts{}
}
