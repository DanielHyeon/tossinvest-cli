//go:build tossos_testseams

package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 8.7.2 를 값으로 잰다: 선언된 활성화가 없거나 만료·폐기되면 그
// 시장의 entry worker 넷을 OFF 로 되돌리고, 공유 안전·계보 상태는 건드리지 않는다.
//
// **닫는 구멍.** 8.7.1 의 관문은 로드 오류를 종류 없이 "관문 없음" 으로 접었다. 관문이
// 없으면 기존 시장 단위 경로가 돌므로, 사람이 매니페스트에서 끈 가족이 활성화가 만료·
// 폐기되는 순간 기존 경로로 **되살아났다**. 만료가 진입을 넓힌 것이다.
//
// **판별자는 배포 핀의 존재다** (Manager 승인, 2026-09-27). 핀이 없으면 4-가족
// 런타임이 배포되지 않은 것이고 기존 경로가 그대로 돈다 — 오늘 생산(핀 0 건, 측정)의
// 동작이다. 핀이 있으면 무엇이 틀렸든 네 가족은 OFF 이고, 관문이 범위를 전부 지워
// 그 시장의 신규 진입만 닫는다(FAMILY_GATE_CLOSED).

const (
	rollbackRouteDigest = "sha256:" + "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	rollbackRiskDigest  = "sha256:" + "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// rollbackLoader 는 **생산 활성화 로더**(seam nil)가 도는 KR 로더다. env 는 제안
// 권한 시험의 것에 두 줄(가족 핀·위험 정책 핀)을 더한다.
func rollbackLoader(t *testing.T, now time.Time, lanes *strategyLaneRuntime, extra map[string]string) *strategyProposalAuthorityLoader {
	t.Helper()
	loader := testStrategyProposalLoader(t).withStrategyLanes(lanes)
	base := loader.getenv
	loader.getenv = func(name string) string {
		if value, ok := extra[name]; ok {
			return value
		}
		return base(name)
	}
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig,
		targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX,
	) (strategyproposal.ProductionBatchAuthority, error) {
		return arbitrationBatch(t, config, targets, now, "005930",
			[]string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID}), nil
	}
	if loader.loadActivation != nil {
		t.Fatal("이 시험은 생산 활성화 로더가 도는 것을 잰다 — seam 이 서 있으면 아무것도 재지 않는다")
	}
	return loader
}

func rollbackRoutes(t *testing.T, now time.Time) strategyRouteAuthorityPair {
	t.Helper()
	routes := arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	routes.kr.snapshot.ManifestDigest = rollbackRouteDigest
	return routes
}

// writeFamilyActivation 은 **도구가 쓰는 바로 그 인코더**로 매니페스트 바이트를 만들어
// 설정 디렉터리에 `0400` 으로 쓰고, 그 바이트의 핀을 돌려준다.
//
// 결속 값은 엔진이 실제로 넘기는 값에서 온다(경로 digest·보정·달력·빌드·위험 핀).
// 하나라도 어긋나면 아래 "검증된" 대조군이 빨개지므로, 만료·폐기 행이 결속 불일치가
// 아니라 정말로 만료·폐기로 거절됐다는 것을 그 대조군이 보증한다.
func writeFamilyActivation(t *testing.T, dir string, now time.Time, issued, expires time.Time, revoked bool) string {
	t.Helper()
	return writeFamilyActivationCalibrated(t, dir, arbitrationCalibrationForTest.CalibrationDigest, issued, expires, revoked)
}

