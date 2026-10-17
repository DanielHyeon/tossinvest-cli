//go:build !tossos_testseams

package engine

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
)

// loadShadow 는 shadow 단계가 매니페스트를 읽는 유일한 자리다(a112 7.3.1). 생산 빌드의 정의는 적재기 한 줄이고, 고장 주입 seam 은
// strategy_lane_shadow_load_testseam.go(tossos_testseams 빌드)에만 있다 — 생산 바이너리에 함수 필드를 두지 않는다(lane step seam 과 같은 규칙).
func (runtime *strategyLaneRuntime) loadShadow(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
	return strategyshadow.LoadProductionFamilyShadow(ctx, config)
}
