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
	"Valid": true, "Key": true, "State": true, "FirstSeenAt": true, "LastSeenAt": true, "ValidUntil": true, "Chase": true,
	"CandidateLifeID": true, "ThresholdVersion": true, "SetDigest": true,
	"EvidenceDigest": true, "ApprovedAt": true,
	"MarketString": true, "SymbolString": true, "StateString": true, "CandidateLifeIDString": true,
	"FirstSeenUnixNano": true, "LastSeenUnixNano": true, "ValidUntilUnixNano": true, "ApprovedAtUnixNano": true,
}

// Intentionally empty in a046: there is no production ApprovedCandidate reader.
// a047 must add its pure strategy package and a non-empty reason in the same diff
// that adds the first reader. Stale permissions fail below.
var approvedCandidateBoundaries = map[string]string{
	"internal/strategy": "a047 value-only immutable handoff; no authority, clock, callback, or mutable input",
}

// strategyengine is the audited sanitizing boundary: it accepts the sealed
// ApprovedSnapshot and emits an opaque Decision whose fields cannot be minted
// by downstream packages. The engine itself remains tainted and is audited;
// imports of its opaque result do not inherit candidate authority.
var approvedCandidateSanitizers = map[string]bool{"internal/strategyengine": true}

// a046 intentionally defines no authority-bridge exemption. If a047 introduces
// a typed Guardian/decision bridge, it must replace the unconditional rejection
// below together with a test that compares the complete reachable authority
// root/path set for exact equality (both added and missing roots must fail).

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

func pureApprovedCandidateBoundaryViolations(packageRel, module string, fset *token.FileSet, files []*ast.File) []string {
	return typeCheckPureApprovedCandidateBoundary(packageRel, module, fset, files)
}

var authorityMethodNames = map[string]bool{
	"Amend": true, "Authorize": true, "Cancel": true, "Dispatch": true,
	"Execute": true, "Place": true, "Plan": true, "RecordAttempt": true,
	"Submit": true,
}

func authorityInterfaceDeclarations(packageRel string, files []*ast.File) []string {
	var findings []string
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			typeSpec, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			contract, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}
			for _, field := range contract.Methods.List {
				for _, name := range field.Names {
					if authorityMethodNames[name.Name] {
						findings = append(findings, packageRel+" approved-candidate taint forbids authority interface "+typeSpec.Name.Name+"."+name.Name)
					}
				}
			}
			return true
		})
	}
	return findings
}