func writeFamilyActivationCalibrated(t *testing.T, dir, calibration string, issued, expires time.Time, revoked bool) string {
	t.Helper()
	data, err := strategyrouter.EncodeProductionFamilyActivation(strategyrouter.FamilyActivationDocument{
		Market: strategyrouter.MarketKR, Generation: 1,
		RouteManifestDigest: rollbackRouteDigest, CalibrationDigest: calibration,
		CalendarVersion: "calendar-" + string(StrategyMarketKR), RiskPolicyDigest: rollbackRiskDigest,
		BuildDigest: strategyRuntimeBuildDigest(), ProtectionReadyMinGeneration: 1,
		Actor: "test-human", ApprovedAt: issued, IssuedAt: issued, ExpiresAt: expires, Revoked: revoked,
		On: []strategyrouter.Family{strategyrouter.FamilyContinuation},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, strategyrouter.ProductionFamilyActivationFileName(strategyrouter.MarketKR))
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// 선언된 활성화가 쓸 수 없게 되면 그 시장은 기존 경로로 넓어지지 않고 닫힌다.
//
// **모든 행이 생산 활성화 로더를 지난다**(seam nil) — 시험이 조종석에 앉으면 "생산
// 로더를 건너뛴다" 변이가 통과한다. 행은 셋으로 갈린다.
//
//	핀 없음                → 기존 경로 (점수 1위 REVERSAL) — 오늘 생산
//	핀 + 검증됨(지속형만 ON) → 관문이 순위를 바꿈 (CONTINUATION) — 결속이 맞다는 대조군
//	핀 + 그 밖의 모든 결함   → 네 가족 OFF, 시장 닫힘
//
// 편집 전 코드에서 셋째 줄의 행들은 전부 첫째 줄의 값(REVERSAL 선택)을 냈다 — 사람이
// 끈 REVERSAL 이 만료·폐기로 되살아나는 넓힘이다.
func TestADeclaredActivationThatLapsesRollsItsMarketBackInsteadOfWidening(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	const (
		legacy     = "legacy"
		promoted   = "promoted"
		rolledBack = "rolled back"
	)
	for name, entry := range map[string]struct {
		setup func(t *testing.T, dir string) map[string]string
		want  string
		err   error
	}{
		"핀 없음 — 4-가족 런타임 미배포": {
			setup: func(*testing.T, string) map[string]string { return map[string]string{} },
			want:  legacy, err: strategyrouter.ErrProductionFamilyActivationUndeclared},
		"핀 + 검증된 활성화": {
			setup: func(t *testing.T, dir string) map[string]string {
				pin := writeFamilyActivation(t, dir, now, now.Add(-time.Hour), now.Add(time.Hour), false)
				return map[string]string{strategyFamilyActivationKRManifestDigestEnv: pin}
			},
			want: promoted},
		"핀 + 만료": {
			setup: func(t *testing.T, dir string) map[string]string {
				pin := writeFamilyActivation(t, dir, now, now.Add(-2*time.Hour), now.Add(-time.Hour), false)
				return map[string]string{strategyFamilyActivationKRManifestDigestEnv: pin}
			},
			want: rolledBack, err: strategyrouter.ErrProductionFamilyActivationExpired},
		"핀 + 폐기": {
			setup: func(t *testing.T, dir string) map[string]string {
				pin := writeFamilyActivation(t, dir, now, now.Add(-time.Hour), now.Add(time.Hour), true)
				return map[string]string{strategyFamilyActivationKRManifestDigestEnv: pin}
			},
			want: rolledBack, err: strategyrouter.ErrProductionFamilyActivationRevoked},
		"핀 + 파일 없음": {
			setup: func(*testing.T, string) map[string]string {
				return map[string]string{strategyFamilyActivationKRManifestDigestEnv: "sha256:" + strings.Repeat("c", 64)}
			},
			want: rolledBack, err: strategyrouter.ErrProductionFamilyActivationUnavailable},
		"핀 + 핀 뒤에 바뀐 바이트": {
			setup: func(t *testing.T, dir string) map[string]string {
				writeFamilyActivation(t, dir, now, now.Add(-time.Hour), now.Add(time.Hour), false)
				return map[string]string{strategyFamilyActivationKRManifestDigestEnv: "sha256:" + strings.Repeat("c", 64)}
			},
			want: rolledBack, err: strategyrouter.ErrProductionFamilyActivationUnavailable},
	} {
		t.Run(name, func(t *testing.T) {
			runtime, _ := familyGateFixture(t)
			probe := testStrategyProposalLoader(t)
			env := entry.setup(t, probe.configDir)
			env[strategyRiskKRManifestDigestEnv] = rollbackRiskDigest
			loader := rollbackLoader(t, now, runtime, env)
			loader.configDir = probe.configDir
			routes := rollbackRoutes(t, now)

			activation, err := loader.loadFamilyActivation(context.Background(), StrategyMarketKR,
				routeReadySchedulePair(now).forMarket(StrategyMarketKR), routes.forMarket(StrategyMarketKR), now)
			if entry.err == nil && err != nil {
				t.Fatalf("검증되어야 할 활성화가 거절됐다: %v — 대조군이 무너지면 아래 행들이 무엇 때문에 거절됐는지 모른다", err)
			}
			if entry.err != nil && !errors.Is(err, entry.err) {
				t.Fatalf("로드 오류=%v, want %v — 이 행은 다른 축을 잰다", err, entry.err)
			}
			if (err == nil) != activation.Verified() {
				t.Fatalf("오류=%v 인데 검증=%v", err, activation.Verified())
			}

			authority := loader.collect(context.Background(), routeReadySchedulePair(now), routes,
				proposalFXPair(now)).forMarket(StrategyMarketKR)
			switch entry.want {
			case legacy:
				if got := selectedFamily(t, authority, "005930"); got != strategyrouter.FamilyReversal {
					t.Fatalf("선택=%s, want REVERSAL — 핀이 없는 시장은 오늘과 같아야 한다", got)
				}
				if authority.snapshot.GatedCount != 0 {
					t.Fatalf("미배포 시장에서 관문이 %d 건을 멈췄다", authority.snapshot.GatedCount)
				}
			case promoted:
				if got := selectedFamily(t, authority, "005930"); got != strategyrouter.FamilyContinuation {
					t.Fatalf("선택=%s, want CONTINUATION — 지속형만 켠 활성화가 순위를 바꿔야 한다", got)
				}
			case rolledBack:
				if authority.snapshot.Ready {
					t.Fatalf("시장이 열려 있다 (선택=%d 건) — 선언된 활성화가 쓸 수 없는데 기존 경로로 넓어졌다",
						len(authority.entries))
				}
				if authority.snapshot.Reason != StrategyProposalFamilyGateClosed {
					t.Fatalf("reason=%s, want %s", authority.snapshot.Reason, StrategyProposalFamilyGateClosed)
				}
				want := []string{string(strategyworker.OutcomeDormant)}
				if strings.Join(authority.snapshot.GatedOutcomes, ",") != strings.Join(want, ",") {
					t.Fatalf("멈춘 종류=%v, want %v — 되돌린 레인은 '안 켰다' 이지 고장이 아니다",
						authority.snapshot.GatedOutcomes, want)
				}
				if authority.familyActivation().Verified() {
					t.Fatal("되돌린 시장이 검증된 활성화를 싣고 나갔다")
				}
				if _, handedOff := authority.dispatchHandoff().Single(); handedOff {
					t.Fatal("되돌린 시장이 공유 dispatch 경계에 값을 건넸다")
				}
			}
		})
	}
}

// 되돌림은 레인의 잠금을 건드리지 않는다.
//
// 잠긴 레인이 되돌림 동안 풀리면 활성화가 다시 선 순간 고장 난 가족이 복구 증거 없이
// 진입을 연다. 그래서 순서대로 잰다: 잠그고 → 되돌리고(시장 닫힘, LATCHED 가 보임) →
// 다시 검증된 활성화 → 잠긴 가족은 여전히 멈추고 이웃이 이긴다.
func TestRollingBackLeavesALatchedLaneLatched(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	runtime, all := familyGateFixture(t)
	for _, lane := range runtime.lanesFor(StrategyMarketKR) {
		if lane.Key().Family != strategyrouter.FamilyReversal {
			continue
		}
		if _, locked := lane.Fail("a measured fault", true); !locked {
			t.Fatal("비정상 실패 하나로 레인이 잠기지 않았다")
		}
	}
	probe := testStrategyProposalLoader(t)
	loader := rollbackLoader(t, now, runtime, map[string]string{
		strategyFamilyActivationKRManifestDigestEnv: "sha256:" + strings.Repeat("c", 64),
		strategyRiskKRManifestDigestEnv:             rollbackRiskDigest})
	loader.configDir = probe.configDir
	rolled := loader.collect(context.Background(), routeReadySchedulePair(now), rollbackRoutes(t, now),
		proposalFXPair(now)).forMarket(StrategyMarketKR)
	if rolled.snapshot.Ready || rolled.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("되돌린 시장: ready=%v reason=%s", rolled.snapshot.Ready, rolled.snapshot.Reason)
	}
	kinds := append([]string(nil), rolled.snapshot.GatedOutcomes...)
	sort.Strings(kinds)
	if strings.Join(kinds, ",") != string(strategyworker.OutcomeDormant)+","+string(strategyworker.OutcomeLatched) {
		t.Fatalf("멈춘 종류=%v, want [DORMANT LATCHED] — 잠긴 레인이 '안 켰다' 로 보이면 복구 증거가 필요한 상태가 사라진다", kinds)
	}
	for _, lane := range runtime.lanesFor(StrategyMarketKR) {
		if lane.Key().Family == strategyrouter.FamilyReversal && lane.Health() != strategyworker.LaneLatched {
			t.Fatalf("되돌림 뒤 역전형 레인 상태=%s, want LATCHED", lane.Health())
		}
	}
	// 활성화가 다시 서도 잠금은 그대로다.
	authority := collectUnderGate(t, all, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if got := selectedFamily(t, authority, "005930"); got != strategyrouter.FamilyContinuation {
		t.Fatalf("선택=%s, want CONTINUATION — 되돌림이 잠금을 풀었다", got)
	}
}

// 되돌림 경로는 **읽기만** 한다.
//
// 되돌림은 진입만 닫아야 한다. 이 네 함수가 원장·게이트웨이·레인 잠금·취소 함수를
// 부를 수 있으면 "safety 루프와 계보를 보존한다" 는 문장이 코드가 아니라 기대가 된다.
// 금지 목록이 아니라 **허용 목록**으로 센다 — 금지 목록은 새 철자를 못 본다.
func TestTheRollbackPathOnlyReads(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_family_activation.go"))
	allowed := map[string]bool{
		// installed
		"gate.activation.Verified": true,
		// admit — 되돌린 관문이 실제로 레인을 만지는 자리다. 묻기(Owns)와 돌리기(Run)만 허용한다.
		"gate.installed": true, "lane.Owns": true, "lane.Run": true,
		// familyGateFor
		"load": true, "errors.Is": true, "activation.Verified": true, "loader.lanes.lanesFor": true,
		// loadFamilyActivation
		"strategyMarketCalibrationDigest": true, "strategyrouter.LoadProductionFamilyActivation": true,
		"strategyRouterMarket": true, "strings.TrimSpace": true, "loader.getenv": true, "strategyRuntimeBuildDigest": true,
	}
	seen := 0
	for _, name := range []string{"installed", "admit", "familyGateFor", "loadFamilyActivation"} {
		decl := engineFuncDecl(t, file, name)
		ast.Inspect(decl.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee := calleeText(call.Fun)
			seen++
			if !allowed[callee] {
				t.Errorf("%s 가 허용 목록 밖의 %q 를 부른다 — 되돌림 경로가 읽기 밖의 일을 할 수 있다", name, callee)
			}
			return true
		})
	}
	if seen == 0 {
		t.Fatal("센 호출이 0 이다 — 이 시험은 아무것도 재지 않는다")
	}
}

func calleeText(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return calleeText(value.X) + "." + value.Sel.Name
	default:
		return "<dynamic>"
	}
}

