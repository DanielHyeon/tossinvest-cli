package breakoutlane
import("testing";"fmt";"encoding/json";"os";"math/rand")
func TestReviewAEquivalence(t *testing.T){
 out,err:=os.Create(os.Getenv("REVIEW_OUT"));if err!=nil{t.Fatal(err)};defer out.Close()
 flags,err:=os.Create(os.Getenv("REVIEW_OUT")+".flags");if err!=nil{t.Fatal(err)};defer flags.Close()
 emit:=func(name string,d Decision){
 if !validDecision(d){t.Fatalf("invalid seal %s",name)}
 p:=d.Provenance();fmt.Fprintf(flags,"%s %v\n",name,p.RVOLAt1200000)
 p.RVOLAt1200000=false
 data,_:=json.Marshal(struct{Name,Phase,Refusal,Proposal,Seal,Digest,Diagnostic string;Candidate,Final uint64;P Provenance}{name,d.Phase(),string(d.Refusal()),d.ProposalID(),d.seal,d.snapshotDigest,string(d.diagnostic),d.CandidateQuantity(),d.FinalQuantity(),p})
 fmt.Fprintln(out,string(data))
 }
 r:=rand.New(rand.NewSource(11285))
 for n:=0;n<2400;n++{
 i:=fixtureInput(t);i.Bars=append([]ClosedBar(nil),i.Bars[:15]...)
 // Opening range boundary plus multiple lower-only, first touch, admission and retest/reclaim bars.
 if n%5==0 {i.Bars[14]=fixtureBar(t,15,100,90,100,1_300_000,100_000)}
 count:=1+r.Intn(20)
 for j:=0;j<count;j++{
 close:=uint64(98+r.Intn(5));rv:=[]uint64{1_000_000,1_199_999,1_200_000,1_300_000,1_499_999,1_500_000,2_000_000,2_500_000}[r.Intn(8)]
 wick:=[]uint64{0,350_000,350_001}[r.Intn(3)]
 i.Bars=append(i.Bars,fixtureBar(t,uint64(16+j),111,90,close,rv,wick))
 }
 s:=snapshot(t,i);d:=Evaluate(s,nil);name:=fmt.Sprintf("%04d",n);emit(name+"-fresh",d)
 emit(name+"-reuse",Evaluate(s,&d))
 // Trusted old decisions with old diagnostic flag remain valid and must not alter decision fields.
 old:=d;old.provenance.RVOLAt1200000=false
 emit(name+"-old-prior",Evaluate(s,&old))
 b:=i.Bars[15].value;b.Revision++;b.RVOLPPM=1_300_000;i.Bars[15]=ClosedBar{value:b}
 emit(name+"-correction",Evaluate(snapshot(t,i),&d))
 }

 cases:=[]struct{name string; values [][2]uint64; rangeRVOL bool}{
 {"multiple-lower-only",[][2]uint64{{101,1200000},{101,1300000},{101,1499999}},false},
 {"lower-before-entry",[][2]uint64{{101,1200000},{101,1300000},{101,1500000},{99,1000000},{101,1000000}},false},
 {"lower-after-entry",[][2]uint64{{101,1500000},{99,1200000},{101,1300000}},false},
 {"first-touch-before-lower",[][2]uint64{{100,1300000},{101,1200000}},false},
 {"first-touch-same-bar",[][2]uint64{{100,1300000}},false},
 {"opening-range-boundary",[][2]uint64{{100,1000000}},true},
 {"lower-then-invalidated",[][2]uint64{{101,1200000},{101,1500000},{89,1000000}},false},
 {"lower-then-timeout",[][2]uint64{{101,1200000},{101,1500000},{103,1000000},{103,1000000},{103,1000000},{103,1000000},{103,1000000},{103,1000000},{103,1000000},{103,1000000},{103,1000000}},false},
 }
 for _,tc:=range cases{
 i:=fixtureInput(t);i.Bars=append([]ClosedBar(nil),i.Bars[:15]...)
 if tc.rangeRVOL{i.Bars[14]=fixtureBar(t,15,100,90,100,1300000,100000)}
 for j,v:=range tc.values{i.Bars=append(i.Bars,fixtureBar(t,uint64(j+16),111,80,v[0],v[1],100000))}
 s:=snapshot(t,i);d:=Evaluate(s,nil);emit(tc.name+"-fresh",d);emit(tc.name+"-reuse",Evaluate(s,&d))
 old:=d;old.provenance.RVOLAt1200000=false;emit(tc.name+"-old-prior",Evaluate(s,&old))
 b:=i.Bars[15].value;b.Revision++;b.RVOLPPM=1300000;i.Bars[15]=ClosedBar{value:b};emit(tc.name+"-correction",Evaluate(snapshot(t,i),&d))
 }
 t.Log("A 2400 random + 8 named snapshots x 4 = 9632 decisions")
}
