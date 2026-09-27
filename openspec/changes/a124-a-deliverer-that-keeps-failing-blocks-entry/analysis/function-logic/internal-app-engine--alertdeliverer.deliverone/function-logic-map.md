# Function Logic Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go` (270-340)
- Revision: current — a124 구현 로트(2026-09-27) 편집 뒤 재추출; source_sha256 `5791a31af9d2407935ea71edbdbf9a14cbd2025fce5da9866b1d586f8e8da0cf`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력, `analysis/harness/flm_a124.py` 가 다시 만듦)
- Risk scan: `risk-pattern-report.md`
- Extractor counts: AST branches 11 · returns 4 · calls 20
- Exact AST return positions: 276:3, 297:3, 303:3, 337:3
- 편집 전 판(base `4798d399`)의 분석은 이 파일의 git 이력에 있고, 분기 대응은 아래 표다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `alert` | 나열 시점의 PENDING 행 — `Attempts` 는 판정에 쓰지 않음(D1) | `cycle` | 나열 뒤 정착되면 임차가 「이미 정산됨」 |
| `d.led().ClaimAlertByID` | 임차 3 종 결과 | 원장 | 오류 → 행별 기록 실패 계수(D8) |
| `d.Publisher` | nil 허용 — nil 도 실패 시도(D3) | 배선 | nil → 고정 원인 문구로 실패 기록 |

## Branches and early returns

| Branch | AST anchor | Source text at anchor |
|---|---|---|
| B1 | if at 272:2 | `if err != nil {` |
| B2 | switch at 278:2 | `switch claim.Disposition {` |
| B3 | case at 279:2 | `case journal.ClaimAcquired:` |
| B4 | case at 283:2 | `case journal.ClaimHeldElsewhere:` |
| B5 | if at 292:3 | `if d.reportHeld(alert.ID, claim.ExpiresAt) {` |
| B6 | case at 298:2 | `default:` |
| B7 | if at 305:2 | `if claim.Stole {` |
| B8 | if at 314:2 | `if d.Publisher == nil {` |
| B9 | else at 324:9 | `} else {` |
| B10 | if at 331:3 | `if perr != nil {` |
| B11 | if at 335:2 | `if perr != nil {` |

### 편집 전 → 편집 뒤 분기 대응 (difflib — 분기 줄 소스 텍스트 정렬)

| Base branch | Base anchor | Base source | Current branch |
|---|---|---|---|
| b1 | if at 272·2 | `}` | — (없어짐 · 하위 함수로 옮김) |
| b2 | switch at 278·2 | `// forgetLapsedHeld drops episodes whose lease has already run out.` | — (없어짐 · 하위 함수로 옮김) |
| b3 | case at 279·2 | `//` | — (없어짐 · 하위 함수로 옮김) |
| b4 | case at 283·2 | `func (d *alertDeliverer) forgetLapsedHeld(now time.Time) {` | — (없어짐 · 하위 함수로 옮김) |
| b5 | if at 292·3 | `if d.Interval <= 0 {` | — (없어짐 · 하위 함수로 옮김) |
| b6 | case at 298·2 | `func (d *alertDeliverer) batch() int {` | — (없어짐 · 하위 함수로 옮김) |
| b7 | if at 305·2 | `func (d *alertDeliverer) claimant() string {` | — (없어짐 · 하위 함수로 옮김) |
| b8 | if at 314·2 | `func (d *alertDeliverer) logf(event obs.EventType, err error, detail string, args ...any) {` | — (없어짐 · 하위 함수로 옮김) |
| b9 | else at 324·9 | `}` | — (없어짐 · 하위 함수로 옮김) |
| b10 | if at 331·3 | `` | — (없어짐 · 하위 함수로 옮김) |
| b11 | if at 335·2 | `` | — (없어짐 · 하위 함수로 옮김) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract |
|---|---|---|
| `d.led().ClaimAlertByID` | 이 행의 발송 권한 | 오류 → `countRecordFailure` |
| `d.forgetHeld` · `d.reportHeld` · `d.forgetRecordRun` | 보고 · 계수 기록 정리 | 메모리 |
| `d.Publisher.Publish` | 원격 전송 1 회 | 잠금 없이 — 재시도는 다음 사이클 |
| `d.recordFailedAttempt` | 실패 기록 → 반납 → 판정(D1 실패 기록 표) | 판정은 원칙 E |
| `d.recordDelivery` | 전달 정산 → 실패면 즉시 판정(D1 전달 정산 표) | 임차 유지 |
| `d.logf` | 기존 줄 — 토큰 없음 | 관측만 |

## State mutations and fallbacks

- 원장 쓰기는 전부 하위 함수에 있다. 이 함수 자체는 메모리 기록(`heldReported`, `recordRuns`)만 바꾼다.
- publisher 가 없을 때도 이제 실패 시도가 기록된다(`alertNoPublisherCause`) — 편집 전에는 로그와 반납뿐이었다(D3).
- 임차 결과가 「이미 정산됨」이면 그 행의 기록 실패 연속을 지운다(Y3).

## Safety conclusion

- Safe edit boundary: 판정은 하위 함수로 옮겼고 이 함수는 흐름만 가른다. 잠금을 쥐지 않는다.
- High-risk impact: yes — 전달 실패 → 진입 차단 · 모드 승격의 입구. 새 문은 없음(오늘 안 잠기던 자리가 잠긴다).