// 주문 lease 는 4-가족 활성화의 남은 수명을 넘어 살지 못한다.
//
// 네 행을 함께 세운다: 활성화 없음(기준 TTL D0), 수명이 긴 활성화(D0 그대로 — 늘지
// 않는다), 수명 10초(min(D0, 10초) — 줄어든다), 수명 0(admission **앞**에서 거절,
// 주문 0, 캠페인 FLAT). 하나만 세우면 "언제나 깎는다" 나 "언제나 거절한다" 판본이 통과한다.
func TestTheOrderLeaseCannotOutliveTheFamilyActivation(t *testing.T) {
	measureIn := func(t *testing.T, market StrategyMarket, expiresIn *time.Duration) (time.Duration, int, error) {
		t.Helper()
		cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
		authority := proposals.forMarket(market)
		if expiresIn != nil {
			routerMarket := strategyRouterMarket(market)
			authority.activation = strategyrouter.FamilyActivationExpiringForTest(routerMarket, 1,
				strategyrouter.AllFourFamiliesForTest(routerMarket), 1, proposals.observedAt.Add(*expiresIn))
			if !authority.activation.Verified() {
				t.Fatal("시험용 활성화가 검증되지 않았다")
			}
		}
		if market == StrategyMarketKR {
			proposals.kr = authority
		} else {
			proposals.us = authority
		}
		cycle.proposals = proposals
		result, handedOff := authority.dispatchHandoff().Single()
		if !handedOff {
			t.Fatal("배선이 건네줄 제안을 만들지 못했다")
		}
		_, err := cycle.dispatch(context.Background(), deliverForTest(t, result))
		spy.mu.Lock()
		defer spy.mu.Unlock()
		if len(spy.calls) == 0 {
			cas, casErr := j.CurrentPositionCampaignCAS(context.Background(), result.Lineage.AccountRef,
				string(result.Lineage.Market), result.Lineage.Symbol)
			if casErr != nil || cas.Claimed || cas.State != "FLAT" {
				t.Fatalf("거절 뒤 캠페인=%+v err=%v, want 손대지 않은 FLAT — admission 앞에서 거절해야 한다", cas, casErr)
			}
			return 0, 0, err
		}
		lease, leaseErr := j.LookupStrategyDispatchLease(context.Background(), spy.calls[0].Lease.LeaseID)
		if leaseErr != nil {
			t.Fatal(leaseErr)
		}
		return lease.ExpiresAt.Sub(lease.IssuedAt), len(spy.calls), err
	}
	measure := func(t *testing.T, expiresIn *time.Duration) (time.Duration, int, error) {
		t.Helper()
		return measureIn(t, StrategyMarketKR, expiresIn)
	}
	base, placed, err := measure(t, nil)
	if err != nil || placed != 1 {
		t.Fatalf("활성화 없는 기준 dispatch: placed=%d err=%v", placed, err)
	}
	// 기준값을 **숫자로** 못 박는다. 같은 코드에서 잰 기준과 견주기만 하면 30초 상한을
	// 넓히는 변이가 기준과 함께 움직여 통과한다(8.7.2 적대 리뷰 B 의 B1).
	if base != 30*time.Second {
		t.Fatalf("기준 TTL=%v, want 30s — lease 절대 상한이 바뀌었다", base)
	}
	long, short, dead := time.Hour, 10*time.Second, time.Duration(0)
	if got, placed, err := measure(t, &long); err != nil || placed != 1 || got != base {
		t.Fatalf("수명 1시간: ttl=%v placed=%d err=%v, want %v — lease 가 늘거나 줄면 안 된다", got, placed, err, base)
	}
	if got, placed, err := measure(t, &short); err != nil || placed != 1 || got != short {
		t.Fatalf("수명 10초: ttl=%v placed=%d err=%v, want %v", got, placed, err, short)
	}
	if _, placed, err := measure(t, &dead); placed != 0 || !errors.Is(err, strategyrouter.ErrProductionFamilyActivationExpired) {
		t.Fatalf("수명 0: placed=%d err=%v, want 주문 0 + expired", placed, err)
	}
	// US 도 **자기 시장의** 활성화로 깎인다(8.7.2 적대 리뷰 B 의 B2 — 시장을 KR 로 고정한
	// 변이가 KR 만 재는 시험을 통과했다).
	if got, placed, err := measureIn(t, StrategyMarketUS, &short); err != nil || placed != 1 || got != short {
		t.Fatalf("US 수명 10초: ttl=%v placed=%d err=%v, want %v", got, placed, err, short)
	}
}

