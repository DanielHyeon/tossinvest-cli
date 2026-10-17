# Function Logic Map: `ExitObserver.noteDelay`

- Source: `internal/app/engine/exitloop.go` (1569-1593)
- AST evidence: `ast.json` — branches 2, returns 2, calls 12, assignments 4
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.delayedSince[positionID]` | 프로세스 메모리 map | 이 함수 · `clearDelay` | 부재면 "시작"으로 해석 |
| `o.delayAlerted[positionID]` | 프로세스 메모리 map | 이 함수 · `clearDelay` | true면 재발화 없음 |
| `o.delayBound()` | `DefaultExitLiquidationDelayBound = 30s` (`:106`) | 설정 | — |
| `why` | 사람이 읽을 사유 문자열 | 호출자 2곳 | 알림 본문·`detail` 필드에 실린다 |

**호출자는 저장소 전체에서 둘뿐이다**: `record:1146`(working order 미정리),
`submit:1302`(`ReasonSymbolInFlight`). 브로커 거부 경로에는 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | 기존 테스트 |
|---|---|---|---|---|
| B1 `:1572` | `!running` (시계 미가동) | `delayedSince[id] = now` | `return` `:1574` | `:883` 1차 관측 |
| B2 `:1576` | `now-since < bound` 또는 `delayAlerted[id]` | 없음 | `return` `:1577` | `:900` (한계 내 0건) |
| — `:1579` | 한계 초과 & 미발화 | `delayAlerted[id]=true`, `alert(EventExitLiquidationDelayed)` | (끝) | `:906`,`:910` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.clk.Now` `:1570` | 경과 계산 | 오류 없음 | AST |
| `o.delayBound` `:1576` | 한계 | 오류 없음 | AST |
| `o.alert` `:1580` | critical 알림 | 내부에서 삼킴(`:1600-1607`) | AST |
| `o.label` `:1583` | 종목 표시명 | 오류 없음 | AST |

`o.alert` → `Notifier.Notify` → `deliver`는 **동기**다. publisher가 설정돼 있고 실패하면
`DefaultCriticalAttempts=3` × `RetryDelay` 만큼 exit 루프를 잡는다. `delayAlerted` latch가
episode당 1회로 묶는다.

## State mutations and fallbacks

- `delayedSince`·`delayAlerted` 두 map만 건드린다. 원장·브로커 접촉 없음
- **재시작 시 두 map이 소실된다** → 재시작 루프 중에는 한계에 영원히 도달하지 않는다
- fallback 없음

## Safety conclusion

- **Safe edit boundary**: 순수 계측. 판정·발의·제출 어디에도 닿지 않는다
- **High-risk impact**: no (이 함수 자체는). 다만 **호출 지점**의 추가·이동은 high-risk다
- 이 함수는 **고장나 있지 않다** — `:883` 테스트가 31초 후 critical 발화를 확인한다.
  거부 경로에서 안 뜨는 이유는 `record:1150`의 `clearDelay`가 매 주기 초기화하기 때문이다
