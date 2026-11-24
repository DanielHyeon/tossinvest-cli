# Function Logic Map: `newAttestationWatch`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--newattestationwatch.ast.json`.

- Source: `internal/app/engine/attestation_watch.go` (262-282)
- AST evidence: `ast.json` — branches **3** / returns **2** / calls **3**, source SHA-256 `85663d710d3b71b71b99efc3a3e60849154d430713ce898d154af5cf0f0e7510`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func newAttestationWatch`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 엔진 Context 에서 감시를 만듦(게이트 · 파일 · 계좌 · 알림기).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Automation` | NewContext(기동 시점 값) | 227-228 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@264) | `if c == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@267) | `if w.clock == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@278) | `if c.Notifier != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 3개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 감시 구조체 채움.

## Safety conclusion

- a199 편집(적용됨): 기동 때의 청산 전용 표지를 감시에 옮김 — 「지금 청산 전용으로 돌고 있다」의 출처(D6 로 프로세스 수명 동안 불변).
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
