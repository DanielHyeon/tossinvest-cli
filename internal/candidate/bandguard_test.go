package candidate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// bandguard_test.go carries the selector TestNoFunctionThatProducesAVerdictCanSeeAShadowBand
// is built from.
//
// It is a separate file for the reason the change's tasks give for every new test:
// a helper added to band_test.go shares a diff hunk with whatever function it lands
// beside, and the Function Logic Map gate then asks for evidence about a function
// nobody touched. The split is a placement rule, not a way around the gate — the two
// tests this change actually rewrites stay in band_test.go and carry their own maps.
//
// # What these helpers exist to fix
//
// The check they serve was believed to make a band unreachable from a verdict, and
// on 2026-07-28 a reviewer put one line into assessOne —
//
//	if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() }
//
// — and `go test ./...` stayed green across 51 packages. A number nobody had
// approved decided a veto and nothing failed. Two blind spots produced that:
// the selector read only *ast.Ident results, so `*VetoState`, `[]Chase` and
// `map[K]Chase` were invisible; and the verdict set held only VetoState and Chase,
// so assessOne — whose result is a Verdict, and the only place a Chase and a band
// are in scope together — was never examined at all.

// verdictProducers is the floor on how many verdict-producing functions the check
// must still be reaching.
//
// It is a floor rather than an exact count because adding a verdict-producing
// function is ordinary and losing the check's reach is not. `checked == 0` was the
// only guard before, and a selector that had stopped matching all but one function
// would have passed it: the count would go from a number nobody had written down to
// a smaller number nobody had written down, and the check would keep reporting
// success over the functions it no longer read. That is not hypothetical — it is
// the exact shape of the defect above, where the number was 10 and should have
// been 12.
//
// The value is the count observed on 2026-07-29 with the widened selector
// (`go test -run TestNoFunctionThatProducesAVerdictCanSeeAShadowBand -v` logs it).
// If a verdict-producing function is deliberately removed, this number moves down in
// the same commit and the commit message says which function went.
const verdictProducers = 12

// verdictResultNames returns the type names a function's results are built from.
//
// A result is not always a bare identifier, and the version of this check that read
// only *ast.Ident was blind to every other spelling: `*VetoState`, `[]Chase` and
// `map[K]Chase` all name a verdict type and none of them is an ident. `Assess`
// returns `([]Verdict, error)` and was invisible for that reason alone.
func verdictResultNames(expr ast.Expr) []string {
	switch t := expr.(type) {
	case *ast.Ident:
		return []string{t.Name}
	case *ast.StarExpr:
		return verdictResultNames(t.X)
	case *ast.ArrayType:
		return verdictResultNames(t.Elt)
	case *ast.MapType:
		return append(verdictResultNames(t.Key), verdictResultNames(t.Value)...)
	default:
		return nil
	}
}

// assignedTo collects the expressions a function writes to, so that assembling a
// band into a verdict can be told apart from reading one.
//
// The distinction is forced by the shape of the code rather than chosen: a Verdict
// has two band fields and something has to fill them, so "a verdict-producing
// function may not name a band field at all" is a rule no assembly can satisfy.
// What such a function must never do is *read* one, because reading is the only way
// a band can reach a decision — and the review's one line read one.
func assignedTo(fn *ast.FuncDecl) map[ast.Expr]bool {
	out := map[ast.Expr]bool{}
	ast.Inspect(fn, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok {
			for _, e := range as.Lhs {
				out[e] = true
			}
		}
		return true
	})
	return out
}

// --- the hop (§5.1) --------------------------------------------------------------
//
// TestNoFunctionThatProducesAVerdictCanSeeAShadowBand above is a rule about ONE
// function: the one that produces the verdict. On 2026-07-29 the independent review
// put a second function between the record and the decision —
//
//	func chaseWorthy(v Verdict) bool { return v.ExtendedBand.Crossed("6") }
//	// inside assessOne
//	if chaseWorthy(v) { v.Chase.Extended = RaisedVeto() }
//
// — and 51 packages, `go vet` and the check above all passed. The verdict-producing
// function names nothing forbidden, and the helper returns no verdict type so it was
// never examined. The requirement is not "must not read directly"; it is that a band
// must be UNREACHABLE from a decision, and inserting a hop does not remove a path.
//
// What follows closes the hop by forbidding the conversion itself.

