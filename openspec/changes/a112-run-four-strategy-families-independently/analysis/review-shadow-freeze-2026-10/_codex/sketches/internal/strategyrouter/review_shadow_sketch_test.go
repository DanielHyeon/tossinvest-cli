package strategyrouter
import("go/parser";"go/token";"testing")
// Review-only same-package analogue of a shadow helper; does not call a loader or encoder.
type reviewActivationTwin FamilyActivation
func reviewShadowSideDoor() any {
 descriptors:=productionRouteDescriptors(MarketKR)
 state:=map[familyLaneKey]productionFamilyActivationDescriptor{}
 for id,d:=range descriptors { state[familyLaneKey{family:d.Family,laneID:id,laneVersion:d.LaneVersion}]=productionFamilyActivationDescriptor{Family:d.Family,LaneID:id,LaneVersion:d.LaneVersion,Desired:StateOn,Effective:StateOn} }
 return FamilyActivation(reviewActivationTwin{market:MarketKR,generation:1,state:state})
}
func TestReviewShadowSamePackageMint(t *testing.T){
 a:=reviewShadowSideDoor().(FamilyActivation)
 if !a.Verified(){t.Fatal("not verified")}
 on:=0;for id,d:=range productionRouteDescriptors(MarketKR){if a.Effective(MarketKR,d.Family,id,d.LaneVersion)==StateOn{on++}}
 if on!=4{t.Fatalf("on=%d",on)}
 parsed,err:=parser.ParseFile(token.NewFileSet(),"review_shadow_sketch_test.go",nil,0);if err!=nil{t.Fatal(err)}
 counts:=countFamilyActivationAuthoringReferences(parsed,"internal/strategyrouter/review_shadow_sketch.go")
 for name,n:=range counts{if n!=0{t.Fatalf("encoder guard caught %s=%d",name,n)}}
 t.Logf("same-package twin conversion -> any -> FamilyActivation: Verified=%v ON=%d; encoder references=%v; file writes=0",a.Verified(),on,counts)
}
