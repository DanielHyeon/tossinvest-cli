package strategyarbiter

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// goldenPath 는 a112 의 동결 골든이다. 이 파일은 "내용을 바꾸려면 Manager 가 쓴
// OpenSpec amendment 와 새 manifest/receipt 가 필요하다"고 스스로 적고 있다.
const goldenPath = "../../openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens/four-family-runtime-v1.json"

// 구현이 내보내는 거절 코드는 동결 골든의 목록과 **정확히 같아야 한다.**
//
// 이 테스트가 있는 이유: 처음 구현은 설계 산문만 읽고 코드 이름을 지어냈고,
// 골든이 이미 이름을 정해 둔 것을 못 봤다. 골든을 파일에서 읽어 대조하면
// 같은 실수가 두 번 일어날 수 없다 — 사람이 옮겨 적을 자리가 없기 때문이다.
func TestTheRefusalContractIsExactlyTheFrozenGolden(t *testing.T) {
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read the frozen golden: %v", err)
	}
	var golden struct {
		RefusalEnums struct {
			Arbitration []string `json:"arbitration"`
		} `json:"refusal_enums"`
		Arbitration struct {
			RefusalMapping              map[string]string `json:"refusal_mapping"`
			SingletonWithoutCalibration string            `json:"singleton_without_calibration_refusal"`
			ScorePPM                    struct {
				Integer bool  `json:"integer"`
				Minimum int64 `json:"minimum"`
				Maximum int64 `json:"maximum"`
			} `json:"score_ppm"`
		} `json:"arbitration"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode the frozen golden: %v", err)
	}

	implemented := []string{string(RefusalUncalibrated), string(RefusalTie), string(RefusalMultipleOwner),
		string(RefusalStaleOwner), string(RefusalStaleEnvelope), string(RefusalSealMismatch)}
	want := append([]string(nil), golden.RefusalEnums.Arbitration...)
	sort.Strings(want)
	got := append([]string(nil), implemented...)
	sort.Strings(got)
	if len(want) != len(got) {
		t.Fatalf("the golden names %d arbitration refusals, the implementation exports %d: %v vs %v", len(want), len(got), want, got)
	}
	for index := range want {
		if want[index] != got[index] {
			t.Fatalf("refusal contract drift at %d: golden %q, implementation %q (golden=%v, implementation=%v)",
				index, want[index], got[index], want, got)
		}
	}

	// 매핑표가 이름 붙인 코드도 모두 구현에 있어야 한다.
	exported := make(map[string]bool, len(implemented))
	for _, value := range implemented {
		exported[value] = true
	}
	for cause, code := range golden.Arbitration.RefusalMapping {
		if !exported[code] {
			t.Fatalf("the golden maps %q to %q, which the implementation does not export", cause, code)
		}
	}
	if golden.Arbitration.SingletonWithoutCalibration != string(RefusalUncalibrated) {
		t.Fatalf("the golden refuses an uncalibrated singleton with %q, the implementation uses %q",
			golden.Arbitration.SingletonWithoutCalibration, RefusalUncalibrated)
	}
}
