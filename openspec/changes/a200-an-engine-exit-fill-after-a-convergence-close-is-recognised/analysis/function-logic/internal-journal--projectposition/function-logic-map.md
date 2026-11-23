# Function Logic Map: `ProjectPosition` (편집 뒤 판)

- Source: `internal/journal/position_projection.go` (65-110), 작업 트리 source sha256 `e7e7d30f6153…`. 편집 전: `../../pre-edit-ast/internal-journal--projectposition.ast.json`(65-105, sha `5b62d380b860`, 분기 5).
- AST evidence: `ast.json`(분기 6, return 7: 68 · 75 · 80 · 98 · 105 · 107 · 109). 명령 `go run ./tools/logic-map --file internal/journal/position_projection.go --func ProjectPosition`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버리지: 편집 전 `../../coverage/pre-edit-ProjectPosition-blocks.txt`, 편집 뒤 `../../coverage/after-ProjectPosition-blocks.txt`.
- 재번호: B1~B5 번호 · 줄 유지, 새 B6 if@104(B5 안). 옛 return@102 → return@107, 옛 return@104 → return@109.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx *ApplyTx` | 체결 tx 안에서만 살아 있는 핸들 | `runApplyHooks` 280 | 닫힌 핸들이면 첫 읽기(resolveFillOrigin)가 오류 → B1 |
| `fill AppliedFill` | RecordFill 이 만든 값. 이 체결의 `fill_events` 행(delta>0)은 이미 같은 tx 에 INSERT 됨(fills.go:466 → 훅 :551) | fills.go | 못 읽는 decimal → B4 |
| 생산 진입 조건 | RecordFill 은 `locallyOwned`(소유 확정)일 때만 훅을 부름(fills.go:550) | fills.go | 외부 주문은 이 함수에 안 옴 |
| 불변식 1 | 투영은 (계좌 · 시장 · 종목)의 최신 인스턴스(`currentInstance`)만 움직임 | position_projection.go:275-306 | — |
| 불변식 2 | 거절이면 투영을 쓰지 않음(수량 그대로) — 인정이어도 같음 | B5 → return@105 · @107 | — |
| 불변식 3(a200) | 인정은 차단을 쓰지 않고 표지 한 줄만, 인정이 아니면 오늘의 차단 경로 그대로 | B6 if@104 | 판정 읽기 오류는 tx 오류(체결 롤백) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@67 | `resolveFillOrigin` 오류 | 없음(IC 행은 resolveFillOrigin 안에서) | return@68 err | 편집 무관 |
| B2 if@70 | 소유 intent 없음(외부 · IC) | 없음 | return@75 nil | 편집 무관(RecordFill 경로에선 도달 안 함 — 커버 hit 0) |
| B3 if@79 | `currentInstance` 오류 | 없음 | return@80 err | 편집 무관 |
| B4 if@94 | `position.Apply` 오류(못 읽는 decimal) | 없음 | return@98 감싼 err | 편집 무관 |
| B5 if@101 | `outcome.Reconcile()` — 전이표 거절 | 아래 B6 또는 `reconcile_states` 종목 행 INSERT(활성 행 없을 때) | return@107 `enterReconcileInTx` 결과 | 편집 지점 — 인정 아니면 오늘과 같음 |
| B6 if@104 | 판정 읽기 오류 또는 인정(`recogniseAdjustmentCloseFill`) | 인정이면 표지 `exit_events` 한 줄(새 파일 안) | return@105 err(인정이면 nil) | 새 갈래 — 인정 · 오류 |
| (return@109) | 허용 전이 | `positions` UPDATE/INSERT | `writeProjection` 결과 | 편집 무관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `resolveFillOrigin`@66 | 소유 intent · 범위 · 방향 | 오류 그대로. 인정 판정은 이 결과(`origin.IntentID`)를 받고 다시 유도하지 않음 | CodeGraph + AST |
| `currentInstance`@78 | 최신 인스턴스 | 오류 그대로 | CodeGraph + AST |
| `position.Apply`@83 | 전이표 판정 | 오류 감쌈 | AST |
| `fmt.Errorf`@98 | B4 감쌈 | — | AST |
| `outcome.Reconcile`@101 | 거절 여부 | 순수 | AST |
| `recogniseAdjustmentCloseFill`@104 | 인정 판정 · 표지 쓰기(새 파일 a200_adjustment_close_fill_recognition.go) | 읽기 · 쓰기 오류를 그대로 돌려줌 — B6 이 tx 오류로 올림 | 구조 시험 `TestA200TheRecognitionIsCalledOnceAndReadsNoFillLedger`(생산 호출 이 한 자리) |
| `enterReconcileInTx`@107 | 거절 → 차단 행 | 쓰기 오류면 tx 롤백 | CodeGraph(호출자 1) |
| `writeProjection`@109 | 허용 전이 쓰기 | 쓰기 오류면 tx 롤백 | AST |
| live config | 없음 | — | — |

## State mutations and fallbacks

- 쓰기 셋: B6 의 표지 행(인정일 때만), B5 의 차단 행(인정 아닐 때 · 활성 행이 있으면 쓰지 않음), return@109 의 투영 행. 전부 체결 tx 안이라 실패하면 체결 스냅숏까지 롤백.
- fallback 없음: 판정이 「아님」이면 오늘의 차단 경로 그대로(문구 · 원인 무변).

## Safety conclusion

- Safe edit boundary: B5 안 한 자리(B6). 판정 조건 넷(design D2)과 사실 읽기는 새 파일. `pending_intent_id` 철자 · 체결 누적 표 읽기 · `resolveFillOrigin` 재호출 0(구조 시험).
- High-risk impact: yes — 체결 투영 · 대사 차단(안전 부트스트랩 5). 인정은 차단을 만들지 않는 쪽으로만 바뀌고 수량은 바꾸지 않음.
