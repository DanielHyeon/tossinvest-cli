package position

// provenance.go classifies a difference between the account and the projection
// (change add-core-domain task 6.2; position-ledger "조정 이벤트": 근거·분류
// (외부/수동/미상)·기대 이전 값).
//
// The three kinds are not three guesses at the same fact. They are three
// different facts, and the audit trail has to keep them apart:
//
//	EXTERNAL  the account holds something no local instance ever opened. The
//	          position is folded in so the ledger stops lying about the account,
//	          but it has no entry decision and therefore no entry stop, which is
//	          why it is not an exit-policy target (design D4).
//	MANUAL    a person did it and said so. Only an operator-declared adjustment
//	          earns this, because "a human probably did it" is a guess and this
//	          column is evidence.
//	UNKNOWN   a local instance exists and the difference cannot be attributed to
//	          anything recorded. Not a placeholder for "we did not look": it is
//	          the recorded verdict that attribution failed
//	          (internal/journal/core_domain.go).

// Kind is an adjustment's provenance classification. The values are the ones
// position_adjustments.kind's CHECK constraint admits.
type Kind string

const (
	KindExternal Kind = "EXTERNAL"
	KindManual   Kind = "MANUAL"
	KindUnknown  Kind = "UNKNOWN"
)

// ValidKind reports whether k is one of the three.
func ValidKind(k Kind) bool {
	switch k {
	case KindExternal, KindManual, KindUnknown:
		return true
	default:
		return false
	}
}

// ProvenanceInputs is what is known about a difference at the moment it is
// classified.
type ProvenanceInputs struct {
	// LocalInstance reports that the projection already has an instance of this
	// symbol on this account — open or closed, because a closed one still means
	// the engine has traded here.
	LocalInstance bool
	// OperatorDeclared reports that a person raised this adjustment and recorded
	// why. It is the only path to MANUAL.
	OperatorDeclared bool
}

// Classify names where a difference came from.
//
// The order is the order of the evidence. An operator's declaration is a
// statement of fact and outranks everything inferred. Absent that, a symbol the
// engine has never traded is external by construction — there is no local event
// it could have come from. What is left is a difference on a symbol the engine
// *has* traded, which is precisely the case nothing in the journal explains, and
// naming it UNKNOWN is the honest answer rather than the unhelpful one.
func Classify(in ProvenanceInputs) Kind {
	switch {
	case in.OperatorDeclared:
		return KindManual
	case !in.LocalInstance:
		return KindExternal
	default:
		return KindUnknown
	}
}

// ExitEligible reports whether a position acquired this way may be managed by
// the exit policy.
//
// Only the engine's own positions are. An externally or manually acquired
// position has no entry decision and therefore no entry stop, and D5's whole
// first correction is that the t0 baseline *is* the entry stop: a position with
// no baseline is a position the ratchet cannot protect, and pretending otherwise
// would put a made-up stop on somebody else's shares.
func ExitEligible(entryDecisionID string) bool { return orEmpty(entryDecisionID) != "" }