// 되돌린 관문은 레인이 없어도 선다.
//
// 레인이 없으면 선 관문의 admit 이 모든 제안을 주인 없음(REFUSED)으로 멈춘다 — 되돌림에서는
// 그것이 맞는 값이다. 레인 조건을 관문에 걸면 레인이 없는 로더에서 되돌림이 기존 경로
// 통과로 바뀐다(검증된 활성화의 같은 경우는
// `TestAVerifiedGateWithoutLanesClosesInsteadOfFallingBackToTheLegacyPath`).
func TestARolledBackGateClosesTheMarketEvenWithoutLanes(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	probe := testStrategyProposalLoader(t)
	loader := rollbackLoader(t, now, nil, map[string]string{
		strategyFamilyActivationKRManifestDigestEnv: "sha256:" + strings.Repeat("c", 64),
		strategyRiskKRManifestDigestEnv:             rollbackRiskDigest})
	loader.configDir = probe.configDir
	authority := loader.collect(context.Background(), routeReadySchedulePair(now), rollbackRoutes(t, now),
		proposalFXPair(now)).forMarket(StrategyMarketKR)
	if authority.snapshot.Ready || authority.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("레인 없는 되돌림: ready=%v reason=%s — 기존 경로로 통과했다", authority.snapshot.Ready, authority.snapshot.Reason)
	}
	if strings.Join(authority.snapshot.GatedOutcomes, ",") != string(strategyworker.OutcomeRefused) {
		t.Fatalf("멈춘 종류=%v, want [REFUSED] — 레인이 없으면 주인 없는 제안이다", authority.snapshot.GatedOutcomes)
	}
}

