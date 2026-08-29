# Function Logic Map: `RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque precheck | non-zero q_final, sealed collection, transaction and projected lineage | preceding Guardian precheck | typed refusal; zero write |
| issue-time account FX | exact same frozen authority and fresh | official/test-sealed authority | currency refusal; zero write |
| recollected exposure snapshot | current account-wide snapshot | `Entry.Collect` | bounded journal recollection or error |
| optional weekly binding | opaque copy already checked against the sealed lane/market/campaign | preceding Guardian precheck | journal revalidates exact durable row in same transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | incomplete opaque precheck | none | risk-calculation refusal | zero-value precheck test |
| B2 | account FX expired/changed | none | currency refusal | paired FX tests |
| B3 | issue-time policy/FX invalid | none | currency refusal | paired FX tests |
| B4 | recollection or exposure calculation fails | no transaction committed | wrapped error | atomic failure tests |
| B5 | exact authority | one journal transaction | opaque receipt | paired six-lane tests |
| B6 | weekly binding changed after precheck or no longer matches durable reservation | no transaction committed | weekly reservation conflict | paired weekly atomic tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `qFinalAccountBaseFXAt` / `qFinalPolicyAt` | revalidate frozen issue-time FX | no fallback except tagged legacy KR test seam | CodeGraph + AST |
| `risk.EntryExposureValue` | calculate exact base reservation | fail closed | CodeGraph + AST |
| `RecordQFinalCampaignFirstLegWithRecollection` | atomic q_final + strategy + campaign write | bounded recollection; no standalone issuance | CodeGraph + AST |

## State mutations and fallbacks

- Mutates only through the single atomic journal method. For a weekly lane the same request includes the exact opaque reservation binding. It mints neither dispatch lease nor Gateway/broker authority.

## Safety conclusion

- Safe edit boundary: issue only an opaque precheck produced by the same Guardian path.
- High-risk impact: yes — atomic financial authority write; race and zero-write refusal tests required.
