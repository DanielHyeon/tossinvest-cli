package strategyrouter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 태스크 8.8.3 [a112 결정 61]: **검증기가 만들지 않은 바이트**를 읽는 유일한 시험이다.
//
// 왜 이 시험이 따로 있어야 하는가. 같은 패키지의 다른 로더 시험은 전부
// `json.Marshal(body)` 로 바이트를 만들고 검증기도 같은 호출을 한다. 그래서 "사람이
// 만들 수 있는 바이트" 와 "검증기가 받는 바이트" 가 갈려도 그 시험들은 전부 초록이다.
// 그리고 승격 경로 시험은 파일을 아예 안 지나간다(`FamilyActivationForTest` 가
// 구조체를 직접 주조한다). 즉 파일 → 핀 → 승격 경로는 이 시험 하나만 끝까지 돈다.
//
// **바이트가 먼저다.** 골든은 `tools/a112-family-activation` 이 낸 커밋된 바이트이고
// 이 시험은 그것을 재생성하지 않는다. 아래 digest 는 그 커밋된 파일의 SHA-256 을
// **손으로 적은 리터럴**이다 — 실행 중에 계산해서 자기 자신과 비교하면 어떤 바이트든
// 통과한다(8.7.1 이 그렇게 충족 불가능한 결속을 초록으로 통과시켰다).
const goldenFamilyActivationKRDigest = "sha256:" +
	"9056f9e53972240c2fa69a174f9e97582031ad8129a9f54b759f297f237f0dbe"

// 골든이 결속하는 다섯 값. 매니페스트 파일과 이 목록 **양쪽에** 적혀 있고, 시험은
// 파일에서 읽지 않는다. 읽으면 파일이 무엇을 말하든 결속이 성립해 버린다.
const (
	goldenFamilyActivationRouteDigest = "sha256:" +
		"0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
	goldenFamilyActivationRiskDigest = "sha256:" +
		"a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"
	goldenFamilyActivationCalibration = "sha256:calibration-golden-kr-v1"
	goldenFamilyActivationCalendar    = "kr-regular-2026.09"
	goldenFamilyActivationBuild       = "tossos-golden-build-1"
	goldenFamilyActivationGeneration  = 3
	goldenFamilyActivationFloor       = 4
)

// goldenFamilyActivationObservedAt 은 골든의 issued_at 과 expires_at 사이다.
var goldenFamilyActivationObservedAt = time.Date(2026, 9, 4, 1, 0, 0, 0, time.UTC)

