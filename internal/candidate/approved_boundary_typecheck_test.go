package candidate

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

var safePureBoundaryBuiltins = map[string]bool{
	"cap": true,
	"len": true,
	"max": true,
	"min": true,
}

type pureBoundaryImporter struct {
	candidatePath string
	candidatePkg  *types.Package
}

func (i pureBoundaryImporter) Import(path string) (*types.Package, error) {
	if path == i.candidatePath {
		return i.candidatePkg, nil
	}
	return nil, fmt.Errorf("pure boundary import %q is not allowed", path)
}

func syntheticCandidatePackage(path string) *types.Package {
	pkg := types.NewPackage(path, "candidate")
	approvedName := types.NewTypeName(token.NoPos, pkg, "ApprovedCandidate", nil)
	approved := types.NewNamed(approvedName, types.NewStruct(nil, nil), nil)
	pkg.Scope().Insert(approvedName)
	for name := range approvedCandidateAccessors {
		result := types.Typ[types.String]
		if name == "Valid" {
			result = types.Typ[types.Bool]
		} else if name == "FirstSeenUnixNano" || name == "LastSeenUnixNano" ||
			name == "ValidUntilUnixNano" || name == "ApprovedAtUnixNano" {
			result = types.Typ[types.Int64]
		}
		receiver := types.NewVar(token.NoPos, pkg, "approved", approved)
		signature := types.NewSignatureType(
			receiver,
			nil,
			nil,
			types.NewTuple(),
			types.NewTuple(types.NewVar(token.NoPos, pkg, "", result)),
			false,
		)
		approved.AddMethod(types.NewFunc(token.NoPos, pkg, name, signature))
	}
	orderedCodesResult := types.NewArray(types.Typ[types.String], 1)
	orderedCodesSignature := types.NewSignatureType(
		nil, nil, nil, types.NewTuple(),
		types.NewTuple(types.NewVar(token.NoPos, pkg, "", orderedCodesResult)), false,
	)
	pkg.Scope().Insert(types.NewFunc(token.NoPos, pkg, "OrderedVetoCodes", orderedCodesSignature))
	sourceName := types.NewTypeName(token.NoPos, pkg, "Source", nil)
	sourceInterface := types.NewInterfaceType(nil, nil)
	sourceInterface.Complete()
	types.NewNamed(sourceName, sourceInterface, nil)
	pkg.Scope().Insert(sourceName)
	pkg.MarkComplete()
	return pkg
}

func exactApprovedCandidateType(typ types.Type, candidatePath string) bool {
	typ = types.Unalias(typ)
	named, ok := typ.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == candidatePath && named.Obj().Name() == "ApprovedCandidate"
}

func pureBoundaryTypeReason(typ types.Type, candidatePath string, visiting map[types.Type]bool) string {
	if typ == nil {
		return "unresolved type"
	}
	typ = types.Unalias(typ)
	if exactApprovedCandidateType(typ, candidatePath) {
		return ""
	}
	if visiting[typ] {
		return "recursive type cycle"
	}
	visiting[typ] = true
	defer delete(visiting, typ)

	switch value := typ.(type) {
	case *types.Basic:
		if value.Kind() == types.UnsafePointer || value.Kind() == types.Uintptr ||
			value.Kind() == types.Invalid || value.Kind() == types.UntypedNil {
			return value.String() + " is not an immutable scalar"
		}
		info := value.Info()
		if info&(types.IsBoolean|types.IsInteger|types.IsFloat|types.IsComplex|types.IsString) == 0 {
			return value.String() + " is not an immutable scalar"
		}
		return ""
	case *types.Named:
		if object := value.Obj(); object != nil && object.Pkg() != nil && object.Pkg().Path() == candidatePath {
			return "candidate." + object.Name() + " is not exact candidate.ApprovedCandidate"
		}
		if parameters := value.TypeParams(); parameters != nil && parameters.Len() != 0 &&
			(value.TypeArgs() == nil || value.TypeArgs().Len() == 0) {
			return "generic type contains unresolved type parameters"
		}
		if arguments := value.TypeArgs(); arguments != nil {
			for index := 0; index < arguments.Len(); index++ {
				if reason := pureBoundaryTypeReason(arguments.At(index), candidatePath, visiting); reason != "" {
					return fmt.Sprintf("type argument %d: %s", index, reason)
				}
			}
		}
		return pureBoundaryTypeReason(value.Underlying(), candidatePath, visiting)
	case *types.Struct:
		for index := 0; index < value.NumFields(); index++ {
			field := value.Field(index)
			if reason := pureBoundaryTypeReason(field.Type(), candidatePath, visiting); reason != "" {
				return "field " + field.Name() + ": " + reason
			}
		}
		return ""
	case *types.Array:
		if reason := pureBoundaryTypeReason(value.Elem(), candidatePath, visiting); reason != "" {
			return "fixed-array element: " + reason
		}
		return ""
	case *types.Pointer:
		return "pointer is capability-bearing"
	case *types.Map:
		return "map is mutable"
	case *types.Slice:
		return "slice is mutable"
	case *types.Chan:
		return "channel carries synchronization capability"
	case *types.Signature:
		return "function signature is executable capability"
	case *types.Interface:
		return "interface (including any/error) is capability-bearing"
	case *types.TypeParam:
		return "type parameter is not a closed value type"
	case *types.Union:
		return "union constraint is not a runtime value type"
	case *types.Tuple:
		return "tuple is not a boundary value type"
	default:
		return fmt.Sprintf("unsupported type %T", typ)
	}
}