// shadowRecordTypes are the names a shadow record is declared with, before this
// package's own type declarations are folded in.
//
// They are the whole family and not only ShadowBand: a []BandCrossing is the record
// with its wrapper removed, and a BandTally is the record summed. Handing any of them
// back is handing back a record.
//
// This is a SEED and not the answer. recordTypeClosure below grows it, because a
// name is not what makes something a record — see there.
var shadowRecordTypes = map[string]bool{
	"ShadowBand": true, "BandCrossing": true, "BandTally": true, "BandQuantile": true,
}

// recordTypeClosure is every type name in this package that IS a shadow record.
//
// # Why the seed above is not enough (S2)
//
// The second independent review spelled a record two ways this file could not read,
// and neither needed a void function or a hop:
//
//	type b4band = ShadowBand              // an alias: the SAME type, another name
//	type b5box struct{ ShadowBand }       // an embedding: promoted fields, another name
//
// `func take(v Verdict) b4band { return v.ExtendedBand }` returns a ShadowBand — the
// compiler knows it and this file did not, because the check asked whether the
// RESULT TYPE WAS SPELLED "ShadowBand". `take(v).Measured` then read a record out of
// an expression nothing here recognised as one. Judging a type by the letters in its
// name is the same class of mistake as judging a function by the letters in its body,
// and it is the class this whole file exists to stop being made about bands.
//
// So the set is closed instead of listed. A declared type is a record when the type
// it is declared FROM is one — which covers aliases and defined types together,
// through pointers, slices, maps and channels — or when it embeds one, because an
// embedded record's fields are readable straight off the outer value. Running that to
// a fixed point catches `type a = ShadowBand; type b = a` as well, and needs no name
// written down anywhere.
//
// What deliberately does NOT make a type a record is holding one in a NAMED field.
// Verdict holds two and must not become a record: it is the thing a record is
// assembled into, and the requirement protects that assembly. A named field is
// reached through a selector this walk already recognises (bandShape.fields); an
// embedded one is reached through a selector that names the outer type, which is the
// hole.
func recordTypeClosure(pkgs map[string]*ast.Package) map[string]bool {
	out := map[string]bool{}
	for name := range shadowRecordTypes {
		out[name] = true
	}
	specs := typeSpecs(pkgs)
	// A fixed point, because nothing says `type b = a` is declared after `type a =
	// ShadowBand`. The cap is only there so a bug cannot hang a test.
	for round := 0; round < 16; round++ {
		before := len(out)
		for _, s := range specs {
			if out[s.name] || !buildsOnARecord(s.typ, out) {
				continue
			}
			out[s.name] = true
		}
		if len(out) == before {
			break
		}
	}
	return out
}

// declaredType is one `type X …` in this package.
type declaredType struct {
	name string
	typ  ast.Expr
}

// typeSpecs is every type declaration in the package, in one pass.
func typeSpecs(pkgs map[string]*ast.Package) []declaredType {
	var out []declaredType
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						out = append(out, declaredType{ts.Name.Name, ts.Type})
					}
				}
			}
		}
	}
	return out
}

// outwardTypeClosure is every type name in this package that a caller's value can be
// written THROUGH — pointers, slices, maps, channels, funcs and interfaces, and any
// name declared on top of one.
func outwardTypeClosure(pkgs map[string]*ast.Package) map[string]bool {
	specs := typeSpecs(pkgs)
	out := map[string]bool{}
	for round := 0; round < 16; round++ {
		before := len(out)
		for _, s := range specs {
			if out[s.name] {
				continue
			}
			if isOutwardType(s.typ, out) {
				out[s.name] = true
			}
		}
		if len(out) == before {
			break
		}
	}
	return out
}

// isOutwardType reports that a value of this type is a channel an answer can leave
// through: something the callee can write into and the caller can then read.
//
// An array is not one (it is copied); a slice is. A struct passed by value is not
// one; a pointer to it is. Named types are looked up in `named`, which is
// outwardTypeClosure's own fixed point, so `type sink *Verdict` is not a way to spell
// this differently.
func isOutwardType(expr ast.Expr, named map[string]bool) bool {
	switch t := expr.(type) {
	case *ast.StarExpr, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType:
		return true
	case *ast.Ellipsis:
		return true
	case *ast.ArrayType:
		return t.Len == nil // a slice, not an array
	case *ast.ParenExpr:
		return isOutwardType(t.X, named)
	case *ast.Ident:
		return named[t.Name]
	case *ast.SelectorExpr:
		return named[t.Sel.Name]
	}
	return false
}

