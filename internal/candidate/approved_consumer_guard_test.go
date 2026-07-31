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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var approvedCandidateSymbols = map[string]bool{
	"ApprovedCandidate":       true,
	"AssessApprovedCandidate": true,
}

// These selectors add diagnostic detail after import-graph tainting. They do not
// decide which packages are tainted: Valid is also a protected-stop predicate in
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

// A package joining approved-candidate-derived decisions to authority requires a
// permission distinct from the pure reader boundary. a047 may add only a reviewed
// Guardian/decision bridge here, with a non-empty reason. A general boundary
// permission cannot authorize execution, risk, ledger, or engine reachability.
var approvedCandidateAuthorityBridges = map[string]string{}

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
	_, blocked := forbidden[rel]
	return blocked
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

func transitiveAuthorityDependency(graph map[string][]string, start string) ([]string, bool) {
	return transitiveDependency(graph, start, forbiddenApprovedBoundaryPackage)
}

func auditApprovedCandidateBoundaries(root, module string, files []string) ([]string, error) {
	type productionFile struct {
		rel, packageRel string
		file            *ast.File
	}
	var parsedFiles []productionFile
	imports := make(map[string][]string)
	directReaders := make(map[string]bool)
	allPackages := make(map[string]bool)
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
		allPackages[packageRel] = true
		parsedFiles = append(parsedFiles, productionFile{rel: rel, packageRel: packageRel, file: parsed})
		candidatePkg, importsCandidate := candidateImportName(parsed, module)
		if importsCandidate && candidatePkg == "." {
			directReaders[packageRel] = true
		} else if namesApprovedCandidateSymbol(parsed, candidatePkg) {
			directReaders[packageRel] = true
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
	tainted := make(map[string]bool, len(directReaders))
	for packageRel := range directReaders {
		tainted[packageRel] = true
	}
	taintedFrom := make(map[string][]string)
	for packageRel := range allPackages {
		if directReaders[packageRel] {
			continue
		}
		path, reachesReader := transitiveDependency(imports, packageRel, func(dependency string) bool {
			return directReaders[dependency]
		})
		if reachesReader {
			tainted[packageRel] = true
			taintedFrom[packageRel] = path
		}
	}

	var findings []string
	for packageRel := range directReaders {
		reason, allowed := approvedCandidateBoundaries[packageRel]
		if !allowed {
			detail := ""
			if len(accessorFiles[packageRel]) != 0 {
				detail = "; package-level accessors in " + strings.Join(accessorFiles[packageRel], ", ")
			}
			findings = append(findings, packageRel+
				" reads ApprovedCandidate but is not an explicit pure boundary"+detail)
		}
		if allowed && strings.TrimSpace(reason) == "" {
			findings = append(findings, packageRel+" has an empty approved-candidate boundary reason")
		}
	}
	for packageRel := range approvedCandidateBoundaries {
		if !directReaders[packageRel] {
			findings = append(findings, packageRel+" is allowlisted but no production file reads ApprovedCandidate")
		}
	}
	for packageRel := range tainted {
		path, bad := transitiveAuthorityDependency(imports, packageRel)
		reason, bridged := approvedCandidateAuthorityBridges[packageRel]
		if bad && !bridged {
			detail := ""
			if len(taintedFrom[packageRel]) != 0 {
				detail = "; approved-candidate taint via " + strings.Join(taintedFrom[packageRel], " -> ")
			}
			findings = append(findings, packageRel+" reaches an authority package: "+
				strings.Join(path, " -> ")+detail)
		}
		if bad && bridged && strings.TrimSpace(reason) == "" {
			findings = append(findings, packageRel+" has an empty approved-candidate authority bridge reason")
		}
		if !bad && bridged {
			findings = append(findings, packageRel+" has an authority bridge permission but reaches no authority package")
		}
	}
	for packageRel := range approvedCandidateAuthorityBridges {
		if !tainted[packageRel] {
			findings = append(findings, packageRel+" has an authority bridge permission but no approved-candidate taint")
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

func TestApprovedCandidateBoundaryRejectsReversePrimitiveLaundering(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root := t.TempDir()
	files := map[string]string{
		"internal/strategy/strategy.go": `package strategy
import cv "` + module + `/internal/candidate"
func Eligible(in cv.VetoInputs, set cv.ThresholdSet) (bool, error) {
	_, err := cv.AssessApprovedCandidate(in, set)
	return err == nil, err
}`,
		"internal/testengine/engine.go": `package testengine
import (
	"` + module + `/internal/strategy"
	"` + module + `/internal/execgw"
)
func Run() { _, _ = strategy.Eligible; _ = execgw.Submit }`,
	}
	var rels []string
	for rel, src := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating fixture directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
			t.Fatalf("writing fixture %s: %v", rel, err)
		}
		rels = append(rels, rel)
	}

	oldBoundaries := approvedCandidateBoundaries
	oldBridges := approvedCandidateAuthorityBridges
	defer func() {
		approvedCandidateBoundaries = oldBoundaries
		approvedCandidateAuthorityBridges = oldBridges
	}()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy":   "pure bool/error strategy boundary fixture",
		"internal/testengine": "a general boundary permission must not authorize execution",
	}
	approvedCandidateAuthorityBridges = map[string]string{}

	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing reverse primitive fixture: %v", err)
	}
	want := "internal/testengine reaches an authority package: internal/testengine -> internal/execgw"
	if !containsFinding(findings, want) {
		t.Fatalf("reverse primitive laundering findings = %v, want %q", findings, want)
	}

	approvedCandidateBoundaries = map[string]string{
		"internal/strategy": "pure bool/error strategy boundary fixture",
	}
	approvedCandidateAuthorityBridges = map[string]string{
		"internal/testengine": " ",
	}
	findings, err = auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing empty bridge fixture: %v", err)
	}
	if !containsFinding(findings, "internal/testengine has an empty approved-candidate authority bridge reason") {
		t.Fatalf("empty bridge findings = %v", findings)
	}

	approvedCandidateAuthorityBridges = map[string]string{
		"internal/testengine": "explicit Guardian/decision bridge fixture",
		"internal/stale":      "must not survive without approved-candidate taint",
	}
	findings, err = auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing explicit bridge fixture: %v", err)
	}
	if !containsFinding(findings, "internal/stale has an authority bridge permission but no approved-candidate taint") {
		t.Fatalf("stale bridge findings = %v", findings)
	}
	for _, finding := range findings {
		if strings.Contains(finding, "internal/testengine") {
			t.Fatalf("explicit bridge still rejected: %v", findings)
		}
	}
}

func TestApprovedCandidateBoundaryDetectsTransitiveJournalDependency(t *testing.T) {
	graph := map[string][]string{
		"internal/strategy": {"internal/journal"},
	}
	path, bad := transitiveAuthorityDependency(graph, "internal/strategy")
	if !bad {
		t.Fatal("approved boundary dependency guard missed strategy -> journal")
	}
	if got := strings.Join(path, " -> "); got != "internal/strategy -> internal/journal" {
		t.Fatalf("dependency path = %q", got)
	}
}

func TestApprovedCandidateBoundaryDetectsAllAuthorityRoots(t *testing.T) {
	if len(forbidden) < 10 {
		t.Fatalf("authoritative forbidden registry unexpectedly small: %d", len(forbidden))
	}
	for authority := range forbidden {
		if !forbiddenApprovedBoundaryPackage(authority) {
			t.Errorf("authority root %s is absent from the approved-boundary guard", authority)
		}
	}
	if forbiddenApprovedBoundaryPackage("internal/strategy") {
		t.Error("unrelated pure strategy package classified as an authority root")
	}
}

func containsFinding(findings []string, want string) bool {
	for _, finding := range findings {
		if strings.HasPrefix(finding, want) {
			return true
		}
	}
	return false
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
