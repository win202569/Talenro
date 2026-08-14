package errorreport

import "context"

// LocalProvider is a deterministic local/test sink. It performs no I/O and
// never serializes report contents.
type LocalProvider struct{}

// NewLocalProvider creates the local/test reporter provider.
func NewLocalProvider() *LocalProvider { return new(LocalProvider) }

// Send accepts a validated batch without retaining it.
func (*LocalProvider) Send(ctx context.Context, reports []Report) error {
	if ctx == nil || ctx.Err() != nil || len(reports) == 0 {
		return ErrInvalidConfiguration
	}
	return nil
}
