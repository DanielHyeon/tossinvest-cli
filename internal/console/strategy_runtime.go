package console

import (
	"context"
	"net/http"
)

type StrategyRuntimeReading struct {
	LaneDesired, AutoStartDesired, GateApproved, LiveApproved                    bool
	ProtectionWired, CandidateProvenance, SchedulerClaim, SourceManifestVerified bool
	KillSwitch                                                                   bool
	Reason                                                                       string
}
type StrategyRuntimeReader interface {
	Read(context.Context) (StrategyRuntimeReading, error)
}
type strategyRuntimePage struct {
	Nav                                                                                                                      string
	LoadErr, Unwired                                                                                                         bool
	LaneDesired, LaneEffective, AutoStartDesired, AutoStartEffective, GateDesired, GateEffective, LiveDesired, LiveEffective string
	Protection, Candidate, Scheduler, SourceManifest, Reason                                                                 string
}

func (strategyRuntimePage) Refresh() bool { return false }
func (c *Console) handleStrategyRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다", "전략 상태는 GET/HEAD만 허용한다. 아무것도 전송되지 않았다.")
		return
	}
	reading := StrategyRuntimeReading{Reason: "source_manifest_unavailable"}
	page := strategyRuntimePage{Nav: "optimization", Unwired: c.opts.StrategyRuntime == nil}
	if c.opts.StrategyRuntime != nil {
		value, err := c.opts.StrategyRuntime.Read(r.Context())
		if err != nil {
			page.LoadErr = true
		} else {
			reading = value
		}
	}
	page.LaneDesired = onOff(reading.LaneDesired)
	page.AutoStartDesired = onOff(reading.AutoStartDesired)
	page.GateDesired = onOff(reading.GateApproved)
	page.LiveDesired = onOff(reading.LiveApproved)
	page.Protection = verified(reading.ProtectionWired, "UNWIRED")
	page.Candidate = verified(reading.CandidateProvenance, "READ_ONLY")
	page.Scheduler = verified(reading.SchedulerClaim, "READ_ONLY")
	page.SourceManifest = verified(reading.SourceManifestVerified, "NOT_CONFIGURED")
	effective := reading.LaneDesired && reading.AutoStartDesired && reading.GateApproved && reading.LiveApproved && reading.ProtectionWired && reading.CandidateProvenance && reading.SchedulerClaim && reading.SourceManifestVerified && !reading.KillSwitch
	page.LaneEffective = onOff(effective)
	page.AutoStartEffective = onOff(effective)
	page.GateEffective = onOff(reading.GateApproved && !reading.KillSwitch)
	page.LiveEffective = onOff(effective)
	page.Reason = strategyReason(reading.Reason, effective)
	c.render(w, "strategy-runtime", page)
}
func verified(ok bool, closed string) string {
	if ok {
		return "VERIFIED"
	}
	return closed
}
func strategyReason(raw string, effective bool) string {
	if effective {
		return "entry_permitted"
	}
	switch raw {
	case "protection_unwired", "candidate_provenance_absent", "scheduler_claim_absent", "source_manifest_unavailable", "kill_switch", "activation_manifest_absent":
		return raw
	default:
		return "activation_manifest_absent"
	}
}
