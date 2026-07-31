package candidate

// This file is the a047-facing half of consumer_guard_test.go. The older guard
// is file-local because cmd/tossctl already contains order authority. An
// ApprovedCandidate is different: its first production consumer must be a pure
// package boundary, so a same-package helper cannot launder .Valid() into a file
// that reaches orders, and the boundary cannot import an order package through a
// helper dependency either.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

var approvedCandidateSymbols = map[string]bool{
	"ApprovedCandidate":       true,
	"AssessApprovedCandidate": true,
}

// These selectors are considered only inside a package that directly names the
// approved type or constructor. Valid is also a protected-stop predicate in
// internal/exitpolicy, so a repository-wide spelling scan would manufacture a
// security permission for an unrelated safety package.
var approvedCandidateAccessors = map[string]bool{
	"Valid": true, "Key": true, "FirstSeenAt": true, "Chase": true,
	"CandidateLifeID": true, "ThresholdVersion": true, "SetDigest": true,
	"EvidenceDigest": true, "ApprovedAt": true,
}

// Intentionally empty in a046: there is no production ApprovedCandidate reader.
// a047 must add its pure strategy package and a non-empty reason in the same diff
// that adds the first reader. Stale permissions fail below.
var approvedCandidateBoundaries = map[string]string{}

func namesApprovedCandidateSymbol(file *ast.File, pkg string) bool {
	if file == nil || pkg == "" {
		return false
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if ok && id.Name == pkg && approvedCandidateSymbols[sel.Sel.Name] {
			found = true
		}
		return true
	})
	return found
}

func namesApprovedCandidateAccessor(file *ast.File) bool {
	if file == nil {
		return false
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if ok && approvedCandidateAccessors[sel.Sel.Name] {
			found = true
		}
		return true
	})
	return found
}

func forbiddenApprovedBoundaryPackage(rel string) bool {
	for _, suffix := range orderPackagePaths {
		if rel == strings.TrimPrefix(suffix, "/") {
			return true
		}
	}
	return false
}

func transitiveDependency(graph map[string][]string, start string, matches func(string) bool) ([]string, bool) {
	type pathNode struct {
		name string
		path []string
	}
	queue := []pathNode{{name: start, path: []string{start}}}
	seen := map[string]bool{}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current.name] {
			continue
		}
		seen[current.name] = true
		if matches(current.name) {
			return current.path, true
		}
		for _, dependency := range graph[current.name] {
			path := append(append([]string(nil), current.path...), dependency)
			queue = append(queue, pathNode{name: dependency, path: path})
		}
	}
	return nil, false
}

func transitiveOrderDependency(graph map[string][]string, start string) ([]string, bool) {
	return transitiveDependency(graph, start, forbiddenApprovedBoundaryPackage)
}

func auditApprovedCandidateBoundaries(root, module string, files []string) ([]string, error) {
	type productionFile struct {
		rel, packageRel string
		file            *ast.File
	}
	var parsedFiles []productionFile
	imports := make(map[string][]string)
	readers := make(map[string]bool)
	accessorFiles := make(map[string][]string)
	fset := token.NewFileSet()
	for _, rel := range files {
		if strings.HasSuffix(rel, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, filepath.Join(root, rel), nil,
			parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", rel, err)
		}
		packageRel := filepath.ToSlash(filepath.Dir(rel))
		parsedFiles = append(parsedFiles, productionFile{rel: rel, packageRel: packageRel, file: parsed})
		candidatePkg, importsCandidate := candidateImportName(parsed, module)
		if importsCandidate && candidatePkg == "." {
			readers[packageRel] = true
		} else if namesApprovedCandidateSymbol(parsed, candidatePkg) {
			readers[packageRel] = true
		}
		for _, spec := range parsed.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if path == module {
				imports[packageRel] = append(imports[packageRel], ".")
			} else if strings.HasPrefix(path, module+"/") {
				imports[packageRel] = append(imports[packageRel], strings.TrimPrefix(path, module+"/"))
			}
		}
	}
	for _, record := range parsedFiles {
		if namesApprovedCandidateAccessor(record.file) {
			accessorFiles[record.packageRel] = append(accessorFiles[record.packageRel], record.rel)
		}
	}
	directReaders := make(map[string]bool, len(readers))
	for packageRel := range readers {
		directReaders[packageRel] = true
	}
	launderedFrom := make(map[string][]string)
	for packageRel := range accessorFiles {
		if readers[packageRel] {
			continue
		}
		path, reachesReader := transitiveDependency(imports, packageRel, func(dependency string) bool {
			return directReaders[dependency]
		})
		if reachesReader {
			readers[packageRel] = true
			launderedFrom[packageRel] = path
		}
	}

	var findings []string
	for packageRel := range readers {
		reason, allowed := approvedCandidateBoundaries[packageRel]
		if !allowed {
			detail := ""
			if len(accessorFiles[packageRel]) != 0 {
				detail = "; package-level accessors in " + strings.Join(accessorFiles[packageRel], ", ")
			}
			if len(launderedFrom[packageRel]) != 0 {
				detail += "; reaches approved boundary via " + strings.Join(launderedFrom[packageRel], " -> ")
			}
			findings = append(findings, packageRel+
				" reads ApprovedCandidate but is not an explicit pure boundary"+detail)
		}
		if allowed && strings.TrimSpace(reason) == "" {
			findings = append(findings, packageRel+" has an empty approved-candidate boundary reason")
		}
		if path, bad := transitiveOrderDependency(imports, packageRel); bad {
			findings = append(findings, packageRel+" reaches an order package: "+strings.Join(path, " -> "))
		}
	}
	for packageRel := range approvedCandidateBoundaries {
		if !readers[packageRel] {
			findings = append(findings, packageRel+" is allowlisted but no production file reads ApprovedCandidate")
		}
	}
	return findings, nil
}