func TestApprovedCandidateTaintRejectsAuthorityInterface(t *testing.T) {
	parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", `package lane
type Gateway interface { Place() error }
type Reader interface { ReadSymbolState() error }`, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	findings := authorityInterfaceDeclarations("internal/lane", []*ast.File{parsed})
	if !findingContains(findings, "Gateway.Place") || findingContains(findings, "Reader.ReadSymbolState") {
		t.Fatalf("authority findings=%v", findings)
	}
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
	filesByPackage := make(map[string][]*ast.File)
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
		filesByPackage[packageRel] = append(filesByPackage[packageRel], parsed)
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
		path, reachesReader := transitiveDependencyWithStop(imports, packageRel, func(dependency string) bool {
			return directReaders[dependency]
		}, func(dependency string) bool {
			return dependency != packageRel && approvedCandidateSanitizers[dependency]
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
		findings = append(findings,
			pureApprovedCandidateBoundaryViolations(packageRel, module, fset, filesByPackage[packageRel])...)
	}
	for packageRel := range approvedCandidateBoundaries {
		if !directReaders[packageRel] {
			findings = append(findings, packageRel+" is allowlisted but no production file reads ApprovedCandidate")
		}
	}
	for packageRel := range tainted {
		findings = append(findings, authorityInterfaceDeclarations(packageRel, filesByPackage[packageRel])...)
		path, bad := transitiveAuthorityDependency(imports, packageRel)
		if bad {
			detail := ""
			if len(taintedFrom[packageRel]) != 0 {
				detail = "; approved-candidate taint via " + strings.Join(taintedFrom[packageRel], " -> ")
			}
			findings = append(findings, packageRel+" reaches an authority package: "+
				strings.Join(path, " -> ")+detail)
		}
	}
	return findings, nil
}

func transitiveDependencyWithStop(graph map[string][]string, start string, matches, stop func(string) bool) ([]string, bool) {
	type node struct {
		name string
		path []string
	}
	queue := []node{{name: start, path: []string{start}}}
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
		if stop(current.name) {
			continue
		}
		for _, dependency := range graph[current.name] {
			path := append(append([]string(nil), current.path...), dependency)
			queue = append(queue, node{name: dependency, path: path})
		}
	}
	return nil, false
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
func Eligible(approved cv.ApprovedCandidate) bool { return approved.Valid() }`,
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
	defer func() { approvedCandidateBoundaries = oldBoundaries }()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy":   "pure bool/error strategy boundary fixture",
		"internal/testengine": "a general boundary permission must not authorize execution",
	}
	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing reverse primitive fixture: %v", err)
	}
	want := "internal/testengine reaches an authority package: internal/testengine -> internal/execgw"
	if !containsFinding(findings, want) {
		t.Fatalf("reverse primitive laundering findings = %v, want %q", findings, want)
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

func TestApprovedCandidateAuthorityReachCannotBeAllowedByBoundaryReason(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/strategy/strategy.go": `package strategy
import cv "` + module + `/internal/candidate"
func Eligible(approved cv.ApprovedCandidate) bool { return approved.Valid() }`,
		"internal/testengine/engine.go": `package testengine
import (
	"` + module + `/internal/strategy"
	"` + module + `/internal/journal"
	"` + module + `/internal/execgw"
)
func Run() { _ = strategy.Eligible; _, _ = journal.Open, execgw.Submit }`,
	})
	oldBoundaries := approvedCandidateBoundaries
	defer func() { approvedCandidateBoundaries = oldBoundaries }()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy":   "pure value-only strategy fixture",
		"internal/testengine": "journal-only reason must not mask a later execgw root",
	}

	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing reason-only bridge fixture: %v", err)
	}
	if !containsFinding(findings, "internal/testengine reaches an authority package:") {
		t.Fatalf("reason-only bridge masked authority expansion: %v", findings)
	}
}

func TestApprovedCandidatePureBoundaryRejectsInjectedAuthority(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/strategy/strategy.go": `package strategy
import (
	"net/http"
	cv "` + module + `/internal/candidate"
)
var packageCallback func()
type Submitter interface { Submit() error }
type Evaluator struct {
	Callback func()
	Client *http.Client
}
func EvaluateAndSubmit(approved cv.ApprovedCandidate, submitter Submitter, callback func()) bool {
	type LocalSubmitter interface { Submit() error }
	var localCallback func()
	local := func() {}
	_, _, _ = LocalSubmitter(nil), localCallback, local
	callback()
	_ = submitter.Submit()
	return approved.Valid()
}`,
	})
	oldBoundaries := approvedCandidateBoundaries
	defer func() { approvedCandidateBoundaries = oldBoundaries }()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy": "fixture exercises the pure evaluator contract",
	}

	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing injected authority fixture: %v", err)
	}
	for _, want := range []string{
		"imports net/http",
		"variable packageCallback",
		"type Submitter",
		"type LocalSubmitter",
		"type Evaluator: field Callback",
		"parameter callback",
		"variable localCallback",
		"forbids function literal",
		"forbids free or injected function call",
	} {
		if !findingContains(findings, want) {
			t.Errorf("pure-boundary findings = %v, want substring %q", findings, want)
		}
	}
}

func TestApprovedCandidatePureBoundaryAllowsValueOnlyEvaluator(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/strategy/strategy.go": `package strategy
import cv "` + module + `/internal/candidate"
type Decision struct { CandidateLifeID string; Eligible bool }
func Evaluate(approved cv.ApprovedCandidate) Decision {
	return Decision{CandidateLifeID: approved.CandidateLifeID(), Eligible: approved.Valid()}
}`,
	})
	oldBoundaries := approvedCandidateBoundaries
	defer func() { approvedCandidateBoundaries = oldBoundaries }()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy": "value-only input to immutable result fixture",
	}
	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing value-only fixture: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("value-only evaluator rejected: %v", findings)
	}
}

func TestApprovedCandidateBoundaryParserAndDotImportControls(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/dotreader/reader.go": `package dotreader
import . "` + module + `/internal/candidate"
func Evaluate(approved ApprovedCandidate) bool { return approved.Valid() }`,
	})
	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing dot-import fixture: %v", err)
	}
	if !containsFinding(findings, "internal/dotreader reads ApprovedCandidate") {
		t.Fatalf("dot-import direct reader was not detected: %v", findings)
	}

	badRoot, badRels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/broken/broken.go": "package broken\nfunc {",
	})
	if _, err := auditApprovedCandidateBoundaries(badRoot, module, badRels); err == nil {
		t.Fatal("malformed source did not fail the boundary audit closed")
	}
}

func writeApprovedBoundaryFixture(t *testing.T, files map[string]string) (string, []string) {
	t.Helper()
	root := t.TempDir()
	rels := make([]string, 0, len(files))
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
	return root, rels
}

func findingContains(findings []string, want string) bool {
	for _, finding := range findings {
		if strings.Contains(finding, want) {
			return true
		}
	}
	return false
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
