package main
import("encoding/json";"os";"testing")
func TestReviewShadowEvidence(t *testing.T){
 targets:=map[string][]string{
 "internal/strategyworker/worker.go":{"FamilyWorker.Run","FamilyWorker.owns"},
 "internal/app/engine/strategy_proposal_authority.go":{"strategyProposalAuthorityLoader.collectMarket"},
 "internal/app/engine/strategy_lane_runtime.go":{"strategyLaneRuntime.evaluate","strategyLaneInputs"},
 "internal/app/engine/strategy_family_activation.go":{"strategyFamilyGate.admit","strategyProposalAuthorityLoader.familyGateFor"},
 "internal/strategyrouter/production_family_activation.go":{"LoadProductionFamilyActivation"},
 "internal/strategyprojection/lanes.go":{"validateLane"},
 }
 for path,names:=range targets {for _,name:=range names { evidence,err:=analyze("../../"+path,name);if err!=nil{t.Fatal(err)};data,_:=json.MarshalIndent(evidence,"","  ");if err=os.WriteFile("../../review-experiments/ast-"+name+".json",data,0600);err!=nil{t.Fatal(err)};t.Log(name)}}
}