func TestApprovedCandidateGuardRejectsAliasedErrNilOrderConsumer(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	src := `package main
import (
	cv "` + module + `/internal/candidate"
	eg "` + module + `/internal/execgw"
)
func f(in cv.VetoInputs, set cv.ThresholdSet) {
	_, err := cv.AssessApprovedCandidate(in, set)
	if err == nil {
		eg.Submit()
	}
}`
	parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", src,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	pkg, imports := candidateImportName(parsed, module)
	if !imports || pkg != "cv" {
		t.Fatalf("candidate alias = (%q, %t), want (cv, true)", pkg, imports)
	}
	if !namesVerdict(parsed, pkg) {
		t.Fatal("verdict detector missed aliased AssessApprovedCandidate guarded only by err == nil")
	}
	if verb, bad := namesAnOrderVerb(parsed, module); !bad {
		t.Fatal("order detector missed the order selector in approved-candidate fixture")
	} else {
		t.Logf("positive control detected approved constructor plus %s", verb)
	}
}

func TestApprovedCandidateBoundaryDetectsHelperAccessorLaundering(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	helper := `package strategy
import cv "` + module + `/internal/candidate"
func latest() cv.ApprovedCandidate { return cv.ApprovedCandidate{} }`
	consumer := `package strategy
func enabled() bool {
	approved := latest()
	return approved.Valid()
}`
	helperFile, err := parser.ParseFile(token.NewFileSet(), "helper.go", helper,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing helper fixture: %v", err)
	}
	consumerFile, err := parser.ParseFile(token.NewFileSet(), "consumer.go", consumer,
		parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing consumer fixture: %v", err)
	}
	if !namesApprovedCandidateSymbol(helperFile, "cv") {
		t.Fatal("approved boundary detector missed the helper's aliased ApprovedCandidate type")
	}
	if !namesApprovedCandidateAccessor(consumerFile) {
		t.Fatal("approved boundary detector missed .Valid() in another file of the same package")
	}
	graph := map[string][]string{"internal/orchestrator": {"internal/strategy"}}
	path, reachesReader := transitiveDependency(graph, "internal/orchestrator", func(dependency string) bool {
		return dependency == "internal/strategy"
	})
	if !reachesReader || strings.Join(path, " -> ") != "internal/orchestrator -> internal/strategy" {
		t.Fatalf("cross-package accessor laundering path = %v, detected=%t", path, reachesReader)
	}
}

func TestApprovedCandidateBoundaryDetectsTransitiveOrderDependency(t *testing.T) {
	graph := map[string][]string{
		"internal/strategy":       {"internal/strategyhelper"},
		"internal/strategyhelper": {"internal/execgw"},
	}
	path, bad := transitiveOrderDependency(graph, "internal/strategy")
	if !bad {
		t.Fatal("approved boundary dependency guard missed strategy -> helper -> execgw")
	}
	if got := strings.Join(path, " -> "); got != "internal/strategy -> internal/strategyhelper -> internal/execgw" {
		t.Fatalf("dependency path = %q", got)
	}
}

func TestApprovedCandidateConsumersStayInsidePureBoundaries(t *testing.T) {
	root := moduleRoot(t)
	findings, err := auditApprovedCandidateBoundaries(root, modulePath(t, root), goFilesUnder(t, root))
	if err != nil {
		t.Fatalf("auditing approved-candidate boundaries: %v", err)
	}
	for _, finding := range findings {
		t.Error(finding)
	}
}
