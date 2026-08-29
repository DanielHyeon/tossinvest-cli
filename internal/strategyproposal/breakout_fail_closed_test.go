package strategyproposal

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// 이 파일에는 빌드 태그가 없다. 태그가 붙으면 CI 가 이 검사를 한 번도 돌리지 않기 때문이다.
// 돌파-되돌림 레인은 필요한 증거(ATR·RVOL·윗꼬리·거래량 확장)가 아예 저장되지도 만들어지지도
// 않아서, 값을 지어내는 대신 입구에서 닫아 두었다. 아래 두 검사가 그 잠금을 지킨다.

// 돌파 레인은 다른 어떤 검사보다 먼저, 돌파 전용 사유로 거절되어야 한다.
func TestBuildLaneInputRefusesBreakoutLanesWithTheBreakoutReason(t *testing.T) {
	for _, laneID := range []string{breakoutlane.KRLaneID, breakoutlane.USLaneID} {
		_, weekly, err := buildLaneInput(context.Background(), ProductionConfig{}, productionScope{LaneID: laneID},
			strategyevidence.Snapshot{}, officialfx.Evidence{}, nil)
		if !errors.Is(err, ErrBreakoutEvidenceUnavailable) {
			t.Fatalf("레인 %q: 오류가 %v 인데 ErrBreakoutEvidenceUnavailable 이어야 한다", laneID, err)
		}
		if weekly != nil {
			t.Fatalf("레인 %q: 거절인데 주간 예약 결과가 새어 나왔다", laneID)
		}
	}
}

// 위 검사만으로는 "돌파라서 막혔다"를 증명하지 못한다. 똑같이 텅 빈 스코프를 돌파가 아닌
// 레인으로 넣으면 *다른* 사유로 거절되어야, 앞의 거절이 돌파 잠금 때문이었다고 말할 수 있다.
func TestBuildLaneInputRefusesNonBreakoutLanesForADifferentReason(t *testing.T) {
	_, _, err := buildLaneInput(context.Background(), ProductionConfig{}, productionScope{LaneID: continuationlane.KRContinuationLaneID},
		strategyevidence.Snapshot{}, officialfx.Evidence{}, nil)
	if !errors.Is(err, ErrProductionProposalUnavailable) {
		t.Fatalf("오류가 %v 인데 ErrProductionProposalUnavailable 이어야 한다", err)
	}
	if errors.Is(err, ErrBreakoutEvidenceUnavailable) {
		t.Fatal("돌파가 아닌 레인이 돌파 잠금 사유로 거절되었다 — 잠금이 레인을 가리지 않고 있다")
	}
}
