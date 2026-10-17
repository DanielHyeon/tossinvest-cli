# Function Logic Map: `alertDeliverer.cycle`

- Source: `internal/app/engine/alertdelivery.go` (235-264)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `5791a31af9d2407935ea71edbdbf9a14cbd2025fce5da9866b1d586f8e8da0cf`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 4 · returns 3 · calls 12
- Exact AST return positions: 246:3, 259:4, 263:2
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | Run 의 수명 ctx | Run | 행 사이에서 취소를 봄 — 끝나면 조용히 반환 |
| `d.led().PendingAlertsForDelivery(ctx, batch, alertAttemptLimit)` | PENDING 행, 한도 아래 먼저 · 오래된 것 먼저 | 원장(`outbox.go` 새 메서드) | 오류 → 나열 실패 계수(D8) 뒤 오류 반환 |
| `d.batch()` | > 0 (기본 10) | `alertDeliveryBatch` | 잘림 판정의 기준 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 243:2 | `if err != nil {` |
| B2 | if at 249:2 | `if len(pending) < d.batch() {` |
| B3 | range at 254:2 | `for _, alert := range pending {` |
| B4 | if at 258:3 | `if ctx.Err() != nil {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 243·2 | `` | — (없어짐 · 하위 함수로 옮김) |
| b2 | if at 249·2 | `func (d *alertDeliverer) release(ctx context.Context, id int64, token string) {` | — (없어짐 · 하위 함수로 옮김) |
| b3 | range at 254·2 | `}` | — (없어짐 · 하위 함수로 옮김) |
| b4 | if at 258·3 | `// answered yes.` | — (없어짐 · 하위 함수로 옮김) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `d.forgetLapsedHeld` | 만료된 「남이 쥠」 보고 기록 정리 | 메모리 |
| `d.led().PendingAlertsForDelivery` | 굶주림 없는 선택(R2, AA4 — `PendingAlerts` 불변) | 오류 → `countListFailure` |
| `d.countListFailure` | 나열 실패의 연속(D8 AA2) | ctx 취소면 세지 않음 |
| `d.pruneRecordRuns` | 완전한 나열에 없는 행의 기록 실패 계수 삭제(Q5) | 잘린 나열에서는 부르지 않음 |
| `d.deliverOne` | 행 하나의 생애 | 행 실패는 사이클 실패가 아님 |

## State mutations and fallbacks

- 나열이 성공하면 나열 실패 연속을 지움(`listRun`, `listSeen`).
- 완전한 나열(`len < batch`)일 때만 기록 실패 계수를 거름 — 잘린 나열은 PENDING 이탈의 증거가 아님.
- 어떤 잠금도 쥐지 않음. 원장 연산 · 발행은 전부 `deliverOne` 과 그 아래에서.

## Safety conclusion

- Safe edit boundary: 선택 순서와 나열 실패 계수만 더함 — 행을 버리는 길은 없음(한도 행도 선택에 남음).
- High-risk impact: yes — 진입 차단(전달 실패 사유)의 판정 입력을 만든다. 차단 방향으로만 틀림.
