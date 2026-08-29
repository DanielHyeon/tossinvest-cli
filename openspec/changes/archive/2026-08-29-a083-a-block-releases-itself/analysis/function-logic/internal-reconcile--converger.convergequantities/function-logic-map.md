# Function Logic Map: `Converger.ConvergeQuantities`

- Source: `internal/reconcile/converge.go` (lines 136–254, pre-edit)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 원장 조정 발행. 면제 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff.Quantities` | 심볼별 로컬/브로커 수량 | `Comparer.Compare` | 비면 즉시 반환 (no-op) |
| `diff.AsOf` | RFC3339. **비면 거부** | `Compare` (`snap.AsOf` 포맷) | 149행에서 오류 반환 — 조정이 체결과 순서를 다툴 수 없다 |
| `c.Journal` | 필수 | 주입 | nil이면 오류 |
| `c.Credit` | 선택 (`*Tracker`) | 주입 | nil이면 수렴만 하고 차단은 운영자 전용이 된다 |
| `account` | `c.AccountRef` 또는 `diff.AccountRef` | 주입/비교 | 둘 다 비면 오류 |
| `instances` | 계좌의 포지션 투영 | `Journal.Positions` | 읽기 실패 시 pass 중단 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (138) | `len(diff.Quantities) == 0` | — | 빈 report, nil | 기존 |
| (141) | `c == nil \|\| c.Journal == nil` | — | 오류 | 기존 |
| (147) | `account == ""` | — | 오류 | 기존 |
| (150) | `asOf == ""` | — | 오류 — 조정을 체결과 정렬할 수 없다 | 기존 |
| (167) | `soleLiveMarket`가 단일 venue를 못 고름 | `report.Refused[symbol]` | continue (다른 심볼은 계속) | 기존 |
| (196) | `ErrAdjustmentStale` | — | pass 중단, 커밋된 것은 report에 남음 | 기존 |
| (201) | 기타 조정 오류 | — | pass 중단 | 기존 |
| (212) | `result.ClosedExitState` | `report.Closed++`, 관리 포지션 외부 청산 알림 | 알림 실패는 모아서 반환 | 기존 |
| **(238)** | `len(credited) > 0` | `report.Credited`, **`c.Credit.AdjustmentApplied(credited...)`** | — | **편집 site** — `asOf`를 함께 전달한다 (4.1) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.Journal.Positions` | 심볼의 live venue 판정 | 오류 시 pass 중단 | `converge.go:156` |
| `c.Journal.FillWatermark` | compare-and-append 증거 | 오류 시 pass 중단 | `converge.go:173` |
| `c.Journal.ApplyPositionAdjustment` | 투영을 계좌값으로 수렴 | `ExpectedPrevQuantity` + watermark 불변 재검증. 어긋나면 `ErrAdjustmentStale` | `converge.go:177` |
| **`c.Credit.AdjustmentApplied`** | 다음 재조회가 해제할 수 있게 한다 | 무오류, in-memory | `mismatch.go:314` — 프로덕션 구현은 `*Tracker` 하나 |
| `c.Alert.ManagedPositionClosedExternally` | 관리 포지션이 0이 됐음을 알린다 | 실패는 모아서 반환 | `converge.go:222` |

## State mutations and fallbacks

- 원장: 심볼당 조정 1건 (compare-and-append). 이 함수는 원장 외 상태를 갖지 않는다
- `c.Credit`의 in-memory credit 집합 — **편집 대상**
- fallback: 한 심볼의 refuse가 다른 심볼의 수렴을 막지 않는다. stale은 pass를 멈춘다

## Safety conclusion

- **Safe edit boundary**: 238행의 credit 호출 인자만. 조정 발행·compare-and-append·
  refuse·stale·알림 경로는 건드리지 않는다.
- **High-risk impact**: **yes** (원장 조정 경로에 있다). 다만 이 change가 바꾸는 것은
  credit 호출의 인자 하나이며, 어떤 조정이 발행되는지도 어떤 순서로 커밋되는지도
  바뀌지 않는다. `asOf`는 이미 같은 함수 안에서 `BrokerAsOf`로 원장에 기록되고 있다.
