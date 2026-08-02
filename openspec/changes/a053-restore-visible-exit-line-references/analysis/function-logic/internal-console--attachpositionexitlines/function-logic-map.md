# Function Logic Map: `attachPositionExitLines`

- Source: `internal/console/portfolio_pages.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Rows | broker-only, managed, released; KR/US | positions read model | no evidence produces no price |
| Snapshot | fresh/stale/absent | persisted effective snapshot | stale/unknown actionable values remain dashes |
| Raw exit state | exact decimal strings, lifecycle generation >=1 | journal | never copied into actionable `ExitLine` |
| Runtime candidate policy | effective-known stop pct only | engine runtime | unknown/invalid settings produce no percentage |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` line 103 | each position row | local row projection | continues | empty/mixed rows |
| B2 | `if` line 107 | no exit state | projects effective-known candidate or typed unknown only | continues | KR/US candidate/runtime tests |
| B3 | `if` line 116 | lifecycle generation differs | clears raw/actionable values and emits mismatch reference | continues | generation mismatch test |
| B4 | `if` line 130 | released lifecycle | suppresses actionable line; only validated raw stays historical | continues | released tests |
| B5 | `if` line 141 | released evidence is allowlisted | stores non-effective historical evidence | continues | released valid/corrupt tests |
| B6 | `if` line 153 | canonical snapshot exists | passes canonical snapshot to fail-closed adapter | continues | fresh/stale tests |
| B7 | `if` line 168 | required lifecycle is unverified | clears canonical/raw price and identity | continues | canonical lifecycle-unknown test |
| B8 | `if` line 175 | projector allowlists legacy raw | stores non-effective evidence | continues | legacy/corrupt tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ExitSnapshotView.WithFreshness` | classify canonical evidence | pure clock comparison | journal tests |
| `operatorview.BuildExitLine` | canonical actionable projection | stale/unknown fail closed | operatorview tests |
| `operatorview.BuildExitLineReference` | non-effective raw/plan projection | mismatch/unknown returns no prices | a053 tests |
| `hasStoredExitEvidence` | detect exact persisted raw strings | no recomputation | current AST |

## State mutations and fallbacks

- Mutates only `positionRow.ExitLine`, `StoredExit`, and the new read-only reference field.
- No price arithmetic occurs in console code.
- Known generation mismatch, required-but-unverified lifecycle, and corrupt snapshot state suppress raw price output.

## Safety conclusion

- Safe edit boundary: display-only projection using exact journal/runtime evidence.
- High-risk impact: no direct order side effect; price evidence must remain non-actionable unless canonical and fresh.
