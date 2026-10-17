//go:build tossos_testseams

package riskbucket

// a112 6.1 — family 결속(Manager 판정 (C), 2026-10-01): 서명 위험 정책의 strategy 항목 risk_id 는 **한 family 를 함의**하고 적재기가 그것을 강제한다.
// 한 risk_id 를 서로 다른 family 의 레인이 공유하거나(family 버킷 경계가 무너짐), family 를 해소할 수 없는 레인을 적으면 정책 전체를 거절한다 —
// 결함 등급(그 시장 위험 미준비 · `ErrProductionRiskSnapshotUnavailable`), 범위 국소 거절(`ErrProductionRiskScopeRefused`)이 아니다.
// lane → family 는 strategyrouter 정본 표에서 유도(`strategyrouter.ProductionLaneFamily`).

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112Resign 은 정책 본문을 고쳐 같은 키로 다시 서명하고 파일 · digest 를 바꾼다.
func a112Resign(t *testing.T, fixture productionRiskFixture, mutate func(*productionRiskPolicyBody)) productionRiskFixture {
	t.Helper()
	body := fixture.body
	body.Strategies = append([]productionRiskStrategyPolicy(nil), body.Strategies...)
	mutate(&body)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(productionRiskPolicyManifest{productionRiskPolicyBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519Sign(fixture, bodyJSON))})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fixture.filePath, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.filePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fixture.filePath, 0o400); err != nil {
		t.Fatal(err)
	}
	fixture.body = body
	fixture.config.ManifestDigest = productionRiskDigest(data)
	return fixture
}

func a112KRLane(t *testing.T, family strategyrouter.Family) strategyflow.Descriptor {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.Market != strategyrouter.MarketKR {
			continue
		}
		if got, ok := strategyrouter.ProductionLaneFamily(strategyrouter.MarketKR, descriptor.LaneID); ok && got == family {
			return descriptor
		}
	}
	t.Fatalf("no KR descriptor of family %s", family)
	return strategyflow.Descriptor{}
}

func a112Load(fixture productionRiskFixture) error {
	_, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	return err
}

func TestARiskIDSharedByTwoFamiliesRefusesThePolicy(t *testing.T) {
	now := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	fixture := newProductionRiskFixture(t, MarketKR, now)
	if err := a112Load(fixture); err != nil {
		t.Fatalf("arrangement: the base policy must load: %v", err)
	}
	own := fixture.body.Strategies[0]
	reversal := a112KRLane(t, strategyrouter.FamilyReversal)
	if reversal.Horizon != strategyrouter.HorizonShort {
		t.Fatalf("arrangement: reversal lane horizon %s", reversal.Horizon)
	}
	// 대조: 다른 family 의 레인이 **다른** risk_id 를 쓰면 수락.
	distinct := a112Resign(t, fixture, func(body *productionRiskPolicyBody) {
		body.Strategies = append(body.Strategies, productionRiskStrategyPolicy{LaneID: reversal.LaneID, LaneVersion: reversal.LaneVersion,
			Horizon: HorizonShort, RiskID: "reversal", RiskVersion: "reversal-risk-v1", LimitMinor: "5000000"})
	})
	if err := a112Load(distinct); err != nil {
		t.Fatalf("control: two families with their own risk ids must load: %v", err)
	}
	// 같은 risk_id 를 두 family 가 공유 → 정책 전체 거절(결함, 범위 거절 아님).
	shared := a112Resign(t, fixture, func(body *productionRiskPolicyBody) {
		body.Strategies = append(body.Strategies, productionRiskStrategyPolicy{LaneID: reversal.LaneID, LaneVersion: reversal.LaneVersion,
			Horizon: HorizonShort, RiskID: own.RiskID, RiskVersion: own.RiskVersion, LimitMinor: own.LimitMinor})
	})
	err := a112Load(shared)
	if err == nil || !errors.Is(err, ErrProductionRiskSnapshotUnavailable) || errors.Is(err, ErrProductionRiskScopeRefused) {
		t.Fatalf("risk id shared by two families: err=%v — want the policy refused as a fault (not scope-local)", err)
	}
}

