package engine

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
)

// a112 0.5 리뷰 시험#11: 엔진 쪽 진입 0 블록 — 부재 묶음은 결속을 붙여도 부재 값이고(조정 앞 닫힘 = 「관측 없음」),
// nil 런타임 · nil 시계 · nil Context 는 아무것도 띄우지 않음(무 panic).
func TestAnAbsentShadowBatchStaysAbsentAndNilHoldersStartNothing(t *testing.T) {
	config := strategyshadow.Config{Market: strategyrouter.MarketKR, ManifestDigest: "sha256:pin", ConfigDir: "/x"}
	if got := (strategyShadowBatch{}).boundTo(config); !reflect.DeepEqual(got, strategyShadowBatch{}) {
		t.Fatalf("an absent batch bound to a config became %+v — it must stay the absent value", got)
	}
	var nilContext *Context
	if nilContext.strategyLaneRuntimeIfAny() != nil {
		t.Fatal("a nil Context returned a lane runtime")
	}
	var nilRuntime *strategyLaneRuntime
	nilRuntime.startShadowStep(context.Background(), clock.NewFake(time.Unix(0, 0)), StrategyMarketKR)
	runtime := &strategyLaneRuntime{shadowInFlight: map[StrategyMarket]bool{}}
	runtime.startShadowStep(context.Background(), nil, StrategyMarketKR)
	if runtime.shadowInFlight[StrategyMarketKR] {
		t.Fatal("a nil clock started a shadow step")
	}
}
