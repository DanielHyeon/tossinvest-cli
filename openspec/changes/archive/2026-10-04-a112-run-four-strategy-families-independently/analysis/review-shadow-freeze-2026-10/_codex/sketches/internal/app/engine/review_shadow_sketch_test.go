//go:build tossos_testseams
package engine
import("context";"testing";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection";"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane";"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter")
func TestReviewShadowInputTap(t *testing.T){
 lanes,_:=familyGateFixture(t)
 for _,tc:=range []struct{name string;err error}{{"undeclared",strategyrouter.ErrProductionFamilyActivationUndeclared},{"declared_closed",strategyrouter.ErrProductionFamilyActivationUnavailable}} {
 a:=collectUnderLoad(t,strategyrouter.FamilyActivation{},tc.err,lanes,"005930",continuationlane.KRContinuationLaneID,reversallane.KRReversalLaneID)
 inputs:=strategyLaneInputs("acct",a)
 names:=[]string{};for _,i:=range inputs{names=append(names,i.Proposal.Result.Lineage.Symbol+":"+i.Proposal.Result.Lineage.LaneID)}
 t.Logf("%s: reason=%s lane inputs=%v",tc.name,a.snapshot.Reason,names)
 if tc.name=="declared_closed"&&len(inputs)!=0{t.Fatal("expected filtered inputs")}
 if tc.name=="undeclared"&&len(inputs)!=2{t.Fatalf("expected one winner per 2 symbols, got %d",len(inputs))}
 }
}

func TestReviewReadValidationBoundary(t *testing.T){
 c,lanes,_:=a112LaneProjectionContext(t)
 if err:=lanes.evaluate(context.Background(),StrategyMarketKR,1,strategyrouter.FamilyActivation{},nil);err!=nil{t.Fatal(err)}
 for k,o:=range lanes.observed{o.Desired="INVALID_REVIEW_STATE";lanes.observed[k]=o;break}
 got,err:=c.Read(context.Background());if err!=nil{t.Fatalf("Read validated unexpectedly: %v",err)}
 if err:=strategyprojection.Validate(got);err==nil{t.Fatal("control validator did not reject")}else{t.Logf("Context.Read returned nil error after dynamic overlay; explicit Validate rejected: %v",err)}
}