// 핀이 없는 시장은 보정 합의 여부와 상관없이 미선언이다.
//
// 앞 판본의 적재기는 보정이 합의되지 않으면 **핀을 보기 전에** Unavailable 을 돌려줬다.
// 새 판별에서 그 답은 "선언했는데 쓸 수 없다" 이므로, 그대로 두면 핀 없는 생산 시장이
// 보정 불일치 한 번으로 닫힌다.
func TestAnUndeclaredMarketStaysUndeclaredWhateverItsCalibrationSays(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	runtime, _ := familyGateFixture(t)
	loader := rollbackLoader(t, now, runtime, map[string]string{})
	routes := rollbackRoutes(t, now).forMarket(StrategyMarketKR)
	routes.entries = nil // 합의할 보정이 없다 → strategyMarketCalibrationDigest 가 거짓
	if _, agreed := strategyMarketCalibrationDigest(routes); agreed {
		t.Fatal("이 시험은 보정이 합의되지 않은 상태를 재는 것이다")
	}
	schedule := routeReadySchedulePair(now).forMarket(StrategyMarketKR)
	if _, err := loader.loadFamilyActivation(context.Background(), StrategyMarketKR, schedule, routes, now); !errors.Is(err,
		strategyrouter.ErrProductionFamilyActivationUndeclared) {
		t.Fatalf("err=%v, want undeclared", err)
	}
	if gate := loader.familyGateFor(context.Background(), StrategyMarketKR, schedule, routes, now); gate.installed() {
		t.Fatal("핀이 없는 시장에 관문이 섰다")
	}
}

// 적재기가 오류 없이 검증 안 된 값을 주면 그것도 되돌림이다.
//
// 생산 적재기는 그런 답을 내지 않는다(오류가 없으면 언제나 검증된 값이다). 그래도 그
// 조합이 기존 경로로 읽히면 적재기 쪽 결함 하나가 넓힘이 되므로 fail-closed 로 둔다.
func TestAnUnverifiedActivationWithoutAnErrorStillRollsBack(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	authority := collectUnderLoad(t, strategyrouter.FamilyActivation{}, nil, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if authority.snapshot.Ready || authority.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("ready=%v reason=%s — 검증 안 된 활성화가 기존 경로로 통과했다", authority.snapshot.Ready, authority.snapshot.Reason)
	}
}

