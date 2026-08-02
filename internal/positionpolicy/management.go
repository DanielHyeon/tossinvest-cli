package positionpolicy

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// RuntimeReader is the deliberately narrow, read-only runtime contract used by
// cross-process read adapters. It cannot express lifecycle or reconcile
// mutation and is separate from the position-policy command client.
type RuntimeReader interface {
	Runtime(context.Context) (ManagementRuntime, error)
}

// AdoptionBlockingTrackerProjection identifies the deliberately narrow block
// list exposed by the engine. It is the same runtime projection consulted by
// adoption; it is not a complete journal reconciliation ledger.
const AdoptionBlockingTrackerProjection = "adoption-blocking tracker projection"

// AdoptionSettings is a transport-neutral copy of the adoption settings an
// engine instance actually loaded. It carries no file or mutation capability.
type AdoptionSettings struct {
	Enabled        bool     `json:"enabled"`
	DefaultStopPct float64  `json:"default_stop_pct"`
	IncludeSymbols []string `json:"include_symbols"`
	ExcludeSymbols []string `json:"exclude_symbols"`
	Rejected       string   `json:"rejected,omitempty"`
}

func NewAdoptionSettings(enabled bool, stopPct float64, include, exclude []string,
	rejected string) AdoptionSettings {
	return AdoptionSettings{
		Enabled: enabled, DefaultStopPct: stopPct,
		IncludeSymbols: normalizeSymbols(include), ExcludeSymbols: normalizeSymbols(exclude),
		Rejected: sanitizeText(rejected),
	}
}

func normalizeSymbols(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	out := make([]string, 0, len(symbols))
	for _, raw := range symbols {
		symbol := strings.ToUpper(strings.TrimSpace(raw))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		out = append(out, symbol)
	}
	sort.Strings(out)
	return out
}

func (a AdoptionSettings) Included(symbol string) bool {
	return containsSymbol(a.IncludeSymbols, symbol)
}

func (a AdoptionSettings) Excluded(symbol string) bool {
	return containsSymbol(a.ExcludeSymbols, symbol)
}

func containsSymbol(symbols []string, raw string) bool {
	want := strings.ToUpper(strings.TrimSpace(raw))
	for _, symbol := range symbols {
		if strings.EqualFold(strings.TrimSpace(symbol), want) {
			return true
		}
	}
	return false
}

type ReconcileScope string

const (
	ScopeAccount ReconcileScope = "account"
	ScopeMarket  ReconcileScope = "market"
	ScopeSymbol  ReconcileScope = "symbol"
)

// ReconcileBlock intentionally omits an account identifier. The private engine
// control plane is single-account, so account scope is sufficient and avoids
// transporting a raw account reference to UI/read adapters.
type ReconcileBlock struct {
	Scope     ReconcileScope `json:"scope"`
	Market    string         `json:"market,omitempty"`
	Symbol    string         `json:"symbol,omitempty"`
	Reason    string         `json:"reason"`
	Detail    string         `json:"detail,omitempty"`
	StartedAt time.Time      `json:"started_at,omitempty"`
	Permanent bool           `json:"permanent"`
}

func NewReconcileBlock(scope ReconcileScope, market, symbol, reason, detail string,
	startedAt time.Time, permanent bool) ReconcileBlock {
	return ReconcileBlock{
		Scope: scope, Market: strings.ToLower(strings.TrimSpace(market)),
		Symbol: strings.ToUpper(strings.TrimSpace(symbol)), Reason: sanitizeText(reason),
		Detail: sanitizeText(detail), StartedAt: startedAt.UTC(), Permanent: permanent,
	}
}

func (b ReconcileBlock) Covers(market, symbol string) bool {
	switch b.Scope {
	case ScopeAccount:
		return true
	case ScopeMarket:
		return strings.EqualFold(strings.TrimSpace(b.Market), strings.TrimSpace(market))
	case ScopeSymbol:
		return strings.EqualFold(strings.TrimSpace(b.Symbol), strings.TrimSpace(symbol))
	default:
		// A future/invalid scope must not make a potentially covering safety block
		// disappear in an older reader.
		return true
	}
}

type ManagementRuntime struct {
	Effective      AdoptionSettings `json:"effective"`
	EffectiveKnown bool             `json:"effective_known"`
	Blocks         []ReconcileBlock `json:"blocks"`
	BlockSource    string           `json:"block_source"`
}

type ManagementStatus string

