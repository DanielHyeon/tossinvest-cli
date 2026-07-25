package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/trading"

// Test-only accessors for the sealed mutators (task 2.5).
//
// The fields are unexported so no production caller can bypass the
// ExecutionGateway. The WTS-isolation tests still have to drive every mutation
// verb directly — proving the engine's conditional-order path never reaches the
// web session is the whole point of that suite — so the seam exists here, in a
// _test.go file, which means it does not exist in the built binary.

// BrokerForTest returns the engine's order mutator. TESTS ONLY.
func (c *Context) BrokerForTest() trading.Broker { return c.broker }

// ConditionalForTest returns the engine's conditional-order mutator. TESTS ONLY.
func (c *Context) ConditionalForTest() ConditionalMutator { return c.conditional }
