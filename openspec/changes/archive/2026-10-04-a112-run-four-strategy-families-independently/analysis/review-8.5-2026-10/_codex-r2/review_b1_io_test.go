package strategyrouter
import("context";"errors";"testing";"os";"path/filepath")
func TestReviewB1ReadSentinels(t *testing.T){for _,mode:=range []string{"missing","directory","symlink","wrong-mode","empty","unknown-json-field","bad-json"}{t.Run(mode,func(t *testing.T){
 f:=newFamilyActivationFixture(t);c:=f.write(t,MarketKR,f.body(MarketKR));p:=filepath.Join(f.dir,ProductionFamilyActivationFileName(MarketKR))
 must:=func(e error){if e!=nil{t.Fatal(e)}}
 switch mode{
 case "missing":must(os.Remove(p))
 case "directory":must(os.Remove(p));must(os.Mkdir(p,0700))
 case "symlink":must(os.Rename(p,p+".target"));must(os.Symlink(p+".target",p))
 case "wrong-mode":must(os.Chmod(p,0600))
 case "empty":f.writeRaw(t,MarketKR,nil);c.ManifestDigest=productionRouteDigest(nil)
 case "unknown-json-field":raw:=[]byte(`{"pretend-secret-field":1}`);f.writeRaw(t,MarketKR,raw);c.ManifestDigest=productionRouteDigest(raw)
 case "bad-json":raw:=[]byte(`{"`);f.writeRaw(t,MarketKR,raw);c.ManifestDigest=productionRouteDigest(raw)
 }
 a,e:=LoadProductionFamilyActivation(context.Background(),c)
 if !errors.Is(e,ErrProductionFamilyActivationUnavailable)||errors.Is(e,ErrProductionFamilyActivationUndeclared)||a.Verified(){t.Fatalf("sentinel path escaped: %v",e)}
 t.Logf("mode=%s unavailable=true undeclared=false verified=false",mode)
 })}}
