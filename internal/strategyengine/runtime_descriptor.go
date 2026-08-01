package strategyengine

// RuntimeDescriptor is the typed, read-only a050 provider boundary. It carries
// no callbacks or mutable values; a050 may compose it after integration without
// giving a047 a control-plane dependency.
type RuntimeDescriptor struct {
	Category string
	Sections [4]RuntimeSection
	Fields   [12]RuntimeField
	Blockers [9]RuntimeBlocker
}
type RuntimeSection struct{ ID, Label, ActionOwner string }
type RuntimeField struct{ Key, Label, Help, Default, Desired, Effective, Unit, Range, Provenance, ApplyTiming string }
type RuntimeBlocker struct {
	Key, Label string
	Desired    RuntimeState
	Effective  RuntimeState
	Freshness  RuntimeState
	Reason     RuntimeRefusal
}

// RuntimeState is the closed vocabulary shared by the a047 authority snapshot
// and every read-only transport. A transport may style these values, but it may
// not reinterpret a boolean or calculate a replacement effective state.
type RuntimeState string

const (
	RuntimeStateOff           RuntimeState = "OFF"
	RuntimeStateOn            RuntimeState = "ON"
	RuntimeStateNotConfigured RuntimeState = "NOT_CONFIGURED"
	RuntimeStateUnwired       RuntimeState = "UNWIRED"
	RuntimeStateMissing       RuntimeState = "MISSING"
	RuntimeStateStale         RuntimeState = "STALE"
	RuntimeStateRefused       RuntimeState = "REFUSED"
	RuntimeStateVerified      RuntimeState = "VERIFIED"
	RuntimeStateUnapproved    RuntimeState = "UNAPPROVED"
	RuntimeStateUnknown       RuntimeState = "UNKNOWN"
	RuntimeStateReady         RuntimeState = "READY"
	RuntimeStateHealthy       RuntimeState = "HEALTHY"
	RuntimeStateLive          RuntimeState = "LIVE"
	RuntimeStateValid         RuntimeState = "VALID"
	RuntimeStateUnobserved    RuntimeState = "UNOBSERVED"
)

func (s RuntimeState) Valid() bool {
	switch s {
	case RuntimeStateOff, RuntimeStateOn, RuntimeStateNotConfigured, RuntimeStateUnwired,
		RuntimeStateMissing, RuntimeStateStale, RuntimeStateRefused, RuntimeStateVerified,
		RuntimeStateUnapproved, RuntimeStateUnknown, RuntimeStateReady, RuntimeStateHealthy,
		RuntimeStateLive, RuntimeStateValid, RuntimeStateUnobserved:
		return true
	default:
		return false
	}
}

// RuntimeRefusal is ordered by the authority which produced the snapshot. The
// console displays FirstRefusal verbatim from this closed set and never derives
// it by walking blocker booleans in a second, transport-owned order.
type RuntimeRefusal string

const (
	RuntimeRefusalNone                     RuntimeRefusal = ""
	RuntimeRefusalProtectionUnwired        RuntimeRefusal = "protection_unwired"
	RuntimeRefusalCandidateProvenance      RuntimeRefusal = "candidate_provenance_absent"
	RuntimeRefusalSchedulerClaim           RuntimeRefusal = "scheduler_claim_absent"
	RuntimeRefusalSourceManifest           RuntimeRefusal = "source_manifest_unavailable"
	RuntimeRefusalGuardianUnapproved       RuntimeRefusal = "guardian_unapproved"
	RuntimeRefusalReconciliationUnhealthy  RuntimeRefusal = "reconciliation_unhealthy"
	RuntimeRefusalOperatingModeNotLive     RuntimeRefusal = "operating_mode_not_live"
	RuntimeRefusalKillSwitch               RuntimeRefusal = "kill_switch"
	RuntimeRefusalActivationManifestAbsent RuntimeRefusal = "activation_manifest_absent"
	RuntimeRefusalActivationManifestExpiry RuntimeRefusal = "activation_manifest_expired"
	RuntimeRefusalReadFailed               RuntimeRefusal = "runtime_read_failed"
)

func (r RuntimeRefusal) Valid() bool {
	switch r {
	case RuntimeRefusalNone, RuntimeRefusalProtectionUnwired, RuntimeRefusalCandidateProvenance,
		RuntimeRefusalSchedulerClaim, RuntimeRefusalSourceManifest, RuntimeRefusalGuardianUnapproved,
		RuntimeRefusalReconciliationUnhealthy, RuntimeRefusalOperatingModeNotLive,
		RuntimeRefusalKillSwitch, RuntimeRefusalActivationManifestAbsent,
		RuntimeRefusalActivationManifestExpiry, RuntimeRefusalReadFailed:
		return true
	default:
		return false
	}
}