func appendPureTypeFinding(findings []string, packageRel, path string, typ types.Type, candidatePath string) []string {
	if reason := pureBoundaryTypeReason(typ, candidatePath, map[types.Type]bool{}); reason != "" {
		findings = append(findings, packageRel+" pure boundary type contract rejects "+path+": "+reason)
	}
	return findings
}

func boundaryParents(file *ast.File) map[ast.Node]ast.Node {
	parents := map[ast.Node]ast.Node{}
	stack := []ast.Node{}
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) != 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})
	return parents
}

func forbiddenAssignmentKind(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.ParenExpr:
		return forbiddenAssignmentKind(value.X)
	case *ast.StarExpr:
		return "dereference assignment"
	case *ast.IndexExpr, *ast.IndexListExpr:
		return "index assignment"
	case *ast.SelectorExpr:
		return "selector assignment"
	}
	return ""
}

func allowedApprovedAccessorCall(call *ast.CallExpr, info *types.Info, candidatePath string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !approvedCandidateAccessors[selector.Sel.Name] {
		return false
	}
	selection := info.Selections[selector]
	return selection != nil && selection.Kind() == types.MethodVal &&
		exactApprovedCandidateType(selection.Recv(), candidatePath)
}

func allowedPureBuiltinCall(call *ast.CallExpr, info *types.Info) bool {
	identifier, ok := call.Fun.(*ast.Ident)
	if !ok || !safePureBoundaryBuiltins[identifier.Name] {
		return false
	}
	_, builtin := info.Uses[identifier].(*types.Builtin)
	return builtin
}

