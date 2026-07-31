package candidate

import (
	"strings"
	"testing"
)

func auditPureBoundaryFixture(t *testing.T, declarations string) []string {
	t.Helper()
	module := modulePath(t, moduleRoot(t))
	source := `package strategy
import cv "` + module + `/internal/candidate"
` + declarations
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/strategy/strategy.go": source,
	})
	old := approvedCandidateBoundaries
	defer func() { approvedCandidateBoundaries = old }()
	approvedCandidateBoundaries = map[string]string{
		"internal/strategy": "type-checked pure boundary fixture",
	}
	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing pure boundary fixture: %v", err)
	}
	return findings
}

func requirePureFinding(t *testing.T, findings []string, want string) {
	t.Helper()
	for _, finding := range findings {
		if strings.Contains(finding, want) {
			return
		}
	}
	t.Fatalf("findings = %v, want substring %q", findings, want)
}

func TestApprovedCandidatePureBoundaryRejectsAliasedAndNamedCapabilities(t *testing.T) {
	tests := []struct {
		name, declaration string
	}{
		{"pointer alias", `type Bad = *int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"named pointer", `type Bad *int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"map alias", `type Bad = map[string]int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"named map", `type Bad map[string]int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"slice alias", `type Bad = []int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"named slice", `type Bad []int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"channel alias", `type Bad = chan int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
		{"named channel", `type Bad chan int
func Evaluate(a cv.ApprovedCandidate, bad Bad) bool { return a.Valid() }`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requirePureFinding(t, auditPureBoundaryFixture(t, test.declaration), "type contract rejects")
		})
	}
}

func TestApprovedCandidatePureBoundaryRejectsNestedAndGenericCapabilities(t *testing.T) {
	tests := []struct {
		name, declaration string
	}{
		{"any and named callback assertion", `type AnyAlias = any
type Callback func()
func Evaluate(a cv.ApprovedCandidate, value AnyAlias) bool {
	value.(Callback)()
	return a.Valid()
}`},
		{"embedded named pointer and interface", `type Mutable struct { N int }
type Ref = *Mutable
type Runner interface { Run() }
type Envelope struct { Ref; Runner }
func Evaluate(a cv.ApprovedCandidate, value Envelope) bool { return a.Valid() }`},
		{"method value and free helper", `func helper(value bool) bool { return value }
func Evaluate(a cv.ApprovedCandidate) bool {
	valid := a.Valid
	return helper(valid())
}`},
		{"named capability result", `type Callback func()
func Evaluate(a cv.ApprovedCandidate) Callback { return func() {} }`},
		{"error result", `func Evaluate(a cv.ApprovedCandidate) error { return nil }`},
		{"candidate selector interface", `func Evaluate(a cv.ApprovedCandidate, source cv.Source) bool { return a.Valid() }`},
		{"capability generic", `type Box[T any] struct { Value T }
func Evaluate(a cv.ApprovedCandidate, box Box[func()]) bool { return a.Valid() }`},
		{"fixed array hides pointer", `type Ref = *int
type Envelope struct { Values [2]Ref }
func Evaluate(a cv.ApprovedCandidate, value Envelope) bool { return a.Valid() }`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requirePureFinding(t, auditPureBoundaryFixture(t, test.declaration), "type contract rejects")
		})
	}
}

func TestApprovedCandidatePureBoundaryRejectsScalarPackageStateAndPackageCall(t *testing.T) {
	findings := auditPureBoundaryFixture(t, `var approved bool
func Evaluate(a cv.ApprovedCandidate) bool {
	approved = a.Valid()
	_ = cv.OrderedVetoCodes()
	return approved
}`)
	requirePureFinding(t, findings, "package variable approved")
	requirePureFinding(t, findings, "free or injected function call")
}

func TestApprovedCandidatePureBoundaryRejectsMutationAndExecutionForms(t *testing.T) {
	findings := auditPureBoundaryFixture(t, `type Decision struct { Eligible bool }
type Worker struct{}
func (Worker) Run() {}
func init() {}
func helper() {}
func Evaluate(a cv.ApprovedCandidate, pointer *int, values []int, output chan int) Decision {
	*pointer = 1
	values[0] = 1
	output <- 1
	defer helper()
	go helper()
	return Decision{Eligible: a.Valid()}
}`)
	for _, want := range []string{
		"method declaration",
		"init function",
		"dereference assignment",
		"index assignment",
		"channel send",
		"defer statement",
		"go statement",
		"free or injected function call",
	} {
		requirePureFinding(t, findings, want)
	}
}

func TestApprovedCandidatePureBoundaryRejectsSelectorAssignmentAndExternalImport(t *testing.T) {
	module := modulePath(t, moduleRoot(t))
	root, rels := writeApprovedBoundaryFixture(t, map[string]string{
		"internal/strategy/strategy.go": `package strategy
import (
	"unsafe"
	cv "` + module + `/internal/candidate"
)
type Decision struct { Eligible bool }
func Evaluate(a cv.ApprovedCandidate, value Decision) Decision {
	value.Eligible = uintptr(unsafe.Pointer(nil)) == 0
	return value
}`,
	})
	old := approvedCandidateBoundaries
	defer func() { approvedCandidateBoundaries = old }()
	approvedCandidateBoundaries = map[string]string{"internal/strategy": "external import fixture"}
	findings, err := auditApprovedCandidateBoundaries(root, module, rels)
	if err != nil {
		t.Fatalf("auditing external import fixture: %v", err)
	}
	requirePureFinding(t, findings, "imports unsafe")
	requirePureFinding(t, findings, "selector assignment")
}

func TestApprovedCandidatePureBoundaryAllowsScalarStructAndFixedArray(t *testing.T) {
	findings := auditPureBoundaryFixture(t, `type Decision struct {
	Candidate cv.ApprovedCandidate
	Codes [2]string
	Eligible bool
	Score int
}
func Evaluate(a cv.ApprovedCandidate, score int) Decision {
	localScore := score
	return Decision{
		Candidate: a,
		Codes: [2]string{a.CandidateLifeID(), ""},
		Eligible: a.Valid(),
		Score: max(localScore, 0),
	}
}`)
	if len(findings) != 0 {
		t.Fatalf("scalar/value-only evaluator rejected: %v", findings)
	}
}
