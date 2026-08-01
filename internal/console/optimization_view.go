package console

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

// OptimizationCommander is the console's complete settings-lifecycle
// authority. The type contains no broker, journal, lane, gate, kill-switch or
// LIVE mutation operation. cmd/tossctl owns the durable implementation.
type OptimizationCommander interface {
	Read(context.Context) (strategyopt.View, error)
	Preview(context.Context, strategyopt.PreviewRequest) (strategyopt.Preview, error)
	PreviewRollback(context.Context, strategyopt.RollbackPreviewRequest) (strategyopt.Preview, error)
	Apply(context.Context, strategyopt.ApplyRequest) (strategyopt.ApplyResult, error)
}

// ExitPolicySettings remains a read-only compatibility projection until the
// engine consumes the canonical optimization snapshot. Save is retained on the
// legacy adapter type for binary compatibility, but a050 never calls it.
type ExitPolicySettings interface {
	Load() (config.ExitPolicy, error)
	Save(config.ExitPolicy) error
}

type optimizationPolicyView struct {
	exitpolicy.CommonPolicy
	Selected bool
	Token    string
}

type optimizationCategoryView struct {
	strategyopt.CategoryDescriptor
	Active    bool
	Available bool
	Status    string
}

type optimizationFieldView struct {
	Key, Label, Description, Unit, Default, Desired, Effective    string
	ApplyTiming, Safety, Owner, PolicyID, PolicyVersion, Evidence string
	Control                                                       settingmeta.ControlKind
	Options                                                       []settingmeta.Option
}

type optimizationPage struct {
	Nav, CSRF, Notice, Warning     string
	Selected                       strategyopt.Category
	Categories                     []optimizationCategoryView
	LifecycleWired, LifecycleReady bool
	LifecycleErr                   string
	Snapshot                       strategyopt.Snapshot
	Evidence                       strategyopt.Evidence
	Fields                         []optimizationFieldView
	History                        []strategyopt.Snapshot
	Audit                          []strategyopt.AuditEvent
	LegacyCurrent                  config.ExitPolicy
	LegacyLoadErr                  string
	EngineRunning                  bool
	Policies                       []optimizationPolicyView
}

func (p optimizationPage) Overview() bool { return p.Selected == strategyopt.CategoryOverview }
func (p optimizationPage) ExitProtection() bool {
	return p.Selected == strategyopt.CategoryExitProtection
}
func (p optimizationPage) PositionManagement() bool {
	return p.Selected == strategyopt.CategoryPositionManagement
}
func (p optimizationPage) CandidateFilters() bool {
	return p.Selected == strategyopt.CategoryCandidateFilters
}
func (p optimizationPage) StrategyRuntime() bool {
	return p.Selected == strategyopt.CategoryStrategyRuntime
}
func (p optimizationPage) PerformanceHistory() bool {
	return p.Selected == strategyopt.CategoryPerformanceHistory
}

type optimizationPreviewPage struct {
	Nav, CSRF string
	Preview   strategyopt.Preview
	WaitSecs  int
}

func (optimizationPreviewPage) Refresh() bool { return false }

func optimizationCategoryViews(selected strategyopt.Category, writable bool) []optimizationCategoryView {
	views := make([]optimizationCategoryView, 0, len(strategyopt.Categories()))
	for _, category := range strategyopt.Categories() {
		view := optimizationCategoryView{CategoryDescriptor: category, Active: category.ID == selected}
		switch category.ID {
		case strategyopt.CategoryOverview:
			view.Available, view.Status = true, "읽기 전용 요약"
		case strategyopt.CategoryExitProtection:
			view.Available = writable
			if writable {
				view.Status = "owner descriptor 연결됨"
			} else {
				view.Status = "command seam 미배선 · 읽기 전용"
			}
		case strategyopt.CategoryPositionManagement:
			view.Available, view.Status = true, "a044 전용 command 화면으로 분리"
		case strategyopt.CategoryCandidateFilters:
			view.Available, view.Status = true, "a046 evidence activation 전까지 읽기 전용"
		case strategyopt.CategoryStrategyRuntime:
			view.Status = "a047 owner descriptor 미통합 · OFF/read-only"
		case strategyopt.CategoryPerformanceHistory:
			view.Status = "a049 provider 미통합 · unavailable"
		}
		views = append(views, view)
	}
	return views
}

