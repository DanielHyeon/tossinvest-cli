package strategyengine

// RuntimeDescriptor is the typed, read-only a050 provider boundary. It carries
// no callbacks or mutable values; a050 may compose it after integration without
// giving a047 a control-plane dependency.
type RuntimeDescriptor struct {
	Category string
	Sections [4]RuntimeSection
	Fields   [11]RuntimeField
	Blockers [4]RuntimeBlocker
}
type RuntimeSection struct{ ID, Label, ActionOwner string }
type RuntimeField struct{ Key, Label, Help, Default, Desired, Effective, Unit, Range, Provenance, ApplyTiming string }
type RuntimeBlocker struct{ Key, Label, Desired, Effective, Reason string }

func DormantRuntimeDescriptor() RuntimeDescriptor {
	return RuntimeDescriptor{
		Category: "strategy-runtime",
		Sections: [4]RuntimeSection{{ID: "parameters", Label: "전략 파라미터", ActionOwner: "a050"}, {ID: "lane", Label: "lane 상태", ActionOwner: "a050"}, {ID: "autostart", Label: "자동 기동", ActionOwner: "a050"}, {ID: "live", Label: "LIVE 주문 승인", ActionOwner: "a050"}},
		Fields: [11]RuntimeField{
			field("min_vwap_slope_pct", "최소 VWAP 기울기", "0.08", "%", "0 이상", "bar 평가 시"), field("ema_touch_tolerance_pct", "EMA9 접촉 허용", "0.25", "%", "0 이상", "bar 평가 시"), field("min_forward_space_pct", "최소 LVN 전방 공간", "1.2", "%", "0 이상", "bar 평가 시"), field("min_expected_rr", "최소 기대 RR", "1.5", "R", "0 이상", "bar 평가 시"), field("tangled_band_pct", "얽힘 band", "0.35", "%", "0 이상", "bar 평가 시"), field("max_band_expansion_rate", "최대 band 확장률", "1.8", "배", "0 이상", "bar 평가 시"), field("hard_stop_pct", "하드 스톱", "0.7", "%", "고정", "결정 생성 시"), field("partial_take_profit_at_r", "목표", "3.0", "R", "고정", "결정 생성 시"), field("skip_open_minutes", "시초 제외", "10", "분", "고정", "세션 평가 시"), field("max_signal_age_seconds", "최대 신호 나이", "15", "초", "0~15", "결정 생성 시"), field("max_entry_price_drift_pct", "최대 진입 괴리", "0.20", "%", "0~0.20", "dispatch 전"),
		},
		Blockers: [4]RuntimeBlocker{{Key: "protection", Label: "a045 ProtectionReady", Desired: "WIRED", Effective: "UNWIRED", Reason: "protection_unwired"}, {Key: "candidate", Label: "a046 후보 provenance", Desired: "VERIFIED", Effective: "READ_ONLY", Reason: "activation_evidence_absent"}, {Key: "scheduler", Label: "a048 scheduler/calendar", Desired: "VERIFIED", Effective: "READ_ONLY", Reason: "activation_claim_absent"}, {Key: "source", Label: "StockOS source manifest", Desired: "VERIFIED", Effective: "NOT_CONFIGURED", Reason: "source_manifest_unavailable"}},
	}
}
func field(key, label, value, unit, validRange, timing string) RuntimeField {
	return RuntimeField{Key: key, Label: label, Help: "서버가 고정한 conservative v1 값", Default: value, Desired: value, Effective: "미구성", Unit: unit, Range: validRange, Provenance: SourceCommit + " / " + FrozenSourceSetDigest, ApplyTiming: timing}
}