func TestAStrategyEntryWhoseLaneHasNoFamilyRefusesThePolicy(t *testing.T) {
	now := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	fixture := newProductionRiskFixture(t, MarketKR, now)
	unknown := a112Resign(t, fixture, func(body *productionRiskPolicyBody) {
		body.Strategies = append(body.Strategies, productionRiskStrategyPolicy{LaneID: "kr-unknown-lane", LaneVersion: "v1",
			Horizon: HorizonShort, RiskID: "unknown", RiskVersion: "unknown-risk-v1", LimitMinor: "5000000"})
	})
	err := a112Load(unknown)
	if err == nil || !errors.Is(err, ErrProductionRiskSnapshotUnavailable) || errors.Is(err, ErrProductionRiskScopeRefused) {
		t.Fatalf("lane without a family: err=%v — want the policy refused as a fault", err)
	}
}

// weekly horizon 핀(Manager 판정 2026-10-01): weekly family 레인(horizon WEEKLY)은 위험 버킷 horizon(SHORT/MEDIUM)에 매핑되지 않으므로 적재기가
// 「unsupported horizon」으로 거절하는 것이 **의도된 현재 상태**다 — 매핑은 weekly 활성화 로트의 freeze 몫(ROADMAP). 미래 편집이 그것을 무음으로
// SHORT 에 욱여넣으면 이 시험이 실패한다. 거절은 결함 등급(범위 국소 아님).
func TestAWeeklyLaneIsRefusedForItsHorizonUntilAMappingIsDecided(t *testing.T) {
	now := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	fixture := newProductionRiskFixture(t, MarketKR, now)
	weekly := a112KRLane(t, strategyrouter.FamilyWeeklyValue)
	if weekly.Horizon != strategyrouter.HorizonWeekly {
		t.Fatalf("arrangement: weekly lane horizon %s", weekly.Horizon)
	}
	result, err := strategyflow.AcceptedResultForAuthorityTest(weekly, "acct-production-risk", "005930", "campaign-weekly", 8,
		"100", "95", "120", now.Add(-time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	fixture.input.Result = result
	err = a112Load(fixture)
	if err == nil || !strings.Contains(err.Error(), "unsupported horizon") || errors.Is(err, ErrProductionRiskScopeRefused) {
		t.Fatalf("weekly lane: err=%v — want the pinned 'unsupported horizon' fault until the weekly horizon mapping is decided", err)
	}
}

func ed25519Sign(fixture productionRiskFixture, message []byte) []byte {
	return ed25519.Sign(fixture.private, message)
}

// 시장 축: family 는 **정책의 시장** 표에서 해소한다 — US 정책은 US 레인으로 적재되고(대조), KR 정책에 US 레인을 적으면 그 레인은 이 시장에서
// family 가 없으므로 정책 거절.
func TestTheFamilyIsResolvedInThePolicysOwnMarket(t *testing.T) {
	now := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	us := newProductionRiskFixture(t, MarketUS, now)
	if err := a112Load(us); err != nil {
		t.Fatalf("control: the US policy with its own US lane must load: %v", err)
	}
	kr := newProductionRiskFixture(t, MarketKR, now)
	usLane := us.body.Strategies[0]
	foreign := a112Resign(t, kr, func(body *productionRiskPolicyBody) {
		body.Strategies = append(body.Strategies, productionRiskStrategyPolicy{LaneID: usLane.LaneID, LaneVersion: usLane.LaneVersion,
			Horizon: usLane.Horizon, RiskID: "us-lane-in-kr", RiskVersion: "x-risk-v1", LimitMinor: "5000000"})
	})
	if err := a112Load(foreign); err == nil || !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("a US lane in the KR policy: err=%v — want the policy refused", err)
	}
}
