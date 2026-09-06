# Function Logic Map: `SourcePolicy.validate`

- Source: `internal/strategyevidence/source.go`
- AST evidence: `ast.json` (extracted at the re-baselined comparison base)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.Authority` | one of the four known authorities | `authorityValid` | `ErrSourceDisabled` |
| `p.ContractVerified` | true for a minted official policy | `MintSourcePolicy` | `ErrSourceUnavailable` for KRX, `ErrSourceDisabled` otherwise |
| required strings | non-blank version, endpoint, endpoint version, method, schema, access contract, request identity | `MintSourcePolicy` | `ErrSourceDisabled` |
| numeric bounds | positive, request deadline <= operation deadline, and within the frozen official contract caps | `testdata/official_contracts.json` and the two minted policies | `ErrSourceDisabled` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the authority is not one of the four known ones | none | `ErrSourceDisabled` | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B2 | the contract is not verified | none | see B3 | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B3 | an unverified contract belongs to KRX | none | `ErrSourceUnavailable` (KRX) / `ErrSourceDisabled` | `TestKRXStaysUnavailableWhileItsContractIsNotFrozen` |
| B4 | walk the seven strings that must be present | none | falls through when all are set | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B5 | one of those strings is blank | none | `ErrSourceDisabled` | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B6 | an access-contract, positivity or deadline-ordering bound fails | none | `ErrSourceDisabled` | `TestEveryPolicyFieldHasAZeroCallRefusal` |
| B7 | a page size or an absolute contract cap is out of range | none | `ErrSourceDisabled` | `TestEveryPolicyFieldHasAZeroCallRefusal` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `authorityValid` | reject an unknown source authority | pure | AST |
| `strings.TrimSpace` | treat whitespace-only policy strings as absent | pure | AST |

## State mutations and fallbacks

- Pure predicate: no state is written and no request is made.
- There is no clamping fallback — an out-of-range policy is refused, never trimmed to the cap.

## Safety conclusion

- High-risk: this predicate is what stands between a mis-specified policy and an outbound request to an official regulator endpoint.
- The completion pass added B7. Before it, the numeric fields were only checked for `> 0` and `PageSize` was not checked at all, so a policy could declare a 1 TiB response limit or 100000 calls per window and still fetch (issues.md I17).
- Fail-closed scope, stated: B7 rejects only configurations outside the frozen official contracts. Both policies this module can mint (SEC and OpenDART) sit inside every cap, and `TestMintedPoliciesSurviveTheirOwnBounds` measures that, so B7 rejects no normal input.
