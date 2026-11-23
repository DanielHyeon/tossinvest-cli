# Function Logic Map: `ProjectPosition` (편집 전 판)

- Source: `internal/journal/position_projection.go` (65-105), base `884be2a74` 소스 sha256 `5b62d380b860…`.
- AST evidence: `ast.json`(분기 5, return 6: 68 · 75 · 80 · 98 · 102 · 104). 명령 `go run ./tools/logic-map --file internal/journal/position_projection.go --func ProjectPosition`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버리지(편집 전): `../../coverage/pre-edit-ProjectPosition-blocks.txt`.
- 편집 계획(a200 design D1 · D2): B5(if@101) 안에서 enterReconcileInTx@102 를 부르기 **직전**에 새 파일의 인정 판정을 한 번 부르고, 인정이면 차단 쓰기 대신 표지 행을 쓰고 nil 을 돌려줌. 다른 분기 · 호출 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx *ApplyTx` | 체결 tx 안에서만 살아 있는 핸들 | `runApplyHooks` 280 | 닫힌 핸들이면 첫 읽기(resolveFillOrigin)가 오류 → B1 |
| `fill AppliedFill` | RecordFill 이 만든 값. 이 체결의 `fill_events` 행(delta>0)은 이미 같은 tx 에 INSERT 됨(fills.go:466 → 훅 :551) | fills.go | 못 읽는 decimal → B4 |
| 생산 진입 조건 | RecordFill 은 `locallyOwned`(소유 확정)일 때만 훅을 부름(fills.go:550) | fills.go | 외부 주문은 이 함수에 안 옴 |
| 불변식 1 | 투영은 (계좌 · 시장 · 종목)의 최신 인스턴스(`currentInstance`)만 움직임 | position_projection.go:270-301 | — |
| 불변식 2 | 거절이면 투영을 쓰지 않음(수량 그대로) | B5 → return@102 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@67 | `resolveFillOrigin` 오류 | 없음(IC 행은 resolveFillOrigin 안에서) | return@68 err | 편집 무관 |
| B2 if@70 | 소유 intent 없음(외부 · IC) | 없음 | return@75 nil | 편집 무관(RecordFill 경로에선 도달 안 함 — 커버 hit 0) |
| B3 if@79 | `currentInstance` 오류 | 없음 | return@80 err | 편집 무관 |
| B4 if@94 | `position.Apply` 오류(못 읽는 decimal) | 없음 | return@98 감싼 err | 편집 무관 |
| B5 if@101 | `outcome.Reconcile()` — 전이표 거절(OVERSELL · UNATTRIBUTED_SELL · SELL_ON_CLOSED · ENTRY_WHILE_CLOSING) | `reconcile_states` 종목 행 INSERT(활성 행 없을 때) | return@102 `enterReconcileInTx` 결과 | **편집 지점** — 인정이면 차단 대신 표지, 아니면 오늘과 같음 |
| (return@104) | 허용 전이 | `positions` UPDATE/INSERT | `writeProjection` 결과 | 편집 무관 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `resolveFillOrigin`@66 | 소유 intent · 범위 · 방향 | 오류 그대로 반환. 인정 판정은 이 결과(`origin.IntentID`)를 받고 다시 유도하지 않음 | CodeGraph + AST |
| `currentInstance`@78 | 최신 인스턴스 | 오류 그대로 | CodeGraph + AST |
| `position.Apply`@83 | 전이표 판정 | 오류 감쌈 | AST |
| `fmt.Errorf`@98 | B4 감쌈 | — | AST |
| `outcome.Reconcile`@101 | 거절 여부 | 순수 | AST |
| `enterReconcileInTx`@102 | 거절 → 차단 행 | 쓰기 오류면 tx 롤백 | CodeGraph(호출자 1) |
| `writeProjection`@104 | 허용 전이 쓰기 | 쓰기 오류면 tx 롤백 | AST |
| live config | 없음 | — | — |

## State mutations and fallbacks

- 쓰기 둘: B5 의 차단 행(enterReconcileScopeInTx, 활성 행이 있으면 쓰지 않음), return@104 의 투영 행. 둘 다 체결 tx 안이라 실패하면 체결 스냅숏까지 롤백.
- fallback 없음. 편집 뒤 B5 안에 갈래 하나가 생김: 인정 → 표지 행(`exit_events`, 새 action, 증거 열 · `proposed_intent_id` 비움) + nil, 투영 무변.

## Safety conclusion

- Safe edit boundary: B5 안 한 자리. 판정은 새 파일의 함수(원장 읽기 + 순수 비교), `pending_intent_id` 는 `ApplyTx.PendingState` 로만 읽음(구조 시험). 판정 읽기가 오류면 tx 오류(체결 롤백 — 오늘의 쓰기 오류와 같은 방향), 판정이 「아님」이면 오늘의 경로 그대로.
- High-risk impact: yes — 체결 투영 · 대사 차단(안전 부트스트랩 5). 인정은 차단을 만들지 않는 쪽으로만 바뀌고 수량은 바꾸지 않음.