// buildsOnARecord reports that a type declaration's right-hand side makes the
// declared name another spelling of a record.
func buildsOnARecord(expr ast.Expr, known map[string]bool) bool {
	if st, ok := expr.(*ast.StructType); ok {
		if st.Fields == nil {
			return false
		}
		for _, f := range st.Fields.List {
			// Only embedded fields. A named field holding a record is reached through
			// its own name and is bandShape.fields' job; treating it as a record here
			// would make Verdict a record and forbid the assembly.
			if len(f.Names) == 0 && namesAnyOf(f.Type, known) {
				return true
			}
		}
		return false
	}
	return namesAnyOf(expr, known)
}

// pinnedRecordFields are the struct fields this check must still recognise as
// holding a record.
//
// bandShape computes the field set from the package rather than hard-coding it, and
// that computation deliberately drops a field NAME that is also declared with a
// non-record type somewhere (Acceleration.Crossings is []ThresholdCrossing and
// ShadowBand.Crossings is []BandCrossing — with no type checker the two are one
// name). The dropping is what keeps the acceleration family out of this check, and
// it is also a way to disarm it: declaring `ExtendedBand int` on any struct in this
// package would make the verdict's own slot stop counting as a record. These are
// pinned so that doing so fails here instead of quietly widening the hole.
//
// Four rather than two (S7). The second independent review found that Bands — the
// cycle result's map of tallies — and Quantiles also hold records and neither was
// defended, so the disarming move was available on them too and nothing said so.
// The omission was only that nobody had listed them; they cost nothing to add.
var pinnedRecordFields = []string{"SeenLateBand", "ExtendedBand", "Bands", "Quantiles"}

// pinnedRecordReaders are functions this check must still be finding a record read
// inside.
//
// It is §0.2's floor in the shape this check needs. The failure mode of a structural
// walk is not a wrong answer, it is silence: an expression form that stops being
// recognised takes every function built on it out of the check at once, and the
// check then reports success over the code it no longer reads. These two are the
// package's plainest record readers — a method on the record and the function that
// sums records — and if either stops registering, the walk has lost its grip.
var pinnedRecordReaders = []string{"ShadowBand.Crossed", "TallyBands"}

// bandShape is what one pass over the package's declarations knows about records.
type bandShape struct {
	// fields are struct field names that unambiguously hold a shadow record.
	fields map[string]bool
	// producers are functions whose results include a shadow record.
	producers map[string]bool
	// recordResults says, per function and per result position, whether that result
	// is a record. `a, b := shadowBandsFor(…)` binds both names; `a, b, c :=
	// TallyVerdicts(…)` binds only the third, because the first two are tallies of
	// chases and accelerations and calling them records would be a false alarm.
	recordResults map[string][]bool
	// voidFuncs are functions declared in this package that return nothing. A call
	// to one cannot be carrying a record onwards, because there is nothing for it to
	// carry the record to — whatever it did with the record it did by side effect.
	voidFuncs map[string]bool
	// types is recordTypeClosure's answer: every name in this package that spells a
	// record, aliases and embeddings included.
	types map[string]bool
	// globals are package-level names holding a record. A package var is the one
	// place a record can sit without any function declaring it, so seeding every
	// function with these is what stops `var lastBand ShadowBand` plus a reader
	// somewhere else from being a scope this walk cannot see.
	globals map[string]bool
	// packageVars are ALL package-level var names, record-valued or not. Writing to
	// one is a way out of a function that its signature does not mention.
	packageVars map[string]bool
	// outwardTypes are type names declared in this package whose underlying type is
	// a pointer, slice, map, channel, func or interface — the named spellings of the
	// kinds isOutwardType recognises structurally.
	outwardTypes map[string]bool
}

// typeIdents returns the identifiers a type expression is built out of.
func typeIdents(expr ast.Expr) []string {
	switch t := expr.(type) {
	case nil:
		return nil
	case *ast.Ident:
		return []string{t.Name}
	case *ast.StarExpr:
		return typeIdents(t.X)
	case *ast.ArrayType:
		return typeIdents(t.Elt)
	case *ast.MapType:
		return append(typeIdents(t.Key), typeIdents(t.Value)...)
	case *ast.Ellipsis:
		return typeIdents(t.Elt)
	case *ast.ChanType:
		return typeIdents(t.Value)
	case *ast.ParenExpr:
		return typeIdents(t.X)
	case *ast.SelectorExpr:
		return []string{t.Sel.Name}
	default:
		return nil
	}
}

// namesAnyOf reports that a type expression is built out of one of the given type
// names, through any number of pointers, slices, maps and channels.
func namesAnyOf(expr ast.Expr, names map[string]bool) bool {
	for _, name := range typeIdents(expr) {
		if names[name] {
			return true
		}
	}
	return false
}

