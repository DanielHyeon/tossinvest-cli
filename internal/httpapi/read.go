package httpapi

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

var readResourceNames = [...]string{
	"engine", "positions", "orders", "candidates", "performance", "settings", "optimization",
}

func ReadResourceNames() []string {
	return append([]string(nil), readResourceNames[:]...)
}

// Reader is the complete read authority accepted by the API router. Its return
// values contain no callable capability. In particular, the interface cannot
// submit/cancel/amend orders or write a journal, gate, lane, kill switch, or
// activation manifest.
type Reader interface {
	Engine(context.Context) (EngineResource, error)
	Positions(context.Context) (PositionsResource, error)
	Orders(context.Context) (OrdersResource, error)
	Candidates(context.Context) (CandidatesResource, error)
	Performance(context.Context) (performance.DashboardView, error)
	Settings(context.Context) (SettingsResource, error)
	Optimization(context.Context) (optimization.View, error)
}

type PerformanceResource struct {
	Query          PerformanceQuery       `json:"query"`
	NewestSourceAt *time.Time             `json:"newestSourceAt"`
	Aggregates     []PerformanceAggregate `json:"aggregates"`
	States         PerformanceStates      `json:"states"`
}

type PerformanceQuery struct {
	AsOf          time.Time `json:"asOf"`
	PeriodDays    int       `json:"periodDays"`
	Market        string    `json:"market"`
	Lane          string    `json:"lane"`
	CompleteOnly  bool      `json:"completeOnly"`
	MinimumSample int       `json:"minimumSample"`
}

type PerformanceStates struct {
	Complete           int `json:"complete"`
	LinkMissing        int `json:"linkMissing"`
	NotMeasured        int `json:"notMeasured"`
	InsufficientSample int `json:"insufficientSample"`
}

type PerformanceAggregate struct {
	Market                string              `json:"market"`
	LaneID                string              `json:"laneId"`
	LaneVersion           string              `json:"laneVersion"`
	PolicyID              string              `json:"policyId"`
	PolicyVersion         string              `json:"policyVersion"`
	Samples               int                 `json:"samples"`
	Status                performance.Status  `json:"status"`
	NetPnL                string              `json:"netPnl"`
	AverageR              string              `json:"averageR"`
	WinRate               string              `json:"winRate"`
	ProfitFactor          string              `json:"profitFactor"`
	MaxDrawdown           string              `json:"maxDrawdown"`
	SlippagePct           string              `json:"slippagePct"`
	MFEPct                string              `json:"mfePct"`
	MAEPct                string              `json:"maePct"`
	Markout5Pct           string              `json:"markout5Pct"`
	Markout15Pct          string              `json:"markout15Pct"`
	Markout30Pct          string              `json:"markout30Pct"`
	SemanticsVersion      string              `json:"semanticsVersion"`
	ObservationProvenance string              `json:"observationProvenance"`
	Metrics               []PerformanceMetric `json:"metrics"`
}

type PerformanceMetric struct {
	Key        string             `json:"key"`
	Label      string             `json:"label"`
	Help       string             `json:"help"`
	Unit       string             `json:"unit"`
	Value      string             `json:"value"`
	Samples    int                `json:"samples"`
	Status     performance.Status `json:"status"`
	Provenance string             `json:"provenance"`
}

