//go:build tossos_testseams

package strategyrouter

// a127 D4 — route 적재기의 실제 원장(`journal.Open`) 수락 시험은 journal 을 import 해야 해서 외부 시험 패키지에 둔다. 그 시험이 쓸 서명 매니페스트를
// 여기서 만든다(시험 seam 빌드에만 존재). 매니페스트 본문은 내부 픽스처 `productionRouteFixture.body` 와 같아야 하며, 그 일치는
// TestA127RouteManifestSeamMatchesTheInternalFixture 가 단언한다(둘째 철자가 갈라지면 실패).

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// productionRouteBodyForTest 는 시장 하나의 네 가족 route 매니페스트 본문임(범위 하나, owner revision 1 — 빈 원장의 재구성 값).
func productionRouteBodyForTest(market Market, now time.Time) productionRouteBody {
	symbol := map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market]
	timezone := map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[market]
	prefix := map[Market]string{MarketKR: "kr", MarketUS: "us"}[market]
	lane := func(family Family, horizon Horizon, laneID, laneVersion, name string, scorePPM uint32) productionRouteCandidate {
		return productionRouteCandidate{Family: family, Horizon: horizon, LaneID: laneID, LaneVersion: laneVersion,
			ScorePPM: scorePPM, Eligible: true, Desired: StateOn, Effective: StateOn,
			EvidenceDigest: prefix + "-" + name + "-evidence", ConfigDigest: prefix + "-" + name + "-config"}
	}
	lanes := []productionRouteCandidate{
		lane(FamilyContinuation, HorizonShort, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, "continuation", 300_000),
		lane(FamilyReversal, HorizonShort, reversallane.USReversalLaneID, reversallane.LaneVersionV1, "reversal", 200_000),
		lane(FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, "weekly", 100_000),
		lane(FamilyBreakoutRetest, HorizonShort, breakoutlane.USLaneID, breakoutlane.LaneVersionV1, "breakout", 50_000),
	}
	if market == MarketKR {
		lanes = []productionRouteCandidate{
			lane(FamilyContinuation, HorizonShort, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, "continuation", 300_000),
			lane(FamilyReversal, HorizonShort, reversallane.KRReversalLaneID, reversallane.LaneVersionV1, "reversal", 200_000),
			lane(FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, "weekly", 100_000),
			lane(FamilyBreakoutRetest, HorizonShort, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, "breakout", 50_000),
		}
	}
	return productionRouteBody{SchemaVersion: productionRouteSchema, Domain: productionRouteDomain, SignatureAlgorithm: productionRouteAlgorithm,
		KeyID: "route-key-v1", Generation: 1, AccountRef: "acct", Market: market,
		MarketRevision: 1, ActivationDigest: "activation-" + string(market), ActivationExpiresAt: now.Add(time.Hour).Format(time.RFC3339Nano),
		CalendarGeneration: "calendar-generation-" + string(market), CalendarDigest: "calendar-digest-" + string(market), Timezone: timezone,
		SessionScope: "REGULAR", ConfigVersion: "scheduler-config-" + string(market),
		ArbitrationScoreVersion: "arbitration-score-v1", CalibrationDigest: "sha256:calibration-" + string(market),
		Actor: "human-approver", ObservedAt: now.Add(-time.Minute).Format(time.RFC3339Nano),
		FreshUntil: now.Add(30 * time.Minute).Format(time.RFC3339Nano), Scopes: []productionRouteScope{{Symbol: symbol, PositionGeneration: 1, OwnerRevision: 1, Candidates: lanes}}}
}

// SignedProductionRouteConfigForTest 는 dir 에 시장 하나의 서명 route 매니페스트(0400)를 쓰고, 그것을 journalPath 의 원장과 함께 검증할 config 를
// 돌려줌. journalSchema 는 config 에 주입할 원장 스키마 버전 — 호출자(외부 시험)가 `journal.SchemaVersion` 을 넘김.
func SignedProductionRouteConfigForTest(dir, journalPath string, market Market, now time.Time, journalSchema int) (ProductionRouteConfig, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return ProductionRouteConfig{}, err
	}
	body := productionRouteBodyForTest(market, now)
	canonical, err := json.Marshal(body)
	if err != nil {
		return ProductionRouteConfig{}, err
	}
	data, err := json.Marshal(productionRouteManifest{productionRouteBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))})
	if err != nil {
		return ProductionRouteConfig{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, ProductionRouteFileName(market)), data, 0o400); err != nil {
		return ProductionRouteConfig{}, err
	}
	scope := body.Scopes[0]
	return ProductionRouteConfig{ConfigDir: dir, JournalPath: journalPath, AccountRef: body.AccountRef, Market: market, Symbol: scope.Symbol,
		PositionGeneration: scope.PositionGeneration, ManifestDigest: productionRouteDigest(data), TrustedKeyID: body.KeyID, TrustedKey: public,
		ObservedAt: now, ActivationDigest: body.ActivationDigest, CalendarGeneration: body.CalendarGeneration, CalendarDigest: body.CalendarDigest,
		SchedulerConfigVersion: body.ConfigVersion, ActivationExpiresAt: now.Add(time.Hour), JournalSchemaVersion: journalSchema}, nil
}