// namesARecord reports that a type expression names a record type — the four
// spellings above and every alias, defined type and embedding this package builds on
// them.
func (sh bandShape) namesARecord(expr ast.Expr) bool {
	return namesAnyOf(expr, sh.types)
}

// resultNames flattens a function's declared results into type-name lists, one per
// returned value.
func resultLists(fn *ast.FuncDecl) []ast.Expr {
	if fn.Type.Results == nil {
		return nil
	}
	var out []ast.Expr
	for _, r := range fn.Type.Results.List {
		n := 1
		if len(r.Names) > 0 {
			n = len(r.Names)
		}
		for i := 0; i < n; i++ {
			out = append(out, r.Type)
		}
	}
	return out
}

// newBandShape reads the package's declarations once.
func newBandShape(pkgs map[string]*ast.Package) bandShape {
	sh := bandShape{
		fields:        map[string]bool{},
		producers:     map[string]bool{},
		recordResults: map[string][]bool{},
		voidFuncs:     map[string]bool{},
		globals:       map[string]bool{},
		packageVars:   map[string]bool{},
		// First, because every question below is asked of them.
		types:        recordTypeClosure(pkgs),
		outwardTypes: outwardTypeClosure(pkgs),
	}
	holds, alsoHoldsOther := map[string]bool{}, map[string]bool{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				st, ok := n.(*ast.StructType)
				if !ok || st.Fields == nil {
					return true
				}
				for _, f := range st.Fields.List {
					names := make([]string, 0, len(f.Names))
					for _, id := range f.Names {
						names = append(names, id.Name)
					}
					if len(names) == 0 {
						// An embedded field is named by its own type.
						if ids := typeIdents(f.Type); len(ids) > 0 {
							names = append(names, ids[len(ids)-1])
						}
					}
					for _, name := range names {
						if sh.namesARecord(f.Type) {
							holds[name] = true
						} else {
							alsoHoldsOther[name] = true
						}
					}
				}
				return true
			})
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				results := resultLists(fn)
				per := make([]bool, 0, len(results))
				any := false
				for _, r := range results {
					isRecord := sh.namesARecord(r)
					per = append(per, isRecord)
					any = any || isRecord
				}
				sh.recordResults[fn.Name.Name] = per
				if len(results) == 0 {
					sh.voidFuncs[fn.Name.Name] = true
				}
				if any {
					sh.producers[fn.Name.Name] = true
				}
			}
		}
	}
	for name := range holds {
		if !alsoHoldsOther[name] {
			sh.fields[name] = true
		}
	}
	// Package-level names come last, because deciding whether `var x = f()` holds a
	// record needs the producer set above.
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || (gen.Tok != token.VAR && gen.Tok != token.CONST) {
					continue
				}
				for _, spec := range gen.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, id := range vs.Names {
						if gen.Tok == token.VAR {
							sh.packageVars[id.Name] = true
						}
						if sh.namesARecord(vs.Type) {
							sh.globals[id.Name] = true
							continue
						}
						if i < len(vs.Values) && sh.isRecordExpr(vs.Values[i], sh.globals) {
							sh.globals[id.Name] = true
						}
					}
				}
			}
		}
	}
	return sh
}

// isRecordExpr reports that an expression evaluates to a shadow record.
//
// It is syntax and not types, so it is an under-approximation in one direction only:
// every form it recognises really is a record, and a form it does not recognise —
// an interface, a channel receive, a record round-tripped through fmt — is a residue
// written down in the change's issues.md rather than claimed as covered.
func (sh bandShape) isRecordExpr(expr ast.Expr, bound map[string]bool) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return bound[t.Name]
	case *ast.ParenExpr:
		return sh.isRecordExpr(t.X, bound)
	case *ast.StarExpr:
		return sh.isRecordExpr(t.X, bound)
	case *ast.UnaryExpr:
		return sh.isRecordExpr(t.X, bound)
	case *ast.IndexExpr:
		return sh.isRecordExpr(t.X, bound)
	case *ast.SliceExpr:
		return sh.isRecordExpr(t.X, bound)
	case *ast.SelectorExpr:
		return sh.fields[t.Sel.Name]
	case *ast.CompositeLit:
		return sh.namesARecord(t.Type)
	case *ast.TypeAssertExpr:
		return sh.namesARecord(t.Type)
	case *ast.CallExpr:
		switch fun := t.Fun.(type) {
		case *ast.Ident:
			switch fun.Name {
			case "make", "new":
				return len(t.Args) > 0 && sh.namesARecord(t.Args[0])
			case "append":
				return len(t.Args) > 0 && sh.isRecordExpr(t.Args[0], bound)
			default:
				return sh.producers[fun.Name]
			}
		case *ast.SelectorExpr:
			return sh.producers[fun.Sel.Name]
		}
		return false
	}
	return false
}