func PerformanceFrom(view performance.DashboardView) PerformanceResource {
	out := PerformanceResource{
		Query: PerformanceQuery{AsOf: view.Query.AsOf.UTC(), PeriodDays: view.Query.PeriodDays,
			Market: view.Query.Market, Lane: view.Query.Lane, CompleteOnly: view.Query.CompleteOnly,
			MinimumSample: view.Query.MinimumSample},
		States: PerformanceStates{Complete: view.States.Complete, LinkMissing: view.States.LinkMissing,
			NotMeasured: view.States.NotMeasured, InsufficientSample: view.States.InsufficientSample},
		Aggregates: make([]PerformanceAggregate, 0, len(view.Aggregates)),
	}
	if !view.NewestSourceAt.IsZero() {
		at := view.NewestSourceAt.UTC()
		out.NewestSourceAt = &at
	}
	for _, aggregate := range view.Aggregates {
		projected := PerformanceAggregate{
			Market: aggregate.Market, LaneID: aggregate.LaneID, LaneVersion: aggregate.LaneVersion,
			PolicyID: aggregate.PolicyID, PolicyVersion: aggregate.PolicyVersion, Samples: aggregate.Samples,
			Status: aggregate.Status, NetPnL: aggregate.NetPnL, AverageR: aggregate.AverageR,
			WinRate: aggregate.WinRate, ProfitFactor: aggregate.ProfitFactor, MaxDrawdown: aggregate.MaxDrawdown,
			SlippagePct: aggregate.SlippagePct, MFEPct: aggregate.MFEPct, MAEPct: aggregate.MAEPct,
			Markout5Pct: aggregate.Markout5Pct, Markout15Pct: aggregate.Markout15Pct,
			Markout30Pct: aggregate.Markout30Pct, SemanticsVersion: aggregate.SemanticsVersion,
			ObservationProvenance: aggregate.ObservationProvenance,
			Metrics:               make([]PerformanceMetric, 0, len(aggregate.Metrics)),
		}
		for _, metric := range aggregate.Metrics {
			projected.Metrics = append(projected.Metrics, PerformanceMetric{Key: metric.Key, Label: metric.Label,
				Help: metric.Help, Unit: metric.Unit, Value: metric.Value, Samples: metric.Samples,
				Status: metric.Status, Provenance: metric.Provenance})
		}
		out.Aggregates = append(out.Aggregates, projected)
	}
	return out
}

type OptimizationResource struct {
	Version            uint64                       `json:"version"`
	EffectiveVersion   uint64                       `json:"effectiveVersion"`
	SettingsDigest     string                       `json:"settingsDigest"`
	Categories         []OptimizationCategory       `json:"categories"`
	Fields             []OptimizationField          `json:"fields"`
	PositionManagement PositionManagementDescriptor `json:"positionManagement"`
	CandidateFilters   []CandidateFilterMarket      `json:"candidateFilters"`
	StrategyRuntime    RuntimeDescriptor            `json:"strategyRuntime"`
	Evidence           OptimizationEvidence         `json:"evidence"`
	History            []OptimizationSnapshot       `json:"history"`
	Audit              []OptimizationAuditEvent     `json:"audit"`
}

type OptimizationCategory struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Purpose  string `json:"purpose"`
	ReadOnly bool   `json:"readOnly"`
}

type OptimizationField struct {
	Category           string                      `json:"category"`
	Key                string                      `json:"key"`
	Label              string                      `json:"label"`
	Description        string                      `json:"description"`
	Type               settingmeta.ValueType       `json:"type"`
	Unit               string                      `json:"unit"`
	Control            settingmeta.ControlKind     `json:"control"`
	Default            State                       `json:"default"`
	Desired            State                       `json:"desired"`
	Effective          State                       `json:"effective"`
	Choices            []OptimizationChoice        `json:"choices"`
	ApplyTiming        settingmeta.ApplyTiming     `json:"applyTiming"`
	SafetyDirection    settingmeta.SafetyDirection `json:"safetyDirection"`
	Owner              string                      `json:"owner"`
	Provenance         Provenance                  `json:"provenance"`
	ConfigurationError string                      `json:"configurationError"`
}

type OptimizationChoice struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Recommended bool   `json:"recommended"`
}

type OptimizationEvidence struct {
	Status     optimization.EvidenceStatus `json:"status"`
	Digest     string                      `json:"digest"`
	ObservedAt *time.Time                  `json:"observedAt"`
	Missing    []string                    `json:"missing"`
}

type OptimizationSnapshot struct {
	Version                  uint64                  `json:"version"`
	EffectiveVersion         uint64                  `json:"effectiveVersion"`
	Desired                  map[string]string       `json:"desired"`
	Effective                map[string]string       `json:"effective"`
	SettingsDigest           string                  `json:"settingsDigest"`
	EvidenceDigest           string                  `json:"evidenceDigest"`
	ActivationManifestDigest string                  `json:"activationManifestDigest"`
	EffectiveEntry           bool                    `json:"effectiveEntry"`
	RestartRequired          bool                    `json:"restartRequired"`
	Actor                    string                  `json:"actor"`
	Reason                   optimization.ReasonCode `json:"reason"`
	AuditID                  string                  `json:"auditId"`
	CreatedAt                time.Time               `json:"createdAt"`
}

