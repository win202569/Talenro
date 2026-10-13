package nodeissuermapping

import "context"

func ownClaims(ctx context.Context, in Claims) (out Claims, err error) {
	if ctx == nil {
		return Claims{}, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return Claims{}, e
	}
	out = in
	out.Catalog.Entries = nil
	out.Association.SerialBytes = nil
	done := false
	defer func() {
		if !done {
			clearClaims(&out)
		}
	}()
	out.Catalog.Entries = copySlice(in.Catalog.Entries)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Association.SerialBytes = copyRaw(in.Association.SerialBytes)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearClaims(in *Claims) {
	if in == nil {
		return
	}
	clear(in.Catalog.Entries)
	clear(in.Association.SerialBytes)
	*in = Claims{}
}