// boundRecords are the names inside one function that hold a shadow record.
//
// The receiver, the parameters and the named results come from their declared
// types; everything else is reached by running the assignments to a fixed point, so
// that `b := v.ExtendedBand` and `a, _ := shadowBandsFor(…)` and `for _, b := range
// in` all put the record where the walk below can see it. Without this the check
// would read one spelling and the next reviewer would use a local.
func (sh bandShape) boundRecords(fn *ast.FuncDecl) map[string]bool {
	bound := map[string]bool{}
	for name := range sh.globals {
		bound[name] = true
	}
	declare := func(list *ast.FieldList) {
		if list == nil {
			return
		}
		for _, f := range list.List {
			if !sh.namesARecord(f.Type) {
				continue
			}
			for _, id := range f.Names {
				bound[id.Name] = true
			}
		}
	}
	declare(fn.Recv)
	declare(fn.Type.Params)
	declare(fn.Type.Results)

	bind := func(lhs, rhs []ast.Expr) {
		if len(rhs) == 1 && len(lhs) > 1 {
			bindOne := func(i int) {
				if i >= len(lhs) {
					return
				}
				if id, ok := lhs[i].(*ast.Ident); ok {
					bound[id.Name] = true
				}
			}
			if call, ok := rhs[0].(*ast.CallExpr); ok {
				// One call filling several names, position by position.
				name := ""
				switch fun := call.Fun.(type) {
				case *ast.Ident:
					name = fun.Name
				case *ast.SelectorExpr:
					name = fun.Sel.Name
				}
				for i, isRecord := range sh.recordResults[name] {
					if isRecord {
						bindOne(i)
					}
				}
				return
			}
			// The comma-ok forms — a type assertion, a map index, a channel receive —
			// put the value in the first name and a bool in the second. Without this,
			// `b, _ := boxed.(ShadowBand)` is a laundering step: the record comes back
			// out of an interface into a name this walk does not follow.
			if sh.isRecordExpr(rhs[0], bound) {
				bindOne(0)
			}
			return
		}
		for i, l := range lhs {
			if i >= len(rhs) || !sh.isRecordExpr(rhs[i], bound) {
				continue
			}
			if id, ok := l.(*ast.Ident); ok {
				bound[id.Name] = true
			}
		}
	}

	// A fixed point, because `a := v.ExtendedBand; b := a` needs two passes and
	// nothing says a function is written in dependency order. It converges in as many
	// rounds as the longest chain; the cap is only there so a bug cannot hang a test.
	for round := 0; round < 16; round++ {
		before := len(bound)
		ast.Inspect(fn, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				bind(s.Lhs, s.Rhs)
			case *ast.ValueSpec:
				if sh.namesARecord(s.Type) {
					for _, id := range s.Names {
						bound[id.Name] = true
					}
				}
				lhs := make([]ast.Expr, 0, len(s.Names))
				for _, id := range s.Names {
					lhs = append(lhs, id)
				}
				bind(lhs, s.Values)
			case *ast.RangeStmt:
				if s.Value == nil || !sh.isRecordExpr(s.X, bound) {
					return true
				}
				if id, ok := s.Value.(*ast.Ident); ok {
					bound[id.Name] = true
				}
			}
			return true
		})
		if len(bound) == before {
			break
		}
	}
	return bound
}

// compositeLabel names a composite literal the way a failure has to read it.
func compositeLabel(lit *ast.CompositeLit) string {
	if ids := typeIdents(lit.Type); len(ids) > 0 {
		return ids[len(ids)-1] + "{…}"
	}
	return "a composite literal"
}

// calleeName is the bare name a call is made through, for both `f(x)` and `p.f(x)`.
func calleeName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	case *ast.IndexExpr: // an instantiated generic: f[T](x)
		if id, ok := fun.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

