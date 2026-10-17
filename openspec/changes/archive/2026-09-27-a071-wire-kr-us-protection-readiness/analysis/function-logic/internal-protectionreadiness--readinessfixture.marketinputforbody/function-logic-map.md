# Function Logic Map: `readinessFixture.marketInputForBody`

- Source: `internal/protectionreadiness/attestation_test.go` (273-296)
- Revision: current — HEAD `648df8ef`; source_sha256 `2fa5888a4648fb4546bddf1970c4f8f22f7ab68b6953e7ed687a4c10fcc2fb48`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 4 · returns 1 · calls 13
- Exact AST return positions: 295:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| attestation body and private key | fixture-owned exact signed data | test fixture | fatal test on canonical/file/supervisor construction error |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 276:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B2 | if at 281:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B3 | if at 285:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |
| B4 | if at 292:2 | `if err != nil {` | _test.go — 커버 계측 밖(분기 진입 여부 미측정) |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | canonicalization or JSON/file/binding construction fails | test-only allocation | fatal test | fixture consumers |
| N2 | construction succeeds | creates sealed in-memory fixture | return exact scope/file/supervisor | attestation suites |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| canonical/sign/file/binding helpers | create cryptographically valid fixture | fatal on error | CodeGraph + AST |

## State mutations and fallbacks

- Test-only in-memory/file value construction; exact quantity bounds copied from signed body.

## Safety conclusion

- Safe edit boundary: fixture scope mirrors body exactly
- High-risk impact: no (test helper)
