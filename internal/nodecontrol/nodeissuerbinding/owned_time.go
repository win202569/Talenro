package nodeissuerbinding

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetimebinding"
)

func ownTimeBinding(ctx context.Context, in statetimebinding.Input) (out statetimebinding.Input, err error) {
	out = in
	out.Policy = statepolicybinding.Input{}
	out.Time.AttestationEnvelope = nil
	out.Time.MetadataEnvelope = nil
	done := false
	defer func() {
		if !done {
			clearTimeBinding(&out)
		}
	}()
	if ctx == nil {
		return out, ErrInvalidInput
	}
	if e := ctx.Err(); e != nil {
		return out, e
	}
	out.Policy, err = ownPolicyBinding(ctx, in.Policy)
	if err != nil {
		return out, err
	}
	out.Time.AttestationEnvelope = copyRaw(in.Time.AttestationEnvelope)
	out.Time.MetadataEnvelope = copyRaw(in.Time.MetadataEnvelope)
	out.Time.Time.Now = normalizeTime(in.Time.Time.Now)
	out.Time.Time.Floor = normalizeTime(in.Time.Time.Floor)
	if e := ctx.Err(); e != nil {
		return out, e
	}
	done = true
	return out, nil
}
func clearTimeBinding(in *statetimebinding.Input) {
	if in == nil {
		return
	}
	clearPolicyBinding(&in.Policy)
	clear(in.Time.AttestationEnvelope)
	clear(in.Time.MetadataEnvelope)
	*in = statetimebinding.Input{}
}