// 핀이 있는 시장에서 보정이 합의되지 않으면 되돌림이다 — 그리고 그 거절은 이제
// strategyrouter 의 결속 형식 검사가 한다.
//
// 앞 판본의 엔진 적재기는 이 입력을 핀을 보기도 전에 Unavailable 로 막았다(그리고 관문이
// 그것을 "관문 없음" 으로 접어 기존 경로가 돌았다). 그 조기 반환을 들어낸 뒤 이 입력을
// 막는 것은 `LoadProductionFamilyActivation` 의 `productionRouteIdentity(config.CalibrationDigest)`
// 이다. 그 검사가 빠지면 **보정 값이 빈 매니페스트**가 빈 보정 결속과 등식으로 맞아
// 검증되어 버린다 — 이 시험이 그 매니페스트를 실제로 써서 그 문을 잰다.
func TestADeclaredMarketWhoseCalibrationDoesNotAgreeIsRolledBack(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	runtime, _ := familyGateFixture(t)
	probe := testStrategyProposalLoader(t)
	pin := writeFamilyActivationCalibrated(t, probe.configDir, "", now.Add(-time.Hour), now.Add(time.Hour), false)
	loader := rollbackLoader(t, now, runtime, map[string]string{
		strategyFamilyActivationKRManifestDigestEnv: pin, strategyRiskKRManifestDigestEnv: rollbackRiskDigest})
	loader.configDir = probe.configDir
	routes := rollbackRoutes(t, now).forMarket(StrategyMarketKR)
	routes.entries = nil
	schedule := routeReadySchedulePair(now).forMarket(StrategyMarketKR)
	activation, err := loader.loadFamilyActivation(context.Background(), StrategyMarketKR, schedule, routes, now)
	if !errors.Is(err, strategyrouter.ErrProductionFamilyActivationUnavailable) || activation.Verified() {
		t.Fatalf("err=%v verified=%v — 보정이 합의되지 않은 선언이 검증됐다", err, activation.Verified())
	}
	gate := loader.familyGateFor(context.Background(), StrategyMarketKR, schedule, routes, now)
	if !gate.installed() || !gate.rolledBack {
		t.Fatalf("installed=%v rolledBack=%v — 선언된 시장이 기존 경로로 돌아갔다", gate.installed(), gate.rolledBack)
	}
}

// SUBMITTING 은 4-가족 활성화가 만료된 뒤 최종 검사를 통과할 수 없다 (Codex P1 수정).
//
// lease TTL 은 상대 시간이고 원장은 그것을 자기 시각에 더하므로, lease 행의 명목 만료는
// 파도→발급 사이 δ 만큼 활성화 만료를 넘을 수 있다. 그 창을 닫는 것은 게이트웨이가 브로커
// 바이트 **전**에 부르는 `FinalAuthorityCheck` 다(`execgw.Gateway` 의 `call` 클로저; 오류면
// 브로커 0 건 — `TestStrategyGatewayRechecksSourceBackedActivationAfterSubmittingFencePairedKRUS` 가 잰다). 이 시험은
// dispatch 가 그 클로저에 가족 만료를 **실시계로** 싣는지를 잰다: 시계를 파도 뒤로 움직이며
// 같은 클로저가 만료 직전에는 통과하고 만료 순간부터 거절하는지.
func TestSubmittingCannotPassTheFinalCheckAfterTheFamilyActivationExpires(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	wave := proposals.observedAt
	current := wave.Add(5 * time.Second)
	cycle.now = func() time.Time { return current }
	kr := proposals.forMarket(StrategyMarketKR)
	kr.activation = strategyrouter.FamilyActivationExpiringForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR), 1, wave.Add(10*time.Second))
	proposals.kr = kr
	cycle.proposals = proposals
	result, handedOff := kr.dispatchHandoff().Single()
	if !handedOff {
		t.Fatal("배선이 건네줄 제안을 만들지 못했다")
	}
	if _, err := cycle.dispatch(context.Background(), deliverForTest(t, result)); err != nil {
		t.Fatalf("수명이 남은 활성화의 dispatch 가 거절됐다: %v", err)
	}
	spy.mu.Lock()
	calls := append([]execgw.StrategyPlaceRequest(nil), spy.calls...)
	spy.mu.Unlock()
	if len(calls) != 1 || calls[0].FinalAuthorityCheck == nil {
		t.Fatalf("게이트웨이 요청=%d 건 — 최종 검사를 잴 수 없다", len(calls))
	}
	final := calls[0].FinalAuthorityCheck
	for _, step := range []struct {
		at      time.Duration
		expired bool
	}{{9 * time.Second, false}, {10*time.Second - time.Nanosecond, false}, {10 * time.Second, true}, {11 * time.Second, true}} {
		current = wave.Add(step.at)
		err := final(context.Background())
		if step.expired != errors.Is(err, strategyrouter.ErrProductionFamilyActivationExpired) || (!step.expired && err != nil) {
			t.Fatalf("파도+%v: 최종 검사 err=%v, want expired=%v", step.at, err, step.expired)
		}
	}
}