type OptimizationAuditEvent struct {
	ID             int64                   `json:"id"`
	AuditID        string                  `json:"auditId"`
	Version        uint64                  `json:"version"`
	CandidateID    string                  `json:"candidateId"`
	Key            string                  `json:"key"`
	BeforeOptionID string                  `json:"beforeOptionId"`
	AfterOptionID  string                  `json:"afterOptionId"`
	Actor          string                  `json:"actor"`
	Reason         optimization.ReasonCode `json:"reason"`
	CreatedAt      time.Time               `json:"createdAt"`
}

func OptimizationFrom(view optimization.View) OptimizationResource {
	out := OptimizationResource{
		Version: view.Snapshot.Version, EffectiveVersion: view.Snapshot.EffectiveVersion,
		SettingsDigest:     view.Snapshot.SettingsDigest,
		Categories:         make([]OptimizationCategory, 0, len(optimization.Categories())),
		Fields:             make([]OptimizationField, 0, len(view.Registry.All())),
		PositionManagement: positionManagementFrom(positionpolicy.Descriptor()),
		CandidateFilters:   candidateFiltersFrom(candidate.CandidateFilterMarkets()),
		StrategyRuntime:    runtimeDescriptorFrom(strategyengine.DormantRuntimeDescriptor()),
		Evidence: OptimizationEvidence{Status: view.Evidence.Status, Digest: view.Evidence.Digest,
			Missing: append([]string(nil), view.Evidence.Missing...)},
		History: make([]OptimizationSnapshot, 0, len(view.History)),
		Audit:   make([]OptimizationAuditEvent, 0, len(view.Audit)),
	}
	if out.Evidence.Missing == nil {
		out.Evidence.Missing = []string{}
	}
	if !view.Evidence.ObservedAt.IsZero() {
		at := view.Evidence.ObservedAt.UTC()
		out.Evidence.ObservedAt = &at
	}
	for _, category := range optimization.Categories() {
		out.Categories = append(out.Categories, OptimizationCategory{ID: string(category.ID), Label: category.Label,
			Purpose: category.Purpose, ReadOnly: category.ReadOnly})
	}
	for _, registered := range view.Registry.All() {
		descriptor := registered.Descriptor
		field := OptimizationField{Category: string(registered.Category), Key: descriptor.Key, Label: descriptor.Label,
			Description: descriptor.Description, Type: descriptor.Type, Unit: descriptor.Unit, Control: descriptor.Control,
			Default:   stateFrom(descriptor.Default, descriptor.Options),
			Desired:   stateForOption(view.Snapshot.Desired[descriptor.Key], descriptor.Default, descriptor.Options),
			Effective: stateForOption(view.Snapshot.Effective[descriptor.Key], descriptor.Effective, descriptor.Options),
			Choices:   make([]OptimizationChoice, 0, len(descriptor.Options)), ApplyTiming: descriptor.ApplyTiming,
			SafetyDirection: descriptor.SafetyDirection, Owner: descriptor.Provenance.OwnerChange,
			Provenance: provenanceFrom(descriptor.Provenance), ConfigurationError: registered.ConfigurationError}
		for _, option := range descriptor.Options {
			field.Choices = append(field.Choices, OptimizationChoice{ID: option.ID, Label: option.Label,
				Description: option.Description, Value: option.Value, Recommended: option.Recommended})
		}
		out.Fields = append(out.Fields, field)
	}
	for _, snapshot := range view.History {
		out.History = append(out.History, optimizationSnapshotFrom(snapshot))
	}
	for _, event := range view.Audit {
		out.Audit = append(out.Audit, OptimizationAuditEvent{ID: event.ID, AuditID: event.AuditID,
			Version: event.Version, CandidateID: event.CandidateID, Key: event.Key,
			BeforeOptionID: event.BeforeOptionID, AfterOptionID: event.AfterOptionID, Actor: event.Actor,
			Reason: event.Reason, CreatedAt: event.CreatedAt.UTC()})
	}
	return out
}

func stateForOption(optionID string, fallback settingmeta.State, options []settingmeta.Option) State {
	if optionID == "" {
		return stateFrom(fallback, options)
	}
	for _, option := range options {
		if option.ID == optionID {
			return State{Kind: string(settingmeta.StateValue), OptionID: option.ID, Value: option.Value, Display: option.Label}
		}
	}
	return State{Kind: "invalid", OptionID: optionID, Display: "unknown option"}
}