func optimizationFieldViews(fields []strategyopt.RegisteredField, snapshot strategyopt.Snapshot) []optimizationFieldView {
	out := make([]optimizationFieldView, 0, len(fields))
	for _, registered := range fields {
		d := registered.Descriptor
		out = append(out, optimizationFieldView{Key: d.Key, Label: d.Label, Description: d.Description,
			Unit: displayUnit(d.Unit), Default: displaySettingState(d.Default, d.Options),
			Desired:     displayOption(snapshot.Desired[d.Key], d.Default, d.Options),
			Effective:   displayOption(snapshot.Effective[d.Key], d.Effective, d.Options),
			ApplyTiming: displayApplyTiming(d.ApplyTiming), Safety: displaySafety(d.SafetyDirection),
			Owner: d.Provenance.OwnerChange, PolicyID: d.Provenance.PolicyID, PolicyVersion: d.Provenance.PolicyVersion,
			Evidence: d.Provenance.EvidenceDigest, Control: d.Control, Options: append([]settingmeta.Option(nil), d.Options...)})
	}
	return out
}

func displayOption(optionID string, fallback settingmeta.State, options []settingmeta.Option) string {
	if optionID == "" {
		return displaySettingState(fallback, options)
	}
	for _, option := range options {
		if option.ID == optionID {
			return option.Label + " · " + option.ID
		}
	}
	return "거부됨 · unknown option " + optionID
}

func displaySettingState(state settingmeta.State, options []settingmeta.Option) string {
	if state.Kind == settingmeta.StateValue {
		return displayOption(state.OptionID, settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "미승인"}, options)
	}
	return state.Display
}

func displayUnit(unit string) string {
	if strings.TrimSpace(unit) == "" {
		return "해당 없음"
	}
	return unit
}
func displayApplyTiming(t settingmeta.ApplyTiming) string {
	return map[settingmeta.ApplyTiming]string{
		settingmeta.ApplyImmediate: "즉시", settingmeta.ApplyNextEvaluation: "다음 평가",
		settingmeta.ApplyNextEngineStart: "다음 엔진 기동", settingmeta.ApplyNewPositionOnly: "신규 포지션만"}[t]
}
func displaySafety(s settingmeta.SafetyDirection) string {
	return map[settingmeta.SafetyDirection]string{
		settingmeta.SafetySaferWhenHigher: "값이 높을수록 보수적", settingmeta.SafetySaferWhenLower: "값이 낮을수록 보수적",
		settingmeta.SafetyNeutral: "중립", settingmeta.SafetyContextual: "상황별 위험 확인 필요"}[s]
}
func waitSeconds(preview strategyopt.Preview) int {
	if preview.RiskConfirmationRequired {
		return 3
	}
	return 0
}

func (c *Console) writeOptimizationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, strategyopt.ErrVersionConflict):
		c.refuse(w, http.StatusPreconditionFailed, "설정 version 충돌", "최신 desired/effective 값을 다시 읽고 preview부터 시작하라. draft를 자동 재적용하지 않았다.")
	case errors.Is(err, strategyopt.ErrCapabilityTooEarly):
		c.refuse(w, http.StatusTooEarly, "3초 위험 확인 대기 중", "before/after를 확인한 뒤 같은 승인 버튼을 다시 눌러라.")
	case errors.Is(err, strategyopt.ErrCapabilityExpired):
		c.refuse(w, http.StatusGone, "preview 만료", "현재값에서 새 preview를 만들어라. 아무것도 변경되지 않았다.")
	case errors.Is(err, strategyopt.ErrCapabilityInvalid):
		c.refuse(w, http.StatusForbidden, "preview capability 거부", "서버가 발급한 현재 preview capability만 사용할 수 있다.")
	case errors.Is(err, strategyopt.ErrConfirmationRequired):
		c.refuse(w, http.StatusBadRequest, "위험 변경 확인 필요", "영향 범위 확인 checkbox를 선택하라. 문구 입력은 필요 없다.")
	case errors.Is(err, strategyopt.ErrInsufficientEvidence):
		c.refuse(w, http.StatusConflict, "성과 근거 부족", "추천 candidate는 만들지 않았다. owner registry의 보수적 server preset만 별도로 검토할 수 있다.")
	default:
		c.refuse(w, http.StatusBadRequest, "최적화 요청 거부", err.Error()+" 아무것도 변경되지 않았다.")
	}
}