// 수집 때는 유효했지만 admission 때 만료된 활성화는 admission **앞**에서 멈춘다.
// 그리고 실시계가 없는 dispatch 주기는 검증된 가족 활성화를 가진 시장에서만 거절한다.
func TestAnActivationThatExpiresAfterTheWaveStopsBeforeAdmission(t *testing.T) {
	type row struct {
		expiresIn time.Duration // 0 이면 활성화 없음
		now       func(wave time.Time) func() time.Time
		wantPlace int
	}
	for name, entry := range map[string]row{
		"파도 뒤 만료 — 실시계가 만료를 넘었다": {expiresIn: 10 * time.Second,
			now: func(wave time.Time) func() time.Time { return func() time.Time { return wave.Add(10 * time.Second) } }, wantPlace: 0},
		"실시계 없음 + 검증된 활성화 — fail-closed": {expiresIn: time.Hour, now: nil, wantPlace: 0},
		"실시계 없음 + 활성화 없음 — 기존 경로 그대로":    {expiresIn: 0, now: nil, wantPlace: 1},
	} {
		t.Run(name, func(t *testing.T) {
			cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
			wave := proposals.observedAt
			cycle.now = nil
			if entry.now != nil {
				cycle.now = entry.now(wave)
			}
			kr := proposals.forMarket(StrategyMarketKR)
			if entry.expiresIn != 0 {
				kr.activation = strategyrouter.FamilyActivationExpiringForTest(strategyrouter.MarketKR, 1,
					strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR), 1, wave.Add(entry.expiresIn))
			}
			proposals.kr = kr
			cycle.proposals = proposals
			result, _ := kr.dispatchHandoff().Single()
			_, err := cycle.dispatch(context.Background(), deliverForTest(t, result))
			spy.mu.Lock()
			placed := len(spy.calls)
			spy.mu.Unlock()
			if placed != entry.wantPlace {
				t.Fatalf("브로커 요청=%d, want %d (err=%v)", placed, entry.wantPlace, err)
			}
			if entry.wantPlace == 0 {
				if err == nil {
					t.Fatal("거절했는데 오류가 없다")
				}
				cas, casErr := j.CurrentPositionCampaignCAS(context.Background(), result.Lineage.AccountRef,
					string(result.Lineage.Market), result.Lineage.Symbol)
				if casErr != nil || cas.Claimed || cas.State != "FLAT" {
					t.Fatalf("거절 뒤 캠페인=%+v err=%v — admission 앞에서 멈춰야 한다", cas, casErr)
				}
			}
		})
	}
}

// dispatch 의 두 가족 만료 검사 자리를 구조로 못 박는다: admission 앞에 하나, 최종 검사
// 클로저 안에 하나. 행동 시험은 "검사가 있다" 를 재고 이것은 "그 자리에 있다" 를 잰다.
func TestTheDispatchJudgesFamilyExpiryBeforeAdmissionAndInsideTheFinalCheck(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_dispatch_cycle.go"))
	decl := engineMethodDecl(t, file, "strategyDispatchCycle", "dispatch")
	admitAt, ceilingBeforeAdmit, ceilingInFinal := token.NoPos, token.NoPos, false
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.KeyValueExpr:
			if key, ok := value.Key.(*ast.Ident); ok && key.Name == "FinalAuthorityCheck" {
				ast.Inspect(value.Value, func(inner ast.Node) bool {
					if call, ok := inner.(*ast.CallExpr); ok && strings.HasSuffix(calleeText(call.Fun), ".LeaseCeiling") {
						ceilingInFinal = true
					}
					return true
				})
				return false
			}
		case *ast.CallExpr:
			callee := calleeText(value.Fun)
			if callee == "cycle.firstLeg.admit" && admitAt == token.NoPos {
				admitAt = value.Pos()
			}
			if strings.HasSuffix(callee, ".LeaseCeiling") && ceilingBeforeAdmit == token.NoPos {
				ceilingBeforeAdmit = value.Pos()
			}
		}
		return true
	})
	if admitAt == token.NoPos || ceilingBeforeAdmit == token.NoPos || ceilingBeforeAdmit > admitAt {
		t.Fatalf("admission 앞 가족 만료 검사가 없다 (LeaseCeiling@%d admit@%d)", ceilingBeforeAdmit, admitAt)
	}
	if !ceilingInFinal {
		t.Fatal("FinalAuthorityCheck 클로저에 가족 만료 검사가 없다 — SUBMITTING 이 만료 뒤에도 최종 검사를 통과한다")
	}
}

func engineMethodDecl(t *testing.T, file *ast.File, receiver, name string) *ast.FuncDecl {
	t.Helper()
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Recv == nil || function.Name.Name != name || len(function.Recv.List) != 1 {
			continue
		}
		expression := function.Recv.List[0].Type
		if star, isStar := expression.(*ast.StarExpr); isStar {
			expression = star.X
		}
		if ident, isIdent := expression.(*ast.Ident); isIdent && ident.Name == receiver {
			return function
		}
	}
	t.Fatalf("%s.%s 가 없다", receiver, name)
	return nil
}

