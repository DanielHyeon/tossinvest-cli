package recheck

import (
 "go/ast"
 "go/parser"
 "go/token"
 "go/types"
 "path/filepath"
 "runtime"
 "strconv"
 "testing"
)

// Review model only: no production code or order path is changed.
// Extends the round-2 recursive type walk with v3 body Types and Uses.
func bodyUses(info *types.Info, body *ast.BlockStmt, targets ...types.Type) bool {
 found:=false
 ast.Inspect(body,func(n ast.Node)bool {
  if n==nil{return true}
  for _,target:=range targets {
   if expr,ok:=n.(ast.Expr);ok && carries(info.TypeOf(expr),target){found=true}
   if id,ok:=n.(*ast.Ident);ok {
    obj:=info.Uses[id]
    if obj!=nil && carries(obj.Type(),target){found=true}
   }
  }
  return true
 })
 return found
}

func TestV3SeparatedCarrierAndAccessorMutation(t *testing.T) {
 p,info,file:=checked(t,`package shape
 type ShadowInput struct{ proposal int }
 type FamilyShadow struct{ enabled bool }
 type strategyShadowBatch struct{ inputs []ShadowInput }
 type strategyProposalMarketAuthority struct{ entries []int }
 func collectMarket()(strategyProposalMarketAuthority,strategyShadowBatch){return strategyProposalMarketAuthority{},strategyShadowBatch{}}
 func(a strategyProposalMarketAuthority) dispatchHandoffs()[]int{return a.entries}
 func(a strategyProposalMarketAuthority) authorityForOwnerScope()int{return len(a.entries)}
 // Mutation: no carrier field added to authority; new accessor launders the type.
 func(a strategyProposalMarketAuthority) shadowAccessor()FamilyShadow{return FamilyShadow{}}
 func(a strategyProposalMarketAuthority) dispatchAttack()bool{return a.shadowAccessor().enabled}
 func(a strategyProposalMarketAuthority) dispatchMethodValue()any{f:=a.shadowAccessor;return f}
 type Alias = ShadowInput
 type Box[T any] struct{ Value T }
 func transportAlias()any{return Box[Alias]{}}
 func transportErasure()any{return any(ShadowInput{})}
 `)
 var targets []types.Type
 for _,name:=range []string{"ShadowInput","FamilyShadow","strategyShadowBatch"}{targets=append(targets,p.Scope().Lookup(name).Type())}
 authority:=p.Scope().Lookup("strategyProposalMarketAuthority").Type()
 for _,target:=range targets{if carries(authority,target){t.Fatal("authority still contains shadow")}}
 want:=map[string]bool{"collectMarket":true,"dispatchHandoffs":false,"authorityForOwnerScope":false,"shadowAccessor":true,"dispatchAttack":true,"dispatchMethodValue":true,"transportAlias":true,"transportErasure":true}
 for _,decl:=range file.Decls {
  fn,ok:=decl.(*ast.FuncDecl);if !ok{continue}
  got:=bodyUses(info,fn.Body,targets...)
  if got!=want[fn.Name.Name]{t.Fatalf("%s: got %v want %v",fn.Name.Name,got,want[fn.Name.Name])}
  t.Logf("%s body use=%v",fn.Name.Name,got)
 }
}

type fixtureImporter map[string]*types.Package
func(i fixtureImporter)Import(path string)(*types.Package,error){return i[path],nil}
func TestV3OpaqueInputRejectsForeignMintAndConversion(t *testing.T){
 fs:=token.NewFileSet()
 f,err:=parser.ParseFile(fs,"worker.go",`package strategyworker; type ShadowInput struct{ proposal int }; type Input struct{ Proposal int }`,0);if err!=nil{t.Fatal(err)}
 worker,err:=new(types.Config).Check("strategyworker",fs,[]*ast.File{f},nil);if err!=nil{t.Fatal(err)}
 for _,src:=range []string{
  `package engine; import "strategyworker"; var _ = strategyworker.ShadowInput{proposal:1}`,
  `package engine; import "strategyworker"; func admit(strategyworker.Input){}; func attack(s strategyworker.ShadowInput){admit(s)}`,
 }{
  f,err:=parser.ParseFile(fs,"attack.go",src,0);if err!=nil{t.Fatal(err)}
  _,err=(&types.Config{Importer:fixtureImporter{"strategyworker":worker}}).Check("engine",fs,[]*ast.File{f},nil)
  if err==nil{t.Fatal("opaque attack compiled")};t.Log("expected rejection:",err)
 }
}

func importsFile(t *testing.T,path,dependency string)bool{
 t.Helper();f,err:=parser.ParseFile(token.NewFileSet(),path,nil,parser.ImportsOnly);if err!=nil{t.Fatal(err)}
 for _,spec:=range f.Imports{value,err:=strconv.Unquote(spec.Path.Value);if err!=nil{t.Fatal(err)};if value==dependency{return true}}
 return false
}
func TestV3WholeClosureBanConflictsWithRequiredRouter(t *testing.T){
 if !importsFile(t,"../internal/strategyrouter/production_family_activation.go","encoding/json"){t.Fatal("router path changed")}
 for _,dependency:=range []string{"reflect","unsafe"}{
  if !importsFile(t,filepath.Join(runtime.GOROOT(),"src/encoding/json/decode.go"),dependency){t.Fatalf("json/decode no longer imports %s",dependency)}
  t.Logf("required strategyshadow -> strategyrouter -> encoding/json -> %s; whole dependency closure ban cannot pass",dependency)
 }
}