func stateFrom(state settingmeta.State, options []settingmeta.Option) State {
	if state.Kind == settingmeta.StateValue {
		return stateForOption(state.OptionID, settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "unapproved"}, options)
	}
	return State{Kind: string(state.Kind), Display: state.Display}
}

func provenanceFrom(value settingmeta.Provenance) Provenance {
	return Provenance{OwnerChange: value.OwnerChange, PolicyID: value.PolicyID, PolicyVersion: value.PolicyVersion,
		PolicyDigest: value.PolicyDigest, EvidenceDigest: value.EvidenceDigest}
}

func optimizationSnapshotFrom(value optimization.Snapshot) OptimizationSnapshot {
	return OptimizationSnapshot{Version: value.Version, EffectiveVersion: value.EffectiveVersion,
		Desired: cloneStringMap(value.Desired), Effective: cloneStringMap(value.Effective), SettingsDigest: value.SettingsDigest,
		EvidenceDigest: value.EvidenceDigest, ActivationManifestDigest: value.ActivationManifestDigest,
		EffectiveEntry: value.EffectiveEntry, RestartRequired: value.RestartRequired, Actor: value.Actor,
		Reason: value.Reason, AuditID: value.AuditID, CreatedAt: value.CreatedAt.UTC()}
}

func cloneStringMap(value map[string]string) map[string]string {
	if value == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}

type StopOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Decimal string `json:"decimal"`
}

type PositionManagementDescriptor struct {
	Category             string       `json:"category"`
	PositionSection      string       `json:"positionSection"`
	AutoAdoptionSection  string       `json:"autoAdoptionSection"`
	AutoEnabledDefault   bool         `json:"autoEnabledDefault"`
	AutoEnabledDesired   bool         `json:"autoEnabledDesired"`
	AutoEnabledEffective bool         `json:"autoEnabledEffective"`
	StopDefault          string       `json:"stopDefault"`
	StopDesired          string       `json:"stopDesired"`
	StopEffective        string       `json:"stopEffective"`
	StopOptions          []StopOption `json:"stopOptions"`
	IncludeDefault       []string     `json:"includeDefault"`
	ExcludeDefault       []string     `json:"excludeDefault"`
	ExcludePrecedence    string       `json:"excludePrecedence"`
	ApplyTiming          string       `json:"applyTiming"`
	Provenance           string       `json:"provenance"`
	OneShareBehavior     string       `json:"oneShareBehavior"`
}

func positionManagementFrom(value positionpolicy.ManagementDescriptor) PositionManagementDescriptor {
	out := PositionManagementDescriptor{Category: value.Category, PositionSection: value.PositionSection,
		AutoAdoptionSection: value.AutoAdoptionSection, AutoEnabledDefault: value.AutoEnabledDefault,
		AutoEnabledDesired: value.AutoEnabledDesired, AutoEnabledEffective: value.AutoEnabledEffective,
		StopDefault: value.StopDefault, StopDesired: value.StopDesired, StopEffective: value.StopEffective,
		StopOptions:    make([]StopOption, 0, len(value.StopOptions)),
		IncludeDefault: append([]string(nil), value.IncludeDefault...), ExcludeDefault: append([]string(nil), value.ExcludeDefault...),
		ExcludePrecedence: value.ExcludePrecedence, ApplyTiming: value.ApplyTiming, Provenance: value.Provenance,
		OneShareBehavior: value.OneShareBehavior}
	for _, option := range value.StopOptions {
		out.StopOptions = append(out.StopOptions, StopOption{ID: option.ID, Label: option.Label, Decimal: option.Decimal})
	}
	if out.IncludeDefault == nil {
		out.IncludeDefault = []string{}
	}
	if out.ExcludeDefault == nil {
		out.ExcludeDefault = []string{}
	}
	return out
}

type CandidateFilterMarket struct {
	Market  string                      `json:"market"`
	Session string                      `json:"session"`
	Filters []CandidateFilterDescriptor `json:"filters"`
}

