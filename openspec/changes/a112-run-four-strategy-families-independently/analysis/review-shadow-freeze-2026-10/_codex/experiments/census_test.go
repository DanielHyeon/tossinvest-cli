package reviewexperiments
import("go/ast";"go/parser";"go/token";"go/types";"testing")
type seamPackage struct {handoff,delivered types.Type}
func (s seamPackage) carries(typ types.Type) bool {
	return s.carriesSeen(typ, map[types.Type]bool{})
}

func (s seamPackage) carriesSeen(typ types.Type, seen map[types.Type]bool) bool {
	if typ == nil || seen[typ] {
		return false
	}
	seen[typ] = true
	if types.Identical(typ, s.handoff) || types.Identical(typ, s.delivered) {
		return true
	}
	switch value := typ.(type) {
	case *types.Named:
		return s.carriesSeen(value.Underlying(), seen)
	case *types.Pointer:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Slice:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Array:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Map:
		return s.carriesSeen(value.Key(), seen) || s.carriesSeen(value.Elem(), seen)
	case *types.Chan:
		return s.carriesSeen(value.Elem(), seen)
	case *types.Struct:
		for i := 0; i < value.NumFields(); i++ {
			if s.carriesSeen(value.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Signature:
		return s.carriesSeen(value.Params(), seen) || s.carriesSeen(value.Results(), seen)
	case *types.Tuple:
		for i := 0; i < value.Len(); i++ {
			if s.carriesSeen(value.At(i).Type(), seen) {
				return true
			}
		}
	case *types.Interface:
		for i := 0; i < value.NumMethods(); i++ {
			if s.carriesSeen(value.Method(i).Type(), seen) {
				return true
			}
		}
	case *types.TypeParam:
		return s.carriesSeen(value.Constraint(), seen)
	}
	return false
}


func TestReviewCensusAliasAndErasure(t *testing.T){
 source:=`package specimen
type Envelope struct{ N int }
type Delivered struct{ N int }
type Box struct{ E Envelope }
type Hidden = Box
type ShadowAlias struct{ H Hidden }
type ShadowAny struct{ H any }
type ShadowDirect struct{ E Envelope }
type ShadowGeneric[T any] struct{ H T }
type ShadowInst = ShadowGeneric[Envelope]
`
 fs:=token.NewFileSet();f,e:=parser.ParseFile(fs,"specimen.go",source,0);if e!=nil{t.Fatal(e)}
 pkg,e:=new(types.Config).Check("specimen",fs,[]*ast.File{f},nil);if e!=nil{t.Fatal(e)}
 s:=seamPackage{pkg.Scope().Lookup("Envelope").Type(),pkg.Scope().Lookup("Delivered").Type()}
 for _,n:=range []string{"ShadowDirect","ShadowAlias","ShadowAny","ShadowInst"}{t.Logf("%s carries=%v",n,s.carries(pkg.Scope().Lookup(n).Type()))}
 if !s.carries(pkg.Scope().Lookup("ShadowDirect").Type()){t.Fatal("positive control missed")}
 if s.carries(pkg.Scope().Lookup("ShadowAny").Type()){t.Fatal("expected erased static type")}
}
