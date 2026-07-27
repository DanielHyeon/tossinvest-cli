package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

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

// OfficialClientForTest returns the concrete official client behind the
// read-only Official field (task 4.2). TESTS ONLY — production code that needs a
// broker write goes through internal/execgw.Gateway.
func (c *Context) OfficialClientForTest() *official.Client { return c.official }

// SetJournalProberForTest overrides the filesystem probe the journal's
// durability guard uses (task 7.2). TESTS ONLY.
//
// The engine profile opens the journal, so the ext4/xfs/btrfs allowlist is a
// startup condition — and TMPDIR is not necessarily on an allowlisted
// filesystem. internal/journal's own tests solve this the same way; the
// difference is only that the seam is reached through a method, because the
// field itself must not exist in the built binary's API.
func (o *Options) SetJournalProberForTest(p journal.FSProber) { o.journalFSProber = p }

// SetProtectionReadyForTest satisfies interlock clause 6 (task 5.2). TESTS ONLY.
//
// The clause is unmeetable in the built binary — profileProtection is a constant
// that says this build has no broker-resident protective execution, and the
// change that wires it flips that constant. Without a seam, every test about an
// *accepted* gate (the audit acceptance line, the verified structured log, the
// toggle trail) would become a test about a refusal, and the acceptance path
// would go untested until 2c. The seam lives here, in a _test.go file, so no
// production caller can claim a capability the build does not have.
func (o *Options) SetProtectionReadyForTest() {
	ready := ProtectionWired
	o.protectionOverride = &ready
}

// TradingServiceForTest returns the policy-enforcing order service the gateway
// wraps (task 7.4). TESTS ONLY.
//
// The field is unexported now, because a caller holding it can mutate with no
// GuardianDecision and no journal record. The WTS-isolation and pre-check suites
// have to drive the service directly — proving the engine's order verbs never
// reach the web session is the point of one and the pre-check is what the other
// is about — so the seam lives here, in a _test.go file, and therefore not in the
// built binary.
func (c *Context) TradingServiceForTest() *trading.Service { return c.tradingService }
