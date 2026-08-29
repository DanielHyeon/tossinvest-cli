package console

import (
	"context"
	"net/http"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type MultiMarketStrategyRuntimeReader interface {
	Read(context.Context) (strategyprojection.Snapshot, error)
}

type multiMarketStrategyRuntimePage struct {
	chrome
	SchemaVersion string
	GeneratedAt   string
	// 운영자가 활성화 매니페스트에 옮겨 적는 두 숫자다. 시장별이 아니라 엔진 프로세스
	// 하나의 사실이라서 시장 카드가 아니라 상단 contract 절에 있다 — a112 결정 54.
	ConfigDigest string
	BuildDigest  string
	LoadErr      bool
	Unwired      bool
	Fields       []strategyprojection.FieldDescriptor
	Markets      []multiMarketStrategyRuntimeView
}

func (multiMarketStrategyRuntimePage) Refresh() bool { return false }

type multiMarketStrategyRuntimeView struct {
	Market, Status, StatusClass, ErrorCode                       string
	LaneID, LaneVersion, LaneDesired, LaneEffective              string
	EvidenceID, EvidenceDigest, EvidenceFreshness                string
	CampaignID, LegID, RiskBucket, RiskPolicyVersion, RiskStatus string
	SchedulerDesired, SchedulerEffective                         string
	CalendarSource, CalendarVersion, CalendarFreshness           string
	ActivationDesired, ActivationEffective, ActivationDigest     string
	ActivationStatus, ProtectionReadiness, ProtectionRefusal     string
	ReconciliationStatus, ReconciliationRefusal, FirstRefusal    string
	ObservedAt                                                   string
}

func (c *Console) buildMultiMarketStrategyRuntimePage(r *http.Request) multiMarketStrategyRuntimePage {
	generatedAt := c.now().UTC()
	snapshot := strategyprojection.DormantSnapshot(generatedAt)
	page := multiMarketStrategyRuntimePage{chrome: c.chromeOnRequest("optimization-sub"),
		Unwired: c.opts.StrategyRuntime == nil}
	if c.opts.StrategyRuntime != nil {
		value, err := c.opts.StrategyRuntime.Read(r.Context())
		if err != nil || strategyprojection.Validate(value) != nil {
			page.LoadErr = true
			snapshot = strategyprojection.UnavailableSnapshot(generatedAt)
		} else {
			snapshot = strategyprojection.Clone(value)
		}
	}
	page.project(snapshot)
	return page
}

func (page *multiMarketStrategyRuntimePage) project(snapshot strategyprojection.Snapshot) {
	page.SchemaVersion = snapshot.SchemaVersion
	page.GeneratedAt = runtimeProjectionTime(snapshot.GeneratedAt)
	// 엔진이 보낸 값만 쓴다. 콘솔이 자기 바이너리에서 같은 값을 계산해 빈칸을 채우면,
	// 두 build 가 다를 때 엔진이 거절할 매니페스트를 운영자가 만들게 된다.
	page.ConfigDigest = projectionValue(snapshot.Runtime.ConfigDigest)
	page.BuildDigest = projectionValue(snapshot.Runtime.BuildDigest)
	page.Fields = strategyprojection.Registry()
	page.Markets = make([]multiMarketStrategyRuntimeView, 0, 2)
	for _, item := range strategyprojection.OrderedMarkets(snapshot) {
		view := multiMarketStrategyRuntimeView{Market: string(item.Market), Status: string(item.Status), StatusClass: runtimeProjectionClass(string(item.Status)),
			LaneID: projectionValue(item.Lane.ID), LaneVersion: projectionValue(item.Lane.Version), LaneDesired: string(item.Lane.Desired), LaneEffective: string(item.Lane.Effective),
			EvidenceID: projectionValue(item.Evidence.ID), EvidenceDigest: projectionValue(item.Evidence.Digest), EvidenceFreshness: string(item.Evidence.Freshness),
			CampaignID: projectionValue(item.Campaign.ID), LegID: projectionValue(item.Campaign.LegID), RiskBucket: projectionValue(item.HorizonRisk.Bucket),
			RiskPolicyVersion: projectionValue(item.HorizonRisk.PolicyVersion), RiskStatus: string(item.HorizonRisk.Status),
			SchedulerDesired: string(item.Scheduler.Desired), SchedulerEffective: string(item.Scheduler.Effective),
			CalendarSource: projectionValue(item.Scheduler.CalendarSource), CalendarVersion: projectionValue(item.Scheduler.CalendarVersion), CalendarFreshness: string(item.Scheduler.CalendarFreshness),
			ActivationDesired: string(item.Activation.Desired), ActivationEffective: string(item.Activation.Effective), ActivationDigest: projectionValue(item.Activation.Digest),
			ActivationStatus: string(item.Activation.Status), ProtectionReadiness: string(item.Protection.Readiness), ProtectionRefusal: string(item.Protection.Refusal),
			ReconciliationStatus: string(item.Reconciliation.Status), ReconciliationRefusal: string(item.Reconciliation.Refusal),
			FirstRefusal: string(item.FirstRefusal), ObservedAt: runtimeProjectionTimePointer(item.ObservedAt)}
		if item.Error != nil {
			view.ErrorCode = string(item.Error.Code)
		}
		page.Markets = append(page.Markets, view)
	}
}

func projectionValue(value *string) string {
	if value == nil {
		return "not_observed"
	}
	return *value
}

func runtimeProjectionTimePointer(value *time.Time) string {
	if value == nil {
		return "관측 없음"
	}
	return runtimeProjectionTime(*value)
}

func runtimeProjectionTime(value time.Time) string {
	if value.IsZero() {
		return "관측 없음"
	}
	return value.UTC().Format(time.RFC3339)
}

func runtimeProjectionClass(value string) string {
	if value == string(strategyprojection.StatusCurrent) || value == string(strategyprojection.ProtectionWired) || value == string(strategyprojection.ReconciliationHealthy) {
		return "ok"
	}
	return "bad"
}
