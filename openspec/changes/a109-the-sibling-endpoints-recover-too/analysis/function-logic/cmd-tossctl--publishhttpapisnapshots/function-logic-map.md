# Function Logic Map: `publishHTTPAPISnapshots`

- Source: `cmd/tossctl/httpapi.go` (700-729)
- AST evidence: `ast.json` — AST 분기 5 · return 2 · defer 1
  (source_sha256 는 ast.json 이 정본, a109 §2-fix F3 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 §2-fix 편집 대상**: 인자 하나(`strategyRuntime`)와 tick 의 첫 줄 하나. 기존 분기
  구조(B1–B5)는 **바꾸지 않는다** — 바꾼 것은 tick 이 하는 일의 순서다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 데몬 수명 | `runHTTPAPI` 의 publisherCtx | 종료 시 루프 return (B2) |
| `stream` | non-nil | `httpapi.NewStream` | `Publish` 실패는 루프 종료 (B5) |
| `snapshots` | non-nil | `newHTTPAPISnapshotCache` | `Refresh` 실패는 **이번 tick 건너뛰기** (B3) |
| `strategyRuntime` | nil 가능 | `strategyRuntimeReaderFor` | nil 이면 부재 판정이 true 를 주고 아무 일도 안 일어난다 |
| `interval` | > 0 | 호출자 (운영 30s) | 0 이하면 `time.NewTicker` 가 패닉 — 호출자 계약 |

**불변식 1**: 같은 스냅샷을 두 번 싣지 않는다(B4 의 sha256 비교). 변화 없는 이벤트를
SSE 구독자에게 보내면 그것이 곧 화면 깜빡임이다.

**불변식 2 (a109 §2-fix F3)**: **재부착 깨우기는 집계 성공과 독립이다.** 이 루프는 요청이
없어도 도는 유일한 것이고, 그래서 전략 reader 재부착 시도의 **상시 구동원**이다. 깨우기를
`Refresh` 성공 뒤에 두면 전략과 무관한 조회 하나(옵티마이저 DB 등)의 고장이 재부착을
영원히 잠근다 — `httpAPIReader.Snapshot` 은 앞의 일곱 조회 중 하나만 실패해도 전략 블록
앞에서 끝나고, 여기 B3 이 `continue` 한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:706) | `for` — 루프 본체 | 없음 | — | 아래 전부 |
| B2 (:707) | `select` — ctx 종료 vs tick | 없음 | ctx 종료면 return (:709) | `TestTheReattachWakeSurvivesABrokenAggregate` (cancel 후 종료를 기다린다) |
| **B3 (:716)** | 집계 `Refresh` 실패 | **없음 — 깨우기는 이미 끝났다** | `continue` | `TestTheReattachWakeSurvivesABrokenAggregate` |
| B4 (:720) | 이전과 같은 digest | 없음 | `continue` | 기존 stream 테스트 |
| B5 (:723) | `stream.Publish` 실패 | 없음 | return (:724) — 루프 종료 | 기존 |

`defer ticker.Stop()` (:703) 하나. 새 ticker 는 **두지 않았다**: 재부착 시도에는 자기
rate limit(기본 30s)이 이미 있고, 주기가 둘이면 어느 쪽이 시도를 깨웠는지 알 수 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httpapi.StrategyRuntimeAbsent` | **값을 안 쓴다** — 재부착 시도를 깨우려고 부른다 | 막히지 않는다(깨우기는 goroutine 하나를 띄우고 즉시 돌아온다) | AST · tick 첫 줄 |
| `snapshots.Refresh` | 집계 재생성 | 실패는 이번 tick 건너뛰기 (B3) | AST · :716 |
| `sha256.Sum256` | 변화 판정 | 순수 함수 | AST · :718 |
| `stream.Publish` | SSE 발행 | 실패는 루프 종료 (B5) | AST · :723 |

## State mutations and fallbacks

- 루프 지역 상태 둘(`previous`·`known`)뿐. 프로세스 상태는 바꾸지 않는다.
- 재부착 wrapper 의 상태 변경은 **wrapper 안에서** 일어난다 — 이 함수는 질문만 한다.

## Safety conclusion

- Safe edit boundary: tick 본문의 **첫 줄**과 인자 하나. 분기 구조는 그대로다.
- High-risk impact: **no** — 조회 데몬의 발행 루프이고 주문 경로가 아니다.
- 금지: 깨우기를 `Refresh` 성공 **안으로** 되돌리는 것(F3 이 지운 결합), 새 ticker 추가,
  부재 판정의 반환값으로 발행을 가르는 것(부재는 발행과 무관하다).
