package nodeleafbinding

import "talenro.local/platform/internal/nodecontrol/nodecerttime"

func ownEvidence(in nodecerttime.Summary) (out nodecerttime.Summary) {
	out = in
	out.Time.Policy.Binding.Proposed.Target.Links = nil
	out.Time.Policy.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearEvidence(&out)
		}
	}()
	out.Time.Policy.Binding.Proposed.Target.Links = copySlice(in.Time.Policy.Binding.Proposed.Target.Links)
	out.Time.Policy.Binding.Proposed.Coverage = copySlice(in.Time.Policy.Binding.Proposed.Coverage)
	done = true
	return out
}
func clearEvidence(in *nodecerttime.Summary) {
	if in == nil {
		return
	}
	clear(in.Time.Policy.Binding.Proposed.Target.Links)
	clear(in.Time.Policy.Binding.Proposed.Coverage)
	*in = nodecerttime.Summary{}
}