// 생산 조립이 dispatch 주기에 실시계를 넣는다.
//
// 넣지 않으면 검증된 가족 활성화를 가진 시장의 주문이 전부 거절된다(fail-closed) — 안전한
// 방향이지만 기능이 조용히 죽는다. 생산에서 `newStrategyDispatchCycle` 을 부르는 자리는 이
// 함수 하나다(측정: 비시험 파일 grep 1 건). 조립 전체를 행동으로 세우려면 원장·게이트웨이·
// 권한 여덟이 필요하므로, 그 대입 한 줄을 구조로 못 박는다.
func TestTheProductionAssemblyGivesTheDispatchCycleTheRealClock(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_entry_supervisor.go"))
	decl := engineMethodDecl(t, file, "Context", "NewPairedStrategyEntryProductionAssembly")
	found := 0
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		if calleeText(assign.Lhs[0]) == "dispatchCycle.now" && calleeText(assign.Rhs[0]) == "clk.Now" {
			found++
		}
		return true
	})
	if found != 1 {
		t.Fatalf("dispatchCycle.now = clk.Now 대입=%d 건, want 1", found)
	}
}

// 검증된 활성화의 관문도 레인이 없으면 기존 경로로 내려가지 않는다 (8.7.2 적대 리뷰 A).
//
// 8.7.1 의 관문은 "검증됨 && 레인 있음" 이었고 레인이 비면 기존 경로가 돌았다. 그 갈래에서
// 지속형만 켠 활성화가 역전형 선택을 허락했다 — 사람이 끈 가족이다.
func TestAVerifiedGateWithoutLanesClosesInsteadOfFallingBackToTheLegacyPath(t *testing.T) {
	onlyContinuation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		map[string]bool{continuationlane.KRContinuationLaneID: true})
	authority := collectUnderLoad(t, onlyContinuation, nil, nil, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if authority.snapshot.Ready || authority.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("ready=%v reason=%s — 레인 없는 선언 시장이 기존 경로로 통과했다", authority.snapshot.Ready, authority.snapshot.Reason)
	}
}

// 핀이 있는 시장은 주기가 취소돼도 기존 경로로 가지 않는다 (8.7.2 적대 리뷰 B 의 B3).
func TestADeclaredMarketWhoseCycleIsCancelledRollsBack(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	runtime, _ := familyGateFixture(t)
	probe := testStrategyProposalLoader(t)
	pin := writeFamilyActivation(t, probe.configDir, now, now.Add(-time.Hour), now.Add(time.Hour), false)
	loader := rollbackLoader(t, now, runtime, map[string]string{
		strategyFamilyActivationKRManifestDigestEnv: pin, strategyRiskKRManifestDigestEnv: rollbackRiskDigest})
	loader.configDir = probe.configDir
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	gate := loader.familyGateFor(cancelled, StrategyMarketKR, routeReadySchedulePair(now).forMarket(StrategyMarketKR),
		rollbackRoutes(t, now).forMarket(StrategyMarketKR), now)
	if !gate.installed() || !gate.rolledBack {
		t.Fatalf("installed=%v rolledBack=%v — 취소된 주기가 선언된 시장을 기존 경로로 돌렸다", gate.installed(), gate.rolledBack)
	}
}

// 최종 검사 안의 스케줄 재검증이 만료를 가로질러 걸려도 가족 만료가 잡힌다 (Codex 재리뷰 P1).
//
// 재검증은 달력을 다시 읽는 I/O 다. 가족 만료를 그 **앞**에서 보면, 재검증 동안 만료가
// 지나가도 클로저가 통과를 돌려준다. 여기서는 재검증이 시계를 만료 너머로 옮긴다.
func TestAFamilyActivationThatExpiresDuringTheScheduleRevalidationStillStopsTheOrder(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	wave := proposals.observedAt
	current := wave.Add(5 * time.Second)
	cycle.now = func() time.Time { return current }
	kr := proposals.forMarket(StrategyMarketKR)
	kr.activation = strategyrouter.FamilyActivationExpiringForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR), 1, wave.Add(10*time.Second))
	proposals.kr = kr
	cycle.proposals = proposals
	result, _ := kr.dispatchHandoff().Single()
	if _, err := cycle.dispatch(context.Background(), deliverForTest(t, result)); err != nil {
		t.Fatalf("수명이 남은 활성화의 dispatch 가 거절됐다: %v", err)
	}
	spy.mu.Lock()
	final := spy.calls[0].FinalAuthorityCheck
	spy.mu.Unlock()
	cycle.revalidateSchedule = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error {
		current = wave.Add(11 * time.Second) // 재검증 I/O 가 만료를 가로지른다
		return nil
	}
	current = wave.Add(9 * time.Second)
	if err := final(context.Background()); !errors.Is(err, strategyrouter.ErrProductionFamilyActivationExpired) {
		t.Fatalf("재검증 동안 만료된 활성화: err=%v, want expired — 가족 만료를 재검증 앞에서만 본다", err)
	}
}
