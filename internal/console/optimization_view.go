package console

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

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
	RecoverConflict(context.Context, string) (strategyopt.ConflictView, error)
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
	ConfigurationError                                            string
	Control                                                       settingmeta.ControlKind
	Options                                                       []settingmeta.Option
}

type optimizationPage struct {
	chrome
	CSRF, Notice, Warning          string
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
	ProtectionWired                bool
	ProtectionLoadErr              string
	Protections                    []ProtectionStatus
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

// ExitPolicyWritable keeps the only a050 mutation surface fail-closed. The
// preset catalogue remains useful as read-only reference material when the
// owner descriptor is absent or invalid, but it must not produce a preview.
func (p optimizationPage) ExitPolicyWritable() bool {
	if !p.LifecycleReady {
		return false
	}
	for _, field := range p.Fields {
		if field.Key == "exit.common-policy" {
			return field.ConfigurationError == "" && field.Control != settingmeta.ControlReadOnly
		}
	}
	return false
}

type optimizationPreviewPage struct {
	chrome
	CSRF               string
	Preview            strategyopt.Preview
	WaitSecs           int
	Waiting            bool
	NotBeforeUnixMilli int64
}

const optimizationPreviewScript = `(function(){
"use strict";
var root=document.querySelector("[data-risk-preview]");
if(!root){return;}
var button=root.querySelector("[data-risk-submit]");
var confirmation=root.querySelector("[data-risk-confirm]");
var live=root.querySelector("[data-risk-countdown]");
var deadline=Number(root.getAttribute("data-not-before-ms"));
function tick(){
  var remaining=Math.max(0,Math.ceil((deadline-Date.now())/1000));
  var confirmed=!confirmation||confirmation.checked;
  button.disabled=remaining!==0||!confirmed;
  if(remaining===0){live.textContent=confirmed?"승인 가능":"확인 항목을 선택하세요";return;}
  live.textContent=remaining+"초 남음";
  window.setTimeout(tick,250);
}
if(confirmation){confirmation.addEventListener("change",tick);}
tick();
}());`

var optimizationPreviewCSP = func() string {
	digest := sha256.Sum256([]byte(optimizationPreviewScript))
	return consoleHTMLCSP + "; script-src 'sha256-" + base64.StdEncoding.EncodeToString(digest[:]) + "'"
}()

func (c *Console) previewPage(preview strategyopt.Preview) optimizationPreviewPage {
	remaining := preview.NotBefore.Sub(c.now())
	waiting := preview.RiskConfirmationRequired && remaining > 0
	waitSecs := 0
	if waiting {
		waitSecs = int((remaining + time.Second - 1) / time.Second)
	}
	return optimizationPreviewPage{chrome: c.chromeOnRequest("optimization"), CSRF: c.csrf, Preview: preview,
		WaitSecs: waitSecs, Waiting: waiting, NotBeforeUnixMilli: preview.NotBefore.UnixMilli()}
}

func (c *Console) renderOptimizationPreview(w http.ResponseWriter, preview strategyopt.Preview) {
	c.renderHTML(w, http.StatusOK, "optimization-preview", c.previewPage(preview), optimizationPreviewCSP)
}

type optimizationConflictRow struct {
	Key, Attempted, LatestDesired, LatestEffective string
}

type optimizationConflictPage struct {
	chrome
	CSRF                          string
	Category                      strategyopt.Category
	BaseVersion, LatestVersion    uint64
	Rows                          []optimizationConflictRow
	CanRepreview                  bool
	RepreviewKey, RepreviewOption string
}

func (c *Console) newOptimizationConflictPage(csrf string, conflict strategyopt.ConflictView) optimizationConflictPage {
	page := optimizationConflictPage{chrome: c.chromeOnRequest("optimization"), CSRF: csrf, Category: conflict.Category,
		BaseVersion: conflict.BaseVersion, LatestVersion: conflict.Latest.Version}
	for _, change := range conflict.Attempted {
		field, ok := conflict.Registry.Field(change.Key)
		if !ok {
			continue
		}
		page.Rows = append(page.Rows, optimizationConflictRow{Key: change.Key,
			Attempted:       displayOption(change.AfterOptionID, field.Descriptor.Default, field.Descriptor.Options),
			LatestDesired:   displayOption(conflict.Latest.Desired[change.Key], field.Descriptor.Default, field.Descriptor.Options),
			LatestEffective: displayOption(conflict.Latest.Effective[change.Key], field.Descriptor.Effective, field.Descriptor.Options)})
	}
	if len(conflict.Attempted) == 1 {
		attempted := conflict.Attempted[0]
		field, ok := conflict.Registry.Field(attempted.Key)
		page.CanRepreview = ok && attempted.AfterOptionID != "" &&
			field.Descriptor.ValidateOption(attempted.AfterOptionID) == nil &&
			conflict.Latest.Desired[attempted.Key] != attempted.AfterOptionID
		if page.CanRepreview {
			page.RepreviewKey, page.RepreviewOption = attempted.Key, attempted.AfterOptionID
		}
	}
	return page
}

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
			view.Available = true
			view.Status = "a049 결정적 성과 · 읽기 전용 (근거 상태는 본문 참조)"
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
			Evidence: d.Provenance.EvidenceDigest, ConfigurationError: registered.ConfigurationError,
			Control: d.Control, Options: append([]settingmeta.Option(nil), d.Options...)})
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

func (c *Console) writeOptimizationApplyError(w http.ResponseWriter, r *http.Request, capability string, err error) {
	if !errors.Is(err, strategyopt.ErrVersionConflict) {
		c.writeOptimizationError(w, err)
		return
	}
	conflict, recoveryErr := c.opts.Optimization.RecoverConflict(r.Context(), capability)
	if recoveryErr != nil {
		c.writeOptimizationError(w, err)
		return
	}
	c.renderHTML(w, http.StatusPreconditionFailed, "optimization-conflict",
		c.newOptimizationConflictPage(c.csrf, conflict), consoleHTMLCSP)
}
