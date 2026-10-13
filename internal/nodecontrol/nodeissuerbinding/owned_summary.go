package nodeissuerbinding

import "talenro.local/platform/internal/nodecontrol/nodeleafbinding"

func ownEvidence(in nodeleafbinding.Summary) (out nodeleafbinding.Summary) {
	out = in
	out.Evidence.Time.Policy.Binding.Proposed.Target.Links = nil
	out.Evidence.Time.Policy.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearEvidence(&out)
		}
	}()
	out.Evidence.Time.Policy.Binding.Proposed.Target.Links = copySlice(in.Evidence.Time.Policy.Binding.Proposed.Target.Links)
	out.Evidence.Time.Policy.Binding.Proposed.Coverage = copySlice(in.Evidence.Time.Policy.Binding.Proposed.Coverage)
	done = true
	return out
}
func clearEvidence(in *nodeleafbinding.Summary) {
	if in == nil {
		return
	}
	clear(in.Evidence.Time.Policy.Binding.Proposed.Target.Links)
	clear(in.Evidence.Time.Policy.Binding.Proposed.Coverage)
	*in = nodeleafbinding.Summary{}
}
