// Package optimization owns the transport-neutral, versioned settings lifecycle.
// It deliberately has no broker, order, journal, lane, gate, or process-control
// dependency. Applying a settings candidate therefore cannot create LIVE authority.
package optimization

import (
	"context"
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type Category string

const (
	CategoryOverview           Category = "overview"
	CategoryExitProtection     Category = "exit-protection"
	CategoryPositionManagement Category = "position-management"
	CategoryCandidateFilters   Category = "candidate-filters"
	CategoryStrategyRuntime    Category = "strategy-runtime"
	CategoryPerformanceHistory Category = "performance-history"
)

type CategoryDescriptor struct {
	ID       Category
	Label    string
	Purpose  string
	ReadOnly bool
}

var categories = []CategoryDescriptor{
	{CategoryOverview, "개요", "설정 상태와 분리된 운영 권한 상태를 한눈에 확인합니다.", true},
	{CategoryExitProtection, "익절·보호", "신규 관리 포지션에 적용할 익절·보호 정책을 검토합니다.", false},
	{CategoryPositionManagement, "종목별 관리", "현재 포지션의 상속·override와 lifecycle을 확인합니다.", false},
	{CategoryCandidateFilters, "후보 필터", "승인된 threshold와 성과 근거의 준비 상태를 확인합니다.", false},
	{CategoryStrategyRuntime, "전략·실행", "시장·일정과 lane 실행 상태를 서로 분리해 확인합니다.", false},
	{CategoryPerformanceHistory, "성과·이력", "결정적 성과, 설정 이력과 rollback 후보를 검토합니다.", true},
}

func Categories() []CategoryDescriptor {
	out := make([]CategoryDescriptor, len(categories))
	copy(out, categories)
	return out
}

func ParseCategory(raw string) (Category, bool) {
	for _, category := range categories {
		if string(category.ID) == raw {
			return category.ID, true
		}
	}
	return CategoryOverview, false
}

var (
	ErrInvalidRegistry       = errors.New("optimization: invalid parameter registry")
	ErrInvalidCandidate      = errors.New("optimization: invalid candidate")
	ErrInsufficientEvidence  = errors.New("optimization: insufficient evidence")
	ErrVersionConflict       = errors.New("optimization: settings version conflict")
	ErrCapabilityInvalid     = errors.New("optimization: preview capability invalid")
	ErrCapabilityTooEarly    = errors.New("optimization: preview capability not active yet")
	ErrCapabilityExpired     = errors.New("optimization: preview capability expired")
	ErrConfirmationRequired  = errors.New("optimization: risk confirmation required")
	ErrHistoricalKeyInactive = errors.New("optimization: historical key is no longer active")
)

type DescriptorProvider interface {
	OwnerChange() string
	Descriptors(context.Context) ([]settingmeta.FieldDescriptor, error)
}

type ProviderBinding struct {
	Category Category
	Provider DescriptorProvider
}

type EvidenceStatus string

const (
	EvidenceComplete     EvidenceStatus = "complete"
	EvidenceInsufficient EvidenceStatus = "insufficient"
	EvidenceUnavailable  EvidenceStatus = "unavailable"
	EvidenceStale        EvidenceStatus = "stale"
)

type Evidence struct {
	Status     EvidenceStatus
	Digest     string
	ObservedAt time.Time
	Missing    []string
}

type EvidenceProvider interface {
	ReadEvidence(context.Context) (Evidence, error)
}

type CandidateSource string

const (
	SourceServerPreset CandidateSource = "server-preset"
	SourceEvidence     CandidateSource = "evidence-backed"
	SourceRollback     CandidateSource = "rollback"
)

type ReasonCode string

const (
	ReasonServerPreset ReasonCode = "operator-server-preset"
	ReasonRollback     ReasonCode = "operator-rollback"
)

type OptionChange struct {
	Key            string                      `json:"key"`
	BeforeOptionID string                      `json:"before_option_id,omitempty"`
	AfterOptionID  string                      `json:"after_option_id"`
	Category       Category                    `json:"category"`
	ApplyTiming    settingmeta.ApplyTiming     `json:"apply_timing"`
	Safety         settingmeta.SafetyDirection `json:"safety"`
}

type Snapshot struct {
	Version                  uint64            `json:"version"`
	EffectiveVersion         uint64            `json:"effective_version"`
	Desired                  map[string]string `json:"desired"`
	Effective                map[string]string `json:"effective"`
	SettingsDigest           string            `json:"settings_digest"`
	EvidenceDigest           string            `json:"evidence_digest,omitempty"`
	ActivationManifestDigest string            `json:"activation_manifest_digest,omitempty"`
	EffectiveEntry           bool              `json:"effective_entry"`
	RestartRequired          bool              `json:"restart_required"`
	Actor                    string            `json:"actor"`
	Reason                   ReasonCode        `json:"reason"`
	AuditID                  string            `json:"audit_id"`
	CreatedAt                time.Time         `json:"created_at"`
}

type PreviewRequest struct {
	BaseVersion uint64
	Category    Category
	Changes     map[string]string
	Source      CandidateSource
	Reason      ReasonCode
}

type RollbackPreviewRequest struct {
	BaseVersion   uint64
	TargetVersion uint64
	Category      Category
}

type Preview struct {
	CandidateID                string
	BaseVersion                uint64
	Category                   Category
	Changes                    []OptionChange
	Evidence                   Evidence
	Capability                 string
	NotBefore                  time.Time
	ExpiresAt                  time.Time
	RestartRequired            bool
	RiskConfirmationRequired   bool
	ExistingPositionsUnchanged bool
	LiveStateUnchanged         bool
	EffectiveEntryAfterApply   bool
}

type ApplyRequest struct {
	Capability string
	Confirmed  bool
}

type ApplyResult struct {
	Snapshot Snapshot
	Replayed bool
}

// ConflictView is a read-only recovery projection for a stale apply. It keeps
// the immutable attempted candidate beside the latest snapshot so a transport
// can offer an explicit new preview without retrying the stale command.
type ConflictView struct {
	BaseVersion uint64
	Category    Category
	Attempted   []OptionChange
	Latest      Snapshot
	Registry    Registry
}

type AuditEvent struct {
	ID             int64
	AuditID        string
	Version        uint64
	CandidateID    string
	Key            string
	BeforeOptionID string
	AfterOptionID  string
	Actor          string
	Reason         ReasonCode
	CreatedAt      time.Time
}

type View struct {
	Registry Registry
	Snapshot Snapshot
	History  []Snapshot
	Audit    []AuditEvent
	Evidence Evidence
}

// Commander is the complete optimization write authority exposed to a
// transport. It has no method for LIVE, lane, gate, kill switch, broker, or
// journal mutation.
type Commander interface {
	Read(context.Context) (View, error)
	Preview(context.Context, PreviewRequest) (Preview, error)
	PreviewRollback(context.Context, RollbackPreviewRequest) (Preview, error)
	Apply(context.Context, ApplyRequest) (ApplyResult, error)
	RecoverConflict(context.Context, string) (ConflictView, error)
}
