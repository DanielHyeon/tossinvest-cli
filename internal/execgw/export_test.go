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

// SetProtectionReadyForTest satisfies interlock clause 6 on the Options being
// built. TESTS ONLY.
//
// The clause is unmeetable in the built binary — ProfileProtection is a constant
// saying this build leaves no protective order at the broker, and the change that
// wires protective execution flips it. Without a way past it every test about an
// *accepted* raising mutation (idempotency, IN_DOUBT resolution, reservations,
// the round trip) would become a test about this one refusal, and those paths
// would go untested until that change lands.
//
// This is the convenience form for a single Options value. Tests in other
// packages set Options.ProtectionOverrideForTest directly, because an
// export_test.go declaration is visible only inside its own package — see
// WiredProtectionForTest.
func (o *Options) SetProtectionReadyForTest() {
	ready := ProtectionWired
	o.ProtectionOverrideForTest = &ready
}

// This package's test binary judges by WIRED unless a gateway says otherwise.
//
// Nearly every suite here drives a buy, because a buy is the mutation with the
// most machinery behind it, and none of them is about whether this build can
// leave a stop at the broker. Flipping the default once is the alternative to
// nineteen construction sites each carrying a field that says nothing about what
// they test. The two suites that *are* about clause 6 ask for the build's own
// answer with UnwiredProtectionForTest.
//
// It cannot leak: export_test.go is compiled only into `go test` for this
// package.
func init() { defaultProtection = ProtectionWired }
