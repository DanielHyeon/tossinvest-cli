//go:build tossos_testseams
package engine
import("context";"testing";"time";"os";"path/filepath";"strings";"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter")
func TestReviewB2ProductionReadDelay(t *testing.T){
 now:=time.Date(2026,9,3,1,2,3,0,time.UTC);runtime,_:=familyGateFixture(t)
 dir:=t.TempDir();env:=map[string]string{strategyFamilyActivationKRManifestDigestEnv:"sha256:"+strings.Repeat("c",64),strategyRiskKRManifestDigestEnv:"sha256:"+strings.Repeat("e",64)}
 loader:=newStrategyProposalAuthorityLoader(dir,filepath.Join(dir,"evidence.db"),filepath.Join(dir,"journal.db"),"acct",func(n string)string{return env[n]}).withStrategyLanes(runtime)
 routes:=arbitrationRoutePair(t,now,familyScoresForTest(strategyrouter.MarketKR),"005930",continuationlane.KRContinuationLaneID);routes.kr.snapshot.ManifestDigest="sha256:"+strings.Repeat("d",64)
 fx:=proposalFXPair(now);fx.kr.snapshot.Ready=false
 marker:=filepath.Join(dir,strategyrouter.ProductionFamilyActivationFileName(strategyrouter.MarketKR));t.Setenv("REVIEW_READ_DELAY_PATH",marker)
 ctx,cancel:=context.WithCancel(context.Background());defer cancel()
 done:=make(chan strategyProposalAuthorityPair,1);go func(){done<-loader.collect(ctx,routeReadySchedulePair(now),routes,fx)}()
 deadline:=time.After(2*time.Second)
 for {
 select{case got:=<-done:t.Logf("PRODUCTION_IO entered=false reason=%s",got.kr.snapshot.Reason);return;case <-deadline:t.Fatal("neither reader entry nor return");default:}
 if _,err:=os.Stat(marker+".entered");err==nil{break};time.Sleep(time.Millisecond)
 }
 cancel()
 select{case <-done:t.Fatal("reader delay unexpectedly bypassed");case <-time.After(150*time.Millisecond):t.Log("PRODUCTION_IO entered=true held_after_cancel=true barrier=150ms")}
 if err:=os.WriteFile(marker+".release",[]byte("release"),0600);err!=nil{t.Fatal(err)}
 select{case got:=<-done:t.Logf("PRODUCTION_IO after_release reason=%s entries=%d",got.kr.snapshot.Reason,len(got.kr.entries));case <-time.After(2*time.Second):t.Fatal("release did not unblock")}
}
