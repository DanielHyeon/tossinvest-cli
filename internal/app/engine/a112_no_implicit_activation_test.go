//go:build tossos_testseams

package engine

// a112 4.5(-d · -e) 감사 보강(Manager 판정 2026-10-04): 「암묵 desired/effective/LIVE 활성화 없음」 을 생산 적재기로 잰다.
//
// 감사(audit-3.8-4.5.md) 시점의 핀은 성분 층이었다 — 레인 런타임에 영값 활성화를 **직접** 넣어 DORMANT 를 봤고(TestEveryFamilyLaneIsDormant…),
// 「후보가 Desired/Effective ON 을 말하는 유효한 짝 매니페스트가 있어도 레인이 OFF 다」 를 실 적재기로 잰 시험은 없었다. 여기서는:
//   ① 실 원장(journal.Open)과 서명된 경로 매니페스트(KR · US, 네 가족, 후보 전부 Desired · Effective ON)를 실 LoadProductionRouteAuthority 로
//      적재한다 — 매니페스트가 정말 ON 을 말한다는 전제를 먼저 확인.
//   ② 활성화 핀이 없는 생산 적재기의 관문(familyGateFor — 실 loadFamilyActivation → LoadProductionFamilyActivation)은 미선언: 검증된 활성화 0,
//      되돌림도 아님(토글 OFF = upstream).
//   ③ 그 관문으로 레인 여덟을 돌리면 전부 DORMANT · 봉투 0, 투영은 여덟 행 OFF/OFF/UNOBSERVED.
//   ④ 그 사이 설정 디렉터리 · 원장 파일의 바이트가 한 바이트도 바뀌지 않는다 — 활성화 매니페스트도, 운영 설정(LIVE 토글 포함)도 쓰이지 않음.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112TreeDigest 는 디렉터리 아래 모든 정규 파일의 「경로 · 바이트 digest」 목록이다.
func a112TreeDigest(t *testing.T, root string) []string {
	t.Helper()
	var entries []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, _ := filepath.Rel(root, path)
		entries = append(entries, rel+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(entries)
	return entries
}

func TestAPairedFourFamilyRouteManifestThatSaysOnPromotesNoLaneWithoutAnActivation(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.db")
	writable, err := journal.Open(context.Background(), journal.Options{Path: journalPath,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}

	// ① 짝 매니페스트가 정말 ON 을 말한다.
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		config, err := strategyrouter.SignedProductionRouteConfigForTest(dir, journalPath, market, now, journal.SchemaVersion)
		if err != nil {
			t.Fatal(err)
		}
		authority, err := strategyrouter.LoadProductionRouteAuthority(context.Background(), config)
		if err != nil {
			t.Fatalf("%s: LoadProductionRouteAuthority: %v", market, err)
		}
		candidates := authority.Request().Candidates
		lanes := map[string]bool{}
		for _, candidate := range candidates {
			if candidate.Desired != strategyrouter.StateOn || candidate.Effective != strategyrouter.StateOn {
				t.Fatalf("arrangement: %s candidate %s is not Desired/Effective ON — the premise would be vacuous", market, candidate.LaneID)
			}
			lanes[candidate.LaneID] = true
		}
		if len(candidates) != 4 || len(lanes) != 4 || len(authority.FamilyScores()) != 4 {
			t.Fatalf("arrangement: %s manifest has %d candidates, %d lanes, %d family scores — want four families", market, len(candidates), len(lanes), len(authority.FamilyScores()))
		}
	}

	// 생산에서 이 설정 디렉터리(c.Paths.ConfigDir)는 운영 설정 파일(trading.allow_live_order_actions — LIVE 토글)도 담는다. 같은 자리에 꺼진
	// 운영 설정을 두고 바이트가 그대로인지 잰다.
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"trading":{"allow_live_order_actions":false}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	before, beforeJournal := a112TreeDigest(t, dir), a112TreeDigest(t, filepath.Dir(journalPath))
	if len(before) < 4 { // config.json · 경로 매니페스트 둘 · 원장(같은 디렉터리)
		t.Fatalf("arrangement: the config tree holds %d files — the byte comparison would read nothing", len(before))
	}

	// ② 활성화 핀 없는 생산 적재기(실 loadFamilyActivation)의 관문은 미선언이다.
	runtime, _ := laneRuntimeFixture(t)
	env := map[string]string{} // 활성화 digest 핀 없음 — 오늘 생산과 같다
	loader := newStrategyProposalAuthorityLoader(dir, filepath.Join(dir, "evidence.db"), journalPath, "acct",
		func(key string) string { return env[key] }).withStrategyLanes(runtime)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		gate := loader.familyGateFor(context.Background(), market, routeReadySchedulePair(now).forMarket(market),
			strategyRouteMarketAuthority{market: market}, now)
		if gate.activation.Verified() || gate.rolledBack || gate.installed() {
			t.Fatalf("%s: a manifest that says ON produced gate verified=%v rolledBack=%v installed=%v — want undeclared",
				market, gate.activation.Verified(), gate.rolledBack, gate.installed())
		}
		// ③ 그 관문으로 레인을 돌린다.
		runtime.evaluate(context.Background(), market, 0, gate.activation, nil, strategyShadowBatch{})
	}
	observations := runtime.observations()
	if len(observations) != 8 {
		t.Fatalf("observations=%d, want 8", len(observations))
	}
	for _, observation := range observations {
		if observation.Outcome != strategyworker.OutcomeDormant || observation.Emitted {
			t.Fatalf("%v: outcome=%s emitted=%v — a lane woke up without a signed activation", observation.Key, observation.Outcome, observation.Emitted)
		}
	}
	rows := runtime.projection()
	if len(rows) != 8 {
		t.Fatalf("projection rows=%d, want 8", len(rows))
	}
	for _, row := range rows {
		if row.Desired != strategyprojection.StateOff || row.Effective != strategyprojection.StateOff || row.Runtime != strategyprojection.LaneRuntimeUnobserved {
			t.Fatalf("%s projects %s/%s/%s, want OFF/OFF/UNOBSERVED", row.LaneID, row.Desired, row.Effective, row.Runtime)
		}
	}

	// ④ 설정 디렉터리(경로 매니페스트 · 활성화 매니페스트 자리 · 운영 설정 자리)와 원장 바이트가 그대로다.
	if after := a112TreeDigest(t, dir); !a112EqualStrings(before, after) {
		t.Fatalf("the config directory changed:\n before %v\n after  %v", before, after)
	}
	if after := a112TreeDigest(t, filepath.Dir(journalPath)); !a112EqualStrings(beforeJournal, after) {
		t.Fatal("the journal changed while no activation was declared")
	}
}

func a112EqualStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
