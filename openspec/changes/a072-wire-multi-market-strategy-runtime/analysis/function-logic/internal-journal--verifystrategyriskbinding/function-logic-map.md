# Function Logic Map: `verifyStrategyRiskBinding`

- Source: `internal/journal/strategy_lineage.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Pre-edit contract: `../../pre-edit/strategyflow-first-leg-canonical-projection-spec.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `decision.RiskPreimage` / `RiskHash` | canonical registered `RiskIntent`; exact hash | journal decision issuance | return before transaction writes |
| `lineage.DecisionPayload` | legacy v1, historical explicit v2, or current explicit v3 | versioned projection verifier | unknown schema, unknown field, trailing bytes, non-canonical bytes fail closed |
| v1 payload binding | existing identity/digest and field comparison unchanged | `strategyengine.DecisionRecord` mirror | legacy mismatch error; no compatibility rewrite |
| v2 inner result | verified `strategyflow-accepted:v1`, sealed identities, fixed production router, one of six KR/US descriptors | `internal/strategyflow` projection verifier | tamper/cross-market/unsupported lane fail closed |
| v2 RiskIntent | exact account/market/symbol/entry/stop/target; positive integer q_final <= sealed candidate quantity; full Guardian policy independent from lane execution policy | canonical RiskIntent + sealed terms | drift or oversizing fails before journal write |
| persisted lineage columns | exact deterministic projection of v2 payload | journal projection contract | any column/payload substitution fails closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Risk preimage hash mismatch | none | `RiskIntent hash mismatch` | existing divergent risk test |
| B2 | canonical preimage parse fails | none | wrapped parse error | existing production issuance rejection |
| B3 | parsed preimage is not `RiskIntent` | none | type error | direct verifier type test |
| B4 | version discriminator reports invalid explicit schema | none | schema error | outer unknown/trailing/whitespace tests |
| B5 | payload carries an explicit version | none | enter explicit v2/v3 branch; no downgrade | compatibility and paired tests |
| B6 | explicit schema is neither v2 nor v3 | none | unsupported schema error | unknown-version test |
| B7 | v2 canonical RiskIntent account differs from decision account | none | account mismatch error | direct decision-account drift test |
| L1 | no schema: legacy strict decode/canonical/digest/identity/field checks | none | delegated legacy error or `nil` | existing untagged lineage suite |
| V1 | exact v2: outer/inner canonical bytes, digest or seals mismatch | none | delegated v2 projection error | inner/outer tamper tests |
| V2 | exact v2: RiskIntent scope/prices/policy mismatch, or q_final is non-integer/zero/above candidate | none | exact-binding error | paired drift, downsizing and oversizing tests |
| V3 | exact v2: persisted `StrategyDecisionLineage` column differs | none | exact projection error | column drift table |
| V4 | all selected-version checks pass | none | `nil` | v1 compatibility + KR/US six-lane success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `HashPreimage` | authenticate immutable RiskIntent bytes | deterministic, no I/O | AST B1 |
| `ParsePreimage` | strict registered preimage parsing | deterministic error; no retry/I/O | AST B2-B3 |
| legacy JSON decoder/marshal/hash | retain v1 byte canonicality and identity exactly | fail closed, no mutation | AST B4-B8 pre-edit |
| `verifyStrategyflowRiskBinding` | validate schema-aware v2/v3 outer and inner projection | deterministic, no mutation or authority | paired price-unit pre-edit contract |
| caller `RecordStrategyDecisionAndReserve` | validates before durable strategy decision/reservation writes | error rolls back / prevents rows | CodeGraph caller |
| caller `prepareQFinalCampaignFirstLeg` | validates before first-leg transaction begins | error returns zero first-leg authority rows | CodeGraph caller |

## State mutations and fallbacks

- This verifier owns no database handle and performs no state mutation, Gateway/broker call, lease mint,
  activation or toggle operation.
- Version dispatch is additive: absent schema retains legacy v1; exact v2 retains historical literal-minor
  semantics; exact v3 selects current major-decimal semantics. No explicit schema can downgrade to v1.
- No value is normalized into authority. Canonical bytes and sealed identities are checked, not repaired.

## Safety conclusion

- Safe edit boundary: common RiskIntent authentication followed by exact version dispatch; legacy body is
  extracted without semantic change, and all new behavior lives in a separate v2 verifier file.
- High-risk impact: yes. This function gates durable first-leg admission, so paired KR/US RED tests,
  tamper/rollback tests, legacy regression, race, vet, SDD and gate checks are mandatory.
