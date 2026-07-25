package trading

import "github.com/JungHoonGhae/tossinvest-cli/internal/domain"

// MutationResult is an alias for domain.MutationResult, which is where the type
// now lives (see internal/domain/mutation.go for why).
//
// An alias, not a named type: every existing caller — internal/client,
// internal/official, internal/hybrid, internal/ops, internal/output and
// cmd/tossctl — refers to it as trading.MutationResult, and the engine packages
// added later refer to the domain type. Both must be the identical type or the
// two halves of the codebase could not exchange results without conversion.
type MutationResult = domain.MutationResult