// installGoldenFamilyActivation 은 커밋된 골든을 로더가 요구하는 자리·모드로 놓는다.
//
// 바이트를 손대지 않고 그대로 복사한다. 여기서 한 글자라도 다시 만들면 이 시험이
// 재는 것이 "커밋된 바이트" 가 아니라 "이 시험이 만든 바이트" 가 된다.
func installGoldenFamilyActivation(t *testing.T, market Market) (string, []byte) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", ProductionFamilyActivationFileName(market)))
	if err != nil {
		t.Fatalf("커밋된 골든 매니페스트를 읽지 못했다: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, ProductionFamilyActivationFileName(market))
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	return dir, data
}

func goldenFamilyActivationConfig(dir string) FamilyActivationConfig {
	return FamilyActivationConfig{
		ConfigDir: dir, Market: MarketKR,
		ManifestDigest:      goldenFamilyActivationKRDigest,
		ObservedAt:          goldenFamilyActivationObservedAt,
		RouteManifestDigest: goldenFamilyActivationRouteDigest,
		CalibrationDigest:   goldenFamilyActivationCalibration,
		CalendarVersion:     goldenFamilyActivationCalendar,
		BuildDigest:         goldenFamilyActivationBuild,
		RiskPolicyDigest:    goldenFamilyActivationRiskDigest,
	}
}

// 커밋된 골든의 바이트가 손으로 적은 핀과 같고, 그 바이트가 네 레인을 승격한다.
//
// 두 단언이 붙어 있는 것이 요점이다. digest 만 재면 "파일이 안 변했다" 이고 승격만
// 재면 "무언가 통과했다" 인데, 이 change 가 알아야 하는 것은 **핀이 가리키는 바로
// 그 바이트가 승격한다** 이다.
func TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes(t *testing.T) {
	dir, data := installGoldenFamilyActivation(t, MarketKR)
	digest := sha256.Sum256(data)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != goldenFamilyActivationKRDigest {
		t.Fatalf("골든 바이트의 digest 가 손으로 적은 핀과 다르다\n  파일 = %s\n  핀   = %s\n"+
			"골든을 다시 만들었다면 tools/a112-family-activation 이 낸 값을 이 리터럴에 옮겨 적을 것",
			got, goldenFamilyActivationKRDigest)
	}
	activation, err := LoadProductionFamilyActivation(context.Background(), goldenFamilyActivationConfig(dir))
	if err != nil {
		t.Fatalf("커밋된 골든이 거절됐다: %v", err)
	}
	if !activation.Verified() {
		t.Fatal("골든이 검증되지 않았다")
	}
	if activation.Generation() != goldenFamilyActivationGeneration {
		t.Fatalf("generation = %d, want %d", activation.Generation(), goldenFamilyActivationGeneration)
	}
	if activation.ProtectionReadyMinGeneration() != goldenFamilyActivationFloor {
		t.Fatalf("protection floor = %d, want %d",
			activation.ProtectionReadyMinGeneration(), goldenFamilyActivationFloor)
	}
	// 레인 ID 를 손으로 적지 않는다 — 적으면 표가 바뀔 때 이 시험이 옛 이름 넷을
	// 들고 "승격했다" 는 초록을 낸다.
	table := productionRouteDescriptors(MarketKR)
	if len(table) != 4 {
		t.Fatalf("KR 서술자 표가 넷이 아니다: %d", len(table))
	}
	for laneID, descriptor := range table {
		if got := activation.Effective(MarketKR, descriptor.Family, laneID, descriptor.LaneVersion); got != StateOn {
			t.Fatalf("레인 %s 의 effective = %v, want %v", laneID, got, StateOn)
		}
	}
}

// 골든이 받아들여진 바이트에는 열쇠도 서명도 없다 [a112 결정 61].
//
// 이 단언이 없으면 서명 필드를 되살린 판본이 위 시험을 통과한다(골든을 함께 다시
// 만들면 되므로). 결정이 바꾼 것은 값이 아니라 **무엇이 신뢰 앵커인가** 이고,
// 그것은 바이트에 무엇이 없는지로만 읽을 수 있다.
func TestTheAcceptedGoldenCarriesNoKeyAndNoSignature(t *testing.T) {
	_, data := installGoldenFamilyActivation(t, MarketKR)
	for _, banned := range []string{"signature", "key_id", "ed25519", "public_key"} {
		if strings.Contains(strings.ToLower(string(data)), banned) {
			t.Fatalf("골든 매니페스트에 %q 가 남아 있다 — 결정 61 은 신뢰 앵커를 digest 핀 하나로 두었다", banned)
		}
	}
}

// 핀을 정한 뒤 바이트가 바뀌면 아무것도 승격하지 않는다.
//
// 서명이 사라진 뒤로 digest 핀이 **유일한** 앵커다. 파일 쪽에서 재는 이유: config
// 쪽 불일치는 이미 다른 시험이 재고 있고, 실제로 일어나는 일은 배포 뒤 파일이
// 바뀌는 것이다.
func TestGoldenBytesThatDriftFromThePinPromoteNothing(t *testing.T) {
	dir, data := installGoldenFamilyActivation(t, MarketKR)
	path := filepath.Join(dir, ProductionFamilyActivationFileName(MarketKR))
	tampered := strings.Replace(string(data), goldenFamilyActivationCalendar, "kr-regular-2026.10", 1)
	if tampered == string(data) {
		t.Fatal("변조가 바이트를 안 바꿨다 — 이 시험은 아무것도 재지 않는다")
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(tampered), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionFamilyActivation(context.Background(),
		goldenFamilyActivationConfig(dir)); err == nil {
		t.Fatal("핀과 다른 바이트가 승격했다")
	}
}

// 골든을 만든 그 입력이다 — 손으로 적었다 [8.8.3 P1 수정, 2026-09-04 독립 적대 리뷰].
//
// 도구를 돌려 그 출력을 기대값으로 삼지 않는다. 그렇게 하면 저작 경로가 무엇을
// 내든 통과한다 — 8.7.1 이 실행 중인 시스템에서 기대값을 읽어 충족 불가능한
// 결속을 초록으로 통과시킨 것과 같은 모양이다.
const goldenFamilyActivationActor = "golden-fixture-approver"

var (
	goldenFamilyActivationApprovedAt = time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	goldenFamilyActivationIssuedAt   = time.Date(2026, 9, 4, 0, 30, 0, 0, time.UTC)
	goldenFamilyActivationExpiresAt  = time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)
)

func goldenFamilyActivationDocument() FamilyActivationDocument {
	return FamilyActivationDocument{
		Market: MarketKR, Generation: goldenFamilyActivationGeneration,
		RouteManifestDigest:          goldenFamilyActivationRouteDigest,
		CalibrationDigest:            goldenFamilyActivationCalibration,
		CalendarVersion:              goldenFamilyActivationCalendar,
		RiskPolicyDigest:             goldenFamilyActivationRiskDigest,
		BuildDigest:                  goldenFamilyActivationBuild,
		ProtectionReadyMinGeneration: goldenFamilyActivationFloor,
		Actor:                        goldenFamilyActivationActor,
		ApprovedAt:                   goldenFamilyActivationApprovedAt,
		IssuedAt:                     goldenFamilyActivationIssuedAt,
		ExpiresAt:                    goldenFamilyActivationExpiresAt,
		On: []Family{FamilyContinuation, FamilyReversal,
			FamilyWeeklyValue, FamilyBreakoutRetest},
	}
}

// 저작 경로가 **지금도** 그 커밋된 바이트를 낸다.
//
// 왜 이것이 위 시험과 별개인가. 위 시험은 커밋된 바이트를 **검증기가 받아들인다**
// 는 것만 잰다. 그것은 골든이 로더 직렬화기의 드리프트 검출기라는 뜻일 뿐,
// **저작 경로**에 대해서는 아무것도 말하지 않는다. 독립 적대 리뷰 전까지
// `EncodeProductionFamilyActivation` 을 부르는 시험이 하나도 없었다 — 참조는 셋
// (정의·도구·가드)이고 그중 호출은 도구 하나뿐이라, `body()` 의 서술자 유도와
// 정렬은 어느 시험에서도 실행되지 않았다. 그래서 정렬을 지운 변이가 스위트 전체를
// 초록으로 통과했다(review.md 의 M3 행을 그 측정으로 정정한다).
func TestTheAuthoringEncoderStillProducesTheCommittedGoldenBytes(t *testing.T) {
	_, committed := installGoldenFamilyActivation(t, MarketKR)
	produced, err := EncodeProductionFamilyActivation(goldenFamilyActivationDocument())
	if err != nil {
		t.Fatalf("저작 경로가 바이트를 못 냈다: %v", err)
	}
	if !bytes.Equal(produced, committed) {
		t.Fatalf("도구가 내는 바이트가 커밋된 골든과 다르다 — 사람이 만들 수 있는 "+
			"바이트와 검증기가 받는 바이트가 갈렸다\n  도구 = %s\n  골든 = %s",
			produced, committed)
	}
}

// 같은 문서는 몇 번을 돌려도 같은 바이트를 낸다.
//
// 결함이 사는 축은 값이 아니라 **순서**다. 서술자는 map(`productionRouteDescriptors`)
// 순회로 모이고 Go 는 range 마다 순서를 무작위로 돌리므로, 정렬이 없으면 같은 입력이
// 실행마다 다른 바이트를 낸다. 그러면 운영자가 배포한 매니페스트의 digest 를 그
// 입력으로부터 다시 유도할 수 없다 — 사후에 "이 핀이 무엇을 승인했는가" 를 답할
// 방법이 사라진다.
//
// 64 회인 이유: 네 서술자의 순열은 24 가지이므로 정렬이 없을 때 64 번이 전부 같은
// 순서로 나올 확률은 24^-63 이다. 하나로는 24 분의 1 로 우연히 초록이 된다.
func TestTheAuthoringEncoderOrdersDescriptorsTheSameWayEveryRun(t *testing.T) {
	document := goldenFamilyActivationDocument()
	first, err := EncodeProductionFamilyActivation(document)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt < 64; attempt++ {
		again, err := EncodeProductionFamilyActivation(document)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(again, first) {
			t.Fatalf("%d 번째 실행이 다른 바이트를 냈다 — 서술자 순서가 map 순회에 "+
				"기대고 있다\n  1회차 = %s\n  %d회차 = %s", attempt+1, first, attempt+1, again)
		}
	}
}
