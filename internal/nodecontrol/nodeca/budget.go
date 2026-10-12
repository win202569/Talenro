package nodeca

import "context"

func preflight(ctx context.Context, in Input) error {
	if ctx == nil {
		return ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(in.DER) > 65536 || len(in.DeclaredSPKI) > 8192 {
		return ErrLimitExceeded
	}
	if len(in.DER) == 0 || len(in.DeclaredSPKI) == 0 {
		return ErrInvalidInput
	}
	return nil
}
