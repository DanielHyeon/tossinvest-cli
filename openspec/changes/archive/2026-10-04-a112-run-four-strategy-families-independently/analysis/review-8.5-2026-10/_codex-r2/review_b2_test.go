//go:build tossos_testseams
package engine
import (
 "context"
 "testing"
 "time"
 "github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
 "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)
func TestReviewB2NilGetenv(t *testing.T) {
 now:=time.Date(2026,9,3,1,2,3,0,time.UTC)
 loader:=testStrategyProposalLoader(t)
 loader.getenv=nil
 routes:=arbitrationRoutePair(t,now,familyScoresForTest(strategyrouter.MarketKR),"005930",continuationlane.KRContinuationLaneID)
 fx:=proposalFXPair(now); fx.kr.snapshot.Ready=false
 got:=loader.collect(context.Background(),routeReadySchedulePair(now),routes,fx).kr
 t.Logf("NIL_GETENV reason=%s ready=%v entries=%d",got.snapshot.Reason,got.snapshot.Ready,len(got.entries))
 if got.snapshot.Ready || len(got.entries)!=0 {t.Fatal("closed market opened")}
}
func TestReviewB2CancelledAndDelayed(t *testing.T) {
 for _,mode:=range []string{"cancelled","delayed-ignore-cancel"} {t.Run(mode,func(t *testing.T){
 now:=time.Date(2026,9,3,1,2,3,0,time.UTC)
 loader:=testStrategyProposalLoader(t)
 routes:=arbitrationRoutePair(t,now,familyScoresForTest(strategyrouter.MarketKR),"005930",continuationlane.KRContinuationLaneID)
 fx:=proposalFXPair(now); fx.kr.snapshot.Ready=false
 ctx,cancel:=context.WithCancel(context.Background()); defer cancel()
 entered:=make(chan struct{},2); release:=make(chan struct{})
 loader.loadActivation=func(c context.Context,_ StrategyMarket,_ strategyScheduleMarketAuthority,_ strategyRouteMarketAuthority,_ time.Time)(strategyrouter.FamilyActivation,error){
 entered<-struct{}{}
 if mode=="delayed-ignore-cancel" {<-release}
 return strategyrouter.FamilyActivation{},c.Err()
 }
 if mode=="cancelled" {cancel()}
 done:=make(chan strategyProposalAuthorityPair,1)
 start:=time.Now()
 go func(){done<-loader.collect(ctx,routeReadySchedulePair(now),routes,fx)}()
 if mode=="delayed-ignore-cancel" {
 select {
 case got:=<-done: t.Logf("DELAY early_return=true reason=%s loads=%d",got.kr.snapshot.Reason,len(entered));close(release);return
 case <-entered:
 }
 cancel()
 select {case got:=<-done: close(release);t.Fatalf("unexpected bounded return after activation entered: %+v",got.kr.snapshot)
 case <-time.After(150*time.Millisecond):t.Log("DELAY held_after_cancel=true barrier=150ms")}
 close(release)
 }
 select {case got:=<-done:t.Logf("RESULT mode=%s reason=%s ready=%v entries=%d elapsed=%v",mode,got.kr.snapshot.Reason,got.kr.snapshot.Ready,len(got.kr.entries),time.Since(start))
 case <-time.After(2*time.Second):t.Fatal("collect did not return after release")}
 })}
}