type CandidateFilterDescriptor struct {
	Key             string                   `json:"key"`
	Category        string                   `json:"category"`
	Label           string                   `json:"label"`
	Help            string                   `json:"help"`
	Market          string                   `json:"market"`
	Session         string                   `json:"session"`
	DefaultState    candidate.ThresholdState `json:"defaultState"`
	DesiredState    candidate.ThresholdState `json:"desiredState"`
	EffectiveState  candidate.ThresholdState `json:"effectiveState"`
	DesiredValue    string                   `json:"desiredValue"`
	EffectiveValue  string                   `json:"effectiveValue"`
	Unit            string                   `json:"unit"`
	ValidRange      string                   `json:"validRange"`
	Direction       string                   `json:"direction"`
	SampleState     string                   `json:"sampleState"`
	EvidenceState   string                   `json:"evidenceState"`
	ApplyTiming     string                   `json:"applyTiming"`
	ReadOnly        bool                     `json:"readOnly"`
	CASRequired     bool                     `json:"casRequired"`
	Provenance      string                   `json:"provenance"`
	LegacyValue     string                   `json:"legacyValue"`
	MissingEvidence []string                 `json:"missingEvidence"`
	PreviewContract string                   `json:"previewContract"`
}

func candidateFiltersFrom(values []candidate.CandidateFilterMarket) []CandidateFilterMarket {
	out := make([]CandidateFilterMarket, 0, len(values))
	for _, value := range values {
		market := CandidateFilterMarket{Market: value.Market, Session: value.Session,
			Filters: make([]CandidateFilterDescriptor, 0, len(value.Filters))}
		for _, field := range value.Filters {
			market.Filters = append(market.Filters, CandidateFilterDescriptor{Key: field.Key, Category: field.Category,
				Label: field.Label, Help: field.Help, Market: field.Market, Session: field.Session,
				DefaultState: field.DefaultState, DesiredState: field.DesiredState, EffectiveState: field.EffectiveState,
				DesiredValue: field.DesiredValue, EffectiveValue: field.EffectiveValue, Unit: field.Unit,
				ValidRange: field.ValidRange, Direction: field.Direction, SampleState: field.SampleState,
				EvidenceState: field.EvidenceState, ApplyTiming: field.ApplyTiming, ReadOnly: field.ReadOnly,
				CASRequired: field.CASRequired, Provenance: field.Provenance, LegacyValue: field.LegacyValue,
				MissingEvidence: append([]string(nil), field.MissingEvidence...), PreviewContract: field.PreviewContract})
		}
		out = append(out, market)
	}
	return out
}

type RuntimeDescriptor struct {
	Category string           `json:"category"`
	Sections []RuntimeSection `json:"sections"`
	Fields   []RuntimeField   `json:"fields"`
	Blockers []RuntimeBlocker `json:"blockers"`
}
type RuntimeSection struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	ActionOwner string `json:"actionOwner"`
}
type RuntimeField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Help        string `json:"help"`
	Default     string `json:"default"`
	Desired     string `json:"desired"`
	Effective   string `json:"effective"`
	Unit        string `json:"unit"`
	Range       string `json:"range"`
	Provenance  string `json:"provenance"`
	ApplyTiming string `json:"applyTiming"`
}
type RuntimeBlocker struct {
	Key       string                        `json:"key"`
	Label     string                        `json:"label"`
	Desired   strategyengine.RuntimeState   `json:"desired"`
	Effective strategyengine.RuntimeState   `json:"effective"`
	Freshness strategyengine.RuntimeState   `json:"freshness"`
	Reason    strategyengine.RuntimeRefusal `json:"reason"`
}

func runtimeDescriptorFrom(value strategyengine.RuntimeDescriptor) RuntimeDescriptor {
	out := RuntimeDescriptor{Category: value.Category, Sections: make([]RuntimeSection, 0, len(value.Sections)),
		Fields: make([]RuntimeField, 0, len(value.Fields)), Blockers: make([]RuntimeBlocker, 0, len(value.Blockers))}
	for _, item := range value.Sections {
		out.Sections = append(out.Sections, RuntimeSection{ID: item.ID, Label: item.Label, ActionOwner: item.ActionOwner})
	}
	for _, item := range value.Fields {
		out.Fields = append(out.Fields, RuntimeField{Key: item.Key, Label: item.Label, Help: item.Help, Default: item.Default, Desired: item.Desired, Effective: item.Effective, Unit: item.Unit, Range: item.Range, Provenance: item.Provenance, ApplyTiming: item.ApplyTiming})
	}
	for _, item := range value.Blockers {
		out.Blockers = append(out.Blockers, RuntimeBlocker{Key: item.Key, Label: item.Label, Desired: item.Desired, Effective: item.Effective, Freshness: item.Freshness, Reason: item.Reason})
	}
	return out
}
