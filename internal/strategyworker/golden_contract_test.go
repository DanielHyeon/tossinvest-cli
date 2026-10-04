package strategyworker

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// goldenPath 는 a112 의 동결 골든이다. 이 파일은 "내용을 바꾸려면 Manager 가 쓴
// OpenSpec amendment 와 새 manifest/receipt 가 필요하다"고 스스로 적고 있다.
//
// 구현이 설계 산문만 읽고 계약 이름을 지어낸 적이 있어서, 계약 값은 언제나
// 이 파일에서 읽어 대조한다. 사람이 옮겨 적을 자리가 없으면 같은 실수가
// 반복될 수 없다.
const goldenPath = "../../openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens/four-family-runtime-v1.json"

type goldenFile struct {
	Descriptors []struct {
		Market      string `json:"market"`
		Family      string `json:"family"`
		LaneID      string `json:"lane_id"`
		LaneVersion string `json:"lane_version"`
		Horizon     string `json:"horizon"`
		Desired     string `json:"desired"`
		Effective   string `json:"effective"`
		Runtime     string `json:"runtime"`
	} `json:"descriptors"`
	WorkerKeyFields []string `json:"worker_key_fields"`
	WorkerCount     int      `json:"worker_count"`
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
	if len(golden.Descriptors) == 0 || golden.WorkerCount == 0 || len(golden.WorkerKeyFields) == 0 {
		t.Fatal("the golden decoded empty, so every comparison below would pass for the wrong reason")
	}
	return golden
}

// worker 열쇠는 골든이 적은 **네** 필드여야 한다.
//
// 조정자 큐의 여덟 필드 열쇠와 헷갈리기 쉬운 자리다. 골든은 둘을 다른
// 이름으로 따로 적어 두었고(`worker_key_fields` 와 `queue.dedup_key_fields`),
// 이 시험은 앞의 것만 본다.
func TestTheWorkerKeyIsExactlyTheFourFieldsTheGoldenFroze(t *testing.T) {
	want := readGolden(t).WorkerKeyFields
	got := KeyFields()
	if len(want) != len(got) {
		t.Fatalf("the golden names %d worker key fields, the implementation names %d: %v vs %v",
			len(want), len(got), want, got)
	}
	for index := range want {
		if want[index] != got[index] {
			t.Fatalf("worker key drift at %d: golden %q, implementation %q (golden=%v, implementation=%v)",
				index, want[index], got[index], want, got)
		}
	}
	// 이름만 맞고 값이 빠지면 이름표는 거짓말이 된다. 값의 개수도 같아야 한다.
	if parts := (Key{}).Parts(); len(parts) != len(got) {
		t.Fatalf("the key names %d fields but carries %d values", len(got), len(parts))
	}
}

// 생산 worker 는 골든이 적은 여덟이고, 순서·레인·지평·상태까지 같아야 한다.
//
// 개수만 세지 않는 이유: 여덟 개를 아무렇게나 넣어도 개수는 맞는다. 골든의
// descriptors 배열은 `operator_compatibility.additive_children` 이 말하는
// "lanes[8] fixed deterministic order" 그 자체이므로 순서까지 대조한다.
func TestProductionWorkersAreExactlyTheEightTheGoldenFroze(t *testing.T) {
	golden := readGolden(t)
	workers := ProductionWorkers()

	if golden.WorkerCount != len(golden.Descriptors) {
		t.Fatalf("the golden disagrees with itself: worker_count %d, descriptors %d",
			golden.WorkerCount, len(golden.Descriptors))
	}
	if len(workers) != golden.WorkerCount {
		t.Fatalf("the golden froze %d workers, the implementation builds %d",
			golden.WorkerCount, len(workers))
	}

	for index, want := range golden.Descriptors {
		got := workers[index]
		key := got.Key()
		if string(key.Market) != want.Market || string(key.Family) != want.Family ||
			key.LaneID != want.LaneID || key.LaneVersion != want.LaneVersion {
			t.Errorf("worker %d drifted from the golden: golden %s/%s/%s/%s, implementation %s/%s/%s/%s",
				index, want.Market, want.Family, want.LaneID, want.LaneVersion,
				key.Market, key.Family, key.LaneID, key.LaneVersion)
		}
		if string(got.Horizon()) != want.Horizon {
			t.Errorf("worker %d horizon: golden %q, implementation %q", index, want.Horizon, got.Horizon())
		}
		// desired/effective 는 여기서 단언하지 않는다(a112 8.8.4 항목 3). 앞 판은 영값 활성화로 `Desired(none)` 을 물었는데, 영값은
		// `lookup` 이 첫 줄에서 OFF 를 돌려주는 상수라 골든의 OFF 와 상수-대-상수였다(어떤 worker 열쇠 · 어떤 상태 표에도 같은 답 — 공허).
		// 골든의 OFF 가 **그 열쇠의 기본값**이고 서명 활성화만이 그것을 뒤집는다는 행동 단언은 태그 파일
		// `a112_golden_desired_effective_test.go` 에 있다(활성화 seam 이 태그 아래에 있다).
		if string(got.Runtime()) != want.Runtime {
			t.Errorf("worker %d runtime: golden %s, implementation %s", index, want.Runtime, got.Runtime())
		}
	}
}

// 여덟 열쇠는 서로 달라야 한다.
//
// 순서 대조만 있으면 같은 열쇠를 두 번 넣고 골든의 두 줄과 각각 맞춰 보는
// 실수를 못 잡는다 — 그러면 한 레인은 worker 가 둘이고 다른 레인은 없다.
func TestEveryProductionWorkerKeyIsDistinct(t *testing.T) {
	seen := make(map[Key]int)
	for index, worker := range ProductionWorkers() {
		if first, exists := seen[worker.Key()]; exists {
			t.Fatalf("workers %d and %d share one key: %v", first, index, worker.Key().Parts())
		}
		seen[worker.Key()] = index
	}
	if len(seen) != len(ProductionWorkers()) {
		t.Fatalf("counted %d distinct keys for %d workers", len(seen), len(ProductionWorkers()))
	}
}

// 시장과 가족은 라우터가 아는 값이어야 한다.
//
// 골든 대조는 문자열이 같은지만 본다. 문자열이 같아도 그 값이 이 모듈의
// 라우터에게 뜻이 없으면 worker 는 아무 데도 못 붙는다.
func TestEveryProductionWorkerUsesKnownRouterValues(t *testing.T) {
	for index, worker := range ProductionWorkers() {
		key := worker.Key()
		if key.Market != strategyrouter.MarketKR && key.Market != strategyrouter.MarketUS {
			t.Errorf("worker %d has a market the router does not know: %q", index, key.Market)
		}
		if !key.Family.Known() {
			t.Errorf("worker %d has a family the router does not know: %q", index, key.Family)
		}
		if worker.Horizon() != strategyrouter.HorizonShort && worker.Horizon() != strategyrouter.HorizonWeekly {
			t.Errorf("worker %d has a horizon the router does not know: %q", index, worker.Horizon())
		}
	}
}
