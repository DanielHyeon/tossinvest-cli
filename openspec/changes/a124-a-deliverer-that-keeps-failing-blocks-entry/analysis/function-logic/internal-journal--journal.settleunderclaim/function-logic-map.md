# Function Logic Map: `Journal.settleUnderClaim`

- Source: `internal/journal/alert_claim.go` (314-355)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `b123231116301de8e6a2270a3061c91a75db20c34bc355c38184af7cabf55577`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 9 · returns 10 · calls 17
- Exact AST return positions: 318:3, 322:3, 328:3, 332:3, 339:4, 342:4, 344:3, 349:3, 352:3, 354:2
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `token` | 비어 있지 않음 | 임차 | 빈 토큰 → 오류 |
| `stmt` · `args` | id · PENDING · 토큰 CAS 로 끝나는 UPDATE | 세 호출자 | 0 행이면 이유를 다시 읽음 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 317:2 | `if strings.TrimSpace(token) == "" {` |
| B2 | if at 321:2 | `if err != nil {` |
| B3 | if at 327:2 | `if err != nil {` |
| B4 | if at 331:2 | `if err != nil {` |
| B5 | if at 334:2 | `if n == 1 {` |
| B6 | if at 338:3 | `if err != nil {` |
| B7 | if at 341:3 | `if err := tx.Commit(); err != nil {` |
| B8 | if at 348:2 | `if err != nil {` |
| B9 | if at 351:2 | `if err := tx.Commit(); err != nil {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 317·2 | `if err != nil {` | B2 |
| b2 | if at 321·2 | `` | — (없어짐 · 하위 함수로 옮김) |
| b3 | if at 327·2 | `if err != nil {` | B6 |
| b4 | if at 331·2 | `if err := tx.Commit(); err != nil {` | B7 |
| b5 | if at 334·2 | `return SettleResult{Outcome: SettleApplied}, nil` | — (없어짐 · 하위 함수로 옮김) |
| b6 | if at 338·3 | `if err != nil {` | B8 |
| b7 | if at 341·3 | `if err := tx.Commit(); err != nil {` | B9 |
| b8 | if at 348·2 | `// nothing, and turns "zero rows" into the reason the sender has to act on.` | — (없어짐 · 하위 함수로 옮김) |
| b9 | if at 351·2 | `if err != nil {` | — (없어짐 · 하위 함수로 옮김) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `tx.ExecContext` | 정산 한 문장 | 오류 → 롤백 |
| `readSettledAttemptsTx` | **a124**: 적용된 정산의 attempts 를 같은 트랜잭션에서 읽음(D1) | 실패 → 롤백 · 오류(정산 없던 일, Z4) |
| `tx.Commit` | 커밋 | 실패 → 오류 |
| `explainSettleTx` | 0 행의 이유(LeaseLost · AlreadySettled · NotFound) | — |

## State mutations and fallbacks

- 적용 경로에서 커밋 **전에** 같은 트랜잭션으로 attempts 를 읽어 `SettleResult.Attempts` 에 담는다. 연결이 하나라 `j.db` 로 읽으면 자기 트랜잭션을 기다림.
- 스키마 무변경. 세 호출자의 SQL · CAS 불변. 적용되지 않은 결과와 오류에서 Attempts 는 0.

## Safety conclusion

- Safe edit boundary: 가산 필드 + 트랜잭션 안 SELECT 한 줄. 실패하면 정산이 없던 일이 된다(원자성 시험).
- High-risk impact: yes — 원장(alert outbox) 정산 경로.
