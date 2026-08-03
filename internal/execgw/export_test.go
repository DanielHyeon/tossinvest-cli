package execgw

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// export_test.go is the one seam the Guardian's tests need and the built binary
// must not have.
//
// The property being tested is negative — "the chain is evaluated exactly once
// per issuance, and the re-collection loop does not run it again" (design D1) —
// and the only honest way to test a negative about a call is to count the calls.
// Counting needs a substitute evaluator; a substitute evaluator on the exported
// API would be a knob that lets a caller replace the risk chain with one that
// allows everything. So the seam lives here, in a _test.go file, which means it
// does not exist outside `go test`.

// SetChainForTest replaces the chain evaluator and returns the restore
// function. TESTS ONLY, and not for parallel tests: the variable is
// package-scoped.
func SetChainForTest(fn func(risk.Input) risk.Decision) func() {
	previous := evaluateChain
	evaluateChain = fn
	return func() { evaluateChain = previous }
}

// SetFXAuthorityForTest supplies the q_final package's private test proof. The
// compiled product has neither this method nor any raw-FXEvidence authority
// setter; production requests must carry opaque officialfx.Evidence.
func (r *QFinalEntryIssuance) SetFXAuthorityForTest(evidence riskbucket.FXEvidence) {
	r.testFXAuthority = evidence
	r.testFXAuthoritySet = true
}

func (r *QFinalEntryIssuance) ClearFXAuthorityForTest() {
	r.testFXAuthority = riskbucket.FXEvidence{}
	r.testFXAuthoritySet = false
}

// This package's test binary supplies an opaque, non-scalar harness for legacy
// gateway tests whose subject is not readiness.
//
// Nearly every suite here drives a buy because that mutation carries the most
// machinery. The readiness-specific tests opt into the real adapter instead.
//
// It cannot leak: export_test.go is compiled only into `go test` for this
// package.
func init() {
	defaultProtectionCheckForTest = func(_ context.Context, market string, previous protectionCheckpoint) (protectionCheckpoint, *RejectedError) {
		return protectionCheckpoint{testIdentity: "legacy-test:" + market}, nil
	}
}
