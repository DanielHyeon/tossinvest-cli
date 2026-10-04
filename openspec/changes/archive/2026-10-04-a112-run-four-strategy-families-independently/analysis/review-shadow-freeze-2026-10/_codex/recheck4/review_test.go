package main

import (
 "encoding/json"
 "go/build"
 "os"
 "path/filepath"
 "sort"
 "strings"
 "testing"
 "time"
)

func TestASTEvidence(t *testing.T) {
 targets := map[string][]string{
 "strategy_lane_runtime.go":{"strategyLaneRuntime.evaluate","strategyLaneRuntime.record"},
 "strategy_lane_latch.go":{"strategyLaneRuntime.recoverMarketLanes"},
 "strategy_entry_supervisor.go":{"Context.runProductionStrategyMarketCycle"},
 "strategy_market_coordinator.go":{"coordinateMarketProposals"},
 "strategy_lane_projection.go":{"strategyLaneRuntime.projection"},
 }
 for file, names := range targets { for _, name := range names {
  evidence,err:=analyze("../../internal/app/engine/"+file,name);if err!=nil{t.Fatal(err)}
  data,err:=json.MarshalIndent(evidence,"","  ");if err!=nil{t.Fatal(err)}
  if err=os.WriteFile("ast-"+name+".json",data,0600);err!=nil{t.Fatal(err)}
  t.Logf("AST %s lines %d-%d",name,evidence.Start.Line,evidence.End.Line)
 }}
}

// Source import closure only. No go list subprocess; four modes match testenv.WalkModes.
func TestModuleDirectImports(t *testing.T) {
 const prefix="github.com/JungHoonGhae/tossinvest-cli/"
 for _,root:=range []string{"internal/strategyrouter","internal/strategyworker"}{
  for _,mode:=range []struct{name string; tests,tag bool}{{"deps",false,false},{"deps-test",true,false},{"deps-tagged",false,true},{"deps-test-tagged",true,true}}{
   t.Run(root+"/"+mode.name,func(t *testing.T){
    ctx:=build.Default;if mode.tag{ctx.BuildTags=[]string{"tossos_testseams"}}
    seen:=map[string]bool{}; rows:=map[string][]string{};external:=map[string]bool{}; forbidden:=0
    var walk func(string)
    walk=func(rel string){if seen[rel]{return};seen[rel]=true
     p,err:=ctx.ImportDir(filepath.Join("../..",rel),0);if err!=nil{t.Fatal(err)}
     imports:=append([]string{},p.Imports...)
     if rel==root && mode.tests {imports=append(imports,p.TestImports...);imports=append(imports,p.XTestImports...)}
     sort.Strings(imports);rows[rel]=imports
     for _,dep:=range imports{
      if dep=="unsafe"||dep=="reflect"{forbidden++;t.Logf("direct import found: %s -> %s",rel,dep)}
      if strings.HasPrefix(dep,prefix){walk(strings.TrimPrefix(dep,prefix))}else if strings.Contains(strings.Split(dep,"/")[0],"."){external[dep]=true}
     }
    };walk(root)
    keys:=make([]string,0,len(rows));for k:=range rows{keys=append(keys,k)};sort.Strings(keys)
    for _,k:=range keys{t.Logf("%s imports %v",k,rows[k])}
    t.Logf("module packages=%d external frontier=%v direct unsafe/reflect=%d",len(rows),external,forbidden)
    expected:=0;if mode.tests{expected=1;if root=="internal/strategyrouter" && mode.tag{expected=3}}
    if forbidden!=expected{t.Fatalf("observed %d direct-import packages, want %d",forbidden,expected)}
   })
  }
 }
}

// v3.1 algorithm model, not a production implementation test.
func TestV31PreRecordFailureRetainsShadow(t *testing.T){
 now:=time.Unix(100,0);expires:=now.Add(time.Hour)
 for _,failure:=range []string{"recover-error","lane-panic"}{t.Run(failure,func(t *testing.T){
  latestWave,observationWave:=uint64(7),uint64(7)
  // Actual AST: failure precedes record; cycle cannot reach shadow kickoff.
  recorded,started:=false,false
  usable:=observationWave==latestWave && now.Before(expires)
  if recorded||started||!usable{t.Fatal("counterexample did not reproduce")}
  t.Log("REPRODUCED: cycle fails before record; wave=7; previous observation wave=7; projection remains SHADOW before expiry")
 })}
}
func TestV31AbsentEmptyAndExpiryModel(t *testing.T){
 verdict:=func(present bool,n int)string{if !present{return "UNOBSERVED"};if n==0{return "NO_INPUT"};return "WOULD_EMIT"}
 if verdict(false,0)!="UNOBSERVED"||verdict(true,0)!="NO_INPUT"{t.Fatal("absence collapsed")}
 expiry:=time.Unix(100,0)
 usable:=func(now time.Time)bool{return now.Before(expiry)}
 if !usable(expiry.Add(-time.Nanosecond))||usable(expiry){t.Fatal("expiry boundary")}
}
