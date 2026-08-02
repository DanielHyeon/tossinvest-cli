package operatorview

import (
	"math"
	"strconv"
	"strings"
)

// ExitLineReferenceKind identifies non-actionable context shown beside the
// canonical ExitLine. A reference can explain saved ledger facts or a running
// adoption policy, but it never becomes current protection by itself.
type ExitLineReferenceKind string

const (
	ExitLineReferenceLegacyRaw          ExitLineReferenceKind = "LEGACY_RAW"
	ExitLineReferenceAdoptionPlan       ExitLineReferenceKind = "ADOPTION_PLAN"
	ExitLineReferenceGenerationMismatch ExitLineReferenceKind = "GENERATION_MISMATCH"
	ExitLineReferenceRuntimeUnknown     ExitLineReferenceKind = "RUNTIME_UNKNOWN"
	ExitLineReferenceLifecycleUnknown   ExitLineReferenceKind = "LIFECYCLE_UNKNOWN"
)

// StoredExitReference is the raw t0/progress evidence retained by legacy
// journals. Values remain strings so the view never rounds or recomputes a
// broker-native decimal.
type StoredExitReference struct {
	EntryPrice          string
	InitialStop         string
	Baseline            string
	HighWater           string
	LifecycleGeneration int64
}

// ExitLineReferenceSource contains already-selected evidence. In particular,
// EffectiveStopPct must come from the running engine snapshot, not desired
// configuration, broker average price, or the latest quote.
type ExitLineReferenceSource struct {
	Market                      string
	CanonicalSnapshotPresent    bool
	CanonicalSnapshotStale      bool
	EvidencePresent             bool
	EvidenceLifecycleGeneration int64
	Raw                         *StoredExitReference
	LifecycleKnown              bool
	LifecycleProofRequired      bool
	CurrentLifecycleGeneration  int64
	ManagementStatus            string
	EffectiveSettingsKnown      bool
	EffectiveStopPct            float64
	Released                    bool
	UnknownReason               string
}

// ExitLineReferenceView is deliberately non-actionable. EffectiveKnown is
// always false: even an ADOPTION_PLAN carries only the running percentage and
// cannot claim a price until adoption persists a matching lifecycle baseline.
type ExitLineReferenceView struct {
	Kind           ExitLineReferenceKind
	Label          string
	EffectiveKnown bool
	Market         string
	Currency       string
	EntryPrice     string
	InitialStop    string
	Baseline       string
	HighWater      string
	StopPercent    string
	Basis          string
	Reason         string
}

func (v ExitLineReferenceView) Present() bool { return v.Kind != "" }
func (v ExitLineReferenceView) LegacyRaw() bool {
	return v.Kind == ExitLineReferenceLegacyRaw
}
func (v ExitLineReferenceView) AdoptionPlan() bool {
	return v.Kind == ExitLineReferenceAdoptionPlan
}
func (v ExitLineReferenceView) GenerationMismatch() bool {
	return v.Kind == ExitLineReferenceGenerationMismatch
}
func (v ExitLineReferenceView) RuntimeUnknown() bool {
	return v.Kind == ExitLineReferenceRuntimeUnknown
}
func (v ExitLineReferenceView) LifecycleUnknown() bool {
	return v.Kind == ExitLineReferenceLifecycleUnknown
}

