package strategyprojectionrpc
import("context";"encoding/json";"io";"net/http";"strings";"testing";"time";"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection")
type reviewMemoryTransport struct{body string}
func (m reviewMemoryTransport) RoundTrip(*http.Request)(*http.Response,error){return &http.Response{StatusCode:200,Body:io.NopCloser(strings.NewReader(m.body)),Header:make(http.Header)},nil}
func TestReviewShadowRPCWithoutNetwork(t *testing.T){
 s:=strategyprojection.DormantSnapshot(time.Date(2026,10,4,0,0,0,0,time.UTC));raw,_:=json.Marshal(s)
 var doc map[string]any;if err:=json.Unmarshal(raw,&doc);err!=nil{t.Fatal(err)}
 lanes:=doc["lanes"].([]any);lanes[0].(map[string]any)["shadowOutcome"]="WOULD_EMIT"
 read:=func()error{body,_:=json.Marshal(doc);c:=&Client{baseURL:"http://memory.invalid",token:strings.Repeat("x",40),http:&http.Client{Transport:reviewMemoryTransport{string(body)}}};_,err:=c.Read(context.Background());return err}
 if err:=read();err!=nil{t.Fatal(err)};t.Log("unknown shadowOutcome: accepted by actual Client.Read (in-memory transport)")
 lanes[0].(map[string]any)["runtime"]="SHADOW"
 if err:=read();err==nil{t.Fatal("SHADOW unexpectedly accepted")}else{t.Logf("SHADOW enum: rejected: %v",err)}
}
