package execgw

import "github.com/JungHoonGhae/tossinvest-cli/internal/risk"

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