const (
	ManagementStatusUnknown          ManagementStatus = "UNKNOWN"
	ManagementStatusManaged          ManagementStatus = "MANAGED"
	ManagementStatusExcluded         ManagementStatus = "EXCLUDED"
	ManagementStatusReconcileBlocked ManagementStatus = "RECONCILE_BLOCKED"
	ManagementStatusAdoptionPending  ManagementStatus = "ADOPTION_PENDING"
	ManagementStatusUnmanaged        ManagementStatus = "UNMANAGED"
)

type ManagementReason string

const (
	ManagementReasonJournalUnavailable ManagementReason = "JOURNAL_UNAVAILABLE"
	ManagementReasonRuntimeUnavailable ManagementReason = "RUNTIME_UNAVAILABLE"
	ManagementReasonManaged            ManagementReason = "MANAGED"
	ManagementReasonOperatorReleased   ManagementReason = "OPERATOR_RELEASED"
	ManagementReasonExcluded           ManagementReason = "EXCLUDED"
	ManagementReasonReconcileBlock     ManagementReason = "RECONCILE_BLOCK"
	ManagementReasonAdoptionCandidate  ManagementReason = "ADOPTION_CANDIDATE"
	ManagementReasonNotSelected        ManagementReason = "NOT_SELECTED"
)

type ManagementInput struct {
	Market       string
	Symbol       string
	JournalKnown bool
	Managed      bool
	Released     bool
	Runtime      ManagementRuntime
}

type ManagementProjection struct {
	Status           ManagementStatus `json:"status"`
	StatusKnown      bool             `json:"status_known"`
	Label            string           `json:"label"`
	Reason           ManagementReason `json:"reason"`
	Included         bool             `json:"included"`
	Excluded         bool             `json:"excluded"`
	Candidate        bool             `json:"candidate"`
	DesignationKnown bool             `json:"designation_known"`
	Block            *ReconcileBlock  `json:"block,omitempty"`
}

// ProjectManagement is the sole priority table shared by web and API rows.
// Managed journal evidence remains authoritative when runtime reads fail, but
// no desired setting is allowed to masquerade as running effective state.
func ProjectManagement(input ManagementInput) ManagementProjection {
	if !input.JournalKnown {
		return ManagementProjection{Status: ManagementStatusUnknown,
			Label: "관리 여부 불명", Reason: ManagementReasonJournalUnavailable}
	}
	if input.Managed {
		out := ManagementProjection{Status: ManagementStatusManaged, StatusKnown: true,
			Label: "엔진 관리", Reason: ManagementReasonManaged}
		stampDesignation(&out, input)
		return out
	}
	if input.Released {
		out := ManagementProjection{Status: ManagementStatusUnmanaged, StatusKnown: true,
			Label: "관리 외(운영자 해제)", Reason: ManagementReasonOperatorReleased}
		stampDesignation(&out, input)
		return out
	}
	if !input.Runtime.EffectiveKnown {
		return ManagementProjection{Status: ManagementStatusUnknown,
			Label: "관리 여부 불명", Reason: ManagementReasonRuntimeUnavailable}
	}

	out := ManagementProjection{StatusKnown: true, DesignationKnown: true}
	stampDesignation(&out, input)
	if out.Excluded {
		out.Status, out.Label, out.Reason = ManagementStatusExcluded, "관리 제외", ManagementReasonExcluded
		return out
	}
	if out.Candidate {
		for index := range input.Runtime.Blocks {
			block := input.Runtime.Blocks[index]
			if block.Covers(input.Market, input.Symbol) {
				out.Status, out.Label, out.Reason = ManagementStatusReconcileBlocked,
					"관리 편입 · 대사 차단으로 대기", ManagementReasonReconcileBlock
				block.Detail = sanitizeText(block.Detail)
				block.Reason = sanitizeText(block.Reason)
				out.Block = &block
				return out
			}
		}
		out.Status, out.Label, out.Reason = ManagementStatusAdoptionPending,
			"관리 편입 · 편입 예약됨", ManagementReasonAdoptionCandidate
		return out
	}
	out.Status, out.Label, out.Reason = ManagementStatusUnmanaged,
		"관리 외(미편입)", ManagementReasonNotSelected
	return out
}

func stampDesignation(out *ManagementProjection, input ManagementInput) {
	if !input.Runtime.EffectiveKnown {
		return
	}
	out.DesignationKnown = true
	out.Included = input.Runtime.Effective.Included(input.Symbol)
	out.Excluded = input.Runtime.Effective.Excluded(input.Symbol)
	out.Candidate = !out.Excluded && (input.Runtime.Effective.Enabled || out.Included)
}

func sanitizeText(value string) string {
	value = strings.Join(strings.Fields(strings.ToValidUTF8(value, "�")), " ")
	const maxRunes = 240
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes])
}
