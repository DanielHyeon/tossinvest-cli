# Function Logic Map: `strategyLaneRuntime.runLane`

- Source: `internal/app/engine/strategy_lane_runtime.go`
- Source SHA-256: `0935960f05276aa2bf972734943b879a5c56b3525f2083aeed9f5b2e3855c620`
- Signature: `strategyLaneRuntime.runLane(params=4, results=1)`
- Source range: `305:1`–`335:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 레인 안에서 도는 일의 자리 · 생산 정의 본문 · seam 빌드 태그 · 정의 수는 `TestOnlyThePackageLevelStepEverRunsInsideALane` 이 못 박는다 — 변이 R12(생산 정의가 훅을 봄) · R13(seam 태그 확장) CAUGHT(`analysis/measurements/lot-7.5/mutation-7.5.tsv`).

## Branches and early returns

- Exact AST return nodes: `315:3, 334:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 313:2 | 투입 거절(DISABLED · FULL) → 건강만 싣고 반환 |
| B2 | if | 329:2 | 유계 사이클 오류 → 실패 문장 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.Key` | 310:46 |
| `lane.Desired` | 310:67 |
| `lane.Effective` | 310:103 |
| `strategyLaneEvidenceDigest` | 311:57 |
| `lane.Offer` | 312:24 |
| `lane.Health` | 314:24 |
| `lane.RunBounded` | 320:20 |
| `runtime.laneStepFor` | 320:48 |
| `bounded.Err.Error` | 330:25 |
| `lane.Health` | 332:23 |

## State mutations and fallbacks

- 레인 상태는 Offer · RunBounded 만 바꾼다(편집 전과 같음).

## Safety conclusion

- High-risk 인접(레인 런타임 동시성) — 주문 · 원장 · 활성화 쓰기 없음. 레인끼리 상태 공유 0(Lane 구조), goroutine 하나가 레인 하나, 관측은 자기 색인 칸에만. 생산 레인은 전부 DORMANT(서명 매니페스트 0).

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
