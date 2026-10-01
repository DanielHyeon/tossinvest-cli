# Function Logic Map: `strategyLaneRuntime.runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `95611377145456c4906da5ebb95eca97dd025df0c1fb255ebc443c1ce12d5d36`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `286:1`–`316:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 레인 안에서 도는 일의 자리 · 생산 정의 본문 · seam 빌드 태그 · 정의 수는 `TestOnlyThePackageLevelStepEverRunsInsideALane` 이 못 박는다 — 변이 R12(생산 정의가 훅을 봄) · R13(seam 태그 확장) CAUGHT(`analysis/measurements/lot-7.5/mutation-7.5.tsv`).

## Branches and early returns

- Exact AST return nodes: `296:3, 315:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 294:2 | 투입 거절(DISABLED · FULL) → 건강만 싣고 반환 |
| B2 | if | 310:2 | 유계 사이클 오류 → 실패 문장 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 291:46 |
| `lane.Desired` | 291:67 |
| `lane.Effective` | 291:103 |
| `strategyLaneEvidenceDigest` | 292:57 |
| `lane.Offer` | 293:24 |
| `lane.Health` | 295:24 |
| `lane.RunBounded` | 301:20 |
| `runtime.laneStepFor` | 301:48 |
| `bounded.Err.Error` | 311:25 |
| `lane.Health` | 313:23 |

## State mutations and fallbacks

- 레인 상태는 Offer · RunBounded 만 바꾼다(편집 전과 같음).

## Safety conclusion

- High-risk 인접(레인 런타임 동시성) — 주문 · 원장 · 활성화 쓰기 없음. 레인끼리 상태 공유 0(Lane 구조), goroutine 하나가 레인 하나, 관측은 자기 색인 칸에만. 생산 레인은 전부 DORMANT(서명 매니페스트 0).
