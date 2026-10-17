//go:build tossos_testseams
package engine
import("context";"testing";"time";"os";"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff")
func TestReviewB2ClosedDownstream(t *testing.T){
 now:=time.Date(2026,9,3,1,2,3,0,time.UTC);runtime,activation:=familyGateFixture(t)
 loader:=testStrategyProposalLoader(t).withStrategyLanes(runtime)
 loader.loadActivation=func(context.Context,StrategyMarket,strategyScheduleMarketAuthority,strategyRouteMarketAuthority,time.Time)(strategyrouter.FamilyActivation,error){return activation,nil}
 loader.load=func(_ context.Context,c strategyproposal.ProductionConfig,targets []strategyproposal.ProductionTarget,_ interfaceOfficialFX)(strategyproposal.ProductionBatchAuthority,error){return arbitrationBatch(t,c,targets,now,"005930",[]string{continuationlane.KRContinuationLaneID}),nil}
 routes:=arbitrationRoutePair(t,now,familyScoresForTest(strategyrouter.MarketKR),"005930",continuationlane.KRContinuationLaneID)
 fx:=proposalFXPair(now);forced:=os.Getenv("REVIEW_FORCED_BRANCH");if forced==""{fx.kr.snapshot.Ready=false}
 got:=loader.collect(context.Background(),routeReadySchedulePair(now),routes,fx).kr
 if got.snapshot.Ready || len(got.entries)!=0 {t.Fatalf("closed scope opened %+v",got.snapshot)}
 if forced!="" && (got.snapshot.Reason!=StrategyProposalInternalFailure || !got.familyActivation().Verified()){t.Fatalf("forced branch did not carry verified gate: %+v",got.snapshot)}
 callbacks:=0;for _,h:=range got.dispatchHandoffs(){if err:=h.Deliver(func(strategyhandoff.Delivered)error{callbacks++;return nil});err!=nil{t.Fatal(err)}}
 if callbacks!=0{t.Fatal("closed scope delivered to dispatch")}
 if err:=runtime.evaluate(context.Background(),StrategyMarketKR,1,got.familyActivation(),strategyLaneInputs("acct",got));err!=nil{t.Fatal(err)}
 on,emitted:=0,0;for _,o:=range runtime.observations(){if o.Key.Market!=strategyrouter.MarketKR{continue};if o.Effective==strategyrouter.StateOn{on++};if o.Emitted{emitted++}}
 if emitted!=0{t.Fatal("closed scope emitted lane proposal")}
 t.Logf("DOWNSTREAM branch=%s reason=%s detail=%s verified=%v entries=%d on=%d emitted=%d dispatch_callbacks=%d",forced,got.snapshot.Reason,got.snapshot.ArbitrationDetail,got.familyActivation().Verified(),len(got.entries),on,emitted,callbacks)
}
