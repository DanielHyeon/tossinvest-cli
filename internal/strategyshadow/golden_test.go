package strategyshadow

// a112 7.3.1 SHADOW 골든(브리프 v3.3 §1 — 8.8.3 「커밋된 골든은 소비자만 잰다」 교훈).
//
// 두 방향을 따로 잰다: (1) 커밋된 바이트(도구 `tools/a112-family-shadow` 가 낸 것)가 **손으로 적은** 핀과 같고 로더를 통과해 두 가족을
// shadow 한다(소비자). (2) 골든 문서를 **만드는** 함수(EncodeProductionFamilyShadow — 도구가 부르는 바로 그 함수)를 불러 같은 바이트가
// 나온다(생산자 — 정렬 삭제 같은 변이가 여기서 잡힌다).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// goldenFamilyShadowKRDigest 는 커밋된 골든 파일의 SHA-256 을 손으로 옮긴 리터럴이다(실행 중 계산 금지).
const goldenFamilyShadowKRDigest = "sha256:" +
	"09e183e5cb709defb8e5053ad0cd8c6787ce7863b1f23e2044f2723e84327e39"

// goldenFamilyShadowDocument 는 골든을 만든 도구 인자 그대로다(docs/operations.md 의 예시와 같은 값).
func goldenFamilyShadowDocument() Document {
	return Document{Market: strategyrouter.MarketKR, Generation: 1,
		RouteManifestDigest: shadowRouteDigest, CalibrationDigest: shadowCalibration, CalendarVersion: shadowCalendar,
		RiskPolicyDigest: shadowRiskDigest, BuildDigest: shadowBuild, Actor: "golden-operator",
		ApprovedAt: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), IssuedAt: time.Date(2026, 10, 5, 0, 30, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 10, 5, 23, 0, 0, 0, time.UTC),
		Shadow:    []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal}}
}

func readGoldenFamilyShadow(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", ProductionFamilyShadowFileName(strategyrouter.MarketKR)))
	if err != nil {
		t.Fatalf("committed golden shadow manifest: %v", err)
	}
	return data
}

func TestTheCommittedGoldenShadowManifestMatchesItsPinAndShadowsTwoLanes(t *testing.T) {
	data := readGoldenFamilyShadow(t)
	digest := sha256.Sum256(data)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != goldenFamilyShadowKRDigest {
		t.Fatalf("golden bytes digest %s differs from the hand-copied pin %s — copy the value the tool printed", got, goldenFamilyShadowKRDigest)
	}
	dir, _ := install(t, data, ProductionFamilyShadowFileName(strategyrouter.MarketKR))
	config := shadowConfig(dir, goldenFamilyShadowKRDigest)
	config.ObservedAt = time.Date(2026, 10, 5, 1, 0, 0, 0, time.UTC)
	shadow, err := LoadProductionFamilyShadow(context.Background(), config)
	if err != nil || !shadow.Verified() || shadow.Generation() != 1 {
		t.Fatalf("the committed golden was refused: verified=%v generation=%d err=%v", shadow.Verified(), shadow.Generation(), err)
	}
	count := 0
	for _, lane := range strategyrouter.SharedProductionRouteDescriptors(strategyrouter.MarketKR) {
		if shadow.Shadowed(strategyrouter.MarketKR, lane.Family, lane.LaneID, lane.LaneVersion) {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("the golden shadows %d lanes, want 2", count)
	}
}

func TestTheEncoderReproducesTheCommittedGoldenBytes(t *testing.T) {
	want := readGoldenFamilyShadow(t)
	got, err := EncodeProductionFamilyShadow(goldenFamilyShadowDocument())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("the encoder no longer produces the committed golden bytes\n got: %s\nwant: %s", got, want)
	}
	// 가족 순서가 입력 순서를 따라가지 않는다(정규 바이트는 입력 순서와 무관).
	reordered := goldenFamilyShadowDocument()
	reordered.Shadow = []strategyrouter.Family{strategyrouter.FamilyReversal, strategyrouter.FamilyContinuation}
	if again, err := EncodeProductionFamilyShadow(reordered); err != nil || !bytes.Equal(again, want) {
		t.Fatalf("reordered input changed the canonical bytes: err=%v", err)
	}
}
