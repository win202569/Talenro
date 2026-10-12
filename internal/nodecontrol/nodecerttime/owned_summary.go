package nodecerttime

import "talenro.local/platform/internal/nodecontrol/statetimebinding"

func ownTimeSummary(in statetimebinding.Summary) (out statetimebinding.Summary) {
	out = in
	out.Policy.Binding.Proposed.Target.Links = nil
	out.Policy.Binding.Proposed.Coverage = nil
	done := false
	defer func() {
		if !done {
			clearTimeSummary(&out)
		}
	}()
	out.Policy.Binding.Proposed.Target.Links = copySlice(in.Policy.Binding.Proposed.Target.Links)
	out.Policy.Binding.Proposed.Coverage = copySlice(in.Policy.Binding.Proposed.Coverage)
	done = true
	return out
}
func clearTimeSummary(in *statetimebinding.Summary) {
	if in == nil {
		return
	}
	clear(in.Policy.Binding.Proposed.Target.Links)
	clear(in.Policy.Binding.Proposed.Coverage)
	*in = statetimebinding.Summary{}
}