func DormantRuntimeDescriptor() RuntimeDescriptor {
	return RuntimeDescriptor{
		Category: "strategy-runtime",
		Sections: [4]RuntimeSection{{ID: "parameters", Label: "전략 파라미터", ActionOwner: "a050"}, {ID: "lane", Label: "lane 상태", ActionOwner: "a050"}, {ID: "autostart", Label: "자동 기동", ActionOwner: "a050"}, {ID: "live", Label: "LIVE 주문 승인", ActionOwner: "a050"}},
		Fields: [12]RuntimeField{
			field("min_vwap_slope_pct", "최소 VWAP 기울기", "닫힌 5분봉의 VWAP 기울기가 이 값보다 낮으면 진입을 거부한다.", "0.08", "%", "0 이상", "bar 평가 시"),
			field("ema_touch_tolerance_pct", "EMA9 접촉 허용", "저가가 EMA9에 충분히 가까운 눌림인지 판정하는 고정 허용폭이다.", "0.25", "%", "0 이상", "bar 평가 시"),
			field("min_forward_space_pct", "최소 LVN 전방 공간", "다음 거래량 저항대까지 확보해야 하는 최소 상승 여유다.", "1.2", "%", "0 이상", "bar 평가 시"),
			field("min_expected_rr", "최소 기대 RR", "구조상 기대 손익비가 이 값보다 낮으면 진입을 거부한다.", "1.5", "R", "0 이상", "bar 평가 시"),
			field("tangled_band_pct", "얽힘 band", "분리 score가 0.35 미만이면 얽힘으로 거부한다. score가 작을수록 분리가 부족하다.", "0.35", "%", "0 이상", "bar 평가 시"),
			field("max_band_expansion_rate", "최대 band 확장률", "변동성 band가 이 배수보다 넓으면 과도한 확장으로 진입을 거부한다.", "1.8", "배", "0 이상", "bar 평가 시"),
			field("hard_stop_pct", "하드 스톱", "진입 결정에 고정되는 최초 손절 거리다.", "0.7", "%", "고정", "결정 생성 시"),
			field("partial_take_profit_at_r", "목표", "진입 결정에 고정되는 목표 수익 배수다.", "3.0", "R", "고정", "결정 생성 시"),
			field("skip_open_minutes", "시초 제외", "정규장 개장 직후 이 시간 동안은 새 진입을 평가하지 않는다.", "10", "분", "고정", "세션 평가 시"),
			field("max_signal_age_seconds", "최대 신호 나이", "닫힌 봉 이후 신호가 유효한 최대 시간이다.", "15", "초", "0~15", "결정 생성 시"),
			field("max_entry_price_drift_pct", "최대 진입 괴리", "결정 가격에서 이 비율보다 멀어진 주문은 dispatch 전에 거부한다.", "0.20", "%", "0~0.20", "dispatch 전"),
			field("symbol_state_stale_seconds", "종목 상태 최대 나이", "HALT·LIMIT·MANAGED 권위 상태가 이 시간보다 오래되면 신규 진입을 거부한다.", "30", "초", "고정", "후보 평가 시"),
		},
		Blockers: [9]RuntimeBlocker{
			{Key: "source", Label: "StockOS source manifest", Desired: RuntimeStateVerified, Effective: RuntimeStateNotConfigured, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalSourceManifest},
			{Key: "candidate", Label: "a046 후보 provenance", Desired: RuntimeStateVerified, Effective: RuntimeStateMissing, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalCandidateProvenance},
			{Key: "scheduler", Label: "a048 scheduler/calendar", Desired: RuntimeStateVerified, Effective: RuntimeStateMissing, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalSchedulerClaim},
			{Key: "protection", Label: "a045 ProtectionReady", Desired: RuntimeStateReady, Effective: RuntimeStateUnwired, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalProtectionUnwired},
			{Key: "guardian", Label: "Guardian 승인", Desired: RuntimeStateVerified, Effective: RuntimeStateMissing, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalGuardianUnapproved},
			{Key: "reconciliation", Label: "Reconciliation health", Desired: RuntimeStateHealthy, Effective: RuntimeStateMissing, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalReconciliationUnhealthy},
			{Key: "operating-mode", Label: "Operating mode", Desired: RuntimeStateLive, Effective: RuntimeStateNotConfigured, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalOperatingModeNotLive},
			{Key: "kill-switch", Label: "Kill switch", Desired: RuntimeStateOff, Effective: RuntimeStateUnknown, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalKillSwitch},
			{Key: "activation-manifest", Label: "Activation manifest", Desired: RuntimeStateValid, Effective: RuntimeStateNotConfigured, Freshness: RuntimeStateUnobserved, Reason: RuntimeRefusalActivationManifestAbsent},
		},
	}
}
func field(key, label, help, value, unit, validRange, timing string) RuntimeField {
	return RuntimeField{Key: key, Label: label, Help: help, Default: value, Desired: value, Effective: "미구성", Unit: unit, Range: validRange, Provenance: "commit " + SourceCommit + " / source-set sha256:" + FrozenSourceSetDigest + " / lane " + LaneID + "@" + LaneVersion, ApplyTiming: timing}
}
