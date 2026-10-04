package strategyrouter
import("testing";"context";"encoding/json";"errors";"fmt";"os";"math/rand";"time";"strings")
func TestReviewB1Acceptance(t *testing.T){
 out,e:=os.Create(os.Getenv("REVIEW_OUT"));if e!=nil{t.Fatal(e)};defer out.Close()
 f:=newFamilyActivationFixture(t)
 type mutation func(*productionFamilyActivationBody,*FamilyActivationConfig)
 mods:=[]mutation{
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.SchemaVersion="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Domain="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Generation=0},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Market="XX"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.RouteManifestDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.CalibrationDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.CalendarVersion="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.BuildDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.RiskPolicyDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ProtectionReadyMinGeneration=0},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Actor=""},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ApprovedAt="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.IssuedAt="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ExpiresAt="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.IssuedAt=f.now.Add(time.Minute).Format(time.RFC3339Nano)},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ApprovedAt=f.now.Format(time.RFC3339Nano)},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ExpiresAt=f.now.Add(-time.Minute).Format(time.RFC3339Nano)},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.ExpiresAt=f.now.Add(48*time.Hour).Format(time.RFC3339Nano)},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].LaneID="unknown"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Family="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Horizon="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].LaneVersion="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Desired="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Effective="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Desired=StateOff},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[1]=b.Descriptors[0]},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors=b.Descriptors[:3]},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.ConfigDir="relative"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.Market="XX"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.ObservedAt=time.Time{}},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.CalibrationDigest=""},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.RouteManifestDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.RiskPolicyDigest="bad"},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.CalendarVersion=""},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){c.BuildDigest=""},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Revoked=true},
 func(b *productionFamilyActivationBody,c *FamilyActivationConfig){b.Descriptors[0].Desired=StateOff;b.Descriptors[0].Effective=StateOff},
 }
 sets:=[][]int{nil};for j:=range mods{sets=append(sets,[]int{j});for k:=j+1;k<len(mods);k++{sets=append(sets,[]int{j,k})}}
 r:=rand.New(rand.NewSource(884));for n:=0;n<2000;n++{var set []int;for j:=range mods {if r.Intn(9)==0 {set=append(set,j)}};sets=append(sets,set)}
 accepted:=0
 for n,set:=range sets {
 b:=f.body(MarketKR);raw,_:=json.Marshal(b);c:=f.config(MarketKR,b,raw)
 for _,j:=range set{mods[j](&b,&c)}
 raw,_=json.Marshal(b);f.writeRaw(t,MarketKR,raw);c.ManifestDigest=productionRouteDigest(raw)
 for variant:=0;variant<6;variant++{
 cc:=c;ctx:=context.Background()
 switch variant{case 1:cc.ManifestDigest=" \t";case 2:cc.ManifestDigest="sha256:"+strings.Repeat("0",64);case 3:ctx=nil;case 4:var cancel context.CancelFunc;ctx,cancel=context.WithCancel(ctx);cancel();case 5:f.writeRaw(t,MarketKR,append(raw,' '));cc.ManifestDigest=productionRouteDigest(append(raw,' '))}
 a,err:=LoadProductionFamilyActivation(ctx,cc)
 class:="accepted";if err!=nil{class="other";for _,s:=range []struct{e error;n string}{{ErrProductionFamilyActivationUndeclared,"undeclared"},{ErrProductionFamilyActivationUnavailable,"unavailable"},{ErrProductionFamilyActivationExpired,"expired"},{ErrProductionFamilyActivationRevoked,"revoked"},{context.Canceled,"canceled"}}{if errors.Is(err,s.e){class=s.n;break}}}
 if a.Verified(){accepted++}
 fmt.Fprintf(out,"%04d/%d %s verified=%v generation=%d\n",n,variant,class,a.Verified(),a.Generation())
 }
 }
 t.Logf("B1 combinations=%d observations=%d accepted=%d",len(sets),len(sets)*6,accepted)
 if accepted==0{t.Fatal("vacuous acceptance corpus")}
}
func TestReviewB1DiagnosticPayload(t *testing.T){
 f:=newFamilyActivationFixture(t);b:=f.body(MarketKR);b.Descriptors[0].LaneID="synthetic-account-123\nFORGED_RECORD";c:=f.write(t,MarketKR,b)
 _,err:=LoadProductionFamilyActivation(context.Background(),c)
 t.Logf("lane_error=%q undeclared=%v",err,errors.Is(err,ErrProductionFamilyActivationUndeclared))
}