// carriesRecordsSafely reports that handing a record to this callee cannot be the
// conversion, because whatever the callee gives back is a record or is nothing.
//
// A callee this walk cannot see the declaration of — anything in another package,
// and any generic whose result is a type parameter — answers false. That is
// deliberate and it is the third bypass class the battery found: `fmt.Sprintf("%v",
// band)` and `identity(band)` both take a record out through a call and bring it
// back as a value with no record anywhere in its type, which is precisely the
// conversion the requirement forbids, spelled so that nothing else in this file
// would notice.
//
// # A callee that returns nothing is not carrying either (S1)
//
// This used to answer true for a void callee, by running a loop over an empty result
// list and finding nothing that was not a record. Vacuous truth is the wrong answer
// to "is this call carrying the record onwards": a function with no results has
// nowhere to carry it TO, so whatever it did with the record it did by side effect,
// and a side effect is exactly the channel this file was blind to.
func (sh bandShape) carriesRecordsSafely(call *ast.CallExpr) bool {
	name := calleeName(call)
	switch name {
	case "append", "len", "cap", "copy", "make", "new", "delete":
		// Builtins that either hand the same records back or answer about the
		// container and never about a record's contents.
		return true
	}
	results, known := sh.recordResults[name]
	if !known || sh.voidFuncs[name] {
		return false
	}
	for _, isRecord := range results {
		if !isRecord {
			return false
		}
	}
	return true
}

// writesOutward reports that a function has somewhere to put an answer that is not
// one of its results.
//
// # Why an exemption needs this and could not be told without it
//
// Both exemptions below name a SURFACE — "its results are records", "its receiver is
// a record" — and read that surface as "so the answer cannot get out". Go disagrees
// twice over, and the battery for this round found both:
//
//	func (b ShadowBand) decide(v *Verdict)                 // receiver is a record
//	func hand(s Sighting, v *Verdict) (ShadowBand, ShadowBand) // results are records
//
// Each is exempt under the old rule, each reads a band, and each writes a Chase
// through a pointer the caller owns. The caller reads nothing and names nothing, so
// both guards are silent about the whole. That is the third bypass's shape a third
// time, which is why the answer here is not another enumerated case but the question
// the exemptions should have been asking all along: can this function put anything
// anywhere its results do not say it does?
//
// Two channels are recognised. A parameter of a type a value can be written through
// — pointer, slice, map, channel, func, interface, and any name declared on one —
// unless that type names a record, because handing []ShadowBand to TallyBands is
// carrying records and not carrying a decision. And an assignment to a package-level
// variable, which is how a function with no parameters at all answers outward.
//
// Nothing in this package pays for it: every record reader here takes values
// (MeasureExtendedBand) or a slice of records (TallyBands), and none assigns to a
// package var. There are no exceptions and no allowlist.
func (sh bandShape) writesOutward(fn *ast.FuncDecl) bool {
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			if isOutwardType(p.Type, sh.outwardTypes) && !sh.namesARecord(p.Type) {
				return true
			}
		}
	}
	local := localNames(fn)
	written := false
	ast.Inspect(fn, func(n ast.Node) bool {
		var targets []ast.Expr
		switch s := n.(type) {
		case *ast.AssignStmt:
			if s.Tok == token.DEFINE {
				return true
			}
			targets = s.Lhs
		case *ast.IncDecStmt:
			targets = []ast.Expr{s.X}
		case *ast.SendStmt:
			targets = []ast.Expr{s.Chan}
		default:
			return true
		}
		for _, target := range targets {
			if id := rootIdent(target); id != "" && sh.packageVars[id] && !local[id] {
				written = true
			}
		}
		return true
	})
	return written
}

// rootIdent is the name at the base of an assignment target: `x`, `x.f`, `x[i].f`.
func rootIdent(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return rootIdent(t.X)
	case *ast.IndexExpr:
		return rootIdent(t.X)
	case *ast.StarExpr:
		return rootIdent(t.X)
	case *ast.ParenExpr:
		return rootIdent(t.X)
	}
	return ""
}

// localNames are the names one function declares, so that a local shadowing a
// package var is not read as a write to the package var.
func localNames(fn *ast.FuncDecl) map[string]bool {
	out := map[string]bool{}
	declare := func(list *ast.FieldList) {
		if list == nil {
			return
		}
		for _, f := range list.List {
			for _, id := range f.Names {
				out[id.Name] = true
			}
		}
	}
	declare(fn.Recv)
	declare(fn.Type.Params)
	declare(fn.Type.Results)
	ast.Inspect(fn, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			if s.Tok != token.DEFINE {
				return true
			}
			for _, l := range s.Lhs {
				if id, ok := l.(*ast.Ident); ok {
					out[id.Name] = true
				}
			}
		case *ast.ValueSpec:
			for _, id := range s.Names {
				out[id.Name] = true
			}
		case *ast.FuncLit:
			declare(s.Type.Params)
			declare(s.Type.Results)
		}
		return true
	})
	return out
}

