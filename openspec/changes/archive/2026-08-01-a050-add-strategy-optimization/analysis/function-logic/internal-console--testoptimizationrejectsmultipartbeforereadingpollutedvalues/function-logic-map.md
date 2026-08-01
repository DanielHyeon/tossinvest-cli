# Function Logic Map: `TestOptimizationRejectsMultipartBeforeReadingPollutedValues`

- Source: `internal/console/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| multipart request | otherwise plausible optimization fields plus invented field and valid session/CSRF | adversarial fixture | 400 before preview/apply commander calls |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | write every multipart fixture field; writer error | test buffer only | fatal on setup error | this test |
| B3 | multipart writer close fails | test buffer only | fatal | this test |
| B4 | request construction fails | none | fatal | this test |
| B5 | client request fails | HTTP test only | fatal | this test |
| B6 | status is not 400 or commander was called | test failure only | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| multipart writer | constructs parser-confusion payload | test-only | fixture |
| console HTTP route | exercises real session/mutation wrapper chain | one request/no retry | status and zero-call assertions |

## State mutations and fallbacks

- Only the fake commander counters could observe a downstream call; both must remain zero. No production or trading state is available.

## Safety conclusion

- Safe edit boundary: adversarial parser/content-type regression test.
- High-risk impact: no direct side effect; validates a high-risk mutation gate.
