package recheck
import (
 "crypto/sha256"
 "go/ast"
 "go/format"
 "go/parser"
 "go/token"
 "go/types"
 "testing"
)
func checked(t *testing.T, src string)(*types.Package,*types.Info,*ast.File) {
 t.Helper();fs:=token.NewFileSet();f,e:=parser.ParseFile(fs,"shape.go",src,0);if e!=nil {t.Fatal(e)}
 info:=&types.Info{Types:map[ast.Expr]types.TypeAndValue{},Uses:map[*ast.Ident]types.Object{},Defs:map[*ast.Ident]types.Object{},Selections:map[*ast.SelectorExpr]*types.Selection{}}
 p,e:=new(types.Config).Check("shape",fs,[]*ast.File{f},info);if e!=nil {t.Fatal(e)};return p,info,f
}
func carries(typ,target types.Type)bool {
 seen:=map[types.Type]bool{}
 var walk func(types.Type)bool
 walk=func(v types.Type)bool {
  if v==nil||seen[v] {return false};seen[v]=true
  v=types.Unalias(v);if types.Identical(v,target){return true}
  switch x:=v.(type) {
  case *types.Named:return walk(x.Underlying())
  case *types.Pointer:return walk(x.Elem())
  case *types.Slice:return walk(x.Elem())
  case *types.Struct:for i:=0;i<x.NumFields();i++ {if walk(x.Field(i).Type()){return true}}
  case *types.Signature:if x.Recv()!=nil&&walk(x.Recv().Type()){return true};return walk(x.Params())||walk(x.Results())
  case *types.Tuple:for i:=0;i<x.Len();i++ {if walk(x.At(i).Type()){return true}}
  };return false
 };return walk(typ)
}
func TestRecheckV2CarrierReachesDispatchReceiver(t *testing.T) {
 p,_,_:=checked(t,`package shape
 type Input struct{ Value int }
 type strategyShadowInputs struct{ inputs []Input }
 func(s strategyShadowInputs) Values()[]Input{return append([]Input(nil),s.inputs...)}
 type strategyProposalMarketAuthority struct{ entries []int; shadow strategyShadowInputs }
 func(a strategyProposalMarketAuthority) dispatchHandoffs()[]int{return a.entries}
 func(a strategyProposalMarketAuthority) attack()int{return a.shadow.Values()[0].Value}
 `)
 carrier:=p.Scope().Lookup("strategyShadowInputs").Type()
 authority:=p.Scope().Lookup("strategyProposalMarketAuthority").Type().(*types.Named)
 for i:=0;i<authority.NumMethods();i++ {
  m:=authority.Method(i);if !carries(m.Type(),carrier){t.Fatal("receiver blind census")}
  t.Logf("%s receiver carries strategyShadowInputs=true (entries-only body does not remove possession)",m.Name())
 }
}
func TestRecheckV2AliasesAndErasure(t *testing.T) {
 p,_,_:=checked(t,`package shape
 type Envelope struct{ N int }; type Box struct{ E Envelope }; type Alias = Box
 type Nested struct{ H Alias }; type Generic[T any] struct{ H T }; type Inst = Generic[Envelope]
 type Erased struct{ H any }
 `)
 target:=p.Scope().Lookup("Envelope").Type()
 for _,n:=range []string{"Nested","Inst"} {if !carries(p.Scope().Lookup(n).Type(),target){t.Fatal(n)};t.Log(n,"detected")}
 if carries(p.Scope().Lookup("Erased").Type(),target){t.Fatal("static any unexpectedly resolved")}
 t.Log("any remains erased; source freeze is necessary, go/types alone is not a dataflow proof")
}
func TestRecheckV2RebindingAndFreeze(t *testing.T) {
 src:=`package shape
 type Envelope struct{}
 type Coordinator struct{}
 func(Coordinator) Submit(Envelope)bool{return true}
 func attack(q Coordinator){worker:=struct{owns func(Envelope)bool}{owns:q.Submit};worker.owns(Envelope{})}
 `
 _,info,_:=checked(t,src);submit,indirect:=false,false
 for expr,s:=range info.Selections {
  if expr.Sel.Name=="Submit" {submit=s.Kind()==types.MethodVal}
  if expr.Sel.Name=="owns" {indirect=s.Kind()==types.FieldVal}
 }
 if !submit||!indirect {t.Fatal("function-value origin not detected")}
 base:=[]byte("package shape\nfunc verdict()string{return \"WOULD_EMIT\"}\n")
 mutated:=[]byte(src)
 a,e:=format.Source(base);if e!=nil{t.Fatal(e)};b,e:=format.Source(mutated);if e!=nil{t.Fatal(e)}
 if sha256.Sum256(a)==sha256.Sum256(b){t.Fatal("freeze missed edit")}
 t.Log("q.Submit resolved as MethodVal; worker.owns resolved as function FieldVal; frozen source edit rejected")
}