// BuildExitLineReference applies the shared console/API authority matrix.
// Canonical snapshots stay solely in ExitLine; stale canonical evidence is not
// replaced with raw values. Cross-lifecycle evidence is hidden completely.
func BuildExitLineReference(source ExitLineReferenceSource) ExitLineReferenceView {
	market := strings.ToUpper(strings.TrimSpace(source.Market))
	currency := marketCurrency(market)
	evidenceGeneration := source.EvidenceLifecycleGeneration
	if source.Raw != nil && evidenceGeneration == 0 {
		evidenceGeneration = source.Raw.LifecycleGeneration
	}
	evidencePresent := source.EvidencePresent || source.CanonicalSnapshotPresent || hasRawReference(source.Raw)
	if evidencePresent && source.LifecycleProofRequired && !source.LifecycleKnown {
		return emptyExitLineReference(ExitLineReferenceLifecycleUnknown, market, currency,
			"관리 세대 확인 불가", "현재 관리 세대를 확인할 수 없어 저장된 가격 근거를 표시하지 않는다")
	}
	if evidencePresent && source.LifecycleKnown && evidenceGeneration > 0 &&
		source.CurrentLifecycleGeneration > 0 && evidenceGeneration != source.CurrentLifecycleGeneration {
		return emptyExitLineReference(ExitLineReferenceGenerationMismatch, market, currency,
			"기준선 근거 세대 불일치", "현재 관리 세대와 저장된 기준선 세대가 달라 가격 근거를 표시하지 않는다")
	}

	if hasRawReference(source.Raw) && (source.Released || !source.CanonicalSnapshotPresent) {
		if !source.CanonicalSnapshotPresent && !rawReferenceAllowed(source.UnknownReason) {
			return ExitLineReferenceView{}
		}
		view := emptyExitLineReference(ExitLineReferenceLegacyRaw, market, currency,
			"저장된 원장 기준선 · 현재 실효 미확인", "현재 실효선 아님")
		if source.Released {
			view.Label = "저장된 과거 원장 기준선 · 관리 해제됨"
		}
		view.EntryPrice = dashIfEmpty(source.Raw.EntryPrice)
		view.InitialStop = dashIfEmpty(source.Raw.InitialStop)
		view.Baseline = dashIfEmpty(source.Raw.Baseline)
		view.HighWater = dashIfEmpty(source.Raw.HighWater)
		return view
	}

	// A canonical snapshot, including a stale one, owns its own fail-closed
	// rendering. Promoting the raw tuple here would bypass its freshness verdict.
	if source.CanonicalSnapshotPresent {
		return ExitLineReferenceView{}
	}

	status := strings.ToUpper(strings.TrimSpace(source.ManagementStatus))
	if (status == "ADOPTION_PENDING" || status == "RECONCILE_BLOCKED") &&
		source.EffectiveSettingsKnown && validReferenceStop(source.EffectiveStopPct) {
		view := emptyExitLineReference(ExitLineReferenceAdoptionPlan, market, currency,
			"기준선 미생성 · 엔진 보호 미적용", "가격은 편입 시 확정")
		view.StopPercent = strconv.FormatFloat(source.EffectiveStopPct*100, 'f', -1, 64) + "%"
		return view
	}
	if status == "UNKNOWN" && !source.EffectiveSettingsKnown {
		return emptyExitLineReference(ExitLineReferenceRuntimeUnknown, market, currency,
			"기준선·정책 폭 알 수 없음", "실행 중 엔진 설정 또는 관리 상태를 확인할 수 없음")
	}
	return ExitLineReferenceView{}
}

func rawReferenceAllowed(reason string) bool {
	switch strings.TrimSpace(reason) {
	case "legacy_snapshot_absent", "not_evaluated_yet":
		return true
	default:
		return false
	}
}

func emptyExitLineReference(kind ExitLineReferenceKind, market, currency, label, reason string) ExitLineReferenceView {
	return ExitLineReferenceView{
		Kind: kind, Label: label, Market: market, Currency: currency,
		EntryPrice: dash, InitialStop: dash, Baseline: dash, HighWater: dash,
		Basis: reason, Reason: reason,
	}
}

func hasRawReference(raw *StoredExitReference) bool {
	return raw != nil && (strings.TrimSpace(raw.EntryPrice) != "" ||
		strings.TrimSpace(raw.InitialStop) != "" || strings.TrimSpace(raw.Baseline) != "" ||
		strings.TrimSpace(raw.HighWater) != "")
}

func validReferenceStop(value float64) bool {
	return value > 0 && value < 1 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func marketCurrency(market string) string {
	switch market {
	case "KR", "KRX", "KOSPI", "KOSDAQ":
		return "KRW"
	case "US", "NASDAQ", "NYSE", "AMEX":
		return "USD"
	default:
		return dash
	}
}
