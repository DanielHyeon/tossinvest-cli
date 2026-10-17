package main

import (
 "encoding/json"
 "errors"
 "os"
 "reflect"
 "sync"
 "testing"
 "time"
)

// v3.2 기반 최소 모델: 기존 회귀 보존, v3.3 §5.1 두 주장만 별도 검증. 생산 통합 시험이 아님.
const poll = 5*time.Second
const cycleLimit = 30*time.Second
const stepDeadline = 2*time.Second
const maxAge = 2*(poll+cycleLimit+stepDeadline)
type stamp struct { wave, epoch uint64 }
type observation struct { stamp; at, expiry time.Time }
type model struct { mu sync.RWMutex; stamp; obs *observation }
func (m *model) snapshot() stamp { m.mu.RLock(); defer m.mu.RUnlock(); return m.stamp }
func (m *model) invalidate() { m.mu.Lock(); defer m.mu.Unlock(); m.epoch++; m.obs=nil }
func (m *model) record() { m.mu.Lock(); defer m.mu.Unlock(); m.wave++ }
func (m *model) publish(s stamp, now, expiry time.Time) bool {
 m.mu.Lock(); defer m.mu.Unlock()
 if m.epoch!=s.epoch || m.wave!=s.wave { return false }
 m.obs=&observation{s,now,expiry};return true
}
func (m *model) usable(now time.Time) bool {
 m.mu.RLock(); defer m.mu.RUnlock()
 return m.obs!=nil && m.obs.wave==m.wave && now.Before(m.obs.expiry) && now.Sub(m.obs.at)<maxAge
}
func (m *model) cycle(run func() error) error {
 returnedNil:=false
 defer func(){if !returnedNil {m.invalidate()}}()
 err:=run();returnedNil=err==nil;return err
}
func seeded(t *testing.T) (*model,time.Time) {
 t.Helper();now:=time.Unix(100,0);m:=&model{stamp:stamp{7,0}}
 if !m.publish(m.snapshot(),now,now.Add(time.Hour)){t.Fatal("seed")};return m,now
}
func TestASTEvidence(t *testing.T) {
 targets:=map[string][]string{
 "strategy_lane_runtime.go":{"strategyLaneRuntime.evaluate","strategyLaneRuntime.record","strategyLaneRuntime.runLane"},
 "strategy_lane_latch.go":{"strategyLaneRuntime.recoverMarketLanes"},
 "strategy_lane_projection.go":{"strategyLaneRuntime.projection","strategyLaneProjection"},
 "strategy_entry_supervisor.go":{"Context.runProductionStrategyMarketCycle","Context.NewRefreshingPairedStrategyEntrySupervisor","Context.productionStrategyWorker","invokeStrategyCycle","invokeBoundedStrategyCycle","StrategyEntrySupervisor.runMarket"},
 }
 for file,names:=range targets { for _,name:=range names {
  e,err:=analyze("../../internal/app/engine/"+file,name);if err!=nil{t.Fatal(err)}
  b,err:=json.MarshalIndent(e,"","  ");if err!=nil{t.Fatal(err)}
  if err=os.WriteFile("ast-"+name+".json",b,0600);err!=nil{t.Fatal(err)}
  t.Logf("%s: %d-%d",name,e.Start.Line,e.End.Line)
 }}
}
func TestV32PreRecordFailuresAndLatePublish(t *testing.T) {
 for _,kind:=range []string{"recover-error","lane-panic"}{t.Run(kind,func(t *testing.T){
  m,now:=seeded(t);old:=m.snapshot();sentinel:=errors.New(kind)
  run:=func()error{if kind=="lane-panic"{panic(sentinel)};return sentinel}
  var got error;var recovered any
  func(){defer func(){recovered=recover()}();got=m.cycle(run)}()
  if kind=="lane-panic" {if recovered!=sentinel{t.Fatal("panic changed")}} else if got!=sentinel{t.Fatal("error changed")}
  if m.snapshot().wave!=7 || m.snapshot().epoch!=1 || m.usable(now){t.Fatal("pre-record failure retained shadow")}
  if m.publish(old,now,now.Add(time.Hour)) || m.usable(now){t.Fatal("late publication revived shadow")}
  t.Log("wave=7 unchanged; epoch=1; immediate UNOBSERVED; late publication rejected; original error/panic identity preserved")
 })}
}
func TestV32SuccessPreservesObservation(t *testing.T){
 m,now:=seeded(t);calls:=0
 if err:=m.cycle(func()error{calls++;return nil});err!=nil||calls!=1||m.snapshot().epoch!=0||!m.usable(now){t.Fatal("success invalidated")}
}
func TestV32AgeAndExpiryBoundaries(t *testing.T){
 if maxAge!=74*time.Second{t.Fatal(maxAge)}
 m,now:=seeded(t)
 if !m.usable(now.Add(maxAge-time.Nanosecond))||m.usable(now.Add(maxAge)){t.Fatal("age boundary")}
 expiry:=now.Add(time.Second)
 m.publish(m.snapshot(),now,expiry)
 if !m.usable(expiry.Add(-time.Nanosecond))||m.usable(expiry){t.Fatal("expiry boundary")}
 t.Log("74s equality rejects; 74s-1ns accepts; expiry equality rejects")
}
func TestV32CASOrdersAndMarketIsolation(t *testing.T){
 for _,invalidateFirst:=range []bool{false,true}{
  a,now:=seeded(t);b,_:=seeded(t);s:=a.snapshot()
  if invalidateFirst{a.invalidate();if a.publish(s,now,now.Add(time.Hour)){t.Fatal("stale epoch accepted")}}else{a.publish(s,now,now.Add(time.Hour));a.invalidate()}
  if a.usable(now)||!b.usable(now){t.Fatal("isolation/order")}
 }
 m,now:=seeded(t);old:=m.snapshot();m.record()
 if m.publish(old,now,now.Add(time.Hour)){t.Fatal("stale wave accepted")}
}
func TestV32ConcurrentPublishInvalidate(t *testing.T){
 for i:=0;i<100;i++{m,now:=seeded(t);s:=m.snapshot();var wg sync.WaitGroup;wg.Add(2)
  go func(){defer wg.Done();m.publish(s,now,now.Add(time.Hour))}()
  go func(){defer wg.Done();m.invalidate()}()
  wg.Wait();if m.usable(now){t.Fatal("resurrection")}
 }
}
func TestV33HealthyCycleIntendedGap(t *testing.T){
 m,now:=seeded(t)
 // Healthy next cycle: 5s poll + just under 30s cycle, then just under 2s SHADOW.
 recordAt:=now.Add(poll+cycleLimit-time.Nanosecond);publicationAt:=recordAt.Add(stepDeadline-time.Nanosecond)
 if !m.usable(recordAt){t.Fatal("age already rejected healthy old observation")}
 m.record()
 if m.usable(recordAt.Add(time.Nanosecond)){t.Fatal("expected new-wave gap not reproduced")}
 if !m.publish(m.snapshot(),publicationAt,now.Add(time.Hour))||!m.usable(publicationAt){t.Fatal("healthy publication")}
 t.Log("v3.3 INTENDED GAP (former v3.2 P1-E): t=35s-1ns record wave 8 => UNOBSERVED; t=37s-2ns publish wave 8 => SHADOW, despite age < 74s and no failure")
}

