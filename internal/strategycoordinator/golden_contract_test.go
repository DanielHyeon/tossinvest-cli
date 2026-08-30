package strategycoordinator

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// goldenPath 는 a112 의 동결 골든이다. 이 파일은 "내용을 바꾸려면 Manager 가 쓴
// OpenSpec amendment 와 새 manifest/receipt 가 필요하다"고 스스로 적고 있다.
const goldenPath = "../../openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens/four-family-runtime-v1.json"

type goldenFile struct {
	Descriptors []struct {
		Market string `json:"market"`
		Family string `json:"family"`
	} `json:"descriptors"`
	CoordinatorKeyFields []string `json:"coordinator_key_fields"`
	CoordinatorCount     int      `json:"coordinator_count"`
	WorkerCount          int      `json:"worker_count"`
	Queue                struct {
		DedupKeyFields []string `json:"dedup_key_fields"`
	} `json:"queue"`
}

func readGolden(t *testing.T) goldenFile {
	t.Helper()
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read the frozen golden: %v", err)
	}
	var golden goldenFile
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode the frozen golden: %v", err)
	}
	return golden
}

// 중복 제거 열쇠는 동결 골든이 적은 여덟 필드와 **정확히** 같아야 한다.
//
// 이 테스트가 있는 이유: 앞선 로트에서 구현이 설계 산문만 읽고 계약 이름을
// 지어냈고, 골든이 이미 이름을 정해 둔 것을 못 봤다. 골든을 파일에서 읽어
// 대조하면 사람이 옮겨 적을 자리가 없어 같은 실수가 반복될 수 없다.
func TestTheDedupKeyIsExactlyTheEightFieldsTheGoldenFroze(t *testing.T) {
	want := readGolden(t).Queue.DedupKeyFields
	got := DedupKeyFields()
	if len(want) != len(got) {
		t.Fatalf("the golden names %d dedup fields, the implementation names %d: %v vs %v",
			len(want), len(got), want, got)
	}
	for index := range want {
		if want[index] != got[index] {
			t.Fatalf("dedup key drift at %d: golden %q, implementation %q (golden=%v, implementation=%v)",
				index, want[index], got[index], want, got)
		}
	}
	// 이름만 맞고 값이 빠지면 이름표는 거짓말이 된다. 값의 개수도 같아야 한다.
	if parts := (Key{}).Parts(); len(parts) != len(got) {
		t.Fatalf("the key names %d fields but carries %d values", len(got), len(parts))
	}
}

// 여덟 필드는 저마다 열쇠를 갈라야 한다. 하나라도 열쇠에 안 들어가 있으면
// 서로 다른 제안이 같은 칸을 차지해 하나가 조용히 사라진다.
func TestEveryDedupFieldChangesTheKey(t *testing.T) {
	base := Key{Scope: ownerScopeForKeyTest("acct", "KR", "005930", 1),
		Family: "CONTINUATION", LaneID: "lane", LaneVersion: "v1", SnapshotDigest: "sha256:a"}
	mutations := map[string]Key{
		"account":             withScope(base, "other", "KR", "005930", 1),
		"market":              withScope(base, "acct", "US", "005930", 1),
		"symbol":              withScope(base, "acct", "KR", "000660", 1),
		"position_generation": withScope(base, "acct", "KR", "005930", 2),
		"family":              mutate(base, func(k *Key) { k.Family = "REVERSAL" }),
		"lane_id":             mutate(base, func(k *Key) { k.LaneID = "other-lane" }),
		"lane_version":        mutate(base, func(k *Key) { k.LaneVersion = "v2" }),
		"snapshot_digest":     mutate(base, func(k *Key) { k.SnapshotDigest = "sha256:b" }),
	}
	for _, field := range DedupKeyFields() {
		changed, ok := mutations[field]
		if !ok {
			t.Fatalf("no mutation covers the golden field %q", field)
		}
		if changed == base {
			t.Fatalf("changing %q left the dedup key unchanged", field)
		}
	}
}

// 골든이 조정자를 시장마다 하나씩 정확히 둘로, 레인을 시장마다 넷으로 정한다.
// 상수를 손으로 적지 않고 골든 파일에서 세어 대조한다.
func TestTheCoordinatorShapeMatchesTheGolden(t *testing.T) {
	golden := readGolden(t)
	if len(golden.CoordinatorKeyFields) != 1 || golden.CoordinatorKeyFields[0] != "market" {
		t.Fatalf("coordinator_key_fields=%v, want exactly [market]", golden.CoordinatorKeyFields)
	}
	perMarket := make(map[string]int, golden.CoordinatorCount)
	for _, descriptor := range golden.Descriptors {
		perMarket[descriptor.Market]++
	}
	if len(perMarket) != golden.CoordinatorCount {
		t.Fatalf("the golden describes %d markets but names %d coordinators", len(perMarket), golden.CoordinatorCount)
	}
	for market, count := range perMarket {
		if count != LanesPerMarket {
			t.Fatalf("the golden gives %s %d lanes, the implementation assumes %d", market, count, LanesPerMarket)
		}
	}
	if golden.WorkerCount != golden.CoordinatorCount*LanesPerMarket {
		t.Fatalf("worker_count=%d, but %d coordinators x %d lanes = %d", golden.WorkerCount,
			golden.CoordinatorCount, LanesPerMarket, golden.CoordinatorCount*LanesPerMarket)
	}
}

func ownerScopeForKeyTest(account, market, symbol string, generation uint64) strategyrouter.OwnerKey {
	return strategyrouter.OwnerKey{AccountRef: account, Market: strategyrouter.Market(market),
		Symbol: symbol, PositionGeneration: generation}
}

func withScope(base Key, account, market, symbol string, generation uint64) Key {
	base.Scope = ownerScopeForKeyTest(account, market, symbol, generation)
	return base
}

func mutate(base Key, apply func(*Key)) Key {
	apply(&base)
	return base
}
