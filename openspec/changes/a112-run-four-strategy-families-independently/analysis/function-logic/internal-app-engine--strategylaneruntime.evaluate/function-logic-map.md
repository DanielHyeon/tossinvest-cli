# Function Logic Map: `strategyLaneRuntime.evaluate`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `strategyLaneRuntime.evaluate(params=5, results=1)`
- Source range: `199:1`–`258:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- join 은 시장 주기 안이다(분리된 레인 goroutine 금지 — Manager 조건 ①). 각 레인은 RunBounded 마감 시한으로 끊기므로 join ≤ 마감 시한 1 회.
- 버려진 사이클의 step goroutine 은 step 이 돌아올 때까지 산다(strategyworker.invokeBounded — 결과 채널 버퍼 1 이라 막히지 않고 끝남). 생산 step 은 순수 메모리 평가라 곧 끝나고, 버림은 비정상이라 레인이 즉시 잠기므로 레인당 최대 하나(조건 ② — review 「7.5」).
- 레인 goroutine 의 panic 은 삼키지 않는다 — join 뒤 다시 던져 invokeStrategyCycle 이 받는다(다른 goroutine 의 panic 은 그 경로가 못 잡는다).

## Branches and early returns

- Exact AST return nodes: `203:3, 209:3, 253:3, 257:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 202:2 | nil 런타임 |
| B2 | if | 208:2 | 복구 실패 → 오류 |
| B3 | range | 228:2 | **(편집)** 레인마다 goroutine 하나 · join(시장 주기 안) — 멈춘 레인이 이웃을 세우지 않음, 시장 지연 = 최댓값 |
| B4 | range | 230:3 | 레인 입력 찾기 |
| B5 | if | 231:4 | 자기 제안 소유 — 제안 하나 → 레인 하나 |
| B6 | range | 244:2 | **(새)** 레인 goroutine 들의 panic 값 순회 |
| B7 | if | 245:3 | **(새)** panic 이 있으면 시장 주기 goroutine 에서 다시 던짐(순차 때와 같은 회복 경로) |
| B8 | if | 252:2 | 잠금 기록 실패 → 오류 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `runtime.recoverMarketLanes` | 208:12 |
| `runtime.lanesFor` | 211:11 |
| `make` | 225:18 |
| `len` | 225:50 |
| `make` | 226:12 |
| `len` | 226:24 |
| `lane.Owns` | 231:7 |
| `join.Add` | 236:3 |
| `(unnamed)` | 237:6 |
| `join.Done` | 238:10 |
| `(unnamed)` | 239:10 |
| `recover` | 239:35 |
| `runtime.runLane` | 240:26 |
| `join.Wait` | 243:2 |
| `panic` | 246:4 |
| `runtime.record` | 249:2 |
| `runtime.persistMarketLatches` | 252:12 |
| `runtime.clk.Now` | 252:76 |
| `runtime.staleLatchError` | 257:9 |

## State mutations and fallbacks

- 레인 상태는 각 레인의 Offer · RunBounded 만 바꾼다(레인당 goroutine 하나). 관측 맵 · 물결 번호는 record 가 쓰기 잠금으로.

## Safety conclusion

- High-risk 인접(레인 런타임 동시성) — 주문 · 원장 · 활성화 쓰기 없음. 레인끼리 상태 공유 0(Lane 구조), goroutine 하나가 레인 하나, 관측은 자기 색인 칸에만. 생산 레인은 전부 DORMANT(서명 매니페스트 0).