func TestV32CycleTraceMatchesBaseline(t *testing.T){
 for _,mode:=range []string{"success","pre-record-error","pre-record-panic","post-record-error"}{
  sentinel:=errors.New(mode)
  execute:=func(wrapped bool)([]string,error,any){
   m,_:=seeded(t);trace:=[]string{};var got error;var recovered any
   run:=func()error{
    trace=append(trace,"recover-lanes")
    if mode=="pre-record-error"{return sentinel}
    trace=append(trace,"lane-run")
    if mode=="pre-record-panic"{panic(sentinel)}
    m.record();trace=append(trace,"record","persist-latches")
    if mode=="post-record-error"{return sentinel}
    trace=append(trace,"dispatch");return nil
   }
   func(){defer func(){recovered=recover()}();if wrapped{got=m.cycle(run)}else{got=run()}}()
   return trace,got,recovered
  }
  a,ae,ap:=execute(false);b,be,bp:=execute(true)
  if !reflect.DeepEqual(a,b)||ae!=be||ap!=bp{t.Fatalf("%s trace/error/panic changed",mode)}
  t.Logf("%s: %v, unchanged error/panic",mode,b)
 }
}

// 파도와 만료를 고정하여 나이 조건을 독립 검증함.
func TestV33SameWaveAgeOnlyAcceptance(t *testing.T) {
 m,now:=seeded(t)
 for _,age:=range []time.Duration{0,poll+cycleLimit-time.Nanosecond,poll+cycleLimit+stepDeadline} {
  if !m.usable(now.Add(age)){t.Fatalf("same wave rejected at age %v",age)}
 }
 if m.snapshot().wave!=7 {t.Fatal("wave changed")}
 t.Log("same wave: age 0, 35s-1ns, 37s accepted; age bound alone does not reject")
}