// mayReadRecords reports that a function is allowed to look inside a record.
//
// This is the whole rule, and it is local: a function that reads a shadow record
// must hand back a record, or be a method on the record itself, and either way must
// have no way to write an answer outward. A function that reads one and does
// anything else with the answer IS the conversion from measurement to decision —
// that is the shape of every bypass found so far, and it is the shape the
// requirement's third scenario describes.
//
// The rule needs no call graph, which is why it was preferred to walking one: a
// transitive check has to decide what to do about assessOne → shadowBandsFor →
// MeasureExtendedBand, where every hop is legitimate, and every answer to that is an
// exception list. Here shadowBandsFor is allowed because it returns two records and
// BandTally.Collapsed because it is a method on a tally, and neither needed naming.
//
// # Returning nothing was an exemption and it was the third bypass (S1)
//
// The rule used to end "…, hand back nothing, …". The premise was that a function
// which reads a record and returns nothing cannot move a decision. In Go that is
// false, and it is not subtly false:
//
//	func annotate(v *Verdict) {
//	    if v.ExtendedBand.Crossed("6") { v.Chase.Extended = RaisedVeto() }
//	}
//	// assessOne: annotate(&v)
//
// A pointer parameter, a package variable, a channel, a map, a closure — Go has
// several ways for an answer to leave a function that are not its results, and the
// exemption covered all of them at once. The worst spelling is the one above: no
// helper, no hop, and it is the most natural refactor there is ("move the band-to-
// chase wiring out of assessOne"). It passed both guards, `go vet` and 51 packages.
//
// So the exemption is gone rather than qualified. Enumerating the ways out is the
// same mistake one level down — it was already tried twice on the read side — and
// there was nothing to trade for it: no function in this package reads a record and
// returns nothing. A void reader that turns out to be wanted has one honest
// spelling, which is a method on the record type.
func (sh bandShape) mayReadRecords(fn *ast.FuncDecl) bool {
	if sh.writesOutward(fn) {
		return false
	}
	if fn.Recv != nil && len(fn.Recv.List) > 0 && sh.namesARecord(fn.Recv.List[0].Type) {
		return true
	}
	results := resultLists(fn)
	if len(results) == 0 {
		return false
	}
	for _, r := range results {
		if !sh.namesARecord(r) {
			return false
		}
	}
	return true
}

