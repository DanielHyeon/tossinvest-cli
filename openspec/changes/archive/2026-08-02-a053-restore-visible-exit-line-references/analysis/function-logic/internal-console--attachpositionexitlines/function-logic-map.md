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
| Runtime candidate policy | effective-known stop pct only; desired designation may prove candidacy but never effectiveness | engine runtime + desired include membership | unavailable/invalid runtime produces typed unknown with no percentage or price |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `range` line 103 | each position row | local row projection | continues | empty/mixed rows |
| B2 | `if` line 108 | shared pending-designation predicate is true: broker holding is unmanaged, non-released, known, desired-designated, non-excluded, and management projection is absent | normalizes reference status to unknown, never pending | continues | unavailable-commander US plus managed/released/designated collision tests |
| B3 | `if` line 114 | no exit state | projects effective-known candidate or typed unknown only | continues | KR/US candidate/runtime tests |
| B4 | `if` line 123 | lifecycle generation differs | clears raw/actionable values and emits mismatch reference | continues | generation mismatch test |
| B5 | `if` line 137 | released lifecycle | suppresses actionable line; only validated raw stays historical | continues | released tests |
| B6 | `if` line 148 | released evidence is allowlisted | stores non-effective historical evidence | continues | released valid/corrupt tests |
| B7 | `if` line 160 | canonical snapshot exists | passes canonical snapshot to fail-closed adapter | continues | fresh/stale tests |
| B8 | `if` line 175 | required lifecycle is unverified | clears canonical/raw price and identity | continues | canonical lifecycle-unknown test |
| B9 | `if` line 182 | projector allowlists legacy raw | stores non-effective evidence | continues | legacy/corrupt tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ExitSnapshotView.WithFreshness` | classify canonical evidence | pure clock comparison | journal tests |
| `operatorview.BuildExitLine` | canonical actionable projection | stale/unknown fail closed | operatorview tests |
| `operatorview.BuildExitLineReference` | non-effective raw/plan projection | mismatch/unknown returns no prices | a053 tests |
| `positionRow.PendingDesignation` | keep desired-only fallback identical in status, reason, and reference projections | pure predicate; managed/released/unknown/excluded rows are false | post-deploy collision regression + released truth table |
| `hasStoredExitEvidence` | detect exact persisted raw strings | no recomputation | current AST |

## State mutations and fallbacks

- Mutates only `positionRow.ExitLine`, `StoredExit`, and the new read-only reference field.
- No price arithmetic occurs in console code.
- Desired inclusion can establish that a row is awaiting management, but an unavailable commander can only produce `RUNTIME_UNKNOWN`; desired/default stop percentages are never promoted.
- Known generation mismatch, required-but-unverified lifecycle, and corrupt snapshot state suppress raw price output.

## Safety conclusion

- Safe edit boundary: display-only projection using exact journal/runtime evidence.
- High-risk impact: no direct order side effect; price evidence must remain non-actionable unless canonical and fresh.