func typeCheckPureApprovedCandidateBoundary(
	packageRel, module string,
	fset *token.FileSet,
	files []*ast.File,
) []string {
	candidatePath := module + "/internal/candidate"
	sealedSnapshotBoundary := false
	if packageRel == "internal/strategy" {
		for _, file := range files {
			ast.Inspect(file, func(node ast.Node) bool {
				if spec, ok := node.(*ast.TypeSpec); ok && spec.Name.Name == "ApprovedSnapshot" {
					sealedSnapshotBoundary = true
				}
				return true
			})
		}
	}
	var findings []string
	for _, file := range files {
		for _, spec := range file.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if path != candidatePath {
				findings = append(findings, packageRel+" pure boundary imports "+path+
					"; only "+candidatePath+" is allowed")
			}
		}
	}

	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	var typeErrors []error
	configuration := types.Config{
		Importer: pureBoundaryImporter{
			candidatePath: candidatePath,
			candidatePkg:  syntheticCandidatePackage(candidatePath),
		},
		Error: func(err error) { typeErrors = append(typeErrors, err) },
	}
	_, checkErr := configuration.Check(module+"/"+packageRel, fset, files, info)
	if checkErr != nil {
		for _, typeErr := range typeErrors {
			findings = append(findings, packageRel+" pure boundary type check failed: "+typeErr.Error())
		}
		if len(typeErrors) == 0 {
			findings = append(findings, packageRel+" pure boundary type check failed: "+checkErr.Error())
		}
	}

	for _, file := range files {
		parents := boundaryParents(file)
		ast.Inspect(file, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.GenDecl:
				if value.Tok == token.VAR {
					if _, topLevel := parents[value].(*ast.File); topLevel {
						for _, spec := range value.Specs {
							declaration := spec.(*ast.ValueSpec)
							for _, name := range declaration.Names {
								findings = append(findings, packageRel+" pure boundary forbids package variable "+name.Name)
							}
						}
					}
				}
			case *ast.TypeSpec:
				if object := info.Defs[value.Name]; object != nil {
					findings = appendPureTypeFinding(findings, packageRel,
						"type "+value.Name.Name, object.Type(), candidatePath)
				}
			case *ast.FuncDecl:
				if sealedSnapshotBoundary && !allowedApprovedSnapshotDeclaration(value) {
					findings = append(findings, packageRel+" pure boundary forbids laundering declaration "+value.Name.Name)
				}
				if value.Recv != nil && !approvedSnapshotMethod(value) {
					findings = append(findings, packageRel+" pure boundary forbids method declaration "+value.Name.Name)
				}
				if value.Name.Name == "init" {
					findings = append(findings, packageRel+" pure boundary forbids init function")
				}
				if function, ok := info.Defs[value.Name].(*types.Func); ok {
					if signature, ok := function.Type().(*types.Signature); ok {
						if receiver := signature.Recv(); receiver != nil {
							findings = appendPureTypeFinding(findings, packageRel,
								"receiver of "+value.Name.Name, receiver.Type(), candidatePath)
						}
						for index := 0; index < signature.Params().Len(); index++ {
							parameter := signature.Params().At(index)
							findings = appendPureTypeFinding(findings, packageRel,
								"parameter "+parameter.Name(), parameter.Type(), candidatePath)
						}
						for index := 0; index < signature.Results().Len(); index++ {
							result := signature.Results().At(index)
							name := result.Name()
							if name == "" {
								name = fmt.Sprintf("#%d", index)
							}
							findings = appendPureTypeFinding(findings, packageRel,
								"result "+name, result.Type(), candidatePath)
						}
						if parameters := signature.TypeParams(); parameters != nil && parameters.Len() != 0 {
							findings = append(findings, packageRel+" pure boundary type contract rejects generic function "+value.Name.Name)
						}
					}
				}
			case *ast.ValueSpec:
				for _, name := range value.Names {
					if object := info.Defs[name]; object != nil {
						findings = appendPureTypeFinding(findings, packageRel,
							"variable "+name.Name, object.Type(), candidatePath)
					}
				}
			case *ast.AssignStmt:
				for _, left := range value.Lhs {
					if kind := forbiddenAssignmentKind(left); kind != "" {
						findings = append(findings, packageRel+" pure boundary forbids "+kind)
					}
				}
				for index, right := range value.Rhs {
					findings = appendPureTypeFinding(findings, packageRel,
						fmt.Sprintf("assignment value #%d", index), info.TypeOf(right), candidatePath)
				}
			case *ast.IncDecStmt:
				if kind := forbiddenAssignmentKind(value.X); kind != "" {
					findings = append(findings, packageRel+" pure boundary forbids "+kind)
				}
			case *ast.SendStmt:
				findings = append(findings, packageRel+" pure boundary forbids channel send")
			case *ast.GoStmt:
				findings = append(findings, packageRel+" pure boundary forbids go statement")
			case *ast.DeferStmt:
				findings = append(findings, packageRel+" pure boundary forbids defer statement")
			case *ast.FuncLit:
				findings = append(findings, packageRel+" pure boundary forbids function literal")
			case *ast.TypeAssertExpr:
				findings = appendPureTypeFinding(findings, packageRel,
					"asserted type", info.TypeOf(value.Type), candidatePath)
			case *ast.CallExpr:
				if allowedApprovedAccessorCall(value, info, candidatePath) || allowedPureBuiltinCall(value, info) {
					break
				}
				findings = append(findings, packageRel+" pure boundary forbids free or injected function call")
			case *ast.SelectorExpr:
				selection := info.Selections[value]
				if selection == nil {
					break
				}
				if _, method := selection.Obj().(*types.Func); !method {
					break
				}
				if call, direct := parents[value].(*ast.CallExpr); direct && call.Fun == value {
					break
				}
				findings = append(findings, packageRel+" pure boundary forbids method value or expression "+value.Sel.Name)
			}
			return true
		})
	}
	return findings
}

func approvedSnapshotMethod(function *ast.FuncDecl) bool {
	wantResults := map[string]string{
		"Valid":              "bool",
		"Market":             "string",
		"Symbol":             "string",
		"State":              "string",
		"CandidateLifeID":    "string",
		"ThresholdVersion":   "string",
		"SetDigest":          "string",
		"EvidenceDigest":     "string",
		"FirstSeenUnixNano":  "int64",
		"LastSeenUnixNano":   "int64",
		"ValidUntilUnixNano": "int64",
		"ApprovedAtUnixNano": "int64",
	}
	if function == nil || function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	receiver, ok := function.Recv.List[0].Type.(*ast.Ident)
	if !ok || receiver.Name != "ApprovedSnapshot" {
		return false
	}
	want, allowed := wantResults[function.Name.Name]
	if !allowed || function.Type.Params.NumFields() != 0 || function.Type.Results == nil || function.Type.Results.NumFields() != 1 {
		return false
	}
	result, ok := function.Type.Results.List[0].Type.(*ast.Ident)
	return ok && result.Name == want
}

func allowedApprovedSnapshotDeclaration(function *ast.FuncDecl) bool {
	if function == nil {
		return false
	}
	if function.Recv != nil {
		return approvedSnapshotMethod(function)
	}
	return function.Name.Name == "SealApproved"
}