// TestNoFunctionTurnsAShadowRecordIntoSomethingElse closes the hop.
//
// # What it forbids
//
// Reading into a shadow record — any selector applied to a record-valued expression
// — from a function that does not hand a record back. Reading Crossed is the obvious
// one; reading Measured or Value is the same act with a different spelling, and the
// review's own bypass would have survived a check that only knew the word Crossed.
//
// Handing a record to a callee that gives back something which is not a record is
// the same act one step further out. `fmt.Sprintf("%v", band)` reads no field and
// `identity(band)` reads no field, and both bring the record back as a value with no
// record anywhere in its type. Those are refused too, which is why a callee has to
// be one whose results are all records. A callee with no results at all is refused
// for the reason S1 gives: it has nowhere to carry the record TO, so whatever it did
// with it, it did by side effect. Putting a record into a composite literal of a
// type that is not a record is the same act again, spelled without a call.
//
// # What it deliberately allows
//
// Carrying. A record passed whole to a record-returning call, appended whole to a
// slice, stored whole in a field or assigned whole to the verdict's own slot never
// becomes a decision: TallyVerdicts collects both bands out of every verdict and
// hands them to TallyBands, and assessOne fills the two slots from shadowBandsFor.
// Neither looks inside one, and forbidding those would forbid the assembly the
// requirement explicitly protects. Nothing in the package needed an exception.
//
// # Why "must return a record" is not a way out
//
// A function may read a record if it hands one back, so a helper can encode its
// answer in WHICH record it returns. That answer is still a record, and whoever
// wants a decision out of it has to read into it — which puts them under this same
// rule, and under the vocabulary rule above if they produce a verdict. The chain
// terminates in a record every time or it fails here.
//
// # What it does not reach
//
// This is syntax, not types, and it is an early warning rather than the guarantee.
// The guarantee is TestChangingTheScaleChangesNoVerdict, which enumerates no path at
// all. What is outside THIS check is written down in the change's issues.md rather
// than implied, because a guard that has been walked around three times does not get
// to claim it is now complete:
//
//   - other packages. Only internal/candidate is parsed, and RaisedVeto is exported.
//   - test files, which both guards skip.
//   - reflection. reflect.ValueOf(verdict).FieldByName("ExtendedBand") reaches the
//     record without any expression here being record-valued.
//   - a record reachable only through a field name that some other struct in this
//     package also declares with a non-record type. That collision is what keeps
//     Acceleration.Crossings out of here, and pinnedRecordFields is the only part of
//     it that is defended.
//   - anything that reaches a record without naming it: unsafe, or a struct whose
//     record-typed field is itself named ambiguously.
func TestNoFunctionTurnsAShadowRecordIntoSomethingElse(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("parsing the package: %v", err)
	}
	shape := newBandShape(pkgs)

	for _, field := range pinnedRecordFields {
		if !shape.fields[field] {
			t.Fatalf("%s is no longer recognised as holding a shadow record; a field of that "+
				"name declared with some other type anywhere in this package disarms this "+
				"check for every record reached through it. The verdict's own two slots are "+
				"where the 2026-07-28 and 2026-07-29 bypasses both went, and the other two "+
				"hold the tallies a report is built from", field)
		}
	}

	readers := map[string]bool{}
	for _, pkg := range pkgs {
		for path, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				bound := shape.boundRecords(fn)
				var read, handed []string
				ast.Inspect(fn, func(n ast.Node) bool {
					switch e := n.(type) {
					case *ast.SelectorExpr:
						if shape.isRecordExpr(e.X, bound) {
							read = append(read, e.Sel.Name)
						}
					case *ast.CallExpr:
						if shape.carriesRecordsSafely(e) {
							return true
						}
						for _, arg := range e.Args {
							if shape.isRecordExpr(arg, bound) {
								handed = append(handed, calleeName(e))
							}
						}
					case *ast.CompositeLit:
						// A composite literal is a call by another spelling: the record
						// goes in and something that is not a record comes out holding
						// it. `box{v.ExtendedBand}` was in no case of this walk, which is
						// half of how the embedding bypass got through.
						if shape.namesARecord(e.Type) {
							return true
						}
						for _, elt := range e.Elts {
							value := elt
							if kv, ok := elt.(*ast.KeyValueExpr); ok {
								if id, ok := kv.Key.(*ast.Ident); ok && shape.fields[id.Name] {
									// Assembly: a record going into a field declared to
									// hold one. This is the shape the requirement protects.
									continue
								}
								value = kv.Value
							}
							if shape.isRecordExpr(value, bound) {
								handed = append(handed, compositeLabel(e))
							}
						}
					}
					return true
				})
				if len(read) == 0 && len(handed) == 0 {
					continue
				}
				readers[funcLabel(fn)] = true
				if shape.mayReadRecords(fn) {
					continue
				}
				var what []string
				if len(read) > 0 {
					sort.Strings(read)
					what = append(what, "reads "+strings.Join(dedupe(read), ", ")+" out of one")
				}
				if len(handed) > 0 {
					sort.Strings(handed)
					what = append(what, "hands one to "+strings.Join(dedupe(handed), ", "))
				}
				t.Errorf("%s:%s takes a shadow record apart (%s) and does not hand a record "+
					"back. A function that reads a record and does anything else with the "+
					"answer is the conversion from measurement to decision — put one of these "+
					"between a band and a Chase and every other guard here is true on each "+
					"side of it and false about the whole. It must return a record or be a "+
					"method on the record. Returning nothing is not a way out: a pointer "+
					"parameter, a package variable or a channel carries the answer out just "+
					"as well, and that exemption is how the 2026-07-29 bypass passed",
					filepath.Base(path), funcLabel(fn), strings.Join(what, "; "))
			}
		}
	}

	for _, want := range pinnedRecordReaders {
		if !readers[want] {
			t.Fatalf("%s no longer registers as reading a shadow record, so this walk has "+
				"stopped recognising the expression forms the package is written in and is "+
				"reporting success over code it cannot see. Found: %v",
				want, sortedKeys(readers))
		}
	}
	t.Logf("record readers: %v", sortedKeys(readers))
}

// funcLabel names a function the way the failures above have to read it, with the
// receiver in front so that ShadowBand.Crossed is not just "Crossed".
func funcLabel(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	ids := typeIdents(fn.Recv.List[0].Type)
	if len(ids) == 0 {
		return fn.Name.Name
	}
	return ids[len(ids)-1] + "." + fn.Name.Name
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func sortedKeys(in map[string]bool) []string {
	out := make([]string, 0, len(in))
	for k := range in {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
