package nodeissuermapping

import "talenro.local/platform/internal/nodecontrol/nodeissuerbinding"

func ownEvidence(in nodeissuerbinding.Summary) (out nodeissuerbinding.Summary) {
	out = in
	out.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links = nil
	out.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearEvidence(&out)
		}
	}()
	out.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links = copySlice(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links)
	out.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage = copySlice(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage)
	done = true
	return out
}
func clearEvidence(in *nodeissuerbinding.Summary) {
	if in == nil {
		return
	}
	clear(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Target.Links)
	clear(in.Evidence.Evidence.Time.Policy.Binding.Proposed.Coverage)
	*in = nodeissuerbinding.Summary{}
}
