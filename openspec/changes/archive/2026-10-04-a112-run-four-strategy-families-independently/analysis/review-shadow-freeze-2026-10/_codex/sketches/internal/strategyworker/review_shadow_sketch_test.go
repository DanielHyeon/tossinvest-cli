//go:build tossos_testseams
package strategyworker

import (
 "go/ast"
 "go/parser"
 "go/token"
 "testing"
 "github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
 "github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
 "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)
// Review-only surrogate for the proposed opaque router FamilyShadow. No activation state.
type reviewFamilyShadow struct { permitted bool }
type reviewShadowCycle struct { Outcome string }
func reviewLoadShadow() reviewFamilyShadow { return reviewFamilyShadow{permitted:true} }
func (worker FamilyWorker) reviewShadow(shadow reviewFamilyShadow,input Input) reviewShadowCycle {
 if !shadow.permitted {return reviewShadowCycle{Outcome:"NOT_SHADOWED"}}
 if !worker.owns(input.Proposal) {return reviewShadowCycle{Outcome:"NOT_THIS_LANE"}}
 return reviewShadowCycle{Outcome:"WOULD_EMIT"}
}
var reviewSink *strategycoordinator.MarketCoordinator
// Same receiver, arguments, result; string-based call allowlist sees worker.owns.
func (worker FamilyWorker) reviewShadowAttack(shadow reviewFamilyShadow,input Input) reviewShadowCycle {
 if !shadow.permitted {return reviewShadowCycle{Outcome:"NOT_SHADOWED"}}
 workerBinding := worker
 _ = workerBinding
 {
  worker := struct{owns func(strategycoordinator.Envelope) strategycoordinator.Admission}{owns:reviewSink.Submit}
  worker.owns(strategycoordinator.Envelope{Scope:input.Scope,SnapshotDigest:input.SnapshotDigest,Proposal:input.Proposal})
 }
 return reviewShadowCycle{Outcome:"WOULD_EMIT"}
}
func reviewCallee(e ast.Expr) string {switch v:=e.(type){case *ast.Ident:return v.Name;case *ast.SelectorExpr:return reviewCallee(v.X)+"."+v.Sel.Name;default:return "<dynamic>"}}
func TestReviewShadowStructure(t *testing.T){
 f:=newFixture(allKRFamilies()...);input:=f.input(t,continuationlane.KRContinuationLaneID)
 worker:=workerFor(t,strategyrouter.MarketKR,strategyrouter.FamilyContinuation)
 shadow:=reviewLoadShadow()
 if worker.reviewShadow(reviewFamilyShadow{},input).Outcome!="NOT_SHADOWED"{t.Fatal("zero shadow admitted")}
 if worker.Run(noActivation(),input).Outcome!=OutcomeDormant{t.Fatal("baseline not dormant")}
 if worker.reviewShadow(shadow,input).Outcome!="WOULD_EMIT"{t.Fatal("valid proposal missing")}
 // The declared signature has no activation input, so cannot distinguish ON/OFF.
 on:=promoting(t,strategyrouter.MarketKR,strategyrouter.FamilyContinuation)
 if worker.Effective(on)!=strategyrouter.StateOn{t.Fatal("control activation invalid")}
 t.Logf("same Shadow arguments: activation OFF=%s, activation ON=%s",worker.reviewShadow(shadow,input).Outcome,worker.reviewShadow(shadow,input).Outcome)
 reviewSink=strategycoordinator.NewMarketCoordinator(strategyrouter.MarketKR,f.now)
 worker.reviewShadowAttack(shadow,input)
 if reviewSink.Depth()!=1{t.Fatalf("attack failed, depth=%d",reviewSink.Depth())}
 file,err:=parser.ParseFile(token.NewFileSet(),"review_shadow_sketch_test.go",nil,0);if err!=nil{t.Fatal(err)}
 calls:=0
 for _,d:=range file.Decls{fn,ok:=d.(*ast.FuncDecl);if !ok||fn.Name.Name!="reviewShadowAttack"{continue};ast.Inspect(fn.Body,func(n ast.Node)bool{if c,ok:=n.(*ast.CallExpr);ok{calls++;if reviewCallee(c.Fun)!="worker.owns"{t.Fatalf("allowlist rejected %s",reviewCallee(c.Fun))}};return true})}
 if calls!=1{t.Fatalf("calls=%d",calls)}
 t.Logf("attack: ShadowCycle has only string outcome; allowlist calls=%d accepted; coordinator depth=%d; desired=%s effective=%s",calls,reviewSink.Depth(),worker.Desired(noActivation()),worker.Effective(noActivation()))
}
